package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
)

func (app *Application) putHandler(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		writeBadRequest(w, errors.New("path value `key` is not specified"))
		return
	}
	v, err := io.ReadAll(r.Body)
	if err != nil {
		writeServerError(w, err)
		return
	}
	app.Store.Put(key, v)
	resp := struct {
		Key string `json:"key"`
	}{
		Key: key,
	}
	writeJSON(w, resp, http.StatusCreated)
}

func (app *Application) getHandler(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		writeBadRequest(w, errors.New("path value `key` is not specified"))
		return
	}
	v, ok := app.Store.Get(key)
	resp := struct {
		Exists bool   `json:"exists"`
		Key    string `json:"key"`
		Value  []byte `json:"value"`
	}{Key: key, Exists: ok, Value: v}
	writeJSON(w, resp, http.StatusOK)
}

func (app *Application) deleteHandler(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		writeBadRequest(w, errors.New("path value `key` is not specified"))
		return
	}
	app.Store.Delete(key)
	resp := struct {
		Key string `json:"key"`
	}{Key: key}
	writeJSON(w, resp, http.StatusOK)
}

func writeJSON(w http.ResponseWriter, v any, code int) error {
	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	w.WriteHeader(code)
	w.Write(data)
	return nil
}

func writeError(w http.ResponseWriter, err error, code int) {
	v := map[string]any{"error": err.Error()}
	writeJSON(w, v, code)
}

func writeBadRequest(w http.ResponseWriter, err error) {
	writeError(w, err, http.StatusBadRequest)
}

func writeServerError(w http.ResponseWriter, err error) {
	writeError(w, errors.New("internal server error"), http.StatusInternalServerError)
	log.Println(err)
}
