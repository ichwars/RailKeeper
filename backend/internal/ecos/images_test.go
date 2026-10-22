package ecos

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"golang.org/x/image/bmp"
)

func testImageBMP() []byte {
	data := make([]byte, ImageBMPSize)
	copy(data, "BM")
	for offset, value := range map[int]uint32{2: ImageBMPSize, 10: 54, 14: 40, 18: 190, 22: 40, 26: 0x180001, 34: 22880} {
		binary.LittleEndian.PutUint32(data[offset:], value)
	}
	for y := 0; y < 40; y++ {
		for x := 0; x < 190; x++ {
			index := 54 + y*572 + x*3
			copy(data[index:], []byte{182, 182, 182})
		}
	}
	return data
}

func TestImageValidationAndPaletteVerification(t *testing.T) {
	data := testImageBMP()
	if err := ValidateImageBMP(data); err != nil {
		t.Fatal(err)
	}
	for _, bad := range [][]byte{nil, data[:100], append(bytes.Clone(data), 0)} {
		if ValidateImageBMP(bad) == nil {
			t.Fatal("accepted malformed BMP")
		}
	}
	compressed := bytes.Clone(data)
	compressed[30] = 1
	if ValidateImageBMP(compressed) == nil {
		t.Fatal("accepted compressed BMP")
	}
	palette := image.NewPaletted(image.Rect(0, 0, 190, 40), color.Palette{color.RGBA{182, 182, 182, 255}})
	var pngBytes bytes.Buffer
	_ = png.Encode(&pngBytes, palette)
	if !VerifyImage(data, pngBytes.Bytes()) {
		t.Fatal("palette PNG did not match")
	}
	pngBytes.Reset()
	_ = png.Encode(&pngBytes, image.NewRGBA(image.Rect(0, 0, 190, 40)))
	if VerifyImage(data, pngBytes.Bytes()) {
		t.Fatal("accepted transparent/different image")
	}
	if VerifyImage(data, []byte("<html>login</html>")) {
		t.Fatal("accepted HTML image")
	}
}

func TestImageVerificationAcceptsExactRGB332ConversionAndRejectsOneChangedPixel(t *testing.T) {
	palette := color.Palette{}
	for r := 0; r < 8; r++ {
		for g := 0; g < 8; g++ {
			for _, b := range []uint8{0, 73, 182, 255} {
				palette = append(palette, color.RGBA{uint8((r*255 + 3) / 7), uint8((g*255 + 3) / 7), b, 255})
			}
		}
	}
	source := image.NewRGBA(image.Rect(0, 0, 190, 40))
	device := image.NewPaletted(source.Bounds(), palette)
	for y := 0; y < 40; y++ {
		for x := 0; x < 190; x++ {
			r, g, b := uint8(x*17+y), uint8(x+y*23), uint8(x*3+y*19)
			source.SetRGBA(x, y, color.RGBA{r, g, b, 255})
			device.SetColorIndex(x, y, (r>>5)<<5|(g>>5)<<2|b>>6)
		}
	}
	var input, output bytes.Buffer
	_ = bmp.Encode(&input, source)
	_ = png.Encode(&output, device)
	if !VerifyImage(input.Bytes(), output.Bytes()) {
		t.Fatal("rejected exact RGB332 conversion")
	}
	device.SetColorIndex(37, 17, device.ColorIndexAt(37, 17)^1)
	output.Reset()
	_ = png.Encode(&output, device)
	if VerifyImage(input.Bytes(), output.Bytes()) {
		t.Fatal("accepted changed pixel")
	}
}

func imagePage(page int, occupied func(int) bool) string {
	var body strings.Builder
	for i := page * 50; i < (page+1)*50; i++ {
		if occupied(i) {
			fmt.Fprintf(&body, `<img src="/loco/image?type=user&amp;index=%d">`, i)
		} else {
			fmt.Fprintf(&body, `<a href="/fcgi/ecos-uploadimage?id=%d&lang=en&start=%d&items=50">Upload</a>`, i, page)
		}
	}
	body.WriteString(`<a href="/fcgi/ecos-locoimages?type=user&start=1&items=50">Next</a>`)
	return body.String()
}

func imageForm(index int) string {
	return fmt.Sprintf(`<form action="/fcgi/ecos-douploadimage" enctype="multipart/form-data" method="POST">
<input name="desc"><input name="type" value="m"><input name="uploaded_file" type="file">
<input name="lang" value="en"><input name="id" value="%d"><input name="start" value="%d"><input name="items" value="50"></form>`, index, index/50)
}

func TestImageUploadUsesOnlyKnownEndpointAndRechecksSlot(t *testing.T) {
	for _, occupied := range []bool{false, true} {
		t.Run(strconv.FormatBool(occupied), func(t *testing.T) {
			posts := 0
			gets := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/fcgi/ecos-locoimages":
					gets++
					_, _ = io.WriteString(w, imagePage(0, func(i int) bool { return i == 0 || occupied && i == 1 }))
				case "/fcgi/ecos-uploadimage":
					_, _ = io.WriteString(w, imageForm(1))
				case "/fcgi/ecos-douploadimage":
					if r.Method != "POST" {
						t.Error("not POST")
					}
					posts++
					if err := r.ParseMultipartForm(64 << 10); err != nil {
						t.Error(err)
						return
					}
					defer func() { _ = r.MultipartForm.RemoveAll() }()
					if r.FormValue("id") != "1" || r.FormValue("type") != "d" || r.FormValue("desc") != "BR 218" {
						t.Error("wrong form fields")
					}
					f, _, err := r.FormFile("uploaded_file")
					if err != nil {
						t.Error(err)
						return
					}
					defer func() { _ = f.Close() }()
					data, _ := io.ReadAll(f)
					if !bytes.Equal(data, testImageBMP()) {
						t.Error("changed uploaded image")
					}
				default:
					t.Errorf("unexpected request %s", r.URL)
					w.WriteHeader(404)
				}
			}))
			defer server.Close()
			web := &ImageWeb{base: server.URL, Host: "192.168.1.2", client: server.Client()}
			err := web.Upload(t.Context(), 1, "BR 218", "d", testImageBMP())
			if occupied {
				if !errors.Is(err, ErrImageSlotOccupied) || posts != 0 {
					t.Fatalf("occupied slot: err=%v posts=%d", err, posts)
				}
			} else if err != nil || posts != 1 {
				t.Fatalf("err=%v posts=%d", err, posts)
			}
			if gets != 1 {
				t.Errorf("slot rechecks=%d", gets)
			}
		})
	}
}

func TestImagePageRequiresCompleteUnambiguousInventory(t *testing.T) {
	page := imagePage(0, func(i int) bool { return i == 0 })
	states, _, err := parseImagePage([]byte(page), 0)
	if err != nil || !states[0] || states[1] {
		t.Fatalf("states=%v err=%v", states, err)
	}
	for _, bad := range []string{"<html>Login</html>", page + `<a href="/fcgi/ecos-uploadimage?id=0">x</a>`, strings.Replace(page, "id=1&", "id=500&", 1)} {
		if _, _, err := parseImagePage([]byte(bad), 0); err == nil {
			t.Fatal("accepted incomplete/duplicate/out-of-range inventory")
		}
	}
}

func TestImageFindFreeSlotFollowsOnlyBoundedPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page, _ := strconv.Atoi(r.URL.Query().Get("start"))
		_, _ = io.WriteString(w, imagePage(page, func(i int) bool { return i < 51 }))
	}))
	defer server.Close()
	web := &ImageWeb{base: server.URL, client: server.Client()}
	if slot, err := web.FindFreeSlot(t.Context()); err != nil || slot != 51 {
		t.Fatalf("slot=%d err=%v", slot, err)
	}
}

func TestImageWebRejectsNonLANTargets(t *testing.T) {
	for _, host := range []string{"127.0.0.1", "::1", "169.254.169.254", "8.8.8.8", "http://192.168.1.2/", "192.168.1.2@other", "192.168.1.2:80"} {
		if _, err := NewImageWeb(context.Background(), host); err == nil {
			t.Errorf("accepted %s", host)
		}
	}
}

func TestImageWebNeverFollowsRedirectsOrRetriesFailedUpload(t *testing.T) {
	for _, mode := range []string{"redirect", "changed_form", "lost_reply"} {
		t.Run(mode, func(t *testing.T) {
			posts, redirects := 0, 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/fcgi/ecos-locoimages":
					_, _ = io.WriteString(w, imagePage(0, func(int) bool { return false }))
				case "/fcgi/ecos-uploadimage":
					if mode == "redirect" {
						http.Redirect(w, r, "/login", http.StatusFound)
						return
					}
					form := imageForm(1)
					if mode == "changed_form" {
						form = strings.Replace(form, "ecos-douploadimage", "other-endpoint", 1)
					}
					_, _ = io.WriteString(w, form)
				case "/fcgi/ecos-douploadimage":
					posts++
					w.WriteHeader(http.StatusBadGateway)
				case "/login":
					redirects++
					_, _ = io.WriteString(w, "login")
				default:
					t.Errorf("unexpected path %s", r.URL)
				}
			}))
			defer server.Close()
			web, err := NewImageWeb(t.Context(), "192.168.1.2")
			if err != nil {
				t.Fatal(err)
			}
			web.base = server.URL
			if err := web.Upload(t.Context(), 1, "Test", "m", testImageBMP()); err == nil {
				t.Fatal("accepted incompatible/failed upload")
			}
			want := 0
			if mode == "lost_reply" {
				want = 1
			}
			if posts != want || redirects != 0 {
				t.Fatalf("posts=%d followed redirects=%d", posts, redirects)
			}
		})
	}
}

func TestImageWebReadsInternalPreviewAndRequiresStoredUploadSlot(t *testing.T) {
	var pngBytes bytes.Buffer
	_ = png.Encode(&pngBytes, image.NewRGBA(image.Rect(0, 0, 190, 40)))
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/loco/image":
			if r.URL.Query().Get("type") != "internal" || r.URL.Query().Get("index") != "118" {
				t.Errorf("wrong image query %s", r.URL)
			}
			_, _ = w.Write(pngBytes.Bytes())
		case "/fcgi/ecos-locoimages":
			_, _ = io.WriteString(w, imagePage(0, func(i int) bool { return i == 1 }))
		default:
			t.Errorf("unexpected path %s", r.URL)
		}
	}))
	defer server.Close()
	web := &ImageWeb{base: server.URL, client: server.Client()}
	if _, err := web.ReadImage(t.Context(), "system", 118); err != nil {
		t.Fatal(err)
	}
	if err := web.CheckStoredSlot(t.Context(), 1); err != nil {
		t.Fatal(err)
	}
	if err := web.CheckStoredSlot(t.Context(), 2); err == nil {
		t.Fatal("accepted an empty upload slot")
	}
}
