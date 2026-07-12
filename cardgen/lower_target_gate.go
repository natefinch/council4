package cardgen

import (
	"github.com/natefinch/council4/mtg/game"
)

// castBranchGate classifies how an instruction's effect condition gates the
// instruction on the resolving spell's cast branch. gated reports whether the
// instruction resolves only on one gift or kicker branch; when gated is true,
// gate names that branch. ok is false when the condition mentions a cast-branch
// state the assigner cannot cleanly map to a single gate (both gift and kicker,
// for example), so the caller fails closed rather than approximate.
func castBranchGate(inst *game.Instruction) (gate game.TargetGate, gated bool, ok bool) {
	if !inst.Condition.Exists {
		return game.TargetGateAlways, false, true
	}
	condition := inst.Condition.Val.Condition
	if !condition.Exists {
		return game.TargetGateAlways, false, true
	}
	cond := condition.Val
	giftGated := cond.GiftPromised
	kickerGated := cond.SpellWasKicked
	switch {
	case giftGated && kickerGated:
		return game.TargetGateAlways, false, false
	case giftGated:
		if cond.Negate {
			return game.TargetGateGiftNotPromised, true, true
		}
		return game.TargetGateGiftPromised, true, true
	case kickerGated:
		if cond.Negate {
			return game.TargetGateSpellNotKicked, true, true
		}
		return game.TargetGateSpellKicked, true, true
	default:
		return game.TargetGateAlways, false, true
	}
}

// cardIndexToTargetSlot maps a card-domain target index (the runtime numbers card
// references by target-slot width among card-allowing specs only) back to its
// TargetSpec slice slot. Like the object domain, a spec that admits up to N cards
// occupies N consecutive card indices, so the mapping accumulates each
// card-allowing spec's width. It reports false for an index past the last
// card-allowing spec.
func cardIndexToTargetSlot(targets []game.TargetSpec, index int) (int, bool) {
	if index < 0 {
		return 0, false
	}
	cumulative := 0
	for i := range targets {
		if targets[i].Allow&game.TargetAllowCard == 0 {
			continue
		}
		width := max(targets[i].MaxTargets, 1)
		if index < cumulative+width {
			return i, true
		}
		cumulative += width
	}
	return 0, false
}

// assignTargetGates gates each of a spell mode's target specs on the cast branch
// its referencing instructions require, so a target used only by a gift-promised
// or kicked clause participates in announcement only on that branch (CR 601.2c).
// It walks every instruction's target references in both numbering domains,
// records the gate under which each referenced spec is used, and assigns:
//   - TargetGateAlways to a spec any ungated instruction references (a target the
//     spell always needs), and to a spec no instruction references;
//   - the single cast-branch gate to a spec referenced only under that one gate.
//
// It returns ok false — leaving the card to fail closed — when a target-bearing
// primitive's references cannot be expressed, when a spec is referenced under two
// different cast-branch gates (which one gate field cannot represent), or when a
// referenced index falls outside the spec list. Non-gift, non-kicker spells never
// set a cast-branch condition, so every spec resolves to TargetGateAlways and the
// spec list is returned unchanged.
func assignTargetGates(targets []game.TargetSpec, sequence []game.Instruction) ([]game.TargetSpec, bool) {
	if len(targets) == 0 {
		return targets, true
	}
	// Classify every instruction's cast-branch gate first. A spell with no
	// cast-branch-gated instruction (the overwhelming majority — every non-gift,
	// non-kicker card) needs no gating and no target-reference walk, so it is
	// returned untouched and stays byte-identical. Only when a gated instruction
	// is present does the reference walk run, whose failure then fails the card
	// closed rather than approximate the gating.
	gatesByInst := make([]instructionGate, len(sequence))
	anyGated := false
	for i := range sequence {
		gate, gated, ok := castBranchGate(&sequence[i])
		if !ok {
			return nil, false
		}
		gatesByInst[i] = instructionGate{gate: gate, gated: gated}
		if gated {
			anyGated = true
		}
	}
	if !anyGated {
		return targets, true
	}
	slots := make([]slotGates, len(targets))
	for i := range sequence {
		inst := &sequence[i]
		gate := gatesByInst[i]
		objectSlots := map[int]bool{}
		cardSlots := map[int]bool{}
		referencedTarget := false
		record := func(kind targetIndexKind, idx int) (int, bool) {
			referencedTarget = true
			if kind == targetIndexCard {
				cardSlots[idx] = true
			} else {
				objectSlots[idx] = true
			}
			return idx, true
		}
		_, wok := transformPrimitiveTargetIndices(inst.Primitive, record)
		if !wok {
			if referencedTarget {
				// The walker reached at least one target reference in this
				// primitive but could not fully express it. Because a
				// cast-branch-gated instruction is present in this sequence, a
				// target-bearing primitive we cannot analyze is unsafe in both
				// directions: a gated instruction whose target we drop would
				// leave that target unconditionally required (announced on every
				// cast), and an ungated instruction whose always-required
				// reference we drop would wrongly gate an always-required spec
				// (dropping it on the base branch). Fail closed rather than emit
				// an approximate gate assignment.
				return nil, false
			}
			// The primitive references no announced target (a "scry 2" gated on
			// "if this spell was kicked", or an each/all group form). It cannot
			// affect any spec's gate, so skip it; the sequence's other
			// instructions still drive gating exactly as before.
			continue
		}
		for objIndex := range objectSlots {
			slot, ok := objectIndexToTargetSlot(targets, objIndex)
			if !ok {
				return nil, false
			}
			slots[slot].record(gate.gate, gate.gated)
		}
		for cardIndex := range cardSlots {
			slot, ok := cardIndexToTargetSlot(targets, cardIndex)
			if !ok {
				return nil, false
			}
			slots[slot].record(gate.gate, gate.gated)
		}
	}
	out := append([]game.TargetSpec(nil), targets...)
	changed := false
	for i := range out {
		gate, ok := slots[i].resolve()
		if !ok {
			return nil, false
		}
		if gate == game.TargetGateAlways {
			continue
		}
		out[i].Gate = gate
		changed = true
	}
	if !changed {
		return targets, true
	}
	return out, true
}

// instructionGate carries an instruction's classified cast-branch gate through
// the two-pass gate assignment: the first pass classifies, the second maps
// references to spec slots only when at least one instruction is gated.
type instructionGate struct {
	gate  game.TargetGate
	gated bool
}

// slotGates accumulates the cast-branch gates under which a single target spec is
// referenced across a spell's instructions. always records any reference by an
// ungated instruction; mask records the set of distinct cast-branch gates.
type slotGates struct {
	always bool
	mask   uint8
}

const (
	slotGateKicked uint8 = 1 << iota
	slotGateNotKicked
	slotGatePromised
	slotGateNotPromised
)

// record folds one referencing instruction's gate into the accumulator. An
// ungated reference marks the spec always-required; a cast-branch-gated reference
// records that gate in the mask.
func (s *slotGates) record(gate game.TargetGate, gated bool) {
	if !gated {
		s.always = true
		return
	}
	switch gate {
	case game.TargetGateSpellKicked:
		s.mask |= slotGateKicked
	case game.TargetGateSpellNotKicked:
		s.mask |= slotGateNotKicked
	case game.TargetGateGiftPromised:
		s.mask |= slotGatePromised
	case game.TargetGateGiftNotPromised:
		s.mask |= slotGateNotPromised
	default:
	}
}

// resolve reduces a spec's accumulated references to the single TargetGate the
// spec should carry, reporting ok false when no single gate can express the set.
// A spec referenced by any ungated instruction, referenced across both branches
// of a mechanic (kicked and not-kicked, or promised and not-promised — the "deals
// X, or Y instead" pattern in which the target is required either way), or
// referenced by nothing resolves to TargetGateAlways. A spec referenced under
// exactly one cast-branch gate carries that gate. A spec referenced under gates
// from two different mechanics (kicked and promised, say) cannot be expressed by a
// single gate field, so it fails closed.
func (s *slotGates) resolve() (game.TargetGate, bool) {
	kickerFull := s.mask&slotGateKicked != 0 && s.mask&slotGateNotKicked != 0
	giftFull := s.mask&slotGatePromised != 0 && s.mask&slotGateNotPromised != 0
	if s.always || kickerFull || giftFull || s.mask == 0 {
		return game.TargetGateAlways, true
	}
	hasKicker := s.mask&(slotGateKicked|slotGateNotKicked) != 0
	hasGift := s.mask&(slotGatePromised|slotGateNotPromised) != 0
	if hasKicker && hasGift {
		return game.TargetGateAlways, false
	}
	switch s.mask {
	case slotGateKicked:
		return game.TargetGateSpellKicked, true
	case slotGateNotKicked:
		return game.TargetGateSpellNotKicked, true
	case slotGatePromised:
		return game.TargetGateGiftPromised, true
	case slotGateNotPromised:
		return game.TargetGateGiftNotPromised, true
	default:
		return game.TargetGateAlways, false
	}
}
