package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// dieToExileReplacement returns the lone Damage/ModifyPT instruction and the
// CreateReplacement appended by the would-die exile rider for a single-target
// spell face, failing the test when the face did not lower to that exact shape.
func dieToExileReplacement(t *testing.T, card *ScryfallCard) (game.Mode, game.CreateReplacement) {
	t.Helper()
	face := lowerSingleFace(t, card)
	if !face.SpellAbility.Exists {
		t.Fatalf("%s produced no spell ability", card.Name)
	}
	modes := face.SpellAbility.Val.Modes
	if len(modes) != 1 {
		t.Fatalf("modes = %d, want 1", len(modes))
	}
	mode := modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %d, want 1", len(mode.Targets))
	}
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence length = %d, want 2 (effect + replacement)", len(mode.Sequence))
	}
	create, ok := mode.Sequence[1].Primitive.(game.CreateReplacement)
	if !ok {
		t.Fatalf("second instruction = %#v, want CreateReplacement", mode.Sequence[1].Primitive)
	}
	return mode, create
}

func assertDieToExileRedirect(t *testing.T, create game.CreateReplacement) {
	t.Helper()
	if create.Object != game.TargetPermanentReference(0) {
		t.Fatalf("replacement object = %#v, want target permanent 0", create.Object)
	}
	if create.Duration != game.DurationThisTurn {
		t.Fatalf("duration = %v, want this turn", create.Duration)
	}
	r := create.Replacement
	if r == nil {
		t.Fatal("replacement effect is nil")
	}
	if r.MatchEvent != game.EventZoneChanged {
		t.Fatalf("match event = %v, want zone changed", r.MatchEvent)
	}
	if !r.MatchFromZone || r.FromZone != zone.Battlefield {
		t.Fatalf("from-zone match = %v/%v, want battlefield", r.MatchFromZone, r.FromZone)
	}
	if !r.MatchToZone || r.ToZone != zone.Graveyard {
		t.Fatalf("to-zone match = %v/%v, want graveyard", r.MatchToZone, r.ToZone)
	}
	if r.ReplaceToZone != zone.Exile {
		t.Fatalf("replace-to-zone = %v, want exile", r.ReplaceToZone)
	}
}

// TestLowerLavaCoilDieToExile verifies the canonical single-target damage rider
// "deals N damage to target creature. If that creature would die this turn,
// exile it instead." lowers to the damage instruction followed by a
// battlefield-to-graveyard-to-exile replacement bound to the spell's target for
// the turn.
func TestLowerLavaCoilDieToExile(t *testing.T) {
	t.Parallel()
	mode, create := dieToExileReplacement(t, &ScryfallCard{
		Name:       "Lava Coil",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Lava Coil deals 4 damage to target creature. If that creature would die this turn, exile it instead.",
	})
	damage, ok := mode.Sequence[0].Primitive.(game.Damage)
	if !ok {
		t.Fatalf("first instruction = %#v, want Damage", mode.Sequence[0].Primitive)
	}
	if damage.Amount != game.Fixed(4) {
		t.Fatalf("damage amount = %#v, want 4", damage.Amount)
	}
	if damage.Recipient != game.AnyTargetDamageRecipient(0) {
		t.Fatalf("recipient = %#v, want any-target 0", damage.Recipient)
	}
	assertDieToExileRedirect(t, create)
}

// TestLowerObliteratingBoltDieToExile verifies the "target creature or
// planeswalker" variant lowers identically, redirecting either a creature or a
// planeswalker death to exile.
func TestLowerObliteratingBoltDieToExile(t *testing.T) {
	t.Parallel()
	mode, create := dieToExileReplacement(t, &ScryfallCard{
		Name:       "Obliterating Bolt",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Obliterating Bolt deals 4 damage to target creature or planeswalker. If that creature or planeswalker would die this turn, exile it instead.",
	})
	if _, ok := mode.Sequence[0].Primitive.(game.Damage); !ok {
		t.Fatalf("first instruction = %#v, want Damage", mode.Sequence[0].Primitive)
	}
	assertDieToExileRedirect(t, create)
}

// TestLowerBleedDryDieToExile verifies the rider generalizes beyond damage to a
// -X/-X modify-power/toughness spell ("Target creature gets -13/-13 until end of
// turn. If that creature would die this turn, exile it instead.").
func TestLowerBleedDryDieToExile(t *testing.T) {
	t.Parallel()
	mode, create := dieToExileReplacement(t, &ScryfallCard{
		Name:       "Bleed Dry",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Target creature gets -13/-13 until end of turn. If that creature would die this turn, exile it instead.",
	})
	if _, ok := mode.Sequence[0].Primitive.(game.ModifyPT); !ok {
		t.Fatalf("first instruction = %#v, want ModifyPT", mode.Sequence[0].Primitive)
	}
	assertDieToExileRedirect(t, create)
}

// TestLowerYamabushisFlameDamagedCreatureDieToExile verifies the burn variant
// "deals N damage to any target. If a creature dealt damage this way would die
// this turn, exile it instead." (Yamabushi's Flame, Demonfire) lowers to the
// damage instruction followed by a creature-gated
// battlefield-to-graveyard-to-exile replacement bound to the spell's single
// target, so an any-target spell that instead kills a player or planeswalker
// does not exile it.
func TestLowerYamabushisFlameDamagedCreatureDieToExile(t *testing.T) {
	t.Parallel()
	mode, create := dieToExileReplacement(t, &ScryfallCard{
		Name:       "Yamabushi's Flame",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Yamabushi's Flame deals 3 damage to any target. If a creature dealt damage this way would die this turn, exile it instead.",
	})
	damage, ok := mode.Sequence[0].Primitive.(game.Damage)
	if !ok {
		t.Fatalf("first instruction = %#v, want Damage", mode.Sequence[0].Primitive)
	}
	if damage.Recipient != game.AnyTargetDamageRecipient(0) {
		t.Fatalf("recipient = %#v, want any-target 0", damage.Recipient)
	}
	assertDieToExileRedirect(t, create)
	if !create.Replacement.AffectedObjectMustBeCreature {
		t.Fatal("burn variant should gate the exile redirect on the dying permanent being a creature")
	}
}

// TestLowerLavaCoilDieToExileHasNoCreatureGate proves the demonstrative "that
// creature" rider keeps its ungated redirect (its target is always a creature),
// so only the "a creature dealt damage this way" burn variant carries the gate.
func TestLowerLavaCoilDieToExileHasNoCreatureGate(t *testing.T) {
	t.Parallel()
	_, create := dieToExileReplacement(t, &ScryfallCard{
		Name:       "Lava Coil",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Lava Coil deals 4 damage to target creature. If that creature would die this turn, exile it instead.",
	})
	if create.Replacement.AffectedObjectMustBeCreature {
		t.Fatal("the \"that creature\" rider must not carry the creature gate")
	}
}

// group "each creature" form, which the single-object replacement cannot bind:
// "deals 3 damage to each creature. If a creature dealt damage this way would
// die this turn, exile it instead." must not produce a spell ability.
func TestLowerDieToExileRequiresSingleTarget(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Anger of the Gods",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Anger of the Gods deals 3 damage to each creature. If a creature dealt damage this way would die this turn, exile it instead.",
	})
}
