package ecos

import (
	"encoding/binary"
	"testing"
)

func storedBackupRecordFixture(typ, id uint32, name string, members ...uint32) []byte {
	size, nameOffset := 948, 28
	if typ&0x7fffffff == 11 {
		size, nameOffset = 40+4*len(members), 8
	}
	if typ&0x80000000 != 0 {
		size += 8
	}
	data := make([]byte, 4+size)
	binary.LittleEndian.PutUint32(data[:4], typ)
	body := data[4:]
	binary.LittleEndian.PutUint32(body[:4], id)
	copy(body[nameOffset:nameOffset+20], name)
	if typ&0x7fffffff == 11 {
		binary.LittleEndian.PutUint32(body[36:40], uint32(len(members)))
		for index, member := range members {
			binary.LittleEndian.PutUint32(body[40+index*4:44+index*4], member)
		}
	}
	return data
}

func storedBackupFileFixture(records ...[]byte) []byte {
	data := make([]byte, 52)
	binary.LittleEndian.PutUint32(data[:4], 0x12345678)
	binary.LittleEndian.PutUint32(data[4:8], 151)
	binary.LittleEndian.PutUint32(data[48:52], uint32(len(records)))
	for _, record := range records {
		data = append(data, record...)
	}
	return data
}

func TestBackupStoredLocomotivesDecodeMetadataWithoutInventingCVValidity(t *testing.T) {
	first := storedBackupRecordFixture(5, 57, "Synthetic engine")
	body := first[4:]
	binary.LittleEndian.PutUint32(body[52:56], 1<<31|1<<23|3)
	binary.LittleEndian.PutUint32(body[60:64], 0x8003)
	binary.LittleEndian.PutUint32(body[60+23*12:64+23*12], 0xc002)
	binary.LittleEndian.PutUint32(body[60+24*12:64+24*12], 0x8000) // Not a function slot.
	body[0x168], body[0x169], body[0x170], body[0x267], body[0x268] = 12, 3, 97, 9, 13
	consist := storedBackupRecordFixture(11, 58, "Synthetic consist", 57, 59)
	last := storedBackupRecordFixture(0x80000005, 59, "Synthetic Güter")
	data := storedBackupFileFixture(first, consist, last)
	locomotives, supported := decodeBackupStoredLocomotives(data)
	if !supported || len(locomotives) != 2 {
		t.Fatalf("metadata rejected or consist included: %+v %t", locomotives, supported)
	}
	loco := locomotives[0]
	if loco.ObjectID != 57 || loco.Name != "Synthetic engine" || loco.Protocol != "DCC" ||
		loco.FunctionMask != 1<<31|1<<23|3 || loco.MomentaryMask != 1<<23|1 {
		t.Fatalf("metadata or bounded mode slots changed: %+v", loco)
	}
	if len(loco.UnverifiedCVs) != 3 || loco.UnverifiedCVs[0].Number != 1 ||
		loco.UnverifiedCVs[0].Value != 3 || loco.UnverifiedCVs[1].Number != 8 ||
		loco.UnverifiedCVs[1].Value != 97 || loco.UnverifiedCVs[2].Number != 255 ||
		loco.UnverifiedCVs[2].Value != 9 || loco.UnverifiedCVs[0].IndexCV31 != nil {
		t.Fatalf("zero, CV0, CV256 or inferred pages became unverified readings: %+v", loco.UnverifiedCVs)
	}
	if locomotives[1].Name != "Synthetic Güter" || len(locomotives[1].UnverifiedCVs) != 0 {
		t.Fatal("UTF-8 name, optional UID or empty candidate set was changed")
	}
}

func TestBackupStoredLocomotivesRejectWholeUnsupportedOrMalformedMetadata(t *testing.T) {
	valid := storedBackupFileFixture(storedBackupRecordFixture(5, 57, "Synthetic engine"))
	cases := map[string][]byte{
		"short header": valid[:51],
		"short record": valid[:len(valid)-1],
		"trailing":     append(append([]byte{}, valid...), 0),
		"duplicate id": storedBackupFileFixture(storedBackupRecordFixture(5, 57, "First"),
			storedBackupRecordFixture(11, 57, "Duplicate")),
	}
	mutate := func(offset int, value uint32) []byte {
		data := append([]byte{}, valid...)
		binary.LittleEndian.PutUint32(data[offset:offset+4], value)
		return data
	}
	cases["version"] = mutate(4, 150)
	cases["object limit"] = mutate(48, 5001)
	cases["object count overflow"] = mutate(48, 0xffffffff)
	cases["missing object"] = mutate(48, 2)
	cases["unknown type"] = mutate(52, 99)
	cases["zero object id"] = mutate(56, 0)
	cases["object id limit"] = mutate(56, 2_000_000_001)
	cases["missing uid"] = mutate(52, 0x80000005)
	badName := append([]byte{}, valid...)
	badName[56+28] = 0xff
	cases["invalid utf8"] = badName
	cases["control name"] = storedBackupFileFixture(storedBackupRecordFixture(5, 57, "Bad\nname"))
	cases["empty name"] = storedBackupFileFixture(storedBackupRecordFixture(5, 57, ""))
	badConsist := storedBackupFileFixture(storedBackupRecordFixture(11, 58, "Consist"))
	binary.LittleEndian.PutUint32(badConsist[56+36:56+40], 5001)
	cases["consist member limit"] = badConsist
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			locomotives, supported := decodeBackupStoredLocomotives(data)
			if supported || locomotives != nil {
				t.Fatalf("invalid metadata partially accepted: %+v %t", locomotives, supported)
			}
		})
	}
}

func TestBackupStoredLocomotivesStaySeparateFromVerifiedTextCVs(t *testing.T) {
	record := storedBackupRecordFixture(5, 57, "Synthetic engine")
	record[4+0x169] = 3
	file := storedBackupFileFixture(record)
	members := []backupTestMember{
		{name: "root/loks.bin", data: file},
		{name: "root/usrdata/lococonfig/00057.ecs", data: backupConfigFixture(57, "0:1:0\n")},
	}
	backup, err := decodeBackupArchive(tarBackupFixture(t, members))
	if err != nil || backup.StoredLocomotivesStatus != BackupStoredLocomotivesAvailable ||
		len(backup.StoredLocomotives) != 1 || len(backup.Locomotives) != 1 {
		t.Fatalf("metadata archive failed: %+v %v", backup, err)
	}
	if len(backup.Locomotives[0].Values) != 1 || backup.Locomotives[0].Values[0].Value != 0 ||
		backup.StoredLocomotives[0].UnverifiedCVs[0].Value != 3 {
		t.Fatal("legacy candidate replaced the separate text CV value")
	}
	for _, malformed := range [][]byte{backupSerialFixture(), append(file, 0)} {
		members[0].data = malformed
		backup, err = decodeBackupArchive(tarBackupFixture(t, members))
		if err != nil || backup.StoredLocomotivesStatus != BackupStoredLocomotivesUnavailable ||
			backup.StoredLocomotives != nil || len(backup.Locomotives[0].Values) != 1 || backup.Locomotives[0].Values[0].Value != 0 {
			t.Fatalf("optional unsupported metadata damaged text CVs: %+v %v", backup, err)
		}
	}
}
