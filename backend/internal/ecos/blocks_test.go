package ecos

import "testing"

func TestParseBlocks(t *testing.T) {
	blocks, err := ParseBlocks([]string{
		"<REPLY get(1, info, status)>",
		"1 status[GO]",
		"1 ProtocolVersion[0.5]",
		"<END 0 (OK)>",
		"<EVENT 10>",
		"10 msg[LIST_CHANGED]",
		"1016 appended",
		"<END 0 (OK)>",
	})
	if err != nil {
		t.Fatalf("ParseBlocks returned error: %v", err)
	}
	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(blocks))
	}
	if blocks[0].Kind != BlockReply || blocks[0].ObjectID != 1 {
		t.Fatalf("unexpected reply block: %+v", blocks[0])
	}
	if blocks[1].Kind != BlockEvent || blocks[1].ObjectID != 10 {
		t.Fatalf("unexpected event block: %+v", blocks[1])
	}
}
