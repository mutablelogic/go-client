/*
elevenlabs implements an API client for elevenlabs (https://elevenlabs.io/docs/api-reference/text-to-speech)
*/
package elevenlabs

import (
	"net/http"

	// Packages
	"github.com/mutablelogic/go-client/pkg/client"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Client struct {
	*client.Client
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	endPoint    = "https://api.elevenlabs.io/v1"
	Sample64    = "mp3_44100_64"  // mp3 with 44.1kHz sample rate at 64kbps
	Sample96    = "mp3_44100_96"  // mp3 with 44.1kHz sample rate at 96kbps
	Sample128   = "mp3_44100_128" // default output format, mp3 with 44.1kHz sample rate at 128kbps
	Sample192   = "mp3_44100_192" // mp3 with 44.1kHz sample rate at 192kbps
	SamplePCM16 = "pcm_16000"     // PCM format (S16LE) with 16kHz sample rate
	SamplePCM22 = "pcm_22050"     // PCM format (S16LE) with 22.05kHz sample rate
	SamplePCM24 = "pcm_24000"     // PCM format (S16LE) with 24kHz sample rate
	SamplePCM44 = "pcm_44100"     // PCM format (S16LE) with 44.1kHz sample rate
	SampleU8    = "ulaw_8000"     // μ-law format (sometimes written mu-law, often approximated as u-law) with 8kHz sample rate
)

///////////////////////////////////////////////////////////////////////////////
// SCHEMA

type Request struct {
	client.Payload `json:"-"`
	Model          string `json:"model_id"`
	Text           string `json:"text"`
	VoiceSettings  struct {
		SimilarityBoost float64 `json:"similarity_boost"`
		Stability       float64 `json:"stability"`
		Style           string  `json:"style,omitempty"`
		UseSpeakerBoost bool    `json:"use_speaker_boost"`
	} `json:"voice_settings"`
}

func (r Request) Method() string {
	return http.MethodGet
}

func (r Request) Type() string {
	return ""
}

func (r Request) Accept() string {
	return client.ContentTypeJson
}

type VoicesResponse struct {
	Voices []Voice `json:"voices"`
}

type Voice struct {
	Id          string `json:"voice_id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	PreviewUrl  string `json:"preview_url,omitempty"`
	Category    string `json:"category,omitempty"`
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func New(ApiKey string, opts ...client.ClientOpt) (*Client, error) {
	// Create client
	client, err := client.New(append(opts, client.OptEndpoint(endPoint), client.OptHeader("Xi-Api-Key", ApiKey))...)
	if err != nil {
		return nil, err
	}

	// Return the client
	return &Client{client}, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return current set of voices
func (c *Client) Voices() ([]Voice, error) {
	var request Request
	var response VoicesResponse
	if err := c.Do(request, &response, client.OptPath("voices")); err != nil {
		return nil, err
	}
	return response.Voices, nil
}

// Get returns the current IP address from the API
func (c *Client) TextToSpeech(Text string) ([]byte, error) {
	var request Request
	if err := c.Do(request, nil, client.OptPath("text-to-speech", "test")); err != nil {
		return nil, err
	}
	return nil, nil
}
