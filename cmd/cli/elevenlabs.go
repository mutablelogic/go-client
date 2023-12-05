package main

import (
	"encoding/json"
	"fmt"
	"os"

	// Packages
	"github.com/mutablelogic/go-client/pkg/client"
	"github.com/mutablelogic/go-client/pkg/elevenlabs"
)

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
			{Name: "voices", Description: "Return registered voices", MinArgs: 2, MaxArgs: 2, Fn: ElevenlabsVoices(elevenlabs, flags)},
			{Name: "voice", Description: "Return a voice", MinArgs: 3, MaxArgs: 3, Fn: ElevenlabsVoice(elevenlabs, flags)},
			{Name: "tts", Description: "Text-to-speech", MinArgs: 3, MaxArgs: 3, Fn: ElevenlabsTextToSpeech(elevenlabs, flags)},
		},
	})

	// Return success
	return cmd, nil
}

func ElevenlabsVoices(client *elevenlabs.Client, flags *Flags) CommandFn {
	return func() error {
		if voices, err := client.Voices(); err != nil {
			return err
		} else {
			return flags.Write(voices)
		}
	}
}

func ElevenlabsVoice(client *elevenlabs.Client, flags *Flags) CommandFn {
	return func() error {
		if voice, err := client.Voice(flags.Arg(2)); err != nil {
			return err
		} else if data, err := json.MarshalIndent(voice, "", "  "); err != nil {
			return err
		} else {
			fmt.Printf("%s\n", data)
		}
		return nil
	}
}

func ElevenlabsTextToSpeech(client *elevenlabs.Client, flags *Flags) CommandFn {
	return func() error {
		// Determine the voice to use
		voice, err := flags.GetString("elevenlabs-voice")
		if err != nil {
			return err
		} else if voice == "" {
			return fmt.Errorf("missing argument: -elevenlabs-voice")
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
