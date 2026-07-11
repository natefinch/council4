package game

import "errors"

// ChooseDrawnPayLifeOrTop has Player choose ChooseCount cards in their hand that
// were drawn this turn and, for each chosen card, pay LifeCost life to keep it or
// put it on top of their library ("choose two cards in your hand drawn this turn.
// For each of those cards, pay 4 life or put the card on top of your library.",
// Sylvan Library). When fewer than ChooseCount qualifying cards are in hand, the
// player chooses as many as they can; a chosen card whose owner cannot pay
// LifeCost life is put on top of the library. The per-card pay-or-return choice
// is not expressible through separate instructions, so it resolves as one
// primitive.
type ChooseDrawnPayLifeOrTop struct {
	Player      PlayerReference
	ChooseCount int
	LifeCost    int
}

// Kind implements Primitive for ChooseDrawnPayLifeOrTop.
func (ChooseDrawnPayLifeOrTop) Kind() PrimitiveKind { return PrimitiveChooseDrawnPayLifeOrTop }

func (ChooseDrawnPayLifeOrTop) isPrimitive() {}

func (ChooseDrawnPayLifeOrTop) instructionRefs() primitiveRefs { return primitiveRefs{} }

func (p ChooseDrawnPayLifeOrTop) validatePrimitive(targets []TargetSpec, checkTargets bool) error {
	if p.ChooseCount <= 0 {
		return errors.New("ChooseDrawnPayLifeOrTop requires a positive ChooseCount")
	}
	if p.LifeCost < 0 {
		return errors.New("ChooseDrawnPayLifeOrTop requires a non-negative LifeCost")
	}
	return validatePlayerReference(p.Player, targets, checkTargets)
}
