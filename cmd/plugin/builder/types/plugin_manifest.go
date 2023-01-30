package types

// PluginManifest defines the plugin manifest for plugin metadata
type PluginManifest struct {
	Plugins []PluginMetadata `json:"plugins" yaml:"plugins"`
}

// PluginMetadata specifies plugin metadata
type PluginMetadata struct {
	Name    string `json:"name" yaml:"name"`
	Target  string `json:"target" yaml:"target"`
	Version string `json:"version" yaml:"version"`
	Path    string `json:"path" yaml:"path"`
}
