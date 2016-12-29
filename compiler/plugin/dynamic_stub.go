// +build !go1.8

package plugin

import (
	"fmt"
	"runtime"
)

// frugalPlugin implements FrugalPlugin by acting as a stub since the plugin
// package requires go1.8+.
type frugalPlugin struct {
	name string
}

// Name returns the plugin name.
func (f *frugalPlugin) Name() string {
	return f.name
}

// Lookup returns nil since plugins are not supported in this build.
func (f *frugalPlugin) Lookup(name string) interface{} {
	return nil
}

// LoadPlugins returns an error since plugins are not supported in this build.
func LoadPlugins(names []string) ([]FrugalPlugin, error) {
	return nil, fmt.Errorf("Plugins are not supported in %s", runtime.Version())
}
