package http_server

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/phuslu/log"
	"github.com/spf13/afero"
	"github.com/victormf2/gosyringe"

	"github.com/prigas-dev/backoffice-ai/ComponentGenerator"
	"github.com/prigas-dev/backoffice-ai/feature_generator"
	"github.com/prigas-dev/backoffice-ai/feature_generator/instruction_files"
	"github.com/prigas-dev/backoffice-ai/operations"
)

//go:embed index.html
var templates embed.FS

func Start(ctx context.Context, db *sql.DB, operationsFs afero.Fs) {

	container := gosyringe.NewContainer()

	RegisterServices(container, db, operationsFs)

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
		Parameters map[string]any `json:"parameters"`
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

		executor, err := gosyringe.Resolve[operations.IOperationExecutor](container)
		if err != nil {
			log.Error().Msgf("error instantiating operation executor: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ExecuteOperationErrorResponseBody{
				Success: false,
				Message: fmt.Sprintf("error instantiating operation executor: %v", err),
			})
			return
		}

		result, err := executor.Execute(operationName, requestBody.Parameters)
		if err != nil {
			log.Error().Msgf("error on operation execution: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ExecuteOperationErrorResponseBody{
				Success: false,
				Message: fmt.Sprintf("error on operation execution: %v", err),
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

		featureGenerator, err := gosyringe.Resolve[feature_generator.IFeatureGenerator](container)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = featureGenerator.GenerateFeature(ctx, prompt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(struct{}{})
	})

	// Start the web server
	log.Info().Msg("Server starting on http://localhost:8080")
	log.Fatal().Err(http.ListenAndServe(":8080", nil)).Msg("exit")
}

func RegisterServices(c *gosyringe.Container, db *sql.DB, operationsFs afero.Fs) {

	gosyringe.RegisterValue[*sql.DB](c, db)

	gosyringe.RegisterValue[afero.Fs](c, operationsFs)
	gosyringe.RegisterSingleton[operations.IOperationStore](c, operations.NewFsOperationStore)

	gosyringe.RegisterSingleton[feature_generator.IDatabaseSchemaGenerator](c, feature_generator.NewSqliteSchemaGenerator)
	templateData := &feature_generator.InstructionsTemplateData{
		SystemName:        "Task Manager",
		SystemDescription: "Task Manager is a system for managing a team's tasks.",
		DatabaseHints: `
These are all possible values for a task status:
- done
- todo
- in_progress
`,
		DatabaseEngine:    "sqlite3",
		FeatureJSONSchema: instruction_files.FeatureJSONSchema.Content,
		ErrorJSONSchema:   instruction_files.ErrorJSONSchema.Content,
		ValidFeatureJSON:  instruction_files.ExampleFeatureJSON.Content,
		ValidFeatureFiles: instruction_files.ExampleFeatureFiles,
	}
	gosyringe.RegisterValue[*feature_generator.InstructionsTemplateData](c, templateData)
	gosyringe.RegisterSingleton[feature_generator.IAIGenerator](c, feature_generator.NewAIGenerator)

	gosyringe.RegisterSingleton[feature_generator.IFeatureGenerator](c, feature_generator.NewReactFeatureGenerator)

	gosyringe.RegisterSingleton[operations.IOperationExecutor](c, operations.NewOperationExecutor)
}
