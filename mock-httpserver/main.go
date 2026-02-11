package main

import (
	"encoding/json"
	"flag"
	"fmt"
        "log"
	"net/http"
)

func main() {
	// 1. 引数の定義 (フラグ名, デフォルト値, 説明)
	port := flag.String("p", "8080", "port number to listen on")
	flag.Parse()

	mux := http.NewServeMux()

	// 特定のホストIDパス
	mux.HandleFunc("PUT /api/v0/hosts/{id}", func(w http.ResponseWriter, r *http.Request) {
                log.Printf("Received: %s %s", r.Method, r.URL.Path)
		id := r.PathValue("id")
		renderJSON(w, map[string]string{"id": id})
	})

	// それ以外すべて
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
                log.Printf("Received: %s %s", r.Method, r.URL.Path)
		renderJSON(w, map[string]bool{"success": true})
	})

	// 2. 指定されたポートで起動
	addr := ":" + *port
	fmt.Printf("Server starting on %s...\n", addr)
	
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Printf("Error: %s\n", err)
	}
}

func renderJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}
