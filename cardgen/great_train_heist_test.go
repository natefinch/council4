package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

const greatTrainHeistText = "Spree (Choose one or more additional costs.)\n" +
	"+ {2}{R} — Untap all creatures you control. If it's your combat phase, there is an additional combat phase after this phase.\n" +
	"+ {2} — Creatures you control get +1/+0 and gain first strike until end of turn.\n" +
	"+ {R} — Choose target opponent. Whenever a creature you control deals combat damage to that player this turn, create a tapped Treasure token."

func TestLowerGreatTrainHeistAllModes(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Great Train Heist",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{R}",
		OracleText: greatTrainHeistText,
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	content := face.SpellAbility.Val
	if content.MinModes != 1 || content.MaxModes != 3 || len(content.Modes) != 3 {
		t.Fatalf("modal content = %#v", content)
	}
	for i, want := range []int{3, 2, 1} {
		if !content.Modes[i].Cost.Exists || content.Modes[i].Cost.Val.ManaValue() != want {
			t.Fatalf("mode %d cost = %#v, want mana value %d", i+1, content.Modes[i].Cost, want)
		}
	}
	mode1 := content.Modes[0]
	if len(mode1.Targets) != 0 {
		t.Fatalf("mode 1 targets = %#v, want none", mode1.Targets)
	}
	if len(mode1.Sequence) != 2 {
		t.Fatalf("mode 1 sequence = %#v, want untap then extra combat", mode1.Sequence)
	}
	if _, ok := mode1.Sequence[0].Primitive.(game.Untap); !ok {
		t.Fatalf("mode 1 first primitive = %T, want Untap", mode1.Sequence[0].Primitive)
	}
	extra, ok := mode1.Sequence[1].Primitive.(game.AddExtraPhases)
	if !ok || !extra.Combat || extra.Main || !mode1.Sequence[1].Condition.Exists ||
		!mode1.Sequence[1].Condition.Val.Condition.Val.ControllerCombatPhase {
		t.Fatalf("mode 1 extra combat = %#v condition %#v", extra, mode1.Sequence[1].Condition)
	}
	mode2 := content.Modes[1]
	if len(mode2.Targets) != 0 {
		t.Fatalf("mode 2 targets = %#v, want none", mode2.Targets)
	}
	continuous, ok := mode2.Sequence[0].Primitive.(game.ApplyContinuous)
	if !ok || continuous.Duration != game.DurationUntilEndOfTurn || len(continuous.ContinuousEffects) != 2 {
		t.Fatalf("mode 2 primitive = %#v", mode2.Sequence)
	}
	mode3 := content.Modes[2]
	if len(mode3.Targets) != 1 || len(mode3.Sequence) != 1 {
		t.Fatalf("mode 3 = %#v", mode3)
	}
	delayed, ok := mode3.Sequence[0].Primitive.(game.CreateDelayedTrigger)
	if !ok || !delayed.Trigger.EventPlayer.Exists ||
		delayed.Trigger.EventPlayer.Val != game.TargetPlayerReference(0) ||
		!delayed.Trigger.EventPattern.Exists {
		t.Fatalf("mode 3 delayed trigger = %#v", mode3.Sequence)
	}
	body := delayed.Trigger.Content.Modes[0].Sequence
	if len(body) != 1 {
		t.Fatalf("delayed body = %#v", body)
	}
	create, ok := body[0].Primitive.(game.CreateToken)
	if !ok || !create.EntryTapped {
		t.Fatalf("delayed body primitive = %#v, want tapped token", body[0].Primitive)
	}
}

func TestGenerateGreatTrainHeistSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Great Train Heist",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{R}",
		OracleText: greatTrainHeistText,
	}, "g")
	if err != nil {
		t.Fatal(err)
	}

	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"MinModes: 1,",
		"MaxModes: 3,",
		"ControllerCombatPhase: true,",
		"EventPlayer: opt.Val(game.TargetPlayerReference(0)),",
		"EntryTapped: true,",
		"game.FirstStrike,",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}

func TestLowerTargetedPlayerDelayedCombatDamageTriggerOutsideModal(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Targeted Raid",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{R}",
		OracleText: "Choose target opponent. Whenever a creature you control deals combat damage to that player this turn, create a tapped Treasure token.",
	})
	content := face.SpellAbility.Val
	if len(content.Modes) != 1 || len(content.Modes[0].Targets) != 1 {
		t.Fatalf("content = %#v, want one mode with one player target", content)
	}
	delayed, ok := content.Modes[0].Sequence[0].Primitive.(game.CreateDelayedTrigger)
	if !ok || !delayed.Trigger.EventPlayer.Exists ||
		delayed.Trigger.EventPlayer.Val != game.TargetPlayerReference(0) {
		t.Fatalf("delayed trigger = %#v", content.Modes[0].Sequence)
	}
}
