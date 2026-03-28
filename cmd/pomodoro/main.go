package main

import (
	"log"

	"github.com/lyj404/pomodoro/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
