package main

import (
	"log"
	"net/http"
	"github.com/supme/goElFinder"
	"fmt"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle("/elf/", http.StripPrefix("/elf/", http.FileServer(http.Dir("./elf/"))))

	config := goElFinder.Config{}
	config["l0"] = goElFinder.Volume {
		Root: "./files/1",
		AllowDirs: []string{"/Allow"},
		DenyDirs:  []string{"/Deny"},
		DefaultRight: false,
	}
	config["l1"] = goElFinder.Volume {
		Root: "./files/2",
		DefaultRight: true,
	}
	mux.Handle("/connector", goElFinder.NetHttp(config))

	fmt.Println("Listen on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}