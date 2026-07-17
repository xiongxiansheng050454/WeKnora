package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
)

const nullByte = 0

// ChunkID generates a deterministic content-addressed chunk ID.
// Same (knowledgeID, normalized content, seq, chunkType) always yields the same ID.
func ChunkID(knowledgeID string, content string, seq int, chunkType string) string {
	h := sha256.New()
	h.Write([]byte(knowledgeID))
	h.Write([]byte{nullByte})
	h.Write([]byte(strings.TrimSpace(content)))
	h.Write([]byte{nullByte})
	h.Write([]byte(strconv.Itoa(seq)))
	h.Write([]byte{nullByte})
	h.Write([]byte(chunkType))
	return hex.EncodeToString(h.Sum(nil))
}

// ImageChunkID generates a deterministic ID for image sub-chunks (OCR / caption).
// parentChunkID must itself be deterministic (content-addressed).
func ImageChunkID(parentChunkID string, subType string, content string) string {
	h := sha256.New()
	h.Write([]byte(parentChunkID))
	h.Write([]byte{nullByte})
	h.Write([]byte(subType))
	h.Write([]byte{nullByte})
	h.Write([]byte(strings.TrimSpace(content)))
	return hex.EncodeToString(h.Sum(nil))
}

// ContentHash returns a hex SHA-256 of the input string.
func ContentHash(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// MultiFieldHash returns a hex SHA-256 of all fields joined with null bytes.
func MultiFieldHash(fields ...string) string {
	h := sha256.New()
	for i, f := range fields {
		if i > 0 {
			h.Write([]byte{nullByte})
		}
		h.Write([]byte(f))
	}
	return hex.EncodeToString(h.Sum(nil))
}
