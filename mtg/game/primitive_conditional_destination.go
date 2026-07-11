package game

import (
	"errors"

	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ConditionalDestinationPlace resolves the "look at or reveal a single card, then
// under a gate you may put it into a chosen zone, otherwise put it somewhere
// else" family (Risen Reef, Scholar of New Horizons, Lantern of Revealing, and
// the conditional look-at-top reveal of Sarinth Steelseeker, Traveling Botanist,
// and Territory Culler) as one atomic step. The card is referenced (in practice a
// linked card a preceding look or reveal published), and rests in FromZone. When
// the gate holds and the controller chooses to, the card moves to the "then"
// destination — the battlefield (tapped when EntryTapped) by default, or the Then
// zone when set, revealed first when ThenReveal; otherwise it moves to Else (the
// bottom of Else when ElseBottom), a move the controller may decline when
// ElseOptional, leaving the card in FromZone. A zone.None Else means there is no
// fallback at all: when the put is not made the card is simply left in FromZone
// (the "look at the top card; if it's a land card, you may put it onto the
// battlefield" family with no trailing else clause).
//
// The single primitive exists because the gated put and its fallback cannot be
// composed from a gated optional instruction plus a result-gated move: a skipped
// instruction publishes no result, so a "card stayed" fallback keyed on that
// result would never fire. Carrying the gate and the fallback together lets the
// handler choose then-or-else without an observable intermediate result.
//
// The gate combines an optional card-characteristic test (CardCondition, matched
// against the referenced card) and an optional board-state test (Condition); an
// absent test holds vacuously, and both must hold for the put to be offered.
type ConditionalDestinationPlace struct {
	Card          CardReference
	FromZone      zone.Type
	CardCondition opt.V[CardSelection]
	Condition     opt.V[EffectCondition]
	EntryTapped   bool
	// ThenMandatory performs the matching "then" move without offering a choice.
	// It backs exact routing text such as "put it onto the battlefield tapped if
	// it's a land card. Otherwise, put it into your hand."
	ThenMandatory bool
	// Then is the destination when the gate holds and the controller accepts.
	// The zero value (zone.None) keeps the historical battlefield put; any other
	// non-battlefield zone moves the card there instead. ThenReveal first reveals
	// the card publicly, backing "reveal it and put it into your hand".
	Then         zone.Type
	ThenReveal   bool
	Else         zone.Type
	ElseBottom   bool
	ElseOptional bool
}

// Kind implements Primitive for ConditionalDestinationPlace.
func (ConditionalDestinationPlace) Kind() PrimitiveKind { return PrimitiveConditionalDestinationPlace }

func (ConditionalDestinationPlace) isPrimitive() {}

func (p ConditionalDestinationPlace) instructionRefs() primitiveRefs {
	return cardReferenceRefs(p.Card)
}

func (p ConditionalDestinationPlace) validatePrimitive(_ []TargetSpec, _ bool) error {
	if p.Card.Kind != CardReferenceLinked || p.Card.LinkID == "" {
		return errors.New("conditional destination place requires a linked card reference")
	}
	if p.FromZone == zone.None || p.FromZone == zone.Battlefield || p.FromZone == zone.Stack {
		return errors.New("conditional destination place requires a non-battlefield source zone")
	}
	if p.Else == zone.Battlefield || p.Else == zone.Stack {
		return errors.New("conditional destination place else zone must be no fallback (zone.None) or a non-battlefield destination")
	}
	if p.Then == zone.Battlefield || p.Then == zone.Stack {
		return errors.New("conditional destination place then zone must be the implicit battlefield (zone.None) or a non-battlefield destination")
	}
	if p.ThenReveal && p.Then == zone.None {
		return errors.New("conditional destination place then reveal requires a non-battlefield then zone")
	}
	if p.ElseBottom && p.Else != zone.Library {
		return errors.New("conditional destination place bottom placement requires a library else zone")
	}
	return nil
}

func (ConditionalDestinationPlace) validateCapturedTargetControllerReferences(_ []TargetSpec, _ bool) error {
	return nil
}
