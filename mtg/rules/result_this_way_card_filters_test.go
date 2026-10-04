package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestResultThisWayCardCharacteristicFilters(t *testing.T) {
	for _, noun := range []string{"card", "red card", "noncreature card"} {
		t.Run(noun, func(t *testing.T) {
			defs, diagnostics, err := cardgen.CompileCardDefs(&cardgen.ScryfallCard{
				Name: "Card Filter Probe", Layout: "normal", TypeLine: "Sorcery",
				OracleText: "Discard a card. If a " + noun + " was discarded this way, you gain 3 life.",
			})
			if err != nil || len(diagnostics) != 0 || len(defs) != 1 {
				t.Fatalf("compile = %v, %v", diagnostics, err)
			}
			for _, branch := range []string{"matching", "nonmatching", "no card"} {
				t.Run(branch, func(t *testing.T) {
					g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
					cardType, cardColor := types.Land, color.Red
					if branch == "nonmatching" {
						cardType, cardColor = types.Creature, color.Blue
					}
					if branch != "no card" {
						addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
							Types: []types.Card{cardType}, Colors: []color.Color{cardColor},
						}})
					}
					NewEngine(nil).resolveAbilityContentWithChoices(g, &game.StackObject{Controller: game.Player1},
						defs[0].SpellAbility.Val, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
					want := 40
					if branch == "matching" || (noun == "card" && branch == "nonmatching") {
						want = 43
					}
					if got := g.Players[game.Player1].Life; got != want {
						t.Fatalf("life = %d, want %d", got, want)
					}
				})
			}
		})
	}
}
