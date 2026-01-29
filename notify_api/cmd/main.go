package main

import (
	"fmt"
	"os"

	"github.com/100bench/infr_training/notify_api/app"
)

func main() {
      if err := app.RunApp(); err != nil {
          fmt.Println("application error:", err)
          os.Exit(1)
      }
  }