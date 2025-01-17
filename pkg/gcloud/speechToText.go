package gcloud

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	speech "cloud.google.com/go/speech/apiv2"
	speechpb "cloud.google.com/go/speech/apiv2/speechpb"
	"google.golang.org/api/option"
)

// SpeechToTextConfig は音声認識に必要な設定を保持する構造体
type SpeechToTextConfig struct {
	ProjectID      string
	BucketName     string
	AudioFilePath  string
	LanguageCodes  []string
	TimeoutSeconds int
}

// NewSpeechToTextConfig は環境変数から設定を読み込んで新しいConfigを作成する
func NewSpeechToTextConfig() (*SpeechToTextConfig, error) {
	projectID := os.Getenv("PROJECT_ID")
	bucketName := os.Getenv("BUCKET_NAME")
	audioFilePath := os.Getenv("AUDIO_FILE_PATH")

	if projectID == "" || bucketName == "" || audioFilePath == "" {
		return nil, fmt.Errorf("required environment variables are not set: PROJECT_ID, BUCKET_NAME, AUDIO_FILE_PATH")
	}

	return &SpeechToTextConfig{
		ProjectID:      projectID,
		BucketName:     bucketName,
		AudioFilePath:  audioFilePath,
		TimeoutSeconds: 300,
	}, nil
}

// SpeechToTextV2 は音声をテキストに変換し、トランスクリプトを返す
func SpeechToTextV2(config *SpeechToTextConfig) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.TimeoutSeconds)*time.Second)
	defer cancel()

	client, err := createSpeechClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}
	defer client.Close()

	req, err := createBatchRecognizeRequest(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	return processBatchRecognition(ctx, client, req)
}

func createSpeechClient(ctx context.Context) (*speech.Client, error) {
	credentialsFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credentialsFile == "" {
		return nil, fmt.Errorf("GOOGLE_APPLICATION_CREDENTIALS environment variable is not set")
	}

	return speech.NewClient(ctx, option.WithCredentialsFile(credentialsFile))
}

func createBatchRecognizeRequest(config *SpeechToTextConfig) (*speechpb.BatchRecognizeRequest, error) {
	fileUri := fmt.Sprintf("gs://%s/%s", config.BucketName, config.AudioFilePath)
	projectID := os.Getenv("PROJECT_ID")
	recognizer := fmt.Sprintf("projects/%s/locations/global/recognizers/_", projectID)

	return &speechpb.BatchRecognizeRequest{
		Recognizer: recognizer,
		Config: &speechpb.RecognitionConfig{
			DecodingConfig: &speechpb.RecognitionConfig_AutoDecodingConfig{
				AutoDecodingConfig: &speechpb.AutoDetectDecodingConfig{},
			},
			Model:         "short",
			LanguageCodes: config.LanguageCodes,
			Features: &speechpb.RecognitionFeatures{
				ProfanityFilter:       true,
				EnableWordTimeOffsets: true,
				EnableWordConfidence:  true,
			},
		},
		Files: []*speechpb.BatchRecognizeFileMetadata{
			{
				AudioSource: &speechpb.BatchRecognizeFileMetadata_Uri{
					Uri: fileUri,
				},
			},
		},
		RecognitionOutputConfig: &speechpb.RecognitionOutputConfig{
			Output: &speechpb.RecognitionOutputConfig_InlineResponseConfig{
				InlineResponseConfig: &speechpb.InlineOutputConfig{},
			},
		},
	}, nil
}

func processBatchRecognition(ctx context.Context, client *speech.Client, req *speechpb.BatchRecognizeRequest) ([]string, error) {
	op, err := client.BatchRecognize(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create BatchRecognize: %v", err)
	}

	res, err := op.Wait(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for BatchRecognize: %v", err)
	}

	var transcripts []string
	for _, result := range res.GetResults() {
		transcript, err := processResult(result)
		if err != nil {
			log.Printf("Warning: %v", err)
			continue
		}
		transcripts = append(transcripts, transcript)
	}

	return transcripts, nil
}

func processResult(result *speechpb.BatchRecognizeFileResult) (string, error) {
	ir := result.GetInlineResult()
	if ir == nil {
		return "", fmt.Errorf("no inline result found")
	}

	tr := ir.GetTranscript()
	if tr == nil {
		return "", fmt.Errorf("no transcript found")
	}

	for _, res := range tr.GetResults() {
		alternatives := res.GetAlternatives()
		if len(alternatives) > 0 {
			return alternatives[0].GetTranscript(), nil
		}
	}

	return "", fmt.Errorf("no alternatives found in transcript")
}
