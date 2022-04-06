//go:build !testing
// +build !testing

package log

import "log"

func init() {
	log.SetOutput(defaultWriter)
}
