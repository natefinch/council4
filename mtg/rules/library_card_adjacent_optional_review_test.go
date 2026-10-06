package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
)

func TestCompiledAdjacentOptionalObservationCharacteristic(t *testing.T) {
	for _, producer := range []struct {
		name, verb, property string
	}{
		{"reveal toughness", "reveal", "toughness"},
		{"look power", "look at", "power"},
	} {
		t.Run(producer.name, func(t *testing.T) {
			def := compileUnlessCard(t, cardgen.ScryfallCard{
				Name: "Adjacent Optional Characteristic", Layout: "normal", TypeLine: "Sorcery",
				OracleText: "You may " + producer.verb + " the top card of your library. If it's a creature card, you gain life equal to its " + producer.property + ". You gain 1 life.",
			})
			sequence := def.SpellAbility.Val.Modes[0].Sequence
			if len(sequence) != 3 || !sequence[0].Optional || sequence[1].Optional || sequence[2].Optional {
				t.Fatalf("optional observation changed mandatory consumers/rider: %#v", sequence)
			}
			outputs := libraryCardCharacteristicOutputs(sequence[0].Primitive)
			key := outputs.Toughness
			if producer.property == "power" {
				key = outputs.Power
			}
			gain, ok := sequence[1].Primitive.(game.GainLife)
			if !ok || key == "" || !sequence[0].LocalProducts.HasResult(key) ||
				!sequence[1].ResultGate.Exists || !sequence[1].ResultGate.Val.AmountAvailable ||
				sequence[1].ResultGate.Val.Key != key || !gain.Amount.DynamicAmount().Exists ||
				gain.Amount.DynamicAmount().Val.ResultKey != key {
				t.Fatalf("consumer lacks its exact characteristic/availability: %#v", sequence)
			}
			for _, outcome := range []string{"accepted", "declined", "empty", "noncreature", "unavailable", "zero"} {
				t.Run(outcome, func(t *testing.T) {
					g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
					g.Players[game.Player1].Life = 40
					power, toughness := "2", "3"
					if outcome == "zero" {
						power, toughness = "0", "0"
					}
					if outcome == "unavailable" {
						power, toughness = "*", "*"
					}
					card := compileUnlessCard(t, cardgen.ScryfallCard{
						Name: "Observed Creature", Layout: "normal", TypeLine: "Creature — Horse",
						ManaCost: "{1}{B}", Power: &power, Toughness: &toughness,
					})
					amount := 3
					if producer.property == "power" {
						amount = 2
					}
					if outcome == "zero" {
						amount = 0
					}
					if outcome == "noncreature" {
						card = compileUnlessCard(t, cardgen.ScryfallCard{
							Name: "Divination", Layout: "normal", TypeLine: "Sorcery",
							ManaCost: "{2}{U}", OracleText: "Draw two cards.",
						})
					}
					observed := addCardToLibrary(g, game.Player1, card)
					if outcome == "empty" {
						g.Players[game.Player1].Library.Remove(observed)
					}
					agent := &libraryPaymentAgent{accept: outcome != "declined"}
					obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1}
					NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.SpellAbility.Val,
						[game.NumPlayers]PlayerAgent{game.Player1: agent}, &TurnLog{})
					want := 41
					if outcome == "accepted" || outcome == "zero" {
						want += amount
					}
					if g.Players[game.Player1].Life != want || len(agent.prompts) != 1 {
						t.Errorf("life=%d prompts=%d, want %d/1", g.Players[game.Player1].Life, len(agent.prompts), want)
					}
					if g.Players[game.Player1].Hand.Size() != 0 || len(obj.ResolvedAmounts) != 0 {
						t.Error("scalar consumer moved a card or leaked its sequence-local scalar")
					}
					revealed := outcome != "declined" && outcome != "empty" && producer.verb == "reveal"
					if eventRevealedCard(g, observed, obj.ID) != revealed {
						t.Error("public reveal did not match actual optional observation")
					}
				})
			}
		})
	}
}
