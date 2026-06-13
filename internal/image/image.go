// Package image loads, resizes, and emits images to the terminal via the kitty / iTerm2 / half-block
// protocols, and is also used by the heading renderer to draw font-rasterized headings.
package image

import (
	"bytes"
	"fmt"
	"image"
	"image/png"

	// Register decoders for the image formats we read.
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"

	xdraw "golang.org/x/image/draw"
)

// Load decodes an image file. It returns an error if the file is unreadable or the format is
// unsupported.
func Load(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("unsupported or unreadable image: %s: %w", path, err)
	}
	return img, nil
}

// ResizeExact resizes to exactly w×h pixels with bilinear interpolation.
func ResizeExact(src image.Image, w, h int) *image.RGBA {
	w = max(1, w)
	h = max(1, h)
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	xdraw.BiLinear.Scale(dst, dst.Bounds(), src, src.Bounds(), xdraw.Over, nil)
	return dst
}

// ToPNG encodes an image to PNG bytes (for the kitty/iTerm2 protocols).
func ToPNG(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
