package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestPaidCostLegacyAmountAndPredicateUseIndependentFrozenFacts(t *testing.T) {
	for _, delta := range []int{1, 3} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		source := addCombatPermanent(g, game.Player1, compiledPaidSubjectCard(t,
			"Sacrifice a creature: You gain life equal to the sacrificed creature's power. If the sacrificed creature had toughness 4 or greater, draw a card.", "Artifact"))
		paid := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name: "Original", Types: []types.Card{types.Creature},
			Power: opt.Val(game.PT{Value: 1}), Toughness: opt.Val(game.PT{Value: 2}),
		}})
		g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
			Layer: game.LayerPowerToughnessModify, AffectedObjectID: paid.ObjectID,
			PowerDelta: delta, ToughnessDelta: delta,
		})
		reward := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Reward", Types: []types.Card{types.Land}}})
		setSorcerySpeedTurn(g, game.Player1)
		life := g.Players[game.Player1].Life
		if !engine.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0, nil, 0)) {
			t.Fatal("actual mixed-consumer activation could not pay its cost")
		}
		object, ok := g.Stack.Peek()
		if !ok || len(object.SacrificedAsCostIDs) != 1 || len(object.PaidCostSubjects) != 1 {
			t.Fatal("new predicate metadata replaced the legacy amount interface")
		}
		g.ContinuousEffects = nil
		card, _ := g.GetCardInstance(paid.CardInstanceID)
		card.Def = &game.CardDef{CardFace: game.CardFace{
			Name: "Later", Types: []types.Card{types.Creature},
			Power: opt.Val(game.PT{Value: 9}), Toughness: opt.Val(game.PT{Value: 9}),
		}}
		engine.resolveTopOfStack(g, &TurnLog{})
		if g.Players[game.Player1].Life != life+1+delta ||
			g.Players[game.Player1].Hand.Contains(reward) != (delta == 3) {
			t.Fatal("legacy amount or exact paid predicate reread later characteristics")
		}
	}
}
