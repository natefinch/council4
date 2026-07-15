package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func estridsInvocationCard() *ScryfallCard {
	return &ScryfallCard{
		Name:     "Estrid's Invocation",
		Layout:   "normal",
		ManaCost: "{2}{U}",
		TypeLine: "Enchantment",
		Colors:   []string{"U"},
		OracleText: "You may have this enchantment enter as a copy of an enchantment you control, " +
			"except it has \"At the beginning of your upkeep, you may exile this enchantment. " +
			"If you do, return it to the battlefield under its owner's control.\"",
	}
}

// TestLowerEstridsInvocationEntersAsCopyGrantedAbility proves Estrid's Invocation
// lowers to an optional self enters-as-copy replacement over "an enchantment you
// control" that carries the printed upkeep self-blink ability as a copiable
// granted-ability rider (CR 706.2).
func TestLowerEstridsInvocationEntersAsCopyGrantedAbility(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, estridsInvocationCard())
	if len(face.ReplacementAbilities) != 1 {
		t.Fatalf("replacement abilities = %d, want 1", len(face.ReplacementAbilities))
	}
	replacement := face.ReplacementAbilities[0].Replacement
	if !replacement.EntersAsCopy {
		t.Fatal("replacement is not an enters-as-copy replacement")
	}
	if !replacement.EntersAsCopyOptional {
		t.Fatal("enters-as-copy must be optional (\"You may have ...\")")
	}
	if replacement.EntersAsCopySelection == nil {
		t.Fatal("enters-as-copy selection is nil")
	}
	selection := replacement.EntersAsCopySelection
	if selection.Controller != game.ControllerYou {
		t.Fatalf("selection controller = %v, want ControllerYou", selection.Controller)
	}
	if len(selection.RequiredTypes) != 1 || selection.RequiredTypes[0] != types.Enchantment {
		t.Fatalf("selection required types = %v, want [Enchantment]", selection.RequiredTypes)
	}
	if len(replacement.EntersAsCopyAddAbilities) != 1 {
		t.Fatalf("granted abilities = %d, want 1", len(replacement.EntersAsCopyAddAbilities))
	}
	trigger, ok := replacement.EntersAsCopyAddAbilities[0].(*game.TriggeredAbility)
	if !ok {
		t.Fatalf("granted ability = %T, want *game.TriggeredAbility", replacement.EntersAsCopyAddAbilities[0])
	}
	if trigger.Trigger.Type != game.TriggerAt ||
		trigger.Trigger.Pattern.Event != game.EventBeginningOfStep ||
		trigger.Trigger.Pattern.Step != game.StepUpkeep ||
		trigger.Trigger.Pattern.Controller != game.TriggerControllerYou {
		t.Fatalf("granted trigger = %#v, want your upkeep beginning", trigger.Trigger)
	}
	sequence := trigger.Content.Modes[0].Sequence
	if len(sequence) != 2 {
		t.Fatalf("granted ability sequence = %d instructions, want 2", len(sequence))
	}
	exile, ok := sequence[0].Primitive.(game.Exile)
	if !ok {
		t.Fatalf("sequence[0] = %T, want game.Exile", sequence[0].Primitive)
	}
	if !sequence[0].Optional {
		t.Fatal("self-exile must be optional (\"you may exile\")")
	}
	if exile.ExileLinkedKey == "" {
		t.Fatal("self-exile must carry a linked key so the return can find the same card")
	}
	put, ok := sequence[1].Primitive.(game.PutOnBattlefield)
	if !ok {
		t.Fatalf("sequence[1] = %T, want game.PutOnBattlefield", sequence[1].Primitive)
	}
	linkedKey, ok := put.Source.LinkedKey()
	if !ok {
		t.Fatal("return source is not the linked exiled card (must return under owner's control)")
	}
	if linkedKey != exile.ExileLinkedKey {
		t.Fatalf("return linked key = %q, want the exile's key %q", linkedKey, exile.ExileLinkedKey)
	}
}

// TestGenerateExecutableEstridsInvocation proves the full Estrid's Invocation
// card generates clean executable source wiring the granted upkeep blink onto
// the enters-as-copy replacement.
func TestGenerateExecutableEstridsInvocation(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(estridsInvocationCard(), "e")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.EntersAsCopyWithAddedAbilities(",
		"game.EntersAsCopyReplacement(",
		"types.Enchantment",
		"game.ControllerYou",
		"game.TriggeredAbility{",
		"game.StepUpkeep",
		"game.Exile{",
		"game.PutOnBattlefield{",
		"game.LinkedBattlefieldSource(",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestLowerImmediateSelfBlinkReturnStandalone proves the printed "at the
// beginning of your upkeep, you may exile this permanent; if you do, return it
// to the battlefield under its owner's control" self-blink upkeep trigger lowers
// standalone (the same shape Estrid's Invocation grants onto its copy), exercising
// lowerImmediateSelfBlinkReturn directly rather than through the copy rider.
func TestLowerImmediateSelfBlinkReturnStandalone(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Test Self Blinker",
		Layout:   "normal",
		ManaCost: "{2}{U}",
		TypeLine: "Enchantment",
		Colors:   []string{"U"},
		OracleText: "At the beginning of your upkeep, you may exile this enchantment. " +
			"If you do, return it to the battlefield under its owner's control.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	trigger := face.TriggeredAbilities[0]
	if trigger.Trigger.Type != game.TriggerAt ||
		trigger.Trigger.Pattern.Event != game.EventBeginningOfStep ||
		trigger.Trigger.Pattern.Step != game.StepUpkeep ||
		trigger.Trigger.Pattern.Controller != game.TriggerControllerYou {
		t.Fatalf("trigger = %#v, want your upkeep beginning", trigger.Trigger)
	}
	sequence := trigger.Content.Modes[0].Sequence
	if len(sequence) != 2 {
		t.Fatalf("sequence = %d instructions, want 2", len(sequence))
	}
	exile, ok := sequence[0].Primitive.(game.Exile)
	if !ok {
		t.Fatalf("sequence[0] = %T, want game.Exile", sequence[0].Primitive)
	}
	if !sequence[0].Optional || exile.ExileLinkedKey == "" {
		t.Fatalf("self-exile must be optional and linked, got %#v", sequence[0])
	}
	put, ok := sequence[1].Primitive.(game.PutOnBattlefield)
	if !ok {
		t.Fatalf("sequence[1] = %T, want game.PutOnBattlefield", sequence[1].Primitive)
	}
	linkedKey, ok := put.Source.LinkedKey()
	if !ok || linkedKey != exile.ExileLinkedKey {
		t.Fatalf("return must reuse the exile's linked key, got %#v (exile key %q)", put.Source, exile.ExileLinkedKey)
	}
}
