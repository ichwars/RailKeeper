package ecos

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto"
	_ "crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"

	"golang.org/x/crypto/openpgp"        //nolint:staticcheck // Legacy ECoS OpenPGP format; bounded, integrity-checked read-only compatibility.
	"golang.org/x/crypto/openpgp/packet" //nolint:staticcheck // Legacy ECoS OpenPGP format; bounded, integrity-checked read-only compatibility.
)

var backupFixture struct {
	sync.Once
	entity *openpgp.Entity
	ring   []byte
	err    error
}

func backupTestKeys(t *testing.T) (*openpgp.Entity, []byte) {
	t.Helper()
	backupFixture.Do(func() {
		backupFixture.entity, backupFixture.err = openpgp.NewEntity("Synthetic backup test", "", "fixture@invalid.example", &packet.Config{RSABits: 1024, DefaultHash: crypto.SHA256})
		if backupFixture.err == nil {
			var ring bytes.Buffer
			backupFixture.err = backupFixture.entity.SerializePrivate(&ring, nil)
			backupFixture.ring = ring.Bytes()
		}
	})
	if backupFixture.err != nil {
		t.Fatal(backupFixture.err)
	}
	return backupFixture.entity, backupFixture.ring
}

func encryptBackupFixture(t *testing.T, body []byte) []byte {
	t.Helper()
	entity, _ := backupTestKeys(t)
	var encrypted bytes.Buffer
	writer, err := openpgp.Encrypt(&encrypted, []*openpgp.Entity{entity}, nil, nil, &packet.Config{DefaultCipher: packet.CipherAES128, DefaultHash: crypto.SHA256})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Write(body); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return encrypted.Bytes()
}

func encryptCompressedBackupFixture(t *testing.T, body []byte) []byte {
	t.Helper()
	entity, _ := backupTestKeys(t)
	var encrypted bytes.Buffer
	key := bytes.Repeat([]byte{0x25}, 16) // Synthetic session key, never used outside tests.
	if err := packet.SerializeEncryptedKey(&encrypted, entity.Subkeys[0].PublicKey, packet.CipherAES128, key, nil); err != nil {
		t.Fatal(err)
	}
	ciphertext, err := packet.SerializeSymmetricallyEncrypted(&encrypted, packet.CipherAES128, key, nil)
	if err != nil {
		t.Fatal(err)
	}
	compressed, err := packet.SerializeCompressed(ciphertext, packet.CompressionZLIB, nil)
	if err != nil {
		t.Fatal(err)
	}
	literal, err := packet.SerializeLiteral(compressed, true, "", 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := literal.Write(body); err != nil {
		t.Fatal(err)
	}
	if err := literal.Close(); err != nil {
		t.Fatal(err)
	}
	return encrypted.Bytes()
}

type backupTestMember struct {
	name string
	data []byte
	kind byte
}

func backupSerialFixture() []byte {
	serial := make([]byte, 4)
	binary.LittleEndian.PutUint32(serial, 0x12345678)
	return serial
}

func backupConfigFixture(id int, entries string) []byte {
	return []byte(fmt.Sprintf("lococonfig 0 1 %d 29\n%s", id, entries))
}

func gzipBackupFixture(t *testing.T, data []byte) []byte {
	t.Helper()
	var result bytes.Buffer
	writer := gzip.NewWriter(&result)
	if _, err := writer.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return result.Bytes()
}

func tarBackupFixture(t *testing.T, members []backupTestMember) []byte {
	t.Helper()
	var result bytes.Buffer
	writer := tar.NewWriter(&result)
	for _, member := range members {
		kind := member.kind
		if kind == 0 {
			kind = tar.TypeReg
		}
		header := &tar.Header{Name: member.name, Mode: 0o600, Size: int64(len(member.data)), Typeflag: kind}
		if kind == tar.TypeSymlink || kind == tar.TypeLink {
			header.Linkname = "root/loks.bin"
		}
		if err := writer.WriteHeader(header); err != nil {
			t.Fatal(err)
		}
		if len(member.data) > 0 {
			if _, err := writer.Write(member.data); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return gzipBackupFixture(t, result.Bytes())
}

func validBackupMembers() []backupTestMember {
	return []backupTestMember{
		{name: "root/loks.bin", data: backupSerialFixture()},
		{name: "root/usrdata/lococonfig/01002.ecs", data: backupConfigFixture(1002, "0:1:3\n0:257:0\n4099:257:20\n4104:257:1\n0:700:4\n")},
	}
}

func TestBackupDecryptsAuthenticatedArchiveAndPreservesRegisterIdentity(t *testing.T) {
	_, ring := backupTestKeys(t)
	members := append(validBackupMembers(), backupTestMember{name: "root/usrdata/lococonfig/00001.ecs", data: backupConfigFixture(1, "0:8:151\n")})
	data := encryptBackupFixture(t, tarBackupFixture(t, members))
	result, err := DecodeBackup(bytes.NewReader(data), ring, nil)
	if err != nil || result.Serial != 0x12345678 || len(result.Locomotives) != 2 {
		t.Fatalf("result=%+v error=%v", result, err)
	}
	if result.Locomotives[0].ObjectID != 1 || result.Locomotives[1].ObjectID != 1002 {
		t.Fatal("locomotives not sorted deterministically")
	}
	values := result.Locomotives[1].Values
	if len(values) != 5 || values[0].IndexCV31 != nil || values[4].IndexCV31 != nil ||
		values[1].Value != 0 || *values[1].IndexCV31 != 0 || *values[1].IndexCV32 != 0 ||
		*values[2].IndexCV31 != 16 || *values[2].IndexCV32 != 3 || *values[3].IndexCV32 != 8 {
		t.Fatalf("register identity lost: %+v", values)
	}
}

func TestBackupRequiresEncryptionPrivateKeyAndVerifiedMDC(t *testing.T) {
	entity, ring := backupTestKeys(t)
	body := tarBackupFixture(t, validBackupMembers())
	valid := encryptBackupFixture(t, body)
	tampered := append([]byte(nil), valid...)
	tampered[len(tampered)-1] ^= 1
	var public bytes.Buffer
	if err := entity.Serialize(&public); err != nil {
		t.Fatal(err)
	}
	for name, data := range map[string][]byte{
		"unencrypted archive": body,
		"empty":               nil,
		"truncated envelope":  valid[:len(valid)/2],
		"truncated MDC":       valid[:len(valid)-1],
		"tampered MDC":        tampered,
		"legacy without MDC":  {0xa4, 0},
	} {
		t.Run(name, func(t *testing.T) {
			result, err := DecodeBackup(bytes.NewReader(data), ring, nil)
			if !errors.Is(err, ErrBackup) || result.Serial != 0 || result.Locomotives != nil {
				t.Fatalf("unverified data accepted: %+v %v", result, err)
			}
		})
	}
	for name, keys := range map[string][]byte{"missing key": nil, "malformed key": []byte("not a key"), "public key only": public.Bytes()} {
		t.Run(name, func(t *testing.T) {
			if _, err := DecodeBackup(bytes.NewReader(valid), keys, []byte("not disclosed")); !errors.Is(err, ErrBackupKey) {
				t.Fatalf("expected generic private key error, got %v", err)
			}
		})
	}
}

func TestBackupRejectsUnsafeArchiveMembersAndMissingIdentity(t *testing.T) {
	cases := map[string][]backupTestMember{
		"missing serial":    validBackupMembers()[1:],
		"short serial":      {{name: "root/loks.bin", data: []byte{1}}, validBackupMembers()[1]},
		"zero serial":       {{name: "root/loks.bin", data: []byte{0, 0, 0, 0}}, validBackupMembers()[1]},
		"missing configs":   validBackupMembers()[:1],
		"absolute path":     {{name: "/outside", data: []byte("x")}},
		"traversal":         {{name: "root/../outside", data: []byte("x")}},
		"parent traversal":  {{name: "../outside", data: []byte("x")}},
		"windows separator": {{name: `root\outside`, data: []byte("x")}},
		"symlink":           {{name: "root/link", kind: tar.TypeSymlink}},
		"hardlink":          {{name: "root/link", kind: tar.TypeLink}},
		"duplicate member":  {validBackupMembers()[0], validBackupMembers()[0], validBackupMembers()[1]},
		"duplicate object":  {validBackupMembers()[0], validBackupMembers()[1], {name: "root/usrdata/lococonfig/1002.ecs", data: backupConfigFixture(1002, "0:8:151\n")}},
	}
	for name, members := range cases {
		t.Run(name, func(t *testing.T) {
			if result, err := decodeBackupArchive(tarBackupFixture(t, members)); !errors.Is(err, ErrBackup) || result.Serial != 0 {
				t.Fatalf("unsafe archive accepted: %+v %v", result, err)
			}
		})
	}
}

func TestBackupConfigRejectsAmbiguousAddressesAndMalformedValues(t *testing.T) {
	for name, body := range map[string][]byte{
		"wrong object":       backupConfigFixture(1003, "0:1:3\n"),
		"wrong version":      []byte("lococonfig 0 2 1002 29\n0:1:3\n"),
		"wrong profile":      []byte("lococonfig 0 1 1002 unknown\n0:1:3\n"),
		"duplicate CV":       backupConfigFixture(1002, "0:1:3\n00:01:4\n"),
		"duplicate indexed":  backupConfigFixture(1002, "4099:257:20\n04099:0257:0\n"),
		"CV zero":            backupConfigFixture(1002, "0:0:3\n"),
		"CV too high":        backupConfigFixture(1002, "0:1025:3\n"),
		"byte too high":      backupConfigFixture(1002, "0:1:256\n"),
		"negative byte":      backupConfigFixture(1002, "0:1:-1\n"),
		"signed number":      backupConfigFixture(1002, "0:+1:3\n"),
		"too high page":      backupConfigFixture(1002, "65536:257:1\n"),
		"nonindexed page":    backupConfigFixture(1002, "4099:1:3\n"),
		"nonindexed high CV": backupConfigFixture(1002, "1:700:3\n"),
		"missing value":      backupConfigFixture(1002, "0:1\n"),
		"blank entry":        backupConfigFixture(1002, "0:1:3\n\n"),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := parseBackupConfig(body, 1002); !errors.Is(err, ErrBackup) {
				t.Fatalf("malformed config accepted: %v", err)
			}
		})
	}
}

func TestBackupLimitsEncryptedPlaintextArchiveAndConfigSizes(t *testing.T) {
	_, ring := backupTestKeys(t)
	if _, err := DecodeBackup(bytes.NewReader(make([]byte, EncryptedBackupLimit+1)), ring, nil); !errors.Is(err, ErrBackup) {
		t.Fatal("encrypted upload limit not enforced")
	}
	compressed := encryptCompressedBackupFixture(t, make([]byte, backupPlaintextLimit+1))
	if len(compressed) >= EncryptedBackupLimit {
		t.Fatal("test must exercise plaintext expansion, not the encrypted upload limit")
	}
	if _, err := DecodeBackup(bytes.NewReader(compressed), ring, nil); !errors.Is(err, ErrBackup) {
		t.Fatal("decrypted body limit not enforced")
	}
	if _, err := decodeBackupArchive(gzipBackupFixture(t, make([]byte, backupPlaintextLimit+1))); !errors.Is(err, ErrBackup) {
		t.Fatal("gzip expansion limit not enforced")
	}
	oversized := validBackupMembers()
	oversized[1].data = make([]byte, backupConfigLimit+1)
	if _, err := decodeBackupArchive(tarBackupFixture(t, oversized)); !errors.Is(err, ErrBackup) {
		t.Fatal("config size limit not enforced")
	}
}

func TestBackupLimitsObjectAndArchiveCountsAndAllows8192DistinctCVs(t *testing.T) {
	var lines strings.Builder
	for index := 0; index < CVDatabaseLimit; index++ {
		fmt.Fprintf(&lines, "%d:%d:0\n", index/256, 257+index%256)
	}
	if values, err := parseBackupConfig(backupConfigFixture(1002, lines.String()), 1002); err != nil || len(values) != CVDatabaseLimit {
		t.Fatalf("full register set rejected: %d %v", len(values), err)
	}
	lines.WriteString("32:257:0\n")
	if _, err := parseBackupConfig(backupConfigFixture(1002, lines.String()), 1002); !errors.Is(err, ErrBackup) {
		t.Fatal("CV count limit not enforced")
	}
	members := validBackupMembers()[:1]
	for id := 1; id <= backupLocoLimit+1; id++ {
		members = append(members, backupTestMember{name: fmt.Sprintf("root/usrdata/lococonfig/%d.ecs", id), data: backupConfigFixture(id, "0:1:3\n")})
	}
	if _, err := decodeBackupArchive(tarBackupFixture(t, members)); !errors.Is(err, ErrBackup) {
		t.Fatal("object count limit not enforced")
	}
	members = validBackupMembers()
	for id := 0; id < backupMemberLimit; id++ {
		members = append(members, backupTestMember{name: fmt.Sprintf("root/ignored%d", id)})
	}
	if _, err := decodeBackupArchive(tarBackupFixture(t, members)); !errors.Is(err, ErrBackup) {
		t.Fatal("archive member count limit not enforced")
	}
}

func TestBackupZipMembersAreReadToVerifiedEOF(t *testing.T) {
	var buffer bytes.Buffer
	archive := zip.NewWriter(&buffer)
	for _, member := range append(validBackupMembers(), backupTestMember{name: "root/ignored.bin", data: []byte("ignored bytes still require a verified CRC")}) {
		file, err := archive.Create(member.name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := file.Write(member.data); err != nil {
			t.Fatal(err)
		}
	}
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}
	if result, err := decodeBackupArchive(buffer.Bytes()); err != nil || len(result.Locomotives) != 1 {
		t.Fatalf("valid ZIP rejected: %+v %v", result, err)
	}
	data := append([]byte(nil), buffer.Bytes()...)
	// Corrupt the ignored member's central CRC, which only fails on a full read.
	offset := bytes.LastIndex(data, []byte("PK\x01\x02"))
	if offset < 0 {
		t.Fatal("ZIP central header missing")
	}
	data[offset+16] ^= 0x20
	if _, err := decodeBackupArchive(data); !errors.Is(err, ErrBackup) {
		t.Fatal("corrupt ignored/identity ZIP data accepted")
	}
	// Archive input errors are not exposed with potentially private details.
	if _, err := readBackupBytes(&backupFailureReader{}, 1024); !errors.Is(err, ErrBackup) || strings.Contains(err.Error(), "private") {
		t.Fatalf("unsafe read error: %v", err)
	}
}

type backupFailureReader struct{}

func (*backupFailureReader) Read([]byte) (int, error) {
	return 0, fmt.Errorf("private input detail: %w", io.ErrUnexpectedEOF)
}
