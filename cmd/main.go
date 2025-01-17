package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"speech-and-text/pkg/gcloud"
	"strings"

	"github.com/joho/godotenv"
)

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	bucketName := r.URL.Query().Get("bucket_name")
	audioFilePath := r.URL.Query().Get("audio_file_path")
	languageCodes := r.URL.Query().Get("language_codes")

	if bucketName == "" || audioFilePath == "" || languageCodes == "" {
		http.Error(w, "Missing required parameters: bucket_name, audio_file_path, language_codes", http.StatusBadRequest)
		return
	}

	config := &gcloud.SpeechToTextConfig{
		ProjectID:      os.Getenv("PROJECT_ID"),
		BucketName:     bucketName,
		AudioFilePath:  audioFilePath,
		LanguageCodes:  strings.Split(languageCodes, ","),
		TimeoutSeconds: 300,
	}

	transcripts, err := gcloud.SpeechToTextV2(config)
	if err != nil {
		http.Error(w, fmt.Sprintf("Speech-to-Text processing failed: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string][]string{
		"transcripts": transcripts,
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to marshal JSON response", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func main() {
	// ENVがローカルの場合だけ.envファイルを読み込む
	if os.Getenv("ENV") == "local" {
		err := godotenv.Load()
		if err != nil {
			log.Printf("Error loading .env file: %v", err)
		}
	}

	requiredEnvVars := []string{"PROJECT_ID"}
	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			log.Fatalf("Error: %s environment variable is not set", envVar)
		}
		log.Printf("%s=%s", envVar, os.Getenv(envVar))
	}

	port := os.Getenv("PORT")
	log.Printf("PORT=%s", port)
	if port == "" {
		port = "80"
		log.Printf("Defaulting to port %s", port)
	}

	http.HandleFunc("/", handler)
	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
