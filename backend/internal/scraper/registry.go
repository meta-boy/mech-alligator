package scraper

import (
	"fmt"
	"sync"
)

// Registry manages all registered plugins
type Registry struct {
	plugins map[string]Plugin
	mu      sync.RWMutex
}

// NewRegistry creates a new plugin registry
func NewRegistry() *Registry {
	return &Registry{
		plugins: make(map[string]Plugin),
	}
}

// Register adds a plugin to the registry
func (r *Registry) Register(plugin Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	name := plugin.GetName()
	if _, exists := r.plugins[name]; exists {
		return fmt.Errorf("plugin %s already registered", name)
	}
	
	r.plugins[name] = plugin
	return nil
}

// GetPluginForType finds the first plugin that supports the given site type
func (r *Registry) GetPluginForType(siteType string) (Plugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	for _, plugin := range r.plugins {
		for _, supportedType := range plugin.GetSupportedTypes() {
			if supportedType == siteType {
				return plugin, nil
			}
		}
	}
	
	return nil, fmt.Errorf("no plugin found for site type %s", siteType)
}

// ListPlugins returns all registered plugins
func (r *Registry) ListPlugins() map[string]Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	result := make(map[string]Plugin)
	for name, plugin := range r.plugins {
		result[name] = plugin
	}
	
	return result
}