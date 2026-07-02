package cardgen

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
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
		CoinFlip:  syntax.CoinFlip,
		Vote:      syntax.Vote,
		Modal:     syntax.Modal,
	}
	content, diagnostic := lowerAbilityContent(cardName, ability.Kind, bodyContent, false, &bodySyntax)
	if diagnostic != nil {
		return abilityLowering{}, diagnostic
	}
	spans := []shared.Span{ability.ChapterSpan, syntax.BodySeparatorSpan}
	if syntax.ChapterFlavorSpan != (shared.Span{}) {
		spans = append(spans, syntax.ChapterFlavorSpan)
	}
	for i := range ability.Content.Effects {
		spans = append(spans, ability.Content.Effects[i].Span)
	}
	for _, target := range ability.Content.Targets {
		spans = append(spans, target.Span)
		if target.ChoiceSpan != (shared.Span{}) {
			spans = append(spans, target.ChoiceSpan)
		}
	}
	for _, reference := range ability.Content.References {
		spans = append(spans, reference.Span)
	}
	for _, keyword := range ability.Content.Keywords {
		spans = append(spans, keyword.Span)
	}
	for i := range ability.Content.Conditions {
		spans = append(spans, ability.Content.Conditions[i].Span)
	}
	for _, reminder := range syntax.Reminders {
		spans = append(spans, reminder.Span)
	}
	chapterSpans := spans
	consumed := semanticConsumption{
		targets:    len(ability.Content.Targets),
		effects:    len(ability.Content.Effects),
		keywords:   len(ability.Content.Keywords),
		references: len(ability.Content.References),
		conditions: len(ability.Content.Conditions),
	}
	if len(ability.Content.Modes) > 0 {
		// A modal chapter ("I, II, III — Choose one at random — • ...") carries its
		// targets and effects inside each mode, which lowerModalContent verifies
		// for complete per-option coverage. Credit the whole chapter span so the
		// modal header and bullet tokens count as consumed, mirroring the modal
		// spell shell.
		consumed.modes = len(ability.Content.Modes)
		chapterSpans = append(chapterSpans, ability.Span)
	}
	return abilityLowering{
		chapterAbility: opt.Val(game.ChapterAbility{
			Text:     ability.Text,
			Chapters: slices.Clone(ability.Chapters),
			Content:  content,
		}),
		consumed:    consumed,
		sourceSpans: chapterSpans,
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
	if ability.SourceAbilityCostReduction != nil {
		spans = append(spans, ability.SourceAbilityCostReduction.Span)
	}
	for i := range ability.Content.Effects {
		spans = append(spans, ability.Content.Effects[i].Span)
		if ability.Content.Effects[i].CopyMayChooseNewTargets {
			spans = append(spans, ability.Content.Effects[i].CopyChooseNewTargetsRiderSpan)
		}
		if len(ability.Content.Effects[i].TokenCopyGrantKeywords) != 0 {
			spans = append(spans, ability.Content.Effects[i].TokenCopyGrantRiderSpan)
		}
	}
	for _, target := range ability.Content.Targets {
		spans = append(spans, target.Span)
		if target.ChoiceSpan != (shared.Span{}) {
			spans = append(spans, target.ChoiceSpan)
		}
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
	if ability.ExactSequence != compiler.ExactSequenceUnknown {
		spans = append(spans, ability.Content.Span)
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
	if abilityContentHasTargets(ability.Content) || !abilityContentHasAddManaEffect(ability.Content) {
		return false
	}
	// An add-mana ability whose body contains a condition-gated effect (e.g.
	// "If there are no depletion counters on this land, sacrifice it.") is not a
	// mana ability: the body-owned condition gates a non-mana effect that the
	// mana-ability lowerer cannot express. Route it through the activated-ability
	// path instead.
	if activationConditionOwnedByBody(ability.Content) {
		return false
	}
	return true
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
		CoinFlip:  syntax.CoinFlip,
		Vote:      syntax.Vote,
	}
	content, diagnostic := lowerAbilityContent(cardName, ability.Kind, bodyContent, false, &bodySyntax)
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
		if ability.Content.Effects[i].CopyMayChooseNewTargets {
			spans = append(spans, ability.Content.Effects[i].CopyChooseNewTargetsRiderSpan)
		}
		if len(ability.Content.Effects[i].TokenCopyGrantKeywords) != 0 {
			spans = append(spans, ability.Content.Effects[i].TokenCopyGrantRiderSpan)
		}
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
	content, diagnostic := lowerAbilityContent(cardName, ability.Kind, ability.Content, ability.Optional, syntax)
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

// hasNonSourceSharedReference reports whether the modal ability content carries
// a shared reference that is not the ability's own source permanent. A modal
// trigger body records the trigger subject ("When this creature enters, choose
// one — ...") as a content-level reference bound to the source, but each mode is
// lowered independently from its own content and resolves any "this creature"
// wording through its own per-mode reference, so a shared source reference is
// redundant and safe to allow. Any other shared reference (a target, event
// object, or prior-instruction result carried across the modes) still fails
// closed because the mode-independent lowering cannot thread it.
func hasNonSourceSharedReference(references []compiler.CompiledReference) bool {
	for _, reference := range references {
		if reference.Binding != compiler.ReferenceBindingSource {
			return true
		}
	}
	return false
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
	if len(ctx.content.Modes) == 0 || ctx.content.Modes[0].Modal == nil {
		return unsupported("the semantic modal content has no compiled modal choice")
	}
	modal := ctx.content.Modes[0].Modal
	minModes, maxModes := modal.MinModes, modal.MaxModes
	// An "up to N" header yields minModes 0 (the controller may decline every
	// mode); a fixed "choose N" header has minModes >= 1. Both are valid as long
	// as the upper bound stays within the available modes.
	if minModes < 0 || maxModes < 1 || maxModes < minModes || maxModes > len(ctx.content.Modes) ||
		(minModes == 1 && maxModes == 2 && len(ctx.content.Modes) != 2) {
		return unsupported("the modal choice range does not match the number of modes")
	}
	randomModes := modal.Kind == compiler.CompiledModalChoiceOneAtRandom
	if randomModes {
		// "Choose one at random" selects the single mode with the game's random
		// source rather than letting the controller choose. Only the triggered
		// and Saga-chapter resolution paths honor that random selection, so a
		// random modal in any other context (a modal spell) fails closed rather
		// than silently letting a player pick.
		if minModes != 1 || maxModes != 1 {
			return unsupported("an at-random modal must choose exactly one mode")
		}
		if ctx.enclosingKind != compiler.AbilityChapter && ctx.enclosingKind != compiler.AbilityTriggered {
			return unsupported("the executable source backend lowers at-random modes only in triggered or Saga-chapter abilities")
		}
	}
	var bonus game.ModeChoiceBonus
	switch modal.Bonus.Condition {
	case compiler.ModeChoiceBonusConditionNone:
		if modal.Bonus.AdditionalMaxModes != 0 {
			return unsupported("the modal choice bonus has no supported condition")
		}
	case compiler.ModeChoiceBonusConditionControlsCommander:
		if modal.Bonus.AdditionalMaxModes < 1 {
			return unsupported("the commander modal choice bonus must add at least one mode")
		}
		bonus = game.ModeChoiceBonus{
			Condition:          game.ModeChoiceConditionControlsCommander,
			AdditionalMaxModes: modal.Bonus.AdditionalMaxModes,
		}
	default:
		return unsupported("the modal choice bonus condition is unsupported")
	}
	if ctx.optional ||
		len(ctx.content.Effects) != 0 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		hasNonSourceSharedReference(ctx.content.References) {
		return unsupported("the executable source backend does not support shared targets, effects, keywords, conditions, or references across modes")
	}
	if len(ctx.content.Modes) != len(syntax.Modal.Options) {
		return unsupported("semantic mode count does not match syntax mode count")
	}

	modes := make([]game.Mode, 0, len(ctx.content.Modes))
	// modeReasons collects every mode that fails to lower. Modes are independent,
	// so all of their blockers are reported; the first failing mode stays primary,
	// matching the reason a first-failure bail used to return.
	var modeReasons []shared.Diagnostic
	modeUnsupported := func(detail string) shared.Diagnostic {
		return *contentDiagnostic(ctx, "unsupported ability modes", detail)
	}
	for i, mode := range ctx.content.Modes {
		syntaxMode := syntax.Modal.Options[i]
		bodySyntax := parser.Ability{
			Kind:                   parser.AbilitySpell,
			Span:                   syntaxMode.Body.Span,
			Text:                   syntaxMode.Body.Text,
			Tokens:                 syntaxMode.Body.Tokens,
			Sentences:              syntaxMode.Sentences,
			ConditionBoundaries:    syntaxMode.ConditionBoundaries,
			EventHistoryConditions: syntaxMode.EventHistoryConditions,
			ConditionClauses:       syntaxMode.ConditionClauses,
			ConditionSegments:      syntaxMode.ConditionSegments,
			SemanticReferences:     syntaxMode.SemanticReferences,
			SemanticKeywords:       syntaxMode.SemanticKeywords,
			Reminders:              syntaxMode.Reminders,
			Quoted:                 syntaxMode.Quoted,
			Atoms:                  syntaxMode.Atoms,
		}
		content, diagnostic := lowerAbilityContent(cardName, ctx.enclosingKind, mode.Content, false, &bodySyntax)
		if diagnostic != nil {
			primary := *diagnostic
			additional := primary.Additional
			primary.Additional = nil
			primary.Detail = fmt.Sprintf("mode %d: %s", i+1, primary.Detail)
			modeReasons = append(modeReasons, primary)
			modeReasons = append(modeReasons, additional...)
			continue
		}
		if content.IsModal() || len(content.Modes) != 1 {
			modeReasons = append(modeReasons, modeUnsupported("mode lowering produced unexpected modal content"))
			continue
		}
		if !modalOptionCompletelyRecognized(mode.Content, &syntaxMode) {
			modeReasons = append(modeReasons, modeUnsupported("a modal option contains rules text without complete executable semantics"))
			continue
		}
		loweredMode := content.Modes[0]
		loweredMode.Text = mode.Text
		if modal.Spree {
			if len(mode.SpreeCost) == 0 {
				modeReasons = append(modeReasons, modeUnsupported("a Spree option is missing its additional cost"))
				continue
			}
			loweredMode.Cost = opt.Val(slices.Clone(mode.SpreeCost))
		} else if len(mode.SpreeCost) != 0 {
			modeReasons = append(modeReasons, modeUnsupported("a non-Spree modal option carries an additional cost"))
			continue
		}
		modes = append(modes, loweredMode)
	}
	if len(modeReasons) > 0 {
		return game.AbilityContent{}, combineReasons(modeReasons)
	}
	result := game.AbilityContent{
		Modes:           modes,
		MinModes:        minModes,
		MaxModes:        maxModes,
		ModeChoiceBonus: bonus,
		RandomModes:     randomModes,
	}
	if modal.Escalate {
		if len(modal.EscalateCost) == 0 {
			return unsupported("an Escalate modal is missing its escalate cost")
		}
		result.EscalateCost = opt.Val(slices.Clone(modal.EscalateCost))
	}
	if labeledModal(ctx.content.Modes) && !exactConnectionModes(ctx.content.Modes, result) {
		return unsupported("the labeled modal options do not match the supported exact mode vocabulary and bodies")
	}
	return result, nil
}

func modalOptionCompletelyRecognized(content compiler.AbilityContent, syntax *parser.Mode) bool {
	var spans []shared.Span
	if syntax.Label != nil {
		spans = append(spans, syntax.Label.Span, syntax.Label.SeparatorSpan)
	}
	if syntax.FlavorSpan != (shared.Span{}) {
		spans = append(spans, syntax.FlavorSpan, syntax.FlavorSeparatorSpan)
	}
	if syntax.SpreeCost != nil {
		spans = append(spans, syntax.SpreeCost.Span, syntax.SpreeCost.SeparatorSpan)
	}
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

func labeledModal(modes []compiler.CompiledMode) bool {
	for _, mode := range modes {
		if mode.Label != compiler.CompiledModeLabelNone {
			return true
		}
	}
	return false
}

func exactConnectionModeLabels(modes []compiler.CompiledMode) bool {
	if len(modes) != 3 {
		return false
	}
	want := [...]compiler.CompiledModeLabel{
		compiler.CompiledModeLabelSellContraband,
		compiler.CompiledModeLabelBuyInformation,
		compiler.CompiledModeLabelHireMercenary,
	}
	for i := range want {
		if modes[i].Label != want[i] {
			return false
		}
	}
	return true
}

func exactConnectionModes(compiled []compiler.CompiledMode, lowered game.AbilityContent) bool {
	if !exactConnectionModeLabels(compiled) ||
		lowered.MinModes != 1 || lowered.MaxModes != 3 ||
		lowered.AllowDuplicateModes || len(lowered.Modes) != 3 {
		return false
	}
	return exactCreateTokenLoseLifeMode(lowered.Modes[0], types.Treasure, 1, false) &&
		exactDrawLoseLifeMode(lowered.Modes[1], 2) &&
		exactCreateTokenLoseLifeMode(lowered.Modes[2], types.Shapeshifter, 3, true)
}

func exactDrawLoseLifeMode(mode game.Mode, life int) bool {
	if len(mode.Targets) != 0 || len(mode.Sequence) != 2 {
		return false
	}
	draw, drawOK := mode.Sequence[0].Primitive.(game.Draw)
	return drawOK &&
		draw.Amount.Value() == 1 &&
		draw.Player == game.ControllerReference() &&
		exactInstructionControllerLifeLoss(&mode.Sequence[1], life)
}

func exactCreateTokenLoseLifeMode(mode game.Mode, subtype types.Sub, life int, changeling bool) bool {
	if len(mode.Targets) != 0 || len(mode.Sequence) != 2 {
		return false
	}
	create, ok := mode.Sequence[0].Primitive.(game.CreateToken)
	if !ok || create.Amount.Value() != 1 || create.Recipient.Exists ||
		create.EntryTapped || create.EntryAttacking {
		return false
	}
	def, ok := create.Source.TokenDefRef()
	if !ok || !def.HasSubtype(subtype) {
		return false
	}
	var expected *game.CardDef
	if changeling {
		expected = &game.CardDef{CardFace: game.CardFace{
			Name:            string(types.Shapeshifter),
			Types:           []types.Card{types.Creature},
			Subtypes:        []types.Sub{types.Shapeshifter},
			Power:           opt.Val(game.PT{Value: 3}),
			Toughness:       opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{game.ChangelingStaticBody},
		}}
	} else {
		expected = treasureTokenDef()
	}
	if !reflect.DeepEqual(def, expected) {
		return false
	}
	return exactInstructionControllerLifeLoss(&mode.Sequence[1], life)
}

func exactInstructionControllerLifeLoss(instruction *game.Instruction, amount int) bool {
	lose, ok := instruction.Primitive.(game.LoseLife)
	return ok &&
		lose.Amount.Value() == amount &&
		lose.Player == game.ControllerReference() &&
		lose.PlayerGroup == (game.PlayerGroupReference{})
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
		CostModifiers:       shell.costModifiers,
		ZoneOfFunction:      shell.zoneOfFunction,
		Timing:              shell.timing,
		ActivationCondition: shell.activationCondition,
		Content:             shell.content,
	}
	return result, nil
}

func prepareActivationCondition(ability *compiler.CompiledAbility, syntax *parser.Ability) (opt.V[game.Condition], bool) {
	if ability.ExactSequence == compiler.ExactSequenceConditionalLookAtTopBattlefield {
		// The conditional look-at-top battlefield body keeps its "if it's a
		// <type> card" characteristic gate in the body, where the dedicated
		// exact-sequence lowering consumes it as a typed CardCondition. Leave it
		// in place instead of extracting it as an activation gate.
		*syntax = syntaxWithoutAbilityWord(syntax)
		return opt.V[game.Condition]{}, true
	}
	if len(ability.Content.Conditions) == 0 {
		*syntax = syntaxWithoutAbilityWord(syntax)
		return opt.V[game.Condition]{}, true
	}
	if activationConditionOwnedByBody(ability.Content) {
		// A resolution-time "unless its controller pays" tax ("Counter target
		// spell unless its controller pays {1}.") is part of the body effect, not
		// an activation gate. Leave it in the body so the counter-unless-pays
		// content lowerer reads it, instead of extracting and rejecting it as an
		// unsupported activation condition.
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
		if recognized, ok := recognizeConditionalDestination(ability.Content); ok && recognized.search != nil {
			// The conditional-destination body keeps its gate and else-marker
			// conditions in the body, where the dedicated content lowerer reads
			// the gate. Leave them in place instead of failing closed on the
			// two-condition shape.
			return opt.V[game.Condition]{}, true
		}
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
	// References bound inside the extracted condition (for example the "this
	// creature" of "Activate only if this creature is attacking") belong to the
	// activation gate, not the resolving body. Drop them so body lowerers that
	// reject stray references — like the fixed-card draw lowerer — see only the
	// references their effect actually uses.
	ability.Content.References = slices.DeleteFunc(append([]compiler.CompiledReference(nil), ability.Content.References...), func(reference compiler.CompiledReference) bool {
		return spanCovered(reference.Span, conditionSpan)
	})
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

// activationConditionOwnedByBody reports whether an activated ability's single
// condition is a body-level gate that the ordered-sequence lowerer consumes
// directly, rather than an activation gate. Recognized forms:
//   - "unless its controller pays" tax (counter-unless-pays)
//   - "If <source object matches>, <effect>" conditional body rider (e.g.
//     depletion taplands: "If there are no depletion counters on this land,
//     sacrifice it.")
func activationConditionOwnedByBody(content compiler.AbilityContent) bool {
	if len(content.Conditions) != 1 {
		return false
	}
	condition := content.Conditions[0]
	if condition.Intervening || condition.Resolving {
		return false
	}
	// "Unless its controller pays" body tax.
	if condition.Kind == compiler.ConditionUnless &&
		condition.Predicate == compiler.ConditionPredicateTargetControllerDoesNotPay {
		return true
	}
	// Body-level "If <gate>, <effect>" rider (e.g. "If there are no depletion
	// counters on this land, sacrifice it."). The condition gates a body effect
	// whose span follows or contains the condition, so
	// prepareActivationCondition's span-precedes-all-effects test would reject it
	// as an activation gate. Recognize it by predicate and binding: source object
	// match negated (the "no counters" wording) within an "If" clause is always a
	// body rider, never an activation restriction.
	if condition.Kind == compiler.ConditionIf &&
		condition.Predicate == compiler.ConditionPredicateObjectMatches &&
		condition.ObjectBinding == compiler.ReferenceBindingSource &&
		condition.Negated {
		return true
	}
	return false
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
	case compiler.ActivationTimingNone, compiler.ActivationTimingInstant:
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
	case compiler.ActivationTimingDuringYourTurnBeforeAttackers:
		return game.DuringYourTurnBeforeAttackers, true
	default:
		return game.NoTimingRestriction, false
	}
}
