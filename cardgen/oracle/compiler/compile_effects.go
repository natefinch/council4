package compiler

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
)

func compileTrigger(ability *parser.Ability, _ Context) CompiledTrigger {
	trigger := CompiledTrigger{
		Kind: TriggerUnknown,
	}
	if ability.Trigger == nil {
		return trigger
	}
	trigger.Span = ability.Trigger.Span
	trigger.Text = ability.Trigger.Text
	trigger.Event = ability.Trigger.Event
	switch ability.Trigger.Introduction.Kind {
	case parser.TriggerIntroductionWhen:
		trigger.Kind = TriggerWhen
	case parser.TriggerIntroductionWhenever:
		trigger.Kind = TriggerWhenever
	case parser.TriggerIntroductionAt:
		trigger.Kind = TriggerAt
	default:
	}
	conditions := compileConditions(ability.TriggerConditionSegments, ability.ConditionClauses, ability.EventHistoryConditions)
	for i := range conditions {
		if conditions[i].Intervening {
			condition := conditions[i]
			trigger.Condition = &condition
			break
		}
	}
	switch {
	case ability.Trigger.PhaseStep != nil:
		trigger.Pattern = compilePhaseStepTriggerPattern(
			ability.Trigger.PhaseStep,
			trigger.Kind,
			trigger.Condition,
		)
	case ability.Trigger.PlayerEvent != nil:
		trigger.Pattern = compilePlayerEventTriggerPattern(
			ability.Trigger.PlayerEvent,
			trigger.Kind,
			trigger.Condition,
		)
	case ability.Trigger.TriggerEvent != nil:
		trigger.Pattern = compileTriggerEventPattern(
			ability.Trigger.TriggerEvent,
			trigger.Kind,
			trigger.Condition,
		)
	default:
		trigger.Pattern = TriggerPattern{
			Span:                 ability.Trigger.EventSpan,
			Kind:                 trigger.Kind,
			InterveningCondition: trigger.Condition,
		}
	}
	return trigger
}

func runtimeCardTypeFromParser(cardType parser.CardType) (types.Card, bool) {
	switch cardType {
	case parser.CardTypeArtifact:
		return types.Artifact, true
	case parser.CardTypeBattle:
		return types.Battle, true
	case parser.CardTypeCreature:
		return types.Creature, true
	case parser.CardTypeEnchantment:
		return types.Enchantment, true
	case parser.CardTypeInstant:
		return types.Instant, true
	case parser.CardTypeLand:
		return types.Land, true
	case parser.CardTypePlaneswalker:
		return types.Planeswalker, true
	case parser.CardTypeSorcery:
		return types.Sorcery, true
	default:
		return "", false
	}
}

func runtimeColorFromParser(colorValue parser.Color) (color.Color, bool) {
	switch colorValue {
	case parser.ColorWhite:
		return color.White, true
	case parser.ColorBlue:
		return color.Blue, true
	case parser.ColorBlack:
		return color.Black, true
	case parser.ColorRed:
		return color.Red, true
	case parser.ColorGreen:
		return color.Green, true
	default:
		return "", false
	}
}

// compileConditions builds the semantic conditions for an ability or mode from
// the parser's pre-segmented condition clauses. The parser owns introducer
// recognition, clause segmentation, and rendering; the compiler consumes each
// segment's typed kind, span, and rendered text mechanically and derives no
// meaning from Oracle wording.
func compileConditions(
	segments []parser.ConditionSegment,
	clauses []parser.ConditionClause,
	eventHistories []parser.EventHistoryCondition,
) []CompiledCondition {
	var conditions []CompiledCondition
	for _, segment := range segments {
		condition := CompiledCondition{
			Kind:                  compileConditionIntro(segment.Kind),
			Span:                  segment.Span,
			Text:                  segment.Text,
			Intervening:           segment.Intervening,
			ActivationKeywordSpan: segment.ActivationKeyword,
			NodeID:                segment.NodeID,
			ClauseIndex:           segment.ClauseIndex,
			EventHistoryIndex:     segment.EventHistoryIndex,
		}
		recognizeCondition(&condition, clauses, eventHistories)
		conditions = append(conditions, condition)
	}
	return conditions
}

func compileConditionIntro(kind parser.ConditionIntroKind) ConditionKind {
	switch kind {
	case parser.ConditionIntroIf:
		return ConditionIf
	case parser.ConditionIntroUnless:
		return ConditionUnless
	case parser.ConditionIntroOnlyIf:
		return ConditionOnlyIf
	case parser.ConditionIntroAsLongAs:
		return ConditionAsLongAs
	default:
		return ConditionUnknown
	}
}

func compileEffects(sentences []parser.Sentence) []CompiledEffect {
	var effects []CompiledEffect
	for _, sentence := range sentences {
		if sentence.StaticRule != nil {
			if effect, ok := compileStaticRuleEffect(sentence); ok {
				effects = append(effects, effect)
			}
			continue
		}
		for syntaxIndex := range sentence.Effects {
			syntax := &sentence.Effects[syntaxIndex]
			effects = append(effects, CompiledEffect{
				Kind:               compileEffectKind(syntax.Kind),
				Context:            syntax.Context,
				Connection:         syntax.Connection,
				ConnectionSpan:     syntax.ConnectionSpan,
				Span:               syntax.Span,
				ClauseSpan:         syntax.ClauseSpan,
				Text:               syntax.Text,
				VerbSpan:           syntax.VerbSpan,
				References:         compileTypedReferences(syntax.References),
				SubjectReferences:  compileTypedReferences(syntax.SubjectReferences),
				Targets:            compileTypedTargetList(syntax.Targets),
				SubjectTargets:     compileTypedTargetList(syntax.SubjectTargets),
				Duration:           compileEffectDuration(syntax.Duration),
				DelayedTiming:      compileDelayedTiming(syntax.DelayedTiming),
				Selector:           compileTypedSelection(syntax.Selection),
				Amount:             compileTypedAmount(syntax.Amount),
				PowerDelta:         compileSignedAmount(syntax.PowerDelta),
				ToughnessDelta:     compileSignedAmount(syntax.ToughnessDelta),
				StaticSubject:      compileStaticSubjectKind(syntax.StaticSubject.Kind),
				StaticSubjectSpan:  syntax.StaticSubject.Span,
				Details:            compiledEffectDetails(staticSubjectType(syntax.StaticSubject.SubtypeText, syntax.StaticSubject.Subtype, syntax.StaticSubject.SubtypeKnown), syntax.Symbol),
				CounterKind:        syntax.CounterKind,
				CounterKindKnown:   syntax.CounterKnown,
				FromZone:           syntax.FromZone,
				ToZone:             syntax.ToZone,
				Destination:        syntax.Destination,
				EntersTapped:       syntax.EntersTapped,
				EntersTappedSelf:   syntax.EntersTappedSelf,
				EntersWithCounters: syntax.EntersWithCounters,
				UnderYourControl:   syntax.UnderYourControl,
				CastAsAdventure:    syntax.CastAsAdventure,
				Negated:            syntax.Negated,
				Optional:           syntax.Optional,
				OptionalSpan:       syntax.OptionalSpan,
				Mana: CompiledEffectMana{
					Span:            syntax.Mana.Span,
					Symbols:         slices.Clone(syntax.Mana.Symbols),
					Choice:          syntax.Mana.Choice,
					AnyColor:        syntax.Mana.AnyColor,
					LegacyBodyExact: syntax.Mana.LegacyBodyExact,
				},
				Replacement:             syntax.Replacement,
				Payment:                 compileEffectPayment(syntax.Payment),
				Exact:                   syntax.Exact,
				RequiresOrderedLowering: syntax.RequiresOrderedLowering,
				HasUnrecognizedSibling:  syntax.HasUnrecognizedSibling,
				UnsupportedDetail:       syntax.UnsupportedDetail,
			})
		}
	}
	return effects
}

func compileStaticRuleEffect(sentence parser.Sentence) (CompiledEffect, bool) {
	rule, _, ok := semanticStaticRuleForSyntax(*sentence.StaticRule)
	if !ok {
		return CompiledEffect{}, false
	}
	kind := effectKindForStaticRule(rule)
	if kind == EffectUnknown {
		return CompiledEffect{}, false
	}
	selector := CompiledSelector{}
	if sentence.StaticRule.Subject.Kind == parser.StaticRuleSubjectSourceCreature {
		selector.Kind = SelectorCreature
	}
	return CompiledEffect{
		Kind:     kind,
		Span:     sentence.StaticRule.Span,
		Text:     sentence.Text,
		VerbSpan: sentence.StaticRule.Operation.Span,
		Selector: selector,
		Negated:  sentence.StaticRule.Constraint.Kind == parser.StaticRuleConstraintProhibition,
	}, true
}

func effectKindForStaticRule(rule StaticRuleKind) EffectKind {
	switch rule {
	case StaticRuleCantBlock:
		return EffectCantBlock
	case StaticRuleCantBeBlocked:
		return EffectCantBeBlocked
	case StaticRuleMustAttack:
		return EffectMustAttack
	case StaticRuleCantBeCountered:
		return EffectCantBeCountered
	default:
		return EffectUnknown
	}
}

func staticRuleSentencesOnly(sentences []parser.Sentence) bool {
	if len(sentences) == 0 {
		return false
	}
	for _, sentence := range sentences {
		if sentence.StaticRule == nil {
			return false
		}
	}
	return true
}

func compileStaticRuleReferences(sentences []parser.Sentence) []CompiledReference {
	references := make([]CompiledReference, 0, len(sentences))
	for i, sentence := range sentences {
		references = append(references, CompiledReference{
			Kind:       ReferenceThisObject,
			Span:       sentence.StaticRule.Subject.Span,
			Binding:    ReferenceBindingSource,
			Occurrence: i,
			NodeID:     i,
		})
	}
	return references
}
