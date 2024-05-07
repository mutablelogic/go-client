package openai

// Packages

const (
	defaultAudioModel = "tts-1"
)

///////////////////////////////////////////////////////////////////////////////
// API CALLS
/*
// Creates audio for the given text.
func (c *Client) Speech(voice, text string, opts ...Opt) ([]byte, error) {
	// Create the request
	var request reqSpeech
	request.Model = defaultAudioModel
	request.Voice = voice
	request.Text = text
	for _, opt := range opts {
		if err := opt(&request); err != nil {
			return nil, err
		}
	}

	return nil, nil
}
*/
