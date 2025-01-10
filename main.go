package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/Nukambe/ctu/internal/rtsp_feeds"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type CTU struct {
	Environment string
	JWTSecret   string
	Feeds       map[string]rtsp_feeds.Feed
	hlsDir      string
}

func addCORSHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		next.ServeHTTP(w, r)
	})
}

func addRevalidateCacheHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		next.ServeHTTP(w, r)
	})
}

func main() {
	// load env
	if err := godotenv.Load(); err != nil {
		log.Printf("error loading env variables: %s", err)
		os.Exit(1)
	}

	// Set up signal channel to capture termination signals (Ctrl+C, etc.)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// const variables
	port := 8080
	hlsDir := "./hls"
	ctu := CTU{
		Environment: os.Getenv("ENVIRONMENT"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		Feeds:       rtsp_feeds.GetFeeds(hlsDir),
	}

	// create server
	mux := http.NewServeMux()
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}
	mux.Handle("/hls/",
		addCORSHeaders(
			addRevalidateCacheHeaders(
				http.StripPrefix("/hls",
					http.FileServer(http.Dir(hlsDir))))))
	mux.Handle("GET /feeds", addCORSHeaders(http.HandlerFunc(rtsp_feeds.HandleGetFeeds(ctu.Feeds))))

	// start server
	go func() {
		log.Printf("Starting server on port %d", port)
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %s", err)
		}
	}()

	// Wait for a termination signal
	sig := <-signalChan
	log.Printf("Received signal: %v. Shutting down.", sig)

	// Shutdown HTTP server with a timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %s", err)
	} else {
		log.Println("HTTP server stopped.")
	}

	// Kill FFmpeg processes
	errs := rtsp_feeds.KillFFmpeg(ctu.Feeds)
	if errs != nil {
		for _, err := range errs {
			log.Printf("Error killing FFmpeg: %s", err)
		}
	} else {
		log.Println("All FFmpeg processes stopped.")
	}

	log.Println("Shutdown complete.")
}
