package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	dsn := os.Getenv("DATABASE_URL")
	redisURL := os.Getenv("REDIS_URL")

	var rdb *redis.Client
	if redisURL != "" {
		opts, err := redis.ParseURL(redisURL)
		if err != nil {
			log.Printf("invalid REDIS_URL: %v", err)
		} else {
			rdb = redis.NewClient(opts)
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/api/hello", func(w http.ResponseWriter, r *http.Request) {
		count, cached, err := visitCount(r.Context(), dsn, rdb)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"message":  "hello from galley fullstack test",
			"visits":   count,
			"cached":   cached,
			"hostname": hostnameOrUnknown(),
		})
	})

	log.Printf("listening on :%s (db=%t, cache=%t)", port, dsn != "", rdb != nil)
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}

const cacheKey = "visits:count"
const cacheTTL = 5 * time.Second

// visitCount returns the live visit count, serving from Redis when warm and
// falling back to a Postgres write+count when the cache is cold or absent.
// The bool reports whether the value came from cache.
func visitCount(ctx context.Context, dsn string, rdb *redis.Client) (int, bool, error) {
	if rdb != nil {
		if v, err := rdb.Get(ctx, cacheKey).Int(); err == nil {
			return v, true, nil
		}
	}
	n, err := bumpVisits(ctx, dsn)
	if err != nil {
		return 0, false, err
	}
	if rdb != nil {
		_ = rdb.Set(ctx, cacheKey, strconv.Itoa(n), cacheTTL).Err()
	}
	return n, false, nil
}

func bumpVisits(ctx context.Context, dsn string) (int, error) {
	if dsn == "" {
		return 0, nil
	}
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return 0, err
	}
	defer conn.Close(ctx)

	if _, err := conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS visits (
			id   BIGSERIAL PRIMARY KEY,
			at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		return 0, err
	}
	if _, err := conn.Exec(ctx, `INSERT INTO visits DEFAULT VALUES`); err != nil {
		return 0, err
	}
	var n int
	if err := conn.QueryRow(ctx, `SELECT COUNT(*) FROM visits`).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

func hostnameOrUnknown() string {
	h, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return h
}
