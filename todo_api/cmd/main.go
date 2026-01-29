package main

import (
	"os"
	"fmt"
	"github.com/100bench/infr_training/todo_api/app"
)

func main() {
	if err := app.RunApp(); err != nil {
		fmt.Errorf("Application exited with error: %v", err)
		os.Exit(1)
	}
}