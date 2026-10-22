package ecos

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestImageSlotSearchReusesIdenticalImageBeyondFreeSlotOnAnotherPage(t *testing.T) {
	for _, mode := range []string{"match", "different", "unreadable", "incomplete_inventory"} {
		t.Run(mode, func(t *testing.T) {
			img := image.NewRGBA(image.Rect(0, 0, 190, 40))
			for y := 0; y < 40; y++ {
				for x := 0; x < 190; x++ {
					img.SetRGBA(x, y, color.RGBA{182, 182, 182, 255})
				}
			}
			if mode == "different" {
				img.SetRGBA(0, 0, color.RGBA{0, 0, 0, 255})
			}
			var data bytes.Buffer
			_ = png.Encode(&data, img)
			reads, posts := 0, 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					posts++
					t.Error("wrote during slot search")
				}
				switch r.URL.Path {
				case "/fcgi/ecos-locoimages":
					page, _ := strconv.Atoi(r.URL.Query().Get("start"))
					if mode == "incomplete_inventory" && page == 1 {
						_, _ = io.WriteString(w, "<html>Login</html>")
						return
					}
					_, _ = io.WriteString(w, imagePage(page, func(i int) bool { return i == 57 }))
				case "/loco/image":
					reads++
					if r.URL.Query().Get("index") != "57" {
						t.Error("unoccupied image requested")
					}
					if mode == "unreadable" {
						w.WriteHeader(500)
						return
					}
					_, _ = w.Write(data.Bytes())
				default:
					t.Error("unexpected URL", r.URL)
					w.WriteHeader(404)
				}
			}))
			defer server.Close()
			web := &ImageWeb{base: server.URL, client: server.Client()}
			slot, existing, err := web.FindImageSlot(t.Context(), testImageBMP())
			switch mode {
			case "match":
				if err != nil || slot != 57 || !existing {
					t.Fatal(slot, existing, err)
				}
			case "different":
				if err != nil || slot != 0 || existing {
					t.Fatal(slot, existing, err)
				}
			default:
				if !errors.Is(err, ErrImageWeb) {
					t.Fatal("allowed upload without complete image check", err)
				}
			}
			if posts != 0 || mode != "incomplete_inventory" && reads != 1 {
				t.Fatal("invalid device reads/writes", reads, posts)
			}
		})
	}
}
