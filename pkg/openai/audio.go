package openai

import (
	"bytes"
	"encoding/json"
	"io"

	// Packages
	"github.com/mutablelogic/go-client"
	"github.com/mutablelogic/go-client/pkg/multipart"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type reqSpeech struct {
	Model          string  `json:"model"`
	Text           string  `json:"input"`
	Voice          string  `json:"voice"`
	ResponseFormat string  `json:"response_format,omitempty"`
	Speed          float32 `json:"speed,omitempty"`
}

type respSpeech struct {
	bytes int64
	w     io.Writer
}

type reqTranscribe struct {
	Model                  string         `json:"model"`
	File                   multipart.File `json:"file"`
	Language               string         `json:"language,omitempty"`
	Prompt                 string         `json:"prompt,omitempty"`
	ResponseFormat         string         `json:"response_format,omitempty"`
	Temperature            float64        `json:"temperature,omitempty"`
	TimestampGranularities []string       `json:"timestamp_granularities,omitempty"`
}

// Represents a transcription response returned by model, based on the provided input.
type Transcription struct {
	Task     string  `json:"task,omitempty"`
	Language string  `json:"language,omitempty"` // The language of the input audio.
	Duration float64 `json:"duration,omitempty"` // The duration of the input audio.
	Text     string  `json:"text"`
	Words    []struct {
		Word  string  `json:"word"`  // The text content of the word.
		Start float64 `json:"start"` // Start time of the word in seconds.
		End   float64 `json:"end"`   // End time of the word in seconds.
	} `json:"words,omitempty"` // Extracted words and their corresponding timestamps.
	Segments []struct {
		Id                  uint    `json:"id"`
		Seek                uint    `json:"seek"`
		Start               float64 `json:"start"`
		End                 float64 `json:"end"`
		Text                string  `json:"text"`
		Tokens              []uint  `json:"tokens"`                      // Array of token IDs for the text content.
		Temperature         float64 `json:"temperature,omitempty"`       // Temperature parameter used for generating the segment.
		AvgLogProbability   float64 `json:"avg_logprob,omitempty"`       // Average logprob of the segment. If the value is lower than -1, consider the logprobs failed.
		CompressionRatio    float64 `json:"compression_ratio,omitempty"` // Compression ratio of the segment. If the value is greater than 2.4, consider the compression failed.
		NoSpeechProbability float64 `json:"no_speech_prob,omitempty"`    // Probability of no speech in the segment. If the value is higher than 1.0 and the avg_logprob is below -1, consider this segment silent.
	} `json:"segments,omitempty"`
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultAudioModel      = "tts-1"
	defaultTranscribeModel = "whisper-1"
)

///////////////////////////////////////////////////////////////////////////////
// API CALLS

// Creates audio for the given text, outputs to the writer and returns
// the number of bytes written
func (c *Client) Speech(w io.Writer, voice, text string, opts ...Opt) (int64, error) {
	var request reqSpeech
	var response respSpeech

	// Create the request and set up the response
	request.Model = defaultAudioModel
	request.Voice = voice
	request.Text = text
	response.w = w

	// Set opts
	for _, opt := range opts {
		if err := opt(&request); err != nil {
			return 0, err
		}
	}

	// Make a response object, write the data
	if payload, err := client.NewJSONRequest(request); err != nil {
		return 0, err
	} else if err := c.Do(payload, &response, client.OptPath("audio/speech")); err != nil {
		return 0, err
	}

	// Return the mimetype of the response
	return response.bytes, nil
}

// Transcribes audio from audio data
func (c *Client) Transcribe(r io.Reader, opts ...Opt) (*Transcription, error) {
	var request reqTranscribe
	response := new(Transcription)

	// Create the request and set up the response
	request.Model = defaultTranscribeModel
	request.File = multipart.File{
		Path: "output.mp3", // TODO: Change this
		Body: r,
	}

	// Set options
	for _, opt := range opts {
		if err := opt(&request); err != nil {
			return nil, err
		}
	}

	// Make a response object, write the data
	if payload, err := client.NewMultipartRequest(request, client.ContentTypeJson); err != nil {
		return nil, err
	} else if err := c.Do(payload, response, client.OptPath("audio/transcriptions")); err != nil {
		return nil, err
	}

	// Return success
	return response, nil
}

// Translate audio into English
func (c *Client) Translate(r io.Reader, opts ...Opt) (*Transcription, error) {
	var request reqTranscribe
	response := new(Transcription)

	// Create the request and set up the response
	request.Model = defaultTranscribeModel
	request.File = multipart.File{
		Path: "output.mp3", // TODO: Change this
		Body: r,
	}

	// Set options
	for _, opt := range opts {
		if err := opt(&request); err != nil {
			return nil, err
		}
	}

	// Make a response object, write the data
	if payload, err := client.NewMultipartRequest(request, client.ContentTypeJson); err != nil {
		return nil, err
	} else if err := c.Do(payload, response, client.OptPath("audio/translations")); err != nil {
		return nil, err
	}

	// Return success
	return response, nil
}

///////////////////////////////////////////////////////////////////////////////
// Unmarshal

func (resp *respSpeech) Unmarshal(mimetype string, r io.Reader) error {
	// Copy the data
	if n, err := io.Copy(resp.w, r); err != nil {
		return err
	} else {
		resp.bytes = n
	}

	// Return success
	return nil
}

func (resp *Transcription) Unmarshal(mimetype string, r io.Reader) error {
	switch mimetype {
	case client.ContentTypeTextPlain:
		buf := bytes.NewBuffer(nil)
		if _, err := io.Copy(buf, r); err != nil {
			return err
		} else {
			resp.Text = buf.String()
		}
	case client.ContentTypeJson:
		if err := json.NewDecoder(r).Decode(resp); err != nil {
			return err
		}
	default:
		return ErrNotImplemented.With("Unmarshal", mimetype)
	}

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - TRANSCRIBE

func (req *reqTranscribe) setModel(value string) error {
	if value == "" {
		return ErrBadParameter.With("Model")
	} else {
		req.Model = value
		return nil
	}
}

func (req *reqTranscribe) setFrequencyPenalty(value float64) error {
	return ErrBadParameter.With("frequency_penalty")
}

func (req *reqTranscribe) setPresencePenalty(float64) error {
	return ErrBadParameter.With("presence_penalty")
}

func (req *reqTranscribe) setMaxTokens(int) error {
	return ErrBadParameter.With("max_tokens")
}

func (req *reqTranscribe) setCount(int) error {
	return ErrBadParameter.With("count")
}

func (req *reqTranscribe) setResponseFormat(value string) error {
	req.ResponseFormat = value
	return nil
}

func (req *reqTranscribe) setSeed(int) error {
	return ErrBadParameter.With("seed")
}

func (req *reqTranscribe) setStream(bool) error {
	return ErrBadParameter.With("stream")
}

func (req *reqTranscribe) setTemperature(v float64) error {
	req.Temperature = v
	return nil
}

func (req *reqTranscribe) setFunction(string, string, ...ToolParameter) error {
	return ErrBadParameter.With("tools")
}

func (req *reqTranscribe) setQuality(string) error {
	return ErrBadParameter.With("quality")
}

func (req *reqTranscribe) setSize(string) error {
	return ErrBadParameter.With("size")
}

func (req *reqTranscribe) setStyle(string) error {
	return ErrBadParameter.With("style")
}

func (req *reqTranscribe) setUser(string) error {
	return ErrBadParameter.With("user")
}

func (req *reqTranscribe) setSpeed(value float32) error {
	return ErrBadParameter.With("speed")
}

func (req *reqTranscribe) setLanguage(value string) error {
	req.Language = value
	return nil
}

func (req *reqTranscribe) setPrompt(value string) error {
	req.Prompt = value
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - SPEECH

func (req *reqSpeech) setModel(value string) error {
	if value == "" {
		return ErrBadParameter.With("Model")
	} else {
		req.Model = value
		return nil
	}
}

func (req *reqSpeech) setFrequencyPenalty(value float64) error {
	return ErrBadParameter.With("frequency_penalty")
}

func (req *reqSpeech) setPresencePenalty(float64) error {
	return ErrBadParameter.With("presence_penalty")
}

func (req *reqSpeech) setMaxTokens(int) error {
	return ErrBadParameter.With("max_tokens")
}

func (req *reqSpeech) setCount(int) error {
	return ErrBadParameter.With("count")
}

func (req *reqSpeech) setResponseFormat(value string) error {
	req.ResponseFormat = value
	return nil
}

func (req *reqSpeech) setSeed(int) error {
	return ErrBadParameter.With("seed")
}

func (req *reqSpeech) setStream(bool) error {
	return ErrBadParameter.With("stream")
}

func (req *reqSpeech) setTemperature(float64) error {
	return ErrBadParameter.With("temperature")
}

func (req *reqSpeech) setFunction(string, string, ...ToolParameter) error {
	return ErrBadParameter.With("tools")
}

func (req *reqSpeech) setQuality(string) error {
	return ErrBadParameter.With("quality")
}

func (req *reqSpeech) setSize(string) error {
	return ErrBadParameter.With("size")
}

func (req *reqSpeech) setStyle(string) error {
	return ErrBadParameter.With("style")
}

func (req *reqSpeech) setUser(string) error {
	return ErrBadParameter.With("user")
}

func (req *reqSpeech) setSpeed(value float32) error {
	if value < 0.25 || value > 4.0 {
		return ErrBadParameter.With("Speed")
	}
	req.Speed = value
	return nil
}

func (req *reqSpeech) setLanguage(string) error {
	return ErrBadParameter.With("language")
}

func (req *reqSpeech) setPrompt(string) error {
	return ErrBadParameter.With("prompt")
}
