package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/go-github/github"
	"gopkg.in/rjz/githubhook.v0"
	"log"
	"net/http"
	"os"
)

func EventHandler(w http.ResponseWriter, r *http.Request) {
	secret := []byte(os.Getenv("GITHUB_PRESHARED_SECRET"))
	hook, err := githubhook.Parse(secret, r)
	if err != nil {
		log.Println("Event Handler")
	}
	evt := github.PushEvent{}
	if err := json.Unmarshal(hook.Payload, &evt); err != nil {
		error := fmt.Sprintf("Encountered an error parsing JSON: %s", err)
		errorHandler(w, r, error)
	}
	go EventDetails(evt)
	JSONResponseHandler(w, map[string]string{
		"message": "Processing Webhook event",
		"target":  *evt.Ref,
		"before":  *evt.Before,
		"after":   *evt.After,
	})
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(Response{Status: "error", Message: fmt.Sprintf("Route %s not found with method %s, "+
		"please check request and try again", r.URL.Path, r.Method)})
}

func errorHandler(w http.ResponseWriter, r *http.Request, error string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	_ = r
	json.NewEncoder(w).Encode(Response{Status: "error", Message: fmt.Sprint(error)})
}

func JSONResponseHandler(w http.ResponseWriter, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
