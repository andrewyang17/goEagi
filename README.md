## GoEagi

Package GoEagi provides some fundamental functionalities that work with Asterisk's EAGI. It has the following
features:

- Audio Streaming
- Google's Speech to Text
- Voice Activity Detection
- Speech File Generation
- Commands to Asterisk

### Example Usage
- Asterisk audio streaming + Google's speech to text
```go
package main

import (
	"fmt"
	"github.com/andrewyang17/goEagi"
	"golang.org/x/net/context"
)

func main() {
	eagi, err := goEagi.New()
	if err != nil {
		panic(err)
	}

	googleService, err := goEagi.NewGoogleService("<GoogleSpeechToTextPrivateKey>", "en-GB")
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bridgeStream := make(chan []byte)
	defer close(bridgeStream)

	audioStream := goEagi.AudioStreaming(ctx)
	errStream := googleService.StartStreaming(ctx, bridgeStream)
	googleStream := googleService.SpeechToTextResponse(ctx)

loop:
	for {
		select {
		case <-errStream:
			cancel()
			break loop

		case a := <-audioStream:
			if a.Error != nil {
				cancel()
				break loop
			}

			if len(a.Stream) != 0 {
				bridgeStream <- a.Stream
			}

		case g := <-googleStream:
			if g.Error != nil {
				cancel()
				break loop
			}
            
			// Do whatever you want with the returning transcription,
			// in this case we stdout
			if err := eagi.Verbose(fmt.Sprintf("Transcription: %v\n", g.Transcription)); err != nil {
				panic(err)
            }
		}
	}
}
```

### Commands to Asterisk
- refer [here](https://github.com/zaf/agi)
