package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/prigas-dev/backoffice-ai/features"
	"github.com/victormf2/gosyringe"
)

func CreateFeature(container *gosyringe.Container) {

	ctx := context.Background()

	http.HandleFunc("/create-feature", func(w http.ResponseWriter, r *http.Request) {
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

		featureGenerator, err := gosyringe.Resolve[features.IFeatureGenerator](container)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		feature, err := featureGenerator.GenerateFeature(ctx, prompt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(feature)
	})
}
