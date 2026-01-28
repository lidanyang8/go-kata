package main

import (
	"html/template"
	"log"
	"net/http"
)

func main() {
	tfs, sfs, err := Assets()
	if err != nil {
		log.Fatal(err)
	}

	tmpl, err := template.ParseFS(tfs, "*.html")
	if err != nil {
		log.Fatal(err)
	}

	fileServer := http.FileServer(http.FS(sfs))
	http.Handle("/static/", http.StripPrefix("/static/", fileServer))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "index.html", nil)
	})

	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
