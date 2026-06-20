package cardgen

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

type loweredActivationShell struct {
	text                string
	manaCost            opt.V[cost.Mana]
	additionalCosts     []cost.Additional
	costModifiers       []game.CostModifier
	zoneOfFunction      zone.Type
	timing              game.TimingRestriction
	activationCondition opt.V[game.Condition]
	semanticContent     compiler.AbilityContent
	content             game.AbilityContent
}

func lowerActivationShell(
	cardName string,
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (loweredActivationShell, *shared.Diagnostic) {
	original := ability
	if ability.Cost == nil || len(ability.Cost.Components) == 0 {
		return loweredActivationShell{}, activationDiagnostic(
			original,
			"unsupported activation cost",
			"the executable source backend requires an exact typed activation cost",
		)
	}
	if !rulesFreeAbilityWordLabel(ability.AbilityWord) {
		return loweredActivationShell{}, activationDiagnostic(
			original,
			"unsupported activation ability word",
			"the executable source backend cannot lower this activated ability word",
		)
	}

	activationCondition, ok := prepareActivationCondition(&ability, syntax)
	if !ok {
		return loweredActivationShell{}, activationDiagnostic(
			original,
			"unsupported activation condition",
			"the executable source backend cannot lower every activation condition",
		)
	}
	if !activationCostReferencesSupported(ability.Content.References, ability.Cost) {
		return loweredActivationShell{}, activationDiagnostic(
			original,
			"unsupported activation references",
			"the executable source backend cannot lower every bound reference in this activation cost",
		)
	}
	manaCost, additionalCosts, ok := lowerActivationCostComponents(cardName, ability.Cost)
	if !ok {
		return loweredActivationShell{}, activationDiagnostic(
			original,
			"unsupported activation cost",
			"the executable source backend cannot lower every typed activation cost component",
		)
	}
	timing, ok := lowerActivationTiming(ability.ActivationTiming)
	if !ok {
		return loweredActivationShell{}, activationDiagnostic(
			original,
			"unsupported activation timing",
			"the executable source backend cannot lower this activation timing restriction",
		)
	}
	zoneOfFunction, ok := lowerActivationZone(ability.ActivationZone)
	if !ok {
		return loweredActivationShell{}, activationDiagnostic(
			original,
			"unsupported activation zone",
			"the executable source backend cannot lower this activation zone of function",
		)
	}
	if !channelActivationSupported(ability, zoneOfFunction, additionalCosts) {
		return loweredActivationShell{}, activationDiagnostic(
			original,
			"unsupported Channel ability",
			"the executable source backend requires Channel to discard itself from hand and supports only exact typed search semantics",
		)
	}
	// A graveyard ability has no battlefield source permanent, so the runtime
	// evaluates its activation condition with a nil source. Event-history
	// patterns resolve their controller-relative filters ("you", "an opponent")
	// from that source, so such a condition would fail closed forever and make
	// the ability impossible to activate. Fail closed at lowering instead of
	// emitting a permanently dead ability.
	if zoneOfFunction == zone.Graveyard && activationCondition.Exists && activationCondition.Val.EventHistory.Exists {
		return loweredActivationShell{}, activationDiagnostic(
			original,
			"unsupported activation condition",
			"the executable source backend cannot evaluate an event-history activation condition for a graveyard ability that has no battlefield source",
		)
	}

	bodyTokens := append([]shared.Token(nil), parser.TokensInSpan(syntax.Tokens, syntax.BodySpan)...)
	if len(bodyTokens) == 0 {
		return loweredActivationShell{}, activationDiagnostic(
			original,
			"unsupported activation structure",
			"the executable source backend cannot identify the activated ability body",
		)
	}
	if ability.ActivationTiming != compiler.ActivationTimingNone {
		bodyTokens = slices.DeleteFunc(bodyTokens, func(token shared.Token) bool {
			return spanCovered(token.Span, []shared.Span{ability.ActivationTimingSpan})
		})
	}
	if ability.SourceAbilityCostReduction != nil {
		bodyTokens = slices.DeleteFunc(bodyTokens, func(token shared.Token) bool {
			return spanCovered(token.Span, []shared.Span{ability.SourceAbilityCostReduction.Span})
		})
	}
	if len(bodyTokens) == 0 {
		return loweredActivationShell{}, activationDiagnostic(
			original,
			"unsupported activation structure",
			"the executable source backend cannot identify nonempty activated ability content",
		)
	}

	bodyContent := ability.Content
	bodyContent.References = bodyReferences(ability.Content.References, ability.Cost.Span)
	bodyContent.References = slices.DeleteFunc(bodyContent.References, func(reference compiler.CompiledReference) bool {
		return slices.ContainsFunc(bodyContent.Effects, func(effect compiler.CompiledEffect) bool {
			return effect.Kind == compiler.EffectManaSpendRider && spanCovered(reference.Span, []shared.Span{effect.Span})
		})
	})
	if !activationReferencesSupported(bodyContent) {
		return loweredActivationShell{}, activationDiagnostic(
			original,
			"unsupported activation references",
			"the executable source backend cannot lower every bound reference in this activated ability",
		)
	}
	bodySpan := shared.Span{
		Start: bodyTokens[0].Span.Start,
		End:   bodyTokens[len(bodyTokens)-1].Span.End,
	}
	bodyText := strings.TrimSpace(ability.Text[bodySpan.Start.Offset-ability.Span.Start.Offset : bodySpan.End.Offset-ability.Span.Start.Offset])
	bodyContent.Keywords = keywordsWithinSpan(ability.Content.Keywords, bodySpan)
	if len(bodyContent.Keywords) != len(ability.Content.Keywords) {
		return loweredActivationShell{}, activationDiagnostic(
			original,
			"unsupported activation structure",
			"the executable source backend cannot assign every activated ability keyword to its body",
		)
	}
	bodySyntax := parser.Ability{
		Span:      bodySpan,
		Text:      bodyText,
		Tokens:    bodyTokens,
		Reminders: syntax.Reminders,
		Quoted:    syntax.Quoted,
		Modal:     syntax.Modal,
		Atoms:     syntax.Atoms,
	}
	content, diagnostic := lowerAbilityContent(cardName, bodyContent, false, &bodySyntax)
	if diagnostic != nil {
		if diagnostic.Summary == "unsupported ability modes" {
			diagnostic.Summary = "unsupported activation modes"
		}
		return loweredActivationShell{}, diagnostic
	}

	result := loweredActivationShell{
		text:                original.Text,
		additionalCosts:     additionalCosts,
		zoneOfFunction:      zoneOfFunction,
		timing:              timing,
		activationCondition: activationCondition,
		semanticContent:     bodyContent,
		content:             content,
	}
	if reduction := ability.SourceAbilityCostReduction; reduction != nil {
		selection, ok := dynamicAmountSelection(reduction.CountSelection)
		if !ok {
			return loweredActivationShell{}, activationDiagnostic(
				original,
				"unsupported source-ability cost reduction",
				"the counted battlefield objects are not representable by the runtime selection vocabulary",
			)
		}
		result.costModifiers = []game.CostModifier{{
			Kind:               game.CostModifierAbility,
			PerObjectReduction: reduction.Amount,
			CountSelection:     selection,
		}}
	}
	if manaCost != nil {
		result.manaCost = opt.Val(manaCost)
	}
	return result, nil
}

func channelActivationSupported(ability compiler.CompiledAbility, functionZone zone.Type, additionalCosts []cost.Additional) bool {
	if !strings.EqualFold(ability.AbilityWord, "Channel") {
		return true
	}
	if functionZone != zone.Hand ||
		len(additionalCosts) != 1 ||
		additionalCosts[0].Kind != cost.AdditionalDiscard ||
		additionalCosts[0].Source != zone.Hand ||
		(additionalCosts[0].Amount != 0 && additionalCosts[0].Amount != 1) {
		return false
	}
	for i := range ability.Content.Effects {
		effect := &ability.Content.Effects[i]
		if effect.Kind == compiler.EffectSearch && !effect.Selector.BasicLandType {
			return false
		}
	}
	return true
}

func activationReferencesSupported(content compiler.AbilityContent) bool {
	for i := range content.Effects {
		if content.Effects[i].Kind == compiler.EffectManifestDread && !content.Effects[i].Exact &&
			len(content.References) != 0 {
			return false
		}
	}
	for _, reference := range content.References {
		if reference.Binding == compiler.ReferenceBindingUnsupported ||
			reference.Binding == compiler.ReferenceBindingAmbiguous {
			return false
		}
	}
	for _, mode := range content.Modes {
		if !activationReferencesSupported(mode.Content) {
			return false
		}
	}
	return true
}

func activationCostReferencesSupported(references []compiler.CompiledReference, compiled *compiler.CompiledCost) bool {
	for _, reference := range references {
		if !spanCovered(reference.Span, []shared.Span{compiled.Span}) ||
			reference.Binding == compiler.ReferenceBindingSource {
			continue
		}
		if !slices.ContainsFunc(compiled.Components, func(component compiler.CostComponent) bool {
			return component.Kind == compiler.CostReturn &&
				spanCovered(reference.Span, []shared.Span{component.Span}) &&
				(reference.Pronoun == compiler.ReferencePronounIts || reference.Pronoun == compiler.ReferencePronounTheir)
		}) {
			return false
		}
	}
	return true
}

func activationDiagnostic(ability compiler.CompiledAbility, summary, detail string) *shared.Diagnostic {
	return executableDiagnostic(ability, summary, detail)
}

func lowerActivationZone(activationZone zone.Type) (zone.Type, bool) {
	switch activationZone {
	case zone.Battlefield, zone.Graveyard, zone.Hand:
		return activationZone, true
	default:
		return zone.None, false
	}
}
