package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
)

func TestCompiledSkippedTokenProducerCannotReusePreviousActivation(t *testing.T) {
	t.Parallel()
	def := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Conditional Product", Layout: "normal", TypeLine: "Enchantment",
		OracleText: "{1}: If you have no cards in hand, create a 1/1 green Elf creature token. Tap it.",
	})
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, def)
	engine := NewEngine(nil)
	resolve := func() {
		obj := &game.StackObject{
			ID: g.IDGen.Next(), Kind: game.StackActivatedAbility, Controller: game.Player1,
			SourceID: source.ObjectID, SourceCardID: source.CardInstanceID,
		}
		engine.resolveAbilityContentWithChoices(g, obj, def.ActivatedAbilities[0].Content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	}
	resolve()
	var token *game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.Token {
			if token != nil {
				t.Fatal("single producer created multiple tokens")
			}
			token = permanent
		}
	}
	if token == nil || !token.Tapped {
		t.Fatal("successful producer's own token was not tapped")
	}
	token.Tapped = false
	addCardToHand(g, game.Player1, evidenceCard("Nonempty", 1))
	resolve()
	if token.Tapped || len(g.Battlefield) != 2 {
		t.Fatal("skipped producer affected a previous activation's product or created a new one")
	}
}
