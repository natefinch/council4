package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: Heaven's Gate oracle fafdba94-51cd-413c-82f3-eb294825c6ca,
// {W} Instant, "One or more target creatures become white until end of turn."
// https://api.scryfall.com/cards/461d7c11-3a7d-42c2-bb6b-0a43779e6842/rulings
// returned no rulings. CR 2026-11-06 from the official copy: "one or more"
// needs at least one target and the count is announced on casting (601.2c);
// targets still legal at resolution are affected and illegal ones are not
// (608.2b); a spell's characteristic-changing effect fixes its affected set
// when it begins (611.2c); a new color replaces previous colors (105.3); and
// "until end of turn" effects end in cleanup (514.2). The card definition is
// test-local; registering the card is follow-up card work.
func heavensGateTestDef() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:     "Heaven's Gate",
			ManaCost: opt.Val(cost.Mana{cost.W}),
			Colors:   []color.Color{color.White},
			Types:    []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 99,
					Constraint: "one or more target creatures",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}},
				Sequence: []game.Instruction{{Primitive: game.ApplyContinuous{
					Object: opt.Val(game.AllTargetPermanentsReference(0)),
					ContinuousEffects: []game.ContinuousEffect{{
						Layer:     game.LayerColor,
						SetColors: []color.Color{color.White},
					}},
					Duration: game.DurationUntilEndOfTurn,
				}}},
			}.Ability()),
			OracleText: "One or more target creatures become white until end of turn.",
		},
	}
}

func coloredTestCreature(g *game.Game, controller game.PlayerID, name string, c color.Color) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Colors:    []color.Color{c},
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
}

// newHeavensGateTable puts a red and a green creature under two opponents and
// a blue creature under the caster, with Heaven's Gate in the caster's hand.
func newHeavensGateTable() (*Engine, *game.Game, []*game.Permanent) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.ActivePlayer = game.Player1
	creatures := []*game.Permanent{
		coloredTestCreature(g, game.Player2, "Red Opponent Creature", color.Red),
		coloredTestCreature(g, game.Player3, "Green Opponent Creature", color.Green),
		coloredTestCreature(g, game.Player1, "Blue Own Creature", color.Blue),
	}
	return engine, g, creatures
}

func castHeavensGate(t *testing.T, engine *Engine, g *game.Game, targets []*game.Permanent) bool {
	t.Helper()
	g.Players[game.Player1].ManaPool.Add(mana.W, 1)
	spellID := addCardToHand(g, game.Player1, heavensGateTestDef())
	chosen := make([]game.Target, len(targets))
	for i, permanent := range targets {
		chosen[i] = game.PermanentTarget(permanent.ObjectID)
	}
	return engine.applyAction(g, game.Player1, action.CastSpellFaceFromZone(spellID, zone.Hand, game.FaceFront, chosen, 0, nil))
}

func assertPermanentColors(t *testing.T, g *game.Game, permanent *game.Permanent, want ...color.Color) {
	t.Helper()
	if got := permanentEffectiveColors(g, permanent); !slices.Equal(got, want) {
		t.Fatalf("permanent %d colors = %v, want %v", permanent.ObjectID, got, want)
	}
}

// TestHeavensGateWhitensEveryChosenTargetUntilEndOfTurn casts Heaven's Gate
// through the stack at two opponents' creatures and one of the caster's own.
// Every target becomes white and regains its printed color after cleanup.
func TestHeavensGateWhitensEveryChosenTargetUntilEndOfTurn(t *testing.T) {
	engine, g, creatures := newHeavensGateTable()
	bystander := coloredTestCreature(g, game.Player2, "Untargeted Creature", color.Black)
	if !castHeavensGate(t, engine, g, creatures) {
		t.Fatal("Heaven's Gate cast with three targets failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	for _, creature := range creatures {
		assertPermanentColors(t, g, creature, color.White)
	}
	assertPermanentColors(t, g, bystander, color.Black)

	engine.runCleanupStep(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	assertPermanentColors(t, g, creatures[0], color.Red)
	assertPermanentColors(t, g, creatures[1], color.Green)
	assertPermanentColors(t, g, creatures[2], color.Blue)
}

// TestHeavensGateSkipsTargetsIllegalAtResolution makes one of three targets
// illegal in response; the spell still resolves and whitens the other two.
func TestHeavensGateSkipsTargetsIllegalAtResolution(t *testing.T) {
	for _, tc := range []struct {
		name    string
		respond func(*game.Game, *game.Permanent)
	}{
		{"target leaves the battlefield", func(g *game.Game, target *game.Permanent) {
			if _, ok := destroyPermanent(g, target.ObjectID); !ok {
				t.Fatal("destroying the responding target failed")
			}
		}},
		{"target gains hexproof", func(g *game.Game, target *game.Permanent) {
			g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
				ID: g.IDGen.Next(), AffectedObjectID: target.ObjectID, Layer: game.LayerAbility,
				AddKeywords: []game.Keyword{game.Hexproof}, Duration: game.DurationUntilEndOfTurn,
			})
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			engine, g, creatures := newHeavensGateTable()
			if !castHeavensGate(t, engine, g, creatures) {
				t.Fatal("Heaven's Gate cast with three targets failed")
			}
			illegal := creatures[0]
			tc.respond(g, illegal)
			engine.resolveTopOfStack(g, &TurnLog{})

			assertPermanentColors(t, g, creatures[1], color.White)
			assertPermanentColors(t, g, creatures[2], color.White)
			if permanent, ok := permanentByObjectID(g, illegal.ObjectID); ok {
				assertPermanentColors(t, g, permanent, color.Red)
			}
			for _, effect := range g.ContinuousEffects {
				if effect.AffectedObjectID == illegal.ObjectID && effect.Layer == game.LayerColor {
					t.Fatal("illegal target received Heaven's Gate's color effect")
				}
			}
		})
	}
}

// TestHeavensGateTargetCount covers "one or more": zero targets is not a legal
// cast, and exactly one target resolves normally.
func TestHeavensGateTargetCount(t *testing.T) {
	engine, g, creatures := newHeavensGateTable()
	if castHeavensGate(t, engine, g, nil) {
		t.Fatal("Heaven's Gate was cast with zero targets")
	}

	engine, g, creatures = newHeavensGateTable()
	if !castHeavensGate(t, engine, g, creatures[1:2]) {
		t.Fatal("Heaven's Gate cast with one target failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	assertPermanentColors(t, g, creatures[0], color.Red)
	assertPermanentColors(t, g, creatures[1], color.White)
	assertPermanentColors(t, g, creatures[2], color.Blue)
}

// TestHeavensGateCloneOnStackResolvesIdentically clones the game while Heaven's
// Gate is on the stack and resolves both games.
func TestHeavensGateCloneOnStackResolvesIdentically(t *testing.T) {
	engine, g, creatures := newHeavensGateTable()
	if !castHeavensGate(t, engine, g, creatures[:2]) {
		t.Fatal("Heaven's Gate cast failed")
	}
	clone := g.Clone()
	engine.resolveTopOfStack(g, &TurnLog{})
	engine.resolveTopOfStack(clone, &TurnLog{})

	for _, creature := range creatures {
		cloned, ok := permanentByObjectID(clone, creature.ObjectID)
		if !ok {
			t.Fatalf("clone lost permanent %d", creature.ObjectID)
		}
		original := permanentEffectiveColors(g, creature)
		if got := permanentEffectiveColors(clone, cloned); !slices.Equal(got, original) {
			t.Fatalf("clone permanent %d colors = %v, original = %v", creature.ObjectID, got, original)
		}
	}
	assertPermanentColors(t, clone, mustPermanent(t, clone, creatures[0]), color.White)
	assertPermanentColors(t, clone, mustPermanent(t, clone, creatures[2]), color.Blue)
}
