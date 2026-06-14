package compiler

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
)

func compileTrigger(ability parser.Ability, _ Context) CompiledTrigger {
	trigger := CompiledTrigger{
		Kind: TriggerUnknown,
	}
	if ability.Trigger == nil {
		return trigger
	}
	trigger.Span = ability.Trigger.Span
	trigger.Text = ability.Trigger.Text
	trigger.Event = ability.Trigger.Event.Text
	switch ability.Trigger.Introduction.Kind {
	case parser.TriggerIntroductionWhen:
		trigger.Kind = TriggerWhen
	case parser.TriggerIntroductionWhenever:
		trigger.Kind = TriggerWhenever
	case parser.TriggerIntroductionAt:
		trigger.Kind = TriggerAt
	default:
	}
	conditions := compileConditions(ability.Tokens, true, ability.ConditionBoundaries(), ability.ConditionClauses(), ability.EventHistoryConditions())
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
			Span:                 ability.Trigger.Event.Span,
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
// the parser's typed condition boundaries. It walks the caller's token stream
// only to locate each boundary's clause extent (by token kind) and to render the
// retained clause text; it derives no meaning from Oracle wording. The parser
// owns introducer recognition, duration classification, and the intervening-if
// position, so the compiler matches each boundary to a token by source position
// and consumes its typed kind mechanically.
func compileConditions(
	tokens []shared.Token,
	triggered bool,
	boundaries []parser.ConditionBoundary,
	clauses []parser.ConditionClause,
	eventHistories []parser.EventHistoryCondition,
) []CompiledCondition {
	var conditions []CompiledCondition
	for i := 0; i < len(tokens); i++ {
		boundary, ok := conditionBoundaryAt(boundaries, tokens[i].Span.Start)
		if !ok {
			continue
		}
		end := conditionEnd(tokens, i)
		if boundary.DurationSkip {
			i = end - 1
			continue
		}
		phrase := tokens[i:end]
		condition := CompiledCondition{
			Kind:                  compileConditionIntro(boundary.Kind),
			Span:                  shared.SpanOf(phrase),
			Text:                  joinedSourceText(phrase),
			Intervening:           triggered && boundary.Intervening,
			ActivationKeywordSpan: boundary.ActivationKeyword,
		}
		recognizeCondition(&condition, clauses, eventHistories)
		conditions = append(conditions, condition)
		i = end - 1
	}
	return conditions
}

// conditionBoundaryAt returns the boundary whose introducer begins at position,
// if any. Boundaries are keyed by absolute source position, so a scan stream
// consumes exactly the boundaries whose tokens it walks.
func conditionBoundaryAt(boundaries []parser.ConditionBoundary, position shared.Position) (parser.ConditionBoundary, bool) {
	for _, boundary := range boundaries {
		if boundary.Start.Offset == position.Offset {
			return boundary, true
		}
	}
	return parser.ConditionBoundary{}, false
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

func conditionEnd(tokens []shared.Token, start int) int {
	for i := start; i < len(tokens); i++ {
		if tokens[i].Kind == shared.Period || (i > start && tokens[i].Kind == shared.Comma) {
			return i
		}
	}
	return len(tokens)
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
		})
	}
	return references
}

func abilityBodyTokens(ability parser.Ability) []shared.Token {
	tokens := ability.Tokens
	if ability.AbilityWord != nil {
		if dash := shared.TopLevelIndex(tokens, shared.EmDash); dash >= 0 {
			tokens = tokens[dash+1:]
		}
	}
	switch ability.Kind {
	case parser.AbilityActivated, parser.AbilityLoyalty:
		if colon := shared.TopLevelIndex(tokens, shared.Colon); colon >= 0 {
			return tokens[colon+1:]
		}
	case parser.AbilityTriggered:
		return tokensWithinSpan(ability.Tokens, ability.BodySpan())
	default:
	}
	return tokens
}
