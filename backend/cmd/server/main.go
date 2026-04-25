package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	dsn := os.Getenv("DATABASE_URL")

	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/api/hello", func(w http.ResponseWriter, r *http.Request) {
		count, err := bumpVisits(r.Context(), dsn)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"message":  "hello from galley fullstack test",
			"visits":   count,
			"hostname": hostnameOrUnknown(),
		})
	})

	log.Printf("listening on :%s (db=%t)", port, dsn != "")
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}

// bumpVisits inserts one row into a counter table and returns the
// resulting count. Schema is created lazily on first hit so the service
// boots even if Postgres is still warming up.
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
