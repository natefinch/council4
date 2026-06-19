package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// The static-source frame cache must be transparent: a value computed inside a
// frame must equal the value computed without one. These tests build a board
// that exercises the cached paths (static-ability sources, counters, control
// change, face-down) and assert framed and unframed results agree for every
// permanent, which catches any cache divergence.

func frameDiffCreature(name string, power, toughness int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: power}),
		Toughness: opt.Val(game.PT{Value: toughness}),
	}}
}

// frameDiffBoard builds a four-player board with several overlapping continuous
// effects, a static-ability source, counters, a control change, and a face-down
// permanent.
func frameDiffBoard(t *testing.T) *game.Game {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	reach := frameDiffCreature("Reachy", 1, 3)
	reach.StaticAbilities = append(reach.StaticAbilities, game.ReachStaticBody)
	addCombatPermanent(g, game.Player1, reach)

	bear := addCombatPermanent(g, game.Player1, frameDiffCreature("Bear", 2, 2))
	bear.Counters.Add(counter.PlusOnePlusOne, 2)

	stolen := addCombatPermanent(g, game.Player2, frameDiffCreature("Stolen", 4, 4))

	addCombatPermanent(g, game.Player3, frameDiffCreature("Plain", 3, 3))

	faceDown := addCombatPermanent(g, game.Player4, frameDiffCreature("Hidden", 5, 5))
	faceDown.FaceDown = true

	g.ContinuousEffects = append(g.ContinuousEffects,
		game.ContinuousEffect{
			ID:               1,
			AffectedObjectID: bear.ObjectID,
			Timestamp:        10,
			Layer:            game.LayerPowerToughnessModify,
			PowerDelta:       1,
			ToughnessDelta:   1,
		},
		game.ContinuousEffect{
			ID:               2,
			AffectedObjectID: stolen.ObjectID,
			Timestamp:        20,
			Layer:            game.LayerControl,
			NewController:    opt.Val(game.Player1),
		},
	)
	return g
}

type frameDiffValues struct {
	power       int
	toughness   int
	toughnessOK bool
	controller  game.PlayerID
	abilities   int
	reach       bool
}

func frameDiffSnapshot(g *game.Game, permanent *game.Permanent) frameDiffValues {
	toughness, toughnessOK := effectiveToughness(g, permanent)
	return frameDiffValues{
		power:       effectivePower(g, permanent),
		toughness:   toughness,
		toughnessOK: toughnessOK,
		controller:  effectiveController(g, permanent),
		abilities:   len(permanentEffectiveAbilities(g, permanent)),
		reach:       hasKeyword(g, permanent, game.Reach),
	}
}

func TestStaticSourceFrameIsTransparent(t *testing.T) {
	g := frameDiffBoard(t)

	for _, permanent := range g.Battlefield {
		unframed := frameDiffSnapshot(g, permanent)

		g.BeginStaticSourceFrame()
		framed := frameDiffSnapshot(g, permanent)
		// A second read in the same frame must also match (exercises the hit).
		framedAgain := frameDiffSnapshot(g, permanent)
		g.EndStaticSourceFrame()

		if unframed != framed || unframed != framedAgain {
			t.Errorf("permanent %d: unframed %+v, framed %+v, framed-again %+v",
				permanent.CardInstanceID, unframed, framed, framedAgain)
		}
	}
}

// TestStaticSourceFrameMatchesAfterReopen confirms a fresh frame reflects state
// changes made between frames (the cache does not persist across frames).
func TestStaticSourceFrameMatchesAfterReopen(t *testing.T) {
	g := frameDiffBoard(t)
	bear := g.Battlefield[1]

	g.BeginStaticSourceFrame()
	before := effectivePower(g, bear)
	g.EndStaticSourceFrame()

	bear.Counters.Add(counter.PlusOnePlusOne, 3)

	g.BeginStaticSourceFrame()
	after := effectivePower(g, bear)
	g.EndStaticSourceFrame()

	unframed := effectivePower(g, bear)
	if after != unframed {
		t.Errorf("after adding counters, framed power %d != unframed %d", after, unframed)
	}
	if after <= before {
		t.Errorf("adding +1/+1 counters should raise power: before %d, after %d", before, after)
	}
}
