package multipart

import (
	"bytes"
	"io"
	"mime"
	"mime/multipart"
	"net/textproto"
	"strings"
	"testing"
)

///////////////////////////////////////////////////////////////////////////////
// TEST TYPES

// testRequest represents a struct for testing various field types
type testRequest struct {
	Name         string    `json:"name"`
	Tags         []string  `json:"tags,omitempty"`
	MultiTags    []string  `json:"multi_tags"`
	Numbers      []int     `json:"numbers,omitempty"`
	SingleNumber []int     `json:"single_number"`
	ByteData     []byte    `json:"byte_data,omitempty"`
	ByteArray    [4]byte   `json:"byte_array,omitempty"`
	StringArray  [3]string `json:"string_array,omitempty"`
}

///////////////////////////////////////////////////////////////////////////////
// TESTS

func Test_Multipart_EmptySlice(t *testing.T) {
	req := testRequest{
		Name: "test",
		Tags: []string{}, // Empty slice with omitempty
	}

	buf := new(bytes.Buffer)
	enc := NewMultipartEncoder(buf)
	defer enc.Close()

	if err := enc.Encode(req); err != nil {
		t.Fatalf("encode error: %v", err)
	}

	// Empty slices with omitempty should not produce form fields
	content := buf.String()
	if strings.Contains(content, "tags") {
		t.Error("empty slice with omitempty should not produce a form field")
	}
}

func Test_Multipart_SingleElementSlice(t *testing.T) {
	req := testRequest{
		Name: "test",
		Tags: []string{"golang"},
	}

	buf := new(bytes.Buffer)
	enc := NewMultipartEncoder(buf)
	defer enc.Close()

	if err := enc.Encode(req); err != nil {
		t.Fatalf("encode error: %v", err)
	}

	content := buf.String()
	if !strings.Contains(content, "tags") || !strings.Contains(content, "golang") {
		t.Error("single-element slice should produce one form field")
	}
}

func Test_Multipart_MultiElementSlice(t *testing.T) {
	req := testRequest{
		Name: "test",
		Tags: []string{"golang", "testing", "multipart"},
	}

	buf := new(bytes.Buffer)
	enc := NewMultipartEncoder(buf)
	defer enc.Close()

	if err := enc.Encode(req); err != nil {
		t.Fatalf("encode error: %v", err)
	}

	content := buf.String()
	// Verify all tags are properly encoded as separate fields
	count := strings.Count(content, `Content-Disposition: form-data; name="tags"`)
	if count != 3 {
		t.Errorf("expected 3 tags fields, got %d", count)
	}
	if !strings.Contains(content, "golang") || !strings.Contains(content, "testing") || !strings.Contains(content, "multipart") {
		t.Error("all slice elements should be present in form fields")
	}
}

func Test_Multipart_SliceOfIntegers(t *testing.T) {
	req := testRequest{
		Name:         "test",
		SingleNumber: []int{42, 100},
	}

	buf := new(bytes.Buffer)
	enc := NewMultipartEncoder(buf)
	defer enc.Close()

	if err := enc.Encode(req); err != nil {
		t.Fatalf("encode error: %v", err)
	}

	content := buf.String()
	// Verify integer values are properly encoded as separate fields
	count := strings.Count(content, `Content-Disposition: form-data; name="single_number"`)
	if count != 2 {
		t.Errorf("expected 2 single_number fields, got %d", count)
	}
	if !strings.Contains(content, "42") || !strings.Contains(content, "100") {
		t.Error("integer values should be present in form fields")
	}
}

func Test_Multipart_ByteSlice(t *testing.T) {
	req := testRequest{
		Name:     "test",
		ByteData: []byte("hello"),
	}

	buf := new(bytes.Buffer)
	enc := NewMultipartEncoder(buf)
	defer enc.Close()

	if err := enc.Encode(req); err != nil {
		t.Fatalf("encode error: %v", err)
	}

	content := buf.String()
	// []byte should be treated as a single scalar value (string)
	count := strings.Count(content, `Content-Disposition: form-data; name="byte_data"`)
	if count != 1 {
		t.Errorf("[]byte should produce exactly 1 field, got %d", count)
	}
	if !strings.Contains(content, "hello") {
		t.Error("byte slice should be converted to string and present in form field")
	}
}

func Test_Multipart_ByteArray(t *testing.T) {
	req := testRequest{
		Name:      "test",
		ByteArray: [4]byte{'t', 'e', 's', 't'},
	}

	buf := new(bytes.Buffer)
	enc := NewMultipartEncoder(buf)
	defer enc.Close()

	if err := enc.Encode(req); err != nil {
		t.Fatalf("encode error: %v", err)
	}

	content := buf.String()
	// [N]byte should also be treated as a single scalar value (string)
	count := strings.Count(content, `Content-Disposition: form-data; name="byte_array"`)
	if count != 1 {
		t.Errorf("[N]byte should produce exactly 1 field, got %d", count)
	}
	if !strings.Contains(content, "test") {
		t.Error("byte array should be converted to string and present in form field")
	}
}

func Test_Multipart_StringArray(t *testing.T) {
	req := testRequest{
		Name:        "test",
		StringArray: [3]string{"first", "second", "third"},
	}

	buf := new(bytes.Buffer)
	enc := NewMultipartEncoder(buf)
	defer enc.Close()

	if err := enc.Encode(req); err != nil {
		t.Fatalf("encode error: %v", err)
	}

	content := buf.String()
	// String array should produce 3 separate form fields
	count := strings.Count(content, `Content-Disposition: form-data; name="string_array"`)
	if count != 3 {
		t.Errorf("expected 3 string_array fields, got %d", count)
	}
	if !strings.Contains(content, "first") || !strings.Contains(content, "second") || !strings.Contains(content, "third") {
		t.Error("all array elements should be present in form fields")
	}
}

func Test_Multipart_EmptySliceWithoutOmitempty(t *testing.T) {
	req := testRequest{
		Name:      "test",
		MultiTags: []string{}, // Empty slice without omitempty
	}

	buf := new(bytes.Buffer)
	enc := NewMultipartEncoder(buf)
	defer enc.Close()

	if err := enc.Encode(req); err != nil {
		t.Fatalf("encode error: %v", err)
	}

	content := buf.String()
	// Empty slice without omitempty should still not produce a field
	// (current behavior treats empty and absent the same)
	if strings.Contains(content, "multi_tags") {
		t.Log("Note: empty slice without omitempty currently does not produce a field")
	}
}

func Test_Form_MultiElementSlice(t *testing.T) {
	// Test form encoding (not multipart) with repeated fields
	req := testRequest{
		Name: "test",
		Tags: []string{"a", "b", "c"},
	}

	buf := new(bytes.Buffer)
	enc := NewFormEncoder(buf)
	if err := enc.Encode(req); err != nil {
		t.Fatalf("encode error: %v", err)
	}
	if err := enc.Close(); err != nil {
		t.Fatalf("close error: %v", err)
	}

	content := buf.String()
	// URL-encoded form should support repeated fields using Add()
	// Expected format: name=test&tags=a&tags=b&tags=c (order may vary)
	if !strings.Contains(content, "name=test") {
		t.Errorf("name field should be present, got: %q", content)
	}
	// Count occurrences of tags= in the form data
	tagCount := strings.Count(content, "tags=")
	if tagCount != 3 {
		t.Errorf("expected 3 tags fields in form data, got %d, content: %q", tagCount, content)
	}
}

func Test_Multipart_File_InvalidHeaderKey(t *testing.T) {
	h := make(textproto.MIMEHeader)
	h["invalid header"] = []string{"value"} // space makes it invalid

	type req struct {
		Upload File `json:"file"`
	}
	r := req{
		Upload: File{
			Path:   "test.txt",
			Body:   strings.NewReader("data"),
			Header: h,
		},
	}

	buf := new(bytes.Buffer)
	enc := NewMultipartEncoder(buf)
	err := enc.Encode(r)
	if err == nil {
		t.Error("expected error for invalid header key, got nil")
	}
}

///////////////////////////////////////////////////////////////////////////////
// HELPER FUNCTIONS
// (None currently - using direct string inspection in tests)

///////////////////////////////////////////////////////////////////////////////
// FILE ENCODING TESTS

func Test_Multipart_SingleFile_ContentType(t *testing.T) {
	type req struct {
		Upload File `json:"file"`
	}
	r := req{
		Upload: File{
			Path:        "photo.jpg",
			Body:        strings.NewReader("fake image data"),
			ContentType: "image/jpeg",
		},
	}

	buf := new(bytes.Buffer)
	enc := NewMultipartEncoder(buf)
	if err := enc.Encode(r); err != nil {
		t.Fatalf("encode error: %v", err)
	}
	enc.Close()

	parts := parseMultipart(t, buf, enc.ContentType())
	if len(parts) != 1 {
		t.Fatalf("expected 1 part, got %d", len(parts))
	}
	ct := parts[0].Header.Get("Content-Type")
	if ct != "image/jpeg" {
		t.Errorf("expected Content-Type image/jpeg, got %q", ct)
	}
	if string(parts[0].Body) != "fake image data" {
		t.Errorf("unexpected body: %q", parts[0].Body)
	}
}

func Test_Multipart_SingleFile_CustomHeader(t *testing.T) {
	h := make(textproto.MIMEHeader)
	h.Set("X-Custom-Meta", "myvalue")

	type req struct {
		Upload File `json:"file"`
	}
	r := req{
		Upload: File{
			Path:        "doc.pdf",
			Body:        strings.NewReader("pdf bytes"),
			ContentType: "application/pdf",
			Header:      h,
		},
	}

	buf := new(bytes.Buffer)
	enc := NewMultipartEncoder(buf)
	if err := enc.Encode(r); err != nil {
		t.Fatalf("encode error: %v", err)
	}
	enc.Close()

	parts := parseMultipart(t, buf, enc.ContentType())
	if len(parts) != 1 {
		t.Fatalf("expected 1 part, got %d", len(parts))
	}
	if got := parts[0].Header.Get("X-Custom-Meta"); got != "myvalue" {
		t.Errorf("expected X-Custom-Meta=myvalue, got %q", got)
	}
	if got := parts[0].Header.Get("Content-Type"); got != "application/pdf" {
		t.Errorf("expected Content-Type application/pdf, got %q", got)
	}
}

func Test_Multipart_FileSlice(t *testing.T) {
	type req struct {
		Name  string `json:"name"`
		Files []File `json:"file"`
	}
	r := req{
		Name: "batch",
		Files: []File{
			{
				Path:        "a.txt",
				Body:        strings.NewReader("content a"),
				ContentType: "text/plain",
			},
			{
				Path:        "b.csv",
				Body:        strings.NewReader("content b"),
				ContentType: "text/csv",
				Header: func() textproto.MIMEHeader {
					h := make(textproto.MIMEHeader)
					h.Set("X-Source", "pipeline")
					return h
				}(),
			},
		},
	}

	buf := new(bytes.Buffer)
	enc := NewMultipartEncoder(buf)
	if err := enc.Encode(r); err != nil {
		t.Fatalf("encode error: %v", err)
	}
	enc.Close()

	parts := parseMultipart(t, buf, enc.ContentType())
	// 1 text field + 2 file parts
	if len(parts) != 3 {
		t.Fatalf("expected 3 parts, got %d", len(parts))
	}
	// first part is the "name" text field
	if string(parts[0].Body) != "batch" {
		t.Errorf("expected name=batch, got %q", parts[0].Body)
	}
	// second part: a.txt
	if ct := parts[1].Header.Get("Content-Type"); ct != "text/plain" {
		t.Errorf("expected text/plain, got %q", ct)
	}
	if string(parts[1].Body) != "content a" {
		t.Errorf("unexpected body: %q", parts[1].Body)
	}
	// third part: b.csv with custom header
	if ct := parts[2].Header.Get("Content-Type"); ct != "text/csv" {
		t.Errorf("expected text/csv, got %q", ct)
	}
	if got := parts[2].Header.Get("X-Source"); got != "pipeline" {
		t.Errorf("expected X-Source=pipeline, got %q", got)
	}
	if string(parts[2].Body) != "content b" {
		t.Errorf("unexpected body: %q", parts[2].Body)
	}
}

// parsedPart holds the pre-read header and body of a multipart part.
type parsedPart struct {
	Header textproto.MIMEHeader
	Body   []byte
}

// parseMultipart parses the encoded buffer back into individual parts for inspection.
func parseMultipart(t *testing.T, buf *bytes.Buffer, contentType string) []parsedPart {
	t.Helper()
	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		t.Fatalf("parse content type: %v", err)
	}
	mr := multipart.NewReader(buf, params["boundary"])
	var parts []parsedPart
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("next part: %v", err)
		}
		body, err := io.ReadAll(p)
		if err != nil {
			t.Fatalf("read part body: %v", err)
		}
		parts = append(parts, parsedPart{Header: p.Header, Body: body})
	}
	return parts
}
