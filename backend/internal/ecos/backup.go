package ecos

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"errors"
	"io"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/crypto/openpgp"                  //nolint:staticcheck // Legacy ECoS OpenPGP format; bounded, integrity-checked read-only compatibility.
	pgperrors "golang.org/x/crypto/openpgp/errors" //nolint:staticcheck // Legacy ECoS OpenPGP format; bounded, integrity-checked read-only compatibility.
	"golang.org/x/crypto/openpgp/packet"           //nolint:staticcheck // Legacy ECoS OpenPGP format; bounded, integrity-checked read-only compatibility.
)

const (
	EncryptedBackupLimit = 8 << 20
	backupPlaintextLimit = 32 << 20
	backupMemberLimit    = 10000
	backupLocoLimit      = 64
	backupConfigLimit    = 1 << 20
)

var (
	ErrBackup    = errors.New("invalid ECoS backup")
	ErrBackupKey = errors.New("ECoS backup key unavailable")
	backupPath   = regexp.MustCompile(`^root/usrdata/lococonfig/([0-9]+)\.ecs$`)
)

type Backup struct {
	Serial                  uint32
	Locomotives             []BackupLocomotive
	StoredLocomotives       []BackupStoredLocomotive
	StoredLocomotivesStatus string
}

type BackupLocomotive struct {
	ObjectID int
	Values   []DatabaseCV
}

// DecodeBackup authenticates the complete OpenPGP body before parsing any archive.
// All archive contents remain in memory; no archive path is written to disk.
func DecodeBackup(reader io.Reader, keyring, passphrase []byte) (Backup, error) {
	data, err := readBackupBytes(reader, EncryptedBackupLimit)
	if err != nil || !backupHasIntegrity(data) {
		return Backup{}, ErrBackup
	}
	if len(keyring) == 0 || len(keyring) > 1<<20 || len(passphrase) > 1024 {
		return Backup{}, ErrBackupKey
	}
	keys, err := openpgp.ReadKeyRing(bytes.NewReader(keyring))
	if err != nil || len(keys) == 0 || len(keys) > 64 {
		return Backup{}, ErrBackupKey
	}
	prompted := false
	prompt := func(candidates []openpgp.Key, _ bool) ([]byte, error) {
		if prompted {
			return nil, ErrBackupKey
		}
		prompted = true
		unlocked := false
		for _, key := range candidates {
			if key.PrivateKey != nil && key.PrivateKey.Decrypt(passphrase) == nil {
				unlocked = true
			}
		}
		if !unlocked {
			return nil, ErrBackupKey
		}
		return nil, nil
	}
	message, err := openpgp.ReadMessage(bytes.NewReader(data), keys, prompt, nil)
	if err != nil {
		if errors.Is(err, ErrBackupKey) || errors.Is(err, pgperrors.ErrKeyIncorrect) {
			return Backup{}, ErrBackupKey
		}
		return Backup{}, ErrBackup
	}
	if !message.IsEncrypted || message.IsSymmetricallyEncrypted {
		return Backup{}, ErrBackup
	}
	// Reading until EOF triggers openpgp's MDC verification. The plaintext must
	// never reach an archive parser when that read fails or exceeds the limit.
	plaintext, err := readBackupBytes(message.UnverifiedBody, backupPlaintextLimit)
	if err != nil || message.SignatureError != nil {
		return Backup{}, ErrBackup
	}
	return decodeBackupArchive(plaintext)
}

func backupHasIntegrity(data []byte) bool {
	packets := packet.NewReader(bytes.NewReader(data))
	keys := 0
	for count := 0; count < 32; count++ {
		p, err := packets.Next()
		if err != nil {
			return false
		}
		switch p := p.(type) {
		case *packet.EncryptedKey:
			keys++
		case *packet.SymmetricallyEncrypted:
			return keys > 0 && p.MDC
		default:
			return false
		}
	}
	return false
}

func readBackupBytes(reader io.Reader, limit int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(reader, limit+1))
	if err != nil || int64(len(data)) > limit {
		return nil, ErrBackup
	}
	return data, nil
}

func decodeBackupArchive(data []byte) (Backup, error) {
	state := backupArchiveState{names: map[string]bool{}, objects: map[int]bool{}}
	switch {
	case bytes.HasPrefix(data, []byte{0x1f, 0x8b, 8}):
		input := bytes.NewReader(data)
		gz, err := gzip.NewReader(input)
		if err != nil {
			return Backup{}, ErrBackup
		}
		gz.Multistream(false)
		uncompressed, readErr := readBackupBytes(gz, backupPlaintextLimit)
		closeErr := gz.Close()
		if readErr != nil || closeErr != nil || input.Len() != 0 {
			return Backup{}, ErrBackup
		}
		if err := state.readTar(uncompressed); err != nil {
			return Backup{}, err
		}
	case bytes.HasPrefix(data, []byte("PK\x03\x04")):
		if err := state.readZip(data); err != nil {
			return Backup{}, err
		}
	default:
		return Backup{}, ErrBackup
	}
	if !state.serialFound || len(state.result.Locomotives) == 0 {
		return Backup{}, ErrBackup
	}
	sort.Slice(state.result.Locomotives, func(i, j int) bool {
		return state.result.Locomotives[i].ObjectID < state.result.Locomotives[j].ObjectID
	})
	return state.result, nil
}

type backupArchiveState struct {
	result      Backup
	names       map[string]bool
	objects     map[int]bool
	serialFound bool
	total       int64
	count       int
}

func (state *backupArchiveState) member(name string, size int64, directory bool, reader io.Reader) error {
	state.count++
	canonical := strings.TrimSuffix(name, "/")
	if state.count > backupMemberLimit || size < 0 || size > backupPlaintextLimit ||
		state.total > backupPlaintextLimit-size ||
		name == "" || strings.ContainsAny(name, "\\\x00") || path.IsAbs(name) ||
		path.Clean(canonical) != canonical || canonical == "." || canonical == ".." ||
		strings.HasPrefix(canonical, "../") || state.names[canonical] {
		return ErrBackup
	}
	state.total += size
	state.names[canonical] = true
	if directory {
		if size != 0 {
			return ErrBackup
		}
		return nil
	}
	if strings.HasSuffix(name, "/") {
		return ErrBackup
	}
	if name == "root/loks.bin" {
		if size < 4 {
			return ErrBackup
		}
		data, err := readBackupBytes(reader, backupPlaintextLimit)
		if err != nil || int64(len(data)) != size {
			return ErrBackup
		}
		state.result.Serial = binary.LittleEndian.Uint32(data[:4])
		state.serialFound = true
		if state.result.Serial == 0 {
			return ErrBackup
		}
		state.result.StoredLocomotivesStatus = BackupStoredLocomotivesUnavailable
		if locomotives, supported := decodeBackupStoredLocomotives(data); supported {
			state.result.StoredLocomotives = locomotives
			state.result.StoredLocomotivesStatus = BackupStoredLocomotivesAvailable
		}
	} else if match := backupPath.FindStringSubmatch(name); match != nil {
		objectID, err := strconv.Atoi(match[1])
		if err != nil || objectID < 1 || objectID > 2_000_000_000 || state.objects[objectID] ||
			len(state.result.Locomotives) >= backupLocoLimit || size > backupConfigLimit {
			return ErrBackup
		}
		data, err := readBackupBytes(reader, backupConfigLimit)
		if err != nil || int64(len(data)) != size {
			return ErrBackup
		}
		values, err := parseBackupConfig(data, objectID)
		if err != nil {
			return err
		}
		state.objects[objectID] = true
		state.result.Locomotives = append(state.result.Locomotives, BackupLocomotive{ObjectID: objectID, Values: values})
	}
	// Drain even ignored members to verify ZIP CRCs and truncated file bodies.
	if _, err := io.Copy(io.Discard, reader); err != nil {
		return ErrBackup
	}
	return nil
}

func (state *backupArchiveState) readTar(data []byte) error {
	input := bytes.NewReader(data)
	archive := tar.NewReader(input)
	for {
		member, err := archive.Next()
		if err == io.EOF {
			trailing := data[len(data)-input.Len():]
			if len(bytes.Trim(trailing, "\x00")) != 0 {
				return ErrBackup
			}
			return nil
		}
		if err != nil || (member.Typeflag != tar.TypeReg && member.Typeflag != tar.TypeDir) {
			return ErrBackup
		}
		if err := state.member(member.Name, member.Size, member.Typeflag == tar.TypeDir, archive); err != nil {
			return err
		}
	}
}

func (state *backupArchiveState) readZip(data []byte) error {
	archive, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil || len(archive.File) > backupMemberLimit {
		return ErrBackup
	}
	for _, file := range archive.File {
		if file.UncompressedSize64 > backupPlaintextLimit || (!file.Mode().IsRegular() && !file.FileInfo().IsDir()) {
			return ErrBackup
		}
		reader, err := file.Open()
		if err != nil {
			return ErrBackup
		}
		readErr := state.member(file.Name, int64(file.UncompressedSize64), file.FileInfo().IsDir(), reader)
		closeErr := reader.Close()
		if readErr != nil || closeErr != nil {
			return ErrBackup
		}
	}
	return nil
}

func parseBackupConfig(data []byte, objectID int) ([]DatabaseCV, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 128), 1024)
	if !scanner.Scan() {
		return nil, ErrBackup
	}
	header := strings.Fields(scanner.Text())
	if len(header) != 5 || header[0] != "lococonfig" || header[1] != "0" || header[2] != "1" {
		return nil, ErrBackup
	}
	id, err := parseBackupDecimal(header[3], 2_000_000_000)
	_, profileErr := parseBackupDecimal(header[4], 2_000_000_000)
	if err != nil || profileErr != nil || id != objectID {
		return nil, ErrBackup
	}
	values := make([]DatabaseCV, 0)
	seen := map[[2]int]bool{}
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), ":")
		if len(fields) != 3 || len(values) >= CVDatabaseLimit {
			return nil, ErrBackup
		}
		page, pageErr := parseBackupDecimal(fields[0], 65535)
		number, numErr := parseBackupDecimal(fields[1], 1024)
		value, valueErr := parseBackupDecimal(fields[2], 255)
		address := [2]int{page, number}
		if pageErr != nil || numErr != nil || valueErr != nil || number < 1 || seen[address] ||
			(page != 0 && (number < 257 || number > 512)) {
			return nil, ErrBackup
		}
		seen[address] = true
		cv := DatabaseCV{Number: number, Value: value}
		if number >= 257 && number <= 512 {
			high, low := page/256, page%256
			cv.IndexCV31, cv.IndexCV32 = &high, &low
		}
		values = append(values, cv)
	}
	if scanner.Err() != nil {
		return nil, ErrBackup
	}
	return values, nil
}

func parseBackupDecimal(text string, maximum uint64) (int, error) {
	if text == "" || strings.IndexFunc(text, func(r rune) bool { return r < '0' || r > '9' }) >= 0 {
		return 0, ErrBackup
	}
	value, err := strconv.ParseUint(text, 10, 32)
	if err != nil || value > maximum {
		return 0, ErrBackup
	}
	return int(value), nil
}
