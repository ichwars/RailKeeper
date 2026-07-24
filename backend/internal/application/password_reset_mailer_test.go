package application

import "testing"

func TestCanonicalMailbox(t *testing.T) {
	mailbox, err := canonicalMailbox("RailKeeper User <user@example.test>")
	if err != nil {
		t.Fatal(err)
	}
	if mailbox != "user@example.test" {
		t.Fatalf("expected canonical mailbox, got %q", mailbox)
	}
}

func TestCanonicalMailboxRejectsHeaderInjection(t *testing.T) {
	_, err := canonicalMailbox("victim@example.test\r\nBcc: attacker@example.test")
	if err == nil {
		t.Fatal("expected injected mail header to be rejected")
	}
}
