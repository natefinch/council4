package rules

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// regContinuousCreature builds a vanilla creature card definition with the
// given base power and toughness.
func regContinuousCreature(name string, power, toughness int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: power}),
		Toughness: opt.Val(game.PT{Value: toughness}),
	}}
}

// TestRegLayeredEffectiveValuesCombineSetModifyAndCounters asserts the canonical
// layer pipeline: a set-P/T effect (layer 7b) is overridden by base, then a
// modify effect (layer 7c) adds, then +1/+1 counters add last.
func TestRegLayeredEffectiveValuesCombineSetModifyAndCounters(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatPermanent(g, game.Player1, regContinuousCreature("Layered Beast", 4, 4))
	creature.Counters.Add(counter.PlusOnePlusOne, 3)
	g.ContinuousEffects = append(g.ContinuousEffects,
		game.ContinuousEffect{
			ID:               1,
			AffectedObjectID: creature.ObjectID,
			Timestamp:        10,
			Layer:            game.LayerPowerToughnessSet,
			SetPower:         opt.Val(game.PT{Value: 1}),
			SetToughness:     opt.Val(game.PT{Value: 1}),
		},
		game.ContinuousEffect{
			ID:               2,
			AffectedObjectID: creature.ObjectID,
			Timestamp:        20,
			Layer:            game.LayerPowerToughnessModify,
			PowerDelta:       2,
			ToughnessDelta:   2,
		},
	)

	// set 1/1 -> modify +2/+2 -> 3/3 -> three +1/+1 counters -> 6/6.
	if got := effectivePower(g, creature); got != 6 {
		t.Fatalf("effective power = %d, want 6", got)
	}
	if got, ok := effectiveToughness(g, creature); !ok || got != 6 {
		t.Fatalf("effective toughness = %d ok=%v, want 6 true", got, ok)
	}
}

// TestRegSetPowerToughnessLaterTimestampWins asserts that two set-P/T effects in
// the same layer are ordered by timestamp, so the later one determines the
// value.
func TestRegSetPowerToughnessLaterTimestampWins(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatPermanent(g, game.Player1, regContinuousCreature("Reset Beast", 5, 5))
	g.ContinuousEffects = append(g.ContinuousEffects,
		game.ContinuousEffect{
			ID:               1,
			AffectedObjectID: creature.ObjectID,
			Timestamp:        30,
			Layer:            game.LayerPowerToughnessSet,
			SetPower:         opt.Val(game.PT{Value: 7}),
			SetToughness:     opt.Val(game.PT{Value: 7}),
		},
		game.ContinuousEffect{
			ID:               2,
			AffectedObjectID: creature.ObjectID,
			Timestamp:        10,
			Layer:            game.LayerPowerToughnessSet,
			SetPower:         opt.Val(game.PT{Value: 2}),
			SetToughness:     opt.Val(game.PT{Value: 2}),
		},
	)

	// The later-timestamp (30) set effect wins regardless of slice order.
	if got := effectivePower(g, creature); got != 7 {
		t.Fatalf("effective power = %d, want later-timestamp 7", got)
	}
	if got, ok := effectiveToughness(g, creature); !ok || got != 7 {
		t.Fatalf("effective toughness = %d ok=%v, want 7 true", got, ok)
	}
}

// TestRegTypeColorControlLayersProduceEffectiveValues asserts that control,
// color, and type layers each take effect and compose on one permanent.
func TestRegTypeColorControlLayersProduceEffectiveValues(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:   "Chameleon",
		Colors: []color.Color{color.Green},
		Types:  []types.Card{types.Creature},
	}})
	g.ContinuousEffects = append(g.ContinuousEffects,
		game.ContinuousEffect{
			ID:               1,
			AffectedObjectID: creature.ObjectID,
			Layer:            game.LayerControl,
			NewController:    opt.Val(game.Player2),
		},
		game.ContinuousEffect{
			ID:               2,
			AffectedObjectID: creature.ObjectID,
			Layer:            game.LayerColor,
			AddColors:        []color.Color{color.Red},
		},
		game.ContinuousEffect{
			ID:               3,
			AffectedObjectID: creature.ObjectID,
			Layer:            game.LayerType,
			AddTypes:         []types.Card{types.Artifact},
		},
	)

	if got := effectiveController(g, creature); got != game.Player2 {
		t.Fatalf("effective controller = %v, want Player2", got)
	}
	colors := permanentEffectiveColors(g, creature)
	if !slices.Contains(colors, color.Red) || !slices.Contains(colors, color.Green) {
		t.Fatalf("effective colors = %v, want both green and red", colors)
	}
	if !permanentHasType(g, creature, types.Artifact) || !permanentHasType(g, creature, types.Creature) {
		t.Fatal("effective types should include both Creature and Artifact")
	}
}

// TestRegAbilityLayerKeywordOrderedByTimestamp asserts that within the ability
// layer an add and a remove of the same keyword resolve in timestamp order.
func TestRegAbilityLayerKeywordOrderedByTimestamp(t *testing.T) {
	t.Parallel()
	t.Run("remove after add clears keyword", func(t *testing.T) {
		t.Parallel()
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		creature := addCombatPermanent(g, game.Player1, regContinuousCreature("Grounded", 2, 2))
		g.ContinuousEffects = append(g.ContinuousEffects,
			game.ContinuousEffect{
				ID:               1,
				AffectedObjectID: creature.ObjectID,
				Timestamp:        10,
				Layer:            game.LayerAbility,
				AddKeywords:      []game.Keyword{game.Flying},
			},
			game.ContinuousEffect{
				ID:               2,
				AffectedObjectID: creature.ObjectID,
				Timestamp:        20,
				Layer:            game.LayerAbility,
				RemoveKeywords:   []game.Keyword{game.Flying},
			},
		)
		if hasKeyword(g, creature, game.Flying) {
			t.Fatal("flying should be removed by the later-timestamp effect")
		}
	})
	t.Run("add after remove keeps keyword", func(t *testing.T) {
		t.Parallel()
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		creature := addCombatPermanent(g, game.Player1, regContinuousCreature("Airborne", 2, 2))
		g.ContinuousEffects = append(g.ContinuousEffects,
			game.ContinuousEffect{
				ID:               1,
				AffectedObjectID: creature.ObjectID,
				Timestamp:        10,
				Layer:            game.LayerAbility,
				RemoveKeywords:   []game.Keyword{game.Flying},
			},
			game.ContinuousEffect{
				ID:               2,
				AffectedObjectID: creature.ObjectID,
				Timestamp:        20,
				Layer:            game.LayerAbility,
				AddKeywords:      []game.Keyword{game.Flying},
			},
		)
		if !hasKeyword(g, creature, game.Flying) {
			t.Fatal("flying should be granted by the later-timestamp effect")
		}
	})
}

// TestRegPowerToughnessSwitchLayerAfterModify asserts the P/T-switch layer (7e)
// swaps the power and toughness produced by the modify layer. (The +1/+1
// counter here is symmetric, so it does not by itself pin counter-vs-switch
// ordering; it is included to confirm the switch composes with counters.)
func TestRegPowerToughnessSwitchLayerAfterModify(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatPermanent(g, game.Player1, regContinuousCreature("Topsy Turvy", 2, 4))
	creature.Counters.Add(counter.PlusOnePlusOne, 1)
	g.ContinuousEffects = append(g.ContinuousEffects,
		game.ContinuousEffect{
			ID:               1,
			AffectedObjectID: creature.ObjectID,
			Timestamp:        10,
			Layer:            game.LayerPowerToughnessModify,
			PowerDelta:       1,
		},
		game.ContinuousEffect{
			ID:               2,
			AffectedObjectID: creature.ObjectID,
			Timestamp:        20,
			Layer:            game.LayerPowerToughnessSwitch,
		},
	)

	// base 2/4 -> modify +1/+0 -> 3/4 -> switch -> 4/3 -> +1/+1 counter -> 5/4.
	if got := effectivePower(g, creature); got != 5 {
		t.Fatalf("effective power = %d, want 5", got)
	}
	if got, ok := effectiveToughness(g, creature); !ok || got != 4 {
		t.Fatalf("effective toughness = %d ok=%v, want 4 true", got, ok)
	}
}

// TestRegLayeredEffectiveValuesProperty fuzzes random set/modify/counter
// combinations against an independent reference implementation of the layer
// pipeline, and asserts the floor-at-zero rule for effective power.
func TestRegLayeredEffectiveValuesProperty(t *testing.T) {
	t.Parallel()
	rng := rand.New(rand.NewPCG(610, 1))
	for iteration := range 500 {
		basePower := rng.IntN(6)
		baseToughness := 1 + rng.IntN(6)
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		creature := addCombatPermanent(g, game.Player1, regContinuousCreature("Fuzz Beast", basePower, baseToughness))

		power, toughness := basePower, baseToughness

		var effects []game.ContinuousEffect
		var nextID id.ID = 1
		if rng.IntN(2) == 0 {
			setPower := rng.IntN(6)
			setToughness := rng.IntN(6)
			power, toughness = setPower, setToughness
			effects = append(effects, game.ContinuousEffect{
				ID:               nextID,
				AffectedObjectID: creature.ObjectID,
				Timestamp:        game.Timestamp(nextID),
				Layer:            game.LayerPowerToughnessSet,
				SetPower:         opt.Val(game.PT{Value: setPower}),
				SetToughness:     opt.Val(game.PT{Value: setToughness}),
			})
			nextID++
		}

		modifyCount := rng.IntN(3)
		for range modifyCount {
			powerDelta := rng.IntN(5) - 2
			toughnessDelta := rng.IntN(5) - 2
			power += powerDelta
			toughness += toughnessDelta
			effects = append(effects, game.ContinuousEffect{
				ID:               nextID,
				AffectedObjectID: creature.ObjectID,
				Timestamp:        game.Timestamp(nextID),
				Layer:            game.LayerPowerToughnessModify,
				PowerDelta:       powerDelta,
				ToughnessDelta:   toughnessDelta,
			})
			nextID++
		}

		plusCounters := rng.IntN(4)
		minusCounters := rng.IntN(4)
		if plusCounters > 0 {
			creature.Counters.Add(counter.PlusOnePlusOne, plusCounters)
		}
		if minusCounters > 0 {
			creature.Counters.Add(counter.MinusOneMinusOne, minusCounters)
		}
		power += plusCounters - minusCounters
		toughness += plusCounters - minusCounters

		g.ContinuousEffects = effects

		wantPower := max(0, power)
		if got := effectivePower(g, creature); got != wantPower {
			t.Fatalf("iteration %d: effective power = %d, want %d", iteration, got, wantPower)
		}
		got, ok := effectiveToughness(g, creature)
		if !ok || got != toughness {
			t.Fatalf("iteration %d: effective toughness = %d ok=%v, want %d true", iteration, got, ok, toughness)
		}
	}
}
