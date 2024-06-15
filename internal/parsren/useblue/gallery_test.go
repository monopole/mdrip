package useblue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_attemptToParseGallery(t *testing.T) {
	const (
		galleryEntry = `:gallery
/img/image-0.png
/img/image-1.png`
		lf        = "\n"
		moreStuff = `stuff
stuff
and more stuff`
	)

	testCases := map[string]struct {
		data              []byte
		expectedRemainder []byte
		expectedSize      int
	}{
		"basic happy path": {
			data:              []byte(galleryEntry + lf + lf + moreStuff),
			expectedRemainder: nil,
			expectedSize:      len(galleryEntry),
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			n, b, s := attemptToParseGallery(tc.data)
			assert.Equal(t, tc.expectedRemainder, b)
			assert.Equal(t, tc.expectedSize, s)
			assert.NotNil(t, n)

			assert.Equal(t, 2, len(n.ImageURLS))

			assert.Equal(t, "/img/image-0.png", n.ImageURLS[0])
			assert.Equal(t, "/img/image-1.png", n.ImageURLS[1])
		})
	}
}
