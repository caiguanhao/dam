package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	address := flag.String("address", "127.0.0.1:15161", "listen address")
	port := flag.String("p", "", "serial port")
	flag.Parse()

	openPort(*port)

	log.Println("listening", "http://"+*address)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.Path == "/" {
			w.Write(indexHtml)
			return
		}
		jsonRPCHandler(w, r)
	})
	log.Fatal(http.ListenAndServe(*address, nil))
}
