package rules

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestObjectMatchesTargetCardAfterExile(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		cardType types.Card
		prior    bool
	}{
		{types.Creature, false}, {types.Land, false}, {types.Instant, false},
		{types.Creature, true}, {types.Land, true}, {types.Instant, true},
	} {
		t.Run(fmt.Sprintf("%s/prior=%t", test.cardType, test.prior), func(t *testing.T) {
			t.Parallel()
			cardType := test.cardType
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			source := addCombatCreaturePermanent(g, game.Player1, 2, 2)
			cardID := addCardToGraveyard(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
				Name: "Target Card", Types: []types.Card{cardType},
			}})
			var targets []game.Target
			if test.prior {
				priorCardID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
					Name: "Earlier Target", Types: []types.Card{types.Instant},
				}})
				targets = append(targets, currentCardTarget(t, g, priorCardID))
			}
			slot := len(targets)
			targets = append(targets, currentCardTarget(t, g, cardID))
			obj := &game.StackObject{
				ID: g.IDGen.Next(), Kind: game.StackTriggeredAbility, Controller: game.Player1,
				SourceID: source.ObjectID, HasTriggerEvent: true,
				TriggerEvent: game.Event{Kind: game.EventPermanentEnteredBattlefield, PermanentID: source.ObjectID},
				Targets:      targets,
			}
			resolver := &effectResolver{engine: NewEngine(nil), game: g, obj: obj, log: &TurnLog{}}
			if test.prior {
				resolver.resolveInstruction(&game.Instruction{Primitive: game.MoveCard{
					Card:     game.CardReference{Kind: game.CardReferenceTarget},
					FromZone: zone.Graveyard, Destination: zone.Hand,
				}})
			}
			resolver.resolveInstruction(&game.Instruction{Primitive: game.MoveCard{
				Card:     game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: slot},
				FromZone: zone.Graveyard, Destination: zone.Exile,
			}, PublishResult: "moved"})
			if !g.Players[game.Player2].Exile.Contains(cardID) {
				t.Fatal("target did not move to exile")
			}
			condition := opt.Val(game.Condition{
				Object:              opt.Val(game.TargetCardReference(slot)),
				ObjectMatches:       opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
				TargetCardResultKey: "moved",
			})
			ctx := conditionContext{controller: game.Player1, source: source, obj: obj, event: &obj.TriggerEvent}
			if got := conditionSatisfied(g, ctx, condition); got != (cardType == types.Creature) {
				t.Fatalf("target-card condition = %v for %s", got, cardType)
			}
			before := g.Players[game.Player2].Life
			resolver.resolveInstruction(&game.Instruction{
				Primitive: game.LoseLife{
					Amount: game.Fixed(1), Player: game.ObjectOwnerReference(game.TargetCardReference(slot)),
				},
				Condition: opt.Val(game.EffectCondition{Condition: condition}),
			})
			want := before
			if cardType == types.Creature {
				want--
			}
			if got := g.Players[game.Player2].Life; got != want {
				t.Fatalf("target owner life = %d, want %d", got, want)
			}
			if g.Players[game.Player1].Life != 40 {
				t.Fatal("payoff affected the preceding target's owner")
			}
			condition.Val.Object = opt.Val(game.EventPermanentReference())
			condition.Val.TargetCardResultKey = ""
			if !conditionSatisfied(g, ctx, condition) {
				t.Fatal("event-bound creature condition changed")
			}
			condition.Val.Object = opt.Val(game.TargetPermanentReference(slot))
			if conditionSatisfied(g, ctx, condition) {
				t.Fatal("permanent reference silently accepted a card target")
			}
		})
	}
}

func TestObjectMatchesPermanentCurrentAndLastKnownTypes(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Animated Land", Types: []types.Card{types.Land},
	}})
	obj := &game.StackObject{
		Controller: game.Player1,
		Targets:    []game.Target{game.PermanentTarget(target.ObjectID)},
	}

	condition := opt.Val(game.Condition{
		Object:        opt.Val(game.TargetPermanentReference(0)),
		ObjectMatches: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
	})
	ctx := conditionContext{controller: game.Player1, obj: obj}
	if conditionSatisfied(g, ctx, condition) {
		t.Fatal("printed land unexpectedly matched creature condition")
	}
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID: 1, AffectedObjectID: target.ObjectID, Layer: game.LayerType,
		AddTypes: []types.Card{types.Creature},
	})
	if !conditionSatisfied(g, ctx, condition) {
		t.Fatal("live target condition ignored effective creature type")
	}
	destroyPermanent(g, target.ObjectID)
	if !conditionSatisfied(g, ctx, condition) {
		t.Fatal("departed target condition ignored last-known creature type")
	}
}
