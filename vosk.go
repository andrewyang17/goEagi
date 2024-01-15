// Package goEagi of vosk.go provides a simplified interface
// for calling Vosk Server's speech to text service.
// It provides flexibility to the callers and allow them to
// set their desired configuration.
package goEagi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/gorilla/websocket"
)

// VoskResult is the response from Vosk Speech Recognizer.
type VoskResult struct {
	Result []struct {
		Conf  float64
		End   float64
		Start float64
		Word  string
	}
	Text    string
	Partial string
}

// VoskService is the client for Vosk Speech Recognizer.
type VoskService struct {
	PhraseList  []string        `json:"phrase_list"`
	Words       bool            `json:"words"`
	Client      *websocket.Conn `json:"-"`
	errorStream chan error      `json:"-"`
}

// VoskConfig is the configuration for Vosk Speech Recognizer.
type voskConfig struct {
	Config VoskService `json:"config"`
}

// NewVoskService creates a new VoskService.
func NewVoskService(host string, port string, phraseList []string) (*VoskService, error) {

	h := fmt.Sprintf("%s:%s", host, port)
	u := url.URL{Scheme: "ws", Host: h, Path: ""}

	// Opening websocket connection
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}

	v := VoskService{
		PhraseList: phraseList,
		Client:     c,
	}

	config := voskConfig{
		Config: v,
	}
	configJSON, _ := json.Marshal(config)

	err = c.WriteMessage(websocket.TextMessage, configJSON)
	if err != nil {
		return nil, err
	}

	return &v, nil
}

// StartStreaming starts the streaming to Vosk speech to text service.
// It takes a reading channel of audio stream and sends it as a websocket binary message to Vosk service.
func (v *VoskService) StartStreaming(ctx context.Context, stream <-chan []byte) <-chan error {
	v.errorStream = make(chan error)

	go func() {
		defer close(v.errorStream)
		defer v.Client.Close()

		for {
			select {
			case <-ctx.Done():
				v.Close()
				return

			case buf := <-stream:
				err := v.Client.WriteMessage(websocket.BinaryMessage, buf)
				if err != nil {
					v.errorStream <- fmt.Errorf("streaming error: %v", err)
					return
				}
			}
		}
	}()

	return v.errorStream
}

// Close the websocket connection to Vosk service.
func (v *VoskService) Close() error {
	err := v.Client.WriteMessage(websocket.TextMessage, []byte("{\"eof\" : 1}"))
	return err
}

// SpeechToTextResponse sends the transcription response from Vosk's SpeechToText.
func (v *VoskService) SpeechToTextResponse(ctx context.Context) <-chan VoskResult {
	voskResultStream := make(chan VoskResult)

	go func() {
		defer close(voskResultStream)

		for {
			select {
			case <-ctx.Done():
				return

			default:
				_, msg, err := v.Client.ReadMessage()
				if err != nil {
					v.errorStream <- err
					return
				}

				m := VoskResult{}
				err = json.Unmarshal(msg, &m)
				if err != nil {
					v.errorStream <- err
					return
				}
				if m.Text != "" || m.Partial != "" {
					voskResultStream <- m
				}
			}
		}
	}()

	return voskResultStream
}
