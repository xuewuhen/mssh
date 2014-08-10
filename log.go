package main

import (
	"github.com/skoo87/log4go"
)

func SetLog() {
		
	w2 := log4go.NewConsoleWriter()

	log4go.Register(w2)
	log4go.SetLevel(log4go.INFO)
}
