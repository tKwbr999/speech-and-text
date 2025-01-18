package main

import (
	"encoding/json"
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
		Env:            os.Getenv("ENV"),
	}

	transcripts, err := gcloud.SpeechToTextV2(config)
	if err != nil {
		log.Fatalf("Speech-to-Text processing failed: %v", err)
		http.Error(w, "Speech-to-Text processing failed", http.StatusInternalServerError)
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

func speechToTextHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(10 << 20) // 10 MB limit
	if err != nil {
		http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("audio")
	if err != nil {
		http.Error(w, "Failed to retrieve audio file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 音声データを []byte に読み込む
	audioBytes := make([]byte, 0)
	buf := make([]byte, 1024)
	for {
		n, err := file.Read(buf)
		if err != nil && err.Error() != "EOF" {
			http.Error(w, "Failed to read audio data", http.StatusInternalServerError)
			return
		}
		if n == 0 {
			break
		}
		audioBytes = append(audioBytes, buf[:n]...)
	}

	config := &gcloud.SpeechToTextConfig{
		ProjectID:      os.Getenv("PROJECT_ID"),
		LanguageCodes:  []string{"ja-JP"},
		TimeoutSeconds: 300,
		Env:            os.Getenv("ENV"),
	}

	transcript, err := gcloud.SpeechToTextV2FromBytes(config, audioBytes)
	if err != nil {
		log.Fatalf("Speech-to-Text processing failed: %v", err)
		http.Error(w, "Speech-to-Text processing failed", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"text": transcript,
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to marshal JSON response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
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
	http.HandleFunc("/api/speech-to-text", speechToTextHandler)
	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
