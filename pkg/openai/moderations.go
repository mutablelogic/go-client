package openai

import (
	// Packages
	"context"

	client "github.com/mutablelogic/go-client"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type reqModerations struct {
	options
	Input []string `json:"input"`
}

type responseModerations struct {
	Id      string       `json:"id"`
	Model   string       `json:"model"`
	Results []Moderation `json:"results"`
}

// Moderation represents the moderation of a text, including whether it is flagged
type Moderation struct {
	Flagged    bool `json:"flagged"`
	Categories struct {
		Sexual                 bool `json:"sexual,omitempty"`
		Hate                   bool `json:"hate,omitempty"`
		Harassment             bool `json:"harassment,omitempty"`
		SelfHarm               bool `json:"self-harm,omitempty"`
		SexualMinors           bool `json:"sexual/minors,omitempty"`
		HateThreatening        bool `json:"hate/threatening,omitempty"`
		ViolenceGraphic        bool `json:"violence/graphic,omitempty"`
		SelfHarmIntent         bool `json:"self-harm/intent,omitempty"`
		HarasssmentThreatening bool `json:"harassment/threatening,omitempty"`
		Violence               bool `json:"violence,omitempty"`
	} `json:"categories,omitempty" writer:",wrap"`
	CategoryScores struct {
		Sexual                 float32 `json:"sexual,omitempty"`
		Hate                   float32 `json:"hate,omitempty"`
		Harassment             float32 `json:"harassment,omitempty"`
		SelfHarm               float32 `json:"self-harm,omitempty"`
		SexualMinors           float32 `json:"sexual/minors,omitempty"`
		HateThreatening        float32 `json:"hate/threatening,omitempty"`
		ViolenceGraphic        float32 `json:"violence/graphic,omitempty"`
		SelfHarmIntent         float32 `json:"self-harm/intent,omitempty"`
		HarasssmentThreatening float32 `json:"harassment/threatening,omitempty"`
		Violence               float32 `json:"violence,omitempty"`
	} `json:"category_scores,omitempty" writer:",wrap"`
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultModerationModel = "text-moderation-latest"
)

///////////////////////////////////////////////////////////////////////////////
// API CALLS

// Classifies if text is potentially harmful
func (c *Client) Moderations(ctx context.Context, text []string, opts ...Opt) ([]Moderation, error) {
	var request reqModerations
	var response responseModerations

	// Set options
	request.Model = defaultModerationModel
	request.Input = text
	for _, opt := range opts {
		if err := opt(&request.options); err != nil {
			return nil, err
		}
	}

	// Request->Response
	if payload, err := client.NewJSONRequest(request); err != nil {
		return nil, err
	} else if err := c.DoWithContext(ctx, payload, &response, client.OptPath("moderations")); err != nil {
		return nil, err
	}

	// Return success
	return response.Results, nil
}
