package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

	// Packages
	tablewriter "github.com/djthorpe/go-tablewriter"
	audio "github.com/go-audio/audio"
	wav "github.com/go-audio/wav"
	client "github.com/mutablelogic/go-client"
	elevenlabs "github.com/mutablelogic/go-client/pkg/elevenlabs"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	elName            = "elevenlabs"
	elClient          *elevenlabs.Client
	elExt             = "mp3"
	elBitrate         = uint64(32)    // in kbps
	elSamplerate      = uint64(44100) // in Hz
	elSimilarityBoost = float64(0.0)
	elStability       = float64(0.0)
	elUseSpeakerBoost = false
	elWriteSettings   = false
	reVoiceId         = regexp.MustCompile("^[a-z0-9-]{20}$")
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func elRegister(flags *Flags) {
	// Register flags required
	flags.String(elName, "elevenlabs-api-key", "${ELEVENLABS_API_KEY}", "API Key")
	flags.Float(elName, "similarity-boost", 0, "Similarity boost")
	flags.Float(elName, "stability", 0, "Voice stability")
	flags.Bool(elName, "use-speaker-boost", false, "Use speaker boost")
	flags.Unsigned(elName, "bitrate", 0, "Bit rate (kbps)")
	flags.Unsigned(elName, "samplerate", 0, "Sample rate (kHz)")

	// Register command set
	flags.Register(Cmd{
		Name:        elName,
		Description: "Elevenlabs API",
		Parse:       elParse,
		Fn: []Fn{
			{Name: "voices", Call: elVoices, Description: "Return registered voices"},
			{Name: "voice", Call: elVoice, Description: "Return one voice", MinArgs: 1, MaxArgs: 1, Syntax: "<voice-id>"},
			{Name: "settings", Call: elVoiceSettings, Description: "Return voice settings, or default settings. Set voice settings from -stability, -similarity-boost and -use-speaker-boost flags", MaxArgs: 1, Syntax: "(<voice-id>)"},
			{Name: "say", Call: elTextToSpeech, Description: "Text to speech", MinArgs: 2, Syntax: "<voice-id> <text>..."},
		},
	})
}

func elParse(flags *Flags, opts ...client.ClientOpt) error {
	// Set defaults
	if typ := flags.GetOutExt(); typ != "" {
		elExt = strings.ToLower(flags.GetOutExt())
	}

	// Create the client
	apiKey := flags.GetString("elevenlabs-api-key")
	if apiKey == "" {
		return fmt.Errorf("missing -elevenlabs-api-key flag")
	} else if client, err := elevenlabs.New(apiKey, opts...); err != nil {
		return err
	} else {
		elClient = client
	}

	// Get the bit rate and sample rate
	if bitrate, err := flags.GetValue("bitrate"); err == nil {
		if bitrate_, ok := bitrate.(uint64); ok && bitrate_ > 0 {
			elBitrate = bitrate_
		}
	}
	if samplerate, err := flags.GetValue("samplerate"); err == nil {
		if samplerate_, ok := samplerate.(uint64); ok && samplerate_ > 0 {
			elSamplerate = samplerate_
		}
	}

	// Similarity boost
	if value, err := flags.GetValue("similarity-boost"); err == nil {
		elSimilarityBoost = value.(float64)
		elWriteSettings = true
	} else if !errors.Is(err, ErrNotFound) {
		return err
	}

	// Stability
	if value, err := flags.GetValue("stability"); err == nil {
		elStability = value.(float64)
		elWriteSettings = true
	} else if !errors.Is(err, ErrNotFound) {
		return err
	}

	// Use speaker boost
	if value, err := flags.GetValue("use-speaker-boost"); err == nil {
		elUseSpeakerBoost = value.(bool)
		elWriteSettings = true
	} else if !errors.Is(err, ErrNotFound) {
		return err
	}

	// Return success
	return nil
}

/////////////////////////////////////////////////////////////////////
// API CALL FUNCTIONS

func elVoices(ctx context.Context, w *tablewriter.Writer, args []string) error {
	voices, err := elClient.Voices()
	if err != nil {
		return err
	}
	return w.Write(voices)
}

func elVoice(ctx context.Context, w *tablewriter.Writer, args []string) error {
	if voice, err := elVoiceId(args[0]); err != nil {
		return err
	} else if voice, err := elClient.Voice(voice); err != nil {
		return err
	} else {
		return w.Write(voice)
	}
}

func elVoiceSettings(ctx context.Context, w *tablewriter.Writer, args []string) error {
	var voice string
	if len(args) > 0 {
		if v, err := elVoiceId(args[0]); err != nil {
			return err
		} else {
			voice = v
		}
	}

	// Get voice settings
	settings, err := elClient.VoiceSettings(voice)
	if err != nil {
		return err
	}

	// Modify settings
	if elWriteSettings {
		// We need a voice in order to write the settings
		if voice == "" {
			return ErrBadParameter.With("Missing voice-id")
		}

		// Change parameters
		if elStability != 0.0 {
			settings.Stability = float32(elStability)
		}
		if elSimilarityBoost != 0.0 {
			settings.SimilarityBoost = float32(elSimilarityBoost)
		}
		if elUseSpeakerBoost != settings.UseSpeakerBoost {
			settings.UseSpeakerBoost = elUseSpeakerBoost
		}

		// Set voice settings
		if err := elClient.SetVoiceSettings(voice, settings); err != nil {
			return err
		}
	}

	return w.Write(settings)
}

func elTextToSpeech(ctx context.Context, w *tablewriter.Writer, args []string) error {
	// The voice to use
	voice, err := elVoiceId(args[0])
	if err != nil {
		return err
	}

	// Output format
	opts := []elevenlabs.Opt{}
	if format := elOutputFormat(); format != nil {
		opts = append(opts, format)
	} else {
		return ErrBadParameter.Withf("invalid output format %q", elExt)
	}

	// The text to speak
	text := strings.Join(args[1:], " ")

	// If wav, then wrap in a header
	if elExt == "wav" {
		// Create the writer
		writer := NewAudioWriter(w.Output().(io.WriteSeeker), int(elSamplerate), 1)
		defer writer.Close()

		// Read the data
		if n, err := elClient.TextToSpeech(writer, voice, text, opts...); err != nil {
			return err
		} else {
			elClient.Debugf("elTextToSpeech: generated %v bytes of PCM data", n)
		}
	} else if _, err := elClient.TextToSpeech(w.Output(), voice, text, opts...); err != nil {
		return err
	}

	// Return success
	return nil
}

/////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func elVoiceId(q string) (string, error) {
	if reVoiceId.MatchString(q) {
		return q, nil
	} else if voices, err := elClient.Voices(); err != nil {
		return "", err
	} else {
		for _, v := range voices {
			if strings.EqualFold(v.Name, q) || v.Id == q {
				return v.Id, nil
			}
		}
	}
	return "", ErrNotFound.Withf("%q", q)
}

func elOutputFormat() elevenlabs.Opt {
	switch elExt {
	case "mp3":
		return elevenlabs.OptFormatMP3(uint(elBitrate), uint(elSamplerate))
	case "wav":
		return elevenlabs.OptFormatPCM(uint(elSamplerate))
	case "ulaw":
		return elevenlabs.OptFormatULAW()
	}
	return nil
}

/////////////////////////////////////////////////////////////////////
// AUDIO WRITER

type wavWriter struct {
	enc *wav.Encoder
	buf *bytes.Buffer
	pcm *audio.IntBuffer
}

func NewAudioWriter(w io.WriteSeeker, sampleRate, channels int) *wavWriter {
	this := new(wavWriter)

	// Create a WAV encoder
	this.enc = wav.NewEncoder(w, sampleRate, 16, channels, 1)
	if this.enc == nil {
		return nil
	}

	// Create a buffer for the incoming byte data
	this.buf = bytes.NewBuffer(nil)

	// Make a PCM buffer with a capacity of 4096 samples
	this.pcm = &audio.IntBuffer{
		Format: &audio.Format{
			SampleRate:  this.enc.SampleRate,
			NumChannels: this.enc.NumChans,
		},
		SourceBitDepth: this.enc.BitDepth,
		Data:           make([]int, 0, 4096),
	}

	// Return the writer
	return this
}

func (a *wavWriter) Write(data []byte) (int, error) {
	// Write the data to the buffer
	if n, err := a.buf.Write(data); err != nil {
		return 0, err
	} else if err := a.Flush(); err != nil {
		return 0, err
	} else {
		return n, nil
	}
}

func (a *wavWriter) Flush() error {
	var n int
	var sample [2]byte

	// Read data until we have a full PCM buffer
	for {
		if a.buf.Len() < len(sample) {
			break
		} else if n, err := a.buf.Read(sample[:]); err != nil {
			return err
		} else if n != len(sample) {
			return ErrInternalAppError.With("short read")
		}

		// Append the sample data - Little Endian
		a.pcm.Data = append(a.pcm.Data, int(int16(sample[0])|int16(sample[1])<<8))
		n += 2
	}

	// Write the PCM data
	if n > 0 {
		if err := a.enc.Write(a.pcm); err != nil {
			return err
		}
	}

	// Reset the PCM data
	a.pcm.Data = a.pcm.Data[:0]

	// Return success
	return nil
}

func (a *wavWriter) Close() error {
	if err := a.Flush(); err != nil {
		return err
	}
	return a.enc.Close()
}
