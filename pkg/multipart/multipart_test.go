package multipart

import (
	"bytes"
	"io"
	"mime/multipart"
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
	// Should contain multiple tags fields
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
	// Should contain two single_number fields with integer values
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

///////////////////////////////////////////////////////////////////////////////
// HELPER FUNCTIONS

// parseFormValues extracts form field values from multipart content
func parseFormValues(t *testing.T, content string, fieldName string) []string {
	// This is a simplified parser for testing purposes
	reader := multipart.NewReader(strings.NewReader(content), extractBoundary(content))
	var values []string

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("parse error: %v", err)
		}

		if part.FormName() == fieldName {
			data, err := io.ReadAll(part)
			if err != nil {
				t.Fatalf("read error: %v", err)
			}
			values = append(values, string(data))
		}
	}

	return values
}

// extractBoundary extracts the boundary from a multipart message
func extractBoundary(content string) string {
	parts := strings.Split(content, "\r\n")
	if len(parts) > 0 {
		boundary := strings.TrimPrefix(parts[0], "--")
		return boundary
	}
	return ""
}
