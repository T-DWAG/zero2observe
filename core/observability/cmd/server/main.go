package main

import (
	"log"
	"net/http"
	"os"

	"github.com/T-DWAG/zero2observe/api"
	"github.com/T-DWAG/zero2observe/evaluation"
	"github.com/T-DWAG/zero2observe/model"
	"github.com/T-DWAG/zero2observe/storage"
)

func main() {
	addr := envOr("OBS_HTTP_ADDR", ":8080")

	store := mustStore()
	judge := evaluation.NewJudge(store, &evaluation.FakeCompleter{})
	srv := api.NewServer(store).WithJudge(judge)

	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, srv.Handler()))
}

func mustStore() storage.Storage {
	dsn := os.Getenv("OBS_PG_DSN")
	if dsn == "" {
		log.Printf("storage: memory (set OBS_PG_DSN to use postgres)")
		return storage.NewMemoryStorage()
	}

	db, err := storage.OpenPostgres(dsn)
	if err != nil {
		log.Fatal(err)
	}
	if err := model.AutoMigrate(db); err != nil {
		log.Fatal(err)
	}
	log.Printf("storage: postgres")
	return storage.NewPostgresStorage(db)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
