package rules

import (
	"strconv"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestCompiledCountedSubtypeUnionPreservesCreatureControllerAndOR(t *testing.T) {
	for _, count := range []int{0, 1, 2} {
		t.Run(strconv.Itoa(count), func(t *testing.T) {
			def := compileTargetCountSpell(t, "For each creature you control that's a Wolf or a Werewolf, create a 2/2 green Wolf creature token.")
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			for i := range count {
				subtype := types.Sub("Wolf")
				if i == 1 {
					subtype = "Werewolf"
				}
				addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
					Name: "Matched Creature", Types: []types.Card{types.Creature}, Subtypes: []types.Sub{subtype},
				}})
			}
			for _, excluded := range []struct {
				controller game.PlayerID
				cardType   types.Card
				subtype    types.Sub
			}{
				{game.Player2, types.Creature, "Wolf"},
				{game.Player1, types.Creature, "Human"},
				{game.Player1, types.Artifact, "Werewolf"},
			} {
				addCombatPermanent(g, excluded.controller, &game.CardDef{CardFace: game.CardFace{
					Name: "Excluded Object", Types: []types.Card{excluded.cardType}, Subtypes: []types.Sub{excluded.subtype},
				}})
			}
			spell := addCardToHand(g, game.Player1, def)
			g.Players[game.Player1].Hand.Remove(spell)
			g.Stack.Push(&game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: spell, Controller: game.Player1})
			NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
			tokens := 0
			for _, permanent := range g.Battlefield {
				if permanent.Token && permanent.Controller == game.Player1 {
					tokens++
				}
			}
			if tokens != count {
				t.Fatalf("created tokens=%d, want %d; OR must not become AND or count another controller/noncreature", tokens, count)
			}
		})
	}
}
