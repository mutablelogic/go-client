package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	// Package imports
	"github.com/mutablelogic/go-client/pkg/client"
	"github.com/mutablelogic/go-client/pkg/elevenlabs"
)

/////////////////////////////////////////////////////////////////////
// TYPES

type result struct {
	Path  string `json:"path"`
	Bytes int64  `json:"bytes_written"`
	Mime  string `json:"mime_type,omitempty"`
}

/////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	reVoiceId = regexp.MustCompile("^[a-zA-Z0-9]{20}$")
)

/////////////////////////////////////////////////////////////////////
// REGISTER FUNCTIONS

func ElevenlabsFlags(flags *Flags) {
	flags.String("elevenlabs-api-key", "${ELEVENLABS_API_KEY}", "ElevenLabs API key")
}

func ElevenlabsRegister(cmd []Client, opts []client.ClientOpt, flags *Flags) ([]Client, error) {
	elevenlabs, err := elevenlabs.New(flags.GetString("elevenlabs-api-key"), opts...)
	if err != nil {
		return nil, err
	}

	// Register commands
	cmd = append(cmd, Client{
		ns: "elevenlabs",
		cmd: []Command{
			{Name: "voices", Description: "Return registered voices", MinArgs: 2, MaxArgs: 2, Fn: elVoices(elevenlabs, flags)},
			{Name: "voice", Description: "Return voice information", Syntax: "<voice-id>", MinArgs: 3, MaxArgs: 3, Fn: elVoice(elevenlabs, flags)},
			{Name: "speak", Description: "Create speech from a prompt", Syntax: "<voice> <prompt>", MinArgs: 4, MaxArgs: 4, Fn: elSpeak(elevenlabs, flags)},
		},
	})

	// Return success
	return cmd, nil
}

/////////////////////////////////////////////////////////////////////
// API CALL FUNCTIONS

func elVoices(client *elevenlabs.Client, flags *Flags) CommandFn {
	return func() error {
		if voices, err := client.Voices(); err != nil {
			return err
		} else {
			return flags.Write(voices)
		}
	}
}

func elVoice(client *elevenlabs.Client, flags *Flags) CommandFn {
	return func() error {
		if voice, err := client.Voice(flags.Arg(2)); err != nil {
			return err
		} else {
			return flags.Write(voice)
		}
	}
}

func elSpeak(client *elevenlabs.Client, flags *Flags) CommandFn {
	return func() error {
		voice, err := elGetVoiceId(client, flags.Arg(2))
		if err != nil {
			return err
		}

		// Set options
		opts := []elevenlabs.Opt{}

		// Create the audio
		out := flags.GetOutFilename("speech.mp3", 0)
		file, err := os.Create(out)
		if err != nil {
			return err
		}
		defer file.Close()
		if n, err := client.TextToSpeech(file, voice, flags.Arg(3), opts...); err != nil {
			return err
		} else if err := flags.Write(result{Path: out, Bytes: n}); err != nil {
			return err
		}

		// Return success
		return nil
	}
}

/////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// return a voice-id given a parameter, which can be a voice-id or name
func elGetVoiceId(client *elevenlabs.Client, voice string) (string, error) {
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
	return "", fmt.Errorf("voice not found: %q", voice)
}
