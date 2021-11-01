package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
)

//go:embed spline-chart
var splitChart embed.FS

func main() {
	serverRoot, err := fs.Sub(splitChart, "spline-chart")
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/", http.FileServer(http.FS(serverRoot)))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
