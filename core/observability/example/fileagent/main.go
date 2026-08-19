package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/T-DWAG/zero2observe/collector"
	"github.com/T-DWAG/zero2observe/model"
	"github.com/T-DWAG/zero2observe/storage"
)

func main() {
	q := flag.String("q", "sandbox 里有哪些文件？obs_hints.md 在讲什么？", "用户问题")
	redact := flag.Bool("redact", false, "开启正文打码")
	noContent := flag.Bool("no-content", false, "不落盘正文")
	flag.Parse()

	var store storage.Storage
	if dsn := os.Getenv("OBS_PG_DSN"); dsn != "" {
		db, err := storage.OpenPostgres(dsn)
		if err != nil {
			log.Fatal(err)
		}
		if err := model.AutoMigrate(db); err != nil {
			log.Fatal(err)
		}
		store = storage.NewPostgresStorage(db)
		log.Printf("storage: postgres")
	} else {
		store = storage.NewMemoryStorage()
		log.Printf("storage: memory")
	}

	cfg := collector.Config{
		SessionID: "file-demo",
		UserInput: *q,
		Redact:    *redact,
		NoContent: *noContent,
	}

	traceID, answer, err := runOnce(context.Background(), store, cfg, *q)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("provider: %s\n", envOr("OBS_LLM_PROVIDER", "openai"))
	fmt.Printf("trace_id: %s\n", traceID)
	fmt.Printf("answer:   %s\n", answer)

	tr, err := store.GetTrace(context.Background(), traceID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("spans:    %d  status=%s  tokens=%d\n", tr.SpanCount, tr.Status, tr.TotalTokens)
	fmt.Printf("user_in:  %q\n", tr.UserInput)
}
