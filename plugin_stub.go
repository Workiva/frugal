// +build !go1.8

package main

import (
	"fmt"
	"runtime"
)

// preloadPlugins returns an error if options specifies plugins since they are
// not supported in this build.
// This is currently a workaround to https://github.com/golang/go/issues/17928.
// TODO: remove once workaround is no longer needed.
func preloadPlugins(options map[string]string) error {
	if _, ok := options["plugins"]; ok {
		return fmt.Errorf("Plugins are not supported in %s", runtime.Version())
	}
	return nil
}
