// Package goEagi of audio.go provides functionality on
// audio streaming from file descriptor 3 in Asterisk,
// amplitude computation and
// wav audio generation from audio bytes.

package goEagi

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"syscall"

	"github.com/cryptix/wav"
)

const (
	audioSampleRate     = 8000
	audioBitsPerSample  = 16
	audioBytesPerSample = audioBitsPerSample / 8
	audioChannel        = 1

	defaultFileDescriptorPath = "/dev/fd/3"
)

type AudioResult struct {
	Error  error
	Stream []byte
}

func AudioStreaming(ctx context.Context) <-chan AudioResult {
	audioResultStream := make(chan AudioResult)

	go func() {
		defer close(audioResultStream)

		fd, err := syscall.Open(defaultFileDescriptorPath, syscall.O_RDONLY, 0755)
		if err != nil {
			r := AudioResult{Error: fmt.Errorf("could not open fd3: %v\n", err)}
			audioResultStream <- r
			return
		}

		buf := make([]byte, 1024)

		for {
			select {
			case <-ctx.Done():
				return

			default:
				n, err := syscall.Read(fd, buf)
				if err != nil {
					r := AudioResult{Error: fmt.Errorf("failed to read fd3: %v\n", err)}
					audioResultStream <- r
					return
				}

				if n > 0 {
					audioResultStream <- AudioResult{Stream: buf[:n]}
				}
			}
		}
	}()

	return audioResultStream
}

// ComputeAmplitude analyzes the amplitude of a sample slice of bytes.
func ComputeAmplitude(sample []byte) (float64, error) {
	parseData, err := parseRawData(sample)
	if err != nil {
		return 0, err
	}

	computeRms := rms(parseData)
	maxAmp := maxPossibleAmplitude()
	db := ratioToDb(computeRms, maxAmp)
	return db + 90, nil
}

// GenerateAudio writes a sample slice of bytes into an audio file.
// It returns a location path of an audio which passed in the function parameters.
// Please note that only wav extension is supported.
func GenerateAudio(sample []byte, audioDirectory string, audioName string) (string, error) {
	if fileExtension := filepath.Ext(audioName); fileExtension != ".wav" {
		return "", errors.New("audio name does not contain .wav extension")
	}

	if _, err := os.Stat(audioDirectory); os.IsNotExist(err) {
		if err := os.MkdirAll(audioDirectory, os.ModePerm); err != nil {
			return "", err
		}
	}

	audioPath := audioDirectory + audioName
	file, err := os.Create(audioPath)
	if err != nil {
		return "", fmt.Errorf("failed to create audio path: %v\n", err)
	}
	defer file.Close()

	meta := wav.File{
		NumberOfSamples: uint32(len(sample)),
		SampleRate:      audioSampleRate,
		SignificantBits: audioBitsPerSample,
		Channels:        audioChannel,
	}

	writer, err := meta.NewWriter(file)
	if err != nil {
		return "", err
	}
	defer writer.Close()
	bytesSampleSize := int(meta.SignificantBits) / 8

	for i := 0; i < len(sample); i += bytesSampleSize {
		if err := writer.WriteSample(sample[i : i+bytesSampleSize]); err != nil {
			return "", fmt.Errorf("failed to generate audio: %v\n", err)
		}
	}

	return audioPath, nil
}

// scaleFrame is used in parseRawData.
func scaleFrame(unscaled int) float64 {
	maxV := math.MaxInt16
	return float64(unscaled) / float64(maxV)
}

// bits16ToInt is used in parseRawData.
func bits16ToInt(b []byte) (int, error) {
	if len(b) != 2 {
		return 0, errors.New("slice of bytes must be length of 2")
	}

	var payload int16
	framesPerBuffer := bytes.NewReader(b)
	if err := binary.Read(framesPerBuffer, binary.LittleEndian, &payload); err != nil {
		return 0, err
	}
	return int(payload), nil
}

// parseRawData is used in ComputeAmplitude.
func parseRawData(rawData []byte) ([]float64, error) {
	var frames []float64
	for i := 0; i < len(rawData); i += audioBytesPerSample {
		rawFrame := rawData[i : i+audioBytesPerSample]
		unscaledFrame, err := bits16ToInt(rawFrame)
		if err != nil {
			return nil, err
		}
		scaled := scaleFrame(unscaledFrame)
		frames = append(frames, scaled)
	}
	return frames, nil
}

// rms is used in ComputeAmplitude.
func rms(samples []float64) float64 {
	sampleCount := len(samples) / audioBytesPerSample
	var sumSquare float64

	for _, sample := range samples {
		sumSquare += sample * sample
	}
	return math.Sqrt(sumSquare / float64(sampleCount))
}

// maxPossibleAmplitude is used in ComputeAmplitude.
func maxPossibleAmplitude() float64 {
	maxPossibleVal := math.Pow(2, float64(audioBitsPerSample))
	return maxPossibleVal / 2
}

// ratioToDb is used in ComputeAmplitude.
func ratioToDb(rms, maxAmplitude float64) float64 {
	ratio := rms / maxAmplitude
	return 20 * math.Log10(ratio)
}
