package main

import (
	"log"
	"os"
	"runtime"

	"github.com/lyj404/pomodoro/internal/app"
)

func init() {
	if os.Getenv("GOGC") == "" {
		os.Setenv("GOGC", "200")
	}
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
