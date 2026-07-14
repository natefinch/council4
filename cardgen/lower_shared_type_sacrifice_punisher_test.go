package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// braidsScryfallCard is Braids, Arisen Nightmare as the compiler pipeline sees
// it, used by the lowering tests below.
func braidsScryfallCard() *ScryfallCard {
	ptr := func(s string) *string { return &s }
	return &ScryfallCard{
		Name:      "Braids, Arisen Nightmare",
		Layout:    "normal",
		TypeLine:  "Legendary Creature — Nightmare",
		ManaCost:  "{1}{B}{B}",
		Power:     ptr("3"),
		Toughness: ptr("3"),
		OracleText: "At the beginning of your end step, you may sacrifice an artifact, " +
			"creature, enchantment, land, or planeswalker. If you do, each opponent may " +
			"sacrifice a permanent of their choice that shares a card type with it. For each " +
			"opponent who doesn't, that player loses 2 life and you draw a card.",
	}
}

// TestLowerSharedTypeSacrificePunisher proves Braids, Arisen Nightmare lowers to
// the generic optional-sacrifice-then-punisher pair: an optional controller
// sacrifice of one of the five permanent types that publishes the sacrificed
// permanent as a linked object and its success as a result, followed by an
// each-opponent shared-card-type punisher gated on that success, where each
// punished opponent lets the controller draw a card.
func TestLowerSharedTypeSacrificePunisher(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, braidsScryfallCard())
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	ability := face.TriggeredAbilities[0]
	if got := ability.Trigger.Pattern.Step; got != game.StepEnd {
		t.Fatalf("trigger step = %v, want StepEnd", got)
	}
	if got := ability.Trigger.Pattern.Controller; got != game.TriggerControllerYou {
		t.Fatalf("trigger controller = %v, want TriggerControllerYou", got)
	}
	mode := ability.Content.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence len = %d, want 2", len(mode.Sequence))
	}

	sacrifice, ok := mode.Sequence[0].Primitive.(game.SacrificePermanents)
	if !ok {
		t.Fatalf("instruction 0 = %#v, want SacrificePermanents", mode.Sequence[0].Primitive)
	}
	if !mode.Sequence[0].Optional {
		t.Fatal("controller sacrifice must be optional (\"you may sacrifice\")")
	}
	if mode.Sequence[0].PublishResult == "" {
		t.Fatal("controller sacrifice must publish a result for the punisher gate")
	}
	if sacrifice.PublishLinked == "" {
		t.Fatal("controller sacrifice must publish the sacrificed permanent as a linked object")
	}
	wantTypes := []types.Card{types.Artifact, types.Creature, types.Enchantment, types.Land, types.Planeswalker}
	if got := sacrifice.Selection.RequiredTypesAny; !equalCardTypes(got, wantTypes) {
		t.Fatalf("sacrifice required types = %v, want %v", got, wantTypes)
	}

	punisher, ok := mode.Sequence[1].Primitive.(game.PunisherEachLoseLife)
	if !ok {
		t.Fatalf("instruction 1 = %#v, want PunisherEachLoseLife", mode.Sequence[1].Primitive)
	}
	if !mode.Sequence[1].ResultGate.Exists ||
		mode.Sequence[1].ResultGate.Val.Key != mode.Sequence[0].PublishResult ||
		mode.Sequence[1].ResultGate.Val.Succeeded != game.TriTrue {
		t.Fatalf("punisher gate = %#v, want gated on the sacrifice succeeding", mode.Sequence[1].ResultGate)
	}
	if !punisher.AllowSacrifice {
		t.Fatal("punisher must allow the shared-card-type sacrifice alternative")
	}
	if punisher.SacrificeSelection.SharesCardTypeFromLinked != sacrifice.PublishLinked {
		t.Fatalf("punisher sacrifice selection = %#v, want shares-card-type with the sacrificed permanent", punisher.SacrificeSelection)
	}
	if !punisher.ControllerDrawEach {
		t.Fatal("punisher must let the controller draw a card per punished opponent")
	}
}

// TestLowerSharedTypeSacrificePunisherFailsClosed proves altered wording is not
// lowered as Braids: a different life amount is neither recognized nor lowered
// into the shared-type punisher, so it fails closed rather than producing a
// mis-costed ability.
func TestLowerSharedTypeSacrificePunisherFailsClosed(t *testing.T) {
	t.Parallel()
	card := braidsScryfallCard()
	card.OracleText = "At the beginning of your end step, you may sacrifice an artifact, " +
		"creature, enchantment, land, or planeswalker. If you do, each opponent may " +
		"sacrifice a permanent of their choice that shares a card type with it. For each " +
		"opponent who doesn't, that player loses 3 life and you draw a card."
	face := lowerSingleFaceExpectingUnsupported(t, card)
	for _, ability := range face.TriggeredAbilities {
		for _, instr := range ability.Content.Modes[0].Sequence {
			if _, ok := instr.Primitive.(game.PunisherEachLoseLife); ok {
				t.Fatal("altered wording lowered a punisher instruction, want fail-closed")
			}
		}
	}
}
