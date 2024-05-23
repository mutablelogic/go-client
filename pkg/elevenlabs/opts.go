package elevenlabs

import (
	"fmt"
	"net/url"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type opts struct {
	url.Values
	Model string `json:"model_id,omitempty"`
	Seed  uint   `json:"seed,omitempty"`
}

// Opt is a function which can be used to set options on a request
type Opt func(*opts) error

///////////////////////////////////////////////////////////////////////////////
// OPTIONS

// Set the voice model
func OptModel(v string) Opt {
	return func(o *opts) error {
		o.Model = v
		return nil
	}
}

// Set the deterministic seed
func OptSeed(v uint) Opt {
	return func(o *opts) error {
		o.Seed = v
		return nil
	}
}

// Set the output format
func OptFormat(v string) Opt {
	return func(o *opts) error {
		if o.Values == nil {
			o.Values = make(url.Values)
		}
		if v == "" {
			o.Del("output_format")
		} else {
			o.Set("output_format", v)
		}
		return nil
	}
}

// Set the output format to MP3 given bitrate and samplerate
func OptFormatMP3(bitrate, samplerate uint) Opt {
	return func(o *opts) error {
		switch samplerate {
		case 22050, 44100:
			return OptFormat(fmt.Sprintf("mp3_%v_%v", samplerate, bitrate))(o)
		default:
			return ErrBadParameter.With("OptFormatMP3: invalid sample rate: ", samplerate)
		}
	}
}

// Set the output format to PCM
func OptFormatPCM(samplerate uint) Opt {
	return func(o *opts) error {
		switch samplerate {
		case 16000, 22050, 24000, 44100:
			return OptFormat(fmt.Sprintf("pcm_%v", samplerate))(o)
		default:
			return ErrBadParameter.With("OptFormatPCM: invalid sample rate: ", samplerate)
		}
	}
}

// Set the output format to ULAW
func OptFormatULAW() Opt {
	return func(o *opts) error {
		return OptFormat("ulaw_8000")(o)
	}
}
