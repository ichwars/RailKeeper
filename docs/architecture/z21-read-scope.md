# Z21 read-only scope

## Decision

RailKeeper does not treat the Z21 LAN protocol as a locomotive roster. The current adapter polls only
decoder addresses that are already known to RailKeeper and places the verified address into the common
command-station comparison workspace.

The implementation uses only `LAN_X_GET_LOCO_INFO` (`0x40`, `0xE3`, `0xF0`). It does not send drive,
function, emergency-stop, purge, CV, turnout, broadcast-flag, or layout commands.

## Evidence and limits

The [Z21 LAN Protocol Specification 1.13](https://www.z21.eu/media/Kwc_Basic_DownloadTag_Component/root-en-main_47-1652-959-downloadTag-download/default/d559b9cf/1699290380/z21-lan-protokoll-en.pdf)
documents `LAN_X_GET_LOCO_INFO` as a status poll for one locomotive address. The same request subscribes
the UDP client to that address, but unsolicited updates additionally require broadcast flags. RailKeeper
does not enable those flags and closes each request socket after the matching response.

The protocol response contains an address plus dynamic driving and function state. It does not contain a
locomotive name. Version 1.13 describes an MM indicator introduced with firmware 1.43, but its DB2 bit
diagram is internally inconsistent (`0000BKKK` while the prose names `M`). Until firmware-aware hardware
captures remove that ambiguity, RailKeeper does not infer DCC or Motorola from this field.

Consequently, the preview keeps only:

- the decoder address returned for the requested address;
- an incomplete marker explaining that name and protocol are unavailable.

Speed, direction, busy state, functions F0-F31, and future optional response bytes are validated only as
bounded packet data and then discarded.

## Safety bounds

- Addresses originate from RailKeeper's decoder number or an existing Z21 external mapping.
- Duplicate addresses are queried once in ascending order.
- A read session is rejected above 64 unique addresses instead of silently truncating the preview.
- The complete Z21 read has a 15-second deadline. Each UDP exchange retains its shorter transport timeout.
- Each datagram is limited to 1472 bytes and 32 datagrams per request.
- A response must use the X-BUS header and `LAN_X_LOCO_INFO` type, contain 10-17 payload bytes, have a
  valid XOR checksum, and repeat the requested address.
- Timeout, truncated response, invalid length, invalid checksum, unexpected type, and unexpected address
  remain distinguishable errors.

## Validation status

The repository fixture models the byte layout from protocol version 1.13 and is explicitly marked as not
being a hardware capture. Automated UDP tests cover combined asynchronous datasets, address matching,
checksum and length rejection, bounded input, and integration with the shared conflict workspace.

Validation against real Z21 and z21 hardware, firmware variants, and captured replies is still required
before issue #133 can be considered complete. Read capability remains disabled for the Intellibox 3
adapter because its Z21-compatible mode has not been proven with real hardware.
