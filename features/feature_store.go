package features

import (
	"encoding/json"
	"fmt"
	"path"

	"github.com/prigas-dev/backoffice-ai/operations"
	"github.com/prigas-dev/backoffice-ai/utils"
	"github.com/spf13/afero"
)

type IFeatureStore interface {
	GetAllFeatures() ([]*FeatureManifest, error)
	GetFeature(name string) (*Feature, error)
	AddFeature(feature *Feature) error
}

type FeatureManifest struct {
	Name        string   `json:"name"`
	Label       string   `json:"label"`
	Description string   `json:"description"`
	Operations  []string `json:"operations"`
}

func NewFsFeatureStore(fs FeaturesFs, operationStore operations.IOperationStore, componentStore IComponentStore) IFeatureStore {
	return &FsFeatureStore{
		fs:             fs,
		operationStore: operationStore,
		componentStore: componentStore,
	}
}

type FeaturesFs afero.Fs

type FsFeatureStore struct {
	fs             afero.Fs
	operationStore operations.IOperationStore
	componentStore IComponentStore
}

func (s *FsFeatureStore) GetAllFeatures() ([]*FeatureManifest, error) {
	files, err := afero.ReadDir(s.fs, ".")
	if err != nil {
		return nil, fmt.Errorf("failed to read features directory: %w", err)
	}

	features := []*FeatureManifest{}

	for _, featureDir := range files {
		featureName := featureDir.Name()
		feature, err := s.getFeatureManifest(featureName)
		if err != nil {
			return nil, fmt.Errorf("failed to get feature %s: %w", featureName, err)
		}
		features = append(features, feature)
	}

	return features, nil
}

func (s *FsFeatureStore) getFeatureManifest(name string) (*FeatureManifest, error) {
	featureManifestFile, err := s.fs.Open(path.Join(name, "feature_manifest.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to open feature_manifest.json of feature %s: %w", name, err)
	}
	defer featureManifestFile.Close()

	var featureManifest FeatureManifest
	err = json.NewDecoder(featureManifestFile).Decode(&featureManifest)
	if err != nil {
		return nil, fmt.Errorf("failed to parse feature_manifest.json of feature %s: %w", name, err)
	}

	return &featureManifest, nil
}

func (s *FsFeatureStore) GetFeature(name string) (*Feature, error) {
	featureManifest, err := s.getFeatureManifest(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get manifest of feature %s: %w", name, err)
	}

	operations := []*operations.Operation{}
	for _, operationName := range featureManifest.Operations {

		operation, err := s.operationStore.GetOperation(operationName)
		if err != nil {
			return nil, fmt.Errorf("failed to get operation %s: %w", operationName, err)
		}

		operations = append(operations, operation)
	}

	reactComponentContent, err := s.componentStore.GetComponent(featureManifest.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to read component content of feature %s: %w", featureManifest.Name, err)
	}

	feature := &Feature{
		Name:        featureManifest.Name,
		Label:       featureManifest.Label,
		Description: featureManifest.Description,
		ReactComponent: &ReactComponent{
			TsxCode: string(reactComponentContent),
		},
		ServerOperations: operations,
	}

	return feature, nil
}

func (s *FsFeatureStore) AddFeature(feature *Feature) error {

	err := s.fs.MkdirAll(feature.Name, 0755)
	if err != nil {
		return fmt.Errorf("failed to create feature %s directory: %w", feature.Name, err)
	}

	featureManifestFile, err := s.fs.Create(path.Join(feature.Name, "feature_manifest.json"))
	if err != nil {
		return fmt.Errorf("failed to create feature %s manifest file: %w", feature.Name, err)
	}
	defer featureManifestFile.Close()

	encoder := json.NewEncoder(featureManifestFile)
	encoder.SetIndent("", "  ")

	featureManifest := FeatureManifest{
		Name:        feature.Name,
		Label:       feature.Label,
		Description: feature.Description,
		Operations:  utils.Map(feature.ServerOperations, func(operation *operations.Operation) string { return operation.Name }),
	}
	err = encoder.Encode(featureManifest)
	if err != nil {
		return fmt.Errorf("failed to write feature %s manifest file: %w", feature.Name, err)
	}

	err = s.componentStore.AddComponent(feature.Name, feature.ReactComponent.TsxCode)
	if err != nil {
		return fmt.Errorf("failed to store component for feature: %s: %w", feature.Name, err)
	}

	for _, operation := range feature.ServerOperations {
		err := s.operationStore.AddOperation(operation)
		if err != nil {
			return fmt.Errorf("failed to add operation %s for feature %s: %w", operation.Name, feature.Name, err)
		}
	}

	return nil
}
