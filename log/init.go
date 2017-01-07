package log

import "log"

func init() {
	log.SetOutput(defaultWriter)
}
