package client_test

import (
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	client "github.com/mutablelogic/go-client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

///////////////////////////////////////////////////////////////////////////////
// TextStream — unit tests (no network required)

func Test_TextStream_New(t *testing.T) {
	ts := client.NewTextStream()
	require.NotNil(t, ts)
	assert.Equal(t, "", ts.LastEventID())
	assert.Equal(t, time.Duration(0), ts.RetryDuration())
}

func Test_TextStream_NilCallbackIsNoop(t *testing.T) {
	ts := client.NewTextStream()
	err := ts.Decode(strings.NewReader("data: hello\n\n"), nil)
	assert.NoError(t, err)
}

func Test_TextStream_BasicDataEvent(t *testing.T) {
	ts := client.NewTextStream()
	var events []client.TextStreamEvent
	err := ts.Decode(strings.NewReader("data: hello\n\n"), func(e client.TextStreamEvent) error {
		events = append(events, e)
		return nil
	})
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "hello", events[0].Data)
}

func Test_TextStream_EventWithIdAndType(t *testing.T) {
	ts := client.NewTextStream()
	input := "id: 42\nevent: update\ndata: payload\n\n"
	var events []client.TextStreamEvent
	err := ts.Decode(strings.NewReader(input), func(e client.TextStreamEvent) error {
		events = append(events, e)
		return nil
	})
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "42", events[0].Id)
	assert.Equal(t, "update", events[0].Event)
	assert.Equal(t, "payload", events[0].Data)
	// LastEventID is updated
	assert.Equal(t, "42", ts.LastEventID())
}

func Test_TextStream_RetryField(t *testing.T) {
	ts := client.NewTextStream()
	input := "retry: 3000\ndata: x\n\n"
	err := ts.Decode(strings.NewReader(input), func(e client.TextStreamEvent) error {
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, 3000*time.Millisecond, ts.RetryDuration())
}

func Test_TextStream_RetryInvalidIgnored(t *testing.T) {
	ts := client.NewTextStream()
	input := "retry: notanumber\ndata: x\n\n"
	err := ts.Decode(strings.NewReader(input), func(e client.TextStreamEvent) error {
		return nil
	})
	require.NoError(t, err)
	// Retry was not parseable, retryDuration stays zero
	assert.Equal(t, time.Duration(0), ts.RetryDuration())
}

func Test_TextStream_MultiLineData(t *testing.T) {
	ts := client.NewTextStream()
	input := "data: line1\ndata: line2\ndata: line3\n\n"
	var events []client.TextStreamEvent
	err := ts.Decode(strings.NewReader(input), func(e client.TextStreamEvent) error {
		events = append(events, e)
		return nil
	})
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "line1\nline2\nline3", events[0].Data)
}

func Test_TextStream_CommentLinesIgnored(t *testing.T) {
	ts := client.NewTextStream()
	// Lines starting with ":" are comments
	input := ": this is a comment\ndata: real\n\n"
	var events []client.TextStreamEvent
	err := ts.Decode(strings.NewReader(input), func(e client.TextStreamEvent) error {
		events = append(events, e)
		return nil
	})
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "real", events[0].Data)
}

func Test_TextStream_MultipleEvents(t *testing.T) {
	ts := client.NewTextStream()
	input := "data: first\n\ndata: second\n\n"
	var events []client.TextStreamEvent
	err := ts.Decode(strings.NewReader(input), func(e client.TextStreamEvent) error {
		events = append(events, e)
		return nil
	})
	require.NoError(t, err)
	require.Len(t, events, 2)
	assert.Equal(t, "first", events[0].Data)
	assert.Equal(t, "second", events[1].Data)
}

func Test_TextStream_CallbackEOFStopsCleanly(t *testing.T) {
	ts := client.NewTextStream()
	input := "data: first\n\ndata: second\n\ndata: third\n\n"
	var count int
	err := ts.Decode(strings.NewReader(input), func(e client.TextStreamEvent) error {
		count++
		return io.EOF // stop after first event
	})
	assert.NoError(t, err) // io.EOF from callback → clean stop, no error returned
	assert.Equal(t, 1, count)
}

func Test_TextStream_CallbackErrorPropagates(t *testing.T) {
	ts := client.NewTextStream()
	input := "data: oops\n\n"
	sentinel := errors.New("stop!")
	err := ts.Decode(strings.NewReader(input), func(e client.TextStreamEvent) error {
		return sentinel
	})
	assert.ErrorIs(t, err, sentinel)
}

func Test_TextStream_IsZero_Nil(t *testing.T) {
	var e *client.TextStreamEvent
	assert.True(t, e.IsZero())
}

func Test_TextStream_IsZero_WithData(t *testing.T) {
	e := client.TextStreamEvent{Data: "hello"}
	assert.False(t, e.IsZero())
}

func Test_TextStream_IsZero_EmptyEvent(t *testing.T) {
	e := client.TextStreamEvent{}
	assert.True(t, e.IsZero())
}

func Test_TextStream_JsonDecode(t *testing.T) {
	e := client.TextStreamEvent{Data: `{"name":"alice"}`}
	var v struct{ Name string }
	err := e.Json(&v)
	require.NoError(t, err)
	assert.Equal(t, "alice", v.Name)
}

func Test_TextStream_JsonDecode_EmptyDataIsNoop(t *testing.T) {
	e := client.TextStreamEvent{}
	var v struct{ Name string }
	err := e.Json(&v)
	assert.NoError(t, err)
	assert.Equal(t, "", v.Name)
}

func Test_TextStream_String(t *testing.T) {
	e := client.TextStreamEvent{Event: "ping", Data: "ok"}
	s := e.String()
	assert.NotEmpty(t, s)
}

// Test that the final event (no trailing blank line) is still dispatched.
func Test_TextStream_FinalEventNoTrailingBlankLine(t *testing.T) {
	ts := client.NewTextStream()
	input := "data: last" // no trailing \n\n
	var events []client.TextStreamEvent
	err := ts.Decode(strings.NewReader(input), func(e client.TextStreamEvent) error {
		events = append(events, e)
		return nil
	})
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "last", events[0].Data)
}

func Test_TextStream_LargeDataLine(t *testing.T) {
	ts := client.NewTextStream()
	large := strings.Repeat("x", 256*1024)
	input := "event: result\ndata: " + large + "\n\n"

	var events []client.TextStreamEvent
	err := ts.Decode(strings.NewReader(input), func(e client.TextStreamEvent) error {
		events = append(events, e)
		return nil
	})

	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "result", events[0].Event)
	assert.Equal(t, large, events[0].Data)
}
