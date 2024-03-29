// Package goEagi of vad.go provides functionality on
// detecting voice/speech activity based on audio bytes.

package goEagi

const (
	defaultAmplitudeDetectionThreshold = -27.5
)

type VadResult struct {
	Error     error
	Detected  bool
	Amplitude float64
	Frame     []byte
}

type Vad struct {
	AmplitudeDetectionThreshold float64
}

// NewVad is a constructor of Vad.
// The initialization will use the defaultAmplitudeDetectionThreshold.
func NewVad(amplitudeThreshold float64) *Vad {
	if amplitudeThreshold != 0 {
		return &Vad{AmplitudeDetectionThreshold: amplitudeThreshold}
	}
	return &Vad{AmplitudeDetectionThreshold: defaultAmplitudeDetectionThreshold}
}

// Detect analyzes voice activity for a given slice of bytes.
func (v *Vad) Detect(done <-chan interface{}, stream <-chan []byte) <-chan VadResult {

	vadResultStream := make(chan VadResult)

	go func() {
		defer close(vadResultStream)

		for {
			select {
			case <-done:
				return

			case buf := <-stream:
				amp, err := ComputeAmplitude(buf)
				if err != nil {
					vadResultStream <- VadResult{Error: err}
					return
				}

				if v.AmplitudeDetectionThreshold < amp {
					vadResultStream <- VadResult{Detected: true, Amplitude: amp, Frame: buf}
				}
			}
		}
	}()

	return vadResultStream
}
