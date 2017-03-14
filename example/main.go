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
	mux.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir("./files/"))))

	config := goElFinder.Volumes{}
	config["l0"] = goElFinder.Local {
		Default: true,
		Root: "/home/aagafonov/Golang/myprojects/goElFinder/example/files/1",
		Url: "http://ly.dmbasis.ru:8080/files/1",
		AllowDirs: []string{"/Allow"},
		DenyDirs:  []string{"/Deny"},
		DefaultRight: false,
	}
	config["l1"] = goElFinder.Local {
		Root: "/home/aagafonov/Golang/myprojects/goElFinder/example/files/2",
		Url: "http://ly.dmbasis.ru:8080/files/2",
		DefaultRight: true,
	}
	mux.Handle("/connector", goElFinder.NetHttp(config))

	fmt.Println("Listen on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}