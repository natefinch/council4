package game

import (
	"errors"

	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ConditionalDestinationPlace resolves the "look at or reveal a single card, then
// under a gate you may put it onto the battlefield, otherwise put it somewhere
// else" family (Risen Reef, Scholar of New Horizons, Lantern of Revealing) as one
// atomic step. The card is referenced (in practice a linked card a preceding look
// or reveal published), and rests in FromZone. When the gate holds and the
// controller chooses to, the card enters the battlefield, tapped when EntryTapped;
// otherwise it moves to Else (the bottom of Else when ElseBottom), a move the
// controller may decline when ElseOptional, leaving the card in FromZone.
//
// The single primitive exists because the gated put and its fallback cannot be
// composed from a gated optional instruction plus a result-gated move: a skipped
// instruction publishes no result, so a "card stayed" fallback keyed on that
// result would never fire. Carrying the gate and the fallback together lets the
// handler choose battlefield-or-else without an observable intermediate result.
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
	Else          zone.Type
	ElseBottom    bool
	ElseOptional  bool
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
	if p.Else == zone.None || p.Else == zone.Battlefield || p.Else == zone.Stack {
		return errors.New("conditional destination place requires a non-battlefield else zone")
	}
	if p.ElseBottom && p.Else != zone.Library {
		return errors.New("conditional destination place bottom placement requires a library else zone")
	}
	return nil
}

func (ConditionalDestinationPlace) validateCapturedTargetControllerReferences(_ []TargetSpec, _ bool) error {
	return nil
}
