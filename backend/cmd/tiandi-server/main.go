package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"agent-dou-dizhu/internal/tiandi/demo"
)

func main() {
	service, err := demo.NewService()
	if err != nil {
		log.Fatal(err)
	}

	mux := newMux(service)

	addr := ":8080"
	log.Printf("tiandi server listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func newMux(service *demo.Service) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/game/state", withService(service, handleState))
	mux.HandleFunc("/api/game/reset", withService(service, handleReset))
	mux.HandleFunc("/api/game/action", withService(service, handleAction))
	mux.HandleFunc("/api/game/rules", withService(service, handleRules))
	return mux
}

type handlerFunc func(service *demo.Service, w http.ResponseWriter, r *http.Request) error

func withService(service *demo.Service, handler handlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setCommonHeaders(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if err := handler(service, w, r); err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, errMethodNotAllowed) {
				status = http.StatusMethodNotAllowed
			} else if errors.Is(err, errBadRequest) {
				status = http.StatusBadRequest
			}
			http.Error(w, err.Error(), status)
		}
	}
}

var (
	errMethodNotAllowed = errors.New("method not allowed")
	errBadRequest       = errors.New("bad request")
)

func handleState(service *demo.Service, w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return errMethodNotAllowed
	}

	state, err := service.State()
	if err != nil {
		return err
	}
	return writeJSON(w, state)
}

func handleReset(service *demo.Service, w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return errMethodNotAllowed
	}

	state, err := service.Reset()
	if err != nil {
		return err
	}
	return writeJSON(w, state)
}

func handleAction(service *demo.Service, w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return errMethodNotAllowed
	}

	var req demo.ActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return errors.Join(errBadRequest, err)
	}

	state, err := service.Apply(req)
	if err != nil {
		return errors.Join(errBadRequest, err)
	}
	return writeJSON(w, state)
}

func handleRules(service *demo.Service, w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return errMethodNotAllowed
	}

	return writeJSON(w, service.Rules())
}

func setCommonHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
}

func writeJSON(w http.ResponseWriter, payload any) error {
	return json.NewEncoder(w).Encode(payload)
}
