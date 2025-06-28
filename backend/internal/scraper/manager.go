package scraper

import (
	"context"
	"fmt"
	"time"
)

// Manager orchestrates the scraping operations
type Manager struct {
	registry *Registry
}

// NewManager creates a new scraper manager
func NewManager() *Manager {
	return &Manager{
		registry: NewRegistry(),
	}
}

// RegisterPlugin registers a new plugin
func (m *Manager) RegisterPlugin(plugin Plugin) error {
	return m.registry.Register(plugin)
}

// ScrapeByType automatically selects the appropriate plugin based on source type
func (m *Manager) ScrapeByType(ctx context.Context, req *ScrapeRequest) (*ScrapeResult, error) {
	plugin, err := m.registry.GetPluginForType(req.SourceType)
	if err != nil {
		return nil, fmt.Errorf("failed to find plugin for type %s: %w", req.SourceType, err)
	}

	return m.scrapeWithPlugin(ctx, plugin, req)
}

// ScrapeByPlugin uses a specific plugin by name
func (m *Manager) ScrapeByPlugin(ctx context.Context, pluginName string, req *ScrapeRequest) (*ScrapeResult, error) {
	plugins := m.registry.ListPlugins()
	plugin, ok := plugins[pluginName]
	if !ok {
		return nil, fmt.Errorf("plugin %s not found", pluginName)
	}

	return m.scrapeWithPlugin(ctx, plugin, req)
}

// scrapeWithPlugin performs the actual scraping with timing and error handling
func (m *Manager) scrapeWithPlugin(ctx context.Context, plugin Plugin, req *ScrapeRequest) (*ScrapeResult, error) {
	// Validate request
	if err := plugin.ValidateRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Add timeout if not present
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 5*time.Minute)
		defer cancel()
	}

	// Track timing
	start := time.Now()

	// Perform scraping
	result, err := plugin.Scrape(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("scraping failed: %w", err)
	}

	// Ensure result stats are populated
	if result.Stats.Duration == "" {
		result.Stats.Duration = time.Since(start).String()
	}
	if result.Stats.ProductsFound == 0 {
		result.Stats.ProductsFound = len(result.Products)
	}
	if result.Stats.ErrorCount == 0 {
		result.Stats.ErrorCount = len(result.Errors)
	}
	if result.Stats.Source == "" {
		result.Stats.Source = req.Reseller
	}

	// Count variants if not already counted
	if result.Stats.VariantsFound == 0 {
		for _, product := range result.Products {
			result.Stats.VariantsFound += len(product.Variants)
		}
	}

	return result, nil
}

// GetPluginInfo returns information about a plugin
func (m *Manager) GetPluginInfo(pluginName string) (map[string]interface{}, error) {
	plugins := m.registry.ListPlugins()
	plugin, ok := plugins[pluginName]
	if !ok {
		return nil, fmt.Errorf("plugin %s not found", pluginName)
	}

	return map[string]interface{}{
		"name":            plugin.Name(),
		"supported_types": plugin.SupportedTypes(),
	}, nil
}

// ListAvailablePlugins returns information about all registered plugins
func (m *Manager) ListAvailablePlugins() map[string]map[string]interface{} {
	plugins := m.registry.ListPlugins()
	result := make(map[string]map[string]interface{})

	for name, plugin := range plugins {
		result[name] = map[string]interface{}{
			"name":            plugin.Name(),
			"supported_types": plugin.SupportedTypes(),
		}
	}

	return result
}

// ScrapeMultiplePages scrapes multiple pages by creating separate requests
// This is a utility method for handling pagination at the manager level
func (m *Manager) ScrapeMultiplePages(ctx context.Context, baseReq *ScrapeRequest, maxPages int) (*ScrapeResult, error) {
	var allProducts []ScrapedProduct
	var allErrors []string
	totalVariants := 0

	for page := 1; page <= maxPages; page++ {
		// Clone request and add page parameter
		req := *baseReq
		if req.Options == nil {
			req.Options = make(map[string]string)
		}
		req.Options["page"] = fmt.Sprintf("%d", page)

		result, err := m.ScrapeByType(ctx, &req)
		if err != nil {
			allErrors = append(allErrors, fmt.Sprintf("Page %d: %v", page, err))
			continue
		}

		// If no products found, we've reached the end
		if len(result.Products) == 0 {
			break
		}

		allProducts = append(allProducts, result.Products...)
		allErrors = append(allErrors, result.Errors...)
		totalVariants += result.Stats.VariantsFound
	}

	return &ScrapeResult{
		Products: allProducts,
		Errors:   allErrors,
		Stats: ScrapeStats{
			ProductsFound: len(allProducts),
			VariantsFound: totalVariants,
			ErrorCount:    len(allErrors),
			Source:        baseReq.Reseller,
		},
	}, nil
}

// ValidateRequest validates a scrape request across all plugins
func (m *Manager) ValidateRequest(req *ScrapeRequest) error {
	if req.URL == "" {
		return fmt.Errorf("URL is required")
	}
	if req.SourceType == "" {
		return fmt.Errorf("source type is required")
	}
	if req.Reseller == "" {
		return fmt.Errorf("reseller is required")
	}
	if req.Category == "" {
		return fmt.Errorf("category is required")
	}

	// Find plugin and validate with it
	plugin, err := m.registry.GetPluginForType(req.SourceType)
	if err != nil {
		return fmt.Errorf("no plugin available for source type %s", req.SourceType)
	}

	return plugin.ValidateRequest(req)
}
