package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestImpulseExileWithoutPayingManaCostMarksPlayPermission(t *testing.T) {
	for _, free := range []bool{false, true} {
		t.Run(map[bool]string{false: "normal", true: "free"}[free], func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			cardID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
				Name:     "Exiled Spell",
				Types:    []types.Card{types.Sorcery},
				ManaCost: opt.Val(cost.Mana{cost.O(5)}),
			}})
			obj := &game.StackObject{
				Kind:         game.StackTriggeredAbility,
				Controller:   game.Player1,
				SourceID:     g.IDGen.Next(),
				SourceCardID: g.IDGen.Next(),
			}
			resolveInstruction(NewEngine(nil), g, obj, game.ImpulseExile{
				Player:                game.ControllerReference(),
				Amount:                game.Fixed(1),
				Duration:              game.DurationThisTurn,
				WithoutPayingManaCost: free,
			}, &TurnLog{})

			if !g.Players[game.Player1].Exile.Contains(cardID) {
				t.Fatal("top card was not exiled")
			}
			if got := castFromZoneWithoutPayingManaCost(g, game.Player1, cardID, zone.Exile, game.FaceFront); got != free {
				t.Fatalf("free-cast permission = %v, want %v", got, free)
			}
		})
	}
}
