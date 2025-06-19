package jobs

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/jonathanhecl/gollama"
	"github.com/meta-boy/mech-alligator/internal/domain/job"
)

func getJSONSchema(t reflect.Type) map[string]interface{} {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"tags": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
		},
		"required": []string{"tags"},
	}
	return schema
}

// cleanLLMResponse removes markdown code blocks and extra formatting
func cleanLLMResponse(response string) string {
	// Remove markdown code blocks
	response = strings.TrimSpace(response)

	// Remove ```json and ``` markers
	if strings.HasPrefix(response, "```json") {
		response = strings.TrimPrefix(response, "```json")
	}
	if strings.HasPrefix(response, "```") {
		response = strings.TrimPrefix(response, "```")
	}
	if strings.HasSuffix(response, "```") {
		response = strings.TrimSuffix(response, "```")
	}

	return strings.TrimSpace(response)
}

type TagPayload struct {
	ProductID string `json:"product_id"`
}

type LLMAnswer struct {
	Tags []string `json:"tags" required:"true"`
}

type TagJobHandler struct {
	db     *sql.DB
	ollama *gollama.Gollama
}

var systemPrompt string

func init() {
	schema := getJSONSchema(reflect.TypeOf(LLMAnswer{}))
	schemaJSON, _ := json.Marshal(schema)

	allowedTags := []string{
		"keyboard", "keycaps", "switches", "accessories",
		"linear", "tactile", "clicky", "silent",
		"full_size", "tkl", "compact", "split", "ergonomic",
		"hot_swap", "wireless", "rgb", "programmable",
	}

	allowedTagsJSON, _ := json.Marshal(allowedTags)

	systemPrompt = fmt.Sprintf(`You are a helpful assistant that generates product tags from a predefined list. 
	
JSON Schema: %s

Allowed tags (you must only use tags from this list): %s

Rules:
- Only return tags that exist in the allowed tags list
- Return relevant tags based on the product description
- Do not create new tags or modify existing ones
- Return empty array if no relevant tags are found`, string(schemaJSON), string(allowedTagsJSON))
}

func NewTagJobHandler(db *sql.DB) (*TagJobHandler, error) {
	g := gollama.New("gemma3:12b")
	g.SystemPrompt = systemPrompt

	return &TagJobHandler{db: db, ollama: g}, nil
}

func (h *TagJobHandler) GetType() job.JobType {
	return job.JobTypeTagProduct
}

func (h *TagJobHandler) Handle(ctx context.Context, j *job.Job) error {
	var payload TagPayload
	payloadBytes, err := json.Marshal(j.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	var name, description string
	query := "SELECT name, description FROM products WHERE id = $1"
	err = h.db.QueryRowContext(ctx, query, payload.ProductID).Scan(&name, &description)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("product with id %s not found", payload.ProductID)
		}
		return fmt.Errorf("failed to fetch product details: %w", err)
	}

	prompt := fmt.Sprintf(`
Generate tags for the following product from the predefined list of allowed tags only.

Product name: %s
Description: %s

Instructions:
- Only use tags from the allowed tags list provided in the system prompt
- Select the most relevant tags based on the product description
- Return only valid JSON in the specified format
- Do not include any additional text or markdown
`, name, description)

	output, err := h.ollama.Chat(ctx, prompt)
	if err != nil {
		return fmt.Errorf("failed to get response from ollama: %w", err)
	}

	// Clean the response before unmarshaling
	cleanedResponse := cleanLLMResponse(output.Content)

	var llmResponse LLMAnswer
	if err := json.Unmarshal([]byte(cleanedResponse), &llmResponse); err != nil {
		return fmt.Errorf("failed to unmarshal llm response: %w. Raw response: %s", err, output.Content)
	}

	if len(llmResponse.Tags) > 0 {
		tx, err := h.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		var committed bool
		defer func() {
			if !committed {
				if rollbackErr := tx.Rollback(); rollbackErr != nil {
					fmt.Printf("failed to rollback transaction: %v\n", rollbackErr)
				}
			}
		}()

		stmt, err := tx.PrepareContext(ctx, "INSERT INTO product_tags (product_id, tag) VALUES ($1, $2) ON CONFLICT DO NOTHING")
		if err != nil {
			return fmt.Errorf("failed to prepare statement: %w", err)
		}
		defer func(stmt *sql.Stmt) {
			err := stmt.Close()
			if err != nil {
				fmt.Printf("failed to close statement: %v\n", err)
			} else {
				fmt.Println("statement closed")
			}
		}(stmt)

		for _, tag := range llmResponse.Tags {
			_, err := stmt.ExecContext(ctx, payload.ProductID, tag)
			if err != nil {
				return fmt.Errorf("failed to insert tag '%s': %w", tag, err)
			}
		}

		return tx.Commit()
	}

	return nil
}
