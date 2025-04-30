package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/phuslu/log"
	"github.com/prigas-dev/backoffice-ai/features"
	"github.com/victormf2/gosyringe"
)

func GetAllFeatures(container *gosyringe.Container) {

	http.HandleFunc("/get-all-features", func(w http.ResponseWriter, r *http.Request) {
		featureStore, err := gosyringe.Resolve[features.IFeatureStore](container)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to instance feature store: %v", err), http.StatusInternalServerError)
			return
		}

		features, err := featureStore.GetAllFeatures()
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to get all features: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(map[string]any{
			"features": features,
		})
		if err != nil {
			log.Error().Err(fmt.Errorf("failed to write features JSON: %w", err))
		}
	})
}
