package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func olorinSearingLightCard() *ScryfallCard {
	return &ScryfallCard{
		Name:          "Olórin's Searing Light",
		Layout:        "normal",
		ManaCost:      "{2}{R}",
		TypeLine:      "Instant",
		ColorIdentity: []string{"R"},
		OracleText: "Each opponent exiles a creature with the greatest power among creatures that player controls.\n" +
			"Spell mastery — If there are two or more instant and/or sorcery cards in your graveyard, Olórin's Searing Light deals damage to each opponent equal to the power of the creature they exiled.",
	}
}

func TestLowerOlorinsSearingLightCorrelatedOpponentChoices(t *testing.T) {
	face := lowerSingleFace(t, olorinSearingLightCard())
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	sequence := face.SpellAbility.Val.Modes[0].Sequence
	if len(sequence) != 2 {
		t.Fatalf("sequence = %#v, want exile and damage", sequence)
	}
	exile, ok := sequence[0].Primitive.(game.ExileForEachOpponent)
	if !ok {
		t.Fatalf("first primitive = %#v", sequence[0].Primitive)
	}
	if exile.Chooser.Kind() != game.PlayerReferenceGroupOfferMember ||
		!exile.Required ||
		exile.Extremum != game.PermanentChoiceGreatestPower ||
		!exile.Simultaneous ||
		len(exile.Selection.RequiredTypes) != 1 ||
		exile.Selection.RequiredTypes[0] != types.Creature {
		t.Fatalf("exile = %#v", exile)
	}
	damage, ok := sequence[1].Primitive.(game.Damage)
	if !ok || !sequence[1].ForEachPlayerGroup.Exists {
		t.Fatalf("damage instruction = %#v", sequence[1])
	}
	dynamic := damage.Amount.DynamicAmount()
	if !dynamic.Exists ||
		dynamic.Val.Kind != game.DynamicAmountObjectPower ||
		dynamic.Val.Object.Kind() != game.ObjectReferenceLinkedObject ||
		dynamic.Val.Player == nil ||
		dynamic.Val.Player.Kind() != game.PlayerReferenceGroupOfferMember {
		t.Fatalf("damage amount = %#v", damage.Amount)
	}
	recipient, ok := damage.Recipient.PlayerReference()
	if !ok || recipient.Kind() != game.PlayerReferenceGroupOfferMember {
		t.Fatalf("damage recipient = %#v", damage.Recipient)
	}
	if !sequence[1].Condition.Exists ||
		!sequence[1].Condition.Val.Condition.Exists ||
		sequence[1].Condition.Val.Condition.Val.ControllerGraveyardInstantOrSorceryCountAtLeast != 2 {
		t.Fatalf("damage gate = %#v", sequence[1].Condition)
	}
}

func TestGenerateOlorinsSearingLightExecutableSource(t *testing.T) {
	source, diagnostics, err := GenerateExecutableCardSource(olorinSearingLightCard(), "o")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.ExileForEachOpponent{",
		"Chooser: game.GroupOfferMemberReference()",
		"Extremum: game.PermanentChoiceGreatestPower",
		"Simultaneous: true",
		"ForEachPlayerGroup: opt.Val(game.OpponentsReference())",
		"Object: game.LinkedObjectReference(\"correlated-opponent-exile\")",
		"Player: func() *game.PlayerReference { ref := game.GroupOfferMemberReference(); return &ref }()",
	} {
		if !strings.Contains(spaceCollapsed(source), spaceCollapsed(want)) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}
