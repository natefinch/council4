package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// payRepeatedlyCounterContent mirrors the reusable Adversary-cycle shape: offer
// the controller {1}{G} any number of times during resolution, record the
// payment count, then put that many +1/+1 counters on the source. It exercises
// the PayRepeatedly publish path together with the existing
// DynamicAmountChosenNumber consumer.
func payRepeatedlyCounterContent() game.AbilityContent {
	const countKey = game.ResultKey("paid-count")
	return game.Mode{
		Sequence: []game.Instruction{
			{
				Primitive: game.PayRepeatedly{
					Payment: game.ResolutionPayment{
						ManaCost: opt.Val(cost.Mana{cost.O(1), cost.G}),
					},
					PublishCount: countKey,
					Prompt:       "Pay {1}{G}?",
				},
			},
			{
				Primitive: game.AddCounter{
					Amount: game.Dynamic(game.DynamicAmount{
						Kind:      game.DynamicAmountChosenNumber,
						ResultKey: countKey,
					}),
					Object:      game.SourcePermanentReference(),
					CounterKind: counter.PlusOnePlusOne,
				},
			},
		},
	}.Ability()
}

func pushPayRepeatedlyTrigger(g *game.Game, source *game.Permanent) {
	trigger := game.TriggeredAbility{Content: payRepeatedlyCounterContent()}
	g.Stack.Push(&game.StackObject{
		ID:            g.IDGen.Next(),
		Kind:          game.StackTriggeredAbility,
		SourceID:      source.ObjectID,
		SourceCardID:  source.CardInstanceID,
		Controller:    source.Controller,
		InlineTrigger: &trigger,
	})
}

func payRepeatedlySource(g *game.Game) *game.Permanent {
	return addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Primal Adversary",
		Types: []types.Card{types.Creature},
	}})
}

// TestPayRepeatedlyRecordsPaymentCount proves the controller may pay {1}{G} more
// than once during resolution and that the recorded count drives the
// downstream +1/+1 counter amount. Four Forests fund exactly two payments
// (each {1}{G} costs two mana), so the source gains two counters.
func TestPayRepeatedlyRecordsPaymentCount(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := payRepeatedlySource(g)
	for range 4 {
		addBasicLandPermanent(g, game.Player1, types.Forest)
	}
	pushPayRepeatedlyTrigger(g, source)

	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if got := source.Counters.Get(counter.PlusOnePlusOne); got != 2 {
		t.Fatalf("source has %d +1/+1 counters, want 2", got)
	}
}

// TestPayRepeatedlyRecordsZeroWhenControllerCannotPay proves the count is zero
// when the controller has no mana, leaving the gated payoff with nothing to do.
func TestPayRepeatedlyRecordsZeroWhenControllerCannotPay(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := payRepeatedlySource(g)
	pushPayRepeatedlyTrigger(g, source)

	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if got := source.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("source has %d +1/+1 counters, want 0", got)
	}
}

// TestPayRepeatedlyScalesWithAvailableMana proves the count grows with the mana
// the controller can spend: six Forests fund three {1}{G} payments.
func TestPayRepeatedlyScalesWithAvailableMana(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := payRepeatedlySource(g)
	for range 6 {
		addBasicLandPermanent(g, game.Player1, types.Forest)
	}
	pushPayRepeatedlyTrigger(g, source)

	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if got := source.Counters.Get(counter.PlusOnePlusOne); got != 3 {
		t.Fatalf("source has %d +1/+1 counters, want 3", got)
	}
}
