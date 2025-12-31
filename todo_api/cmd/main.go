package main

import (
	"log"
	"os"
	"github.com/100bench/infr_training/app"
)

func main() {
	if err := app.RunApp(); err != nil {
		log.Printf("Application exited with error: %v", err)
		os.Exit(1)
	}
}