package compiler

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
)

func TestCompileTriggeredAbility(t *testing.T) {
	t.Parallel()
	source := "Whenever a creature enters, if it was cast, draw a card."
	compilation, diagnostics := compileSource(source, pipelineContext{})
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
	compilation, diagnostics := compileSource(source, pipelineContext{})
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

func TestCompileNonPhaseTriggerUsesNormalizedSyntaxTokens(t *testing.T) {
	t.Parallel()
	source := "Whenever a  creature enters, draw a card."
	compilation, diagnostics := compileSource(source, pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	trigger := compilation.Abilities[0].Trigger
	if trigger.Event != "a  creature enters" {
		t.Fatalf("raw event = %q, want exact source metadata", trigger.Event)
	}
	if trigger.Pattern.Event != TriggerEventPermanentEnteredBattlefield {
		t.Fatalf("pattern = %#v, want normalized permanent-enter event", trigger.Pattern)
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
			compilation, diagnostics := compileSource(test.source, pipelineContext{})
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

func TestCompileSpellCastDisjunctionTriggerPatterns(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		source string
		check  func(*testing.T, TriggerPattern)
	}{
		{
			name:   "cast spell card-type disjunction",
			source: "Whenever you cast an artifact or enchantment spell, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventSpellCast ||
					!slices.Equal(pattern.CardSelection.RequiredTypesAny,
						[]TriggerCardType{TriggerCardTypeArtifact, TriggerCardTypeEnchantment}) {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			name:   "cast spell subtype disjunction",
			source: "Whenever you cast an Aura, Equipment, or Vehicle spell, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventSpellCast ||
					!slices.Equal(pattern.CardSelection.SubtypesAny,
						[]TriggerSubtype{types.Aura, types.Equipment, types.Vehicle}) {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(test.source, pipelineContext{})
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

func TestCompileSelfOrAnotherTriggerPattern(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"Whenever this creature or another Ally you control enters, draw a card.",
		pipelineContext{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	trigger := compilation.Abilities[0].Trigger
	if trigger == nil {
		t.Fatal("trigger = nil")
	}
	pattern := trigger.Pattern
	if pattern.Event != TriggerEventPermanentEnteredBattlefield ||
		pattern.Controller != ControllerYou ||
		!pattern.SubjectSelectionOrSelf ||
		pattern.ExcludeSelf ||
		pattern.Source != TriggerSourceAny ||
		!slices.Equal(pattern.SubjectSelection.SubtypesAny, []TriggerSubtype{types.Sub("Ally")}) {
		t.Fatalf("pattern = %#v", pattern)
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
					!slices.Equal(pattern.SubjectSelection.SubtypesAny, []TriggerSubtype{types.Forest}) {
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
					!slices.Equal(pattern.SubjectSelection.SubtypesAny, []TriggerSubtype{types.Clue}) {
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
			source: "Whenever you create a token, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventTokenCreated ||
					pattern.UnionEvent != TriggerEventUnknown ||
					pattern.Player != TriggerPlayerYou ||
					!pattern.SubjectSelection.TokenOnly {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			source: "Whenever you create or sacrifice a token, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventTokenCreated ||
					pattern.UnionEvent != TriggerEventPermanentSacrificed ||
					pattern.Player != TriggerPlayerYou ||
					!pattern.SubjectSelection.TokenOnly {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			source: "Whenever this creature attacks or becomes the target of a spell, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventObjectBecameTarget ||
					pattern.UnionEvent != TriggerEventAttackerDeclared ||
					pattern.Source != TriggerSourceSelf ||
					pattern.StackObject != TriggerStackObjectSpell {
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
		{
			source: "Whenever this creature attacks alone, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventAttackerDeclared ||
					pattern.Source != TriggerSourceSelf ||
					!pattern.AttackAlone {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			source: "Whenever you attack with two or more creatures, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventAttackerDeclared ||
					pattern.Controller != ControllerYou ||
					!pattern.OneOrMore ||
					pattern.AttackerCountAtLeast != 2 {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
		{
			source: "At the beginning of your next upkeep, draw a card.",
			check: func(t *testing.T, pattern TriggerPattern) {
				if pattern.Event != TriggerEventBeginningOfStep ||
					pattern.Step != TriggerStepUpkeep ||
					pattern.Controller != ControllerYou ||
					!pattern.NextOccurrence {
					t.Fatalf("pattern = %#v", pattern)
				}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(test.source, pipelineContext{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			test.check(t, compilation.Abilities[0].Trigger.Pattern)
		})
	}
}

func TestCompileEnterOrDiesUnionTriggerPattern(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"When this creature enters or dies, draw a card.", pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	pattern := compilation.Abilities[0].Trigger.Pattern
	if pattern.Event != TriggerEventPermanentEnteredBattlefield ||
		pattern.UnionEvent != TriggerEventPermanentDied ||
		pattern.Source != TriggerSourceSelf {
		t.Fatalf("pattern = %#v", pattern)
	}
}

func TestCompileSemanticTriggerPatternsFailClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Whenever this creature becomes the target of a spell or ability for the first time each turn, draw a card.",
		"Whenever creature you control becomes tapped, draw a card.",
		"At the beginning of your declare attackers step, draw a card.",
		"At the beginning of the upkeep of enchanted permanent's controller, draw a card.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(source, pipelineContext{})
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
	for _, test := range []struct {
		source  string
		event   TriggerEvent
		player  TriggerPlayerRelation
		ordinal int
	}{
		{
			source:  "Whenever you draw your second card each turn, create a 2/2 black Zombie creature token.",
			event:   TriggerEventCardDrawn,
			player:  TriggerPlayerYou,
			ordinal: 2,
		},
		{
			source:  "When an opponent loses life for the first time each turn, draw a card.",
			event:   TriggerEventLifeLost,
			player:  TriggerPlayerOpponent,
			ordinal: 1,
		},
	} {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(test.source, pipelineContext{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			pattern := compilation.Abilities[0].Trigger.Pattern
			if pattern.Event != test.event ||
				pattern.Player != test.player ||
				pattern.PlayerEventOrdinalThisTurn != test.ordinal {
				t.Fatalf("pattern = %#v", pattern)
			}
		})
	}
}

func TestCompileNamedSelfEnterTriggerPattern(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"When Example Card enters, draw a card.",
		pipelineContext{CardName: "Example Card"},
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
	compilation, diagnostics := compileSource(source, pipelineContext{})
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
	compilation, diagnostics := compileSource(source, pipelineContext{Saga: true})
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
	compilation, diagnostics := compileSource(source, pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	ability := compilation.Abilities[0]
	if !ability.Optional || source[ability.OptionalSpan.Start.Offset:ability.OptionalSpan.End.Offset] != "you may" {
		t.Fatalf("optional ability = %#v", ability)
	}
}
