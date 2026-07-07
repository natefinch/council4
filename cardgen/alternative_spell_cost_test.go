package cardgen

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
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

const damnText = `Destroy target creature. A creature destroyed this way can't be regenerated.
Overload {2}{W}{W} (You may cast this spell for its overload cost. If you do, change "target" in its text to "each.")`

func TestLowerDamnDestroyRiderAndOverload(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Damn",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{B}{B}",
		OracleText: damnText,
	})
	// Normal form: destroy one target creature, with the regeneration rider
	// folded onto the destroy.
	normal := face.SpellAbility.Val.Modes[0]
	if len(normal.Targets) != 1 {
		t.Fatalf("normal targets = %#v, want one", normal.Targets)
	}
	normalDestroy, ok := normal.Sequence[0].Primitive.(game.Destroy)
	if !ok || normalDestroy.Object != game.TargetPermanentReference(0) || !normalDestroy.PreventRegeneration {
		t.Fatalf("normal primitive = %#v, want target destroy preventing regeneration", normal.Sequence[0].Primitive)
	}
	// Overload form: destroy each creature on the battlefield, still preventing
	// regeneration.
	if !face.Overload.Exists || !slices.Equal(face.Overload.Val.Cost, cost.Mana{cost.O(2), cost.W, cost.W}) {
		t.Fatalf("overload = %#v", face.Overload)
	}
	overloaded := face.Overload.Val.SpellAbility.Modes[0]
	if len(overloaded.Targets) != 0 {
		t.Fatalf("overload targets = %#v, want none", overloaded.Targets)
	}
	overloadDestroy, ok := overloaded.Sequence[0].Primitive.(game.Destroy)
	if !ok || !overloadDestroy.Group.Valid() || !overloadDestroy.PreventRegeneration {
		t.Fatalf("overload primitive = %#v, want mass destroy preventing regeneration", overloaded.Sequence[0].Primitive)
	}
	if selection := overloadDestroy.Group.Selection(); !slices.Equal(selection.RequiredTypes, []types.Card{types.Creature}) {
		t.Fatalf("overload selection = %#v, want creatures", overloadDestroy.Group.Selection())
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

// TestLowerOverloadPumpBecomesGroupContinuous verifies that overloading a
// single-target fixed power/toughness pump ("Target creature you control gets
// +2/+2 until end of turn.") turns it into the group ApplyContinuous form the
// runtime already uses for mass pumps, addressing every creature the caster
// controls instead of one target.
func TestLowerOverloadPumpBecomesGroupContinuous(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Stirring Address",
		Layout:   "normal",
		TypeLine: "Instant",
		ManaCost: "{1}{W}",
		OracleText: "Target creature you control gets +2/+2 until end of turn.\n" +
			"Overload {4}{W} (You may cast this spell for its overload cost. If you do, change \"target\" in its text to \"each.\")",
	})
	if !face.Overload.Exists {
		t.Fatalf("overload = %#v", face.Overload)
	}
	normalModify, ok := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.ModifyPT)
	if !ok || normalModify.Object != game.TargetPermanentReference(0) {
		t.Fatalf("normal primitive = %#v", face.SpellAbility.Val.Modes[0].Sequence[0].Primitive)
	}
	overloaded := face.Overload.Val.SpellAbility.Modes[0]
	if len(overloaded.Targets) != 0 {
		t.Fatalf("overload targets = %#v, want none", overloaded.Targets)
	}
	apply, ok := overloaded.Sequence[0].Primitive.(game.ApplyContinuous)
	if !ok || len(apply.ContinuousEffects) != 1 {
		t.Fatalf("overload primitive = %#v", overloaded.Sequence[0].Primitive)
	}
	continuous := apply.ContinuousEffects[0]
	selection := continuous.Group.Selection()
	if continuous.Layer != game.LayerPowerToughnessModify ||
		continuous.PowerDelta != 2 || continuous.ToughnessDelta != 2 ||
		selection.Controller != game.ControllerYou ||
		!slices.Equal(selection.RequiredTypes, []types.Card{types.Creature}) {
		t.Fatalf("overload continuous effect = %#v (group %#v)", continuous, selection)
	}
}

func TestLowerOverloadIsTextBlindAndFailsClosed(t *testing.T) {
	t.Parallel()
	normal := opt.Val(game.Mode{
		Targets: []game.TargetSpec{{
			MinTargets: 1,
			MaxTargets: 1,
			Allow:      game.TargetAllowPermanent,
			Selection: opt.Val(game.Selection{
				RequiredTypesAny: []types.Card{types.Artifact},
				Controller:       game.ControllerNotYou,
			}),
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

func TestGenerateDreadReturnFlashbackSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Dread Return",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{2}{B}{B}",
		OracleText: "Return target creature card from your graveyard to the battlefield.\nFlashback—Sacrifice three creatures. (You may cast this card from your graveyard for its flashback cost. Then exile it.)",
	}, "dreadReturn")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.SimpleKeyword{Kind: game.Flashback}",
		`Label: "Flashback"`,
		"Mechanic: cost.AlternativeMechanicFlashback,",
		"AdditionalCosts: []cost.Additional{",
		"Kind:               cost.AdditionalSacrifice,",
		"PermanentType:      types.Creature,",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
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
	lowered, diagnostic := lowerSpellAlternativeCost("", compiler.CompiledAbility{
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

func TestLowerDeadlyRollick(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Deadly Rollick",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{3}{B}",
		OracleText: commanderAlternativeCostText + "\nExile target creature.",
	}
	face := lowerSingleFace(t, card)
	if len(face.AlternativeCosts) != 1 ||
		face.AlternativeCosts[0].Condition != cost.AlternativeConditionControlsCommander ||
		face.AlternativeCosts[0].ManaCost.Exists {
		t.Fatalf("alternative costs = %#v", face.AlternativeCosts)
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %#v, want one creature target", mode.Targets)
	}
	target := mode.Targets[0]
	if target.MinTargets != 1 ||
		target.MaxTargets != 1 ||
		target.Allow != game.TargetAllowPermanent ||
		!slices.Equal(target.Selection.Val.RequiredTypesAny, []types.Card{types.Creature}) {
		t.Fatalf("target = %#v, want exact one creature", target)
	}
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence = %#v, want one exile", mode.Sequence)
	}
	exile, ok := mode.Sequence[0].Primitive.(game.Exile)
	if !ok || exile.Object != game.TargetPermanentReference(0) {
		t.Fatalf("primitive = %#v, want exile of target reference 0", mode.Sequence[0].Primitive)
	}

	source, diagnostics, err := GenerateExecutableCardSource(card, "d")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"Condition: cost.AlternativeConditionControlsCommander",
		"RequiredTypesAny: []types.Card{types.Creature}",
		"Primitive: game.Exile",
		"Object: game.TargetPermanentReference(0)",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}

func TestLowerCommanderCreatureExilePreservesAdditionalCost(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Test Commander Exile",
		Layout:   "normal",
		TypeLine: "Instant",
		ManaCost: "{3}{B}",
		OracleText: "As an additional cost to cast this spell, discard a card.\n" +
			commanderAlternativeCostText + "\nExile target creature.",
	})
	if len(face.AlternativeCosts) != 1 ||
		face.AlternativeCosts[0].Condition != cost.AlternativeConditionControlsCommander {
		t.Fatalf("alternative costs = %#v", face.AlternativeCosts)
	}
	if len(face.AdditionalCosts) != 1 ||
		face.AdditionalCosts[0].Kind != cost.AdditionalDiscard ||
		face.AdditionalCosts[0].Amount != 1 {
		t.Fatalf("additional costs = %#v", face.AdditionalCosts)
	}
}

func TestCommanderAlternativeCostSiblings(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name string
		body string
	}{
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
	for _, test := range []struct {
		name string
		body string
	}{
		{name: "unsupported exile player graveyard", body: "Target player exiles a card from their graveyard."},
		{name: "subtype unbounded exile", body: "Exile any number of target Goblin creatures."},
		{name: "nonpermanent exile", body: "Exile target spell."},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
				Name:       "Test Commander Alternative",
				Layout:     "normal",
				TypeLine:   "Instant",
				ManaCost:   "{2}{R}",
				OracleText: commanderAlternativeCostText + "\n" + test.body,
			})
			if !face.empty() {
				t.Fatalf("partially lowered unsupported card: %#v", face)
			}
		})
	}
}

func TestLowerForceOfWillPitch(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Force of Will",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{3}{U}{U}",
		OracleText: "You may pay 1 life and exile a blue card from your hand rather than pay this spell's mana cost.\nCounter target spell.",
	})
	if len(face.AlternativeCosts) != 1 {
		t.Fatalf("alternative costs = %#v, want one", face.AlternativeCosts)
	}
	alt := face.AlternativeCosts[0]
	if alt.ManaCost.Exists {
		t.Fatalf("pitch alternative should carry no mana cost: %#v", alt)
	}
	if alt.Condition != cost.AlternativeConditionNone {
		t.Fatalf("condition = %v, want none", alt.Condition)
	}
	if len(alt.AdditionalCosts) != 2 {
		t.Fatalf("additional costs = %#v, want pay-life + exile", alt.AdditionalCosts)
	}
	life := alt.AdditionalCosts[0]
	if life.Kind != cost.AdditionalPayLife || life.Amount != 1 {
		t.Fatalf("life cost = %#v, want pay 1 life", life)
	}
	exile := alt.AdditionalCosts[1]
	if exile.Kind != cost.AdditionalExile ||
		exile.Source != zone.Hand ||
		!exile.MatchCardColor ||
		exile.CardColor != color.Blue ||
		exile.Amount != 1 {
		t.Fatalf("exile cost = %#v, want exile one blue card from hand", exile)
	}
}

func TestLowerForceOfNegationPitchAndCounterExile(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Force of Negation",
		Layout:   "normal",
		TypeLine: "Instant",
		ManaCost: "{1}{U}{U}",
		OracleText: "If it's not your turn, you may exile a blue card from your hand rather than pay this spell's mana cost.\n" +
			"Counter target noncreature spell. If that spell is countered this way, exile it instead of putting it into its owner's graveyard.",
	})
	if len(face.AlternativeCosts) != 1 {
		t.Fatalf("alternative costs = %#v, want one", face.AlternativeCosts)
	}
	alt := face.AlternativeCosts[0]
	if alt.Condition != cost.AlternativeConditionNotYourTurn {
		t.Fatalf("condition = %v, want not-your-turn", alt.Condition)
	}
	if len(alt.AdditionalCosts) != 1 ||
		alt.AdditionalCosts[0].Kind != cost.AdditionalExile ||
		alt.AdditionalCosts[0].CardColor != color.Blue {
		t.Fatalf("additional costs = %#v, want exile one blue card", alt.AdditionalCosts)
	}
	counter, ok := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.CounterObject)
	if !ok || !counter.ExileInstead {
		t.Fatalf("primitive = %#v, want counter with exile instead", face.SpellAbility.Val.Modes[0].Sequence[0].Primitive)
	}
}

func TestLowerFoilDiscardAlternativeCost(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Foil",
		Layout:   "normal",
		TypeLine: "Instant",
		ManaCost: "{2}{U}{U}",
		OracleText: "You may discard an Island card and another card rather than pay this spell's mana cost.\n" +
			"Counter target spell.",
	})
	if len(face.AlternativeCosts) != 1 {
		t.Fatalf("alternative costs = %#v, want one", face.AlternativeCosts)
	}
	alt := face.AlternativeCosts[0]
	if alt.ManaCost.Exists {
		t.Fatalf("discard alternative should carry no mana cost: %#v", alt)
	}
	if alt.Condition != cost.AlternativeConditionNone {
		t.Fatalf("condition = %v, want none", alt.Condition)
	}
	if len(alt.AdditionalCosts) != 2 {
		t.Fatalf("additional costs = %#v, want two discards", alt.AdditionalCosts)
	}
	island := alt.AdditionalCosts[0]
	if island.Kind != cost.AdditionalDiscard ||
		island.Source != zone.Hand ||
		island.Amount != 1 ||
		island.SubtypesAny != (cost.SubtypeSet{types.Island}) {
		t.Fatalf("first discard = %#v, want discard one Island card from hand", island)
	}
	other := alt.AdditionalCosts[1]
	if other.Kind != cost.AdditionalDiscard ||
		other.Source != zone.Hand ||
		other.Amount != 1 ||
		other.SubtypesAny != (cost.SubtypeSet{}) {
		t.Fatalf("second discard = %#v, want discard one unfiltered card from hand", other)
	}
	if _, ok := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.CounterObject); !ok {
		t.Fatalf("primitive = %#v, want counter object", face.SpellAbility.Val.Modes[0].Sequence[0].Primitive)
	}
}
