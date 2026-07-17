package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// mycosynthBecomeCopyAbilityDef builds The Mycosynth Gardens' novel activated
// ability: "{X}, {T}: This land becomes a copy of target nontoken artifact you
// control with mana value X." The lone permanent target is a nontoken artifact
// the controller controls whose mana value is bound to the chosen X via
// ManaValueEqualsX (an exact bound, not at-most), and the effect is a permanent
// BecomeCopy over that target.
func mycosynthBecomeCopyAbilityDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "The Mycosynth Gardens",
		Types: []types.Card{types.Land},
		ActivatedAbilities: []game.ActivatedAbility{{
			Text:            "{X}, {T}: This land becomes a copy of target nontoken artifact you control with mana value X.",
			ManaCost:        opt.Val(cost.Mana{cost.X}),
			AdditionalCosts: cost.Tap,
			ZoneOfFunction:  zone.Battlefield,
			Content: game.Mode{
				Targets: []game.TargetSpec{{
					MinTargets:       1,
					MaxTargets:       1,
					Constraint:       "target nontoken artifact you control with mana value X",
					Allow:            game.TargetAllowPermanent,
					Selection:        opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact}, Controller: game.ControllerYou, NonToken: true}),
					ManaValueEqualsX: true,
				}},
				Sequence: []game.Instruction{{Primitive: game.BecomeCopy{Object: game.TargetPermanentReference(0)}}},
			}.Ability(),
		}},
	}}
}

// TestBodyTargetsSatisfyManaValueXBoundsByExactX verifies the activated-ability
// announcement enforcement for the exact "mana value X" bound: a nontoken
// artifact you control is a legal target only when its mana value equals the
// chosen X, so (unlike the at-most-X bound) neither a smaller nor a larger X
// admits it.
func TestBodyTargetsSatisfyManaValueXBoundsByExactX(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := mycosynthBecomeCopyAbilityDef()
	land := addCombatPermanent(g, game.Player1, def)
	artifact := addArtifactWithManaValue(g, game.Player1, 3)
	body := &def.ActivatedAbilities[0]
	targets := []game.Target{game.PermanentTarget(artifact.ObjectID)}

	cases := []struct {
		name   string
		xValue int
		want   bool
	}{
		{"exact", 3, true},
		{"under", 2, false},
		{"over", 4, false},
		{"zero", 0, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := bodyTargetsSatisfyManaValueX(g, game.Player1, def, land.ObjectID, body, nil, targets, tc.xValue)
			if got != tc.want {
				t.Fatalf("bodyTargetsSatisfyManaValueX(mv3 artifact, X=%d) = %v, want %v", tc.xValue, got, tc.want)
			}
		})
	}
}

// TestBecomeCopyManaValueXResolutionRechecksExactBound verifies CR 608.2b
// re-checking for the exact bound: at resolution the targeted artifact is legal
// only when its mana value still equals the ability's chosen X, and otherwise it
// is deferred so the become-a-copy does nothing (fizzles on the lone target).
func TestBecomeCopyManaValueXResolutionRechecksExactBound(t *testing.T) {
	def := mycosynthBecomeCopyAbilityDef()
	body := &def.ActivatedAbilities[0]
	specs := bodyTargetSpecs(body, nil)

	cases := []struct {
		name      string
		manaValue int
		xValue    int
		wantLegal bool
		wantDefer bool
	}{
		{"still exact", 3, 3, true, false},
		{"now under X", 2, 3, false, true},
		{"now over X", 4, 3, false, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			land := addCombatPermanent(g, game.Player1, def)
			artifact := addArtifactWithManaValue(g, game.Player1, tc.manaValue)
			obj := &game.StackObject{
				Kind:       game.StackActivatedAbility,
				Controller: game.Player1,
				SourceID:   land.ObjectID,
				XValue:     tc.xValue,
				Targets:    []game.Target{game.PermanentTarget(artifact.ObjectID)},
			}
			gotLegal := stackObjectHasAnyLegalTargetsForSpecs(g, def, land.ObjectID, specs, obj)
			if gotLegal != tc.wantLegal {
				t.Fatalf("stackObjectHasAnyLegalTargetsForSpecs(mv%d, X=%d) = %v, want %v", tc.manaValue, tc.xValue, gotLegal, tc.wantLegal)
			}
			deferred := obj.Targets[0].Kind == game.TargetDeferred
			if deferred != tc.wantDefer {
				t.Fatalf("target deferred = %v, want %v", deferred, tc.wantDefer)
			}
		})
	}
}

// TestBecomeCopyManaValueXResolvesToArtifactCopy verifies the composed end
// state: resolving the become-a-copy makes the source land a copy of the
// targeted nontoken artifact, taking its name and artifact type and losing its
// original land type (it copies no land abilities the artifact lacks).
func TestBecomeCopyManaValueXResolvesToArtifactCopy(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := mycosynthBecomeCopyAbilityDef()
	source := addCombatPermanent(g, game.Player1, def)
	target := addArtifactWithManaValue(g, game.Player1, 3)

	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackActivatedAbility,
		Controller: game.Player1,
		SourceID:   source.ObjectID,
		XValue:     3,
		Targets:    []game.Target{game.PermanentTarget(target.ObjectID)},
	}
	r := &effectResolver{engine: NewEngine(nil), game: g, obj: obj, log: &TurnLog{}}

	resolved := handleBecomeCopy(r, game.BecomeCopy{Object: game.TargetPermanentReference(0)})
	if !resolved.succeeded {
		t.Fatal("handleBecomeCopy did not succeed")
	}
	if got := permanentEffectiveName(g, source); got != "Test Artifact" {
		t.Fatalf("effective name = %q, want copied Test Artifact", got)
	}
	if !permanentHasType(g, source, types.Artifact) {
		t.Fatal("source should be an Artifact after copying the artifact")
	}
	if permanentHasType(g, source, types.Land) {
		t.Fatal("source should lose its Land type when copying a non-land artifact")
	}
}
