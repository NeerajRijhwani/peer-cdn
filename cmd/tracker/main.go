package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/NeerajRijhwani/peer-cdn/api"
	"github.com/NeerajRijhwani/peer-cdn/internal/peer"
	"github.com/NeerajRijhwani/peer-cdn/internal/tracker"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

func main() {
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	peerManager := peer.NewManager(rdb, logger)
	tracker := tracker.NewTracker(peerManager, logger)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/announce", api.ApiHandler(api.Announce, tracker))
	r.Get("/checkinfohash", api.ApiHandler(api.Announce, tracker))

	fmt.Println("Tracker starting on :8080...")
	log.Fatal(http.ListenAndServe(":8080", r))
}
