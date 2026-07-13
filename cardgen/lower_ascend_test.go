package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerAscendPermanentKeyword proves the permanent form of ascend
// (CR 702.131b) lowers to the reusable AscendStaticBody static ability alongside
// the rest of the face, the shape Snubhorn Sentry uses ("Ascend" plus a
// city's-blessing continuous static).
func TestLowerAscendPermanentKeyword(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Snubhorn Sentry",
		Layout:   "normal",
		TypeLine: "Creature — Dinosaur Soldier",
		OracleText: "Ascend (If you control ten or more permanents, you get the city's blessing for the rest of the game.)\n" +
			"As long as you have the city's blessing, Snubhorn Sentry gets +3/+0.",
		Power:     new("1"),
		Toughness: new("1"),
	})
	if len(face.StaticAbilities) != 2 {
		t.Fatalf("got %d static abilities, want 2", len(face.StaticAbilities))
	}
	if got := face.StaticAbilities[0].VarName; got != "game.AscendStaticBody" {
		t.Fatalf("ascend static VarName = %q, want game.AscendStaticBody", got)
	}
	if face.SpellAbility.Exists {
		t.Fatal("permanent ascend produced a spell ability")
	}
}

// TestLowerAscendCombatGuard proves Wayward Swordtooth lowers fully: the ascend
// keyword, the extra land-play static, and the "can't attack or block unless you
// have the city's blessing" combat guard all lower without an unsupported
// diagnostic.
func TestLowerAscendCombatGuard(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Wayward Swordtooth",
		Layout:   "normal",
		TypeLine: "Creature — Dinosaur",
		OracleText: "Ascend (If you control ten or more permanents, you get the city's blessing for the rest of the game.)\n" +
			"You may play an additional land on each of your turns.\n" +
			"Wayward Swordtooth can't attack or block unless you have the city's blessing.",
		Power:     new("5"),
		Toughness: new("5"),
	})
	if got := face.StaticAbilities[0].VarName; got != "game.AscendStaticBody" {
		t.Fatalf("ascend static VarName = %q, want game.AscendStaticBody", got)
	}
	var guard *game.StaticAbility
	for i := range face.StaticAbilities {
		body := face.StaticAbilities[i].Body
		for _, effect := range body.RuleEffects {
			if effect.Kind == game.RuleEffectCantAttack {
				guard = &face.StaticAbilities[i].Body
			}
		}
	}
	if guard == nil {
		t.Fatal("combat guard did not lower to a can't-attack rule effect")
	}
	if !guard.Condition.Exists {
		t.Fatal("combat guard has no city's-blessing condition")
	}
	if !guard.Condition.Val.ControllerHasCityBlessing || !guard.Condition.Val.Negate {
		t.Fatalf("combat guard condition = %+v, want negated ControllerHasCityBlessing", guard.Condition.Val)
	}
}

// TestLowerAscendSpell proves the spell form of ascend (CR 702.131a) lowers to a
// GainCityBlessing instruction prepended to the spell's sequence so the blessing
// is gained before the spell's other instructions resolve (Golden Demise).
func TestLowerAscendSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Golden Demise",
		Layout:   "normal",
		TypeLine: "Sorcery",
		OracleText: "Ascend (If you control ten or more permanents, you get the city's blessing for the rest of the game.)\n" +
			"All creatures get -2/-2 until end of turn.",
	})
	for i := range face.StaticAbilities {
		if ascendStaticAbility(face.StaticAbilities[i]) {
			t.Fatal("spell ascend left a permanent ascend static ability")
		}
	}
	if !face.SpellAbility.Exists {
		t.Fatal("spell ascend produced no spell ability")
	}
	modes := face.SpellAbility.Val.Modes
	if len(modes) == 0 || len(modes[0].Sequence) < 2 {
		t.Fatalf("spell sequence too short: %+v", modes)
	}
	first := modes[0].Sequence[0].Primitive
	if first == nil || first.Kind() != game.PrimitiveGainCityBlessing {
		t.Fatalf("first spell instruction = %T, want GainCityBlessing", first)
	}
}
