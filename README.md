# Pomodoro

A desktop Pomodoro timer application built with Go and Fyne, with local SQLite storage.

![Screenshot](/img/img.png)

[中文](/README_zh.md)

## Features

- Pomodoro countdown timer
- Work / Short Break / Long Break modes
- Start, Pause, Reset, Skip controls
- Today's focus statistics
- History viewing and batch deletion
- System notification reminders
- Local settings persistence

## Tech Stack

- Go
- Fyne
- SQLite

## Run

```bash
go run ./cmd/pomodoro/main.go
```

## Build

```bash
# Build with go
go build -o pomodoro ./cmd/pomodoro/main.go

# Build with make (smaller binary with -ldflags)
make build

# Build for all platforms
make build-all
```

## Data Storage

Application settings and history are stored in a local SQLite database.
