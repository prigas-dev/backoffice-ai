package ViewCreator

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"github.com/prigas-dev/backoffice-ai/AiAssistant"
	"github.com/prigas-dev/backoffice-ai/ComponentGenerator"
)

type ReactComponentProps map[string]any

func CreateView(ctx context.Context, db *sql.DB, prompt string) (*ReactComponentProps, error) {
	databaseHints := `
These are all possible values for a task status:
- done
- todo
- in_progress
`
	templateData := AiAssistant.InstructionsTemplateData{
		DatabaseHints: databaseHints,
	}
	p, err := AiAssistant.Assist(ctx, db, prompt, templateData)
	if err != nil {
		return nil, fmt.Errorf("failed to create page component view: %w", err)
	}

	err = SaveViewToJsonFile(p)
	if err != nil {
		return nil, fmt.Errorf("failed to save view json file: %w", err)
	}

	err = ComponentGenerator.GenerateComponentTSX(p.Component.Code, "./http_server/public")
	if err != nil {
		return nil, fmt.Errorf("failed to generate tsx component: %w", err)
	}

	props := ReactComponentProps{}

	for _, query := range p.Queries {
		rows, err := RunQuery(db, query)
		if err != nil {
			return nil, fmt.Errorf("failed to run query %s: %w", query.SQL, err)
		}

		if query.Mode == AiAssistant.SingleRow {
			if len(rows) == 1 {
				props[query.MapToProp] = rows[0]
			}
		} else if query.Mode == AiAssistant.MultipleRows {
			props[query.MapToProp] = rows
		}
	}

	return &props, nil
}

func SaveViewToJsonFile(p *AiAssistant.PageComponentView) error {
	outFile, err := os.Create(fmt.Sprintf("./AiGeneratedViews/%s.json", p.Component.ID))
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

func RunQuery(db *sql.DB, query AiAssistant.Query) ([]ReactComponentProps, error) {
	rows, err := db.Query(query.SQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	scannedRows := []ReactComponentProps{}
	for rows.Next() {
		values := make([]any, len(columns))
		scanArgs := make([]any, len(columns))
		for i := range values {
			scanArgs[i] = &values[i]
		}

		err := rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}

		mapValues := ReactComponentProps{}
		for i, val := range values {
			// Ensuring byte slices are converted to string
			if b, ok := val.([]byte); ok {
				values[i] = string(b)
			}

			mapValues[columns[i]] = values[i]
		}

		scannedRows = append(scannedRows, mapValues)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return scannedRows, nil
}
