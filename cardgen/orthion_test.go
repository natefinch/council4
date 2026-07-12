package cardgen

import (
	"reflect"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func orthionCard() *ScryfallCard {
	return &ScryfallCard{
		Name:       "Orthion, Hero of Lavabrink",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Human Soldier",
		ManaCost:   "{3}{R}",
		Power:      new("3"),
		Toughness:  new("3"),
		OracleText: "{1}{R}, {T}: Create a token that's a copy of another target creature you control. It gains haste. Sacrifice it at the beginning of the next end step. Activate only as a sorcery.\n{6}{R}{R}{R}, {T}: Create five tokens that are copies of another target creature you control. They gain haste. Sacrifice them at the beginning of the next end step. Activate only as a sorcery.",
	}
}

func TestLowerOrthionDelayedCopyTokenSacrifice(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, orthionCard())
	if len(face.ActivatedAbilities) != 2 {
		t.Fatalf("activated abilities = %d, want 2", len(face.ActivatedAbilities))
	}
	for i, wantAmount := range []int{1, 5} {
		mode := face.ActivatedAbilities[i].Content.Modes[0]
		create, ok := mode.Sequence[0].Primitive.(game.CreateToken)
		if !ok || create.Amount.Value() != wantAmount || create.PublishLinked == "" {
			t.Fatalf("ability %d create = %#v, want %d published tokens", i, mode.Sequence[0].Primitive, wantAmount)
		}
		delayed, ok := mode.Sequence[1].Primitive.(game.CreateDelayedTrigger)
		if !ok || delayed.Trigger.Timing != game.DelayedAtBeginningOfNextEndStep {
			t.Fatalf("ability %d delayed = %#v", i, mode.Sequence[1].Primitive)
		}
		sacrifice, ok := delayed.Trigger.Content.Modes[0].Sequence[0].Primitive.(game.Sacrifice)
		if !ok {
			t.Fatalf("ability %d cleanup = %#v, want Sacrifice", i, delayed.Trigger.Content)
		}
		if wantAmount == 1 {
			if !delayed.Trigger.CapturedObject.Exists ||
				delayed.Trigger.CapturedObjectGroup.Exists ||
				sacrifice.Object != game.CapturedObjectReference() {
				t.Fatalf("single-token cleanup = %#v", delayed.Trigger)
			}
			continue
		}
		if delayed.Trigger.CapturedObject.Exists ||
			!delayed.Trigger.CapturedObjectGroup.Exists ||
			!reflect.DeepEqual(sacrifice.Group, game.CapturedObjectsGroup()) {
			t.Fatalf("multi-token cleanup = %#v", delayed.Trigger)
		}
	}
}

func TestGenerateOrthionCapturedGroupCleanup(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(orthionCard(), "o")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"CapturedObjectGroup: opt.Val(game.LinkedObjectReference(\"delayed-sacrifice-1\"))",
		"Group: game.CapturedObjectsGroup()",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}
