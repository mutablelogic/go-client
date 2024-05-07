package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	// Packages
	"github.com/mutablelogic/go-client/pkg/client"
	"github.com/mutablelogic/go-client/pkg/elevenlabs"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

/////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	reVoiceId = regexp.MustCompile("^[a-zA-Z0-9]{20}$")
)

/////////////////////////////////////////////////////////////////////
// REGISTER FUNCTIONS

func ElevenlabsFlags(flags *Flags) {
	flags.String("elevenlabs-api-key", "${ELEVENLABS_API_KEY}", "ElevenLabs API key")
	flags.String("elevenlabs-voice", "", "Voice")
}

func ElevenlabsRegister(cmd []Client, opts []client.ClientOpt, flags *Flags) ([]Client, error) {
	// Get API key
	key, err := flags.GetString("elevenlabs-api-key")
	if err != nil {
		return nil, err
	}

	// Create client
	elevenlabs, err := elevenlabs.New(key, opts...)
	if err != nil {
		return nil, err
	}

	// Register commands
	cmd = append(cmd, Client{
		ns: "elevenlabs",
		cmd: []Command{
			{Name: "voices", Description: "Return registered voices", MinArgs: 2, MaxArgs: 2, Fn: elevenlabsVoices(elevenlabs, flags)},
			{Name: "voice", Description: "Return a voice", Syntax: "<voice>", MinArgs: 3, MaxArgs: 3, Fn: elevenlabsVoice(elevenlabs, flags)},
			{Name: "preview", Description: "Preview a voice", Syntax: "<voice>", MinArgs: 3, MaxArgs: 3, Fn: elevenlabsVoicePreview(elevenlabs, flags)},
			{Name: "say", Description: "Text-to-speech", Syntax: "<voice> <text>", MinArgs: 3, MaxArgs: 3, Fn: elevenlabsTextToSpeech(elevenlabs, flags)},
		},
	})

	// Return success
	return cmd, nil
}

/////////////////////////////////////////////////////////////////////
// API CALL FUNCTIONS

func elevenlabsVoices(client *elevenlabs.Client, flags *Flags) CommandFn {
	return func() error {
		if voices, err := client.Voices(); err != nil {
			return err
		} else {
			return flags.Write(voices)
		}
	}
}

func elevenlabsVoice(client *elevenlabs.Client, flags *Flags) CommandFn {
	return func() error {
		voice, err := elevenlabsGetVoiceId(client, flags.Arg(2))
		if err != nil {
			return err
		} else if voice, err := client.Voice(voice); err != nil {
			return err
		} else {
			return flags.Write(voice)
		}
	}
}

func elevenlabsVoicePreview(client *elevenlabs.Client, flags *Flags) CommandFn {
	return func() error {
		voice, err := elevenlabsGetVoiceId(client, flags.Arg(2))
		if err != nil {
			return err
		} else if voice, err := client.Voice(voice); err != nil {
			return err
		} else if voice.PreviewUrl == "" {
			return ErrNotFound.Withf("%q", flags.Arg(2))
		} else {
			fmt.Println(voice.PreviewUrl)
			return nil
		}
	}
}

func elevenlabsTextToSpeech(client *elevenlabs.Client, flags *Flags) CommandFn {
	return func() error {
		// Determine the voice to use
		voice, err := flags.GetString("elevenlabs-voice")
		if err != nil {
			return err
		} else if voice == "" {
			return fmt.Errorf("missing argument: -elevenlabs-voice")
		}
		voice, err = elevenlabsGetVoiceId(client, voice)
		if err != nil {
			return err
		}

		data, err := client.TextToSpeech(flags.Arg(2), voice)
		if err != nil {
			return err
		} else if _, err := os.Stdout.Write(data); err != nil {
			return err
		}
		return nil
	}
}

/////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// return a voice-id given a parameter, which can be a voice-id or name
func elevenlabsGetVoiceId(client *elevenlabs.Client, voice string) (string, error) {
	if reVoiceId.MatchString(voice) {
		return voice, nil
	} else if voices, err := client.Voices(); err != nil {
		return "", err
	} else {
		for _, v := range voices {
			if strings.EqualFold(v.Name, voice) || v.Id == voice {
				return v.Id, nil
			}
		}
	}
	return "", ErrNotFound.Withf("%q", voice)
}
