package oracle

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestCompileActivatedAbility(t *testing.T) {
	t.Parallel()
	source := "{1}{G}, {T}: Target attacking creature you control gets +2/+2 until end of turn."
	compilation, diagnostics := Compile(source, ParseContext{})
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

func TestCompileAbilityContentSpan(t *testing.T) {
	t.Parallel()
	source := "Draw a card."
	compilation, diagnostics := Compile(source, ParseContext{})
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
		compilation, _ := Compile(source, ParseContext{})
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
		compilation, diagnostics := Compile(source, ParseContext{})
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
		{"once per turn before reminder", "{1}: Draw a card. Activate only once each turn. (This is reminder text.)", ActivationTimingOncePerTurn},
		{
			"sorcery once per turn",
			"{1}: Draw a card. Activate only as a sorcery. Activate only once each turn.",
			ActivationTimingSorceryOncePerTurn,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := Compile(test.text, ParseContext{})
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
	text := "{1}: Draw a card. Activate only during your end step."
	compilation, diagnostics := Compile(text, ParseContext{})
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
			compilation, diagnostics := Compile(test.text, ParseContext{})
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
	compilation, diagnostics := Compile("Tap two untapped artifacts you control: Draw a card.", ParseContext{})
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
	compilation, diagnostics := Compile("Remove two storage counters from this land: Add {G}.", ParseContext{})
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
	compilation, diagnostics := Compile("Pay {E}{E}: Draw a card.", ParseContext{})
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
	compilation, diagnostics := Compile("Return two Islands you control to their owner's hand: Draw a card.", ParseContext{})
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
	compilation, diagnostics := Compile("Reveal X blue cards from your hand, Sacrifice this creature: Draw a card.", ParseContext{})
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
			compilation, diagnostics := Compile(test.text, ParseContext{})
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
	compilation, diagnostics := Compile("Collect evidence 4: Draw a card.", ParseContext{})
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

func TestCompileActivatedAbilityCollectEvidenceRejectsMalformedThresholds(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"Collect evidence 0: Draw a card.",
		"Collect evidence two: Draw a card.",
		"Collect evidence X: Draw a card.",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := Compile(text, ParseContext{})
			if len(diagnostics) == 0 {
				t.Fatal("expected unsupported cost diagnostic")
			}
			if compilation.Abilities[0].Cost.Components[0].Kind != CostUnknown {
				t.Fatalf("cost component = %#v, want CostUnknown", compilation.Abilities[0].Cost.Components[0])
			}
		})
	}
}

func TestCompileTriggeredAbility(t *testing.T) {
	t.Parallel()
	source := "Whenever a creature enters, if it was cast, draw a card."
	compilation, diagnostics := Compile(source, ParseContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	ability := compilation.Abilities[0]
	if ability.Trigger == nil ||
		ability.Trigger.Kind != TriggerWhenever ||
		ability.Trigger.Event != "a creature enters" {
		t.Fatalf("trigger = %#v", ability.Trigger)
	}
	if ability.Trigger.Condition == nil || !ability.Trigger.Condition.Intervening {
		t.Fatalf("intervening condition = %#v", ability.Trigger.Condition)
	}
	if len(ability.Content.Effects) != 1 || ability.Content.Effects[0].Kind != EffectDraw {
		t.Fatalf("effects = %#v", ability.Content.Effects)
	}
}

func TestCompileTriggeredAbilityWithInternalEventComma(t *testing.T) {
	t.Parallel()
	source := "Whenever you cast a noncreature, nonland spell, draw a card."
	compilation, diagnostics := Compile(source, ParseContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	ability := compilation.Abilities[0]
	if ability.Trigger == nil ||
		ability.Trigger.Kind != TriggerWhenever ||
		ability.Trigger.Event != "you cast a noncreature, nonland spell" {
		t.Fatalf("trigger = %#v", ability.Trigger)
	}
	if len(ability.Content.Effects) != 1 || ability.Content.Effects[0].Kind != EffectDraw {
		t.Fatalf("effects = %#v", ability.Content.Effects)
	}
}

func TestCompileSemanticTriggerPatterns(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		source string
		check  func(*testing.T, TriggerPattern)
	}{
		{
			name:   "source self",
			source: "When this creature enters, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Kind != TriggerWhen ||
					pattern.Event != TriggerEventPermanentEnteredBattlefield ||
					pattern.Source != TriggerSourceSelf {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			name:   "source self with capitalized subtype",
			source: "When this Aura enters, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Kind != TriggerWhen ||
					pattern.Event != TriggerEventPermanentEnteredBattlefield ||
					pattern.Source != TriggerSourceSelf {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			name:   "controller and subject Selection",
			source: "Whenever another nontoken creature you control enters, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Controller != ControllerYou ||
					!pattern.ExcludeSelf ||
					!pattern.SubjectSelection.NonToken ||
					!slices.Equal(pattern.SubjectSelection.RequiredTypes, []TriggerCardType{TriggerCardTypeCreature}) {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			name:   "step and active player relation",
			source: "At the beginning of each opponent's draw step, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventBeginningOfStep ||
					pattern.Step != TriggerStepDraw ||
					pattern.Controller != ControllerOpponent {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			name:   "combat damage role",
			source: "Whenever this creature deals combat damage to a creature, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventDamageDealt ||
					pattern.Source != TriggerSourceSelf ||
					pattern.Subject != TriggerSubjectDamageSource ||
					pattern.CombatQualifier != TriggerCombatDamage ||
					pattern.DamageRecipient != TriggerDamageRecipientPermanent ||
					!slices.Equal(pattern.DamageRecipientSelection.RequiredTypes, []TriggerCardType{TriggerCardTypeCreature}) {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			name:   "one or more",
			source: "Whenever one or more artifacts you control enter, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if !pattern.OneOrMore ||
					pattern.Controller != ControllerYou ||
					!slices.Equal(pattern.SubjectSelection.RequiredTypes, []TriggerCardType{TriggerCardTypeArtifact}) {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			name:   "attack subject",
			source: "Whenever a creature an opponent controls attacks, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventAttackerDeclared ||
					pattern.Controller != ControllerOpponent ||
					!slices.Equal(pattern.SubjectSelection.RequiredTypes, []TriggerCardType{TriggerCardTypeCreature}) {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			name:   "attached blocker",
			source: "Whenever equipped creature blocks, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventBlockerDeclared ||
					pattern.Source != TriggerSourceAttachedPermanent {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			name:   "attached permanent dies whenever",
			source: "Whenever equipped creature dies, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Kind != TriggerWhenever ||
					pattern.Event != TriggerEventPermanentDied ||
					pattern.Source != TriggerSourceAttachedPermanent ||
					!slices.Equal(pattern.SubjectSelection.RequiredTypes, []TriggerCardType{TriggerCardTypeCreature}) {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			name:   "tap controller relation",
			source: "Whenever another artifact you control becomes tapped, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventPermanentTapped ||
					pattern.Controller != ControllerYou ||
					!pattern.ExcludeSelf ||
					!slices.Equal(pattern.SubjectSelection.RequiredTypes, []TriggerCardType{TriggerCardTypeArtifact}) {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			name:   "untap subject",
			source: "Whenever a creature you control becomes untapped, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventPermanentUntapped ||
					pattern.Controller != ControllerYou {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			name:   "became target of spell",
			source: "Whenever this creature becomes the target of a spell, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventObjectBecameTarget ||
					pattern.Source != TriggerSourceSelf ||
					pattern.StackObject != TriggerStackObjectSpell {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := Compile(test.source, ParseContext{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			trigger := compilation.Abilities[0].Trigger
			if trigger == nil || trigger.Event == "" {
				t.Fatalf("trigger = %#v", trigger)
			}
			eventText := test.source[trigger.Pattern.Span.Start.Offset:trigger.Pattern.Span.End.Offset]
			if eventText != trigger.Event {
				t.Fatalf("pattern span text = %q, raw diagnostic event = %q", eventText, trigger.Event)
			}
			test.check(t, trigger.Pattern)
		})
	}
}

func TestCompileActionTriggerPatterns(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		check  func(*testing.T, TriggerPattern)
	}{
		{
			source: "Whenever a Forest an opponent controls becomes tapped, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventPermanentTapped ||
					pattern.Controller != ControllerOpponent ||
					!slices.Equal(pattern.SubjectSelection.SubtypesAny, []TriggerSubtype{"forest"}) {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			source: "When this creature is turned face up, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Kind != TriggerWhen ||
					pattern.Event != TriggerEventPermanentTurnedFaceUp ||
					pattern.Source != TriggerSourceSelf {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			source: "Whenever a creature you control becomes the target of a spell or ability an opponent controls, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventObjectBecameTarget ||
					pattern.Controller != ControllerYou ||
					pattern.CauseController != ControllerOpponent ||
					!slices.Equal(pattern.SubjectSelection.RequiredTypes, []TriggerCardType{TriggerCardTypeCreature}) {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			source: "Whenever a player cycles a card, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventCycled || pattern.Player != TriggerPlayerAny {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			source: "Whenever you sacrifice a Clue, you gain 3 life.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventPermanentSacrificed ||
					pattern.Player != TriggerPlayerYou ||
					!slices.Equal(pattern.SubjectSelection.SubtypesAny, []TriggerSubtype{"clue"}) {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			source: "Whenever you scry, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventScry || pattern.Player != TriggerPlayerYou {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			source: "Whenever an opponent activates an ability of a creature or land that isn't a mana ability, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventAbilityActivated ||
					pattern.Player != TriggerPlayerOpponent ||
					!pattern.ExcludeManaAbility ||
					!slices.Equal(pattern.SubjectSelection.RequiredTypesAny, []TriggerCardType{TriggerCardTypeCreature, TriggerCardTypeLand}) {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := Compile(test.source, ParseContext{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			test.check(t, compilation.Abilities[0].Trigger.Pattern)
		})
	}
}

func TestCompileSemanticTriggerPatternsFailClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Whenever this creature attacks alone, draw a card.",
		"Whenever this creature becomes the target of a spell or ability for the first time each turn, draw a card.",
		"Whenever creature you control becomes tapped, draw a card.",
		"At the beginning of your next upkeep, draw a card.",
		"At the beginning of your declare attackers step, draw a card.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := Compile(source, ParseContext{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if pattern := compilation.Abilities[0].Trigger.Pattern; pattern.Event != TriggerEventUnknown {
				t.Fatalf("near-miss pattern = %#v, want unknown event", pattern)
			}
		})
	}
}

func TestCompilePlayerOrdinalTriggerPattern(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := Compile(
		"Whenever you draw your second card each turn, create a 2/2 black Zombie creature token.",
		ParseContext{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	pattern := compilation.Abilities[0].Trigger.Pattern
	if pattern.Event != TriggerEventCardDrawn ||
		pattern.Player != TriggerPlayerYou ||
		pattern.PlayerEventOrdinalThisTurn != 2 {
		t.Fatalf("pattern = %#v", pattern)
	}
}

func TestCompileNamedSelfEnterTriggerPattern(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := Compile(
		"When Example Card enters, draw a card.",
		ParseContext{CardName: "Example Card"},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	pattern := compilation.Abilities[0].Trigger.Pattern
	if pattern.Event != TriggerEventPermanentEnteredBattlefield || pattern.Source != TriggerSourceSelf {
		t.Fatalf("pattern = %#v", pattern)
	}
}

func TestCompileSemanticTriggerPatternReferencesInterveningCondition(t *testing.T) {
	t.Parallel()
	source := "When this creature enters, if it was kicked, draw a card."
	compilation, diagnostics := Compile(source, ParseContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	trigger := compilation.Abilities[0].Trigger
	if trigger.Condition == nil || trigger.Pattern.InterveningCondition != trigger.Condition {
		t.Fatalf("trigger = %#v, want pattern to reference source-spanned intervening condition", trigger)
	}
}

func TestCompileSagaChapterAbility(t *testing.T) {
	t.Parallel()
	source := "II, III — Draw a card."
	compilation, diagnostics := Compile(source, ParseContext{Saga: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Kind != AbilityChapter || !slices.Equal(ability.Chapters, []int{2, 3}) {
		t.Fatalf("ability = %#v", ability)
	}
	if ability.AbilityWord != "" {
		t.Fatalf("ability word = %q, want empty", ability.AbilityWord)
	}
	if len(ability.Content.Effects) != 1 || ability.Content.Effects[0].Kind != EffectDraw {
		t.Fatalf("effects = %#v", ability.Content.Effects)
	}
}

func TestCompileOptionalTriggeredAbility(t *testing.T) {
	t.Parallel()
	source := "When this creature enters, you may draw a card."
	compilation, diagnostics := Compile(source, ParseContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	ability := compilation.Abilities[0]
	if !ability.Optional || source[ability.OptionalSpan.Start.Offset:ability.OptionalSpan.End.Offset] != "you may" {
		t.Fatalf("optional ability = %#v", ability)
	}
}

func TestCompileSelfCannotBlockStaticAbility(t *testing.T) {
	t.Parallel()
	source := "This creature can't block."
	compilation, diagnostics := Compile(source, ParseContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	ability := compilation.Abilities[0]
	if ability.Kind != AbilityStatic ||
		len(ability.Content.Effects) != 1 ||
		ability.Content.Effects[0].Kind != EffectCantBlock ||
		!ability.Content.Effects[0].Negated {
		t.Fatalf("ability = %#v", ability)
	}
	if len(ability.Content.References) != 1 ||
		ability.Content.References[0].Kind != ReferenceThisObject {
		t.Fatalf("references = %#v", ability.Content.References)
	}
}

func TestCompileSelfCannotBeBlockedStaticAbility(t *testing.T) {
	t.Parallel()
	source := "This creature can't be blocked."
	compilation, diagnostics := Compile(source, ParseContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	ability := compilation.Abilities[0]
	if ability.Kind != AbilityStatic ||
		len(ability.Content.Effects) != 1 ||
		ability.Content.Effects[0].Kind != EffectCantBeBlocked ||
		!ability.Content.Effects[0].Negated {
		t.Fatalf("ability = %#v", ability)
	}
	if len(ability.Content.References) != 1 ||
		ability.Content.References[0].Kind != ReferenceThisObject {
		t.Fatalf("references = %#v", ability.Content.References)
	}
}

func TestCompileSelfMustAttackStaticAbility(t *testing.T) {
	t.Parallel()
	source := "This creature attacks each combat if able."
	compilation, diagnostics := Compile(source, ParseContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	ability := compilation.Abilities[0]
	if ability.Kind != AbilityStatic ||
		len(ability.Content.Effects) != 1 ||
		ability.Content.Effects[0].Kind != EffectMustAttack ||
		ability.Content.Effects[0].Negated {
		t.Fatalf("ability = %#v", ability)
	}
	if len(ability.Content.References) != 1 ||
		ability.Content.References[0].Kind != ReferenceThisObject {
		t.Fatalf("references = %#v", ability.Content.References)
	}
	if len(ability.Content.Conditions) != 0 {
		t.Fatalf("intrinsic if-able text became a separate condition: %#v", ability.Content.Conditions)
	}
}

func TestCompileSelfUncounterableStaticAbility(t *testing.T) {
	t.Parallel()
	source := "This spell can't be countered."
	compilation, diagnostics := Compile(source, ParseContext{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	ability := compilation.Abilities[0]
	if ability.Kind != AbilityStatic ||
		len(ability.Content.Effects) != 1 ||
		ability.Content.Effects[0].Kind != EffectCantBeCountered ||
		!ability.Content.Effects[0].Negated {
		t.Fatalf("ability = %#v", ability)
	}
	if len(ability.Content.References) != 1 ||
		ability.Content.References[0].Kind != ReferenceThisObject {
		t.Fatalf("references = %#v", ability.Content.References)
	}
}

func TestCompileReturnToOwnersHand(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := Compile(
		"Return target creature to its owner's hand.",
		ParseContext{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if len(ability.Content.Effects) != 1 || ability.Content.Effects[0].Kind != EffectReturn {
		t.Fatalf("effects = %#v", ability.Content.Effects)
	}
	if len(ability.Content.Targets) != 1 ||
		ability.Content.Targets[0].Selector.Kind != SelectorCreature ||
		ability.Content.Targets[0].Text != "target creature to its owner's hand" {
		t.Fatalf("targets = %#v", ability.Content.Targets)
	}
	if len(ability.Content.References) != 1 ||
		ability.Content.References[0].Kind != ReferencePronoun ||
		ability.Content.References[0].Text != "its" {
		t.Fatalf("references = %#v", ability.Content.References)
	}
	if len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0 ||
		ability.Content.Effects[0].Negated ||
		ability.Content.Targets[0].Cardinality.Min != 1 ||
		ability.Content.Targets[0].Cardinality.Max != 1 {
		t.Fatalf("ability = %#v", ability)
	}
}

func TestCompileGraveyardReturnZones(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		text     string
		fromZone zone.Type
		toZone   zone.Type
	}{
		{
			name:     "target card to hand",
			text:     "Return target instant or sorcery card from your graveyard to your hand.",
			fromZone: zone.Graveyard,
			toZone:   zone.Hand,
		},
		{
			name:     "target card to library",
			text:     "Put target card from your graveyard on the bottom of your library.",
			fromZone: zone.Graveyard,
			toZone:   zone.Library,
		},
		{
			name:     "opponents graveyard",
			text:     "Return target creature card from an opponent's graveyard to your hand.",
			fromZone: zone.Graveyard,
			toZone:   zone.Hand,
		},
		{
			name:     "self to battlefield",
			text:     "Return this card from your graveyard to the battlefield tapped.",
			fromZone: zone.Graveyard,
			toZone:   zone.Battlefield,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := Compile(tc.text, ParseContext{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			ability := compilation.Abilities[0]
			if len(ability.Content.Effects) != 1 {
				t.Fatalf("effects = %#v", ability.Content.Effects)
			}
			effect := ability.Content.Effects[0]
			if effect.FromZone != tc.fromZone || effect.ToZone != tc.toZone {
				t.Fatalf("zones = %v -> %v, want %v -> %v", effect.FromZone, effect.ToZone, tc.fromZone, tc.toZone)
			}
		})
	}
}

func TestCompileResolutionConditionIsNotIntervening(t *testing.T) {
	t.Parallel()
	source := "When this creature dies, draw a card if you control a Forest."
	compilation, diagnostics := Compile(source, ParseContext{})
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

func TestCompileModalAbility(t *testing.T) {
	t.Parallel()
	source := "Choose one —\n• Destroy target creature.\n• Draw two cards."
	compilation, diagnostics := Compile(source, ParseContext{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if len(ability.Content.Modes) != 2 {
		t.Fatalf("modes = %#v", ability.Content.Modes)
	}
	if ability.Content.Modes[0].Content.Effects[0].Kind != EffectDestroy ||
		len(ability.Content.Modes[0].Content.Targets) != 1 ||
		ability.Content.Modes[1].Content.Effects[0].Kind != EffectDraw {
		t.Fatalf("compiled modes = %#v", ability.Content.Modes)
	}
}

func TestCompileKeywordsAndReminder(t *testing.T) {
	t.Parallel()
	source := "First strike (This creature deals combat damage before other creatures.)\nEquip {2} ({2}: Attach to target creature you control.)"
	compilation, diagnostics := Compile(source, ParseContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if got := compilation.Abilities[0].Content.Keywords; len(got) != 1 || got[0].Name != "First strike" {
		t.Fatalf("first strike = %#v", got)
	}
	equip := compilation.Abilities[1]
	if len(equip.Content.Keywords) != 1 || equip.Content.Keywords[0].Name != "Equip" ||
		equip.Content.Keywords[0].Parameter != "{2}" {
		t.Fatalf("equip = %#v", equip.Content.Keywords)
	}
	if len(equip.Content.Effects) != 0 || len(equip.Content.Targets) != 0 {
		t.Fatalf("reminder text leaked semantics: %#v", equip)
	}
}

func TestCompileDevoidAndReminder(t *testing.T) {
	t.Parallel()
	source := "Devoid (This card has no color.)"
	compilation, diagnostics := Compile(source, ParseContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if len(ability.Content.Keywords) != 1 ||
		ability.Content.Keywords[0].Name != "Devoid" ||
		ability.Content.Keywords[0].Text != "Devoid" {
		t.Fatalf("keywords = %#v", ability.Content.Keywords)
	}
	if len(ability.Content.Effects) != 0 || len(ability.Content.References) != 0 {
		t.Fatalf("reminder text leaked semantics: %#v", ability)
	}
}

func TestCompileReadAheadAndReminder(t *testing.T) {
	t.Parallel()
	source := "Read ahead (Choose a chapter and start with that many lore counters. Add one after your draw step. Skipped chapters don't trigger.)"
	compilation, diagnostics := Compile(source, ParseContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(compilation.Abilities) != 1 {
		t.Fatalf("abilities = %#v, want one", compilation.Abilities)
	}
	ability := compilation.Abilities[0]
	if len(ability.Content.Keywords) != 1 ||
		ability.Content.Keywords[0].Name != "Read ahead" ||
		ability.Content.Keywords[0].Text != "Read ahead" {
		t.Fatalf("keywords = %#v", ability.Content.Keywords)
	}
	if len(ability.Content.Effects) != 0 || len(ability.Content.References) != 0 {
		t.Fatalf("reminder text leaked semantics: %#v", ability)
	}
}

func TestCompileEnchantKeywordParameter(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := Compile("Enchant creature", ParseContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	keywords := compilation.Abilities[0].Content.Keywords
	if len(keywords) != 1 {
		t.Fatalf("keywords = %#v", keywords)
	}
	if keywords[0].Name != "Enchant" ||
		keywords[0].Parameter != "creature" ||
		keywords[0].Text != "Enchant creature" ||
		keywords[0].Span.Start.Offset != 0 ||
		keywords[0].Span.End.Offset != len("Enchant creature") {
		t.Fatalf("enchant keyword = %#v", keywords[0])
	}
}

func TestCompileProtectionKeywordParameter(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := Compile("Protection from red", ParseContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	keywords := compilation.Abilities[0].Content.Keywords
	if len(keywords) != 1 {
		t.Fatalf("keywords = %#v", keywords)
	}
	if keywords[0].Name != "Protection" ||
		keywords[0].Parameter != "red" ||
		keywords[0].Text != "Protection from red" ||
		keywords[0].Span.Start.Offset != 0 ||
		keywords[0].Span.End.Offset != len("Protection from red") {
		t.Fatalf("protection keyword = %#v", keywords[0])
	}
}

func TestCompileProtectionKeywordMultipleColors(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := Compile("Protection from black and from red", ParseContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	keywords := compilation.Abilities[0].Content.Keywords
	if len(keywords) != 1 ||
		keywords[0].Parameter != "black,red" ||
		keywords[0].Text != "Protection from black and from red" ||
		keywords[0].Span.End.Offset != len("Protection from black and from red") {
		t.Fatalf("protection keyword = %#v", keywords)
	}
}

func TestCompileTargetsAndReferences(t *testing.T) {
	t.Parallel()
	source := "Legolas deals damage to up to one target creature you don't control. It gains trample until end of turn."
	compilation, diagnostics := Compile(source, ParseContext{
		CardName: "Legolas",
	})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if len(ability.Content.Targets) != 1 ||
		ability.Content.Targets[0].Cardinality != (TargetCardinality{Min: 0, Max: 1}) ||
		ability.Content.Targets[0].Selector.Controller != ControllerNotYou {
		t.Fatalf("targets = %#v", ability.Content.Targets)
	}
	if len(ability.Content.References) != 2 ||
		ability.Content.References[0].Kind != ReferenceSelfName ||
		ability.Content.References[1].Kind != ReferencePronoun {
		t.Fatalf("references = %#v", ability.Content.References)
	}
}

func TestCompileExactTargetCardinalityAndPluralSelector(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := Compile("Tap two target creatures.", ParseContext{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	targets := compilation.Abilities[0].Content.Targets
	if len(targets) != 1 ||
		targets[0].Cardinality != (TargetCardinality{Min: 2, Max: 2}) ||
		targets[0].Selector.Kind != SelectorCreature {
		t.Fatalf("targets = %#v", targets)
	}
}

func TestCompileThirdPersonEffects(t *testing.T) {
	t.Parallel()
	tests := map[string]EffectKind{
		"Each opponent discards a card.":        EffectDiscard,
		"Target player draws two cards.":        EffectDraw,
		"Its controller sacrifices a creature.": EffectSacrifice,
		"That player searches their library.":   EffectSearch,
		"That creature transforms.":             EffectTransform,
	}
	for source, want := range tests {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := Compile(source, ParseContext{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := compilation.Abilities[0].Content.Effects
			if len(effects) != 1 || effects[0].Kind != want {
				t.Fatalf("effects = %#v, want %v", effects, want)
			}
		})
	}
}

func TestCompileFixedEffectValues(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		context ParseContext
		kind    EffectKind
		amount  int
		symbol  string
	}{
		"Draw two cards.": {
			context: ParseContext{InstantOrSorcery: true},
			kind:    EffectDraw,
			amount:  2,
		},
		"Shock deals 3 damage to any target.": {
			context: ParseContext{CardName: "Shock", InstantOrSorcery: true},
			kind:    EffectDealDamage,
			amount:  3,
		},
		"{T}: Add {G}.": {
			kind:   EffectAddMana,
			amount: 1,
			symbol: "{G}",
		},
	}

	for source, test := range tests {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := Compile(source, test.context)
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := compilation.Abilities[0].Content.Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v", effects)
			}
			effect := effects[0]
			if effect.Kind != test.kind ||
				!effect.Amount.Known ||
				effect.Amount.Value != test.amount ||
				effect.Symbol != test.symbol {
				t.Fatalf("effect = %#v", effect)
			}
		})
	}
}

func TestCompileDelayedEffectTiming(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		kind   EffectKind
		timing game.DelayedTriggerTiming
	}{
		{
			source: "Draw a card at the beginning of the next turn's upkeep.",
			kind:   EffectDraw,
			timing: game.DelayedAtBeginningOfNextUpkeep,
		},
		{
			source: "Exile it at the beginning of the next end step.",
			kind:   EffectExile,
			timing: game.DelayedAtBeginningOfNextEndStep,
		},
		{
			source: "Sacrifice it at the beginning of the next end step.",
			kind:   EffectSacrifice,
			timing: game.DelayedAtBeginningOfNextEndStep,
		},
		{
			source: "Return it to its owner's hand at the beginning of the next end step.",
			kind:   EffectReturn,
			timing: game.DelayedAtBeginningOfNextEndStep,
		},
	}
	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := Compile(tt.source, ParseContext{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := compilation.Abilities[0].Content.Effects
			if len(effects) != 1 || effects[0].Kind != tt.kind || effects[0].DelayedTiming != tt.timing {
				t.Fatalf("effects = %#v, want kind %v timing %v", effects, tt.kind, tt.timing)
			}
		})
	}
}

func TestCompileDelayedBlinkEffects(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := Compile(
		"Exile target creature. Return that card to the battlefield under its owner's control at the beginning of the next end step.",
		ParseContext{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := compilation.Abilities[0].Content.Effects
	if len(effects) != 2 ||
		effects[0].Kind != EffectExile ||
		effects[0].DelayedTiming != 0 ||
		effects[1].Kind != EffectReturn ||
		effects[1].DelayedTiming != game.DelayedAtBeginningOfNextEndStep {
		t.Fatalf("effects = %#v, want immediate exile followed by delayed return", effects)
	}
}

func TestCompileDynamicEffectAmounts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source     string
		context    ParseContext
		kind       DynamicAmountKind
		form       DynamicAmountForm
		multiplier int
		selector   SelectorKind
		controller ControllerKind
		text       string
	}{
		{"Swarm deals damage equal to the number of creatures you control to any target.", ParseContext{CardName: "Swarm", InstantOrSorcery: true}, DynamicAmountCount, DynamicAmountEqual, 1, SelectorCreature, ControllerYou, "equal to the number of creatures you control"},
		{"Swarm deals damage equal to twice the number of lands on the battlefield to any target.", ParseContext{CardName: "Swarm", InstantOrSorcery: true}, DynamicAmountCount, DynamicAmountEqual, 2, SelectorLand, ControllerAny, "equal to twice the number of lands on the battlefield"},
		{"You gain 2 life for each opponent you have.", ParseContext{InstantOrSorcery: true}, DynamicAmountOpponentCount, DynamicAmountForEach, 2, SelectorUnknown, ControllerAny, "for each opponent you have"},
		{"You gain life equal to your life total.", ParseContext{InstantOrSorcery: true}, DynamicAmountControllerLife, DynamicAmountEqual, 1, SelectorUnknown, ControllerAny, "equal to your life total"},
		{"You gain X life, where X is your life total.", ParseContext{InstantOrSorcery: true}, DynamicAmountControllerLife, DynamicAmountWhereX, 1, SelectorUnknown, ControllerAny, "where X is your life total"},
		{"When this creature dies, it deals damage equal to its power to any target.", ParseContext{CardName: "Devil"}, DynamicAmountSourcePower, DynamicAmountEqual, 1, SelectorUnknown, ControllerAny, "equal to its power"},
		{"{T}: Put X +1/+1 counters on target creature, where X is Druid's power.", ParseContext{CardName: "Druid"}, DynamicAmountSourcePower, DynamicAmountWhereX, 1, SelectorUnknown, ControllerAny, "where X is Druid's power"},
		{"{T}: Put X +1/+1 counters on target creature, where X is Fight Bear's power.", ParseContext{CardName: "Fight Bear"}, DynamicAmountSourcePower, DynamicAmountWhereX, 1, SelectorUnknown, ControllerAny, "where X is Fight Bear's power"},
		{"You gain 2 life for each basic land type among lands you control.", ParseContext{InstantOrSorcery: true}, DynamicAmountBasicLandTypes, DynamicAmountForEach, 2, SelectorUnknown, ControllerAny, "for each basic land type among lands you control"},
		{"Flames deals damage equal to the number of basic land types among lands you control to any target.", ParseContext{CardName: "Flames", InstantOrSorcery: true}, DynamicAmountBasicLandTypes, DynamicAmountEqual, 1, SelectorUnknown, ControllerAny, "equal to the number of basic land types among lands you control"},
		{"Flames deals X damage to any target, where X is the number of basic land types among lands you control.", ParseContext{CardName: "Flames", InstantOrSorcery: true}, DynamicAmountBasicLandTypes, DynamicAmountWhereX, 1, SelectorUnknown, ControllerAny, "where X is the number of basic land types among lands you control"},
	}

	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := Compile(test.source, test.context)
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			amount := compilation.Abilities[0].Content.Effects[0].Amount
			if amount.DynamicKind != test.kind ||
				amount.DynamicForm != test.form ||
				amount.Multiplier != test.multiplier ||
				amount.Selector.Kind != test.selector ||
				amount.Selector.Controller != test.controller ||
				amount.Text != test.text {
				t.Fatalf("amount = %#v tokens = %#v", amount, compilation.Syntax.Abilities[0].Tokens)
			}
			if test.kind == DynamicAmountSourcePower && amount.ReferenceSpan == (Span{}) {
				t.Fatal("source-power amount has no reference span")
			}
		})
	}
}

func TestCompileWithCyclingTargetSelector(t *testing.T) {
	t.Parallel()
	source := "Return up to two target cards with cycling from your graveyard to your hand."
	compilation, diagnostics := Compile(source, ParseContext{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	target := compilation.Abilities[0].Content.Targets[0]
	if target.Cardinality.Min != 0 || target.Cardinality.Max != 2 {
		t.Fatalf("cardinality = %#v, want up to two", target.Cardinality)
	}
	if target.Selector.Kind != SelectorCard || target.Selector.Keyword != "Cycling" {
		t.Fatalf("selector = %#v, want card with Cycling", target.Selector)
	}
}

func TestCompileDynamicCardCountWithCyclingInGraveyard(t *testing.T) {
	t.Parallel()
	source := "Flare deals X damage to any target, where X is the number of cards with a cycling ability in your graveyard."
	compilation, diagnostics := Compile(source, ParseContext{CardName: "Flare", InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	amount := compilation.Abilities[0].Content.Effects[0].Amount
	if amount.DynamicKind != DynamicAmountCount ||
		amount.DynamicForm != DynamicAmountWhereX ||
		amount.Selector.Kind != SelectorCard ||
		amount.Selector.Keyword != "Cycling" ||
		amount.Selector.Zone != zone.Graveyard ||
		amount.Selector.Controller != ControllerYou {
		t.Fatalf("amount = %#v, want count of cards with Cycling in your graveyard", amount)
	}
}

func TestCompileNamedCounterKinds(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		kind counter.Kind
	}{
		{"+1/+1", counter.PlusOnePlusOne},
		{"charge", counter.Charge},
		{"first strike", counter.FirstStrike},
		{"poison", counter.Poison},
		{"experience", counter.Experience},
	}
	for _, test := range tests {
		source := "Put a " + test.name + " counter on target permanent."
		compilation, diagnostics := Compile(source, ParseContext{InstantOrSorcery: true})
		if len(diagnostics) != 0 {
			t.Fatalf("%q diagnostics = %#v", source, diagnostics)
		}
		effect := compilation.Abilities[0].Content.Effects[0]
		if !effect.CounterKindKnown || effect.CounterKind != test.kind {
			t.Fatalf("%q counter kind = %v, %v", source, effect.CounterKind, effect.CounterKindKnown)
		}
	}

	compilation, diagnostics := Compile(
		"Put a quest counter on target permanent.",
		ParseContext{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("unknown counter diagnostics = %#v", diagnostics)
	}
	if compilation.Abilities[0].Content.Effects[0].CounterKindKnown {
		t.Fatal("unknown counter kind was recognized")
	}
}

func TestCompileEntersWithCounterKind(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := Compile(
		"This creature enters with three +1/+1 counters on it.",
		ParseContext{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effect := compilation.Abilities[0].Content.Effects[0]
	if effect.Kind != EffectEnterTapped ||
		!effect.CounterKindKnown ||
		effect.CounterKind != counter.PlusOnePlusOne ||
		!effect.Amount.Known ||
		effect.Amount.Value != 3 {
		t.Fatalf("effect = %#v, want fixed +1/+1 ETB counters", effect)
	}
}

func TestCompileNamedCounterKindsRejectsMissingRuntimeMechanics(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"stun", "finality"} {
		source := "Put a " + name + " counter on target creature."
		compilation, diagnostics := Compile(source, ParseContext{InstantOrSorcery: true})
		if len(diagnostics) != 0 {
			t.Fatalf("%q diagnostics = %#v", source, diagnostics)
		}
		effect := compilation.Abilities[0].Content.Effects[0]
		if effect.CounterKindKnown {
			t.Fatalf("%q counter kind was accepted for placement", source)
		}
	}
}

func TestCompileDynamicEffectAmountsRejectsAmbiguousSubjects(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Swarm deals damage equal to the number of cards in your hand to any target.",
		"Swarm deals damage equal to the number of creatures you control plus one to any target.",
		"You gain 2 life for each opponent and creature.",
		"Swarm deals damage equal to creatures you control to any target.",
		"You gain X life, where X is opponent.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := Compile(source, ParseContext{
				CardName:         "Swarm",
				InstantOrSorcery: true,
			})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if amount := compilation.Abilities[0].Content.Effects[0].Amount; amount.DynamicKind != DynamicAmountNone || amount.Known {
				t.Fatalf("amount = %#v, want unsupported", amount)
			}
		})
	}
}

func TestCompileDynamicEffectAmountsRejectsNumberDisagreement(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Draw a card for each creatures you control.",
		"Swarm deals damage equal to the number of creature you control to any target.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := Compile(source, ParseContext{
				CardName:         "Swarm",
				InstantOrSorcery: true,
			})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if amount := compilation.Abilities[0].Content.Effects[0].Amount; amount.DynamicKind != DynamicAmountNone || amount.Known {
				t.Fatalf("amount = %#v, want unsupported", amount)
			}
		})
	}
}

func TestCompileEffectAmountsAreClauseLocal(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		source string
		check  func(*testing.T, []CompiledEffect)
	}{
		{
			name:   "fixed then dynamic effect",
			source: "You gain 2 life, then draw a card for each creature you control.",
			check: func(t *testing.T, effects []CompiledEffect) {
				t.Helper()
				assertFixedEffectAmount(t, effects, EffectGain, 2)
				assertDynamicEffectAmount(t, effects, EffectDraw, DynamicAmountCount)
			},
		},
		{
			name:   "dynamic then fixed effect",
			source: "Draw a card for each creature you control, then you gain 2 life.",
			check: func(t *testing.T, effects []CompiledEffect) {
				t.Helper()
				assertDynamicEffectAmount(t, effects, EffectDraw, DynamicAmountCount)
				assertFixedEffectAmount(t, effects, EffectGain, 2)
			},
		},
		{
			name:   "and separates effects",
			source: "Draw a card for each creature you control and gain 2 life.",
			check: func(t *testing.T, effects []CompiledEffect) {
				t.Helper()
				assertDynamicEffectAmount(t, effects, EffectDraw, DynamicAmountCount)
				assertFixedEffectAmount(t, effects, EffectGain, 2)
			},
		},
		{
			name:   "fixed before condition formula",
			source: "You gain 2 life if the number of creatures you control is greater than 3.",
			check: func(t *testing.T, effects []CompiledEffect) {
				t.Helper()
				assertFixedEffectAmount(t, effects, EffectGain, 2)
			},
		},
		{
			name:   "dynamic before condition amount",
			source: "Draw a card for each creature you control unless your life total is 2.",
			check: func(t *testing.T, effects []CompiledEffect) {
				t.Helper()
				assertDynamicEffectAmount(t, effects, EffectDraw, DynamicAmountCount)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := Compile(test.source, ParseContext{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			test.check(t, compilation.Abilities[0].Content.Effects)
		})
	}
}

func assertFixedEffectAmount(t *testing.T, effects []CompiledEffect, kind EffectKind, value int) {
	t.Helper()
	for _, effect := range effects {
		if effect.Kind == kind {
			if !effect.Amount.Known ||
				effect.Amount.Value != value ||
				effect.Amount.DynamicKind != DynamicAmountNone {
				t.Fatalf("%v amount = %#v, want fixed %d", kind, effect.Amount, value)
			}
			return
		}
	}
	t.Fatalf("effects = %#v, missing %v", effects, kind)
}

func assertDynamicEffectAmount(t *testing.T, effects []CompiledEffect, kind EffectKind, dynamicKind DynamicAmountKind) {
	t.Helper()
	for _, effect := range effects {
		if effect.Kind == kind {
			if effect.Amount.Known || effect.Amount.DynamicKind != dynamicKind {
				t.Fatalf("%v amount = %#v, want dynamic %v", kind, effect.Amount, dynamicKind)
			}
			return
		}
	}
	t.Fatalf("effects = %#v, missing %v", effects, kind)
}

func TestCompileStaticPTBuffSubjects(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source          string
		wantSubject     StaticSubjectKind
		wantSubjectText string
		wantPower       CompiledSignedAmount
		wantToughness   CompiledSignedAmount
	}{
		"enchanted creature": {
			source:          "Enchanted creature gets +2/+2.",
			wantSubject:     StaticSubjectAttachedObject,
			wantSubjectText: "Enchanted creature",
			wantPower:       CompiledSignedAmount{Value: 2, Known: true},
			wantToughness:   CompiledSignedAmount{Value: 2, Known: true},
		},
		"equipped creature": {
			source:          "Equipped creature gets -3/-1.",
			wantSubject:     StaticSubjectAttachedObject,
			wantSubjectText: "Equipped creature",
			wantPower:       CompiledSignedAmount{Value: 3, Known: true, Negative: true},
			wantToughness:   CompiledSignedAmount{Value: 1, Known: true, Negative: true},
		},
		"other creatures you control": {
			source:          "Other creatures you control get +1/+1.",
			wantSubject:     StaticSubjectOtherControlledCreatures,
			wantSubjectText: "Other creatures you control",
			wantPower:       CompiledSignedAmount{Value: 1, Known: true},
			wantToughness:   CompiledSignedAmount{Value: 1, Known: true},
		},
		"creatures you control": {
			source:          "Creatures you control get +0/+2.",
			wantSubject:     StaticSubjectControlledCreatures,
			wantSubjectText: "Creatures you control",
			wantPower:       CompiledSignedAmount{Value: 0, Known: true},
			wantToughness:   CompiledSignedAmount{Value: 2, Known: true},
		},
		"each wall you control": {
			source:          "Each Wall you control gets +0/+2.",
			wantSubject:     StaticSubjectControlledWalls,
			wantSubjectText: "Each Wall you control",
			wantPower:       CompiledSignedAmount{Value: 0, Known: true},
			wantToughness:   CompiledSignedAmount{Value: 2, Known: true},
		},
		"artifacts you control": {
			source:          "Artifacts you control get +1/+1.",
			wantSubject:     StaticSubjectControlledArtifacts,
			wantSubjectText: "Artifacts you control",
			wantPower:       CompiledSignedAmount{Value: 1, Known: true},
			wantToughness:   CompiledSignedAmount{Value: 1, Known: true},
		},
		"tokens you control": {
			source:          "Tokens you control get +1/+1.",
			wantSubject:     StaticSubjectControlledTokens,
			wantSubjectText: "Tokens you control",
			wantPower:       CompiledSignedAmount{Value: 1, Known: true},
			wantToughness:   CompiledSignedAmount{Value: 1, Known: true},
		},
		"creatures your opponents control": {
			source:          "Creatures your opponents control get -1/-0.",
			wantSubject:     StaticSubjectOpponentControlledCreatures,
			wantSubjectText: "Creatures your opponents control",
			wantPower:       CompiledSignedAmount{Value: 1, Known: true, Negative: true},
			wantToughness:   CompiledSignedAmount{Value: 0, Known: true, Negative: true},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := Compile(test.source, ParseContext{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if len(compilation.Abilities) != 1 {
				t.Fatalf("abilities = %d, want 1", len(compilation.Abilities))
			}
			ability := compilation.Abilities[0]
			if len(ability.Content.Effects) != 1 || ability.Content.Effects[0].Kind != EffectModifyPT {
				t.Fatalf("effects = %#v", ability.Content.Effects)
			}
			effect := ability.Content.Effects[0]
			if effect.StaticSubject != test.wantSubject {
				t.Fatalf("static subject = %v, want %v", effect.StaticSubject, test.wantSubject)
			}
			if got := test.source[effect.StaticSubjectSpan.Start.Offset:effect.StaticSubjectSpan.End.Offset]; got != test.wantSubjectText {
				t.Fatalf("subject span text = %q, want %q", got, test.wantSubjectText)
			}
			if effect.PowerDelta != test.wantPower || effect.ToughnessDelta != test.wantToughness {
				t.Fatalf("PT = %+v / %+v, want %+v / %+v", effect.PowerDelta, effect.ToughnessDelta, test.wantPower, test.wantToughness)
			}
		})
	}
}

func TestCompileStaticKeywordGrantSubjects(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source             string
		wantSubject        StaticSubjectKind
		wantSubjectSubtype string
		keywords           []string
	}{
		"enchanted creature": {
			source:      "Enchanted creature has menace.",
			wantSubject: StaticSubjectAttachedObject,
			keywords:    []string{"Menace"},
		},
		"equipped creature": {
			source:      "Equipped creature has flying and first strike.",
			wantSubject: StaticSubjectAttachedObject,
			keywords:    []string{"Flying", "First strike"},
		},
		"double strike": {
			source:      "Equipped creature has double strike.",
			wantSubject: StaticSubjectAttachedObject,
			keywords:    []string{"Double strike"},
		},
		"other creatures": {
			source:      "Other creatures you control have flying.",
			wantSubject: StaticSubjectOtherControlledCreatures,
			keywords:    []string{"Flying"},
		},
		"controlled creatures": {
			source:      "Creatures you control have haste.",
			wantSubject: StaticSubjectControlledCreatures,
			keywords:    []string{"Haste"},
		},
		"controlled artifacts": {
			source:      "Artifacts you control have indestructible.",
			wantSubject: StaticSubjectControlledArtifacts,
			keywords:    []string{"Indestructible"},
		},
		"controlled subtype": {
			source:             "Zombies you control have flying.",
			wantSubject:        StaticSubjectControlledCreatureSubtype,
			wantSubjectSubtype: "Zombies",
			keywords:           []string{"Flying"},
		},
		"other controlled subtype": {
			source:             "Other Dinosaurs you control have haste.",
			wantSubject:        StaticSubjectOtherControlledCreatureSubtype,
			wantSubjectSubtype: "Dinosaurs",
			keywords:           []string{"Haste"},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := Compile(test.source, ParseContext{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			ability := compilation.Abilities[0]
			if len(ability.Content.Effects) != 1 || ability.Content.Effects[0].Kind != EffectGrantKeyword {
				t.Fatalf("effects = %#v", ability.Content.Effects)
			}
			if got := ability.Content.Effects[0].StaticSubject; got != test.wantSubject {
				t.Fatalf("static subject = %v, want %v", got, test.wantSubject)
			}
			if got := ability.Content.Effects[0].StaticSubjectSubtype; got != test.wantSubjectSubtype {
				t.Fatalf("static subject subtype = %q, want %q", got, test.wantSubjectSubtype)
			}
			if len(ability.Content.Keywords) != len(test.keywords) {
				t.Fatalf("keywords = %#v, want %v", ability.Content.Keywords, test.keywords)
			}
			for i, keyword := range ability.Content.Keywords {
				if keyword.Name != test.keywords[i] {
					t.Fatalf("keyword %d = %q, want %q", i, keyword.Name, test.keywords[i])
				}
			}
		})
	}
}

func TestCompileStaticPTBuffWithKeywordHasOneEffect(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := Compile(
		"Creatures you control get +1/+1 and have vigilance.",
		ParseContext{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if len(ability.Content.Effects) != 1 || ability.Content.Effects[0].Kind != EffectModifyPT {
		t.Fatalf("effects = %#v", ability.Content.Effects)
	}
}

func TestCompileStaticDeclarationsCarryClosedGroupSelectionAndLayer(t *testing.T) {
	t.Parallel()
	source := "Creatures your opponents control get -1/-0."
	compilation, diagnostics := Compile(source, ParseContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Static == nil || len(ability.Static.Declarations) != 1 {
		t.Fatalf("static semantics = %#v, want one declaration", ability.Static)
	}
	declaration := ability.Static.Declarations[0]
	if declaration.Kind != StaticDeclarationContinuous ||
		declaration.Continuous.Layer != StaticLayerPowerToughnessModify ||
		declaration.Continuous.Operation != StaticContinuousModifyPowerToughness {
		t.Fatalf("declaration = %#v, want power/toughness continuous declaration", declaration)
	}
	if declaration.Group.Domain != StaticGroupBattlefield ||
		declaration.Group.Selection.Controller != ControllerOpponent ||
		!slices.Equal(declaration.Group.Selection.RequiredTypes, []StaticCardType{StaticCardTypeCreature}) {
		t.Fatalf("group = %#v, want opponent-controlled battlefield creatures", declaration.Group)
	}
	if got := source[declaration.Group.Span.Start.Offset:declaration.Group.Span.End.Offset]; got != "Creatures your opponents control" {
		t.Fatalf("group span = %q", got)
	}
}

func TestCompileStaticDeclarationsCarryConditionsAndRuleDomains(t *testing.T) {
	t.Parallel()
	source := "As long as you control an artifact, this creature has flying."
	compilation, diagnostics := Compile(source, ParseContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	declaration := compilation.Abilities[0].Static.Declarations[0]
	if declaration.Group.Domain != StaticGroupSource ||
		declaration.Condition == nil ||
		declaration.Condition.Predicate != ConditionPredicateControllerControls {
		t.Fatalf("declaration = %#v, want conditional source declaration", declaration)
	}
	if declaration.Continuous.Layer != StaticLayerAbility ||
		declaration.Continuous.Operation != StaticContinuousGrantKeywords {
		t.Fatalf("continuous declaration = %#v", declaration.Continuous)
	}

	compilation, diagnostics = Compile("This spell can't be countered.", ParseContext{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	declaration = compilation.Abilities[0].Static.Declarations[0]
	if declaration.Kind != StaticDeclarationRule ||
		declaration.Rule.Domain != StaticRuleDomainCountering ||
		declaration.Rule.Kind != StaticRuleCantBeCountered ||
		declaration.Rule.Zone != StaticZoneStack {
		t.Fatalf("rule declaration = %#v", declaration)
	}
}

func TestCompileMixedStaticParagraphProducesExactDeclarations(t *testing.T) {
	t.Parallel()
	source := "Delirium — As long as there are four or more card types among cards in your graveyard, Dragon's Rage Channeler gets +2/+2, has flying, and attacks each combat if able."
	compilation, diagnostics := Compile(source, ParseContext{CardName: "Dragon's Rage Channeler"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	if ability.Static == nil || len(ability.Static.Declarations) != 3 {
		t.Fatalf("static semantics = %#v, want three declarations", ability.Static)
	}
	if ability.Static.Declarations[0].Continuous.Layer != StaticLayerPowerToughnessModify ||
		ability.Static.Declarations[1].Continuous.Layer != StaticLayerAbility ||
		ability.Static.Declarations[2].Rule.Domain != StaticRuleDomainAttack ||
		ability.Static.Declarations[2].Rule.Kind != StaticRuleMustAttack {
		t.Fatalf("static declarations = %#v", ability.Static.Declarations)
	}
	for i, declaration := range ability.Static.Declarations {
		if declaration.Group.Domain != StaticGroupSource || declaration.Condition == nil {
			t.Fatalf("declaration %d = %#v, want conditional source declaration", i, declaration)
		}
		if declaration.Span.Start.Offset != 0 || declaration.Span.End.Offset != len(source) {
			t.Fatalf("declaration %d span = %#v, want entire paragraph", i, declaration.Span)
		}
	}
}

func TestCompileStaticDeclarationsFailClosedOnAdjacentSemantics(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source  string
		blocker StaticDeclarationBlocker
	}{
		"duration": {
			source:  "Creatures you control get +1/+1 until end of turn.",
			blocker: StaticDeclarationBlockerDuration,
		},
		"condition": {
			source:  "As long as the moon is full, creatures you control get +1/+1.",
			blocker: StaticDeclarationBlockerCondition,
		},
		"group": {
			source:  "All creatures get +1/+1.",
			blocker: StaticDeclarationBlockerGroup,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := Compile(test.source, ParseContext{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			ability := compilation.Abilities[0]
			if ability.Static == nil {
				t.Fatal("static semantics = nil, want blocker")
			}
			if len(ability.Static.Declarations) != 0 {
				t.Fatalf("static declarations = %#v, want none", ability.Static.Declarations)
			}
			if ability.Static.Blocker != test.blocker {
				t.Fatalf("static blocker = %v, want %v", ability.Static.Blocker, test.blocker)
			}
		})
	}
}

func TestCompileResolvingPTBuffHasNoStaticSubject(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := Compile(
		"Target creature gets +2/+2 until end of turn.",
		ParseContext{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effect := compilation.Abilities[0].Content.Effects[0]
	if effect.StaticSubject != StaticSubjectNone {
		t.Fatalf("static subject = %v, want StaticSubjectNone", effect.StaticSubject)
	}
	if effect.StaticSubjectSpan != (Span{}) {
		t.Fatalf("static subject span = %#v, want zero span", effect.StaticSubjectSpan)
	}
}

func TestCompileSurveilEffect(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := Compile("Surveil 2.", ParseContext{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := compilation.Abilities[0].Content.Effects
	if len(effects) != 1 ||
		effects[0].Kind != EffectSurveil ||
		effects[0].Amount != (CompiledAmount{Value: 2, Known: true}) {
		t.Fatalf("effects = %#v, want surveil 2", effects)
	}
}

func TestCompileInvestigateEffect(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := Compile("Investigate.", ParseContext{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := compilation.Abilities[0].Content.Effects
	if len(effects) != 1 || effects[0].Kind != EffectInvestigate {
		t.Fatalf("effects = %#v, want investigate", effects)
	}
}

func TestCompileProliferateEffect(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := Compile("Proliferate.", ParseContext{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := compilation.Abilities[0].Content.Effects
	if len(effects) != 1 || effects[0].Kind != EffectProliferate {
		t.Fatalf("effects = %#v, want proliferate", effects)
	}
}

func TestCompileRegenerateEffect(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := Compile(
		"Regenerate target creature.",
		ParseContext{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := compilation.Abilities[0].Content.Effects
	if len(effects) != 1 || effects[0].Kind != EffectRegenerate {
		t.Fatalf("effects = %#v, want regenerate", effects)
	}
}

func TestCompileCounterVerbAndNoun(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		wantKinds []EffectKind
	}{
		"Counter target spell.": {
			wantKinds: []EffectKind{EffectCounter},
		},
		"This spell counters target spell.": {
			wantKinds: []EffectKind{EffectCounter},
		},
		"Put two +1/+1 counters on target creature.": {
			wantKinds: []EffectKind{EffectPut},
		},
		"Remove a counter from this permanent: Draw a card.": {
			wantKinds: []EffectKind{EffectDraw},
		},
	}
	for source, test := range tests {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := Compile(source, ParseContext{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := compilation.Abilities[0].Content.Effects
			if len(effects) != len(test.wantKinds) {
				t.Fatalf("effects = %#v, want kinds %v", effects, test.wantKinds)
			}
			for i, want := range test.wantKinds {
				if effects[i].Kind != want {
					t.Fatalf("effect %d = %v, want %v", i, effects[i].Kind, want)
				}
			}
		})
	}
}

func TestCompileExactCounterAbilityTargets(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		text   string
		kind   SelectorKind
	}{
		{"Counter target activated ability.", "target activated ability", SelectorActivatedAbility},
		{"Counter target triggered ability.", "target triggered ability", SelectorTriggeredAbility},
		{"Counter target activated or triggered ability.", "target activated or triggered ability", SelectorActivatedOrTriggeredAbility},
		{"Counter target spell, activated ability, or triggered ability.", "target spell, activated ability, or triggered ability", SelectorSpellActivatedOrTriggeredAbility},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := Compile(test.source, ParseContext{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			targets := compilation.Abilities[0].Content.Targets
			if len(targets) != 1 || targets[0].Text != test.text || targets[0].Selector.Kind != test.kind {
				t.Fatalf("targets = %#v, want text %q kind %v", targets, test.text, test.kind)
			}
		})
	}
}

func TestCompileNegatedEffect(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := Compile("Players can't gain life.", ParseContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := compilation.Abilities[0].Content.Effects
	if len(effects) != 1 || effects[0].Kind != EffectGain || !effects[0].Negated {
		t.Fatalf("effects = %#v", effects)
	}
}

func TestCompileEntersTappedUnlessCondition(t *testing.T) {
	t.Parallel()
	source := "This land enters tapped unless you control two or more basic lands."
	compilation, diagnostics := Compile(source, ParseContext{})
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

func TestCompileArtifactAndEnchantmentEntersTappedReference(t *testing.T) {
	t.Parallel()
	tests := []string{
		"This artifact enters tapped.",
		"This enchantment enters tapped.",
	}
	for _, source := range tests {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := Compile(source, ParseContext{})
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
	source := "Start your engines!"
	compilation, diagnostics := Compile(source, ParseContext{})
	if len(compilation.Abilities) != 1 {
		t.Fatalf("abilities = %#v", compilation.Abilities)
	}
	if len(diagnostics) != 1 ||
		diagnostics[0].Severity != SeverityWarning ||
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
			context := ParseContext{
				CardName:         name,
				InstantOrSorcery: typeLine == "Instant" || typeLine == "Sorcery",
				Planeswalker:     typeLine == "Planeswalker" || typeLine == "Legendary Planeswalker",
			}
			compilation, diagnostics := Compile(source, context)
			for _, diagnostic := range diagnostics {
				if diagnostic.Severity == SeverityError {
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
		{"event subject", "When this creature enters, if it was kicked, draw a card.", ConditionIf, ConditionPredicateEventSubjectWasKicked, false},
		{"activation resource threshold", "{T}: Draw a card. Activate only if you have 10 or more life.", ConditionOnlyIf, ConditionPredicateControllerLifeAtLeast, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			compilation, _ := Compile(test.source, ParseContext{CardName: "Test Bear"})
			if len(compilation.Abilities) != 1 || len(compilation.Abilities[0].Content.Conditions) != 1 {
				t.Fatalf("compilation = %#v", compilation)
			}
			condition := compilation.Abilities[0].Content.Conditions[0]
			if condition.Kind != test.kind ||
				condition.Predicate != test.predicate ||
				condition.Negated != test.negated ||
				condition.Span.Start.Offset >= condition.Span.End.Offset ||
				test.source[condition.Span.Start.Offset:condition.Span.End.Offset] != condition.Text {
				t.Fatalf("condition = %#v", condition)
			}
		})
	}
}

func TestCompileConditionsRejectsNearMissWordingSemantically(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"When this creature enters, if you nearly control an artifact, draw a card.",
		"If a creature dealt damage by this creature this turn would die, exile it instead.",
	} {
		compilation, _ := Compile(source, ParseContext{CardName: "Test Bear"})
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
		{"trigger event subject", "Whenever a creature dies, return it to its owner's hand.", []ReferenceBinding{ReferenceBindingEventPermanent, ReferenceBindingEventPermanent}},
		{"zone-change event subject", "Whenever an artifact is put into a graveyard from the battlefield, return it to its owner's hand.", []ReferenceBinding{ReferenceBindingEventPermanent, ReferenceBindingEventPermanent}},
		{"batched event subject is ambiguous", "Whenever one or more creatures die, return it to its owner's hand.", []ReferenceBinding{ReferenceBindingAmbiguous, ReferenceBindingAmbiguous}},
		{"explicit source in trigger body", "Whenever a creature dies, this creature deals 1 damage to its controller.", []ReferenceBinding{ReferenceBindingSource, ReferenceBindingSource}},
		{"single target occurrence", "Return target creature to its owner's hand.", []ReferenceBinding{ReferenceBindingTarget}},
		{"prior instruction result", "Exile target creature. Return it to the battlefield under its owner's control at the beginning of the next end step.", []ReferenceBinding{ReferenceBindingPriorInstructionResult, ReferenceBindingPriorInstructionResult}},
		{"delayed source", "When this creature enters, exile it at the beginning of the next end step.", []ReferenceBinding{ReferenceBindingSource, ReferenceBindingSource}},
		{"delayed non-self event subject", "When enchanted creature dies, return that card to the battlefield under its owner's control at the beginning of the next end step.", []ReferenceBinding{ReferenceBindingEventPermanent, ReferenceBindingEventPermanent}},
		{"activation cost source", "Remove a counter from it: Draw a card.", []ReferenceBinding{ReferenceBindingSource}},
		{"activation cost prior object", "Tap an untapped creature you control, Remove a +1/+1 counter from it: Draw a card.", []ReferenceBinding{ReferenceBindingAmbiguous}},
		{"activation cost prior source and object", "Remove a charge counter from this artifact, Tap an untapped creature you control, Remove a +1/+1 counter from it: Draw a card.", []ReferenceBinding{ReferenceBindingSource, ReferenceBindingAmbiguous}},
		{"ambiguous pronoun", "It explores.", []ReferenceBinding{ReferenceBindingAmbiguous}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			compilation, _ := Compile(test.source, ParseContext{CardName: "Test Bear", InstantOrSorcery: true})
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

func hasDiagnosticForSpan(diagnostics []Diagnostic, span Span) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Span == span {
			return true
		}
	}
	return false
}
