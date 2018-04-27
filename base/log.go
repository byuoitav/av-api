package base

import (
	"fmt"
	"log"
	"os"
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

	if len(os.Getenv("DEBUG_LOGS")) == 0 {
		/*
			f, err := os.OpenFile("/var/log/av-api.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			if err != nil {
				msg := fmt.Sprintf("Couldn't open log file: %s", err)
				log.Printf(color.HiRedString(msg))
				panic(errors.New(msg))
			}
			defer f.Close()

			log.SetOutput(f)
		*/
	}

	for {
		log.Printf(<-logBuffer)
	}
}
