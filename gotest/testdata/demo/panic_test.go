package gotest

import (
	"log"
	"testing"
)

func TestPanic(t *testing.T) {
	log.Panic("This is panic testcase")
}
