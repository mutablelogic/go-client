package elevenlabs

import (
	// Packages
	client "github.com/mutablelogic/go-client"
)

///////////////////////////////////////////////////////////////////////////////
// SCHEMA

type Model struct {
	Id                   string  `json:"model_id" writer:",width:30"`
	Name                 string  `json:"name" writer:",width:30,wrap"`
	Description          string  `json:"description,omitempty" writer:",wrap"`
	CanBeFineTuned       bool    `json:"can_be_fine_tuned" writer:",width:5"`
	CanDoTextToSpeech    bool    `json:"can_do_text_to_speech" writer:",width:5"`
	CanDoVoiceConversion bool    `json:"can_do_voice_conversion" writer:",width:5"`
	CanUseStyle          bool    `json:"can_use_style" writer:",width:5"`
	CanUseSpeakerBoost   bool    `json:"can_use_speaker_boost" writer:",width:5"`
	ServesProVoices      bool    `json:"serves_pro_voices" writer:",width:5"`
	TokenCostFactor      float32 `json:"token_cost_factor" writer:",width:5,right"`
	RequiresAlphaAccess  bool    `json:"requires_alpha_access,omitempty" writer:",width:5"`
	Languages            []struct {
		Id   string `json:"language_id"`
		Name string `json:"name"`
	} `json:"languages,omitempty" writer:",wrap"`
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return models
func (c *Client) Models() ([]Model, error) {
	var response []Model
	if err := c.Do(nil, &response, client.OptPath("models")); err != nil {
		return nil, err
	}
	return response, nil
}
