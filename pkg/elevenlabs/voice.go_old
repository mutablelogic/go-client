package elevenlabs

import (
	"bytes"
	"fmt"
	"net/http"

	// Packages
	"github.com/djthorpe/go-errors"
	"github.com/mutablelogic/go-client/pkg/client"
)

///////////////////////////////////////////////////////////////////////////////
// SCHEMA

type Voice struct {
	Id          string        `json:"voice_id"`
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	PreviewUrl  string        `json:"preview_url,omitempty"`
	Category    string        `json:"category,omitempty"`
	Settings    VoiceSettings `json:"settings"`
}

type VoiceSettings struct {
	SimilarityBoost float32 `json:"similarity_boost"`
	Stability       float32 `json:"stability"`
	Style           float32 `json:"style,omitempty"`
	UseSpeakerBoost bool    `json:"use_speaker_boost"`
}

///////////////////////////////////////////////////////////////////////////////
// PAYLOADS

type voiceDeleteRequest struct {
	client.Payload `json:"-"`
}

type voiceAddRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

type voicesResponse struct {
	Voices []Voice `json:"voices"`
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return current set of voices
func (c *Client) Voices() ([]Voice, error) {
	var request client.Payload
	var response voicesResponse
	if err := c.Do(request, &response, client.OptPath("voices")); err != nil {
		return nil, err
	}
	return response.Voices, nil
}

// Return a single voice
func (c *Client) Voice(Id string) (Voice, error) {
	var request client.Payload
	var response Voice
	if Id == "" {
		return response, errors.ErrBadParameter.With("Id")
	}
	if err := c.Do(request, &response, client.OptPath("voices", Id)); err != nil {
		return response, err
	}
	return response, nil
}

// Return voice settings. If Id is empty, then return the default voice settings
func (c *Client) VoiceSettings(Id string) (VoiceSettings, error) {
	var request client.Payload
	var response VoiceSettings
	var path client.RequestOpt
	if Id == "" {
		path = client.OptPath("voices", "settings", "default")
	} else {
		path = client.OptPath("voices", Id, "settings")
	}
	if err := c.Do(request, &response, path); err != nil {
		return response, err
	}
	return response, nil
}

// Delete a voice
func (c *Client) DeleteVoice(Id string) error {
	var request voiceDeleteRequest
	if Id == "" {
		return errors.ErrBadParameter.With("Id")
	}
	if err := c.Do(request, nil, client.OptPath("voices", Id)); err != nil {
		return err
	}
	return nil
}

// Add a voice
func (c *Client) AddVoice(Name, Description string) error {
	var request voiceAddRequest

	// Check parameters
	if Name == "" {
		return errors.ErrBadParameter.With("Name")
	}

	// Set request
	request.Name = Name
	request.Description = Description

	// Execute request
	if err := c.Do(request, nil, client.OptPath("voices", "add")); err != nil {
		return err
	}

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PAYLOAD METHODS

func (voiceDeleteRequest) Method() string {
	return http.MethodDelete
}

func (voiceDeleteRequest) Type() string {
	return ""
}

func (voiceDeleteRequest) Accept() string {
	return client.ContentTypeJson
}

func (voiceAddRequest) Method() string {
	return http.MethodPost
}

func (voiceAddRequest) Type() string {
	return client.ContentTypeBinary
}

func (voiceAddRequest) Accept() string {
	return client.ContentTypeForm
}

///////////////////////////////////////////////////////////////////////////////
// MARSHAL

func (v VoiceSettings) Marshal() ([]byte, error) {
	data := new(bytes.Buffer)
	data.Write([]byte(fmt.Sprintf("similarity_boost=%v\n", v.SimilarityBoost)))
	data.Write([]byte(fmt.Sprintf("stability=%v\n", v.Stability)))
	if v.Style != 0 {
		data.Write([]byte(fmt.Sprintf("style=%v\n", v.Style)))
	}
	if v.UseSpeakerBoost {
		data.Write([]byte(fmt.Sprintf("use_speaker_boost=%v\n", v.UseSpeakerBoost)))
	}
	return data.Bytes(), nil
}
