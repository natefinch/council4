package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// blueSunsTwilightSequence mirrors the lowered Blue Sun's Twilight spell
// ability: gain control of the target permanent unconditionally, then — only
// when the resolving spell's chosen {X} is 5 or more — create a token that's a
// copy of that same gained creature (target slot 0). The X threshold gates only
// the copy, never the gain-control.
func blueSunsTwilightSequence() []game.Instruction {
	return []game.Instruction{
		{
			Primitive: game.ApplyContinuous{
				Object: opt.Val(game.TargetPermanentReference(0)),
				ContinuousEffects: []game.ContinuousEffect{{
					Layer:         game.LayerControl,
					NewController: opt.Val(game.Player1),
				}},
				Duration: game.DurationPermanent,
			},
		},
		{
			Primitive: game.CreateToken{
				Amount: game.Fixed(1),
				Source: game.TokenCopyOf(game.TokenCopySpec{
					Source: game.TokenCopySourceObject,
					Object: game.TargetPermanentReference(0),
				}),
			},
			Condition: opt.Val(game.EffectCondition{
				Condition: opt.Val(game.Condition{Aggregates: []game.AggregateComparison{{
					Aggregate: game.AggregateSpellX,
					Op:        compare.GreaterOrEqual,
					Value:     5,
				}}}),
			}),
		},
	}
}

// resolveBlueSunsTwilight casts and resolves the Blue Sun's Twilight sequence
// with the given chosen X against an opponent's creature, returning the game and
// the targeted creature.
func resolveBlueSunsTwilight(t *testing.T, xValue int) (*game.Game, *game.Permanent) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:      "Grizzly Bears",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
	sourceID := g.IDGen.Next()
	g.CardInstances[sourceID] = &game.CardInstance{
		ID: sourceID,
		Def: &game.CardDef{CardFace: game.CardFace{
			Name:  "Blue Sun's Twilight",
			Types: []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: blueSunsTwilightSequence(),
			}.Ability()),
		}},
		Owner: game.Player1,
	}
	g.Stack.Push(&game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   sourceID,
		Controller: game.Player1,
		XValue:     xValue,
		Targets:    []game.Target{game.PermanentTarget(creature.ObjectID)},
	})
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	return g, creature
}

// countCopyTokens counts battlefield token permanents named name controlled by
// controller — i.e. copies of the gained creature under our control.
func countCopyTokens(g *game.Game, name string, controller game.PlayerID) int {
	copies := 0
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanent.Controller == controller &&
			permanent.TokenDef != nil && permanent.TokenDef.Name == name {
			copies++
		}
	}
	return copies
}

// TestBlueSunsTwilightBelowThresholdGainsControlOnly proves that when the chosen
// X is below 5, Blue Sun's Twilight gains control of the target creature but the
// gated copy-of-that-creature does not run, so no token is created.
func TestBlueSunsTwilightBelowThresholdGainsControlOnly(t *testing.T) {
	g, creature := resolveBlueSunsTwilight(t, 4)
	if got := effectiveController(g, creature); got != game.Player1 {
		t.Fatalf("effective controller = %v, want Player1 (gain control always applies)", got)
	}
	if got := countCopyTokens(g, "Grizzly Bears", game.Player1); got != 0 {
		t.Fatalf("copy tokens at X=4 = %d, want 0 (below the X>=5 gate)", got)
	}
}

// TestBlueSunsTwilightAtThresholdGainsControlAndCopies proves that when the
// chosen X is at least 5, Blue Sun's Twilight both gains control of the target
// creature and creates a token that's a copy of it under our control.
func TestBlueSunsTwilightAtThresholdGainsControlAndCopies(t *testing.T) {
	g, creature := resolveBlueSunsTwilight(t, 5)
	if got := effectiveController(g, creature); got != game.Player1 {
		t.Fatalf("effective controller = %v, want Player1", got)
	}
	if got := countCopyTokens(g, "Grizzly Bears", game.Player1); got != 1 {
		t.Fatalf("copy tokens at X=5 = %d, want 1 (X>=5 gate met)", got)
	}
}
