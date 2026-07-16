package compiler

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestCompileResolutionConditionIsNotIntervening(t *testing.T) {
	t.Parallel()
	source := "When this creature dies, draw a card if you control a Forest."
	compilation, diagnostics := compileSource(source, pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Trigger == nil {
		t.Fatal("missing trigger")
	}
	if ability.Trigger.Condition != nil {
		t.Fatalf("resolution condition became trigger condition: %#v", ability.Trigger.Condition)
	}
	if len(ability.Content.Conditions) != 1 || ability.Content.Conditions[0].Intervening {
		t.Fatalf("conditions = %#v", ability.Content.Conditions)
	}
}

func TestCompileEntersTappedUnlessCondition(t *testing.T) {
	t.Parallel()
	source := "This land enters tapped unless you control two or more basic lands."
	compilation, diagnostics := compileSource(source, pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Kind != AbilityReplacement {
		t.Fatalf("kind = %v, want AbilityReplacement", ability.Kind)
	}
	if len(ability.Content.Effects) != 1 || ability.Content.Effects[0].Kind != EffectEnterTapped {
		t.Fatalf("effects = %#v", ability.Content.Effects)
	}
	if len(ability.Content.Conditions) != 1 ||
		ability.Content.Conditions[0].Kind != ConditionUnless ||
		ability.Content.Conditions[0].Text != "unless you control two or more basic lands" {
		t.Fatalf("conditions = %#v", ability.Content.Conditions)
	}
	if len(ability.Content.References) != 1 || ability.Content.References[0].Kind != ReferenceThisObject {
		t.Fatalf("references = %#v", ability.Content.References)
	}
}

func TestCompileEntersTappedIfControlCondition(t *testing.T) {
	t.Parallel()
	source := "If you control two or more other lands, this land enters tapped."
	compilation, diagnostics := compileSource(source, pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Kind != AbilityReplacement {
		t.Fatalf("kind = %v, want AbilityReplacement", ability.Kind)
	}
	if len(ability.Content.Effects) != 1 || ability.Content.Effects[0].Kind != EffectEnterTapped {
		t.Fatalf("effects = %#v", ability.Content.Effects)
	}
	if len(ability.Content.Conditions) != 1 {
		t.Fatalf("conditions = %#v", ability.Content.Conditions)
	}
	condition := ability.Content.Conditions[0]
	// The leading "If" form gates entry on the condition holding, so the
	// condition is an unnegated "if" rather than the trailing "unless" form.
	if condition.Kind != ConditionIf ||
		condition.Intervening ||
		condition.Negated ||
		condition.Predicate != ConditionPredicateControllerControls {
		t.Fatalf("condition = %#v, want unnegated if controller-controls", condition)
	}
}

func TestCompileEntersTappedUnlessLegendaryCreature(t *testing.T) {
	t.Parallel()
	source := "Minas Tirith enters tapped unless you control a legendary creature."
	compilation, diagnostics := compileSource(source, pipelineContext{CardName: "Minas Tirith"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Kind != AbilityReplacement {
		t.Fatalf("kind = %v, want AbilityReplacement", ability.Kind)
	}
	if len(ability.Content.Conditions) != 1 {
		t.Fatalf("conditions = %#v", ability.Content.Conditions)
	}
	condition := ability.Content.Conditions[0]
	if condition.Kind != ConditionUnless ||
		condition.Predicate != ConditionPredicateControllerControls ||
		!condition.Negated {
		t.Fatalf("condition = %#v, want negated unless controller-controls", condition)
	}
	selection := condition.Selection
	if len(selection.RequiredTypes) != 1 || selection.RequiredTypes[0] != types.Creature {
		t.Fatalf("selection = %#v, want single creature type", selection)
	}
	if len(selection.Supertypes) != 1 || selection.Supertypes[0] != types.Legendary {
		t.Fatalf("selection = %#v, want legendary supertype", selection)
	}
}

func TestCompileConditionWorldSupertypeFailsClosed(t *testing.T) {
	t.Parallel()
	// "world" is a recognized supertype atom but is outside the closed condition
	// vocabulary, so the controls predicate must not compile from it.
	source := "This land enters tapped unless you control a world enchantment."
	compilation, _ := compileSource(source, pipelineContext{CardName: "Test Land"})
	if len(compilation.Abilities) == 0 {
		t.Fatalf("compilation = %#v", compilation)
	}
	for _, condition := range compilation.Abilities[0].Content.Conditions {
		if condition.Predicate == ConditionPredicateControllerControls {
			t.Fatalf("condition = %#v, want world supertype to fail closed", condition)
		}
	}
}

func TestCompileArtifactAndEnchantmentEntersTappedReference(t *testing.T) {
	t.Parallel()
	tests := []string{
		"This artifact enters tapped.",
		"This enchantment enters tapped.",
	}
	for _, source := range tests {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(source, pipelineContext{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			ability := compilation.Abilities[0]
			if ability.Kind != AbilityReplacement {
				t.Fatalf("kind = %v, want AbilityReplacement", ability.Kind)
			}
			if len(ability.Content.References) != 1 || ability.Content.References[0].Kind != ReferenceThisObject {
				t.Fatalf("references = %#v", ability.Content.References)
			}
		})
	}
}

func TestCompileUnsupportedConstruct(t *testing.T) {
	t.Parallel()
	source := "Daybound"
	compilation, diagnostics := compileSource(source, pipelineContext{})
	if len(compilation.Abilities) != 1 {
		t.Fatalf("abilities = %#v", compilation.Abilities)
	}
	if len(diagnostics) != 1 ||
		diagnostics[0].Severity != shared.SeverityWarning ||
		diagnostics[0].Span != compilation.Abilities[0].Span {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
}

func TestCompileScryfallCacheHasNoSilentAbilities(t *testing.T) {
	t.Parallel()
	cache := filepath.Join("..", "..", ".cardwork", "deck", "cache", "scryfall")
	paths, err := filepath.Glob(filepath.Join(cache, "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Skip("local Scryfall cache is not present")
	}

	var texts int
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var card cachedParserCard
		if err := json.Unmarshal(data, &card); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		check := func(name, typeLine, source string) {
			t.Helper()
			if source == "" {
				return
			}
			texts++
			context := pipelineContext{
				CardName:         name,
				InstantOrSorcery: typeLine == "Instant" || typeLine == "Sorcery",
				Planeswalker:     typeLine == "Planeswalker" || typeLine == "Legendary Planeswalker",
			}
			compilation, diagnostics := compileSource(source, context)
			for _, diagnostic := range diagnostics {
				if diagnostic.Severity == shared.SeverityError {
					t.Fatalf("%s: compiler error = %#v", name, diagnostic)
				}
			}
			for _, ability := range compilation.Abilities {
				if ability.Kind == AbilityReminder {
					continue
				}
				meaningful := len(ability.Content.Effects) > 0 ||
					len(ability.Content.Keywords) > 0 ||
					len(ability.Content.Modes) > 0
				if meaningful || hasDiagnosticForSpan(diagnostics, ability.Span) {
					continue
				}
				t.Fatalf("%s: silently uncompiled ability %q", name, ability.Text)
			}
		}
		check(card.Name, card.TypeLine, card.OracleText)
		for _, face := range card.CardFaces {
			check(face.Name, face.TypeLine, face.OracleText)
		}
	}
	if texts != 59 {
		t.Fatalf("checked %d non-empty Oracle texts, want 59", texts)
	}
}

func TestCompileConditionsRecognizesClosedSemanticPredicates(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		source    string
		kind      ConditionKind
		predicate ConditionPredicate
		negated   bool
	}{
		{"static Selection", "As long as you control another red creature, this creature has flying.", ConditionAsLongAs, ConditionPredicateControllerControls, false},
		{"negated static Selection", "As long as you control two or fewer other lands, this creature has flying.", ConditionAsLongAs, ConditionPredicateControllerControls, true},
		{"replacement Selection count", "This land enters tapped unless you control two or more basic lands.", ConditionUnless, ConditionPredicateControllerControls, true},
		{"existential opponent at least", "As long as an opponent controls two or more creatures, this creature has flying.", ConditionAsLongAs, ConditionPredicateAnyOpponentControls, false},
		{"event subject", "When this creature enters, if it was kicked, draw a card.", ConditionIf, ConditionPredicateEventSubjectWasKicked, false},
		{"activation resource threshold", "{T}: Draw a card. Activate only if you have 10 or more life.", ConditionOnlyIf, ConditionPredicateControllerLifeAtLeast, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			compilation, _ := compileSource(test.source, pipelineContext{CardName: "Test Bear"})
			if len(compilation.Abilities) != 1 || len(compilation.Abilities[0].Content.Conditions) != 1 {
				t.Fatalf("compilation = %#v", compilation)
			}
			condition := compilation.Abilities[0].Content.Conditions[0]
			if condition.Kind != test.kind ||
				condition.Predicate != test.predicate ||
				condition.Negated != test.negated ||
				condition.Span.Start.Offset >= condition.Span.End.Offset ||
				test.source[condition.Span.Start.Offset:condition.Span.End.Offset] != condition.Text {
				t.Fatalf("condition = %#v, references = %#v", condition, compilation.Abilities[0].Content.References)
			}
		})
	}
}

func TestCompileActivationConditionTotalPower(t *testing.T) {
	t.Parallel()
	source := "{1}{G}: Regenerate this creature. Activate only if creatures you control have total power 8 or greater."
	compilation, _ := compileSource(source, pipelineContext{CardName: "Test Bear"})
	if len(compilation.Abilities) != 1 || len(compilation.Abilities[0].Content.Conditions) != 1 {
		t.Fatalf("compilation = %#v", compilation)
	}
	condition := compilation.Abilities[0].Content.Conditions[0]
	if condition.Kind != ConditionOnlyIf || condition.Predicate != ConditionPredicateControllerControls {
		t.Fatalf("condition = %#v, want only-if controller-controls", condition)
	}
	selection := condition.Selection
	if len(selection.RequiredTypes) != 1 || selection.RequiredTypes[0] != types.Creature {
		t.Fatalf("selection = %#v, want single creature type", selection)
	}
	if !selection.MatchTotalPowerAtLeast || selection.TotalPowerAtLeast != 8 {
		t.Fatalf("selection = %#v, want total power 8", selection)
	}
	if selection.MatchPowerAtLeast {
		t.Fatalf("selection = %#v, total-power qualifier must not set per-permanent power", selection)
	}
}

func TestCompileConditionsRejectsNearMissWordingSemantically(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"When this creature enters, if you nearly control an artifact, draw a card.",
		"If a creature dealt damage by this creature this turn would die, exile it instead.",
		"Whenever you gain life, if it's a creature, draw a card.",
		"As long as an opponent controls no creatures, this creature has flying.",
	} {
		compilation, _ := compileSource(source, pipelineContext{CardName: "Test Bear"})
		condition := compilation.Abilities[0].Content.Conditions[0]
		if condition.Predicate != ConditionPredicateUnsupported {
			t.Fatalf("condition = %#v, want unsupported predicate", condition)
		}
		if got := source[condition.Span.Start.Offset:condition.Span.End.Offset]; got != condition.Text {
			t.Fatalf("condition span text = %q, want %q", got, condition.Text)
		}
	}
}

func TestCompileReferencesBindsConservativeAntecedents(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		source   string
		bindings []ReferenceBinding
	}{
		{"trigger event card", "Whenever a creature dies, return it to its owner's hand.", []ReferenceBinding{ReferenceBindingEventCard, ReferenceBindingEventCard}},
		{"zone-change event card", "Whenever an artifact is put into a graveyard from the battlefield, return it to its owner's hand.", []ReferenceBinding{ReferenceBindingEventCard, ReferenceBindingEventCard}},
		{"batched event subject is ambiguous", "Whenever one or more creatures die, return it to its owner's hand.", []ReferenceBinding{ReferenceBindingAmbiguous, ReferenceBindingAmbiguous}},
		{"explicit source in trigger body", "Whenever a creature dies, this creature deals 1 damage to its controller.", []ReferenceBinding{ReferenceBindingSource, ReferenceBindingSource}},
		{"single target occurrence", "Return target creature to its owner's hand.", []ReferenceBinding{ReferenceBindingTarget}},
		{"optional single target object pronoun", "Tap up to one target creature and put a stun counter on it.", []ReferenceBinding{ReferenceBindingTarget}},
		{"optional single target possessive pronoun", "Destroy up to one target nonland permanent. Its controller draws a card.", []ReferenceBinding{ReferenceBindingAmbiguous}},
		{"plural target object pronoun", "Tap up to two target creatures and put a stun counter on each of them.", []ReferenceBinding{ReferenceBindingAmbiguous}},
		{"prior instruction result", "Exile target creature. Return it to the battlefield under its owner's control at the beginning of the next end step.", []ReferenceBinding{ReferenceBindingPriorInstructionResult, ReferenceBindingPriorInstructionResult}},
		{"reanimated card result", "Put target creature card from a graveyard onto the battlefield under your control. You lose life equal to that card's mana value.", []ReferenceBinding{ReferenceBindingPriorInstructionResult}},
		{"delayed source", "When this creature enters, exile it at the beginning of the next end step.", []ReferenceBinding{ReferenceBindingSource, ReferenceBindingSource}},
		{"delayed non-self event card", "When enchanted creature dies, return that card to the battlefield under its owner's control at the beginning of the next end step.", []ReferenceBinding{ReferenceBindingEventCard, ReferenceBindingEventCard}},
		{"activation cost source", "Remove a counter from it: Draw a card.", []ReferenceBinding{ReferenceBindingSource}},
		{"activation cost prior object", "Tap an untapped creature you control, Remove a +1/+1 counter from it: Draw a card.", []ReferenceBinding{ReferenceBindingAmbiguous}},
		{"activation cost prior source and object", "Remove a charge counter from this artifact, Tap an untapped creature you control, Remove a +1/+1 counter from it: Draw a card.", []ReferenceBinding{ReferenceBindingSource, ReferenceBindingAmbiguous}},
		{"ambiguous pronoun", "It explores.", []ReferenceBinding{ReferenceBindingAmbiguous}},
		{"they in draw trigger", "Whenever an opponent draws a card, they lose 1 life.", []ReferenceBinding{ReferenceBindingEventPlayer}},
		{"they in discard trigger", "Whenever a player discards a card, they lose 2 life.", []ReferenceBinding{ReferenceBindingEventPlayer}},
		{"their in life trigger", "Whenever an opponent gains life, draw a card.", []ReferenceBinding(nil)},
		{"they in life trigger", "Whenever an opponent gains life, they draw a card.", []ReferenceBinding{ReferenceBindingEventPlayer}},
		{"they in scry trigger", "Whenever a player scries, they draw a card.", []ReferenceBinding{ReferenceBindingEventPlayer}},
		{"they in non-player trigger binds permanent", "Whenever a creature attacks, they deal 1 damage to any target.", []ReferenceBinding{ReferenceBindingEventPermanent}},
		{"that player in combat damage trigger", "Whenever this creature deals combat damage to a player, that player discards a card.", []ReferenceBinding{ReferenceBindingSource, ReferenceBindingEventPlayer}},
		{"they in combat damage to player trigger", "Whenever this creature deals combat damage to a player, they draw a card.", []ReferenceBinding{ReferenceBindingSource, ReferenceBindingEventPlayer}},
		{"their in combat damage to player trigger", "Whenever this creature deals combat damage to a player, exile a card from their hand.", []ReferenceBinding{ReferenceBindingSource, ReferenceBindingEventPlayer}},
		{"their possessive keeps target antecedent", "Whenever you cast a spell, target opponent exiles a card from their hand.", []ReferenceBinding{ReferenceBindingTarget}},
		{"died-creature amount binds event permanent", "Whenever a creature you control dies, put X +1/+1 counters on target creature you control, where X is the power of the creature that died.", []ReferenceBinding{ReferenceBindingEventPermanent}},
		{"its power in leaves trigger with player target binds event permanent", "When this creature leaves the battlefield, target opponent loses life equal to its power.", []ReferenceBinding{ReferenceBindingSource, ReferenceBindingEventPermanent}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			compilation, _ := compileSource(test.source, pipelineContext{CardName: "Test Bear", InstantOrSorcery: true})
			references := compilation.Abilities[0].Content.References
			if len(references) != len(test.bindings) {
				t.Fatalf("references = %#v, want bindings %v", references, test.bindings)
			}
			for i, reference := range references {
				if reference.Binding != test.bindings[i] {
					t.Fatalf("reference[%d] = %#v, want binding %v", i, reference, test.bindings[i])
				}
				if got := test.source[reference.Span.Start.Offset:reference.Span.End.Offset]; got != reference.Text {
					t.Fatalf("reference[%d] span text = %q, want %q", i, got, reference.Text)
				}
			}
		})
	}
}

func hasDiagnosticForSpan(diagnostics []shared.Diagnostic, span shared.Span) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Span == span {
			return true
		}
	}
	return false
}

func TestCompileEventHistoryInterveningConditions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		source    string
		predicate ConditionPredicate
		event     TriggerEvent
		window    ConditionEventHistoryWindow
		negated   bool
	}{
		{
			name:      "you attacked this turn",
			source:    "When this creature enters, if you attacked this turn, draw a card.",
			predicate: ConditionPredicateEventHistory,
			event:     TriggerEventAttackerDeclared,
			window:    ConditionEventHistoryWindowCurrentTurn,
		},
		{
			name:      "a creature died this turn",
			source:    "At the beginning of your end step, if a creature died this turn, draw a card.",
			predicate: ConditionPredicateEventHistory,
			event:     TriggerEventPermanentDied,
			window:    ConditionEventHistoryWindowCurrentTurn,
		},
		{
			name:      "you gained life this turn",
			source:    "At the beginning of each end step, if you gained life this turn, draw a card.",
			predicate: ConditionPredicateEventHistory,
			event:     TriggerEventLifeGained,
			window:    ConditionEventHistoryWindowCurrentTurn,
		},
		{
			name:      "an opponent lost life this turn",
			source:    "At the beginning of your end step, if an opponent lost life this turn, draw a card.",
			predicate: ConditionPredicateEventHistory,
			event:     TriggerEventLifeLost,
			window:    ConditionEventHistoryWindowCurrentTurn,
		},
		{
			name:      "you lost life this turn",
			source:    "At the beginning of your end step, if you lost life this turn, draw a card.",
			predicate: ConditionPredicateEventHistory,
			event:     TriggerEventLifeLost,
			window:    ConditionEventHistoryWindowCurrentTurn,
		},
		{
			name:      "an opponent lost life last turn",
			source:    "At the beginning of each upkeep, if an opponent lost life last turn, draw a card.",
			predicate: ConditionPredicateEventHistory,
			event:     TriggerEventLifeLost,
			window:    ConditionEventHistoryWindowPreviousTurn,
		},
		{
			name:      "you lost life last turn",
			source:    "At the beginning of each upkeep, if you lost life last turn, draw a card.",
			predicate: ConditionPredicateEventHistory,
			event:     TriggerEventLifeLost,
			window:    ConditionEventHistoryWindowPreviousTurn,
		},
		{
			name:      "no spells were cast last turn",
			source:    "At the beginning of your upkeep, if no spells were cast last turn, draw a card.",
			predicate: ConditionPredicateEventHistory,
			event:     TriggerEventSpellCast,
			window:    ConditionEventHistoryWindowPreviousTurn,
			negated:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(test.source, pipelineContext{CardName: "Test Bear"})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if len(compilation.Abilities) != 1 {
				t.Fatalf("abilities = %d, want 1", len(compilation.Abilities))
			}
			trigger := compilation.Abilities[0].Trigger
			if trigger == nil || trigger.Condition == nil {
				t.Fatal("trigger condition = nil")
			}
			cond := trigger.Condition
			if !cond.Intervening {
				t.Error("Intervening = false, want true")
			}
			if cond.Kind != ConditionIf {
				t.Errorf("Kind = %v, want ConditionIf", cond.Kind)
			}
			if cond.Predicate != test.predicate {
				t.Errorf("Predicate = %v, want %v", cond.Predicate, test.predicate)
			}
			if cond.EventHistoryPattern == nil {
				t.Fatal("EventHistoryPattern = nil, want non-nil")
			}
			if cond.EventHistoryPattern.Event != test.event {
				t.Errorf("EventHistoryPattern.Event = %v, want %v", cond.EventHistoryPattern.Event, test.event)
			}
			if cond.EventHistoryWindow != test.window {
				t.Errorf("EventHistoryWindow = %v, want %v", cond.EventHistoryWindow, test.window)
			}
			if cond.Negated != test.negated {
				t.Errorf("Negated = %v, want %v", cond.Negated, test.negated)
			}
			if cond.Span.Start.Offset >= cond.Span.End.Offset {
				t.Errorf("Span = %v, want non-empty", cond.Span)
			}
			if got := test.source[cond.Span.Start.Offset:cond.Span.End.Offset]; got != cond.Text {
				t.Errorf("span text = %q, want %q", got, cond.Text)
			}
		})
	}
}

func TestCompileProvenObjectAndControllerInterveningConditions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		condition     string
		predicate     ConditionPredicate
		binding       ReferenceBinding
		threshold     int
		negated       bool
		requiredTypes []types.Card
		subtypes      []string
		tapped        ConditionTriState
		power         int
		excludeSource bool
	}{
		{"event creature", "if it was a creature", ConditionPredicateObjectMatches, ReferenceBindingEventPermanent, 0, false, []types.Card{types.Creature}, nil, ConditionTriAny, 0, false},
		{"event creature contraction", "IF IT'S A CREATURE", ConditionPredicateObjectMatches, ReferenceBindingEventPermanent, 0, false, []types.Card{types.Creature}, nil, ConditionTriAny, 0, false},
		{"event Human", "if it was a Human", ConditionPredicateObjectMatches, ReferenceBindingEventPermanent, 0, false, nil, []string{"Human"}, ConditionTriAny, 0, false},
		{"event counters", "if it had counters on it", ConditionPredicateEventSubjectHadCounters, ReferenceBindingEventPermanent, 0, false, nil, nil, ConditionTriAny, 0, false},
		{"event name unique", "if it doesn't have the same name as another creature you control or a creature card in your graveyard", ConditionPredicateEventSubjectNameUnique, ReferenceBindingEventPermanent, 0, false, nil, nil, ConditionTriAny, 0, false},
		{"untapped artifact source", "if this artifact is untapped", ConditionPredicateObjectMatches, ReferenceBindingSource, 0, false, []types.Card{types.Artifact}, nil, ConditionTriFalse, 0, false},
		{"untapped creature source", "if this creature is untapped", ConditionPredicateObjectMatches, ReferenceBindingSource, 0, false, []types.Card{types.Creature}, nil, ConditionTriFalse, 0, false},
		{"enchantment source", "if this permanent is an enchantment", ConditionPredicateObjectMatches, ReferenceBindingSource, 0, false, []types.Card{types.Enchantment}, nil, ConditionTriAny, 0, false},
		{"source exists", "if this creature is on the battlefield", ConditionPredicateObjectExists, ReferenceBindingSource, 0, false, nil, nil, ConditionTriAny, 0, false},
		{"two Gates", "if you control two or more Gates", ConditionPredicateControllerControls, ReferenceBindingUnsupported, 2, false, nil, []string{"Gate"}, ConditionTriAny, 0, false},
		{"two tapped creatures", "if you control two or more tapped creatures", ConditionPredicateControllerControls, ReferenceBindingUnsupported, 2, false, []types.Card{types.Creature}, nil, ConditionTriTrue, 0, false},
		{"power five creature", "if you control a creature with power 5 or greater", ConditionPredicateControllerControls, ReferenceBindingUnsupported, 0, false, []types.Card{types.Creature}, nil, ConditionTriAny, 5, false},
		{"another power four creature", "if you control another creature with power 4 or greater", ConditionPredicateControllerControls, ReferenceBindingUnsupported, 0, false, []types.Card{types.Creature}, nil, ConditionTriAny, 4, true},
		{"Equipment", "if you control an Equipment", ConditionPredicateControllerControls, ReferenceBindingUnsupported, 0, false, nil, []string{"Equipment"}, ConditionTriAny, 0, false},
		{"no creatures", "if you control no creatures", ConditionPredicateControllerControls, ReferenceBindingUnsupported, 1, true, []types.Card{types.Creature}, nil, ConditionTriAny, 0, false},
		{"three creatures", "if you control three or more creatures", ConditionPredicateControllerControls, ReferenceBindingUnsupported, 3, false, []types.Card{types.Creature}, nil, ConditionTriAny, 0, false},
		{"tapped creature", "if you control a tapped creature", ConditionPredicateControllerControls, ReferenceBindingUnsupported, 0, false, []types.Card{types.Creature}, nil, ConditionTriTrue, 0, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			source := "Whenever a creature dies, " + test.condition + ", draw a card."
			compilation, _ := compileSource(source, pipelineContext{CardName: "Test Relic"})
			if len(compilation.Abilities) != 1 || len(compilation.Abilities[0].Content.Conditions) != 1 {
				t.Fatalf("compilation = %#v", compilation)
			}
			condition := compilation.Abilities[0].Content.Conditions[0]
			if condition.Predicate != test.predicate ||
				condition.ObjectBinding != test.binding ||
				condition.Threshold != test.threshold ||
				condition.Negated != test.negated ||
				condition.Selection.Tapped != test.tapped ||
				condition.Selection.PowerAtLeast != test.power ||
				condition.Selection.ExcludeSource != test.excludeSource ||
				!slices.Equal(condition.Selection.RequiredTypes, test.requiredTypes) ||
				!slices.Equal(condition.Selection.SubtypesAny, test.subtypes) {
				t.Fatalf("condition = %#v, references = %#v", condition, compilation.Abilities[0].Content.References)
			}
			if test.power > 0 && !condition.Selection.MatchPowerAtLeast {
				t.Fatalf("condition = %#v, want power-at-least match", condition)
			}
		})
	}
}

func TestCompileEventHistoryCreatureDiedHasCreatureSelection(t *testing.T) {
	t.Parallel()
	compilation, _ := compileSource("At the beginning of your end step, if a creature died this turn, draw a card.", pipelineContext{CardName: "Test Bear"})
	if len(compilation.Abilities) != 1 {
		t.Fatalf("abilities = %d, want 1", len(compilation.Abilities))
	}
	cond := compilation.Abilities[0].Trigger.Condition
	if cond == nil || cond.Predicate != ConditionPredicateEventHistory {
		t.Fatalf("condition = %#v", cond)
	}
	if cond.EventHistoryPattern == nil {
		t.Fatal("EventHistoryPattern = nil, want non-nil")
	}
	sel := cond.EventHistoryPattern.SubjectSelection
	if len(sel.RequiredTypes) != 1 || sel.RequiredTypes[0] != types.Creature {
		t.Fatalf("SubjectSelection = %#v, want creature", sel)
	}
}

func TestCompileEventHistoryAttackedHasControllerYou(t *testing.T) {
	t.Parallel()
	compilation, _ := compileSource("When this creature enters, if you attacked this turn, draw a card.", pipelineContext{CardName: "Test Bear"})
	if len(compilation.Abilities) != 1 {
		t.Fatalf("abilities = %d, want 1", len(compilation.Abilities))
	}
	cond := compilation.Abilities[0].Trigger.Condition
	if cond == nil || cond.Predicate != ConditionPredicateEventHistory {
		t.Fatalf("condition = %#v", cond)
	}
	if cond.EventHistoryPattern == nil {
		t.Fatal("EventHistoryPattern = nil, want non-nil")
	}
	if cond.EventHistoryPattern.Controller != ControllerYou {
		t.Fatalf("Controller = %v, want ControllerYou", cond.EventHistoryPattern.Controller)
	}
}

func TestCompileConstructedEventHistoryConditionIsTextBlind(t *testing.T) {
	t.Parallel()
	span := shared.Span{
		Start: shared.Position{Offset: 1, Line: 1, Column: 2},
		End:   shared.Position{Offset: 26, Line: 1, Column: 27},
	}
	syntax := []parser.EventHistoryCondition{
		{
			Span:    span,
			Negated: true,
			Window:  parser.EventHistoryWindow{Kind: parser.EventHistoryWindowPreviousTurn},
			TriggerEvent: &parser.TriggerEventClause{
				Kind:  parser.TriggerEventKindSpellCast,
				Actor: parser.TriggerEventActor{Kind: parser.TriggerEventActorPlayer},
			},
		},
	}
	condition := CompiledCondition{
		Kind:              ConditionIf,
		Span:              span,
		Text:              "if you attacked this turn",
		ClauseIndex:       -1,
		EventHistoryIndex: 0,
	}
	recognizeCondition(&condition, nil, syntax)
	if condition.Predicate != ConditionPredicateEventHistory ||
		condition.EventHistoryPattern == nil ||
		condition.EventHistoryPattern.Event != TriggerEventSpellCast {
		t.Fatalf("condition = %#v, want typed spell-cast history", condition)
	}

	condition = CompiledCondition{
		Kind:              ConditionIf,
		Span:              span,
		Text:              "if you attacked this turn",
		ClauseIndex:       -1,
		EventHistoryIndex: -1,
	}
	recognizeCondition(&condition, nil, nil)
	if condition.Predicate != ConditionPredicateUnsupported {
		t.Fatalf("condition = %#v, want text-blind unsupported predicate", condition)
	}
}

func TestCompileConstructedConditionClauseIsTextBlind(t *testing.T) {
	t.Parallel()
	span := shared.Span{
		Start: shared.Position{Offset: 1, Line: 1, Column: 2},
		End:   shared.Position{Offset: 26, Line: 1, Column: 27},
	}
	tests := []struct {
		name      string
		kind      ConditionKind
		clause    parser.ConditionClause
		predicate ConditionPredicate
		threshold int
		negated   bool
	}{
		{
			name: "controls at least maps threshold",
			kind: ConditionIf,
			clause: parser.ConditionClause{
				Span:         span,
				Predicate:    parser.ConditionPredicateControls,
				Scope:        parser.ConditionControlScopeController,
				Comparison:   parser.ConditionComparisonAtLeast,
				CompareValue: 3,
				Selection:    parser.ConditionSelection{RequiredTypes: []parser.TriggerCardType{parser.TriggerCardTypeCreature}},
			},
			predicate: ConditionPredicateControllerControls,
			threshold: 3,
		},
		{
			name: "controls at most inverts negation",
			kind: ConditionIf,
			clause: parser.ConditionClause{
				Span:         span,
				Predicate:    parser.ConditionPredicateControls,
				Scope:        parser.ConditionControlScopeController,
				Comparison:   parser.ConditionComparisonAtMost,
				CompareValue: 2,
				Selection:    parser.ConditionSelection{RequiredTypes: []parser.TriggerCardType{parser.TriggerCardTypeCreature}},
			},
			predicate: ConditionPredicateControllerControls,
			threshold: 3,
			negated:   true,
		},
		{
			name: "unless introducer negates base predicate",
			kind: ConditionUnless,
			clause: parser.ConditionClause{
				Span:      span,
				Predicate: parser.ConditionPredicateControllerLifeAtLeast,
				Threshold: 5,
			},
			predicate: ConditionPredicateControllerLifeAtLeast,
			threshold: 5,
			negated:   true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			condition := CompiledCondition{
				Kind: test.kind,
				Span: span,
				// Deliberately contradictory text: the compiler must derive
				// meaning from the typed clause, never from this source text.
				Text:              "if the sky is green",
				ClauseIndex:       0,
				EventHistoryIndex: -1,
			}
			recognizeCondition(&condition, []parser.ConditionClause{test.clause}, nil)
			if condition.Predicate != test.predicate ||
				condition.Threshold != test.threshold ||
				condition.Negated != test.negated {
				t.Fatalf("condition = %#v, want predicate %v threshold %d negated %v", condition, test.predicate, test.threshold, test.negated)
			}
		})
	}
}

func TestCompileConstructedConditionClauseFailsClosed(t *testing.T) {
	t.Parallel()
	span := shared.Span{
		Start: shared.Position{Offset: 1, Line: 1, Column: 2},
		End:   shared.Position{Offset: 26, Line: 1, Column: 27},
	}
	tests := []struct {
		name   string
		clause parser.ConditionClause
	}{
		{
			name:   "unknown predicate",
			clause: parser.ConditionClause{Span: span, Predicate: parser.ConditionPredicateUnknown},
		},
		{
			name: "selection card type outside closed vocabulary",
			clause: parser.ConditionClause{
				Span:       span,
				Predicate:  parser.ConditionPredicateControls,
				Scope:      parser.ConditionControlScopeController,
				Comparison: parser.ConditionComparisonNone,
				Selection:  parser.ConditionSelection{RequiredTypes: []parser.TriggerCardType{parser.TriggerCardTypeInstant}},
			},
		},
		{
			name: "counter predicate with no counter kind",
			clause: parser.ConditionClause{
				Span:      span,
				Predicate: parser.ConditionPredicateEventSubjectHadNoCounter,
				Counter:   parser.ConditionCounterNone,
			},
		},
		{
			name: "damage source with unknown color",
			clause: parser.ConditionClause{
				Span:      span,
				Predicate: parser.ConditionPredicateDamageByControlledSource,
				Selection: parser.ConditionSelection{ColorsAny: []parser.TriggerColor{parser.TriggerColorUnknown}},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			condition := CompiledCondition{Kind: ConditionIf, Span: span, Text: "if something", ClauseIndex: 0, EventHistoryIndex: -1}
			recognizeCondition(&condition, []parser.ConditionClause{test.clause}, nil)
			if condition.Predicate != ConditionPredicateUnsupported {
				t.Fatalf("condition = %#v, want unsupported predicate", condition)
			}
		})
	}
}
