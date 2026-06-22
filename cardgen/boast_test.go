package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestLowerGoldmawChampionBoast(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Goldmaw Champion",
		Layout:     "normal",
		TypeLine:   "Creature — Dwarf Warrior",
		OracleText: "Boast — {1}{W}: Tap target creature. (Activate only if this creature attacked this turn and only once each turn.)",
		Power:      new("2"),
		Toughness:  new("3"),
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want one Boast ability", len(face.ActivatedAbilities))
	}
	boast := face.ActivatedAbilities[0]
	if boast.Timing != game.OncePerTurn {
		t.Fatalf("timing = %v, want OncePerTurn", boast.Timing)
	}
	if !boast.ActivationCondition.Exists || !boast.ActivationCondition.Val.EventHistory.Exists {
		t.Fatalf("activation condition = %#v, want attacked-this-turn event history", boast.ActivationCondition)
	}
	hist := boast.ActivationCondition.Val.EventHistory.Val
	if hist.Pattern.Event != game.EventAttackerDeclared ||
		hist.Pattern.Source != game.TriggerSourceSelf ||
		hist.Window != game.EventHistoryCurrentTurn {
		t.Fatalf("event history = %#v, want self attacker-declared this turn", hist)
	}
}

func TestLowerBoastRejectsExplicitTimingOrCondition(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Malformed Boaster",
		Layout:     "normal",
		TypeLine:   "Creature — Dwarf Warrior",
		OracleText: "Boast — {1}{W}: Tap target creature. Activate only as a sorcery.",
		Power:      new("2"),
		Toughness:  new("3"),
	})
	if len(diagnostics) == 0 {
		t.Fatal("expected a diagnostic for a Boast ability with an explicit timing restriction")
	}
}
