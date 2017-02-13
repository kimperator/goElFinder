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
	//mux.Handle("/connector", elFinder.NetHttp())
	mux.HandleFunc("/connector", elFinder.NetHttp)
	fmt.Println("Listen on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}