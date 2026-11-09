package application

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"
)

func (s *DigitalCenterService) exchangeZ21UDP(ctx context.Context, target DigitalCenterConnectionInput, command []byte) ([]byte, error) {
	if len(command) < 4 {
		return nil, errors.New("Z21-Befehl war unvollständig")
	}
	expectedHeader := binary.LittleEndian.Uint16(command[2:4])
	return s.exchangeZ21UDPMatching(ctx, target, command, func(packet z21Packet) (bool, error) {
		if packet.Header != expectedHeader {
			return false, nil
		}
		if err := validateZ21PayloadLength(packet.Header, packet.Payload); err != nil {
			return false, err
		}
		return true, nil
	})
}

func (s *DigitalCenterService) exchangeZ21UDPMatching(
	ctx context.Context,
	target DigitalCenterConnectionInput,
	command []byte,
	matcher func(z21Packet) (bool, error),
) ([]byte, error) {
	if len(command) < 4 {
		return nil, errors.New("Z21-Befehl war unvollständig")
	}
	expectedHeader := binary.LittleEndian.Uint16(command[2:4])
	timeout := s.timeout
	if timeout <= 0 {
		timeout = 4 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	dialer := net.Dialer{Timeout: timeout}
	conn, err := dialer.DialContext(ctx, "udp", net.JoinHostPort(target.Host, strconv.Itoa(target.Port)))
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Close() }()
	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	} else {
		_ = conn.SetDeadline(time.Now().Add(timeout))
	}
	if _, err := conn.Write(command); err != nil {
		return nil, err
	}
	buffer := make([]byte, z21MaximumUDPPayload+1)
	unmatchedResponse := false
	for datagram := 0; datagram < z21MaximumDatagrams; datagram++ {
		count, readErr := conn.Read(buffer)
		if readErr != nil {
			if unmatchedResponse {
				return nil, fmt.Errorf("keine passende Z21-Antwort für Header 0x%04X erhalten", expectedHeader)
			}
			return nil, readErr
		}
		if count > z21MaximumUDPPayload {
			return nil, fmt.Errorf("Z21-Antwort überschritt die maximale UDP-Nutzdatenlänge von %d Byte", z21MaximumUDPPayload)
		}
		packets, parseErr := parseZ21Packets(buffer[:count])
		if parseErr != nil {
			return nil, parseErr
		}
		for _, packet := range packets {
			matched, matchErr := matcher(packet)
			if matchErr != nil {
				return nil, matchErr
			}
			if !matched {
				unmatchedResponse = true
				continue
			}
			return append([]byte(nil), packet.Raw...), nil
		}
	}
	return nil, fmt.Errorf("keine passende Z21-Antwort für Header 0x%04X nach %d UDP-Datagrammen erhalten", expectedHeader, z21MaximumDatagrams)
}

func (s *DigitalCenterService) readZ21LocoInfo(
	ctx context.Context,
	target DigitalCenterConnectionInput,
	address int,
) (ECoSRawLocomotive, error) {
	command, err := z21LocoInfoCommand(address)
	if err != nil {
		return ECoSRawLocomotive{}, err
	}
	var result ECoSRawLocomotive
	mismatch := ""
	_, err = s.exchangeZ21UDPMatching(ctx, target, command, func(packet z21Packet) (bool, error) {
		if packet.Header != z21LANXHeader {
			return false, nil
		}
		if len(packet.Payload) == 0 {
			return false, errors.New("Z21-Lokantwort war unvollständig: X-Header fehlt")
		}
		if packet.Payload[0] != 0xEF {
			mismatch = fmt.Sprintf("unerwarteter Z21-X-BUS-Nachrichtentyp 0x%02X", packet.Payload[0])
			return false, nil
		}
		locomotive, parseErr := parseZ21LocoInfo(packet.Payload)
		if parseErr != nil {
			return false, parseErr
		}
		if locomotive.Address != address {
			mismatch = fmt.Sprintf("Z21-Lokantwort für unerwartete Adresse %d", locomotive.Address)
			return false, nil
		}
		result = locomotive
		return true, nil
	})
	if err != nil {
		if mismatch != "" {
			return ECoSRawLocomotive{}, fmt.Errorf("%s: %w", mismatch, err)
		}
		return ECoSRawLocomotive{}, err
	}
	return result, nil
}

func z21SerialNumberCommand() []byte {
	return z21HeaderCommand(z21LANGetSerialNumber)
}

func z21HeaderCommand(header uint16) []byte {
	packet := make([]byte, 4)
	binary.LittleEndian.PutUint16(packet[0:2], 4)
	binary.LittleEndian.PutUint16(packet[2:4], header)
	return packet
}

func z21LocoInfoCommand(address int) ([]byte, error) {
	if address < 1 || address > maxDigitalCenterAddress {
		return nil, fmt.Errorf("Z21-Decoderadresse %d liegt außerhalb 1 bis %d", address, maxDigitalCenterAddress)
	}
	addressMSB := byte(address >> 8)
	if address >= 128 {
		addressMSB |= 0xC0
	}
	command := []byte{0x09, 0x00, 0x40, 0x00, 0xE3, 0xF0, addressMSB, byte(address), 0x00}
	for _, value := range command[4 : len(command)-1] {
		command[len(command)-1] ^= value
	}
	return command, nil
}

func normalizeZ21ReadAddresses(addresses []int) ([]int, error) {
	unique := make(map[int]struct{}, len(addresses))
	for _, address := range addresses {
		if address < 1 || address > maxDigitalCenterAddress {
			return nil, fmt.Errorf("Z21-Decoderadresse %d liegt außerhalb 1 bis %d", address, maxDigitalCenterAddress)
		}
		unique[address] = struct{}{}
	}
	if len(unique) > z21MaximumReadAddresses {
		return nil, fmt.Errorf("Z21-Leseumfang ist auf %d eindeutige Decoderadressen begrenzt", z21MaximumReadAddresses)
	}
	result := make([]int, 0, len(unique))
	for address := range unique {
		result = append(result, address)
	}
	sort.Ints(result)
	return result, nil
}

func parseZ21LocoInfo(payload []byte) (ECoSRawLocomotive, error) {
	if len(payload) < 10 {
		return ECoSRawLocomotive{}, fmt.Errorf("Z21-Lokantwort war unvollständig: erwartete mindestens 10 Nutzdatenbytes, erhielt %d", len(payload))
	}
	if len(payload) > 17 {
		return ECoSRawLocomotive{}, fmt.Errorf("Z21-Lokantwort hatte eine ungültige Nutzdatenlänge von %d Byte", len(payload))
	}
	if payload[0] != 0xEF {
		return ECoSRawLocomotive{}, fmt.Errorf("unerwarteter Z21-X-BUS-Nachrichtentyp 0x%02X", payload[0])
	}
	checksum := byte(0)
	for _, value := range payload {
		checksum ^= value
	}
	if checksum != 0 {
		return ECoSRawLocomotive{}, errors.New("Z21-Lokantwort hatte eine ungültige XOR-Prüfsumme")
	}
	address := int(payload[1]&0x3F)<<8 | int(payload[2])
	if address < 1 || address > maxDigitalCenterAddress {
		return ECoSRawLocomotive{}, fmt.Errorf("Z21-Lokantwort enthielt ungültige Decoderadresse %d", address)
	}
	return ECoSRawLocomotive{
		ObjectID: address,
		Address:  address,
		DetailError: "Z21-LAN stellt weder Loknamen noch ein verlässlich eindeutiges Decoderprotokoll bereit; " +
			"Fahrzustände wurden verworfen.",
	}, nil
}

func decodeZ21ProbeFields(header uint16, payload []byte) map[string]string {
	fields := map[string]string{}
	switch header {
	case z21LANGetSerialNumber:
		if len(payload) == 4 {
			fields["serialNumber"] = strconv.FormatUint(uint64(binary.LittleEndian.Uint32(payload[:4])), 10)
		}
	case z21LANGetCode:
		if len(payload) == 1 {
			fields["centralCode"] = fmt.Sprintf("%d (0x%02X)", payload[0], payload[0])
		}
	case z21LANGetHWInfo:
		if len(payload) == 8 {
			fields["hardwareTypeRaw"] = fmt.Sprintf("0x%08X", binary.LittleEndian.Uint32(payload[:4]))
			fields["firmwareVersionRaw"] = fmt.Sprintf("0x%08X", binary.LittleEndian.Uint32(payload[4:8]))
		}
	}
	return fields
}

func z21TransportFields(provider string) map[string]string {
	fields := map[string]string{"transport": "z21_udp"}
	if provider == "intellibox3" {
		fields["z21UdpStatus"] = "available"
		fields["loconetTcpStatus"] = "planned"
	}
	return fields
}

func validateZ21PayloadLength(header uint16, payload []byte) error {
	expected := 0
	switch header {
	case z21LANGetSerialNumber:
		expected = 4
	case z21LANGetCode:
		expected = 1
	case z21LANGetHWInfo:
		expected = 8
	default:
		return nil
	}
	if len(payload) != expected {
		return fmt.Errorf("Z21-Antwort für Header 0x%04X erwartete Nutzdatenlänge %d, erhielt %d", header, expected, len(payload))
	}
	return nil
}

func formatHex(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	return strings.ToUpper(hex.EncodeToString(data))
}

func parseZ21Packet(packet []byte) (uint16, []byte, error) {
	packets, err := parseZ21Packets(packet)
	if err != nil {
		return 0, nil, err
	}
	if len(packets) != 1 {
		return 0, nil, errors.New("Z21-Antwort enthielt unerwartet mehrere Datensätze")
	}
	return packets[0].Header, packets[0].Payload, nil
}

func parseZ21Packets(datagram []byte) ([]z21Packet, error) {
	if len(datagram) > z21MaximumUDPPayload {
		return nil, fmt.Errorf("Z21-Antwort überschritt die maximale UDP-Nutzdatenlänge von %d Byte", z21MaximumUDPPayload)
	}
	if len(datagram) < 4 {
		return nil, errors.New("Z21-Antwort war unvollständig")
	}
	packets := make([]z21Packet, 0, 1)
	for offset := 0; offset < len(datagram); {
		if len(datagram)-offset < 4 {
			return nil, errors.New("Z21-Antwort enthielt unvollständige nachfolgende Daten")
		}
		length := int(binary.LittleEndian.Uint16(datagram[offset : offset+2]))
		if length < 4 || length > len(datagram)-offset {
			return nil, errors.New("Z21-Antwort hatte eine ungültige Datensatzlänge")
		}
		raw := datagram[offset : offset+length]
		packets = append(packets, z21Packet{
			Raw:     raw,
			Header:  binary.LittleEndian.Uint16(raw[2:4]),
			Payload: raw[4:],
		})
		offset += length
	}
	return packets, nil
}
