package gofrc

import (
	"encoding/base64"
	"math"
	"strings"
)

const (
	PuzzleExpiryOffset     = 13
	NumberOfPuzzlesOffset  = 14
	PuzzleDifficultyOffset = 15
)

type Puzzle struct {
	Signature string
	Base64    string
	Buffer    []byte
	Threshold uint32
	N         int
	Expiry    uint32
}

func NewPuzzle(input string) (*Puzzle, error) {
	parts := strings.Split(input, ".")

	signature, puzzle := parts[0], parts[1]

	buffer, err := base64.StdEncoding.DecodeString(puzzle)
	if err != nil {
		return nil, err
	}

	return &Puzzle{
		Signature: signature,
		Base64:    puzzle,
		Buffer:    buffer,
		N:         int(buffer[NumberOfPuzzlesOffset]),
		Threshold: difficultyToThreshold(buffer[PuzzleDifficultyOffset]),
		Expiry:    uint32(buffer[PuzzleExpiryOffset]) * 300000,
	}, nil
}

func difficultyToThreshold(value uint8) uint32 {
	return uint32(math.Pow(2, (255.999-float64(value))/8.0)) >> 0
}
