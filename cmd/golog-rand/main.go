package main

import (
	"crypto/rand"
	"encoding/base64"
	"flag"
	"log"
	"time"

	"github.com/bingoohuang/golog"
)

func main() {
	interval := flag.Duration("interval", 1*time.Second, "interval lime, like 10s, default 10ms")
	size := flag.Int("size", 1000, "rand string size")
	flag.Parse()

	golog.Setup()

	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	token := make([]byte, *size)

	for range ticker.C {
		rand.Read(token)
		log.Printf("Rand: %s", base64.URLEncoding.EncodeToString(token))
	}
}
