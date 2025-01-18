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
	Env            string
}

// SpeechToTextV2 は音声をテキストに変換し、トランスクリプトを返す
func SpeechToTextV2(config *SpeechToTextConfig) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.TimeoutSeconds)*time.Second)
	defer cancel()

	var client *speech.Client
	var err error
	if config.Env == "local" {
		client, err = createSpeechClient(ctx)
	} else {
		client, err = createSpeechClientWithJSON(ctx)
	}
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
	return speech.NewClient(ctx)
}

func createSpeechClientWithJSON(ctx context.Context) (*speech.Client, error) {
	credentials := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credentials == "" {
		return nil, fmt.Errorf("GOOGLE_APPLICATION_CREDENTIALS environment variable is not set")
	}

	return speech.NewClient(ctx, option.WithCredentialsJSON([]byte(credentials)))
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

// SpeechToTextV2FromBytes はバイト配列の音声をテキストに変換し、トランスクリプトを返す
func SpeechToTextV2FromBytes(config *SpeechToTextConfig, audioBytes []byte) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.TimeoutSeconds)*time.Second)
	defer cancel()

	var client *speech.Client
	var err error
	if config.Env == "local" {
		client, err = createSpeechClient(ctx)
	} else {
		client, err = createSpeechClientWithJSON(ctx)
	}
	if err != nil {
		return "", fmt.Errorf("failed to create client: %v", err)
	}
	defer client.Close()

	req, err := createRecognizeRequestFromBytes(config, audioBytes)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := client.Recognize(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to recognize: %v", err)
	}

	if len(resp.GetResults()) > 0 && len(resp.GetResults()[0].GetAlternatives()) > 0 {
		return resp.GetResults()[0].GetAlternatives()[0].GetTranscript(), nil
	}

	return "", fmt.Errorf("no transcript found")
}

func createRecognizeRequestFromBytes(config *SpeechToTextConfig, audioBytes []byte) (*speechpb.RecognizeRequest, error) {
	projectID := os.Getenv("PROJECT_ID")
	recognizer := fmt.Sprintf("projects/%s/locations/global/recognizers/_", projectID)

	return &speechpb.RecognizeRequest{
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
		AudioSource: &speechpb.RecognizeRequest_Content{
			Content: audioBytes,
		},
	}, nil
}
