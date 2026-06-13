// Package compiler lowers parsed Oracle syntax into semantic intermediate
// representation for card generation.
package compiler

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Compile lowers a parsed syntax document into conservative semantic IR.
func Compile(document parser.Document, context Context) (Compilation, []shared.Diagnostic) {
	compilation := Compilation{Syntax: document}
	var diagnostics []shared.Diagnostic
	for _, ability := range document.Abilities {
		compiled, abilityDiagnostics := compileAbility(ability, context)
		compilation.Abilities = append(compilation.Abilities, compiled)
		diagnostics = append(diagnostics, abilityDiagnostics...)
	}
	return compilation, diagnostics
}

func compileAbility(
	ability parser.Ability,
	context Context,
) (CompiledAbility, []shared.Diagnostic) {
	var diagnostics []shared.Diagnostic
	kind := compileAbilityKind(ability.Kind)
	compiled := CompiledAbility{
		Kind: kind,
		Span: ability.Span,
		Text: ability.Text,
	}
	if ability.AbilityWord != nil {
		compiled.AbilityWord = ability.AbilityWord.Text
	}
	compiled.Chapters = append([]int(nil), ability.Chapters...)
	compiled.ChapterSpan = ability.ChapterSpan
	if ability.CostSyntax() != nil {
		cost := compileCost(*ability.CostSyntax())
		compiled.Cost = &cost
	}
	if kind == AbilityTriggered {
		trigger := compileTrigger(ability, context)
		compiled.Trigger = &trigger
	}
	if ability.Modal != nil {
		for _, mode := range ability.Modal.Options {
			compiledMode, modeDiagnostics := compileMode(mode, context)
			compiled.Content.Modes = append(compiled.Content.Modes, compiledMode)
			diagnostics = append(diagnostics, modeDiagnostics...)
		}
	}

	body := abilityBodyTokens(ability)
	timing, timingSpan := compileActivationTiming(kind, ability.ActivationRestrictions)
	if timing != ActivationTimingNone {
		body = tokensOutsideSpan(body, timingSpan)
		compiled.ActivationTiming = timing
		compiled.ActivationTimingSpan = timingSpan
	}
	tokens := semanticTokens(body, ability.Reminders, ability.Quoted)
	if kind == AbilityTriggered && ability.Optional() {
		compiled.Optional = true
		compiled.OptionalSpan = ability.OptionalSpan()
	}
	if kind == AbilityStatic && staticRuleSentencesOnly(ability.Sentences) {
		compiled.Content.Effects = compileEffects(ability.Sentences)
		applyEffectPaymentsToConditions(compiled.Content.Effects, compiled.Content.Conditions)
		compiled.Content.References = compileStaticRuleReferences(ability.Sentences)
	} else {
		compiled.Content.Keywords = compileKeywords(tokens, ability.Atoms)
		compiled.Content.Targets = compileTypedTargets(ability.Sentences)
		conditionTokens := tokens
		if kind == AbilityTriggered {
			conditionTokens = semanticTokens(ability.Tokens, ability.Reminders, ability.Quoted)
		}
		compiled.Content.Conditions = compileConditions(
			conditionTokens,
			kind == AbilityTriggered,
			ability.ConditionBoundaries(),
			ability.ConditionClauses(),
			ability.EventHistoryConditions(),
		)
		compiled.Content.Effects = compileEffects(ability.Sentences)
		applyEffectPaymentsToConditions(compiled.Content.Effects, compiled.Content.Conditions)
		referenceTokens := semanticTokens(ability.Tokens, ability.Reminders, ability.Quoted)
		if timing != ActivationTimingNone {
			referenceTokens = tokensOutsideSpan(referenceTokens, timingSpan)
		}
		compiled.Content.References = compileReferences(
			referenceTokens,
			ability.Atoms,
		)
		compiled.Content.References = bindReferences(
			compiled.Content.References,
			compiled.Content.Targets,
			compiled.Content.Effects,
			compiled.Trigger,
		)
	}
	compiled.Content.References = bindActivationCostReferences(compiled.Kind, compiled.Cost, compiled.Content.References)
	bindConditionReferences(compiled.Content.Conditions, compiled.Content.References, compiled.Trigger)
	applyEffectReferenceBindings(compiled.Content.Effects, compiled.Content.References)
	recognizeActivationZone(&compiled)
	if compiled.Trigger != nil && compiled.Trigger.Condition != nil {
		for i := range compiled.Content.Conditions {
			if compiled.Content.Conditions[i].Span == compiled.Trigger.Condition.Span {
				condition := compiled.Content.Conditions[i]
				compiled.Trigger.Condition = &condition
				compiled.Trigger.Pattern.InterveningCondition = &condition
				break
			}

		}
	}
	recognizeStaticDeclarations(&compiled, ability)

	for _, mode := range compiled.Content.Modes {
		if len(mode.Content.Effects) == 0 && len(mode.Content.Keywords) == 0 {
			diagnostics = append(diagnostics, unsupportedDiagnostic(mode.Span, mode.Text))
		}
	}
	if kind != AbilityReminder && ability.Modal == nil &&
		len(compiled.Content.Effects) == 0 && len(compiled.Content.Keywords) == 0 &&
		!legacyEffectsPresent(ability.Sentences) &&
		(compiled.Static == nil || len(compiled.Static.Declarations) == 0) {
		diagnostics = append(diagnostics, unsupportedDiagnostic(ability.Span, ability.Text))
	}

	// Set Content.Span from the body token range after shell/timing extraction.
	// This is non-zero even for unrecognized content, and for activated/loyalty
	// abilities it excludes the cost span. For triggered abilities with an
	// optional prefix, advance past "you may" since that is shell semantics.
	compiled.Content.Span = shared.SpanOf(body)
	if compiled.Optional && len(tokens) >= 3 {
		compiled.Content.Span.Start = tokens[2].Span.Start
	}
	if compiled.Cost != nil {
		for _, component := range compiled.Cost.Components {
			if component.Kind == CostUnknown {
				diagnostics = append(diagnostics, shared.Diagnostic{
					Severity: shared.SeverityWarning,
					Summary:  "unsupported cost",
					Detail:   "the compiler preserved this cost component but did not assign executable semantics",
					Span:     component.Span,
				})
			}
		}
	}
	return compiled, diagnostics
}

func legacyEffectsPresent(sentences []parser.Sentence) bool {
	return slices.ContainsFunc(sentences, func(sentence parser.Sentence) bool {
		return sentence.LegacyEffects
	})
}

func applyEffectReferenceBindings(effects []CompiledEffect, references []CompiledReference) {
	for effectIndex := range effects {
		applyReferenceBindings(effects[effectIndex].References, references)
		applyReferenceBindings(effects[effectIndex].SubjectReferences, references)
	}
}

func applyReferenceBindings(ownedReferences, references []CompiledReference) {
	for referenceIndex, owned := range ownedReferences {
		for _, reference := range references {
			if reference.Span == owned.Span {
				ownedReferences[referenceIndex] = reference
				break
			}
		}
	}
}

func compileAbilityKind(kind parser.AbilityKind) AbilityKind {
	switch kind {
	case parser.AbilitySpell:
		return AbilitySpell
	case parser.AbilityActivated:
		return AbilityActivated
	case parser.AbilityLoyalty:
		return AbilityLoyalty
	case parser.AbilityChapter:
		return AbilityChapter
	case parser.AbilityTriggered:
		return AbilityTriggered
	case parser.AbilityReplacement:
		return AbilityReplacement
	case parser.AbilityStatic:
		return AbilityStatic
	case parser.AbilityReminder:
		return AbilityReminder
	default:
		return AbilityUnknown
	}
}

func compileActivationTiming(kind AbilityKind, restrictions []parser.ActivationRestriction) (ActivationTimingKind, shared.Span) {
	if kind != AbilityActivated || len(restrictions) == 0 {
		return ActivationTimingNone, shared.Span{}
	}
	span := shared.Span{
		Start: restrictions[0].Span.Start,
		End:   restrictions[len(restrictions)-1].Span.End,
	}
	compiled := make([]ActivationTimingKind, 0, len(restrictions))
	for i := range restrictions {
		compiled = append(compiled, compileActivationRestriction(&restrictions[i]))
	}
	if len(compiled) == 1 {
		return compiled[0], span
	}
	if len(compiled) == 2 &&
		(compiled[0] == ActivationTimingSorcery && compiled[1] == ActivationTimingOncePerTurn ||
			compiled[0] == ActivationTimingOncePerTurn && compiled[1] == ActivationTimingSorcery) {
		return ActivationTimingSorceryOncePerTurn, span
	}
	return ActivationTimingUnsupported, span
}

func compileActivationRestriction(restriction *parser.ActivationRestriction) ActivationTimingKind {
	switch restriction.Kind {
	case parser.ActivationRestrictionSorceryTiming:
		return ActivationTimingSorcery
	case parser.ActivationRestrictionFrequency:
		if restriction.Frequency.Count.Kind == parser.ActivationFrequencyCountOnce &&
			restriction.Frequency.Period.Kind == parser.ActivationFrequencyPeriodTurn {
			return ActivationTimingOncePerTurn
		}
	case parser.ActivationRestrictionPhaseStep:
		if restriction.PhaseStep.Name.Kind == parser.PhaseStepNameCombat &&
			restriction.PhaseStep.Player.Kind == parser.TriggerPlayerSelectorAny &&
			(restriction.PhaseStep.Quantifier.Kind == parser.PhaseStepQuantifierNone ||
				restriction.PhaseStep.Quantifier.Kind == parser.PhaseStepQuantifierEach) {
			return ActivationTimingDuringCombat
		}
		if restriction.PhaseStep.Name.Kind == parser.PhaseStepNameUpkeep &&
			restriction.PhaseStep.Player.Kind == parser.TriggerPlayerSelectorYou &&
			(restriction.PhaseStep.Quantifier.Kind == parser.PhaseStepQuantifierSingle ||
				restriction.PhaseStep.Quantifier.Kind == parser.PhaseStepQuantifierEachOf) {
			return ActivationTimingDuringUpkeep
		}
	default:
	}
	return ActivationTimingUnsupported
}

func recognizeActivationZone(ability *CompiledAbility) {
	if ability.Kind != AbilityActivated {
		return
	}
	ability.ActivationZone = zone.Battlefield
	if activationCostUsesSourceFromGraveyard(*ability) ||
		contentReturnsSourceFromGraveyard(ability.Content) {
		ability.ActivationZone = zone.Graveyard
	}
}

func activationCostUsesSourceFromGraveyard(ability CompiledAbility) bool {
	if ability.Cost == nil {
		return false
	}
	for _, reference := range ability.Content.References {
		if reference.Binding != ReferenceBindingSource || !spanContains(ability.Cost.Span, reference.Span) {
			continue
		}
		for _, component := range ability.Cost.Components {
			if spanContains(component.Span, reference.Span) &&
				component.SourceZone == zone.Graveyard {
				return true
			}
		}
	}
	return false
}

func contentReturnsSourceFromGraveyard(content AbilityContent) bool {
	for effectIndex := range content.Effects {
		effect := &content.Effects[effectIndex]
		if effect.Kind != EffectReturn || effect.FromZone != zone.Graveyard {
			continue
		}
		for _, reference := range content.References {
			if reference.Binding == ReferenceBindingSource &&
				referenceFollowsEffectVerbInClause(effectIndex, content.Effects, reference.Span) {
				return true
			}
		}
	}
	for _, mode := range content.Modes {
		if contentReturnsSourceFromGraveyard(mode.Content) {
			return true
		}
	}
	return false
}

func referenceFollowsEffectVerbInClause(effectIndex int, effects []CompiledEffect, reference shared.Span) bool {
	effect := effects[effectIndex]
	if reference.Start.Offset < effect.VerbSpan.End.Offset || reference.End.Offset > effect.Span.End.Offset {
		return false
	}
	for i := effectIndex + 1; i < len(effects); i++ {
		next := effects[i]
		if next.Span != effect.Span {
			continue
		}
		if next.VerbSpan.Start.Offset < reference.End.Offset {
			return false
		}
		break
	}
	return true
}

func tokensOutsideSpan(tokens []shared.Token, span shared.Span) []shared.Token {
	return slices.DeleteFunc(append([]shared.Token(nil), tokens...), func(token shared.Token) bool {
		return span.Start.Offset <= token.Span.Start.Offset &&
			span.End.Offset >= token.Span.End.Offset
	})
}

// tokensWithinSpan returns the contiguous run of tokens that lie within span.
// An empty span selects no tokens.
func tokensWithinSpan(tokens []shared.Token, span shared.Span) []shared.Token {
	var result []shared.Token
	for _, token := range tokens {
		if token.Span.Start.Offset >= span.Start.Offset && token.Span.End.Offset <= span.End.Offset {
			result = append(result, token)
		}
	}
	return result
}

func compileMode(
	mode parser.Mode,
	context Context,
) (CompiledMode, []shared.Diagnostic) {
	tokens := semanticTokens(mode.Tokens, mode.Reminders, mode.Quoted)
	targets := compileTypedTargets(mode.Sentences)
	effects := compileEffects(mode.Sentences)
	references := bindReferences(compileReferences(tokens, mode.Atoms), targets, effects, nil)
	applyEffectReferenceBindings(effects, references)
	compiled := CompiledMode{
		Span: mode.Span,
		Text: mode.Text,
		Content: AbilityContent{
			Targets:    targets,
			Conditions: compileConditions(tokens, false, mode.ConditionBoundaries, mode.ConditionClauses(), mode.EventHistoryConditions()),
			Effects:    effects,
			Keywords:   compileKeywords(tokens, mode.Atoms),
			References: references,
		},
	}
	applyEffectPaymentsToConditions(compiled.Content.Effects, compiled.Content.Conditions)
	// Set Content.Span to the mode's full source span: a modal option has no
	// shell cost or trigger, so the mode body IS the content.
	compiled.Content.Span = mode.Span
	return compiled, nil
}

// compileCost maps the parser's typed Cost onto the semantic cost IR. It reads
// typed cost components and never inspects retained cost text to derive meaning.
func compileCost(parserCost parser.Cost) CompiledCost {
	cost := CompiledCost{Span: parserCost.Span, Text: parserCost.Text}
	for _, component := range parserCost.Components {
		cost.Components = append(cost.Components, compileCostComponent(component))
	}
	return cost
}

func compileCostComponent(component parser.CostComponent) CostComponent {
	compiled := CostComponent{
		Kind:             compileCostKind(component.Kind),
		Span:             component.Span,
		Text:             component.Text,
		Symbol:           component.Symbol,
		Amount:           component.Amount,
		Object:           component.Object,
		AmountValue:      component.AmountValue,
		AmountKnown:      component.AmountKnown,
		AmountFromX:      component.AmountFromX,
		ObjectSupertype:  component.ObjectSupertype,
		SupertypeKnown:   component.SupertypeKnown,
		ObjectController: compilerControllerRelation(component.ObjectController),
		RequireTapped:    component.RequireTapped,
		RequireUntapped:  component.RequireUntapped,
		SourceZone:       component.SourceZone,
		ToZone:           component.ToZone,
		SourceSelf:       component.SourceSelf,
		CounterKind:      component.CounterKind,
		CounterKindKnown: component.CounterKindKnown,
		SubtypesAny:      append([]types.Sub(nil), component.SubtypesAny...),
	}
	if component.ObjectColorKnown {
		if mapped, ok := compilerColor(component.ObjectColor); ok {
			compiled.ObjectColor = mapped
			compiled.ObjectColorKnown = true
		}
	}
	applyCostObjectNoun(&compiled, component)
	return compiled
}

// applyCostObjectNoun derives the selector kind and card type from the parser's
// typed object noun. A card object selects SelectorCard with an optional card
// type; a permanent object maps the noun onto its permanent selector.
func applyCostObjectNoun(compiled *CostComponent, component parser.CostComponent) {
	if component.ObjectIsCard {
		compiled.ObjectKind = SelectorCard
		if typ, ok := costCardTypeFromNoun(component.ObjectNoun); ok {
			compiled.ObjectType = typ
			compiled.ObjectTypeKnown = true
		}
		return
	}
	annotateCostObjectNoun(compiled, component.ObjectNoun)
}

func compileCostKind(kind parser.CostComponentKind) CostKind {
	switch kind {
	case parser.CostComponentMana:
		return CostMana
	case parser.CostComponentTap:
		return CostTap
	case parser.CostComponentUntap:
		return CostUntap
	case parser.CostComponentSacrifice:
		return CostSacrifice
	case parser.CostComponentDiscard:
		return CostDiscard
	case parser.CostComponentPayLife:
		return CostPayLife
	case parser.CostComponentExile:
		return CostExile
	case parser.CostComponentRemoveCounter:
		return CostRemoveCounter
	case parser.CostComponentReveal:
		return CostReveal
	case parser.CostComponentTapPermanents:
		return CostTapPermanents
	case parser.CostComponentEnergy:
		return CostEnergy
	case parser.CostComponentReturn:
		return CostReturn
	case parser.CostComponentExert:
		return CostExert
	case parser.CostComponentMill:
		return CostMill
	case parser.CostComponentPutCounter:
		return CostPutCounter
	case parser.CostComponentCollectEvidence:
		return CostCollectEvidence
	case parser.CostComponentLoyalty:
		return CostLoyalty
	default:
		return CostUnknown
	}
}

// annotateCostObjectNoun maps a typed parser object noun onto the semantic
// permanent selector and card type. It consumes the typed noun atom and reads
// no cost text.
func annotateCostObjectNoun(component *CostComponent, noun parser.ObjectNoun) bool {
	switch noun {
	case parser.ObjectNounArtifact:
		component.ObjectKind = SelectorArtifact
		component.ObjectType = types.Artifact
		component.ObjectTypeKnown = true
		return true
	case parser.ObjectNounCreature:
		component.ObjectKind = SelectorCreature
		component.ObjectType = types.Creature
		component.ObjectTypeKnown = true
		return true
	case parser.ObjectNounEnchantment:
		component.ObjectKind = SelectorEnchantment
		component.ObjectType = types.Enchantment
		component.ObjectTypeKnown = true
		return true
	case parser.ObjectNounLand:
		component.ObjectKind = SelectorLand
		component.ObjectType = types.Land
		component.ObjectTypeKnown = true
		return true
	case parser.ObjectNounPermanent:
		component.ObjectKind = SelectorPermanent
		component.PermanentModifier = true
		return true
	case parser.ObjectNounCard:
		component.ObjectKind = SelectorCard
		return true
	default:
		return false
	}
}

func costCardTypeFromNoun(noun parser.ObjectNoun) (types.Card, bool) {
	switch noun {
	case parser.ObjectNounArtifact:
		return types.Artifact, true
	case parser.ObjectNounCreature:
		return types.Creature, true
	case parser.ObjectNounEnchantment:
		return types.Enchantment, true
	case parser.ObjectNounLand:
		return types.Land, true
	default:
		return "", false
	}
}

func compilerCardType(cardType parser.CardType) (types.Card, bool) {
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

func compilerSupertype(supertype parser.Supertype) (types.Super, bool) {
	switch supertype {
	case parser.SupertypeLegendary:
		return types.Legendary, true
	case parser.SupertypeSnow:
		return types.Snow, true
	case parser.SupertypeBasic:
		return types.Basic, true
	case parser.SupertypeWorld:
		return types.World, true
	default:
		return "", false
	}
}

func compilerColor(value parser.Color) (color.Color, bool) {
	switch value {
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

func compilerControllerRelation(relation parser.ControllerRelation) ControllerKind {
	switch relation {
	case parser.ControllerRelationYouControl:
		return ControllerYou
	case parser.ControllerRelationYouDontControl:
		return ControllerNotYou
	case parser.ControllerRelationOpponentControls:
		return ControllerOpponent
	default:
		return ControllerAny
	}
}

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

func compileKeywords(tokens []shared.Token, atoms parser.Atoms) []CompiledKeyword {
	syntaxKeywords := atoms.KeywordsWithin(tokens)
	keywords := make([]CompiledKeyword, 0, len(syntaxKeywords))
	for i := range syntaxKeywords {
		keyword := &syntaxKeywords[i]
		compiled := CompiledKeyword{
			Kind:          keyword.Kind,
			Name:          keyword.Kind.String(),
			Span:          keyword.Span,
			Text:          keyword.Text,
			Parameter:     keyword.Parameter.Text,
			ParameterKind: keyword.Parameter.Kind,
			ManaCost:      keyword.Parameter.ManaCost(),
			Integer:       keyword.Parameter.Integer(),
			EnchantTarget: keyword.Parameter.EnchantTarget(),
		}
		if keyword.Parameter.Kind == parser.KeywordParameterProtection {
			compiled.Protection, compiled.ProtectionKnown = compileProtectionKeyword(keyword.Parameter.Protection())
		}
		keywords = append(keywords, compiled)
	}
	return keywords
}

func compileProtectionKeyword(parameter parser.ProtectionParameter) (game.ProtectionKeyword, bool) {
	families := 0
	for _, present := range []bool{
		parameter.Everything,
		parameter.EachColor,
		parameter.Multicolored,
		parameter.Monocolored,
		len(parameter.FromColors) > 0,
		len(parameter.FromTypes) > 0,
		len(parameter.FromSubtypes) > 0,
	} {
		if present {
			families++
		}
	}
	if families != 1 {
		return game.ProtectionKeyword{}, false
	}
	protection := game.ProtectionKeyword{
		Everything:   parameter.Everything,
		EachColor:    parameter.EachColor,
		Multicolored: parameter.Multicolored,
		Monocolored:  parameter.Monocolored,
		FromSubtypes: append([]types.Sub(nil), parameter.FromSubtypes...),
	}
	for _, value := range parameter.FromColors {
		compiled, ok := compilerColor(value)
		if !ok {
			return game.ProtectionKeyword{}, false
		}
		protection.FromColors = append(protection.FromColors, compiled)
	}
	for _, value := range parameter.FromTypes {
		compiled, ok := runtimeCardTypeFromParser(value)
		if !ok {
			return game.ProtectionKeyword{}, false
		}
		protection.FromTypes = append(protection.FromTypes, compiled)
	}
	return protection, true
}

func compileReferences(tokens []shared.Token, atoms parser.Atoms) []CompiledReference {
	recognized := atoms.ReferencesWithin(tokens)
	references := make([]CompiledReference, 0, len(recognized))
	for _, reference := range recognized {
		references = append(references, CompiledReference{
			Kind:    compileReferenceKind(reference.Kind),
			Pronoun: compileReferencePronoun(reference.Pronoun),
			Span:    reference.Span,
			Text:    joinedSourceText(reference.Tokens),
		})
	}

	return references
}

func compileTypedReferences(recognized []parser.Reference) []CompiledReference {
	references := make([]CompiledReference, 0, len(recognized))
	for _, reference := range recognized {
		references = append(references, CompiledReference{
			Kind:    compileReferenceKind(reference.Kind),
			Pronoun: compileReferencePronoun(reference.Pronoun),
			Span:    reference.Span,
			Text:    joinedSourceText(reference.Tokens),
		})
	}
	return references
}

func compileReferencePronoun(pronoun parser.PronounKind) ReferencePronounKind {
	switch pronoun {
	case parser.PronounIt:
		return ReferencePronounIt
	case parser.PronounIts:
		return ReferencePronounIts
	case parser.PronounThey:
		return ReferencePronounThey
	case parser.PronounTheir:
		return ReferencePronounTheir
	case parser.PronounThem:
		return ReferencePronounThem
	case parser.PronounThose:
		return ReferencePronounThose
	default:
		return ReferencePronounUnknown
	}
}

func compileReferenceKind(kind parser.ReferenceKind) ReferenceKind {
	switch kind {
	case parser.ReferenceSelfName:
		return ReferenceSelfName
	case parser.ReferenceThisObject:
		return ReferenceThisObject
	case parser.ReferenceThatObject:
		return ReferenceThatObject
	case parser.ReferenceThatPlayer:
		return ReferenceThatPlayer
	case parser.ReferencePronoun:
		return ReferencePronoun
	default:
		return ReferenceUnknown
	}
}

func semanticTokens(tokens []shared.Token, reminders, quoted []parser.Delimited) []shared.Token {
	excluded := append(append([]parser.Delimited(nil), reminders...), quoted...)
	result := make([]shared.Token, 0, len(tokens))
	for _, token := range tokens {
		var skip bool
		for _, delimiter := range excluded {
			if token.Span.Start.Offset >= delimiter.Span.Start.Offset &&
				token.Span.End.Offset <= delimiter.Span.End.Offset {
				skip = true
				break
			}
		}
		if !skip {
			result = append(result, token)
		}
	}
	return result
}

func joinedSourceText(tokens []shared.Token) string {
	if len(tokens) == 0 {
		return ""
	}
	var builder strings.Builder
	for i, token := range tokens {
		if i > 0 && needsSemanticSpace(tokens[i-1], token) {
			_ = builder.WriteByte(' ')
		}
		_, _ = builder.WriteString(token.Text)
	}
	return builder.String()
}

func needsSemanticSpace(previous, current shared.Token) bool {
	if current.Kind == shared.Comma || current.Kind == shared.Period || current.Kind == shared.Colon ||
		current.Kind == shared.Semicolon || current.Kind == shared.RightParen ||
		previous.Kind == shared.LeftParen || previous.Kind == shared.Quote || current.Kind == shared.Quote {
		return false
	}
	if previous.Kind == shared.Plus || previous.Kind == shared.Minus || previous.Kind == shared.Slash ||
		current.Kind == shared.Slash {
		return false
	}
	return previous.Kind != shared.Symbol && current.Kind != shared.Symbol
}

func unsupportedDiagnostic(span shared.Span, text string) shared.Diagnostic {
	return shared.Diagnostic{
		Severity: shared.SeverityWarning,
		Summary:  "unsupported Oracle construct",
		Detail:   "the compiler preserved but did not confidently lower: " + text,
		Span:     span,
	}
}
