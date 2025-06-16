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

// ScrapeByType automatically selects the appropriate plugin based on site type
func (m *Manager) ScrapeByType(ctx context.Context, req *ScrapeRequest) (*ScrapeResult, error) {
	plugin, err := m.registry.GetPluginForType(req.SiteType)
	if err != nil {
		return nil, fmt.Errorf("failed to find plugin for type %s: %w", req.SiteType, err)
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
	if err := plugin.Validate(req); err != nil {
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
	
	// Update metadata
	duration := time.Since(start)
	result.Metadata.ScrapedAt = start
	result.Metadata.Duration = duration.String()
	result.Metadata.PluginName = plugin.GetName()
	result.Metadata.PluginVersion = plugin.GetVersion()
	result.Metadata.TotalFound = len(result.Products)
	result.Metadata.TotalErrors = len(result.Errors)
	
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
		"name":                 plugin.GetName(),
		"version":              plugin.GetVersion(),
		"supported_types":      plugin.GetSupportedTypes(),
		"required_credentials": plugin.GetRequiredCredentials(),
		"supported_options":    plugin.GetSupportedOptions(),
	}, nil
}

// ListAvailablePlugins returns information about all registered plugins
func (m *Manager) ListAvailablePlugins() map[string]map[string]interface{} {
	plugins := m.registry.ListPlugins()
	result := make(map[string]map[string]interface{})
	
	for name, plugin := range plugins {
		result[name] = map[string]interface{}{
			"name":                 plugin.GetName(),
			"version":              plugin.GetVersion(),
			"supported_types":      plugin.GetSupportedTypes(),
			"required_credentials": plugin.GetRequiredCredentials(),
			"supported_options":    plugin.GetSupportedOptions(),
		}
	}
	
	return result
}