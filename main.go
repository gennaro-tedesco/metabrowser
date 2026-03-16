package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"metabrowser/scanner"
	"metabrowser/server"
)

//go:embed static
var static embed.FS

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <library-path>\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}

	root, err := filepath.Abs(os.Args[1])
	if err != nil {
		log.Fatalf("invalid path: %v", err)
	}

	if _, err := os.Stat(root); err != nil {
		log.Fatalf("cannot access %s: %v", root, err)
	}

	log.Printf("scanning %s ...", root)
	lib, err := scanner.Scan(root)
	if err != nil {
		log.Fatalf("scan failed: %v", err)
	}
	log.Printf("found %d books", len(lib.Books))

	addr := "localhost:7070"
	if len(os.Args) >= 3 {
		addr = os.Args[2]
	}

	server.Serve(addr, root, lib, static)
}
