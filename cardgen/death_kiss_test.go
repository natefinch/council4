package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLowerDeathKissAbilities(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Death Kiss",
		Layout:   "normal",
		TypeLine: "Creature — Beholder",
		OracleText: "Whenever a creature an opponent controls attacks one of your opponents, double its power until end of turn.\n" +
			"{X}{X}{R}: Monstrosity X.\n" +
			"When this creature becomes monstrous, goad up to X target creatures your opponents control.",
		Power:     new("5"),
		Toughness: new("5"),
	})
	if len(face.TriggeredAbilities) != 2 || len(face.ActivatedAbilities) != 1 {
		t.Fatalf("abilities = %d triggered, %d activated", len(face.TriggeredAbilities), len(face.ActivatedAbilities))
	}

	attack := face.TriggeredAbilities[0]
	pattern := attack.Trigger.Pattern
	if pattern.Event != game.EventAttackerDeclared ||
		pattern.Controller != game.TriggerControllerOpponent ||
		pattern.Player != game.TriggerPlayerOpponent ||
		pattern.AttackRecipient != game.AttackRecipientPlayer ||
		len(pattern.SubjectSelection.RequiredTypes) != 1 ||
		pattern.SubjectSelection.RequiredTypes[0] != types.Creature {
		t.Fatalf("attack pattern = %#v", pattern)
	}
	apply, ok := attack.Content.Modes[0].Sequence[0].Primitive.(game.ApplyContinuous)
	if !ok || apply.Object.Val != game.EventPermanentReference() ||
		apply.Duration != game.DurationUntilEndOfTurn ||
		len(apply.ContinuousEffects) != 1 ||
		!apply.ContinuousEffects[0].DoublePower ||
		apply.ContinuousEffects[0].DoubleToughness {
		t.Fatalf("attack effect = %#v", attack.Content.Modes[0].Sequence)
	}

	monstrous := face.TriggeredAbilities[1]
	if monstrous.Trigger.Pattern.Event != game.EventPermanentBecameMonstrous ||
		monstrous.Trigger.Pattern.Source != game.TriggerSourceSelf {
		t.Fatalf("monstrous pattern = %#v", monstrous.Trigger.Pattern)
	}
	mode := monstrous.Content.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("goad targets = %#v", mode.Targets)
	}
	target := mode.Targets[0]
	if target.MinTargets != 0 ||
		!target.MaxTargetsFromTriggerEventX ||
		!target.Selection.Exists ||
		target.Selection.Val.Controller != game.ControllerOpponent ||
		len(target.Selection.Val.RequiredTypesAny) != 1 ||
		target.Selection.Val.RequiredTypesAny[0] != types.Creature {
		t.Fatalf("goad target = %#v", target)
	}
	goad, ok := mode.Sequence[0].Primitive.(game.Goad)
	if !ok || goad.Object != game.AllTargetPermanentsReference(0) {
		t.Fatalf("goad primitive = %#v", mode.Sequence[0].Primitive)
	}
}

func TestGenerateDeathKissSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Death Kiss",
		Layout:   "normal",
		TypeLine: "Creature — Beholder",
		OracleText: "Whenever a creature an opponent controls attacks one of your opponents, double its power until end of turn.\n" +
			"{X}{X}{R}: Monstrosity X.\n" +
			"When this creature becomes monstrous, goad up to X target creatures your opponents control.",
		Power:     new("5"),
		Toughness: new("5"),
	}, "d")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.EventPermanentBecameMonstrous",
		"MaxTargetsFromTriggerEventX: true",
		"game.AllTargetPermanentsReference(0)",
		"game.DynamicAmountX",
		"DoublePower: true",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}
