package goEagi

import (
	"context"
	"fmt"

	"github.com/Microsoft/cognitive-services-speech-sdk-go/audio"
	"github.com/Microsoft/cognitive-services-speech-sdk-go/speech"
)

// AzureResult is a struct that contains transcription result from Azure Speech to Text service.
type AzureResult struct {
	Transcription string
	Info          string
	IsFinal       bool
	Error         error
}

// AzureService is used to stream audio data to Azure Speech to Text service.
type AzureService struct {
	subscriptionKey    string
	serviceRegion      string
	sourceLanguageCode []string
	recognizer         *speech.SpeechRecognizer
	InputStream        *audio.PushAudioInputStream

	SessionID      string
	SessionStarted bool

	result chan AzureResult
}

// NewAzureService creates a new AzureService instance,
// which is used to stream audio data to Azure Speech to Text service.
// endpoint argument is optional, if provided, then it is used to create speech config for custom speech service/model.
// if it is empty, then subscriptionKey and serviceRegion are used to create the speech config.
func NewAzureService(subscriptionKey string, serviceRegion string, endpoint string, sourceLanguageCode []string) (*AzureService, error) {
	format, err := audio.GetWaveFormatPCM(sampleRate, audioBitsPerSample, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to get default input format: %v\n", err)
	}
	defer format.Close()

	inputStream, err := audio.CreatePushAudioInputStreamFromFormat(format)
	if err != nil {
		return nil, fmt.Errorf("failed to create input stream: %v\n", err)
	}

	audioConfig, err := audio.NewAudioConfigFromStreamInput(inputStream)
	if err != nil {
		return nil, fmt.Errorf("failed to create audio config: %v\n", err)
	}
	defer audioConfig.Close()

	var speechConfig *speech.SpeechConfig

	if endpoint != "" {
		speechConfig, err = speech.NewSpeechConfigFromEndpointWithSubscription(endpoint, subscriptionKey)
	} else {
		speechConfig, err = speech.NewSpeechConfigFromSubscription(subscriptionKey, serviceRegion)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create speech config: %v\n", err)
	}

	var recognizer *speech.SpeechRecognizer

	if len(sourceLanguageCode) == 1 {
		recognizer, err = speech.NewSpeechRecognizerFromSourceLanguage(speechConfig, sourceLanguageCode[0], audioConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create speech recognizer: %v\n", err)
		}
	} else {
		langConfig, err := speech.NewAutoDetectSourceLanguageConfigFromLanguages(sourceLanguageCode)
		if err != nil {
			return nil, fmt.Errorf("failed to create language config: %v\n", err)
		}
		recognizer, err = speech.NewSpeechRecognizerFomAutoDetectSourceLangConfig(speechConfig, langConfig, audioConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create speech recognizer: %v\n", err)
		}
	}

	azure := AzureService{
		subscriptionKey:    subscriptionKey,
		serviceRegion:      serviceRegion,
		sourceLanguageCode: sourceLanguageCode,
		InputStream:        inputStream,
		recognizer:         recognizer,
		result:             make(chan AzureResult),
	}

	azure.recognizer.SessionStarted(azure.sessionStartedHandler)
	azure.recognizer.SessionStopped(azure.sessionStoppedHandler)
	azure.recognizer.Recognizing(azure.recognizingHandler)
	azure.recognizer.Recognized(azure.recognizedHandler)
	azure.recognizer.Canceled(azure.cancelledHandler)

	return &azure, nil
}

func (azure *AzureService) StartStreaming(ctx context.Context, stream <-chan []byte) <-chan error {
	startStream := make(chan error)

	go func() {
		defer close(startStream)
		defer azure.Close()

		startErrCh := azure.recognizer.StartContinuousRecognitionAsync()

		for {
			select {
			case <-ctx.Done():
				return

			case err := <-startErrCh:
				if err != nil {
					startStream <- fmt.Errorf("async start streaming error: %w\n", err)
					return
				}

			case buffer := <-stream:
				if err := azure.InputStream.Write(buffer); err != nil {
					startStream <- fmt.Errorf("streaming error: %w\n", err)
					return
				}
			}
		}
	}()

	return startStream
}

func (azure *AzureService) SpeechToTextResponse(ctx context.Context) <-chan AzureResult {
	transcriptStream := make(chan AzureResult)

	go func() {
		defer close(transcriptStream)

		for {
			select {
			case <-ctx.Done():
				return

			case result := <-azure.result:
				transcriptStream <- result
			}
		}
	}()

	return transcriptStream
}

func (azure *AzureService) Close() {
	azure.InputStream.CloseStream()
	<-azure.recognizer.StopContinuousRecognitionAsync()
	azure.InputStream.Close()
	azure.recognizer.Close()
}

func (azure *AzureService) sessionStartedHandler(event speech.SessionEventArgs) {
	defer event.Close()

	azure.SessionID = event.SessionID
	azure.SessionStarted = true

	azure.result <- AzureResult{
		Info: "azure session started",
	}
}

func (azure *AzureService) sessionStoppedHandler(event speech.SessionEventArgs) {
	defer event.Close()

	azure.SessionID = event.SessionID
	azure.SessionStarted = false

	azure.result <- AzureResult{
		Info: "azure session stopped",
	}
}

func (azure *AzureService) recognizingHandler(event speech.SpeechRecognitionEventArgs) {
	defer event.Close()

	azure.result <- AzureResult{
		Transcription: event.Result.Text,
		IsFinal:       false,
	}
}

func (azure *AzureService) recognizedHandler(event speech.SpeechRecognitionEventArgs) {
	defer event.Close()

	azure.result <- AzureResult{
		Transcription: event.Result.Text,
		IsFinal:       true,
	}
}

func (azure *AzureService) cancelledHandler(event speech.SpeechRecognitionCanceledEventArgs) {
	defer event.Close()

	azure.result <- AzureResult{
		Error: fmt.Errorf("cancelled: %v\n", event.ErrorDetails),
	}
}
