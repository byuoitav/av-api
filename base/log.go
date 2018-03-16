package base

import (
	"fmt"
	"log"
)

var logBuffer chan string

func init() {
	logBuffer = make(chan string, 100000)

	go logger()
}

func Log(format string, values ...interface{}) {

	logBuffer <- fmt.Sprintf(format, values...)
}

func logger() {
	for {
		log.Printf(<-logBuffer)
	}
}
