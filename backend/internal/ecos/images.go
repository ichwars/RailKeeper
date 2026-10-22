package ecos

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"html"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/textproto"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/image/bmp"
)

const ImageBMPSize = 22934 // 54-byte header + 572-byte rows, 190 × 40 RGB24.

var (
	ErrImageFormat       = errors.New("invalid ECoS image format")
	ErrImageWeb          = errors.New("unsupported or unreachable ECoS image web interface")
	ErrImageRead         = errors.New("ECoS image could not be read")
	ErrImageInventory    = errors.New("ECoS image inventory could not be checked")
	ErrImageUploadForm   = errors.New("ECoS image upload form could not be checked")
	ErrImageSlotOccupied = errors.New("ECoS image slot is occupied")
	ErrImageNotStored    = errors.New("ECoS image slot is still free after upload")
	ErrImageSlotsFull    = errors.New("ECoS image slots are full")
	imageLinks           = regexp.MustCompile(`(?i)\b(?:href|src)\s*=\s*["']([^"']+)["']`)
	imageInput           = regexp.MustCompile(`(?is)<input\b[^>]*>`)
	imageAttribute       = regexp.MustCompile(`(?i)\b([a-z_]+)\s*=\s*["']([^"']*)["']`)
)

// ImageWeb only talks to a pinned, configured LAN address on port 80. It never
// follows redirects, device-supplied URLs, proxies, or arbitrary form actions.
type ImageWeb struct {
	Host   string
	base   string
	client *http.Client
}

func (web *ImageWeb) Address() string { return web.Host }

func NewImageWeb(ctx context.Context, host string) (*ImageWeb, error) {
	if strings.ContainsAny(host, "/\\@?#:") && net.ParseIP(host) == nil {
		return nil, ErrImageWeb
	}
	lookup, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	addresses, err := net.DefaultResolver.LookupIPAddr(lookup, host)
	if err != nil || len(addresses) == 0 {
		return nil, ErrImageWeb
	}
	for _, address := range addresses {
		if !address.IP.IsPrivate() || address.IP.IsLoopback() || address.Zone != "" {
			return nil, ErrImageWeb
		}
	}
	ip := addresses[0].IP.String()
	transport := &http.Transport{Proxy: nil, DisableKeepAlives: true,
		DialContext:           (&net.Dialer{Timeout: 4 * time.Second}).DialContext,
		ResponseHeaderTimeout: 5 * time.Second}
	return &ImageWeb{Host: ip, base: "http://" + net.JoinHostPort(ip, "80"), client: &http.Client{
		Transport: transport, Timeout: 8 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}}, nil
}

func ValidateImageBMP(data []byte) error {
	// Exact Windows BITMAPINFOHEADER, positive height, RGB24, BI_RGB and pixel offset.
	if len(data) != ImageBMPSize || string(data[:2]) != "BM" || binary.LittleEndian.Uint32(data[2:6]) != ImageBMPSize {
		return ErrImageFormat
	}
	expected := map[int][]byte{10: {54, 0, 0, 0}, 14: {40, 0, 0, 0}, 18: {190, 0, 0, 0},
		22: {40, 0, 0, 0}, 26: {1, 0, 24, 0}, 30: {0, 0, 0, 0}}
	for offset, value := range expected {
		if !bytes.Equal(data[offset:offset+len(value)], value) {
			return ErrImageFormat
		}
	}
	_, err := bmp.Decode(bytes.NewReader(data))
	if err != nil {
		return ErrImageFormat
	}
	return nil
}

func (web *ImageWeb) request(ctx context.Context, method, path, contentType string, body io.Reader, limit int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, web.base+path, body)
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	req.Header.Set("Cache-Control", "no-cache")
	response, err := web.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrImageWeb, err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return nil, ErrImageWeb
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, limit+1))
	if err != nil || int64(len(data)) > limit {
		return nil, ErrImageWeb
	}
	return data, nil
}

func imageListPath(page int) string {
	return fmt.Sprintf("/fcgi/ecos-locoimages?type=user&lang=en&start=%d&items=50", page)
}

func parseImagePage(data []byte, page int) (map[int]bool, map[int]bool, error) {
	states, pages := map[int]bool{}, map[int]bool{page: true}
	for _, match := range imageLinks.FindAllSubmatch(data, -1) {
		link, err := url.Parse(html.UnescapeString(string(match[1])))
		if err != nil || link.IsAbs() || link.Host != "" {
			continue
		}
		query := link.Query()
		if link.Path == "/fcgi/ecos-locoimages" && query.Get("type") == "user" && query.Get("items") == "50" {
			p, err := strconv.Atoi(query.Get("start"))
			if err == nil && p >= 0 && p < 10 {
				pages[p] = true
			}
		}
		var raw string
		var occupied bool
		switch {
		case link.Path == "/loco/image" && query.Get("type") == "user":
			raw = query.Get("index")
			occupied = true
		case link.Path == "/fcgi/ecos-uploadimage":
			raw = query.Get("id")
		default:
			continue
		}
		index, err := strconv.Atoi(raw)
		if err != nil || index < page*50 || index >= (page+1)*50 {
			return nil, nil, ErrImageWeb
		}
		if _, exists := states[index]; exists {
			return nil, nil, ErrImageWeb
		}
		states[index] = occupied
	}
	// This protocol adapter supports the documented ten/50-slot page layout. A
	// login page or changed markup must never be interpreted as an empty slot.
	if len(states) != 50 {
		return nil, nil, ErrImageWeb
	}
	return states, pages, nil
}

func (web *ImageWeb) FindFreeSlot(ctx context.Context) (int, error) {
	pages := map[int]bool{0: true}
	for page := 0; page < 10; page++ {
		if !pages[page] {
			continue
		}
		data, err := web.request(ctx, http.MethodGet, imageListPath(page), "", nil, 256<<10)
		if err != nil {
			return 0, err
		}
		states, available, err := parseImagePage(data, page)
		if err != nil {
			return 0, err
		}
		for p := range available {
			pages[p] = true
		}
		for index := page * 50; index < (page+1)*50; index++ {
			if !states[index] {
				return index, nil
			}
		}
	}
	return 0, ErrImageSlotsFull
}

func (web *ImageWeb) CheckFreeSlot(ctx context.Context, index int) error {
	return web.checkSlot(ctx, index, false)
}

func (web *ImageWeb) CheckStoredSlot(ctx context.Context, index int) error {
	return web.checkSlot(ctx, index, true)
}

func (web *ImageWeb) checkSlot(ctx context.Context, index int, occupied bool) (err error) {
	defer func() {
		if err != nil && !errors.Is(err, ErrImageSlotOccupied) && !errors.Is(err, ErrImageNotStored) {
			err = fmt.Errorf("%w: slot %d: %w", ErrImageInventory, index, err)
		}
	}()
	if index < 0 || index > 499 {
		return ErrImageWeb
	}
	data, err := web.request(ctx, http.MethodGet, imageListPath(index/50), "", nil, 256<<10)
	if err != nil {
		return err
	}
	states, _, err := parseImagePage(data, index/50)
	if err != nil {
		return err
	}
	if !occupied && states[index] {
		return ErrImageSlotOccupied
	}
	if occupied && !states[index] {
		return ErrImageNotStored
	}
	return nil
}

func (web *ImageWeb) ReadImage(ctx context.Context, kind string, index int) (result []byte, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("%w: %s slot %d: %w", ErrImageRead, kind, index, err)
		}
	}()
	if (kind != "user" && kind != "system") || index < 0 || index > 65535 {
		return nil, ErrImageWeb
	}
	// The inventory calls these "system", but the image endpoint uses "internal".
	if kind == "system" {
		kind = "internal"
	}
	data, err := web.request(ctx, http.MethodGet, fmt.Sprintf("/loco/image?type=%s&index=%d", kind, index), "", nil, 128<<10)
	if err != nil {
		return nil, err
	}
	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || config.Width != 190 || config.Height != 40 {
		return nil, ErrImageWeb
	}
	decoded, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, ErrImageWeb
	}
	var normalized bytes.Buffer
	if err := png.Encode(&normalized, decoded); err != nil {
		return nil, ErrImageWeb
	}
	return normalized.Bytes(), nil
}

func (web *ImageWeb) CheckUploadForm(ctx context.Context, index int) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("%w: slot %d: %w", ErrImageUploadForm, index, err)
		}
	}()
	if index < 0 || index > 499 {
		return ErrImageWeb
	}
	// Check the actual form before writing; do not guess endpoints on a different firmware.
	form, err := web.request(ctx, http.MethodGet, fmt.Sprintf("/fcgi/ecos-uploadimage?id=%d&lang=en&start=%d&items=50", index, index/50), "", nil, 64<<10)
	if err != nil {
		return err
	}
	if !bytes.Contains(form, []byte(`action="/fcgi/ecos-douploadimage"`)) || !bytes.Contains(form, []byte(`enctype="multipart/form-data"`)) || !bytes.Contains(form, []byte(`method="POST"`)) {
		return ErrImageWeb
	}
	fields := map[string]string{}
	for _, input := range imageInput.FindAll(form, -1) {
		attributes := map[string]string{}
		for _, attr := range imageAttribute.FindAllSubmatch(input, -1) {
			attributes[strings.ToLower(string(attr[1]))] = html.UnescapeString(string(attr[2]))
		}
		fields[attributes["name"]] = attributes["value"]
	}
	for _, name := range []string{"desc", "uploaded_file", "type", "lang", "id", "start", "items"} {
		if _, ok := fields[name]; !ok {
			return ErrImageWeb
		}
	}
	if fields["id"] != strconv.Itoa(index) || fields["start"] != strconv.Itoa(index/50) || fields["items"] != "50" {
		return ErrImageWeb
	}
	return nil
}

func (web *ImageWeb) Upload(ctx context.Context, index int, description, kind string, data []byte) error {
	if err := ValidateImageBMP(data); err != nil {
		return err
	}
	if index < 0 || index > 499 || !strings.Contains("sdem", kind) || len(kind) != 1 {
		return ErrImageFormat
	}
	if err := web.CheckUploadForm(ctx, index); err != nil {
		return err
	}
	// Recheck immediately before POST. No occupied slot is intentionally replaced.
	if err := web.CheckFreeSlot(ctx, index); err != nil {
		return err
	}
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	// Follow the station's HTML form order, including the file between the radio
	// selection and hidden fields. Older embedded form parsers can be order-sensitive.
	for _, field := range [][2]string{{"desc", description}, {"type", kind}} {
		if err := writer.WriteField(field[0], field[1]); err != nil {
			return err
		}
	}
	// CreateFormFile uses application/octet-stream. Send the actual BMP media type,
	// as a browser does when submitting the ECoS upload form.
	part, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Disposition": {`form-data; name="uploaded_file"; filename="railkeeper.bmp"`},
		"Content-Type":        {"image/bmp"},
	})
	if err != nil {
		return err
	}
	if _, err = part.Write(data); err != nil {
		return err
	}
	for _, field := range [][2]string{{"lang", "en"}, {"id", strconv.Itoa(index)},
		{"start", strconv.Itoa(index / 50)}, {"items", "50"}, {"submit", "Submit"}} {
		if err := writer.WriteField(field[0], field[1]); err != nil {
			return err
		}
	}
	if err = writer.Close(); err != nil {
		return err
	}
	_, err = web.request(ctx, http.MethodPost, "/fcgi/ecos-douploadimage", writer.FormDataContentType(), &body, 256<<10)
	return err // A missing reply is ambiguous. The caller must never retry this POST.
}
