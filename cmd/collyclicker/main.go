//cmd/collyclicker/main.go

package main

import (
	"collyclicker/internal/app"
	"log"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatalf("error: %v", err)
	}
}
