package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func spiritGuideHandCard(name string, m mana.Color) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  name,
		Types: []types.Card{types.Creature},
		ManaAbilities: []game.ManaAbility{{
			ZoneOfFunction: zone.Hand,
			AdditionalCosts: []cost.Additional{{
				Kind:   cost.AdditionalExileSource,
				Source: zone.Hand,
				Amount: 1,
			}},
			Content: game.Mode{
				Sequence: []game.Instruction{
					{Primitive: game.AddMana{ManaColor: m, Amount: game.Fixed(1)}},
				},
			}.Ability(),
		}},
	}}
}

func TestHandManaAbilityExilesCardAndFloatsMana(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, spiritGuideHandCard("Simian Spirit Guide", mana.R))

	want := action.ActivateAbility(cardID, 0, nil, 0)
	if !containsAction(engine.legalActions(g, game.Player1), want) {
		t.Fatalf("hand mana ability was not exposed as a legal action: %+v", engine.legalActions(g, game.Player1))
	}
	if !engine.applyAction(g, game.Player1, want) {
		t.Fatal("applyAction(hand mana ability) = false, want true")
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.R); got != 1 {
		t.Fatalf("red mana = %d, want 1", got)
	}
	if g.Players[game.Player1].Hand.Contains(cardID) {
		t.Fatal("Spirit Guide remained in hand after activation")
	}
	if !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("Spirit Guide was not exiled by its activation cost")
	}
	if got := g.Stack.Size(); got != 0 {
		t.Fatalf("stack size = %d, want 0 for mana ability", got)
	}
}
