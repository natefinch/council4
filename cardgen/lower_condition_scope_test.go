package cardgen

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

func assertConditionConsumer(t *testing.T, publisher, consumer game.Instruction, negate bool) {
	t.Helper()
	if publisher.PublishCondition == "" || consumer.Condition.Exists ||
		consumer.ConditionGate != publisher.PublishCondition || consumer.ConditionGateNegate != negate {
		t.Fatal("consumer must use the publisher's captured condition with the expected polarity")
	}
}

func TestConditionScopeExpandedUnless(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name: "Expanded Unless", Layout: "normal", TypeLine: "Instant",
		OracleText: "Unless you control a creature with power 2 or greater, put a +1/+1 counter on each of up to two target creatures you control.",
	}
	assertCardPaths(t, card,
		"SpellAbility.Val.Modes[0].Sequence[0].Condition.Val.Condition.Val.Negate = true",
		`SpellAbility.Val.Modes[0].Sequence[0].PublishCondition = "condition-0"`,
		`SpellAbility.Val.Modes[0].Sequence[1].ConditionGate = "condition-0"`,
	)
	assertCardPathsAbsent(t, card, "Sequence[1].Condition.Val", "Sequence[1].ResultGate")
}

func TestConditionOwnershipIgnoresTextAndPositions(t *testing.T) {
	t.Parallel()
	effects := []compiler.CompiledEffect{{ClauseID: 11}, {ClauseID: 23}}
	condition := compiler.CompiledCondition{
		Kind: compiler.ConditionIf, Predicate: compiler.ConditionPredicateControllerHandEmpty,
		Ownership: parser.ConditionOwnership{Scope: parser.ConditionScopeGroup, ClauseIDs: []int{11, 23}},
		Text:      "diagnostic-only", Span: shared.Span{Start: shared.Position{Offset: 10000}},
	}
	gates, reason, ok := matchOrderedSequenceEffectConditions(effects, []compiler.CompiledCondition{condition})
	if !ok || len(gates) != 2 || reason != "" {
		t.Fatalf("typed matching: gates=%#v reason=%q ok=%v", gates, reason, ok)
	}
	for _, ids := range [][]int{nil, {11}, {23, 11}, {11, 11}, {11, 99}} {
		malformed := condition
		malformed.Ownership.ClauseIDs = ids
		if _, reason, ok := matchOrderedSequenceEffectConditions(effects, []compiler.CompiledCondition{malformed}); ok || reason == "" {
			t.Fatalf("malformed ownership %v accepted", ids)
		}
	}
}

func TestConditionScopeActivationRestrictionIsSeparate(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name: "Activation Restriction", Layout: "normal", TypeLine: "Enchantment",
		OracleText: "{1}: Draw a card. Activate only if you have no cards in hand.",
	}
	assertCardPaths(t, card,
		"ActivatedAbilities[0].ActivationCondition.Val.ControllerHandEmpty = true",
	)
	assertCardPathsAbsent(t, card, "ActivatedAbilities[0].Content.Modes[0].Sequence[0].Condition")
}

func TestResolvingBodyOwnershipFailsClosed(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name  string
		scope parser.ConditionScope
		ids   []int
		kind  compiler.ConditionKind
		mana  bool
		want  bool
	}{
		{"owned", parser.ConditionScopeClause, []int{1}, compiler.ConditionIf, false, true},
		{"unknown", parser.ConditionScopeUnknown, nil, compiler.ConditionIf, false, false},
		{"ambiguous", parser.ConditionScopeUnsupported, nil, compiler.ConditionIf, false, false},
		{"unavailable", parser.ConditionScopeClause, []int{99}, compiler.ConditionIf, false, false},
		{"activation only", parser.ConditionScopeClause, []int{1}, compiler.ConditionOnlyIf, false, false},
		{"mixed mana", parser.ConditionScopeClause, []int{1}, compiler.ConditionIf, true, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			content := compiler.AbilityContent{
				Effects: []compiler.CompiledEffect{{ClauseID: 1, Kind: compiler.EffectDraw}},
				Conditions: []compiler.CompiledCondition{{
					Kind: tt.kind, Predicate: compiler.ConditionPredicateControllerHandEmpty,
					Ownership: parser.ConditionOwnership{Scope: tt.scope, ClauseIDs: tt.ids},
				}},
			}
			if tt.mana {
				content.Effects = append(content.Effects, compiler.CompiledEffect{ClauseID: 2, Kind: compiler.EffectAddMana})
			}
			if got := conditionsOwnedByResolvingBody(content); got != tt.want {
				t.Fatalf("body-owned=%v, want %v", got, tt.want)
			}
		})
	}
}

func TestConditionScopeNearMisses(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"If you have no cards in hand and win the game, draw a card, then draw a card.",
		"Draw a card, then discard a card unless you control a Villain.",
		"Unless you pay {1}, draw a card, then discard a card.",
		"Unless you control another Pirate with a hat, draw a card, then discard a card.",
		"Unless you control a Villain, draw a card, then discard a card. Otherwise, you gain 2 life.",
	} {
		assertCardUnsupported(t, &ScryfallCard{Name: "Scope Near Miss", Layout: "normal", TypeLine: "Instant", OracleText: text})
	}
	for _, text := range []string{
		"{T}: Unless you control a Villain, add {G}.",
		"{T}: Draw a card. Unless you control a Villain, add {G}.",
	} {
		assertCardUnsupported(t, &ScryfallCard{Name: "Mana Scope Near Miss", Layout: "normal", TypeLine: "Artifact", OracleText: text})
	}
}

func TestActivatedResolutionOrdinalRemainsUnsupported(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name: "Inner-Flame Igniter", Layout: "normal", TypeLine: "Creature - Elemental Warrior",
		OracleText: "{2}{R}: Creatures you control get +1/+0 until end of turn. If this is the third time this ability has resolved this turn, creatures you control gain first strike until end of turn.",
	}
	assertCardUnsupported(t, card)
	_, diagnostics := lowerExecutableFaces(card)
	if len(diagnostics) != 1 || diagnostics[0].Detail != "resolution ordinals are modeled only for triggered abilities" {
		t.Fatalf("diagnostics=%#v, want explicit unmodeled activated resolution count", diagnostics)
	}
}

func TestConditionEvaluationPreservesCastTargetGates(t *testing.T) {
	t.Parallel()
	for _, condition := range []game.Condition{{SpellWasKicked: true}, {GiftPromised: true}, {SpellWasBargained: true}} {
		publisher := game.Instruction{
			Primitive:        game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
			Condition:        opt.Val(game.EffectCondition{Condition: opt.Val(condition)}),
			PublishCondition: "cast-branch",
		}
		consumer := game.Instruction{
			Primitive:     game.Destroy{Object: game.TargetPermanentReference(0)},
			ConditionGate: "cast-branch",
		}
		expected, _, _ := castBranchGate(&publisher)
		targets := []game.TargetSpec{{Allow: game.TargetAllowPermanent, MinTargets: 1, MaxTargets: 1}}
		got, ok := assignTargetGates(targets, []game.Instruction{publisher, consumer})
		if !ok || got[0].Gate != expected {
			t.Fatalf("cast target gate=%#v ok=%v, want %v", got, ok, expected)
		}
		consumer.ConditionGateNegate = true
		complement, complementOK := complementCastBranchGate(expected)
		got, ok = assignTargetGates(targets, []game.Instruction{publisher, consumer})
		if !ok || !complementOK || got[0].Gate != complement {
			t.Fatalf("complemented cast target gate=%#v ok=%v, want %v", got, ok, complement)
		}
		if _, ok := assignTargetGates(targets, []game.Instruction{consumer}); ok {
			t.Fatal("unavailable cast-branch evaluation was accepted")
		}
	}
}
