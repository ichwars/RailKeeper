package application

import (
	"bytes"
	"compress/zlib"
	"context"
	"encoding/hex"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

func extractPDFText(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	parts := []string{}
	for _, stream := range pdfStreams(data) {
		if text := pdfContentText(stream); text != "" {
			parts = append(parts, text)
		}
	}
	text := strings.Join(parts, "\n")
	if text == "" {
		text = printablePDFText(data)
	}
	return normalizeWhitespacePreservingLines(repairMojibake(text))
}

func extractPDFTextWithOCRFallback(data []byte, articleNumber string) string {
	text := extractPDFText(data)
	if !needsPDFOCRFallback(text, articleNumber) {
		return text
	}
	if ocrText := pdfOCRTextExtractor(data); ocrText != "" {
		return ocrText
	}
	return text
}

func needsPDFOCRFallback(text, articleNumber string) bool {
	text = normalizeWhitespacePreservingLines(text)
	if text == "" || len([]rune(text)) < 40 {
		return true
	}
	if documentTextMatchesArticleNumber(text, articleNumber) && looksLikeSparePartsDocumentText(text, articleNumber) {
		return false
	}
	lower := strings.ToLower(text)
	return !containsAny(lower, []string{
		"ersatzteile", "ersatzteilliste", "spare parts", "pièces de rechange", "náhradní díly", "et-nr", "spare part n", "bezeichnung / description",
	})
}

func extractPDFOCRText(data []byte) string {
	if pdfOCRDisabled() || len(data) == 0 {
		return ""
	}
	tempDir, err := os.MkdirTemp("", "railkeeper-pdf-ocr-*")
	if err != nil {
		return ""
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	inputPath := filepath.Join(tempDir, "input.pdf")
	if err := os.WriteFile(inputPath, data, 0o600); err != nil {
		return ""
	}
	if text := extractPDFOCRTextWithOCRmyPDF(inputPath, tempDir); text != "" {
		return text
	}
	return extractPDFOCRTextWithTesseract(inputPath, tempDir)
}

func pdfOCRDisabled() bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv("RAILKEEPER_PDF_OCR")))
	return value == "0" || value == "false" || value == "off" || value == "disabled"
}

func extractPDFOCRTextWithOCRmyPDF(inputPath, tempDir string) string {
	ocrmypdf, err := exec.LookPath("ocrmypdf")
	if err != nil {
		return ""
	}
	outputPath := filepath.Join(tempDir, "ocr.pdf")
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, ocrmypdf, "--skip-text", "--optimize", "0", "--quiet", inputPath, outputPath)
	if err := cmd.Run(); err != nil {
		return ""
	}
	data, err := os.ReadFile(outputPath)
	if err != nil || len(data) == 0 {
		return ""
	}
	return extractPDFText(data)
}

func extractPDFOCRTextWithTesseract(inputPath, tempDir string) string {
	pdftoppm, err := exec.LookPath("pdftoppm")
	if err != nil {
		return ""
	}
	tesseract, err := exec.LookPath("tesseract")
	if err != nil {
		return ""
	}
	pageLimit := pdfOCRPageLimit()
	prefix := filepath.Join(tempDir, "page")
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	render := exec.CommandContext(ctx, pdftoppm, "-r", "200", "-png", "-f", "1", "-l", strconv.Itoa(pageLimit), inputPath, prefix)
	if err := render.Run(); err != nil {
		return ""
	}
	images, err := filepath.Glob(prefix + "-*.png")
	if err != nil || len(images) == 0 {
		return ""
	}
	sort.Strings(images)
	if len(images) > pageLimit {
		images = images[:pageLimit]
	}
	parts := []string{}
	for _, imagePath := range images {
		if text := runTesseractImageOCR(tesseract, imagePath); text != "" {
			parts = append(parts, text)
		}
	}
	return normalizeWhitespacePreservingLines(repairMojibake(strings.Join(parts, "\n")))
}

func pdfOCRPageLimit() int {
	value, err := strconv.Atoi(strings.TrimSpace(os.Getenv("RAILKEEPER_PDF_OCR_MAX_PAGES")))
	if err != nil || value <= 0 {
		return 4
	}
	if value > 12 {
		return 12
	}
	return value
}

func runTesseractImageOCR(tesseract, imagePath string) string {
	for _, language := range []string{"deu+eng", "eng", ""} {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		args := []string{imagePath, "stdout"}
		if language != "" {
			args = append(args, "-l", language)
		}
		args = append(args, "--psm", "6")
		output, err := exec.CommandContext(ctx, tesseract, args...).Output()
		cancel()
		if err != nil {
			continue
		}
		text := normalizeWhitespacePreservingLines(repairMojibake(string(output)))
		if text != "" {
			return text
		}
	}
	return ""
}

func pdfStreams(data []byte) [][]byte {
	streamPattern := regexp.MustCompile(`(?s)(<<.*?>>)\s*stream\r?\n(.*?)\r?\nendstream`)
	matches := streamPattern.FindAllSubmatch(data, -1)
	streams := [][]byte{}
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		dictionary := strings.ToLower(string(match[1]))
		stream := bytes.Trim(match[2], "\r\n")
		if strings.Contains(dictionary, "flatedecode") {
			reader, err := zlib.NewReader(bytes.NewReader(stream))
			if err != nil {
				continue
			}
			decoded, err := io.ReadAll(io.LimitReader(reader, 8*1024*1024))
			_ = reader.Close()
			if err == nil {
				streams = append(streams, decoded)
			}
			continue
		}
		streams = append(streams, stream)
	}
	return streams
}

func pdfContentText(data []byte) string {
	raw := string(data)
	if !strings.Contains(raw, "BT") || (!strings.Contains(raw, "Tj") && !strings.Contains(raw, "TJ")) {
		return ""
	}
	builder := strings.Builder{}
	newLine := func() {
		current := builder.String()
		if current != "" && !strings.HasSuffix(current, "\n") {
			builder.WriteByte('\n')
		}
	}
	space := func() {
		current := builder.String()
		if current != "" && !strings.HasSuffix(current, " ") && !strings.HasSuffix(current, "\n") {
			builder.WriteByte(' ')
		}
	}
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.Contains(line, " Tm") {
			newLine()
		}
		if match := pdfTextMovePattern.FindStringSubmatch(line); len(match) == 3 {
			x := parsePDFNumber(match[1])
			y := parsePDFNumber(match[2])
			if y < -0.2 || x < -15 {
				newLine()
			} else if x > 0.5 {
				space()
			}
		}
		text := pdfShowText(line)
		if text == "" {
			continue
		}
		if builder.Len() > 0 && !strings.HasSuffix(builder.String(), "\n") && needsPDFTextSpace(builder.String(), text) {
			builder.WriteByte(' ')
		}
		builder.WriteString(text)
	}
	return builder.String()
}

var pdfTextMovePattern = regexp.MustCompile(`(-?\d+(?:\.\d+)?)\s+(-?\d+(?:\.\d+)?)\s+Td\b`)

func parsePDFNumber(value string) float64 {
	var result float64
	var sign float64 = 1
	if strings.HasPrefix(value, "-") {
		sign = -1
		value = strings.TrimPrefix(value, "-")
	}
	parts := strings.SplitN(value, ".", 2)
	for _, char := range parts[0] {
		if char >= '0' && char <= '9' {
			result = result*10 + float64(char-'0')
		}
	}
	if len(parts) == 2 {
		scale := 10.0
		for _, char := range parts[1] {
			if char >= '0' && char <= '9' {
				result += float64(char-'0') / scale
				scale *= 10
			}
		}
	}
	return result * sign
}

func needsPDFTextSpace(current, next string) bool {
	last := []rune(strings.TrimRight(current, " \t"))
	first := []rune(strings.TrimLeft(next, " \t"))
	if len(last) == 0 || len(first) == 0 {
		return false
	}
	return isPDFWordRune(last[len(last)-1]) && isPDFWordRune(first[0])
}

func isPDFWordRune(char rune) bool {
	return char >= '0' && char <= '9' || char >= 'A' && char <= 'Z' || char >= 'a' && char <= 'z' || char >= 128
}

func pdfShowText(line string) string {
	if !strings.Contains(line, "Tj") && !strings.Contains(line, "TJ") {
		return ""
	}
	parts := []string{}
	for index := 0; index < len(line); index++ {
		switch line[index] {
		case '(':
			value, next := readPDFLiteral(line, index)
			if next > index {
				parts = append(parts, value)
				index = next - 1
			}
		case '<':
			if index+1 < len(line) && line[index+1] == '<' {
				continue
			}
			value, next := readPDFHex(line, index)
			if next > index {
				parts = append(parts, value)
				index = next - 1
			}
		}
	}
	return strings.Join(parts, "")
}

func readPDFLiteral(raw string, start int) (string, int) {
	depth := 1
	escaped := false
	for index := start + 1; index < len(raw); index++ {
		char := raw[index]
		if escaped {
			escaped = false
			continue
		}
		if char == '\\' {
			escaped = true
			continue
		}
		if char == '(' {
			depth++
			continue
		}
		if char == ')' {
			depth--
			if depth == 0 {
				return decodePDFLiteralString(raw[start+1 : index]), index + 1
			}
		}
	}
	return "", start
}

func readPDFHex(raw string, start int) (string, int) {
	end := strings.IndexByte(raw[start+1:], '>')
	if end < 0 {
		return "", start
	}
	end += start + 1
	cleaned := regexp.MustCompile(`\s+`).ReplaceAllString(raw[start+1:end], "")
	if len(cleaned) < 2 || len(cleaned)%2 == 1 {
		return "", end + 1
	}
	decoded, err := hex.DecodeString(cleaned)
	if err != nil || len(decoded) == 0 {
		return "", end + 1
	}
	return decodePDFHexText(decoded), end + 1
}

func decodePDFLiteralString(value string) string {
	decoded := []byte{}
	for index := 0; index < len(value); index++ {
		char := value[index]
		if char != '\\' || index+1 >= len(value) {
			decoded = append(decoded, char)
			continue
		}
		index++
		switch value[index] {
		case 'n', 'r', 't':
			decoded = append(decoded, ' ')
		case 'b', 'f':
			continue
		case '(', ')', '\\':
			decoded = append(decoded, value[index])
		default:
			if value[index] >= '0' && value[index] <= '7' {
				end := index + 1
				for end < len(value) && end-index < 3 && value[end] >= '0' && value[end] <= '7' {
					end++
				}
				octal := value[index:end]
				var octalByte byte
				for _, digit := range octal {
					octalByte = octalByte*8 + byte(digit-'0')
				}
				decoded = append(decoded, octalByte)
				index = end - 1
				continue
			}
			decoded = append(decoded, value[index])
		}
	}
	return decodePDFByteString(decoded)
}

func decodePDFHexText(data []byte) string {
	if len(data) >= 2 && data[0] == 0xFE && data[1] == 0xFF {
		builder := strings.Builder{}
		for index := 2; index+1 < len(data); index += 2 {
			r := rune(data[index])<<8 | rune(data[index+1])
			if r >= 32 {
				builder.WriteRune(r)
			}
		}
		return builder.String()
	}
	return decodePDFByteString(data)
}

func decodePDFByteString(data []byte) string {
	if utf8.Valid(data) {
		return string(data)
	}
	builder := strings.Builder{}
	for _, char := range data {
		if char == 0 {
			continue
		}
		builder.WriteRune(rune(char))
	}
	return builder.String()
}

func printablePDFText(data []byte) string {
	builder := strings.Builder{}
	for _, char := range string(data) {
		if char == '\n' || char == '\r' || char == '\t' || char >= 32 && char < utf8.RuneSelf {
			builder.WriteRune(char)
		} else {
			builder.WriteRune(' ')
		}
	}
	return builder.String()
}

func normalizeWhitespacePreservingLines(value string) string {
	lines := []string{}
	for _, line := range regexp.MustCompile(`[\n\r]+`).Split(value, -1) {
		line = normalizeWhitespace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}
