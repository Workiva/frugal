package plugin

// FrugalPlugin wraps a dynamically linked plugin.
type FrugalPlugin interface {
	// Name returns the plugin name.
	Name() string

	// Lookup returns the symbol with the given name or nil if it doesn't
	// exist.
	Lookup(name string) interface{}
}
