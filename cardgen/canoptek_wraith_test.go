package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestLowerCanoptekWraithExactOracle(t *testing.T) {
	t.Parallel()
	const oracle = "Wraith Form — This creature can't be blocked.\nTransdimensional Scout — When this creature deals combat damage to a player, you may pay {3} and sacrifice it. If you do, choose a land you control. Then search your library for up to two basic land cards which have the same name as the chosen land, put them onto the battlefield tapped, then shuffle."
	card := &ScryfallCard{
		Name:       "Canoptek Wraith",
		Layout:     "normal",
		TypeLine:   "Artifact Creature — Wraith",
		OracleText: oracle,
		Power:      new("2"),
		Toughness:  new("1"),
		ManaCost:   "{3}",
	}
	source, sourceDiagnostics, err := GenerateExecutableCardSource(card, "c")
	if err != nil || len(sourceDiagnostics) != 0 {
		t.Fatalf("source generation: err = %v, diagnostics = %#v", err, sourceDiagnostics)
	}
	for line := range strings.SplitSeq(oracle, "\n") {
		if !strings.Contains(source, line) {
			t.Fatalf("generated source did not preserve Oracle line %q", line)
		}
	}

	faces, diagnostics := lowerExecutableFaces(card)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	face := faces[0]
	if len(face.StaticAbilities) != 1 ||
		face.StaticAbilities[0].Body.RuleEffects[0].Kind != game.RuleEffectCantBeBlocked {
		t.Fatalf("face = %#v", face)
	}
	trigger := face.TriggeredAbilities[0]
	if trigger.Trigger.Pattern.Event != game.EventDamageDealt ||
		trigger.Trigger.Pattern.Source != game.TriggerSourceSelf ||
		!trigger.Trigger.Pattern.RequireCombatDamage ||
		trigger.Trigger.Pattern.DamageRecipient != game.DamageRecipientPlayer {
		t.Fatalf("trigger = %#v", trigger.Trigger)
	}
	sequence := trigger.Content.Modes[0].Sequence
	if len(sequence) != 3 {
		t.Fatalf("sequence = %#v, want pay, choose, search", sequence)
	}
	pay, ok := sequence[0].Primitive.(game.Pay)
	if !ok {
		t.Fatalf("first primitive = %T, want game.Pay", sequence[0].Primitive)
	}
	if !pay.Payment.ManaCost.Exists ||
		len(pay.Payment.AdditionalCosts) != 1 ||
		pay.Payment.AdditionalCosts[0].Kind != cost.AdditionalSacrificeSource {
		t.Fatalf("payment = %#v", pay.Payment)
	}
	choose, ok := sequence[1].Primitive.(game.Choose)
	if !ok {
		t.Fatalf("second primitive = %T, want game.Choose", sequence[1].Primitive)
	}
	if choose.Choice.Kind != game.ResolutionChoicePermanent ||
		choose.PublishChoice != game.ResolutionChosenPermanentChoiceKey ||
		choose.Choice.Selection == nil ||
		len(choose.Choice.Selection.RequiredTypes) != 1 ||
		choose.Choice.Selection.RequiredTypes[0] != types.Land {
		t.Fatalf("choice = %#v", choose)
	}
	search, ok := sequence[2].Primitive.(game.Search)
	if !ok {
		t.Fatalf("third primitive = %T, want game.Search", sequence[2].Primitive)
	}
	if search.Amount.Value() != 2 ||
		search.Spec.NameFromChoice != game.ResolutionChosenPermanentChoiceKey ||
		search.Spec.Destination != zone.Battlefield ||
		!search.Spec.EntersTapped {
		t.Fatalf("search = %#v", search)
	}
}
