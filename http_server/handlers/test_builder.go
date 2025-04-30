package handlers

import (
	"fmt"
	"net/http"

	"github.com/prigas-dev/backoffice-ai/frontend"
	"github.com/victormf2/gosyringe"
)

func TestBuilder(container *gosyringe.Container) {
	http.HandleFunc("/build", func(w http.ResponseWriter, r *http.Request) {
		builder, err := gosyringe.Resolve[frontend.IBuilder](container)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to instance IBuilder: %v", err), http.StatusInternalServerError)
		}

		err = builder.BuildFrontend()
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to build frontend: %v", err), http.StatusInternalServerError)
		}

		http.Redirect(w, r, "/", http.StatusFound)
	})
}
