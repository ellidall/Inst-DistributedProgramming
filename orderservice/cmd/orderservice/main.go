package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"

	"orderservise/pkg/server"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	file, err := os.OpenFile("my.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.SetOutput(file)
		defer file.Close()
	}

	serverPort := ":8000"
	killSignalChan := getKillSignalChannel()
	srv := startServer(serverPort)
	waitForKillSignal(killSignalChan)
	err = srv.Shutdown(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}

func startServer(serverPort string) *http.Server {
	router := server.Router()
	srv := &http.Server{Addr: serverPort, Handler: router}
	go func() {
		log.Fatal(srv.ListenAndServe())
	}()

	return srv
}

func getKillSignalChannel() <-chan os.Signal {
	osKillSignalChannel := make(chan os.Signal, 1)
	signal.Notify(osKillSignalChannel, os.Interrupt, syscall.SIGTERM)
	return osKillSignalChannel
}

func waitForKillSignal(killSignalChannel <-chan os.Signal) {
	killSignal := <-killSignalChannel
	switch killSignal {
	case os.Interrupt:
		log.Info("Got SIGINT, shutting down")
	case syscall.SIGTERM:
		log.Info("Got SIGTERM, shutting down")
	}
}
