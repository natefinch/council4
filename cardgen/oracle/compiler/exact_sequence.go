package compiler

import "github.com/natefinch/council4/cardgen/oracle/parser"

// ExactSequenceKind identifies a parser-recognized exact multi-instruction
// resolving sequence. It is the only text-aware input the compiler and lowering
// consume for these bodies: both switch on the kind without inspecting Oracle
// words, so they remain text-blind and fail closed on unknown kinds.
type ExactSequenceKind uint8

const (
	// ExactSequenceUnknown marks content with no recognized exact sequence.
	ExactSequenceUnknown ExactSequenceKind = iota
	// ExactSequenceChosenTypeLibraryTopToHand is the upkeep "look at the top
	// card of your library; if it's a creature card of the chosen type, you may
	// reveal it and put it into your hand" sequence.
	ExactSequenceChosenTypeLibraryTopToHand
)

func compileExactSequenceKind(kind parser.ExactSequenceKind) ExactSequenceKind {
	switch kind {
	case parser.ExactSequenceChosenTypeLibraryTopToHand:
		return ExactSequenceChosenTypeLibraryTopToHand
	default:
		return ExactSequenceUnknown
	}
}
