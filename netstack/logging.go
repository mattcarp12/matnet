package netstack

import (
	"log"
	"os"
)

func NewLogger(name string) *log.Logger {
	return log.New(os.Stdout, "["+name+"] ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
}
