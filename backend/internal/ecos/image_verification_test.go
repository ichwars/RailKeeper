package ecos

import (
	"bytes"
	"encoding/hex"
	"image"
	"image/color"
	"image/png"
	"testing"

	"golang.org/x/image/bmp"
)

func ditherFixture(t *testing.T) ([]byte, *image.Paletted) {
	t.Helper()
	palette := color.Palette{}
	for _, r := range []uint8{0, 36, 73, 109, 146, 182, 219, 255} {
		for _, g := range []uint8{0, 36, 73, 109, 146, 182, 219, 255} {
			for _, b := range []uint8{0, 73, 182, 255} {
				palette = append(palette, color.RGBA{r, g, b, 255})
			}
		}
	}
	source := image.NewRGBA(image.Rect(0, 0, 190, 40))
	device := image.NewPaletted(source.Bounds(), palette)
	for y := 0; y < 40; y++ {
		for x := 0; x < 190; x++ {
			source.SetRGBA(x, y, color.RGBA{182, 182, 182, 255})
			device.SetColorIndex(x, y, 182)
		}
	}
	// Synthetic colour fields touch both image boundaries and include a grey gap,
	// near-grey, tie distances and clamped extremes. The fixed expected indices
	// come from a separate reference calculation, not VerifyImage's implementation.
	rows := []string{
		"00ff00b6b6fa21658aaccd32", "31567ab9b606494d8fb1f53b",
		"397d83a4b62e5171b7d81e26", "41468aadb63355b6bec5054a",
		"496eb3d5b63a7d86a6c92e52", "5196b6d9b6476989cff1355b",
		"759aa3c5b64b6db2b7195a62", "6186a8edb65395bade21466b",
	}
	for _, origin := range []image.Point{{0, 0}, {178, 32}} {
		for y, row := range rows {
			indices, err := hex.DecodeString(row)
			if err != nil || len(indices) != 12 {
				t.Fatal("invalid reference indices")
			}
			for x, index := range indices {
				pixel := color.RGBA{uint8((x*47 + y*13 + 17) % 256),
					uint8((x*29 + y*61 + 91) % 256), uint8((x*83 + y*7 + 43) % 256), 255}
				if x == 4 || x == 7 && y == 3 {
					pixel = color.RGBA{182, 182, 182, 255}
				}
				if y == 0 && x < 4 {
					pixel = []color.RGBA{{18, 18, 36, 255}, {255, 255, 255, 255},
						{0, 0, 0, 255}, {181, 182, 183, 255}}[x]
				}
				source.SetRGBA(origin.X+x, origin.Y+y, pixel)
				device.SetColorIndex(origin.X+x, origin.Y+y, index)
			}
		}
	}
	var data bytes.Buffer
	if err := bmp.Encode(&data, source); err != nil {
		t.Fatal(err)
	}
	return data.Bytes(), device
}

func TestImageVerificationMatchesIntegerDiffusionWithBackgroundErrorReset(t *testing.T) {
	source, device := ditherFixture(t)
	var data bytes.Buffer
	if err := png.Encode(&data, device); err != nil {
		t.Fatal(err)
	}
	if !VerifyImage(source, data.Bytes()) {
		t.Fatal("rejected exact reference dither conversion")
	}
}

func TestImageVerificationRejectsModifiedOrDisplacedDitheredImage(t *testing.T) {
	for _, name := range []string{"foreground pixel", "background pixel", "displaced", "transparent"} {
		t.Run(name, func(t *testing.T) {
			source, device := ditherFixture(t)
			switch name {
			case "foreground pixel":
				device.SetColorIndex(6, 2, device.ColorIndexAt(6, 2)^1)
			case "background pixel":
				device.SetColorIndex(100, 20, 0)
			case "displaced":
				copy(device.Pix[1:], bytes.Clone(device.Pix[:len(device.Pix)-1]))
			case "transparent":
				device.Palette[0] = color.RGBA{0, 0, 0, 0}
			}
			var data bytes.Buffer
			if err := png.Encode(&data, device); err != nil {
				t.Fatal(err)
			}
			if VerifyImage(source, data.Bytes()) {
				t.Fatal("accepted a changed dithered image")
			}
		})
	}
}
