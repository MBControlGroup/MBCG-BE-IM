package util

import (
	"log"
	"runtime/debug"
)

// Try try & catch
func Try(fun func(), handler func(err interface{})) {
	defer func() {
		if err := recover(); err != nil {
			// handle exception
			handler(err)
		}
	}()

	// call user's function
	fun()
}

func logException(err error) {
	log.Println("Got error", err)
	debug.PrintStack()
}
