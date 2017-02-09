package main

import (
	"log"
	"net/http"
	"github.com/supme/elFinder"
	"fmt"
)

func main() {
	http.Handle("/elf/", http.StripPrefix("/elf/", http.FileServer(http.Dir("./elf/"))))
	http.HandleFunc("/connector", elFinder.NetHttp)
	fmt.Println("Listen on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

/*
func checkRights(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {


		fn(w, r)
	}
}
*/