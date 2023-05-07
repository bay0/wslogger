# wslogger

[![Go Report Card](https://goreportcard.com/badge/github.com/bay0/wslogger)](https://goreportcard.com/report/github.com/bay0/wslogger)
![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/bay0/wslogger)
![GitHub Workflow Status (with event)](https://img.shields.io/github/actions/workflow/status/bay0/wslogger/test.yml)
![GitHub last commit](https://img.shields.io/github/last-commit/bay0/wslogger)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/bay0/wslogger)

![Demo GIF of the wslogger in action.](./example/demo.gif)

`wslogger` is a simple WebSocket-based logging helper for Go.

It offers the following functionality:

* Real-time log broadcasting: broadcasts log messages to connected clients in real-time
* Multiple clients support: handles multiple WebSocket clients connected simultaneously
* Buffering: maintains a buffer of log messages to avoid blocking the application
* Customizable logging: easily integrates with popular logging libraries like `logrus` and `zap`

This library defines two main types:

`WSLogger`: the core WebSocket-based logger that handles client connections and message broadcasting

`WSWriter`: an `io.Writer` implementation that writes log messages to the `WSLogger` broadcast channel

## Installation

Use `go get` to install wslogger.

```bash
go get github.com/bay0/wslogger
```

## Usage

To use the library, import the wslogger package and create a new instance of WSLogger using NewWSLogger().

```go
package main

import (
 "io"
 "log"
 "net/http"
 "os"
 "time"

 "github.com/bay0/wslogger"

 "github.com/sirupsen/logrus"
)

func main() {
 wsLogger := wslogger.NewWSLogger()
 wsLogger.Start()

 http.HandleFunc("/ws", wsLogger.HandleConnections)
 http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
  http.ServeFile(w, r, "index.html")
 })

 go func() {
  log.Fatal(http.ListenAndServe(":8000", nil))
 }()

 wsWriter := wsLogger.NewWSWriter()

 logrus.SetFormatter(&logrus.JSONFormatter{
  TimestampFormat: time.RFC3339Nano,
 })

 logrus.SetOutput(io.MultiWriter(os.Stdout, wsWriter))

 for i := 0; i < 100000; i++ {
  logrus.Infof("This is log message #%d", i+1)
  logrus.Warnf("This is log message #%d", i+1)
  logrus.Errorf("This is log message #%d", i+1)

  time.Sleep(1 * time.Second)
 }
}
```

In this example, the wslogger package is integrated with the logrus logging library to broadcast log messages to WebSocket clients.

The WebSocket server listens on port 8000 and serves an index.html file for clients to connect and receive log messages.

Please refer to the provided example HTML and JavaScript code to implement the client-side WebSocket connection and log message rendering.
