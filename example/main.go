package main

import (
	"fmt"
	"github.com/kimperator/goElFinder"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./elf/"))))
	mux.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir("./files/"))))

	volumes := goElFinder.Volumes{}
	volumes["l0"] = goElFinder.Volume{
		Default:      true,
		Root:         "./files/1",
		Url:          "http://localhost:8080/files/1",
		AllowDirs:    []string{"/Allow"},
		DenyDirs:     []string{"/Deny"},
		DefaultRight: false,
	}
	volumes["l1"] = goElFinder.Volume{
		Root:         "./files/2",
		Url:          "http://localhost:8080/files/2",
		DefaultRight: true,
	}
	mux.Handle("/connector", volumes.NetHttp())

	fmt.Println("Listen on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
