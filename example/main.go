package main

import (
	"log"
	"net/http"
	"github.com/supme/elFinder"
	"fmt"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle("/elf/", http.StripPrefix("/elf/", http.FileServer(http.Dir("./elf/"))))

	config := elFinder.Config{}
	config["l0"] = elFinder.Volume {
		Root: "./files",
		AllowDirs: []string{"/Allow"},
		DenyDirs:  []string{"/Deny"},
		DefaultRight: false,
	}
	mux.Handle("/connector", elFinder.NetHttp(config))

	fmt.Println("Listen on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}