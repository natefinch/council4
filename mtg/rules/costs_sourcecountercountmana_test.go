package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// everflowingChalicePermanent adds a permanent whose tap mana ability adds one
// colorless mana for each charge counter on itself (Everflowing Chalice).
func everflowingChalicePermanent(g *game.Game, controller game.PlayerID, charges int) *game.Permanent {
	def := &game.CardDef{CardFace: game.CardFace{
		Name:  "Everflowing Chalice",
		Types: []types.Card{types.Artifact},
	}}
	def.ManaAbilities = append(def.ManaAbilities, game.ManaAbility{
		AdditionalCosts: cost.Tap,
		Content: game.Mode{
			Sequence: []game.Instruction{
				{Primitive: game.AddMana{
					ManaColor: mana.C,
					Amount: game.Dynamic(game.DynamicAmount{
						Kind:        game.DynamicAmountObjectCounters,
						Multiplier:  1,
						CounterKind: counter.Charge,
						Object:      game.SourcePermanentReference(),
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
		panic("everflowing chalice permanent was not created")
	}
	permanent.Counters.Add(counter.Charge, charges)
	return permanent
}

func TestSourceCounterCountManaAbilityScalesWithCounters(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	chalice := everflowingChalicePermanent(g, game.Player1, 3)

	want := action.ActivateAbility(chalice.ObjectID, 0, nil, 0)
	if !engine.applyAction(g, game.Player1, want) {
		t.Fatal("applyAction(everflowing chalice mana ability) = false, want true")
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.C); got != 3 {
		t.Fatalf("colorless mana = %d, want 3 (one per charge counter)", got)
	}
}
