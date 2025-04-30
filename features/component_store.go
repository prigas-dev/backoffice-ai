package features

import (
	_ "embed"
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"

	"github.com/prigas-dev/backoffice-ai/utils"
	"github.com/spf13/afero"
)

type IComponentStore interface {
	AddComponent(name string, tsxCode string) error
	GetComponent(name string) ([]byte, error)
}

func NewFsComponentStore(fs ComponentsFs) IComponentStore {
	return &FsComponentStore{
		fs: fs,
	}
}

type ComponentsFs afero.Fs

type FsComponentStore struct {
	fs ComponentsFs
}

var componentsFolder = path.Join("src", "components")

func (s *FsComponentStore) AddComponent(name string, tsxCode string) error {
	err := s.fs.MkdirAll(componentsFolder, 0755)
	if err != nil {
		return fmt.Errorf("failed to create components folder: %w", err)
	}

	componentFile, err := s.fs.Open(path.Join(componentsFolder, fmt.Sprintf("%s.tsx", name)))
	if err != nil {
		return fmt.Errorf("failed to open component file: %w", err)
	}
	defer componentFile.Close()

	_, err = io.WriteString(componentFile, tsxCode)
	if err != nil {
		return fmt.Errorf("failed to write component file: %w", err)
	}

	err = s.regenerateRoot()
	if err != nil {
		return fmt.Errorf("failed to regenerate root: %w", err)
	}

	return nil
}

//go:embed main.tsx.tmpl
var mainTsxTemplate string

func (s *FsComponentStore) regenerateRoot() error {
	componentNames := []string{}
	err := afero.Walk(s.fs, componentsFolder, func(path string, info fs.FileInfo, err error) error {
		if err == nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(info.Name(), ".tsx") {
			return nil
		}

		componentName := strings.TrimSuffix(info.Name(), ".tsx")

		componentNames = append(componentNames, componentName)
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to retrieve components: %w", err)
	}

	mainTsx, err := utils.DoTemplate("main.tsx", mainTsxTemplate, map[string]any{
		"ComponentNames": componentNames,
	})

	if err != nil {
		return fmt.Errorf("failed to generate main.tsx: %w", err)
	}

	err = afero.WriteFile(s.fs, "main.tsx", []byte(mainTsx), 0755)
	if err != nil {
		return fmt.Errorf("failed to write main.tsx file: %w", err)
	}

	return nil
}

func (s *FsComponentStore) GetComponent(name string) ([]byte, error) {
	content, err := afero.ReadFile(s.fs, path.Join(componentsFolder, fmt.Sprintf("%s.tsx", name)))
	if err != nil {
		return nil, fmt.Errorf("failed to read component %s.tsx: %w", name, err)
	}

	return content, nil
}
