package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestCompileActivatedAbility(t *testing.T) {
	t.Parallel()
	source := "{1}{G}, {T}: Target attacking creature you control gets +2/+2 until end of turn."
	compilation, diagnostics := compileSource(source, pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	ability := compilation.Abilities[0]
	if ability.Cost == nil || len(ability.Cost.Components) != 2 {
		t.Fatalf("cost = %#v", ability.Cost)
	}

	if ability.Cost.Components[0].Kind != CostMana ||
		ability.Cost.Components[0].Symbol != "{1}{G}" ||
		ability.Cost.Components[1].Kind != CostTap {
		t.Fatalf("cost components = %#v", ability.Cost.Components)
	}
	if len(ability.Content.Targets) != 1 {
		t.Fatalf("targets = %#v", ability.Content.Targets)
	}
	target := ability.Content.Targets[0]
	if target.Selector.Kind != SelectorCreature ||
		target.Selector.Controller != ControllerYou ||
		!target.Selector.Attacking {
		t.Fatalf("target selector = %#v", target.Selector)
	}
	if len(ability.Content.Effects) != 1 ||
		ability.Content.Effects[0].Kind != EffectModifyPT ||
		ability.Content.Effects[0].Duration != DurationUntilEndOfTurn {
		t.Fatalf("effects = %#v", ability.Content.Effects)
	}
	if ability.Content.Effects[0].PowerDelta != (CompiledSignedAmount{Value: 2, Known: true}) ||
		ability.Content.Effects[0].ToughnessDelta != (CompiledSignedAmount{Value: 2, Known: true}) {
		t.Fatalf("power/toughness change = %#v", ability.Content.Effects[0])
	}
}

func TestCompileActivatedCostPayLifeCommanderColorIdentity(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"{3}, {T}, Pay life equal to the number of colors in your commanders' color identity: Draw a card.",
		pipelineContext{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Cost == nil || len(ability.Cost.Components) != 3 {
		t.Fatalf("cost = %#v", ability.Cost)
	}
	component := ability.Cost.Components[2]
	if component.Kind != CostPayLife || component.AmountKnown || component.AmountFromX ||
		component.PayLifeAmountDynamic != DynamicAmountCommanderColorCount {
		t.Fatalf("component = %#v", component)
	}
}

func TestCompileDiscardSelfActivationFunctionsFromHand(t *testing.T) {
	t.Parallel()

	compilation, diagnostics := compileSource(
		"Channel — {1}{G}, Discard this card: Destroy target artifact.",
		pipelineContext{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.ActivationZone != zone.Hand {
		t.Fatalf("activation zone = %v, want hand", ability.ActivationZone)
	}
}

func TestCompileSpellAdditionalPayXLifeCost(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"As an additional cost to cast this spell, pay X life.\nAll creatures get -X/-X until end of turn.",
		pipelineContext{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Kind != AbilitySpellAdditionalCost || ability.Cost == nil ||
		len(ability.Cost.Components) != 1 {
		t.Fatalf("ability = %#v", ability)
	}
	component := ability.Cost.Components[0]
	if component.Kind != CostPayLife || !component.AmountFromX || component.AmountKnown {
		t.Fatalf("component = %#v", component)
	}
	effect := compilation.Abilities[1].Content.Effects[0]
	if effect.StaticSubject != StaticSubjectAllCreatures ||
		!effect.PowerDelta.VariableX || !effect.PowerDelta.Negative ||
		!effect.ToughnessDelta.VariableX || !effect.ToughnessDelta.Negative ||
		effect.Duration != DurationUntilEndOfTurn ||
		!effect.Exact {
		t.Fatalf("effect = %#v", effect)
	}
}

func TestCompileAbilityContentSpan(t *testing.T) {
	t.Parallel()
	source := "Draw a card."
	compilation, diagnostics := compileSource(source, pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	span := ability.Content.Span
	if span.Start.Offset < 0 {
		t.Fatalf("content span start = %d, want >= 0", span.Start.Offset)
	}
	if span.End.Offset <= span.Start.Offset {
		t.Fatalf("content span = %#v, want End.Offset > Start.Offset", span)
	}
	if len(ability.Content.Effects) != 1 {
		t.Fatalf("effects = %#v, want one effect", ability.Content.Effects)
	}
	effect := ability.Content.Effects[0]
	if span.Start.Offset > effect.Span.Start.Offset || span.End.Offset < effect.Span.End.Offset {
		t.Fatalf("content span %#v does not cover effect span %#v", span, effect.Span)
	}
}

// TestCompileAbilityContentSpanBodyRange proves that Content.Span is taken from
// the body token range, not just the union of recognized elements, so that:
//   - Unrecognized/unsupported bodies still have a non-zero Content.Span.
//   - Activated-ability Content.Span excludes the cost (everything before the
//     colon) and therefore starts at the body, not at offset 0.
func TestCompileAbilityContentSpanBodyRange(t *testing.T) {
	t.Parallel()
	t.Run("unsupported_body_nonzero_span", func(t *testing.T) {
		t.Parallel()
		// An ability text the compiler cannot recognise into any element still
		// has a body; Content.Span must cover that body.
		source := "Frob the gronk."
		compilation, _ := compileSource(source, pipelineContext{})
		if len(compilation.Abilities) == 0 {
			t.Fatal("expected at least one ability")
		}
		span := compilation.Abilities[0].Content.Span
		if span.Start.Offset < 0 || span.End.Offset <= span.Start.Offset {
			t.Fatalf("expected non-zero Content.Span for unrecognized body, got %#v", span)
		}
		if got := source[span.Start.Offset:span.End.Offset]; got == "" {
			t.Fatal("Content.Span maps to empty source slice")
		}
	})
	t.Run("activated_span_excludes_cost", func(t *testing.T) {
		t.Parallel()
		// For an activated ability the cost is everything up to and including
		// the colon.  Content.Span must start at the body (after the colon),
		// not at offset 0 where the cost begins.
		source := "{T}: Draw a card."
		compilation, diagnostics := compileSource(source, pipelineContext{})
		if len(diagnostics) != 0 {
			t.Fatalf("diagnostics = %#v", diagnostics)
		}
		ability := compilation.Abilities[0]
		if ability.Cost == nil {
			t.Fatal("expected a cost")
		}
		costEnd := ability.Cost.Span.End.Offset
		contentStart := ability.Content.Span.Start.Offset
		if contentStart <= costEnd {
			t.Fatalf("Content.Span.Start (%d) is not after cost end (%d); content span = %#v",
				contentStart, costEnd, ability.Content.Span)
		}
	})
}

func TestCompileActivatedAbilityTiming(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		text string
		want ActivationTimingKind
	}{
		{"sorcery", "{1}: Draw a card. Activate only as a sorcery.", ActivationTimingSorcery},
		{"once per turn", "{1}: Draw a card. Activate only once each turn.", ActivationTimingOncePerTurn},
		{"combat", "{1}: Draw a card. Activate only during combat.", ActivationTimingDuringCombat},
		{"upkeep", "{1}: Draw a card. Activate only during your upkeep.", ActivationTimingDuringUpkeep},
		{"during your turn", "{1}: Draw a card. Activate only during your turn.", ActivationTimingDuringYourTurn},
		{
			"opponent turn unsupported",
			"{1}: Draw a card. Activate only during an opponent's turn.",
			ActivationTimingUnsupported,
		},
		{"once per turn before reminder", "{1}: Draw a card. Activate only once each turn. (This is reminder text.)", ActivationTimingOncePerTurn},
		{"once per turn after reminder", "{1}: Draw a card. (This is reminder text.) Activate only once each turn.", ActivationTimingOncePerTurn},
		{
			"sorcery once per turn",
			"{1}: Draw a card. Activate only as a sorcery. Activate only once each turn.",
			ActivationTimingSorceryOncePerTurn,
		},
		{
			"sorcery once per turn conjoined",
			"{1}: Draw a card. Activate only as a sorcery and only once each turn.",
			ActivationTimingSorceryOncePerTurn,
		},
		{
			"player turn once per turn conjoined unsupported",
			"{1}: Draw a card. Activate only during your turn and only once each turn.",
			ActivationTimingUnsupported,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(test.text, pipelineContext{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			ability := compilation.Abilities[0]
			if ability.ActivationTiming != test.want {
				t.Fatalf("activation timing = %v, want %v", ability.ActivationTiming, test.want)
			}
			if got := test.text[ability.ActivationTimingSpan.Start.Offset:ability.ActivationTimingSpan.End.Offset]; got == "" {
				t.Fatal("activation timing span is empty")
			}
			if len(ability.Content.Effects) != 1 || ability.Content.Effects[0].Kind != EffectDraw {
				t.Fatalf("effects = %#v, want one draw effect", ability.Content.Effects)
			}
			if len(ability.Content.References) != 0 {
				t.Fatalf("references = %#v, want timing references excluded", ability.Content.References)
			}
		})
	}
}

func TestCompileUnsupportedActivationTiming(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"{1}: Draw a card. Activate only during your end step.",
		"{1}: Draw a card. Activate only before combat.",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(text, pipelineContext{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			ability := compilation.Abilities[0]
			if ability.ActivationTiming != ActivationTimingUnsupported {
				t.Fatalf("activation timing = %v, want unsupported", ability.ActivationTiming)
			}
			if len(ability.Content.Effects) != 1 || ability.Content.Effects[0].Kind != EffectDraw {
				t.Fatalf("effects = %#v, want timing restriction excluded from content", ability.Content.Effects)
			}
		})
	}
}

func TestCompileConstructedActivationRestrictions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		restrictions []parser.ActivationRestriction
		want         ActivationTimingKind
	}{
		{
			name: "sorcery timing",
			restrictions: []parser.ActivationRestriction{{
				Kind: parser.ActivationRestrictionSorceryTiming,
			}},
			want: ActivationTimingSorcery,
		},
		{
			name: "once per turn",
			restrictions: []parser.ActivationRestriction{{
				Kind: parser.ActivationRestrictionFrequency,
				Frequency: parser.ActivationFrequencyRestriction{
					Count:  parser.ActivationFrequencyCount{Kind: parser.ActivationFrequencyCountOnce},
					Period: parser.ActivationFrequencyPeriod{Kind: parser.ActivationFrequencyPeriodTurn},
				},
			}},
			want: ActivationTimingOncePerTurn,
		},
		{
			name: "combat",
			restrictions: []parser.ActivationRestriction{{
				Kind: parser.ActivationRestrictionPhaseStep,
				PhaseStep: parser.ActivationPhaseStepRestriction{
					Quantifier: parser.PhaseStepQuantifier{Kind: parser.PhaseStepQuantifierEach},
					Player:     parser.TriggerPlayerSelector{Kind: parser.TriggerPlayerSelectorAny},
					Name:       parser.PhaseStepName{Kind: parser.PhaseStepNameCombat},
				},
			}},
			want: ActivationTimingDuringCombat,
		},
		{
			name: "controller upkeep",
			restrictions: []parser.ActivationRestriction{{
				Kind: parser.ActivationRestrictionPhaseStep,
				PhaseStep: parser.ActivationPhaseStepRestriction{
					Quantifier: parser.PhaseStepQuantifier{Kind: parser.PhaseStepQuantifierEachOf},
					Player:     parser.TriggerPlayerSelector{Kind: parser.TriggerPlayerSelectorYou},
					Name:       parser.PhaseStepName{Kind: parser.PhaseStepNameUpkeep},
				},
			}},
			want: ActivationTimingDuringUpkeep,
		},
		{
			name: "composed",
			restrictions: []parser.ActivationRestriction{
				{
					Kind: parser.ActivationRestrictionFrequency,
					Frequency: parser.ActivationFrequencyRestriction{
						Count:  parser.ActivationFrequencyCount{Kind: parser.ActivationFrequencyCountOnce},
						Period: parser.ActivationFrequencyPeriod{Kind: parser.ActivationFrequencyPeriodTurn},
					},
				},
				{Kind: parser.ActivationRestrictionSorceryTiming},
			},
			want: ActivationTimingSorceryOncePerTurn,
		},
		{
			name: "unsupported",
			restrictions: []parser.ActivationRestriction{{
				Kind: parser.ActivationRestrictionUnsupported,
			}},
			want: ActivationTimingUnsupported,
		},
		{
			name: "unsupported typed phase",
			restrictions: []parser.ActivationRestriction{{
				Kind: parser.ActivationRestrictionPhaseStep,
				PhaseStep: parser.ActivationPhaseStepRestriction{
					Quantifier: parser.PhaseStepQuantifier{Kind: parser.PhaseStepQuantifierSingle},
					Player:     parser.TriggerPlayerSelector{Kind: parser.TriggerPlayerSelectorYou},
					Name:       parser.PhaseStepName{Kind: parser.PhaseStepNameEndStep},
				},
			}},
			want: ActivationTimingUnsupported,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			for i := range test.restrictions {
				test.restrictions[i].Span = shared.Span{
					Start: shared.Position{Offset: 10 + i*20},
					End:   shared.Position{Offset: 20 + i*20},
				}
			}
			got, span := compileActivationTiming(AbilityActivated, test.restrictions)
			if got != test.want {
				t.Fatalf("timing = %v, want %v", got, test.want)
			}
			if span.Start.Offset != 10 || span.End.Offset != 20+(len(test.restrictions)-1)*20 {
				t.Fatalf("span = %#v, want span derived from constructed nodes", span)
			}
		})
	}
}

func TestCompileActivatedAbilityZone(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		text string
		want zone.Type
	}{
		{name: "battlefield", text: "{1}: Draw a card.", want: zone.Battlefield},
		{name: "graveyard self return", text: "{1}: Return this card from your graveyard to your hand.", want: zone.Graveyard},
		{name: "graveyard source cost", text: "Exile this card from your graveyard: Draw a card.", want: zone.Graveyard},
		{name: "battlefield returns target", text: "{1}: Return target card from your graveyard to your hand.", want: zone.Battlefield},
		{
			name: "battlefield source reference in another clause",
			text: "{1}: Exile this card, then return target card from your graveyard to your hand.",
			want: zone.Battlefield,
		},
		{
			name: "modal graveyard self return",
			text: "{1}: Choose one —\n• Return this card from your graveyard to your hand.\n• Draw a card.",
			want: zone.Graveyard,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(test.text, pipelineContext{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if got := compilation.Abilities[0].ActivationZone; got != test.want {
				t.Fatalf("activation zone = %v, want %v", got, test.want)
			}
		})
	}
}

func TestCompileActivatedAbilityTapPermanentsCost(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource("Tap two untapped artifacts you control: Draw a card.", pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Cost == nil || len(ability.Cost.Components) != 1 {
		t.Fatalf("cost = %#v", ability.Cost)
	}
	component := ability.Cost.Components[0]
	if component.Kind != CostTapPermanents || component.Object != "two untapped artifacts you control" {
		t.Fatalf("cost component = %#v, want tap-permanents object", component)
	}
}

func TestCompileActivatedAbilityPluralRemoveCounterCost(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource("Remove two storage counters from this land: Add {G}.", pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Cost == nil || len(ability.Cost.Components) != 1 {
		t.Fatalf("cost = %#v", ability.Cost)
	}
	component := ability.Cost.Components[0]
	if component.Kind != CostRemoveCounter || component.Object != "two storage counters from this land" {
		t.Fatalf("cost component = %#v, want remove-counter object", component)
	}
}

func TestCompileActivatedAbilityEnergyCost(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource("Pay {E}{E}: Draw a card.", pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Cost == nil || len(ability.Cost.Components) != 1 {
		t.Fatalf("cost = %#v", ability.Cost)
	}
	component := ability.Cost.Components[0]
	if component.Kind != CostEnergy || component.Amount != "2" {
		t.Fatalf("cost component = %#v, want two-energy cost", component)
	}
}

func TestCompileActivatedAbilityReturnToHandCost(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource("Return two Islands you control to their owner's hand: Draw a card.", pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Cost == nil || len(ability.Cost.Components) != 1 {
		t.Fatalf("cost = %#v", ability.Cost)
	}
	component := ability.Cost.Components[0]
	if component.Kind != CostReturn || component.Object != "two Islands you control to their owner's hand" {
		t.Fatalf("cost component = %#v, want return object", component)
	}
}

func TestCompileActivatedAbilityRevealCost(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource("Reveal X blue cards from your hand, Sacrifice this creature: Draw a card.", pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Cost == nil || len(ability.Cost.Components) != 2 {
		t.Fatalf("cost = %#v", ability.Cost)
	}
	component := ability.Cost.Components[0]
	if component.Kind != CostReveal || component.Object != "X blue cards from your hand" {
		t.Fatalf("cost component = %#v, want reveal object", component)
	}
}

func TestCompileActivatedAbilityIssue210Costs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		text       string
		wantKind   CostKind
		wantObject string
	}{
		{"Exert this creature: Draw a card.", CostExert, "this creature"},
		{"Mill four cards: Draw a card.", CostMill, "four cards"},
		{"Put a verse counter on this creature: Draw a card.", CostPutCounter, "a verse counter on this creature"},
		{"Put two charge counters on this artifact: Draw a card.", CostPutCounter, "two charge counters on this artifact"},
	}
	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(test.text, pipelineContext{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			ability := compilation.Abilities[0]
			if ability.Cost == nil || len(ability.Cost.Components) != 1 {
				t.Fatalf("cost = %#v", ability.Cost)
			}
			component := ability.Cost.Components[0]
			if component.Kind != test.wantKind || component.Object != test.wantObject {
				t.Fatalf("cost component = %#v, want kind %v object %q", component, test.wantKind, test.wantObject)
			}
		})
	}
}

func TestCompileActivatedAbilityCollectEvidenceCost(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource("Collect evidence 4: Draw a card.", pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Cost == nil || len(ability.Cost.Components) != 1 {
		t.Fatalf("cost = %#v", ability.Cost)
	}
	component := ability.Cost.Components[0]
	if component.Kind != CostCollectEvidence || component.Amount != "4" {
		t.Fatalf("cost component = %#v, want collect evidence 4", component)
	}
}

func TestCompileLoyaltyAbilitySignedAmount(t *testing.T) {
	t.Parallel()
	tests := []struct {
		text  string
		value int
		known bool
		fromX bool
	}{
		{text: "+2: Draw a card.", value: 2, known: true},
		{text: "\u22123: Draw a card.", value: -3, known: true},
		{text: "-3: Draw a card.", value: -3, known: true},
		{text: "0: Draw a card.", value: 0, known: true},
		{text: "\u2212X: Draw a card.", fromX: true},
	}
	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(test.text, pipelineContext{Planeswalker: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			ability := compilation.Abilities[0]
			if ability.Cost == nil || len(ability.Cost.Components) != 1 {
				t.Fatalf("cost = %#v", ability.Cost)
			}
			component := ability.Cost.Components[0]
			if component.Kind != CostLoyalty {
				t.Fatalf("cost kind = %v, want loyalty", component.Kind)
			}
			if component.AmountKnown != test.known ||
				(test.known && component.AmountValue != test.value) ||
				component.AmountFromX != test.fromX {
				t.Fatalf("amount = value %d known %v fromX %v, want value %d known %v fromX %v",
					component.AmountValue, component.AmountKnown, component.AmountFromX,
					test.value, test.known, test.fromX)
			}
		})
	}
}

func TestCompileActivatedAbilityCollectEvidenceRejectsMalformedThresholds(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"Collect evidence 0: Draw a card.",
		"Collect evidence two: Draw a card.",
		"Collect evidence X: Draw a card.",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(text, pipelineContext{})
			if len(diagnostics) == 0 {
				t.Fatal("expected unsupported cost diagnostic")
			}
			if compilation.Abilities[0].Cost.Components[0].Kind != CostUnknown {
				t.Fatalf("cost component = %#v, want CostUnknown", compilation.Abilities[0].Cost.Components[0])
			}
		})
	}
}

// TestCompileActivationKeywordSpan proves the compiler copies the parser's typed
// "Activate" keyword span onto an "Activate only if" condition, so lowering can
// account for that consumed source span without inspecting token spelling. A
// plain "only if" without the keyword leaves the span zero.
func TestCompileActivationKeywordSpan(t *testing.T) {
	t.Parallel()

	source := "{T}: Draw a card. Activate only if you have 10 or more life."
	compilation, _ := compileSource(source, pipelineContext{CardName: "Test Bear"})
	if len(compilation.Abilities) != 1 || len(compilation.Abilities[0].Content.Conditions) != 1 {
		t.Fatalf("compilation = %#v", compilation)
	}
	condition := compilation.Abilities[0].Content.Conditions[0]
	span := condition.ActivationKeywordSpan
	if span == (shared.Span{}) {
		t.Fatal("expected a non-zero ActivationKeywordSpan for \"Activate only if\"")
	}
	if got := source[span.Start.Offset:span.End.Offset]; got != "Activate" {
		t.Fatalf("ActivationKeywordSpan text = %q, want %q", got, "Activate")
	}

	// A triggered intervening-if "if" carries no activation keyword span.
	triggered := "When this creature enters, if you control a Mountain, draw a card."
	compiled, _ := compileSource(triggered, pipelineContext{CardName: "Test Bear"})
	if len(compiled.Abilities) != 1 || len(compiled.Abilities[0].Content.Conditions) != 1 {
		t.Fatalf("compilation = %#v", compiled)
	}
	if span := compiled.Abilities[0].Content.Conditions[0].ActivationKeywordSpan; span != (shared.Span{}) {
		t.Fatalf("intervening-if condition has ActivationKeywordSpan = %#v, want zero", span)
	}
}
