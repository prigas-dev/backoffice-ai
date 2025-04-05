package HttpServer

import (
	"embed"
	"html/template"
	"log"
	"net/http"

	"github.com/prigas-dev/backoffice-ai/ComponentGenerator"
)

//go:embed index.html
var templates embed.FS

func Start() {
	// Parse the HTML template
	tmpl, err := template.ParseFS(templates, "*.html")
	if err != nil {
		log.Fatalf("Error parsing template: %v", err)
	}

	fs := http.FileServer(http.Dir("HttpServer/public"))
	http.Handle("/public/", http.StripPrefix("/public/", fs))

	// Handle the home page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		type PageData struct{}
		// Sample data for our template
		data := PageData{}

		err := tmpl.ExecuteTemplate(w, "index.html", data)
		if err != nil {
			log.Printf("Error executing template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	})

	http.HandleFunc("/generate-component", func(w http.ResponseWriter, r *http.Request) {

		err := ComponentGenerator.GenerateComponent("./HttpServer/public")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusOK)
	})

	// Start the web server
	log.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
