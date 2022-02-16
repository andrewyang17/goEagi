// Package goEagi of google.go provides a simplified interface
// for calling Google's speech to text service.
// It provides flexibility to the callers and allow them to
// set their desired configuration.

package goEagi

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	speech "cloud.google.com/go/speech/apiv1"
	speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1"
)

type GoogleResult struct {
	Error         error
	Transcription string
}

// GoogleService provides information to Google Speech Recognizer
// as well as methods on calling speech to text.
type GoogleService struct {
	languageCode   string
	model          string
	privateKeyPath string
	sampleRate     int32
	enhancedMode   bool
	client         speechpb.Speech_StreamingRecognizeClient
}

// NewGoogleService is a constructor of GoogleService,
// it takes a privateKeyPath to set it in environment with key GOOGLE_APPLICATION_CREDENTIALS,
// and a languageCode, example ["en-GB", "en-US", "en-SG", ...], see [Language Support](https://cloud.google.com/speech-to-text/docs/languages).
func NewGoogleService(privateKeyPath string, languageCode string) (*GoogleService, error) {
	if len(strings.TrimSpace(privateKeyPath)) == 0 {
		return nil, errors.New("private key path is empty")
	}

	err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to set env: %v\n", err)
	}

	g := GoogleService{
		languageCode:   languageCode,
		model:          "phone_call",
		privateKeyPath: privateKeyPath,
		sampleRate:     8000,
		enhancedMode:   true,
	}

	ctx := context.Background()

	client, err := speech.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	g.client, err = client.StreamingRecognize(ctx)
	if err != nil {
		return nil, err
	}

	if err := g.client.Send(&speechpb.StreamingRecognizeRequest{
		StreamingRequest: &speechpb.StreamingRecognizeRequest_StreamingConfig{
			StreamingConfig: &speechpb.StreamingRecognitionConfig{
				Config: &speechpb.RecognitionConfig{
					Encoding:        speechpb.RecognitionConfig_LINEAR16,
					SampleRateHertz: g.sampleRate,
					LanguageCode:    g.languageCode,
					Model:           g.model,
					UseEnhanced:     g.enhancedMode,
				},
			},
		},
	}); err != nil {
		return nil, err
	}

	return &g, nil
}

// StartStreaming takes a reading channel of audio stream and send it
// as a gRPC request to Google service through the initialized client.
// Caller should run it in a goroutine.
func (g *GoogleService) StartStreaming(ctx context.Context, stream <-chan []byte) <-chan error {
	startStream := make(chan error)

	go func() {
		defer close(startStream)

		for {
			select {
			case <-ctx.Done():
				return

			case s := <-stream:
				if err := g.client.Send(&speechpb.StreamingRecognizeRequest{
					StreamingRequest: &speechpb.StreamingRecognizeRequest_AudioContent{
						AudioContent: s,
					},
				}); err != nil {
					startStream <- fmt.Errorf("streaming error: %v\n", err)
					return
				}
			}
		}
	}()

	return startStream
}

// SpeechToTextResponse sends the transcription response from Google's SpeechToText.
func (g *GoogleService) SpeechToTextResponse(ctx context.Context) <-chan GoogleResult {
	googleResultStream := make(chan GoogleResult)

	go func() {
		defer close(googleResultStream)

		for {
			select {
			case <-ctx.Done():
				return

			default:
				resp, err := g.client.Recv()
				if err == io.EOF {
					googleResultStream <- GoogleResult{Error: err}
					return
				}

				if err != nil {
					googleResultStream <- GoogleResult{Error: fmt.Errorf("cannot stream results: %v", err)}
				}

				for _, result := range resp.Results {
					googleResultStream <- GoogleResult{Transcription: result.Alternatives[0].Transcript}
				}
			}
		}
	}()

	return googleResultStream
}
