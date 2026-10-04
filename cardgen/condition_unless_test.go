package cardgen

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestLowerResolvingStateUnless(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		condition compiler.CompiledCondition
		want      game.Condition
	}{
		{
			name: "controller selection",
			condition: compiler.CompiledCondition{
				Predicate: compiler.ConditionPredicateControllerControls,
				Selection: compiler.ConditionSelection{SubtypesAny: []string{"Villain"}},
				Threshold: 1,
			},
			want: game.Condition{ControlsMatching: opt.Val(game.SelectionCount{
				Selection: game.Selection{SubtypesAny: []types.Sub{"Villain"}},
				MinCount:  1,
			})},
		},
		{
			name: "graveyard cards",
			condition: compiler.CompiledCondition{
				Predicate: compiler.ConditionPredicateControllerGraveyardCardCountAtLeast,
				Threshold: 7,
			},
			want: game.Condition{Aggregates: []game.AggregateComparison{{
				Aggregate: game.AggregateControllerGraveyardCardCount, Op: compare.GreaterOrEqual, Value: 7,
			}}},
		},
		{
			name: "graveyard distinct mana values",
			condition: compiler.CompiledCondition{
				Predicate: compiler.ConditionPredicateControllerGraveyardManaValueCountAtLeast,
				Threshold: 5,
			},
			want: game.Condition{Aggregates: []game.AggregateComparison{{
				Aggregate: game.AggregateControllerGraveyardManaValueCount, Op: compare.GreaterOrEqual, Value: 5,
			}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.condition.Kind = compiler.ConditionUnless
			tt.condition.Negated = true
			tt.condition.Text = "diagnostic metadata only"
			got, ok := lowerCondition(tt.condition, conditionContextUnlessEffectGate)
			if !ok {
				t.Fatal("recognized state Unless failed to lower")
			}
			tt.want.Negate = true
			tt.want.Text = tt.condition.Text
			// Selection projection preserves empty slices; compare its semantics
			// separately from the aggregate-only conditions.
			if tt.want.ControlsMatching.Exists {
				if !got.ControlsMatching.Exists ||
					!reflect.DeepEqual(got.ControlsMatching.Val.Selection.SubtypesAny, tt.want.ControlsMatching.Val.Selection.SubtypesAny) ||
					got.ControlsMatching.Val.MinCount != tt.want.ControlsMatching.Val.MinCount ||
					!got.Negate || got.Text != tt.want.Text {
					t.Fatalf("condition = %#v, want %#v", got, tt.want)
				}
			} else if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("condition = %#v, want %#v", got, tt.want)
			}
			for _, ctx := range []conditionLoweringContext{
				conditionContextActivation, conditionContextEntryCounters,
				conditionContextSpellCostReduction, conditionContextInterveningTrigger,
				conditionContextEffectGate,
			} {
				if _, ok := lowerCondition(tt.condition, ctx); ok {
					t.Fatalf("Unless accepted outside resolving context: %v", ctx)
				}
			}
		})
	}
}

func TestResolvingUnlessRejectsOtherSemantics(t *testing.T) {
	t.Parallel()
	for _, predicate := range []compiler.ConditionPredicate{
		compiler.ConditionPredicateUnsupported,
		compiler.ConditionPredicateTargetControllerDoesNotPay,
		compiler.ConditionPredicatePriorInstructionAccepted,
		compiler.ConditionPredicateResultThisWay,
		compiler.ConditionPredicateObjectMatches,
		compiler.ConditionPredicateSpellWasKicked,
		compiler.ConditionPredicateEventHistory,
	} {
		condition := compiler.CompiledCondition{
			Kind: compiler.ConditionUnless, Negated: true, Predicate: predicate,
		}
		if _, ok := lowerCondition(condition, conditionContextUnlessEffectGate); ok {
			t.Fatalf("predicate %v accepted as pure state Unless", predicate)
		}
	}
	for _, condition := range []compiler.CompiledCondition{
		{Kind: compiler.ConditionUnless, Predicate: compiler.ConditionPredicateControllerGraveyardCardCountAtLeast, Threshold: 7},
		{Kind: compiler.ConditionUnless, Negated: true, Intervening: true, Predicate: compiler.ConditionPredicateControllerGraveyardCardCountAtLeast, Threshold: 7},
		{Kind: compiler.ConditionUnless, Negated: true, Predicate: compiler.ConditionPredicateControllerControls, Threshold: 1},
		{Kind: compiler.ConditionUnless, Negated: true, Predicate: compiler.ConditionPredicateControllerGraveyardCardCountAtLeast, Threshold: -1},
		{Kind: compiler.ConditionUnless, Negated: true, SourceInGraveyard: true,
			Predicate: compiler.ConditionPredicateControllerControls, Threshold: 1,
			Selection: compiler.ConditionSelection{SubtypesAny: []string{"Villain"}}},
		{Kind: compiler.ConditionUnless, Negated: true,
			Predicate: compiler.ConditionPredicateControllerControls, Threshold: 1,
			Selection: compiler.ConditionSelection{SubtypesAny: []string{"Pirate"}, ExcludeSource: true}},
		{Kind: compiler.ConditionUnless, Negated: true, Predicate: compiler.ConditionPredicateControllerGraveyardCardCountAtLeast, Threshold: 7,
			Selection: compiler.ConditionSelection{RequiredTypes: []types.Card{types.Creature}}},
	} {
		if _, ok := lowerCondition(condition, conditionContextUnlessEffectGate); ok {
			t.Fatalf("incomplete/qualified condition accepted: %#v", condition)
		}
	}
}

func TestResolvingUnlessClauseOwnership(t *testing.T) {
	t.Parallel()
	span := func(start, end int) shared.Span {
		return shared.Span{Start: shared.Position{Offset: start}, End: shared.Position{Offset: end}}
	}
	condition := compiler.CompiledCondition{
		Kind: compiler.ConditionUnless, Negated: true,
		Predicate: compiler.ConditionPredicateControllerGraveyardCardCountAtLeast,
		Threshold: 7, Span: span(21, 29),
	}
	effect := compiler.CompiledEffect{Span: span(20, 50), ClauseSpan: span(20, 50), VerbSpan: span(30, 35)}
	other := compiler.CompiledEffect{Span: span(0, 20), ClauseSpan: span(0, 20), VerbSpan: span(0, 4)}
	for _, tt := range []struct {
		name       string
		effects    []compiler.CompiledEffect
		conditions []compiler.CompiledCondition
		reason     string
	}{
		{"one owned clause", []compiler.CompiledEffect{other, effect}, []compiler.CompiledCondition{condition}, ""},
		{"no owner", []compiler.CompiledEffect{other}, []compiler.CompiledCondition{condition}, effectGateCategoryNoClause},
		{"shared leading group", []compiler.CompiledEffect{effect, effect}, []compiler.CompiledCondition{condition}, effectGateCategoryMultiClause},
		{"multiple conditions", []compiler.CompiledEffect{effect}, []compiler.CompiledCondition{condition, condition}, effectGateCategoryMultiCondition},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gates, reason, ok := matchOrderedSequenceEffectConditions(tt.effects, tt.conditions)
			if ok != (tt.reason == "") || reason != tt.reason {
				t.Fatalf("ok=%v reason=%q, want reason=%q", ok, reason, tt.reason)
			}
			if ok && (len(gates) != 1 || !gates[1].Condition.Val.Negate) {
				t.Fatalf("gates = %#v, want only the second clause negated", gates)
			}
			if ok {
				if _, reason, accepted := matchSequenceEffectConditions(tt.effects, tt.conditions); accepted || reason != effectGateCategoryKind {
					t.Fatalf("specialized matcher accepted Unless: accepted=%v reason=%q", accepted, reason)
				}
			}
		})
	}
}

func TestResolvingUnlessCardDefs(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, text, gate string
	}{
		{"selection", "Draw a card. You lose 2 life unless you control a Villain.", "ControlsMatching.Val.Selection.SubtypesAny[0] = types.Villain"},
		{"graveyard", "Draw a card. Then discard a card unless there are seven or more cards in your graveyard.", "Aggregates[0].Value = 7"},
		{"mana values", "Draw two cards. Then discard a card unless there are five or more mana values among cards in your graveyard.", "Aggregates[0].Value = 5"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{Name: "Unless Test", Layout: "normal", TypeLine: "Instant", OracleText: tt.text}
			prefix := "CardDef.CardFace.SpellAbility.Val.Modes[0].Sequence[1].Condition.Val.Condition.Val."
			assertCardPaths(t, card, prefix+"Negate = true", prefix+tt.gate)
			assertCardPathsAbsent(t, card, "Sequence[0].Condition", "Sequence[1].ResultGate")
		})
	}
}

func TestResolvingUnlessFullConsumption(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"Draw a card. Discard a card unless you control a Villain and you control an artifact.",
		"Draw a card. Discard a card unless you control a Villain or pay 2 life.",
		"Draw a card. Discard a card unless you pay {1}.",
		"Draw a card. Discard a card unless you choose to.",
		"Draw a card. Discard a card unless a land card was discarded this way.",
		"Draw a card. Discard a card unless you control a Villain with a hat.",
		"Unless you control a Villain, draw a card, then discard a card.",
		"Draw a card, then discard a card unless there are seven or more cards in your graveyard.",
		"Draw a card. You lose 2 life unless this card is in your graveyard and you control a Villain.",
		"Creatures you control gain indestructible until end of turn.\nAddendum — Unless you control a creature with power 2 or greater, put a +1/+1 counter on each of those creatures and they gain vigilance until end of turn.",
		"Unless you control a creature with power 2 or greater, put a +1/+1 counter on each of up to two target creatures you control.",
		"Discard a card unless there are seven or more cards in your graveyard. Otherwise, draw a card.",
		"Draw a card. You lose 2 life unless you control another Pirate.",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			assertCardUnsupported(t, &ScryfallCard{Name: "Unless Near Miss", Layout: "normal", TypeLine: "Instant", OracleText: text})
		})
	}
}

func TestResolvingUnlessShellRouting(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, text, path, typeLine, layout string
		index                              int
	}{
		{"single effect", "You lose 2 life unless you control a Villain.", "SpellAbility.Val", "Instant", "normal", 0},
		{"activated body", "{1}: Draw a card. Then discard a card unless there are seven or more cards in your graveyard.", "ActivatedAbilities[0].Content", "Artifact", "normal", 1},
		{"chapter body", "I — Draw two cards. Then discard a card unless there are five or more mana values among cards in your graveyard.", "ChapterAbilities[0].Content", "Enchantment — Saga", "saga", 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{Name: "Unless Shell Test", Layout: tt.layout, TypeLine: tt.typeLine, OracleText: tt.text}
			path := fmt.Sprintf("%s.Modes[0].Sequence[%d].Condition.Val.Condition.Val.Negate = true", tt.path, tt.index)
			assertCardPaths(t, card, path)
			assertCardPathsAbsent(t, card, "ActivationCondition", "InterveningCondition")
		})
	}
}

func TestResolvingUnlessPreservesIndependentResultGate(t *testing.T) {
	t.Parallel()
	gate := game.InstructionResultGate{Key: "prior-choice", Succeeded: game.TriTrue}
	sequence := []game.Instruction{{
		Primitive:  game.Discard{Amount: game.Fixed(1), Player: game.ControllerReference()},
		ResultGate: opt.Val(gate),
	}}
	condition := game.EffectCondition{Condition: opt.Val(game.Condition{
		Negate: true, Aggregates: []game.AggregateComparison{{
			Aggregate: game.AggregateControllerGraveyardCardCount, Op: compare.GreaterOrEqual, Value: 7,
		}},
	})}
	if reason := applySequenceClauseGates(sequence, 0, map[int]game.EffectCondition{0: condition}, nil, nil); reason != "" {
		t.Fatalf("gate application failed: %s", reason)
	}
	if !reflect.DeepEqual(sequence[0].ResultGate.Val, gate) || !sequence[0].Condition.Val.Condition.Val.Negate {
		t.Fatalf("instruction = %#v, want both independent gates", sequence[0])
	}
	if applyEffectConditionGate(sequence, &condition) {
		t.Fatal("second condition overwrote an existing condition")
	}
}

func TestResolvingUnlessRecognizedSyntax(t *testing.T) {
	t.Parallel()
	document, diagnostics := parser.Parse(
		"Draw two cards. Then discard a card unless there are five or more mana values among cards in your graveyard.",
		parser.Context{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("parse diagnostics = %#v", diagnostics)
	}

	compilation, diagnostics := compiler.Compile(document, compiler.Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("compile diagnostics = %#v", diagnostics)
	}
	content := compilation.Abilities[0].Content
	if len(content.Conditions) != 1 || content.Conditions[0].Kind != compiler.ConditionUnless ||
		!content.Conditions[0].Negated ||
		content.Conditions[0].Predicate != compiler.ConditionPredicateControllerGraveyardManaValueCountAtLeast {
		t.Fatalf("conditions = %#v, want recognized negated mana-value Unless", content.Conditions)
	}
}

func TestResolvingUnlessDoesNotBypassSequenceBlockers(t *testing.T) {
	t.Parallel()
	assertCardUnsupported(t, &ScryfallCard{
		Name: "The Spot's Portal", Layout: "normal", TypeLine: "Instant", ManaCost: "{2}{B}",
		OracleText: "Put target creature on the bottom of its owner's library. You lose 2 life unless you control a Villain.",
	}, "structural — inherited target not remappable")
}

func TestResolvingUnlessManaAbilityFailsClosed(t *testing.T) {
	t.Parallel()
	assertCardUnsupported(t, &ScryfallCard{
		Name: "Unless Mana Test", Layout: "normal", TypeLine: "Artifact",
		OracleText: "{T}: Add {B}. You lose 2 life unless you control a Villain.",
	}, "unsupported activation condition")
}

func TestResolvingUnlessSourceExclusionFailsClosed(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ name, text string }{
		{"Fathom Fleet Boarder", "When this creature enters, you lose 2 life unless you control another Pirate."},
		{"Reaver Drone", "Devoid (This card has no color.)\nAt the beginning of your upkeep, you lose 1 life unless you control another colorless creature."},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertCardUnsupported(t, &ScryfallCard{
				Name: tt.name, Layout: "normal", TypeLine: "Creature — Eldrazi Pirate",
				OracleText: tt.text, Power: new("2"), Toughness: new("2"),
			})
		})
	}
}

func TestResolvingUnlessOtherwiseFailsClosed(t *testing.T) {
	t.Parallel()
	assertCardUnsupported(t, &ScryfallCard{
		Name: "Unless Branch Test", Layout: "normal", TypeLine: "Creature — Rogue",
		OracleText: "When this creature enters, discard a card unless there are seven or more cards in your graveyard. Otherwise, draw a card.",
		Power:      new("1"), Toughness: new("1"),
	}, effectGateCategoryUnlessBranch)
}
