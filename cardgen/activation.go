package cardgen

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
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
	if !isBoastAbilityWord(ability.AbilityWord) &&
		!isMaxSpeedAbilityWord(ability.AbilityWord) &&
		!rulesFreeAbilityWordLabel(ability.AbilityWord) {
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
	if isBoastAbilityWord(ability.AbilityWord) {
		// Boast (CR 702.116) carries its activation restrictions in the keyword
		// itself, not in rules text: it may be activated only if its source
		// attacked this turn and only once each turn. Reject any Boast ability
		// that also declares an explicit timing or condition, or that does not
		// function from the battlefield, so unexpected shapes stay unsupported.
		if timing != game.NoTimingRestriction || activationCondition.Exists || zoneOfFunction != zone.Battlefield {
			return loweredActivationShell{}, activationDiagnostic(
				original,
				"unsupported Boast ability",
				"the executable source backend supports only a battlefield Boast ability with no other activation restriction",
			)
		}
		timing = game.OncePerTurn
		activationCondition = opt.Val(game.BoastActivationCondition())
	}
	if isMaxSpeedAbilityWord(ability.AbilityWord) {
		// The "Max speed" ability word (CR 702.179, the Start your engines! speed
		// subsystem) gates activation on the controller having maximum speed. It
		// imposes no timing of its own, so reject any Max speed ability that also
		// declares an explicit timing or rules-text condition. The
		// ControllerHasMaxSpeed condition is evaluated against the controlling
		// player (not the source permanent), so it functions from the battlefield
		// or from the graveyard (e.g. "Max speed — {3}, Exile this card from your
		// graveyard: Draw a card.").
		if timing != game.NoTimingRestriction || activationCondition.Exists ||
			(zoneOfFunction != zone.Battlefield && zoneOfFunction != zone.Graveyard) {
			return loweredActivationShell{}, activationDiagnostic(
				original,
				"unsupported Max speed ability",
				"the executable source backend supports only a battlefield or graveyard Max speed ability with no other activation restriction",
			)
		}
		activationCondition = opt.Val(game.MaxSpeedActivationCondition())
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
	normalizeExactActivationSelfReturnReferences(&bodyContent)
	if !ability.EvolutionaryLeapRevealUntil &&
		!ability.ProgenitorIconNextFlash &&
		!ability.SelesnyaEulogistPopulate &&
		ability.LifeCharacteristicExchange == nil &&
		!ability.UnlicensedHearseExile &&
		!ability.TemurSabertoothSequence &&
		!activationReferencesSupported(bodyContent) {
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
		CoinFlip:  syntax.CoinFlip,
		Vote:      syntax.Vote,
	}
	content, diagnostic := lowerActivatedBodyContent(cardName, ability, bodyContent, &bodySyntax, bodyText, hasVariableCounterRemovalCost(additionalCosts))
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
			CountSelection:     &selection,
		}}
	}
	if manaCost != nil {
		result.manaCost = opt.Val(manaCost)
	}
	return result, nil
}

func normalizeExactActivationSelfReturnReferences(content *compiler.AbilityContent) {
	if content == nil || len(content.Effects) != 1 ||
		content.Effects[0].Kind != compiler.EffectReturn ||
		!content.Effects[0].Exact ||
		content.Effects[0].ToZone != zone.Hand ||
		len(content.References) == 0 {
		return
	}
	hasSelfName := false
	for i := range content.References {
		reference := &content.References[i]
		switch {
		case reference.Kind == compiler.ReferenceSelfName:
			hasSelfName = true
		case reference.Pronoun == compiler.ReferencePronounIts:
		default:
			return
		}
	}
	if !hasSelfName {
		return
	}
	for i := range content.References {
		content.References[i].Binding = compiler.ReferenceBindingSource
	}
	content.Effects[0].References = slices.Clone(content.References)
}

func isBoastAbilityWord(label string) bool {
	return strings.EqualFold(label, "Boast")
}

func isMaxSpeedAbilityWord(label string) bool {
	return strings.EqualFold(label, "Max speed")
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
		if effect.Kind != compiler.EffectSearch || effect.Selector.BasicLandType {
			continue
		}
		// Beyond Boseiju's opponent-redirected "land card with a basic land type"
		// search, allow the controller's own basic-land fetch — the Kamigawa
		// Channel land cycle (Greater Tanuki et al.): "Search your library for a
		// basic land card, put it onto the battlefield[ tapped], then shuffle."
		// That shape carries no targets and selects a Basic-supertype land; the
		// downstream body lowering validates the full search sequence and fails
		// closed on anything else, so this stays bounded.
		if len(ability.Content.Targets) != 0 ||
			effect.Selector.Kind != compiler.SelectorLand ||
			!slices.Contains(effect.Selector.Supertypes(), types.Basic) {
			return false
		}
	}
	return true
}

func activationReferencesSupported(content compiler.AbilityContent) bool {
	if recognizeCopyLinkedExiledCardCast(content) {
		// The imprint copy/cast lowering ("You may copy the exiled card. If you
		// do, you may cast the copy without paying its mana cost.") identifies the
		// copied card through the source's imprint link and consumes the "its" of
		// "its mana cost" into the free-cast rider, so any residual possessive
		// reference needs no external antecedent resolution.
		return true
	}
	if len(content.Effects) == 1 &&
		content.Effects[0].Kind == compiler.EffectCastAsThoughFlash &&
		content.Effects[0].Exact {
		// The exact timing-permission effect intrinsically owns "you", "spells",
		// and "they"; its lowerer represents the whole sentence as one rule grant.
		return true
	}
	if len(content.Effects) == 1 &&
		content.Effects[0].Kind == compiler.EffectReturn &&
		content.Effects[0].Exact &&
		content.Effects[0].ToZone == zone.Hand &&
		len(content.References) > 0 &&
		!slices.ContainsFunc(content.References, func(reference compiler.CompiledReference) bool {
			return reference.Binding != compiler.ReferenceBindingSource
		}) {
		// Exact self-return effects own both the source-name and possessive
		// references ("Return Arcanis to its owner's hand.").
		return true
	}
	if _, ok := recognizeConditionalDestination(content); ok {
		// The conditional-destination lowering binds the searched card through a
		// linked key and consumes every "it"/"that card" pronoun in the routing
		// sequence, so these references need no external antecedent resolution.
		return true
	}
	for i := range content.Effects {
		if content.Effects[i].Kind == compiler.EffectManifestDread && !content.Effects[i].Exact &&
			len(content.References) != 0 {
			return false
		}
	}
	for _, reference := range content.References {
		if reorderInternalReference(content.Effects, reference) {
			continue
		}
		if returnExiledCardsWithCounterInternalReference(content.Effects, reference) {
			continue
		}
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

// reorderInternalReference reports whether reference is the "them" pronoun
// internal to a self-contained "Look at the top N cards … then put them back in
// any order." (EffectReorderLibraryTop) effect. Such a pronoun refers to the
// looked-at cards and is consumed by the reorder lowering, so it is not a bound
// reference the activation backend must resolve to an external antecedent.
func reorderInternalReference(effects []compiler.CompiledEffect, reference compiler.CompiledReference) bool {
	if reference.Kind != compiler.ReferencePronoun ||
		reference.Pronoun != compiler.ReferencePronounThem {
		return false
	}
	for i := range effects {
		if effects[i].Kind == compiler.EffectReorderLibraryTop &&
			spanCovered(reference.Span, []shared.Span{effects[i].Span}) {
			return true
		}
	}
	return false
}

// returnExiledCardsWithCounterInternalReference reports whether reference is the
// "them" pronoun internal to a self-contained "Put all exiled cards you own with
// <kind> counters on them into your hand." (EffectReturnExiledCardsWithCounter)
// effect. The pronoun refers to the exiled cards the effect returns and is
// consumed by the mass-return lowering, so it is not a bound reference the
// activation backend must resolve to an external antecedent.
func returnExiledCardsWithCounterInternalReference(effects []compiler.CompiledEffect, reference compiler.CompiledReference) bool {
	if reference.Kind != compiler.ReferencePronoun ||
		reference.Pronoun != compiler.ReferencePronounThem {
		return false
	}
	for i := range effects {
		if effects[i].Kind == compiler.EffectReturnExiledCardsWithCounter &&
			spanCovered(reference.Span, []shared.Span{effects[i].Span}) {
			return true
		}
	}
	return false
}

func activationCostReferencesSupported(references []compiler.CompiledReference, compiled *compiler.CompiledCost) bool {
	for _, reference := range references {
		if !spanCovered(reference.Span, []shared.Span{compiled.Span}) ||
			reference.Binding == compiler.ReferenceBindingSource {
			continue
		}
		if costReferenceDenotesSource(reference, compiled) {
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

// costReferenceDenotesSource reports that a reference falls inside a self-sacrifice
// cost component, whose sacrificed object is the ability's own source by
// construction. The "it" of the trailing "sacrifice it" in a combined "Remove N
// counters from this and sacrifice it" cost is such a reference: it denotes the
// source regardless of how the reference binder otherwise classified the bare
// pronoun. The check is restricted to the sacrifice verb, where a self-sacrifice
// can only target the source, so a counter-removal "from it" that denotes a prior
// cost object stays unsupported.
func costReferenceDenotesSource(reference compiler.CompiledReference, compiled *compiler.CompiledCost) bool {
	for _, component := range compiled.Components {
		if component.Kind == compiler.CostSacrifice &&
			component.SourceSelf &&
			spanCovered(reference.Span, []shared.Span{component.Span}) {
			return true
		}
	}
	return false
}

func activationDiagnostic(ability compiler.CompiledAbility, summary, detail string) *shared.Diagnostic {
	return executableDiagnostic(ability, summary, detail)
}

// hasVariableCounterRemovalCost reports whether any additional cost removes a
// player-chosen "one or more" number of counters announced as the ability's X
// (AdditionalRemoveCounter with AmountAtLeastOne). It lets the body lowering
// resolve a "that much"/"that many" anaphor to that announced X.
func hasVariableCounterRemovalCost(additionalCosts []cost.Additional) bool {
	return slices.ContainsFunc(additionalCosts, func(additional cost.Additional) bool {
		return additional.Kind == cost.AdditionalRemoveCounter && additional.AmountAtLeastOne
	})
}

func lowerActivationZone(activationZone zone.Type) (zone.Type, bool) {
	switch activationZone {
	case zone.Battlefield, zone.Graveyard, zone.Hand:
		return activationZone, true
	default:
		return zone.None, false
	}
}
