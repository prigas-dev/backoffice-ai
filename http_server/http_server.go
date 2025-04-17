package http_server

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"

	"github.com/phuslu/log"
	"github.com/spf13/afero"

	"github.com/prigas-dev/backoffice-ai/AiAssistant"
	"github.com/prigas-dev/backoffice-ai/ComponentGenerator"
	"github.com/prigas-dev/backoffice-ai/ViewCreator"
	"github.com/prigas-dev/backoffice-ai/operations"
)

//go:embed index.html
var templates embed.FS

func Start(ctx context.Context, db *sql.DB) {
	// Parse the HTML template
	tmpl, err := template.ParseFS(templates, "*.html")
	if err != nil {
		log.Fatal().Msgf("Error parsing template: %v", err)
	}

	fs := http.FileServer(http.Dir("http_server/public"))
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

	type ExecuteOperationRequestBody struct {
		Arguments map[string]any `json:"arguments"`
	}

	type ExecuteOperationSuccessResponseBody struct {
		Success bool `json:"success"`
		Result  any  `json:"result"`
	}

	type ExecuteOperationErrorResponseBody struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}

	getOperationName := func(r *http.Request) (string, error) {

		pathSegments := strings.Split(
			strings.TrimLeft(r.URL.Path, "/"),
			"/",
		)

		if len(pathSegments) == 0 {
			return "", fmt.Errorf("request path is empty")
		}

		// /operations/execute/{operationName}
		if len(pathSegments) != 3 {
			return "", fmt.Errorf("invalid request path")
		}

		operationName := pathSegments[len(pathSegments)-1]

		return operationName, nil
	}

	err = os.MkdirAll("tmp/operations", 0755)
	if err != nil {
		panic(err)
	}
	files := afero.NewBasePathFs(afero.NewOsFs(), "tmp/operations")
	store := operations.NewFsOperationStore(files)
	executor := operations.NewOperationExecutor(store)

	http.HandleFunc("/operations/execute/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			log.Warn().Msgf("request with invalid method: %v", r.Method)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ExecuteOperationErrorResponseBody{
				Success: false,
				Message: "only POST method is allowed",
			})
			return
		}

		operationName, err := getOperationName(r)
		if err != nil {
			log.Warn().Msgf("request with invalid path %v: %v", r.URL.Path, err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ExecuteOperationErrorResponseBody{
				Success: false,
				Message: err.Error(),
			})
			return
		}

		requestBody := ExecuteOperationRequestBody{}
		err = json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			log.Warn().Msgf("failed to parse request parameters: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ExecuteOperationErrorResponseBody{
				Success: false,
				Message: "failed to parse request parameters",
			})
			return
		}

		result, err := executor.Execute(operationName, requestBody.Arguments)
		if err != nil {
			log.Error().Msgf("error on operation execution: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ExecuteOperationErrorResponseBody{
				Success: false,
				Message: "error on operation execution",
			})
			return
		}

		json.NewEncoder(w).Encode(ExecuteOperationSuccessResponseBody{
			Success: true,
			Result:  result,
		})
	})

	http.HandleFunc("/generate-component", func(w http.ResponseWriter, r *http.Request) {

		err := ComponentGenerator.GenerateComponentSample("./http_server/public")
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

		err = ComponentGenerator.GenerateComponentTSX(p.Component.Code, "./http_server/public")
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
	log.Info().Msg("Server starting on http://localhost:8080")
	log.Fatal().Err(http.ListenAndServe(":8080", nil)).Msg("exit")
}
