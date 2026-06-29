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
	// ExactSequenceBottomHandThenDraw is the spell "put any number of cards from
	// your hand on the bottom (or top) of your library, then draw that many
	// cards[ plus N]" sequence.
	ExactSequenceBottomHandThenDraw
	// ExactSequenceDiscardHandThenDraw is the spell "Discard {your hand | all
	// the cards in your hand}, then draw that many cards." sequence: the
	// controller discards their whole hand, then draws that many cards.
	ExactSequenceDiscardHandThenDraw
	// ExactSequenceConditionalLookAtTopReveal is the triggered "look at the top
	// card of your library; if it's a card of one of the recorded types, you may
	// reveal it and put it into your hand; if you don't, you may put it into your
	// graveyard" sequence. The recorded card types travel on the compiled
	// ability so lowering filters the reveal without reading Oracle words.
	ExactSequenceConditionalLookAtTopReveal
	// ExactSequenceConditionalLookAtTopBattlefield is the triggered "look at the
	// top card of your library; if it's a card of one of the recorded types, you
	// may put it onto the battlefield[ tapped]" sequence, optionally followed by
	// a mandatory "if you don't, put it into your hand" fallback. The recorded
	// card types, the tapped entry rider, and the fallback disposition travel on
	// the compiled ability so lowering routes the card without reading Oracle
	// words.
	ExactSequenceConditionalLookAtTopBattlefield
	// ExactSequenceDrawThenDiscardUnlessType is the spell "Draw N cards. Then
	// discard M cards unless you discard a <type[ or type...]> card." sequence:
	// the controller draws ExactSequenceDrawCount cards, then discards
	// ExactSequenceDiscardCount cards unless they instead discard a single card
	// of one of the recorded exempt types. The counts and exempt types travel on
	// the compiled ability so lowering routes the discard without reading Oracle
	// words.
	ExactSequenceDrawThenDiscardUnlessType
)

func compileExactSequenceKind(kind parser.ExactSequenceKind) ExactSequenceKind {
	switch kind {
	case parser.ExactSequenceChosenTypeLibraryTopToHand:
		return ExactSequenceChosenTypeLibraryTopToHand
	case parser.ExactSequenceBottomHandThenDraw:
		return ExactSequenceBottomHandThenDraw
	case parser.ExactSequenceDiscardHandThenDraw:
		return ExactSequenceDiscardHandThenDraw
	case parser.ExactSequenceConditionalLookAtTopReveal:
		return ExactSequenceConditionalLookAtTopReveal
	case parser.ExactSequenceConditionalLookAtTopBattlefield:
		return ExactSequenceConditionalLookAtTopBattlefield
	case parser.ExactSequenceDrawThenDiscardUnlessType:
		return ExactSequenceDrawThenDiscardUnlessType
	default:
		return ExactSequenceUnknown
	}
}
