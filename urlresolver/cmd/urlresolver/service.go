package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

func runServer(cfg *Config) error {
	router := mux.NewRouter()

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}).Methods("GET")

	router.HandleFunc("/{shortPath}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		path := "/" + vars["shortPath"]

		if longURL, ok := cfg.ShortenUrls[path]; ok {
			http.Redirect(w, r, longURL, http.StatusFound)
			return
		}
		http.NotFound(w, r)
	}).Methods("GET")

	server := &http.Server{
		Addr:              cfg.ServeAddr,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		fmt.Printf("Server starting on %s%s\n", cfg.HostName, server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("Server failed: %v\n", err)
			os.Exit(1)
		}
	}()

	return gracefulShutdown(server)
}

func gracefulShutdown(server *http.Server) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	fmt.Println("\nShutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %v", err)
	}

	fmt.Println("Server stopped gracefully")
	return nil
}
