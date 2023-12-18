package elevenlabs

import (
	"fmt"

	"github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type TextToSpeechOpt func(*textToSpeechRequest) error
type TextToSpeechFormat string

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	MP3_44100_64  TextToSpeechFormat = "mp3_44100_64"  // mp3 with 44.1kHz sample rate at 64kbps
	MP3_44100_96  TextToSpeechFormat = "mp3_44100_96"  // mp3 with 44.1kHz sample rate at 96kbps
	MP3_44100_128 TextToSpeechFormat = "mp3_44100_128" // default output format, mp3 with 44.1kHz sample rate at 128kbps
	MP3_44100_192 TextToSpeechFormat = "mp3_44100_192" // mp3 with 44.1kHz sample rate at 192kbps
	PCM_16000     TextToSpeechFormat = "pcm_16000"     // PCM format (S16LE) with 16kHz sample rate
	PCM_22050     TextToSpeechFormat = "pcm_22050"     // PCM format (S16LE) with 22.05kHz sample rate
	PCM_24000     TextToSpeechFormat = "pcm_24000"     // PCM format (S16LE) with 24kHz sample rate
	PCM_44100     TextToSpeechFormat = "pcm_44100"     // PCM format (S16LE) with 44.1kHz sample rate
	ULAW_8000     TextToSpeechFormat = "ulaw_8000"     // μ-law format (sometimes written mu-law, often approximated as u-law) with 8kHz sample rate
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func OptOutput(format TextToSpeechFormat) TextToSpeechOpt {
	return func(req *textToSpeechRequest) error {
		req.Query.Set("output_format", string(format))
		return nil
	}
}

func OptOptimizeStreamingLatency(value uint8) TextToSpeechOpt {
	return func(req *textToSpeechRequest) error {
		req.Query.Set("optimize_streaming_latency", fmt.Sprint(value))
		return nil
	}
}

func OptModel(id string) TextToSpeechOpt {
	return func(req *textToSpeechRequest) error {
		if id == "" {
			return errors.ErrBadParameter.With("OptModel: id")
		}
		req.ModelId = id
		return nil
	}
}

func OptVoiceSettings(settings VoiceSettings) TextToSpeechOpt {
	return func(req *textToSpeechRequest) error {
		req.Settings = settings
		return nil
	}
}
