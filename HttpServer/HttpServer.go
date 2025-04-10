package HttpServer

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/prigas-dev/backoffice-ai/AiAssistant"
	"github.com/prigas-dev/backoffice-ai/ComponentGenerator"
	"github.com/prigas-dev/backoffice-ai/ViewCreator"
)

//go:embed index.html
var templates embed.FS

func Start(ctx context.Context, db *sql.DB) {
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

		err := ComponentGenerator.GenerateComponentSample("./HttpServer/public")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/test-anthropic", func(w http.ResponseWriter, r *http.Request) {

		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		prompt := r.Form.Get("prompt")
		if len(prompt) == 0 {
			http.Error(w, "prompt is required", http.StatusBadRequest)
			return
		}

		result, err := AiAssistant.Assist(ctx, db, prompt, "")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	http.HandleFunc("/test-load-view", func(w http.ResponseWriter, r *http.Request) {
		file, err := os.Open("AiGeneratedViews/todo_tasks_list.json")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		p := &AiAssistant.PageComponentView{}
		err = json.NewDecoder(file).Decode(p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = ComponentGenerator.GenerateComponentTSX(p.Component.Code, "./HttpServer/public")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		props := map[string]any{}

		for _, query := range p.Queries {
			rows, err := ViewCreator.RunQuery(db, query)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if query.Mode == AiAssistant.SingleRow {
				if len(rows) == 1 {
					props[query.MapToProp] = rows[0]
				}
			} else if query.Mode == AiAssistant.MultipleRows {
				props[query.MapToProp] = rows
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(props)
	})

	http.HandleFunc("/new-view", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		prompt := r.Form.Get("prompt")
		if len(prompt) == 0 {
			http.Error(w, "prompt is required", http.StatusBadRequest)
			return
		}

		props, err := ViewCreator.CreateView(ctx, db, prompt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(props)
	})

	// Start the web server
	log.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
