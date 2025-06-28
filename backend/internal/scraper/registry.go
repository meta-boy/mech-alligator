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

	name := plugin.Name()
	if _, exists := r.plugins[name]; exists {
		return fmt.Errorf("plugin %s already registered", name)
	}

	r.plugins[name] = plugin
	return nil
}

// GetPluginForType finds the first plugin that supports the given source type
func (r *Registry) GetPluginForType(sourceType string) (Plugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, plugin := range r.plugins {
		for _, supportedType := range plugin.SupportedTypes() {
			if supportedType == sourceType {
				return plugin, nil
			}
		}
	}

	return nil, fmt.Errorf("no plugin found for source type %s", sourceType)
}

// GetPlugin returns a specific plugin by name
func (r *Registry) GetPlugin(name string) (Plugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugin, exists := r.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", name)
	}

	return plugin, nil
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

// ListSupportedTypes returns all supported source types across all plugins
func (r *Registry) ListSupportedTypes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	typeSet := make(map[string]bool)
	for _, plugin := range r.plugins {
		for _, sourceType := range plugin.SupportedTypes() {
			typeSet[sourceType] = true
		}
	}

	var types []string
	for sourceType := range typeSet {
		types = append(types, sourceType)
	}

	return types
}
