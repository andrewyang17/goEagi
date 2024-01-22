package goEagi

import (
	"context"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
)

const (
	audioExtension = ".wav"
)

type GoogleTTS struct {
	AudioOutputDirectory string
	LanguageCode         string
	VoiceName            string
}

func NewGoogleTTS(googleCred, audioOutputDir, languageCode, voiceName string) (*GoogleTTS, error) {
	tts := GoogleTTS{
		AudioOutputDirectory: audioOutputDir,
		LanguageCode:         languageCode,
		VoiceName:            voiceName,
	}

	if env := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); env == "" {
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", googleCred)
	}

	if _, err := os.Stat(tts.AudioOutputDirectory); os.IsNotExist(err) {
		if err := os.MkdirAll(tts.AudioOutputDirectory, os.ModePerm); err != nil {
			return nil, fmt.Errorf("failed to create audio output directory: %w", err)
		}
	}

	return &tts, nil
}

// GenerateAudio generates audio file from content.
// It returns audio file path without extension for playback, and error if any.
func (tts *GoogleTTS) GenerateAudio(content string) (string, error) {
	audioName := generateHash(strings.ToLower(content))
	audioFilepathWithoutWavExtension := filepath.Join(tts.AudioOutputDirectory, audioName)
	audioFilepathWithWavExtension := filepath.Join(tts.AudioOutputDirectory, audioName+audioExtension)

	if _, err := os.Stat(audioFilepathWithWavExtension); os.IsExist(err) {
		return audioFilepathWithoutWavExtension, nil
	}

	file, err := os.OpenFile(audioFilepathWithWavExtension, os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	ctx := context.Background()
	client, err := texttospeech.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	req := &texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{
				Text: content,
			},
		},
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: tts.LanguageCode,
			Name:         tts.VoiceName,
		},
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding:   texttospeechpb.AudioEncoding_LINEAR16,
			SampleRateHertz: 8000,
		},
	}

	resp, err := client.SynthesizeSpeech(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to synthesize speech: %w", err)
	}

	if _, err := file.Write(resp.AudioContent); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}
	return audioFilepathWithoutWavExtension, nil
}

// generateHash generates hash from input string.
func generateHash(input string) string {
	hasher := fnv.New32a()
	hasher.Write([]byte(input))
	hashValue := hasher.Sum32()
	hashString := strconv.FormatUint(uint64(hashValue), 10)

	return hashString
}
