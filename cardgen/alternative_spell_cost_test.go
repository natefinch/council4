package cardgen

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const commanderAlternativeCostText = "If you control a commander, you may cast this spell without paying its mana cost."
const vandalblastText = `Destroy target artifact you don't control.
Overload {4}{R} (You may cast this spell for its overload cost. If you do, change "target" in its text to "each.")`

func TestLowerFierceGuardianship(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Fierce Guardianship",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{2}{U}",
		OracleText: commanderAlternativeCostText + "\nCounter target noncreature spell.",
	})
	if len(face.AlternativeCosts) != 1 ||
		face.AlternativeCosts[0].Condition != cost.AlternativeConditionControlsCommander ||
		face.AlternativeCosts[0].ManaCost.Exists {
		t.Fatalf("alternative costs = %#v", face.AlternativeCosts)
	}

	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 ||
		!slices.Equal(mode.Targets[0].Predicate.ExcludedSpellCardTypes, []types.Card{types.Creature}) {
		t.Fatalf("targets = %#v, want target noncreature spell", mode.Targets)
	}

	counter, ok := mode.Sequence[0].Primitive.(game.CounterObject)
	if !ok || counter.Object != game.TargetStackObjectReference(0) {
		t.Fatalf("primitive = %#v, want counter target stack object", mode.Sequence[0].Primitive)
	}
}

func TestLowerVandalblastOverload(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Vandalblast",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{R}",
		OracleText: vandalblastText,
	})
	if !face.Overload.Exists ||
		!slices.Equal(face.Overload.Val.Cost, cost.Mana{cost.O(4), cost.R}) {
		t.Fatalf("overload = %#v", face.Overload)
	}
	normal := face.SpellAbility.Val.Modes[0]
	if len(normal.Targets) != 1 {
		t.Fatalf("normal targets = %#v", normal.Targets)
	}
	normalDestroy, ok := normal.Sequence[0].Primitive.(game.Destroy)
	if !ok || normalDestroy.Object != game.TargetPermanentReference(0) {
		t.Fatalf("normal primitive = %#v", normal.Sequence[0].Primitive)
	}
	overloaded := face.Overload.Val.SpellAbility.Modes[0]
	if len(overloaded.Targets) != 0 {
		t.Fatalf("overload targets = %#v, want none", overloaded.Targets)
	}
	overloadDestroy, ok := overloaded.Sequence[0].Primitive.(game.Destroy)
	selection := overloadDestroy.Group.Selection()
	if !ok || selection.Controller != game.ControllerNotYou ||
		!slices.Equal(selection.RequiredTypes, []types.Card{types.Artifact}) {
		t.Fatalf("overload primitive = %#v", overloaded.Sequence[0].Primitive)
	}
}

func TestLowerOverloadIsTextBlindAndFailsClosed(t *testing.T) {
	t.Parallel()
	normal := opt.Val(game.Mode{
		Targets: []game.TargetSpec{{
			MinTargets: 1,
			MaxTargets: 1,
			Allow:      game.TargetAllowPermanent,
			Predicate: game.TargetPredicate{
				PermanentTypes: []types.Card{types.Artifact},
				Controller:     game.ControllerNotYou,
			},
		}},
		Sequence: []game.Instruction{{Primitive: game.Destroy{Object: game.TargetPermanentReference(0)}}},
	}.Ability())
	spell := &compiler.CompiledAbility{Content: compiler.AbilityContent{
		Targets: []compiler.CompiledTarget{{
			Exact:       true,
			Cardinality: compiler.TargetCardinality{Min: 1, Max: 1},
			Selector: compiler.CompiledSelector{
				Kind:       compiler.SelectorArtifact,
				Controller: compiler.ControllerNotYou,
			},
		}},
	}}
	if _, ok := lowerOverloadSpell(normal, spell); !ok {
		t.Fatal("typed overload semantics unexpectedly depended on Oracle text")
	}
	spell.Content.Targets[0].Exact = false
	if _, ok := lowerOverloadSpell(normal, spell); ok {
		t.Fatal("inexact target unexpectedly lowered")
	}
}

func TestLowerQualifiedOverloadPreservesOrRejectsSelectors(t *testing.T) {
	t.Parallel()
	normal := opt.Val(game.Mode{
		Targets: []game.TargetSpec{{
			MinTargets: 1,
			MaxTargets: 1,
			Allow:      game.TargetAllowPermanent,
		}},
		Sequence: []game.Instruction{{Primitive: game.Destroy{Object: game.TargetPermanentReference(0)}}},
	}.Ability())
	lower := func(selector compiler.CompiledSelector) (game.Selection, bool) {
		t.Helper()
		overloaded, ok := lowerOverloadSpell(normal, &compiler.CompiledAbility{
			Content: compiler.AbilityContent{Targets: []compiler.CompiledTarget{{
				Exact:       true,
				Cardinality: compiler.TargetCardinality{Min: 1, Max: 1},
				Selector:    selector,
			}}},
		})
		if !ok {
			return game.Selection{}, false
		}
		destroy, ok := overloaded.Modes[0].Sequence[0].Primitive.(game.Destroy)
		if !ok {
			return game.Selection{}, false
		}
		return destroy.Group.Selection(), true
	}

	selection, ok := lower(compiler.CompiledSelector{
		Kind:            compiler.SelectorCreature,
		Controller:      compiler.ControllerNotYou,
		Another:         true,
		Attacking:       true,
		Blocking:        true,
		ExcludedKeyword: parser.KeywordFlying,
	})
	if !ok ||
		selection.Controller != game.ControllerNotYou ||
		!selection.ExcludeSource ||
		selection.CombatState != game.CombatStateAttackingOrBlocking ||
		selection.ExcludedKeyword != game.Flying {
		t.Fatalf("qualified overload selection = %#v, %v", selection, ok)
	}
	battleSelection, ok := lower(compiler.CompiledSelector{Kind: compiler.SelectorBattle})
	if !ok || !slices.Equal(battleSelection.RequiredTypes, []types.Card{types.Battle}) {
		t.Fatalf("battle overload selection = %#v, %v", battleSelection, ok)
	}

	for name, selector := range map[string]compiler.CompiledSelector{
		"zone": {
			Kind: compiler.SelectorCreature,
			Zone: zone.Graveyard,
		},
		"basic land type": {
			Kind:          compiler.SelectorLand,
			BasicLandType: true,
		},
		"player or planeswalker": {
			Kind:                 compiler.SelectorPlaneswalker,
			PlayerOrPlaneswalker: true,
		},
		"contradictory tapped state": {
			Kind:     compiler.SelectorCreature,
			Tapped:   true,
			Untapped: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if selection, ok := lower(selector); ok {
				t.Fatalf("unsupported selector broadened to %#v", selection)
			}
		})
	}
}

func TestGenerateVandalblastSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Vandalblast",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{R}",
		OracleText: vandalblastText,
	}, "v")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"Overload: opt.Val(game.OverloadAbility",
		"Cost: cost.Mana{cost.O(4), cost.R}",
		"Controller: game.ControllerNotYou",
		"Primitive: game.Destroy",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}

func TestLowerSafeOverloadSiblings(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name      string
		manaCost  string
		text      string
		cost      cost.Mana
		primitive func(game.Primitive) bool
	}{
		{
			name:     "Blustersquall",
			manaCost: "{U}",
			text:     `Tap target creature you don't control.` + "\n" + `Overload {3}{U} (You may cast this spell for its overload cost. If you do, change "target" in its text to "each.")`,
			cost:     cost.Mana{cost.O(3), cost.U},
			primitive: func(primitive game.Primitive) bool {
				tap, ok := primitive.(game.Tap)
				return ok && tap.Group.Valid()
			},
		},
		{
			name:     "Cyclonic Rift",
			manaCost: "{1}{U}",
			text:     `Return target nonland permanent you don't control to its owner's hand.` + "\n" + `Overload {6}{U} (You may cast this spell for its overload cost. If you do, change "target" in its text to "each.")`,
			cost:     cost.Mana{cost.O(6), cost.U},
			primitive: func(primitive game.Primitive) bool {
				bounce, ok := primitive.(game.Bounce)
				return ok && bounce.Group.Valid()
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       test.name,
				Layout:     "normal",
				TypeLine:   "Instant",
				ManaCost:   test.manaCost,
				OracleText: test.text,
			})
			if !face.Overload.Exists || !slices.Equal(face.Overload.Val.Cost, test.cost) {
				t.Fatalf("overload = %#v", face.Overload)
			}
			mode := face.Overload.Val.SpellAbility.Modes[0]
			if len(mode.Targets) != 0 || !test.primitive(mode.Sequence[0].Primitive) {
				t.Fatalf("overloaded mode = %#v", mode)
			}
		})
	}
}

func TestLowerCommanderAlternativeCostIsTextBlind(t *testing.T) {
	t.Parallel()
	lowered, diagnostic := lowerSpellAlternativeCost(compiler.CompiledAbility{
		Kind: compiler.AbilitySpellAlternativeCost,
		Text: "not Oracle wording",
		AlternativeCost: &compiler.CompiledAlternativeCost{
			Condition:             compiler.AlternativeCostConditionControlsCommander,
			WithoutPayingManaCost: true,
		},
	})
	if diagnostic != nil {
		t.Fatalf("diagnostic = %#v", diagnostic)
	}
	if len(lowered.alternativeCosts) != 1 ||
		lowered.alternativeCosts[0].Condition != cost.AlternativeConditionControlsCommander {
		t.Fatalf("alternative costs = %#v", lowered.alternativeCosts)
	}
}

func TestGenerateFierceGuardianshipSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Fierce Guardianship",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{2}{U}",
		OracleText: commanderAlternativeCostText + "\nCounter target noncreature spell.",
	}, "f")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"AlternativeCosts: []cost.Alternative{",
		"Condition: cost.AlternativeConditionControlsCommander",
		"ExcludedSpellCardTypes: []types.Card{types.Creature}",
		"Primitive: game.CounterObject",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}

func TestCommanderAlternativeCostSiblings(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name string
		body string
	}{
		{name: "Deadly Rollick", body: "Exile target creature."},
		{name: "Flawless Maneuver", body: "Creatures you control gain indestructible until end of turn."},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       test.name,
				Layout:     "normal",
				TypeLine:   "Instant",
				ManaCost:   "{3}{B}",
				OracleText: commanderAlternativeCostText + "\n" + test.body,
			})
			if len(face.AlternativeCosts) != 1 {
				t.Fatalf("alternative costs = %#v", face.AlternativeCosts)
			}
		})
	}
}

func TestCommanderAlternativeCostDoesNotHideUnsupportedBody(t *testing.T) {
	t.Parallel()
	face := lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Deflecting Swat",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{2}{R}",
		OracleText: commanderAlternativeCostText + "\nYou may choose new targets for target spell or ability.",
	})
	if !face.empty() {
		t.Fatalf("partially lowered unsupported card: %#v", face)
	}
}
