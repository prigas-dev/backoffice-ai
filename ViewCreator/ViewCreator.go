package ViewCreator

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/prigas-dev/backoffice-ai/AiAssistant"
	"github.com/prigas-dev/backoffice-ai/ComponentGenerator"
	"github.com/prigas-dev/backoffice-ai/examples"
	"github.com/prigas-dev/backoffice-ai/operations"
	"github.com/prigas-dev/backoffice-ai/schemas"

	_ "embed"
)

type ReactComponentProps map[string]any

func CreateView(ctx context.Context, db *sql.DB, operationStore operations.IOperationStore, prompt string) error {
	templateData := AiAssistant.InstructionsTemplateData{
		SystemName:        "Task Manager",
		SystemDescription: "Task Manager is a system for managing a team's tasks.",
		DatabaseHints: `
These are all possible values for a task status:
- done
- todo
- in_progress
`,
		DatabaseEngine:    "sqlite3",
		FeatureJSONSchema: schemas.FeatureSchema,
		ErrorJSONSchema:   schemas.ErrorSchema,
		ValidFeatureJSON:  examples.FeatureJSON,
		ValidFeatureFiles: []AiAssistant.FeatureFile{
			{
				MarkdownLanguageIdentifier: "typescriptreact",
				Filename:                   examples.ComponentFilename,
				Content:                    examples.ComponentFileContent,
			},
			{
				MarkdownLanguageIdentifier: "javascript",
				Filename:                   examples.GetUsernameFilename,
				Content:                    examples.GetUsernameFileContent,
			},
			{
				MarkdownLanguageIdentifier: "javascript",
				Filename:                   examples.UpdateUsernameFilename,
				Content:                    examples.UpdateUsernameFileContent,
			},
		},
	}
	p, err := AiAssistant.Assist(ctx, db, prompt, templateData)
	if err != nil {
		return fmt.Errorf("failed to create page component view: %w", err)
	}

	err = SaveFeatureToJsonFile(p)
	if err != nil {
		return fmt.Errorf("failed to save view json file: %w", err)
	}

	err = ComponentGenerator.GenerateComponentTSX(p.ReactComponent.TsxCode, "./http_server/public")
	if err != nil {
		return fmt.Errorf("failed to generate tsx component: %w", err)
	}

	for _, operation := range p.ServerOperations {
		err := operationStore.AddOperation(&operation)
		if err != nil {
			return fmt.Errorf("failed to store operation %s: %w", operation.Name, err)
		}
	}

	return nil

}

func SaveFeatureToJsonFile(p *AiAssistant.FeatureStructure) error {
	outFile, err := os.Create(fmt.Sprintf("./AiGeneratedViews/%s.json", uuid.NewString()))
	if err != nil {
		return fmt.Errorf("failed to create view json file: %w", err)
	}
	defer outFile.Close()

	formattedJson, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format json: %w", err)
	}

	_, err = outFile.Write(formattedJson)
	if err != nil {
		return fmt.Errorf("failed to write json file: %w", err)
	}

	return nil
}
