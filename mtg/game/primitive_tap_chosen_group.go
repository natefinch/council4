package game

import "errors"

// TapChosenGroup lets the resolving controller choose any number of permanents
// from ChooseFrom — a group the enclosing ability restricts to the untapped
// permanents they control that match a subtype or other selection — and taps
// each chosen permanent. The number chosen and tapped is published under
// PublishCount as a ResolutionChoiceNumber result so later instructions read it
// through DynamicAmountChosenNumber ("this creature gets +X/+0 until end of turn
// and deals X damage to the player or planeswalker it's attacking", Myr
// Battlesphere).
//
// Choosing zero publishes zero and reports the instruction as not succeeded, so
// a reflexive "If you do" payoff gated on the choice resolves to nothing. The
// choice is made when this instruction resolves, not when the enabling ability
// triggers. Because tapping this way is a resolution action rather than a {T}
// cost, summoning sickness never restricts which matching permanents may be
// tapped, and the ability's own source may be among them when it is untapped and
// matches the selection.
type TapChosenGroup struct {
	ChooseFrom   GroupReference
	PublishCount ResultKey
	Prompt       string
}

// Kind implements Primitive for TapChosenGroup.
func (TapChosenGroup) Kind() PrimitiveKind { return PrimitiveTapChosenGroup }

func (TapChosenGroup) isPrimitive() {}

func (p TapChosenGroup) instructionRefs() primitiveRefs {
	return primitiveRefs{publishesChoice: ChoiceKey(p.PublishCount)}
}

func (p TapChosenGroup) validatePrimitive(_ []TargetSpec, _ bool) error {
	if p.PublishCount == "" {
		return errors.New("TapChosenGroup requires a published count key")
	}
	if !p.ChooseFrom.Valid() {
		return errors.New("TapChosenGroup requires a candidate group")
	}
	return nil
}
