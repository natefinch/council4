package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// manaValueAtMostXSpellDef builds a Dominate-like spell whose lone creature
// target spec bounds the target's mana value by the chosen X via
// ManaValueAtMostX.
func manaValueAtMostXSpellDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Gain Control MV X",
		Types: []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets:       1,
				MaxTargets:       1,
				Constraint:       "target creature with mana value X or less",
				Allow:            game.TargetAllowPermanent,
				Selection:        opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				ManaValueAtMostX: true,
			}},
			Sequence: []game.Instruction{{Primitive: game.ApplyContinuous{
				Object: opt.Val(game.TargetPermanentReference(0)),
				ContinuousEffects: []game.ContinuousEffect{{
					Layer:         game.LayerControl,
					NewController: opt.Val(game.Player1),
				}},
				Duration: game.DurationPermanent,
			}}},
		}.Ability()),
	}}
}

func addCreatureWithManaValue(g *game.Game, controller game.PlayerID, manaValue int) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:     "MV Creature",
		Types:    []types.Card{types.Creature},
		ManaCost: opt.Val(cost.Mana{cost.O(manaValue)}),
	}})
}

// TestSpellTargetsSatisfyManaValueXBoundsTargetByX verifies the ManaValueAtMostX
// cast legality: a target is legal only when its mana value is at most the chosen
// X, so a low X cannot gain control of an expensive creature.
func TestSpellTargetsSatisfyManaValueXBoundsTargetByX(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	mv2 := addCreatureWithManaValue(g, game.Player2, 2)
	mv5 := addCreatureWithManaValue(g, game.Player2, 5)
	card := manaValueAtMostXSpellDef()

	cases := []struct {
		name   string
		target *game.Permanent
		xValue int
		want   bool
	}{
		{"mv2 X=2", mv2, 2, true},
		{"mv2 X=1", mv2, 1, false},
		{"mv2 X=5", mv2, 5, true},
		{"mv5 X=5", mv5, 5, true},
		{"mv5 X=4", mv5, 4, false},
		{"mv5 X=0", mv5, 0, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			targets := []game.Target{game.PermanentTarget(tc.target.ObjectID)}
			got := spellTargetsSatisfyManaValueX(g, game.Player1, card, nil, targets, tc.xValue, game.CastBranch{})
			if got != tc.want {
				t.Fatalf("spellTargetsSatisfyManaValueX(mv target, X=%d) = %v, want %v", tc.xValue, got, tc.want)
			}
		})
	}
}

// TestSpellTargetsSatisfyManaValueXIgnoresNonXSpecs verifies a spell without a
// ManaValueAtMostX spec is unaffected by the bound check.
func TestSpellTargetsSatisfyManaValueXIgnoresNonXSpecs(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	mv5 := addCreatureWithManaValue(g, game.Player2, 5)
	card := &game.CardDef{CardFace: game.CardFace{
		Name:  "Plain Control",
		Types: []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Constraint: "target creature",
				Allow:      game.TargetAllowPermanent,
				Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
			}},
			Sequence: []game.Instruction{{Primitive: game.Destroy{Object: game.TargetPermanentReference(0)}}},
		}.Ability()),
	}}
	if !spellTargetsSatisfyManaValueX(g, game.Player1, card, nil, []game.Target{game.PermanentTarget(mv5.ObjectID)}, 0, game.CastBranch{}) {
		t.Fatal("non-ManaValueAtMostX spell should pass the X-bound check regardless of X")
	}
}

// TestManaValueAtMostXResolutionRechecksBound verifies CR 608.2b re-checking: a
// target whose mana value exceeds the resolving spell's X is no longer legal at
// resolution and is deferred, while a within-bound target stays legal.
func TestManaValueAtMostXResolutionRechecksBound(t *testing.T) {
	specs := spellTargetSpecs(manaValueAtMostXSpellDef(), nil, game.CastBranch{})

	cases := []struct {
		name      string
		manaValue int
		xValue    int
		wantLegal bool
		wantDefer bool
	}{
		{"within bound", 2, 3, true, false},
		{"at bound", 3, 3, true, false},
		{"over bound", 5, 3, false, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			creature := addCreatureWithManaValue(g, game.Player2, tc.manaValue)
			obj := &game.StackObject{
				Kind:       game.StackSpell,
				Controller: game.Player1,
				XValue:     tc.xValue,
				Targets:    []game.Target{game.PermanentTarget(creature.ObjectID)},
			}
			gotLegal := stackObjectHasAnyLegalTargetsForSpecs(g, nil, 0, specs, obj)
			if gotLegal != tc.wantLegal {
				t.Fatalf("stackObjectHasAnyLegalTargetsForSpecs = %v, want %v", gotLegal, tc.wantLegal)
			}
			deferred := obj.Targets[0].Kind == game.TargetDeferred
			if deferred != tc.wantDefer {
				t.Fatalf("target deferred = %v, want %v", deferred, tc.wantDefer)
			}
		})
	}
}
