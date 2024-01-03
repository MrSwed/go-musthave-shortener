package handler

import (
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
)

func (h *Handler) MakeShort() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body == http.NoBody {
			w.WriteHeader(http.StatusBadRequest)
			log.Printf("No body ")
			return
		}
		defer func() { _ = r.Body.Close() }()
		url, err := io.ReadAll(r.Body)
		if err != nil && !errors.Is(err, io.EOF) {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("Error get body %s", err)
			return
		}
		html, err := h.s.NewShort(string(url))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("Error reate new short %s", err)
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		if _, err := w.Write([]byte(html)); err != nil {
			log.Printf("Error: %s", err)
		}
	}
}

func (h *Handler) GetShort() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		params := strings.Split(
			strings.Trim(r.URL.Path, "/"), "/")

		if len(params) > 1 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if newURL, err := h.s.GetFromShort(params[0]); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		} else {
			http.Redirect(w, r, newURL, http.StatusTemporaryRedirect)
			return
		}
	}
}
