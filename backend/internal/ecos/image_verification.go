package ecos

import (
	"bytes"
	"image"
	"sort"

	"golang.org/x/image/bmp"
)

// VerifyImage requires an exact pixel match after a supported ECoS palette
// conversion. It never accepts a merely similar or displaced image.
func VerifyImage(sourceBMP, deviceImage []byte) bool {
	if ValidateImageBMP(sourceBMP) != nil {
		return false
	}
	want, err := bmp.Decode(bytes.NewReader(sourceBMP))
	if err != nil {
		return false
	}
	config, _, err := image.DecodeConfig(bytes.NewReader(deviceImage))
	if err != nil || config.Width != 190 || config.Height != 40 {
		return false
	}
	got, _, err := image.Decode(bytes.NewReader(deviceImage))
	if err != nil {
		return false
	}
	if equalImagePixels(want, got, nil) {
		return true
	}
	paletted, ok := got.(*image.Paletted)
	if !ok || len(paletted.Palette) != 256 {
		return false
	}
	channels := [3]map[uint32]bool{{}, {}, {}}
	colors := map[[3]uint32]bool{}
	for _, color := range paletted.Palette {
		r, g, b, a := color.RGBA()
		if a != 65535 {
			return false
		}
		channels[0][r] = true
		channels[1][g] = true
		channels[2][b] = true
		colors[[3]uint32{r, g, b}] = true
	}
	if len(channels[0]) != 8 || len(channels[1]) != 8 || len(channels[2]) != 4 || len(colors) != 256 {
		return false
	}
	levels := [3][]uint32{}
	for channel, values := range channels {
		for value := range values {
			levels[channel] = append(levels[channel], value)
		}
		sort.Slice(levels[channel], func(i, j int) bool { return levels[channel][i] < levels[channel][j] })
		if levels[channel][0] != 0 || levels[channel][len(levels[channel])-1] != 65535 {
			return false
		}
	}
	for _, nearest := range []bool{false, true} {
		convert := func(r, g, b uint32) (uint32, uint32, uint32) {
			values := [3]uint32{r, g, b}
			for channel, value := range values {
				index := int(value >> 13)
				if channel == 2 {
					index = int(value >> 14)
				}
				if nearest {
					index = 0
					distance := uint32(1 << 17)
					for candidate, level := range levels[channel] {
						var delta uint32
						if value > level {
							delta = value - level
						} else {
							delta = level - value
						}
						if delta < distance {
							index = candidate
							distance = delta
						}
					}
				}
				values[channel] = levels[channel][index]
			}
			return values[0], values[1], values[2]
		}
		if equalImagePixels(want, got, convert) {
			return true
		}
	}
	return verifyECoSDitheredImage(want, got, levels)
}

func equalImagePixels(want, got image.Image, convert func(uint32, uint32, uint32) (uint32, uint32, uint32)) bool {
	for y := 0; y < 40; y++ {
		for x := 0; x < 190; x++ {
			r, g, b, _ := want.At(x, y).RGBA()
			if convert != nil {
				r, g, b = convert(r, g, b)
			}
			rr, gg, bb, a := got.At(x, y).RGBA()
			if a != 65535 || r != rr || g != gg || b != bb {
				return false
			}
		}
	}
	return true
}

// Firmware 4.3.3 applies integer Floyd–Steinberg diffusion in top-down, left-to-right
// order. Exact RGB(182,182,182) background pixels discard incoming error and do not
// diffuse it. The conversion was checked against every pixel of two device reads
// and the original upload, whose SHA-256 matched the audit record.
func verifyECoSDitheredImage(want, got image.Image, paletteLevels [3][]uint32) bool {
	levels := [3][]int{{0, 36, 73, 109, 146, 182, 219, 255},
		{0, 36, 73, 109, 146, 182, 219, 255}, {0, 73, 182, 255}}
	for channel, values := range levels {
		if len(paletteLevels[channel]) != len(values) {
			return false
		}
		for i, value := range values {
			if paletteLevels[channel][i] != uint32(value*257) {
				return false
			}
		}
	}
	// Padding cells receive error falling outside the image and are never read.
	var current, next [192][3]int
	for y := 0; y < 40; y++ {
		for x := 0; x < 190; x++ {
			r, g, b, _ := want.At(x, y).RGBA()
			source := [3]int{int(r >> 8), int(g >> 8), int(b >> 8)}
			expected := source
			if source != [3]int{182, 182, 182} {
				for channel, value := range source {
					// Arithmetic shift rounds negative accumulated error down, unlike
					// Go's signed division. Keep the station's integer rounding.
					adjusted := max(0, min(255, value+(current[x+1][channel]>>4)))
					quantized := levels[channel][0]
					distance := 256
					for _, candidate := range levels[channel] {
						delta := adjusted - candidate
						if delta < 0 {
							delta = -delta
						}
						if delta < distance {
							quantized, distance = candidate, delta
						}
					}
					expected[channel] = quantized
					difference := adjusted - quantized
					current[x+2][channel] += difference * 7
					next[x][channel] += difference * 3
					next[x+1][channel] += difference * 5
					next[x+2][channel] += difference
				}
			}
			rr, gg, bb, alpha := got.At(x, y).RGBA()
			if alpha != 65535 || rr != uint32(expected[0]*257) || gg != uint32(expected[1]*257) || bb != uint32(expected[2]*257) {
				return false
			}
		}
		current, next = next, [192][3]int{}
	}
	return true
}
