// +build go1.8

package plugin

import "plugin"

// frugalPlugin implements FrugalPlugin by wrapping a plugin.Plugin.
type frugalPlugin struct {
	name   string
	plugin *plugin.Plugin
}

// Name returns the plugin name.
func (f *frugalPlugin) Name() string {
	return f.name
}

// Lookup returns the symbol with the given name or nil if it doesn't exist.
func (f *frugalPlugin) Lookup(name string) interface{} {
	symbol, _ := f.plugin.Lookup(name)
	return symbol
}

// LoadPlugins opens the FrugalPlugins with the given names and returns an
// error if any fail to open.
func LoadPlugins(names []string) ([]FrugalPlugin, error) {
	plugins := make([]FrugalPlugin, len(names))
	for i, name := range names {
		p, err := plugin.Open(name)
		if err != nil {
			return nil, err
		}
		plugins[i] = &frugalPlugin{name: name, plugin: p}
	}
	return plugins, nil
}
