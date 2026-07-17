package cardgen

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

const commandeerOracle = "You may exile two blue cards from your hand rather than pay this spell's mana cost.\n" +
	"Gain control of target noncreature spell. You may choose new targets for it. " +
	"(If that spell is an artifact, enchantment, or planeswalker, the permanent enters under your control.)"

func TestLowerCommandeer(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Commandeer",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{5}{U}{U}",
		OracleText: commandeerOracle,
	})
	if len(face.AlternativeCosts) != 1 {
		t.Fatalf("alternative costs = %#v, want one", face.AlternativeCosts)
	}
	alternative := face.AlternativeCosts[0]
	if alternative.Label != "Exile 2 blue cards" ||
		len(alternative.AdditionalCosts) != 1 {
		t.Fatalf("alternative = %#v", alternative)
	}
	exile := alternative.AdditionalCosts[0]
	if exile.Kind != cost.AdditionalExile ||
		exile.Amount != 2 ||
		exile.Source != zone.Hand ||
		!exile.MatchCardColor ||
		exile.CardColor != color.Blue {
		t.Fatalf("pitch exile = %#v, want two blue cards from hand", exile)
	}

	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 || len(mode.Sequence) != 2 {
		t.Fatalf("mode = %#v, want one target and two instructions", mode)
	}
	target := mode.Targets[0]
	if target.Allow != game.TargetAllowStackObject ||
		!slices.Equal(target.Predicate.StackObjectKinds, []game.StackObjectKind{game.StackSpell}) ||
		!slices.Equal(target.Predicate.ExcludedSpellCardTypes, []types.Card{types.Creature}) {
		t.Fatalf("target = %#v, want noncreature spell", target)
	}
	change, ok := mode.Sequence[0].Primitive.(game.ChangeStackObjectController)
	if !ok || change.Object != game.TargetStackObjectReference(0) ||
		change.Controller != game.ControllerReference() {
		t.Fatalf("first instruction = %#v, want stack controller change", mode.Sequence[0])
	}
	retarget, ok := mode.Sequence[1].Primitive.(game.ChooseNewTargets)
	if !ok || retarget.Object != game.TargetStackObjectReference(0) || !mode.Sequence[1].Optional {
		t.Fatalf("second instruction = %#v, want optional retarget", mode.Sequence[1])
	}
}

func TestGenerateCommandeerSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:          "Commandeer",
		Layout:        "normal",
		TypeLine:      "Instant",
		ManaCost:      "{5}{U}{U}",
		OracleText:    commandeerOracle,
		Colors:        []string{"U"},
		ColorIdentity: []string{"U"},
	}, "c")
	if err != nil {
		t.Fatalf("GenerateExecutableCardSource: %v", err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v, want none", diagnostics)
	}
	for _, want := range []string{
		"game.ChangeStackObjectController{",
		"Object:     game.TargetStackObjectReference(0)",
		"Controller: game.ControllerReference()",
		"game.ChooseNewTargets{",
		"Optional: true",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source does not contain %q:\n%s", want, source)
		}
	}
}
