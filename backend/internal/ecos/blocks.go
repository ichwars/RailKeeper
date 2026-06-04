package ecos

import (
	"fmt"
	"strconv"
	"strings"
)

type BlockKind string

const (
	BlockReply BlockKind = "reply"
	BlockEvent BlockKind = "event"
)

type Block struct {
	Kind     BlockKind `json:"kind"`
	ObjectID int       `json:"objectId,omitempty"`
	Header   string    `json:"header"`
	EndLine  string    `json:"endLine"`
	Lines    []string  `json:"lines"`
}

func HasBlockLine(line string) bool {
	line = strings.TrimSpace(line)
	return strings.HasPrefix(line, "<END")
}

func ParseBlocks(lines []string) ([]Block, error) {
	blocks := []Block{}
	current := []string{}
	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}
		current = append(current, line)
		if !HasBlockLine(line) {
			continue
		}
		block, err := parseBlock(current)
		if err != nil {
			return blocks, err
		}
		blocks = append(blocks, block)
		current = []string{}
	}
	return blocks, nil
}

func parseBlock(lines []string) (Block, error) {
	if len(lines) < 2 {
		return Block{}, fmt.Errorf("unvollständiger ECoS-Block")
	}
	header := strings.TrimSpace(lines[0])
	block := Block{
		Header:  header,
		EndLine: strings.TrimSpace(lines[len(lines)-1]),
		Lines:   append([]string(nil), lines...),
	}
	switch {
	case strings.HasPrefix(header, "<REPLY "):
		block.Kind = BlockReply
	case strings.HasPrefix(header, "<EVENT "):
		block.Kind = BlockEvent
	default:
		return Block{}, fmt.Errorf("unbekannter ECoS-Block: %s", header)
	}
	block.ObjectID = parseHeaderObjectID(header)
	return block, nil
}

func parseHeaderObjectID(header string) int {
	start := strings.Index(header, "(")
	if strings.HasPrefix(header, "<EVENT ") {
		start = strings.Index(header, " ")
	}
	if start < 0 {
		return 0
	}
	rest := strings.TrimLeft(header[start+1:], " ")
	end := strings.IndexAny(rest, ",)> ")
	if end >= 0 {
		rest = rest[:end]
	}
	value, _ := strconv.Atoi(strings.TrimSpace(rest))
	return value
}
