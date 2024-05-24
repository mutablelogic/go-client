package client

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"
	"time"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

/////////////////////////////////////////////////////////////////////////////
// TYPES

// Implementation of a text stream, as per
// https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events
type TextStream struct {
	buf *bytes.Buffer
}

// Implementation of a text stream, as per
// https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events#event_stream_format
type TextStreamEvent struct {
	// The event ID to set the EventSource object's last event ID value.
	Id string `json:"id,omitempty"`

	// A string identifying the type of event described
	Event string `json:"event,omitempty"`

	// The data field for the message
	Data string `json:"data"`

	// The reconnection time. If the connection to the server is lost,
	// the client should wait for the specified time before attempting to reconnect.
	Retry time.Duration `json:"retry,omitempty"`
}

// Callback for text stream events, return an error if you want to return from the Decode method
type TextStreamCallback func(TextStreamEvent) error

/////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	// Mime type for text stream
	ContentTypeTextStream = "text/event-stream"
)

/////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new text stream decoder
func NewTextStream() *TextStream {
	return &TextStream{
		buf: new(bytes.Buffer),
	}
}

/////////////////////////////////////////////////////////////////////////////
// STRINGIFY

// Return the text stream event as a string
func (t TextStreamEvent) String() string {
	data, _ := json.MarshalIndent(t, "", "  ")
	return string(data)
}

/////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Decode a text stream. The reader should be a stream of text/event-stream data
// and the method will return when all the data has been scanned, or the callback
// returns an error
func (t *TextStream) Decode(r io.Reader, callback TextStreamCallback) error {
	var event *TextStreamEvent
	scanner := bufio.NewScanner(r)

	// Reset the buffer
	t.buf.Reset()

	// If the callback is nil, then return without doing anything
	if callback == nil {
		return nil
	}

	// Loop through until EOF or error
	for scanner.Scan() {
		data := strings.TrimSpace(scanner.Text())
		if data == "" {
			// Reset the buffer
			t.buf.Reset()

			// Eject the text stream event
			if !event.IsZero() {
				if err := callback(*event); err != nil {
					if errors.Is(err, io.EOF) {
						return nil
					} else {
						return err
					}
				} else {
					event = nil
				}
			}

			// Continue processing
			continue
		}

		// Split the data
		fields := strings.SplitN(data, ":", 2)
		if len(fields) != 2 {
			return ErrUnexpectedResponse.Withf("%q", data)
		}

		// Create a new event if necessary
		if event == nil {
			event = new(TextStreamEvent)
		}

		// Populate the event
		key := strings.TrimSpace(strings.ToLower(fields[0]))
		value := strings.TrimSpace(fields[1])
		switch key {
		case "id":
			event.Id = value
		case "event":
			event.Event = value
		case "data":
			// Concatenate data
			event.Data = event.Data + value
		case "retry":
			// Retry time in milliseconds, ignore if not a number
			if retry, err := strconv.ParseInt(value, 10, 64); err == nil {
				event.Retry = time.Duration(retry) * time.Millisecond
			}
		default:
			// Ignore other fields
		}
	}

	// Eject the final text stream event
	if !event.IsZero() {
		if err := callback(*event); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			} else {
				return err
			}
		}
	}

	// Return any scanner errors
	return scanner.Err()
}

// Return true if the event contains no content
func (t *TextStreamEvent) IsZero() bool {
	if t == nil {
		return true
	}
	return t.Id == "" && t.Event == "" && t.Data == "" && t.Retry == 0
}

// Decode the text stream event data as JSON
func (t *TextStreamEvent) Json(v any) error {
	// Do nothing if there is no data
	if t.Data == "" {
		return nil
	}
	return json.Unmarshal([]byte(t.Data), v)
}
