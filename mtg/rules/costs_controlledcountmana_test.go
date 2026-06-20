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

// cabalCoffersPermanent adds a permanent whose tap mana ability adds one black
// mana for each Swamp its controller has on the battlefield (Cabal Coffers).
func cabalCoffersPermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	def := &game.CardDef{CardFace: game.CardFace{
		Name:  "Cabal Coffers",
		Types: []types.Card{types.Land},
	}}
	def.ManaAbilities = append(def.ManaAbilities, game.ManaAbility{
		AdditionalCosts: cost.Tap,
		Content: game.Mode{
			Sequence: []game.Instruction{
				{Primitive: game.AddMana{
					ManaColor: mana.B,
					Amount: game.Dynamic(game.DynamicAmount{
						Kind:       game.DynamicAmountCountSelector,
						Multiplier: 1,
						Group: game.BattlefieldGroup(game.Selection{
							SubtypesAny: []types.Sub{types.Swamp},
							Controller:  game.ControllerYou,
						}),
					}),
				}},
			},
		}.Ability(),
	})
	cardID := g.IDGen.Next()
	card := &game.CardInstance{ID: cardID, Def: def, Owner: controller}
	g.CardInstances[cardID] = card
	permanent, ok := createCardPermanent(g, card, controller, zone.Stack)
	if !ok {
		panic("cabal coffers permanent was not created")
	}
	return permanent
}

func TestControlledCountManaAbilityScalesWithSwamps(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	coffers := cabalCoffersPermanent(g, game.Player1)
	for range 3 {
		addBasicLandPermanent(g, game.Player1, types.Swamp)
	}
	// An opponent's Swamp must not be counted by "you control".
	addBasicLandPermanent(g, game.Player2, types.Swamp)

	want := action.ActivateAbility(coffers.ObjectID, 0, nil, 0)
	if !engine.applyAction(g, game.Player1, want) {
		t.Fatal("applyAction(cabal coffers mana ability) = false, want true")
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.B); got != 3 {
		t.Fatalf("black mana = %d, want 3 (one per controlled Swamp)", got)
	}
}
