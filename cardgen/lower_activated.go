package cardgen

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

func lowerChapterAbility(
	cardName string,
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (abilityLowering, *shared.Diagnostic) {
	if len(ability.Chapters) == 0 || ability.ChapterSpan == (shared.Span{}) {
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported Saga chapter ability",
			"the executable source backend requires one or more chapter numbers",
		)
	}
	if syntax.BodySpan == (shared.Span{}) {
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported Saga chapter ability",
			"the executable source backend requires an em dash after the chapter numbers",
		)
	}
	bodySpan := syntax.BodySpan
	bodyText := strings.TrimSpace(
		ability.Text[bodySpan.Start.Offset-ability.Span.Start.Offset:],
	)
	bodyContent := ability.Content
	bodyContent.Keywords = keywordsWithinSpan(ability.Content.Keywords, bodySpan)
	if len(bodyContent.Keywords) != len(ability.Content.Keywords) {
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported Saga chapter ability",
			"the executable source backend requires chapter keywords to belong to a supported effect",
		)
	}
	bodySyntax := parser.Ability{
		Span:      bodySpan,
		Text:      bodyText,
		Tokens:    slices.Clone(parser.TokensInSpan(syntax.Tokens, syntax.BodySpan)),
		Reminders: syntax.Reminders,
		Quoted:    syntax.Quoted,
		Atoms:     syntax.Atoms,
	}
	content, diagnostic := lowerAbilityContent(cardName, bodyContent, false, &bodySyntax)
	if diagnostic != nil {
		return abilityLowering{}, diagnostic
	}
	spans := []shared.Span{ability.ChapterSpan, syntax.BodySeparatorSpan}
	for i := range ability.Content.Effects {
		spans = append(spans, ability.Content.Effects[i].Span)
	}
	for _, target := range ability.Content.Targets {
		spans = append(spans, target.Span)
	}
	for _, reference := range ability.Content.References {
		spans = append(spans, reference.Span)
	}
	for _, keyword := range ability.Content.Keywords {
		spans = append(spans, keyword.Span)
	}
	for _, reminder := range syntax.Reminders {
		spans = append(spans, reminder.Span)
	}
	return abilityLowering{
		chapterAbility: opt.Val(game.ChapterAbility{
			Text:     ability.Text,
			Chapters: slices.Clone(ability.Chapters),
			Content:  content,
		}),
		consumed: semanticConsumption{
			targets:    len(ability.Content.Targets),
			effects:    len(ability.Content.Effects),
			keywords:   len(ability.Content.Keywords),
			references: len(ability.Content.References),
		},
		sourceSpans: spans,
	}, nil
}

func lowerEntersPrepared(ability compiler.CompiledAbility, syntax *parser.Ability) (abilityLowering, bool) {
	if ability.Kind != compiler.AbilityStatic ||
		len(ability.Content.Effects) != 1 ||
		ability.Content.Effects[0].Kind != compiler.EffectEnterPrepared ||
		len(ability.Content.References) != 1 ||
		ability.Content.References[0].Binding != compiler.ReferenceBindingSource ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Modes) != 0 ||
		ability.Cost != nil ||
		ability.Trigger != nil {
		return abilityLowering{}, false
	}
	return abilityLowering{
		entersPrepared: true,
		consumed: semanticConsumption{
			effects:    1,
			references: 1,
		},
		sourceSpans: []shared.Span{syntax.Span},
	}, true
}

func lowerActivatedAbilityKind(
	cardName string,
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (abilityLowering, *shared.Diagnostic) {
	if isSemanticManaAbility(ability) {
		manaAbility, diagnostic := lowerManaAbility(cardName, ability, syntax)
		if diagnostic != nil {
			return abilityLowering{}, diagnostic
		}
		spans := []shared.Span{ability.Cost.Span}
		for i := range ability.Content.Effects {
			spans = append(spans, ability.Content.Effects[i].Span)
		}
		spans = append(spans, activationConditionSourceSpans(ability)...)
		if ability.ActivationTiming != compiler.ActivationTimingNone {
			spans = append(spans, ability.ActivationTimingSpan)
		}
		for _, reference := range ability.Content.References {
			spans = append(spans, reference.Span)
		}
		for _, reminder := range syntax.Reminders {
			spans = append(spans, reminder.Span)
		}
		return abilityLowering{
			manaAbility: opt.Val(manaAbility),
			consumed: semanticConsumption{
				cost:       true,
				conditions: len(ability.Content.Conditions),
				effects:    len(ability.Content.Effects),
				references: len(ability.Content.References),
			},
			sourceSpans: spans,
		}, nil
	}
	activatedAbility, diagnostic := lowerActivatedAbility(cardName, ability, syntax)
	if diagnostic != nil {
		return abilityLowering{}, diagnostic
	}
	spans := make(
		[]shared.Span,
		0,
		1+len(ability.Content.Effects)+len(ability.Content.Targets)+len(ability.Content.References)+len(syntax.Reminders),
	)
	spans = append(spans, ability.Cost.Span)
	if ability.ActivationTiming != compiler.ActivationTimingNone {
		spans = append(spans, ability.ActivationTimingSpan)
	}
	for i := range ability.Content.Effects {
		spans = append(spans, ability.Content.Effects[i].Span)
	}
	for _, target := range ability.Content.Targets {
		spans = append(spans, target.Span)
	}
	spans = append(spans, activationConditionSourceSpans(ability)...)
	for _, reference := range ability.Content.References {
		spans = append(spans, reference.Span)
	}
	spans = appendKeywordSpans(spans, ability.Content.Keywords)
	for _, reminder := range syntax.Reminders {
		spans = append(spans, reminder.Span)
	}
	if len(ability.Content.Modes) > 0 {
		spans = append(spans, ability.Span)
	}
	return abilityLowering{
		activatedAbility: opt.Val(activatedAbility),
		consumed: semanticConsumption{
			cost:       true,
			modes:      len(ability.Content.Modes),
			targets:    len(ability.Content.Targets),
			conditions: len(ability.Content.Conditions),
			effects:    len(ability.Content.Effects),
			keywords:   len(ability.Content.Keywords),
			references: len(ability.Content.References),
		},
		sourceSpans: spans,
	}, nil
}

func isSemanticManaAbility(ability compiler.CompiledAbility) bool {
	return !abilityContentHasTargets(ability.Content) && abilityContentHasAddManaEffect(ability.Content)
}

func abilityContentHasAddManaEffect(content compiler.AbilityContent) bool {
	if slices.ContainsFunc(content.Effects, func(effect compiler.CompiledEffect) bool {
		return effect.Kind == compiler.EffectAddMana
	}) {
		return true
	}
	return slices.ContainsFunc(content.Modes, func(mode compiler.CompiledMode) bool {
		return abilityContentHasAddManaEffect(mode.Content)
	})
}

func abilityContentHasTargets(content compiler.AbilityContent) bool {
	if len(content.Targets) != 0 {
		return true
	}
	return slices.ContainsFunc(content.Modes, func(mode compiler.CompiledMode) bool {
		return abilityContentHasTargets(mode.Content)
	})
}

// lowerLoyaltyAbility lowers an AbilityLoyalty into a game.LoyaltyAbility.
// It accepts only exact signed integer loyalty costs and supported single or
// ordered effect bodies. Variable costs (X) are rejected.
func lowerLoyaltyAbility(
	cardName string,
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (abilityLowering, *shared.Diagnostic) {
	const unsupportedDetail = "the executable source backend supports only exact signed loyalty costs with a supported effect body"
	if ability.Cost == nil ||
		len(ability.Cost.Components) != 1 ||
		ability.Cost.Components[0].Kind != compiler.CostLoyalty ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Modes) != 0 ||
		ability.AbilityWord != "" {
		return abilityLowering{}, executableDiagnostic(ability, "unsupported loyalty ability", unsupportedDetail)
	}
	loyaltyComponent := ability.Cost.Components[0]
	if !loyaltyComponent.AmountKnown {
		return abilityLowering{}, executableDiagnostic(ability, "unsupported loyalty ability", "the executable source backend supports only fixed integer loyalty costs, not variable costs")
	}
	loyaltyCost := loyaltyComponent.AmountValue

	if syntax.BodySpan == (shared.Span{}) {
		return abilityLowering{}, executableDiagnostic(ability, "unsupported loyalty ability", unsupportedDetail)
	}
	bodySpan := syntax.BodySpan
	bodyText := strings.TrimSpace(ability.Text[bodySpan.Start.Offset-ability.Span.Start.Offset:])
	bodyContent := ability.Content
	bodyContent.Keywords = keywordsWithinSpan(ability.Content.Keywords, bodySpan)
	if len(bodyContent.Keywords) != len(ability.Content.Keywords) {
		return abilityLowering{}, executableDiagnostic(ability, "unsupported loyalty ability", unsupportedDetail)
	}
	bodySyntax := parser.Ability{
		Span:      bodySpan,
		Text:      bodyText,
		Tokens:    parser.TokensInSpan(syntax.Tokens, syntax.BodySpan),
		Reminders: syntax.Reminders,
		Quoted:    syntax.Quoted,
		Atoms:     syntax.Atoms,
	}
	content, diagnostic := lowerAbilityContent(cardName, bodyContent, false, &bodySyntax)
	if diagnostic != nil {
		return abilityLowering{}, diagnostic
	}

	spans := make(
		[]shared.Span,
		0,
		1+len(ability.Content.Effects)+len(ability.Content.Targets)+len(ability.Content.References)+len(syntax.Reminders),
	)
	spans = append(spans, ability.Cost.Span)
	for i := range ability.Content.Effects {
		spans = append(spans, ability.Content.Effects[i].Span)
	}
	for _, target := range ability.Content.Targets {
		spans = append(spans, target.Span)
	}
	for _, reference := range ability.Content.References {
		spans = append(spans, reference.Span)
	}
	spans = appendKeywordSpans(spans, ability.Content.Keywords)
	for _, reminder := range syntax.Reminders {
		spans = append(spans, reminder.Span)
	}
	return abilityLowering{
		loyaltyAbility: opt.Val(game.LoyaltyAbility{
			Text:        ability.Text,
			LoyaltyCost: loyaltyCost,
			Content:     content,
		}),
		consumed: semanticConsumption{
			cost:       true,
			targets:    len(ability.Content.Targets),
			effects:    len(ability.Content.Effects),
			keywords:   len(ability.Content.Keywords),
			references: len(ability.Content.References),
		},
		sourceSpans: spans,
	}, nil
}

// lowerModalAbility lowers a modal spell/static shell. The modal body itself is
// lowered exclusively through lowerAbilityContent.
func lowerModalAbility(
	cardName string,
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (abilityLowering, *shared.Diagnostic) {
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		ability.Optional ||
		ability.AbilityWord != "" {
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported ability modes",
			"the executable source backend cannot lower this modal ability shell",
		)
	}
	switch ability.Kind {
	case compiler.AbilitySpell, compiler.AbilityStatic:
	default:
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported ability modes",
			"the executable source backend supports only spell or static modal abilities",
		)
	}
	content, diagnostic := lowerAbilityContent(cardName, ability.Content, ability.Optional, syntax)
	if diagnostic != nil {
		return abilityLowering{}, diagnostic
	}
	return abilityLowering{
		spellAbility: opt.Val(content),
		consumed: semanticConsumption{
			modes: len(ability.Content.Modes),
		},
		sourceSpans: []shared.Span{ability.Span},
	}, nil
}

func lowerModalContent(
	cardName string,
	ctx contentCtx,
	syntax *parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func(detail string) (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(ctx, "unsupported ability modes", detail)
	}
	if syntax.Modal == nil {
		return unsupported("the semantic modal content has no matching modal syntax")
	}
	if !syntax.Modal.ChoiceKnown {
		return unsupported("the executable source backend supports only exact \"Choose N\" and \"Choose one or both\" modal headers")
	}
	minModes, maxModes := syntax.Modal.MinModes, syntax.Modal.MaxModes
	if minModes < 1 || maxModes < minModes || maxModes > len(ctx.content.Modes) ||
		(minModes == 1 && maxModes == 2 && len(ctx.content.Modes) != 2) {
		return unsupported("the modal choice range does not match the number of modes")
	}
	if ctx.optional ||
		len(ctx.content.Effects) != 0 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.References) != 0 {
		return unsupported("the executable source backend does not support shared targets, effects, keywords, conditions, or references across modes")
	}
	if len(ctx.content.Modes) != len(syntax.Modal.Options) {
		return unsupported("semantic mode count does not match syntax mode count")
	}

	modes := make([]game.Mode, 0, len(ctx.content.Modes))
	for i, mode := range ctx.content.Modes {
		syntaxMode := syntax.Modal.Options[i]
		bodySyntax := parser.Ability{
			Span:      syntaxMode.Span,
			Text:      syntaxMode.Text,
			Tokens:    syntaxMode.Tokens,
			Reminders: syntaxMode.Reminders,
			Quoted:    syntaxMode.Quoted,
			Atoms:     syntaxMode.Atoms,
		}
		content, diagnostic := lowerAbilityContent(cardName, mode.Content, false, &bodySyntax)
		if diagnostic != nil {
			return game.AbilityContent{}, diagnostic
		}
		if content.IsModal() || len(content.Modes) != 1 {
			return unsupported("mode lowering produced unexpected modal content")
		}
		if !modalOptionCompletelyRecognized(mode.Content, syntaxMode) {
			return unsupported("a modal option contains rules text without complete executable semantics")
		}
		modes = append(modes, content.Modes[0])
	}
	return game.AbilityContent{
		Modes:    modes,
		MinModes: minModes,
		MaxModes: maxModes,
	}, nil
}

func modalOptionCompletelyRecognized(content compiler.AbilityContent, syntax parser.Mode) bool {
	var spans []shared.Span
	for i := range content.Effects {
		spans = append(spans, content.Effects[i].Span)
	}
	for _, target := range content.Targets {
		spans = append(spans, target.Span)
	}
	for _, condition := range content.Conditions {
		spans = append(spans, condition.Span)
	}
	for _, reference := range content.References {
		spans = append(spans, reference.Span)
	}
	spans = appendKeywordSpans(spans, content.Keywords)
	for _, reminder := range syntax.Reminders {
		spans = append(spans, reminder.Span)
	}
	for _, span := range syntax.CoverageSpans() {
		if spanCovered(span, spans) {
			continue
		}
		return false
	}
	return true
}

func lowerActivatedAbility(
	cardName string,
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.ActivatedAbility, *shared.Diagnostic) {
	shell, diagnostic := lowerActivationShell(cardName, ability, syntax)
	if diagnostic != nil {
		return game.ActivatedAbility{}, diagnostic
	}

	result := game.ActivatedAbility{
		Text:                shell.text,
		ManaCost:            shell.manaCost,
		AdditionalCosts:     shell.additionalCosts,
		ZoneOfFunction:      shell.zoneOfFunction,
		Timing:              shell.timing,
		ActivationCondition: shell.activationCondition,
		Content:             shell.content,
	}
	return result, nil
}

func prepareActivationCondition(ability *compiler.CompiledAbility, syntax *parser.Ability) (opt.V[game.Condition], bool) {
	if len(ability.Content.Conditions) == 0 {
		*syntax = syntaxWithoutAbilityWord(syntax)
		return opt.V[game.Condition]{}, true
	}
	if slices.ContainsFunc(ability.Content.Conditions, func(condition compiler.CompiledCondition) bool {
		return condition.Resolving
	}) {
		if !slices.ContainsFunc(ability.Content.Conditions, func(condition compiler.CompiledCondition) bool {
			return !condition.Resolving
		}) {
			return opt.V[game.Condition]{}, true
		}
		return opt.V[game.Condition]{}, false
	}
	if len(ability.Content.Conditions) != 1 {
		return opt.V[game.Condition]{}, false
	}
	condition, ok := lowerCondition(ability.Content.Conditions[0], conditionContextActivation)
	if !ok {
		return opt.V[game.Condition]{}, false
	}
	conditionSpan := []shared.Span{ability.Content.Conditions[0].Span}
	effects := slices.DeleteFunc(append([]compiler.CompiledEffect(nil), ability.Content.Effects...), func(effect compiler.CompiledEffect) bool {
		return spanCovered(effect.VerbSpan, conditionSpan)
	})
	bodyEffects := append([]compiler.CompiledEffect(nil), effects...)
	bodyEffects = appendModeEffects(bodyEffects, ability.Content.Modes)
	if len(bodyEffects) == 0 || slices.ContainsFunc(bodyEffects, func(effect compiler.CompiledEffect) bool {
		return effect.Span.End.Offset > ability.Content.Conditions[0].Span.Start.Offset
	}) {
		return opt.V[game.Condition]{}, false
	}
	ability.Content.Effects = effects
	ability.Content.Conditions = nil
	*syntax = syntaxWithoutAbilityWord(syntax)
	lastEffectEnd := bodyEffects[0].Span.End.Offset
	for i := 1; i < len(bodyEffects); i++ {
		lastEffectEnd = max(lastEffectEnd, bodyEffects[i].Span.End.Offset)
	}
	syntax.Tokens = slices.DeleteFunc(append([]shared.Token(nil), syntax.Tokens...), func(token shared.Token) bool {
		return token.Span.Start.Offset >= lastEffectEnd
	})
	return opt.Val(condition), true
}

func appendModeEffects(effects []compiler.CompiledEffect, modes []compiler.CompiledMode) []compiler.CompiledEffect {
	for _, mode := range modes {
		effects = append(effects, mode.Content.Effects...)
		effects = appendModeEffects(effects, mode.Content.Modes)
	}
	return effects
}

func activationConditionSourceSpans(ability compiler.CompiledAbility) []shared.Span {
	spans := make([]shared.Span, 0, len(ability.Content.Conditions)+1)
	for _, condition := range ability.Content.Conditions {
		spans = append(spans, condition.Span)
		if condition.ActivationKeywordSpan != (shared.Span{}) {
			spans = append(spans, condition.ActivationKeywordSpan)
		}
	}
	return spans
}

func lowerActivationTiming(timing compiler.ActivationTimingKind) (game.TimingRestriction, bool) {
	switch timing {
	case compiler.ActivationTimingNone:
		return game.NoTimingRestriction, true
	case compiler.ActivationTimingSorcery:
		return game.SorceryOnly, true
	case compiler.ActivationTimingOncePerTurn:
		return game.OncePerTurn, true
	case compiler.ActivationTimingSorceryOncePerTurn:
		return game.SorceryOncePerTurn, true
	case compiler.ActivationTimingDuringCombat:
		return game.DuringCombat, true
	case compiler.ActivationTimingDuringUpkeep:
		return game.DuringUpkeep, true
	case compiler.ActivationTimingDuringYourTurn:
		return game.DuringYourTurn, true
	default:
		return game.NoTimingRestriction, false
	}
}
