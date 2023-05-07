package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/bay0/wslogger"
	"github.com/brianvoe/gofakeit"
	"github.com/sirupsen/logrus"
)

type LogLevel string

const (
	Info  LogLevel = "INFO"
	Warn  LogLevel = "WARN"
	Error LogLevel = "ERROR"
)

func randomLogLevel() LogLevel {
	levels := []LogLevel{Info, Warn, Error}
	return levels[rand.Intn(len(levels))]
}

func randomLogMessage(r *rand.Rand) string {
	components := []func() string{
		func() string { return fmt.Sprintf("User ID: %d", gofakeit.Number(1000, 9999)) },
		func() string { return fmt.Sprintf("IP: %s", gofakeit.IPv4Address()) },
		func() string { return fmt.Sprintf("Method: %s", gofakeit.HTTPMethod()) },
		func() string { return fmt.Sprintf("URL: %s", gofakeit.URL()) },
		func() string { return fmt.Sprintf("Error: %s", gofakeit.HackerPhrase()) },
	}

	numComponents := r.Intn(5) + 1
	var message string
	for i := 0; i < numComponents; i++ {
		componentIndex := r.Intn(len(components))
		message += components[componentIndex]() + " "
		components = append(components[:componentIndex], components[componentIndex+1:]...)
	}

	return message
}

func main() {
	gofakeit.Seed(0)
	localRand := rand.New(rand.NewSource(time.Now().UnixNano()))

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

	numLogs := 1000000
	sleepDuration := 1000 * time.Millisecond

	for i := 0; i < numLogs; i++ {
		randomLevel := randomLogLevel()
		message := randomLogMessage(localRand)

		switch randomLevel {
		case Info:
			logrus.Info(message)
		case Warn:
			logrus.Warn(message)
		case Error:
			logrus.Error(message)
		}

		time.Sleep(sleepDuration)
	}
}
