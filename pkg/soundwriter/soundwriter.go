package soundwriter

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"sync"

	// Packages
	"github.com/veandco/go-sdl2/sdl"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Reader struct {
	r        io.Reader
	mimetype string
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	firstBytes = 512
)

var (
	sdlinit sync.Once
)

///////////////////////////////////////////////////////////////////////////////
// CONSTRUCTOR

func NewReader(r io.Reader, opts ...Opt) (*Reader, error) {
	self := new(Reader)
	self.r = r

	// Read first 512 bytes
	buf := make([]byte, firstBytes)
	lr := io.LimitReader(r, firstBytes)
	if _, err := lr.Read(buf); !errors.Is(err, io.EOF) && err != nil {
		return nil, err
	}

	// Determine the mimetype
	mimetype := http.DetectContentType(buf)
	if mediatype, _, err := mime.ParseMediaType(mimetype); err != nil {
		return nil, err
	} else {
		self.mimetype = mediatype
	}

	// Apply the options
	for _, opt := range opts {
		if err := opt(self); err != nil {
			return nil, err
		}
	}

	// Initialise SDL
	var result error
	sdlinit.Do(func() {
		if err := sdl.Init(sdl.INIT_AUDIO); err != nil {
			result = err
		}
	})
	if result != nil {
		return nil, result
	}

	// Return success
	return self, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (r *Reader) MimeType() string {
	return r.mimetype
}
