package gcloud

import "testing"

func TestSpeechToTextV2(t *testing.T) {
	tests := []struct {
		name string
		cfg  *SpeechToTextConfig
	}{
		{
			name: "Test SpeechToText",
			cfg: &SpeechToTextConfig{
				ProjectID:      "test-project",
				BucketName:     "test-bucket",
				AudioFilePath:  "test-audio.raw",
				LanguageCodes:  []string{"en-US"},
				TimeoutSeconds: 60,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := SpeechToTextV2(tt.cfg)
			if err != nil {
				t.Errorf("failed to convert speech to text: %v", err)
			}
		})
	}
}
