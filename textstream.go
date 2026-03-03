package client

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"
	"time"

	// Package imports
	types "github.com/mutablelogic/go-server/pkg/types"
)

/////////////////////////////////////////////////////////////////////////////
// TYPES

// Implementation of a text stream, as per
// https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events
//
// After Decode returns, LastEventID and RetryDuration can be used to reconnect:
//
//	req.Header.Set("Last-Event-ID", stream.LastEventID())
type TextStream struct {
	lastEventID   string
	retryDuration time.Duration
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
	ContentTypeTextStream = types.ContentTypeTextStream
)

/////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new text stream decoder
func NewTextStream() *TextStream {
	return &TextStream{}
}

/////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS — RECONNECT STATE

// LastEventID returns the last event ID received during Decode.
// Send this as the "Last-Event-ID" request header when reconnecting.
func (t *TextStream) LastEventID() string {
	return t.lastEventID
}

// RetryDuration returns the server-requested reconnect delay from the most
// recent "retry:" field, or zero if none was received.
func (t *TextStream) RetryDuration() time.Duration {
	return t.retryDuration
}

/////////////////////////////////////////////////////////////////////////////
// STRINGIFY

// Return the text stream event as a string
func (t TextStreamEvent) String() string {
	return types.Stringify(t)
}

/////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Decode a text stream. The reader should be a stream of text/event-stream data
// and the method will return when all the data has been scanned, or the callback
// returns an error
func (t *TextStream) Decode(r io.Reader, callback TextStreamCallback) error {
	var event *TextStreamEvent
	scanner := bufio.NewScanner(r)

	// If the callback is nil, then return without doing anything
	if callback == nil {
		return nil
	}

	// Loop through until EOF or error
	for scanner.Scan() {
		data := strings.TrimSpace(scanner.Text())
		if data == "" {
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

		// Split at the first colon. A line with no colon has the whole line as
		// the field name with an empty value (SSE spec). Lines starting with ':'
		// are comments and must be silently ignored.
		fields := strings.SplitN(data, ":", 2)
		key := strings.ToLower(strings.TrimSpace(fields[0]))
		value := ""
		if len(fields) == 2 {
			value = strings.TrimSpace(fields[1])
		}
		if key == "" {
			// Comment line — skip
			continue
		}

		// Create a new event if necessary
		if event == nil {
			event = new(TextStreamEvent)
		}

		// Populate the event
		switch key {
		case "id":
			event.Id = value
			t.lastEventID = value
		case "event":
			event.Event = value
		case "data":
			// Per SSE spec, multiple data fields are joined with a newline
			if event.Data == "" {
				event.Data = value
			} else {
				event.Data = event.Data + "\n" + value
			}
		case "retry":
			// Retry time in milliseconds, ignore if not a number
			if retry, err := strconv.ParseInt(value, 10, 64); err == nil {
				event.Retry = time.Duration(retry) * time.Millisecond
				t.retryDuration = event.Retry
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
