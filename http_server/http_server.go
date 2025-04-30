package http_server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/phuslu/log"
	"github.com/spf13/afero"
	"github.com/victormf2/gosyringe"

	"github.com/prigas-dev/backoffice-ai/features"
	"github.com/prigas-dev/backoffice-ai/frontend"
	"github.com/prigas-dev/backoffice-ai/http_server/handlers"
	"github.com/prigas-dev/backoffice-ai/operations"
)

func Start(ctx context.Context) {

	container := gosyringe.NewContainer()

	RegisterServices(container)
	db, err := gosyringe.Resolve[*sql.DB](container)
	if err != nil {
		log.Fatal().Err(fmt.Errorf("failed to instance db: %w", err))
	}
	defer db.Close()

	fs := http.FileServer(http.Dir("http_server/public"))
	http.Handle("/public/", http.StripPrefix("/public/", fs))

	handlers.Index()
	handlers.OperationsExecute(container)
	handlers.CreateFeature(container)
	handlers.GetAllFeatures(container)
	handlers.TestBuilder(container)

	// Start the web server
	log.Info().Msg("Server starting on http://localhost:8080")
	log.Fatal().Err(http.ListenAndServe(":8080", nil)).Msg("exit")
}

func RegisterServices(c *gosyringe.Container) {
	dbPath := "kanban.db"
	// Open database connection
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal().Err(fmt.Errorf("failed to open database: %w", err))
	}

	operationsFolder := "fstore/operations"
	err = os.MkdirAll(operationsFolder, 0755)
	if err != nil {
		log.Fatal().Err(fmt.Errorf("failed to create operations folder: %w", err))
	}
	operationsFs := afero.NewBasePathFs(afero.NewOsFs(), operationsFolder)

	frontendFolder := "frontend"
	componentsFs := afero.NewBasePathFs(afero.NewOsFs(), frontendFolder)

	featuresFolder := "fstore/features"
	err = os.MkdirAll(featuresFolder, 0755)
	if err != nil {
		log.Fatal().Err(fmt.Errorf("failed to create features folder: %w", err))
	}
	featuresFs := afero.NewBasePathFs(afero.NewOsFs(), featuresFolder)

	frontendBuilderConfig := &frontend.BuilderConfig{
		Entrypoint:        "frontend/src/main.tsx",
		DestinationFolder: "http_server/public",
	}

	gosyringe.RegisterValue[*sql.DB](c, db)

	gosyringe.RegisterValue[operations.OperationsFs](c, operationsFs)
	gosyringe.RegisterSingleton[operations.IOperationStore](c, operations.NewFsOperationStore)
	gosyringe.RegisterValue[features.ComponentsFs](c, componentsFs)
	gosyringe.RegisterSingleton[features.IComponentStore](c, features.NewFsComponentStore)
	gosyringe.RegisterValue[features.FeaturesFs](c, featuresFs)
	gosyringe.RegisterSingleton[features.IFeatureStore](c, features.NewFsFeatureStore)

	gosyringe.RegisterSingleton[frontend.IBuilder](c, frontend.NewBuilder)
	gosyringe.RegisterValue[*frontend.BuilderConfig](c, frontendBuilderConfig)

	gosyringe.RegisterSingleton[features.IDatabaseSchemaGenerator](c, features.NewSqliteSchemaGenerator)
	templateData := &features.InstructionsTemplateData{
		SystemName:        "Task Manager",
		SystemDescription: "Task Manager is a system for managing a team's tasks.",
		DatabaseHints: `
These are all possible values for a task status:
- done
- todo
- in_progress
`,
		DatabaseEngine: "sqlite3",
	}
	gosyringe.RegisterValue[*features.InstructionsTemplateData](c, templateData)
	gosyringe.RegisterSingleton[features.IAIGenerator](c, features.NewAIGenerator)
	// gosyringe.RegisterSingleton[features.IAIGenerator](c, NewTestAIGenerator)

	gosyringe.RegisterSingleton[features.IFeatureGenerator](c, features.NewReactFeatureGenerator)

	gosyringe.RegisterSingleton[operations.IOperationExecutor](c, operations.NewOperationExecutor)
}

type TestAIGenerator struct{}

func NewTestAIGenerator() features.IAIGenerator {
	return &TestAIGenerator{}
}

func (g *TestAIGenerator) Generate(ctx context.Context, prompt string) (*features.Feature, error) {
	featureJson, err := os.ReadFile("AiGeneratedViews/b701b9f0-21e7-44fe-ba2a-f488521ecffa.json")
	if err != nil {
		return nil, err
	}

	feature := features.Feature{}
	err = json.Unmarshal(featureJson, &feature)
	if err != nil {
		return nil, err
	}

	return &feature, nil
}
