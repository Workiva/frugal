// +build go1.8

package main

import (
	"plugin"
	"strings"
)

// preloadPlugins opens the Plugins specified by the "plugins" option, if any.
// This is currently a workaround to https://github.com/golang/go/issues/17928.
// TODO: remove once workaround is no longer needed.
func preloadPlugins(options map[string]string) error {
	namesStr, ok := options["plugins"]
	if !ok {
		return nil
	}
	names := strings.Split(namesStr, ":")
	for _, name := range names {
		if _, err := plugin.Open(name); err != nil {
			return err
		}
	}
	return nil
}
