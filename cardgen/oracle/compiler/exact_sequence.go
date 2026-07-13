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
	// ExactSequencePayHandSizeOrCantAttack is the triggered "that opponent may
	// pay {X}, where X is the number of cards in their hand. If they don't, they
	// can't attack you this combat." punisher (Champions of Minas Tirith): the
	// triggering opponent may pay generic mana equal to their hand size and, on
	// non-payment, their creatures can't attack the source's controller for the
	// rest of that combat. The whole body is fixed, so it carries no extra data.
	ExactSequencePayHandSizeOrCantAttack
	// ExactSequenceExtraDrawThenPayLifeOrTop is the triggered draw-step "you may
	// draw N additional cards. If you do, choose M cards in your hand drawn this
	// turn. For each of those cards, pay L life or put the card on top of your
	// library." sequence (Sylvan Library): the controller may draw N extra cards
	// and, if they do, choose M of the cards drawn this turn still in hand and
	// for each pay L life to keep it or put it on top of their library. The
	// counts N, M, and L travel on the compiled ability (ExactSequenceDrawCount,
	// ExactSequenceChooseCount, ExactSequencePayLife) so lowering models the
	// sequence without reading Oracle words.
	ExactSequenceExtraDrawThenPayLifeOrTop
	// ExactSequenceBargainSearchCastPayoff is the spell "Search your library for a
	// card, exile it face down, then shuffle. If this spell was bargained, you may
	// cast the exiled card without paying its mana cost if that spell's mana value
	// is N or less. Put the exiled card into your hand if it wasn't cast this
	// way." sequence (Beseech the Mirror): search the library and exile one card
	// face down, then when the spell was bargained optionally cast the exiled card
	// for free if its mana value is at most ExactSequenceMaxManaValue, otherwise
	// put it into hand. The bound N travels on the compiled ability
	// (ExactSequenceMaxManaValue) so lowering models the sequence without reading
	// Oracle words.
	ExactSequenceBargainSearchCastPayoff
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
	case parser.ExactSequencePayHandSizeOrCantAttack:
		return ExactSequencePayHandSizeOrCantAttack
	case parser.ExactSequenceExtraDrawThenPayLifeOrTop:
		return ExactSequenceExtraDrawThenPayLifeOrTop
	case parser.ExactSequenceBargainSearchCastPayoff:
		return ExactSequenceBargainSearchCastPayoff
	default:
		return ExactSequenceUnknown
	}
}
