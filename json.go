package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func respondWithError(w http.ResponseWriter, code int, msg string, err error) {
	if err != nil {
		log.Println(err)
	}
	if code > 499 {
		log.Printf("Responding with 5XX error: %s", msg)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}

func getMimeExtension(mediaType string) (string, error) {
	mimeExtensions := map[string]string{
		"image/png":        ".png",
		"image/jpeg":       ".jpg",
		"video/mp4":       ".mp4",
	}

	if ext, ok := mimeExtensions[mediaType]; ok {
		return ext, nil
	}
	return "", fmt.Errorf("unsupported media type: %s", mediaType)
}
