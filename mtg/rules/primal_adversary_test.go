package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// primalAdversaryContent mirrors the lowered Primal Adversary enters trigger:
// the controller may pay {1}{G} any number of times, then gains that many
// +1/+1 counters and animates up to that many lands they control into 3/3
// Wolves with haste that are still lands. It exercises the PayRepeatedly publish
// path, the DynamicAmountChosenNumber counter consumer, and the
// ApplyContinuous choose-up-to-from-group selection together.
func primalAdversaryContent(source *game.Permanent) game.AbilityContent {
	const countKey = game.ResultKey("primal-pay")
	dynamic := game.Dynamic(game.DynamicAmount{
		Kind:      game.DynamicAmountChosenNumber,
		ResultKey: countKey,
	})
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
					Amount:      dynamic,
					Object:      game.SourcePermanentReference(),
					CounterKind: counter.PlusOnePlusOne,
				},
			},
			{
				Primitive: game.ApplyContinuous{
					ChooseFrom: game.ObjectControlledGroup(
						game.SourcePermanentReference(),
						game.Selection{RequiredTypes: []types.Card{types.Land}},
					),
					ChooseUpTo: dynamic,
					Prompt:     "Choose lands to animate",
					ContinuousEffects: []game.ContinuousEffect{
						{
							Layer:       game.LayerType,
							AddTypes:    []types.Card{types.Creature},
							AddSubtypes: []types.Sub{types.Wolf},
						},
						{
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Haste},
						},
						{
							Layer:        game.LayerPowerToughnessSet,
							SetPower:     opt.Val(game.PT{Value: 3}),
							SetToughness: opt.Val(game.PT{Value: 3}),
						},
					},
					Duration: game.DurationPermanent,
				},
			},
		},
	}.Ability()
}

func pushPrimalAdversaryTrigger(g *game.Game, source *game.Permanent) {
	trigger := game.TriggeredAbility{Content: primalAdversaryContent(source)}
	g.Stack.Push(&game.StackObject{
		ID:            g.IDGen.Next(),
		Kind:          game.StackTriggeredAbility,
		SourceID:      source.ObjectID,
		SourceCardID:  source.CardInstanceID,
		Controller:    source.Controller,
		InlineTrigger: &trigger,
	})
}

// TestPrimalAdversaryAnimatesPaidNumberOfLands proves the full enters sequence:
// six Forests fund three {1}{G} payments, so the source gains three +1/+1
// counters and three lands become 3/3 Wolf creatures with haste while remaining
// lands. A nil agent greedily takes every offered payment and selection.
func TestPrimalAdversaryAnimatesPaidNumberOfLands(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := payRepeatedlySource(g)
	lands := make([]*game.Permanent, 0, 6)
	for range 6 {
		lands = append(lands, addBasicLandPermanent(g, game.Player1, types.Forest))
	}
	pushPrimalAdversaryTrigger(g, source)

	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if got := source.Counters.Get(counter.PlusOnePlusOne); got != 3 {
		t.Fatalf("source has %d +1/+1 counters, want 3", got)
	}
	animated := 0
	for _, land := range lands {
		if permanentHasType(g, land, types.Creature) {
			animated++
			if !permanentHasType(g, land, types.Land) {
				t.Error("animated land lost its land type, want it retained")
			}
			if got := effectivePower(g, land); got != 3 {
				t.Errorf("animated land power = %d, want 3", got)
			}
			if !hasKeyword(g, land, game.Haste) {
				t.Error("animated land lacks haste")
			}
		}
	}
	if animated != 3 {
		t.Fatalf("animated %d lands, want 3", animated)
	}
}

// TestPrimalAdversaryAnimatesNothingWithoutPayment proves the choose-up-to
// selection is bounded by the payment count: with no mana the controller pays
// zero times, gains no counters, and animates no lands.
func TestPrimalAdversaryAnimatesNothingWithoutPayment(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := payRepeatedlySource(g)
	land := addBasicLandPermanent(g, game.Player1, types.Forest)
	pushPrimalAdversaryTrigger(g, source)

	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if got := source.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("source has %d +1/+1 counters, want 0", got)
	}
	if permanentHasType(g, land, types.Creature) {
		t.Fatal("land became a creature despite no payment")
	}
}
