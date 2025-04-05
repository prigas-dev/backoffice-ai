package ComponentGenerator

import (
	"embed"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"path"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
)

//go:embed sample/*.tsx
var sampleComponents embed.FS

var ErrNoTsxFiles = errors.New("no tsx file found")

func GenerateComponent(publicFolder string) error {

	files, err := sampleComponents.ReadDir("sample")
	if err != nil {
		return fmt.Errorf("failed to read sample directory: %w", err)
	}

	tsxFiles := []string{}
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		info, err := file.Info()
		if err != nil {
			log.Printf("Failed to read file: %v\n", err)
			continue
		}

		isTsxFile := strings.HasSuffix(info.Name(), ".tsx")
		if isTsxFile {
			tsxFiles = append(tsxFiles, fmt.Sprintf("sample/%s", info.Name()))
		}
	}

	if len(tsxFiles) == 0 {
		return ErrNoTsxFiles
	}

	fmt.Printf("tsx files: %+v\n", tsxFiles)

	chosenIndex := rand.Intn(len(tsxFiles))

	file, err := sampleComponents.Open(tsxFiles[chosenIndex])
	if err != nil {
		return fmt.Errorf("failed to open tsx file: %w", err)
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	jsx := string(bytes)

	result := api.Transform(jsx, api.TransformOptions{
		Loader: api.LoaderTSX,
		JSX:    api.JSXAutomatic,
	})

	log.Printf("%d errors and %d warnings\n",
		len(result.Errors), len(result.Warnings))

	if len(result.Errors) > 0 {
		return fmt.Errorf("JSX transform errors: %+v", result.Errors)
	}

	err = os.MkdirAll(publicFolder, 0660)
	if err != nil {
		return fmt.Errorf("failed to create public folder: %w", err)
	}

	output, err := os.Create(path.Join(publicFolder, "component.mjs"))
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer output.Close()

	output.Write(result.Code)

	return nil
}
