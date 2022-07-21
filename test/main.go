package test

import (
	"log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	if os.Getenv("LOCAL_IP") == "" {
		log.Fatalln("LOCAL_IP environment variable must be set")
	}

	os.Exit(m.Run())
}
