package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
)

// castBranchForCast derives the cast branch a spell is being cast on from its
// cast action: whether its Gift keyword action promised a gift and whether its
// kicker cost was paid. The branch decides which cast-branch-gated target specs
// participate in target announcement, validation, and counting (CR 601.2c).
func castBranchForCast(cast action.CastSpellAction) game.CastBranch {
	return game.CastBranch{
		GiftPromised: cast.GiftPromised,
		Kicked:       effectiveKickerCount(cast.KickerPaid, cast.KickerCount) > 0,
		Bargained:    cast.Bargained,
		Bestowed:     cast.Bestowed,
		Offspring:    cast.Offspring,
	}
}

// castBranchForObject derives the cast branch of a spell already on the stack
// from the choices captured when it was cast. A copy inherits these fields, so a
// copy or retarget resolves the same gated specs the original cast did.
func castBranchForObject(obj *game.StackObject) game.CastBranch {
	if obj == nil {
		return game.CastBranch{}
	}
	return game.CastBranch{
		GiftPromised: obj.GiftPromised,
		Kicked:       effectiveKickerCount(obj.KickerPaid, obj.KickerCount) > 0,
		Bargained:    obj.Bargained,
		Bestowed:     obj.Bestowed,
		Offspring:    obj.OffspringPaid,
	}
}

// specsHaveGate reports whether any spec carries a non-default cast-branch gate.
// It is the fast-path guard that keeps ungated spells (every existing card) on
// the identity path through applyCastBranchToSpecs and the target-slot remap.
func specsHaveGate(specs []game.TargetSpec) bool {
	for i := range specs {
		if specs[i].Gate != game.TargetGateAlways {
			return true
		}
	}
	return false
}

// applyCastBranchToSpecs returns the spec list a spell announces, validates, and
// counts targets against on the given cast branch. A spec whose gate is inactive
// for the branch is replaced by an empty spec that reserves its ordinal slot but
// requires no target: enumeration offers it zero targets, counting records zero
// for it, and its chosen-target slot is skipped. A spec whose gate is active, or
// that has no gate, is returned unchanged. Ungated spell lists are returned as-is
// so every existing card enumerates and validates byte-identically.
func applyCastBranchToSpecs(specs []game.TargetSpec, branch game.CastBranch) []game.TargetSpec {
	if !specsHaveGate(specs) {
		return specs
	}
	out := make([]game.TargetSpec, len(specs))
	for i := range specs {
		if specs[i].Gate.ActiveIn(branch) {
			out[i] = specs[i]
			continue
		}
		out[i] = game.TargetSpec{Gate: specs[i].Gate}
	}
	return out
}

// gatedTargetSlot maps a compile-time target-slot index to the position of that
// spec's chosen target within a stack object's compacted Targets slice. Chosen
// targets omit inactive gated specs, whose reserved compile-time width therefore
// disappears from the compacted slice, so the position shifts down by the total
// width of every inactive gated spec whose slot range precedes the requested
// index. With no gated specs the mapping is the identity, so every existing card
// indexes its targets unchanged.
func gatedTargetSlot(specs []game.TargetSpec, branch game.CastBranch, index int) int {
	compileStart := 0
	removed := 0
	for j := range specs {
		width := specTargetWidth(specs[j])
		if index < compileStart+width {
			break
		}
		if specs[j].Gate != game.TargetGateAlways && !specs[j].Gate.ActiveIn(branch) {
			removed += width
		}
		compileStart += width
	}
	return index - removed
}

// specTargetWidth is the number of consecutive flat target slots a spec reserves
// in the compile-time target-index space: one per possible target.
func specTargetWidth(spec game.TargetSpec) int {
	if spec.MaxTargets > 1 {
		return spec.MaxTargets
	}
	return 1
}

// remapTargetSlot maps a resolving spell's compile-time target-slot index to the
// position of that target within obj.Targets, accounting for inactive gated specs
// that reserved no chosen-target slot. Spells without gated targets, and every
// non-spell stack object, take the identity fast path.
func remapTargetSlot(g *game.Game, obj *game.StackObject, index int) int {
	if obj == nil || obj.Kind != game.StackSpell {
		return index
	}
	specs, ok := spellRawTargetSpecsForObject(g, obj)
	if !ok || !specsHaveGate(specs) {
		return index
	}
	return gatedTargetSlot(specs, castBranchForObject(obj), index)
}

// gatedCardSlot is the card-reference-domain analogue of gatedTargetSlot. Card
// references number targets among card-allowing specs only (a spec admitting N
// cards occupies N consecutive card indices), so a card-reference index shifts
// down by the card-slot width of every inactive gated card-allowing spec whose
// card-index range precedes it. With no gated card specs the mapping is the
// identity, so every existing card resolves its card references unchanged.
func gatedCardSlot(specs []game.TargetSpec, branch game.CastBranch, cardIndex int) int {
	compileStart := 0
	removed := 0
	for j := range specs {
		if specs[j].Allow&game.TargetAllowCard == 0 {
			continue
		}
		width := specTargetWidth(specs[j])
		if cardIndex < compileStart+width {
			break
		}
		if specs[j].Gate != game.TargetGateAlways && !specs[j].Gate.ActiveIn(branch) {
			removed += width
		}
		compileStart += width
	}
	return cardIndex - removed
}

// remapCardTargetSlot maps a resolving spell's compile-time card-reference index
// to the card-reference index within obj.Targets, accounting for inactive gated
// card-allowing specs that reserved no chosen-target slot. It mirrors
// remapTargetSlot for the card-reference numbering domain; spells without gated
// targets, and every non-spell stack object, take the identity fast path.
func remapCardTargetSlot(g *game.Game, obj *game.StackObject, cardIndex int) int {
	if obj == nil || obj.Kind != game.StackSpell {
		return cardIndex
	}
	specs, ok := spellRawTargetSpecsForObject(g, obj)
	if !ok || !specsHaveGate(specs) {
		return cardIndex
	}
	return gatedCardSlot(specs, castBranchForObject(obj), cardIndex)
}

// spellRawTargetSpecsForObject returns the full, ungated target-spec list of a
// resolving spell (gates intact), used only to compute the target-slot remap.
func spellRawTargetSpecsForObject(g *game.Game, obj *game.StackObject) ([]game.TargetSpec, bool) {
	if obj.Kind != game.StackSpell {
		return nil, false
	}
	card, ok := stackObjectSpellDef(g, obj)
	if !ok {
		return nil, false
	}
	if obj.Overloaded && card.Overload.Exists {
		card = overloadSpellDef(card)
	}
	return spellTargetSpecsRaw(card, obj.ChosenModes), true
}

// stackObjectSpellDef returns the announced face definition of a spell on the
// stack, matching the lookup stackObjectTargetSpecs uses for the spell case.
func stackObjectSpellDef(g *game.Game, obj *game.StackObject) (*game.CardDef, bool) {
	if card, ok := g.GetCardInstance(obj.SourceID); ok {
		if def, defOK := card.Def.FaceDef(obj.Face); defOK {
			return def, true
		}
	}
	return stackObjectSourceDef(g, obj)
}
