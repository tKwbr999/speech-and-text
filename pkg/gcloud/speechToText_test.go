package gcloud

import "testing"

func TestSpeechToTextV2(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "Test SpeechToText",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SpeechToTextV2()
			if err != nil {
				t.Errorf("failed to convert speech to text: %v", err)
			}
		})
	}
}
