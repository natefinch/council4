package cardgen

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestLibraryCardProducerSemanticRoundTrip(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping generated-package semantic round-trip in short mode")
	}
	cards := []*ScryfallCard{
		{
			Name: "RT Library Spell", Layout: "normal", TypeLine: "Sorcery",
			OracleText: "Look at the top card of your library.",
		},
		{
			Name: "RT Library Trigger", Layout: "normal", TypeLine: "Enchantment",
			OracleText: "At the beginning of your upkeep, you may reveal the top card of your library.",
		},
		{
			Name: "RT Library Player", Layout: "normal", TypeLine: "Artifact",
			OracleText: "{T}: Look at the top card of target opponent's library.",
		},
		{
			Name: "RT Library Modes", Layout: "normal", TypeLine: "Sorcery",
			OracleText: "Choose one \u2014\n\u2022 Reveal the top card of your library.\n\u2022 You may look at the top card of your library.",
		},
		{
			Name: "RT Library Consumption", Layout: "normal", TypeLine: "Sorcery",
			OracleText: "Reveal the top card of your library. Scry 1. If it's a creature card, put it onto the battlefield tapped under your control. Otherwise, put it into your hand.",
		},
		{
			Name: "RT Library Independent", Layout: "normal", TypeLine: "Artifact",
			OracleText: "{T}: Look at the top card of your library. If it's a land card, put it into your hand. Reveal the top card of your library. Scry 1. If it's a Zombie card, you gain 2 life.",
		},
		{
			Name: "RT Library Characteristics", Layout: "normal", TypeLine: "Sorcery",
			OracleText: "Reveal the top card of your library. If it's a creature card, you draw cards equal to its power and you gain life equal to its toughness.",
		},
		{
			Name: "RT Library Payment", Layout: "normal", TypeLine: "Artifact",
			OracleText: "{T}: Look at the top card of target player's library. If it's a nonland card, you may pay 2 life. If you do, put it into that player's graveyard.",
		},
	}
	dir, pkg := writeCardRoundTripPackage(t, cards)
	source := fmt.Sprintf(`package %s

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestActualObservationPrimitives(t *testing.T) {
	spell := RTLibrarySpell().SpellAbility.Val.Modes[0].Sequence[0]
	look, ok := spell.Primitive.(game.LookAtLibraryTop)
	if !ok || look.Player != game.ControllerReference() || look.PublishLinked == "" || spell.Optional {
		t.Fatalf("spell observation = %%#v", spell)
	}
	trigger := RTLibraryTrigger().TriggeredAbilities[0]
	reveal, ok := trigger.Content.Modes[0].Sequence[0].Primitive.(game.Reveal)
	if !ok || reveal.Amount.Value() != 1 || reveal.Player != game.ControllerReference() ||
		reveal.PublishLinked != "" || !trigger.Optional {
		t.Fatalf("trigger observation = %%#v", trigger)
	}
	mode := RTLibraryPlayer().ActivatedAbilities[0].Content.Modes[0]
	look, ok = mode.Sequence[0].Primitive.(game.LookAtLibraryTop)
	if !ok || look.Player != game.TargetPlayerReference(0) || look.PublishLinked == "" || len(mode.Targets) != 1 {
		t.Fatalf("target-player observation = %%#v", mode)
	}
	modes := RTLibraryModes().SpellAbility.Val
	if len(modes.Modes) != 2 || modes.MinModes != 1 || modes.MaxModes != 1 {
		t.Fatalf("mode selection = %%#v", modes)
	}
	reveal, ok = modes.Modes[0].Sequence[0].Primitive.(game.Reveal)
	if !ok || reveal.Amount.Value() != 1 || reveal.PublishLinked != "" {
		t.Fatalf("reveal mode = %%#v", modes.Modes[0])
	}
	look, ok = modes.Modes[1].Sequence[0].Primitive.(game.LookAtLibraryTop)
	if !ok || look.PublishLinked == "" || !modes.Modes[1].Sequence[0].Optional {
		t.Fatalf("optional look mode = %%#v", modes.Modes[1])
	}
	sequence := RTLibraryConsumption().SpellAbility.Val.Modes[0].Sequence
	reveal, ok = sequence[0].Primitive.(game.Reveal)
	if !ok || reveal.PublishLinked != "sequence-effect-0-product" || len(sequence) != 4 {
		t.Fatalf("conditional observation sequence = %%#v", sequence)
	}
	if !sequence[0].LocalProducts.HasLink(reveal.PublishLinked) {
		t.Fatal("generated observation lost its typed local namespace declaration")
	}
	battlefield, ok := sequence[2].Primitive.(game.PutOnBattlefield)
	if !ok || !battlefield.EntryTapped || !battlefield.Recipient.Exists ||
		battlefield.Recipient.Val != game.ControllerReference() {
		t.Fatalf("battlefield consumer = %%#v", sequence[2])
	}
	if key, ok := battlefield.Source.LinkedKey(); !ok || key != reveal.PublishLinked {
		t.Fatal("battlefield consumer did not consume the exact observation")
	}
	hand, ok := sequence[3].Primitive.(game.MoveCard)
	if !ok || hand.FromZone != zone.Library || hand.Destination != zone.Hand ||
		hand.Card.LinkID != string(reveal.PublishLinked) || sequence[3].ConditionGate == "" ||
		!sequence[3].ConditionGateNegate {
		t.Fatalf("otherwise consumer = %%#v", sequence[3])
	}
	sequence = RTLibraryIndependent().ActivatedAbilities[0].Content.Modes[0].Sequence
	look, ok = sequence[0].Primitive.(game.LookAtLibraryTop)
	if !ok || look.PublishLinked != "sequence-effect-0-product" {
		t.Fatalf("first independent observation = %%#v", sequence[0])
	}
	reveal, ok = sequence[2].Primitive.(game.Reveal)
	if !ok || reveal.PublishLinked != "sequence-effect-2-product" || reveal.PublishLinked == look.PublishLinked {
		t.Fatalf("second independent observation = %%#v", sequence[2])
	}
	if !sequence[0].LocalProducts.HasLink(look.PublishLinked) ||
		!sequence[2].LocalProducts.HasLink(reveal.PublishLinked) {
		t.Fatal("generated independent producers lost their separate local namespaces")
	}
	if !sequence[4].Condition.Exists || !sequence[4].Condition.Val.Condition.Exists ||
		!sequence[4].Condition.Val.Condition.Val.Object.Exists ||
		sequence[4].Condition.Val.Condition.Val.Object.Val.LinkID() != string(reveal.PublishLinked) {
		t.Fatalf("independent condition did not consume its exact producer = %%#v", sequence[4])
	}
	sequence = RTLibraryCharacteristics().SpellAbility.Val.Modes[0].Sequence
	reveal = sequence[0].Primitive.(game.Reveal)
	outputs := reveal.PublishCharacteristics
	if outputs.Power != "sequence-effect-0-power" || outputs.Toughness != "sequence-effect-0-toughness" ||
		!sequence[0].LocalProducts.HasResult(outputs.Power) || !sequence[0].LocalProducts.HasResult(outputs.Toughness) {
		t.Fatalf("generated characteristic outputs = %%#v", sequence[0])
	}
	draw := sequence[1].Primitive.(game.Draw).Amount.DynamicAmount()
	gain := sequence[2].Primitive.(game.GainLife).Amount.DynamicAmount()
	if !draw.Exists || !gain.Exists ||
		draw.Val.Kind != game.DynamicAmountPreviousEffectResult || draw.Val.ResultKey != outputs.Power ||
		gain.Val.Kind != game.DynamicAmountPreviousEffectResult || gain.Val.ResultKey != outputs.Toughness ||
		!sequence[1].ResultGate.Val.AmountAvailable || sequence[1].ResultGate.Val.Key != outputs.Power ||
		!sequence[2].ResultGate.Val.AmountAvailable || sequence[2].ResultGate.Val.Key != outputs.Toughness {
		t.Fatal("generated consumers lost their separate scalar availability gates")
	}
	sequence = RTLibraryPayment().ActivatedAbilities[0].Content.Modes[0].Sequence
	pay := sequence[1].Primitive.(game.Pay)
	if sequence[1].Optional || len(pay.Payment.AdditionalCosts) != 1 ||
		pay.Payment.AdditionalCosts[0].Kind != cost.AdditionalPayLife || pay.Payment.AdditionalCosts[0].Amount != 2 ||
		!pay.Payment.Payer.Exists || pay.Payment.Payer.Val != game.ControllerReference() ||
		sequence[2].ResultGate.Val.Key != sequence[1].PublishResult || sequence[2].ResultGate.Val.Succeeded != game.TriTrue {
		t.Fatalf("generated payment confused the payer, decision, or success = %%#v", sequence)
	}
}
`, pkg)
	if err := os.WriteFile(filepath.Join(dir, "semantic_test.go"), []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.CommandContext(context.Background(), "go", "test", "-count=1", "./"+filepath.Base(dir))
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("generated observation semantic round-trip: %v\n%s", err, out)
	}
}
