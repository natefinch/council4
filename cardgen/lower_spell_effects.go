package cardgen

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func lowerSelfCardGraveyardReturn(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 1 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectReturn ||
		!effect.Exact ||
		effect.FromZone != zone.Graveyard ||
		effect.Negated ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		effect.UnderYourControl ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		!selfCardGraveyardReturnReferences(ctx.content.References) {
		return game.AbilityContent{}, false
	}
	sourceCard, ok := lowerCardReference(ctx.content.References[0], referenceLoweringContext{AllowSource: true})
	if !ok {
		return game.AbilityContent{}, false
	}
	switch effect.ToZone {
	case zone.Hand:
		if effect.EntersTapped || effect.CounterKindKnown || effect.Amount.Known {
			return game.AbilityContent{}, false
		}
		return game.Mode{Sequence: []game.Instruction{{Primitive: game.MoveCard{
			Card:        sourceCard,
			FromZone:    zone.Graveyard,
			Destination: zone.Hand,
		}}}}.Ability(), true
	case zone.Battlefield:
		if effect.CounterKindKnown &&
			(effect.CounterKind != counter.PlusOnePlusOne || !effect.Amount.Known || effect.Amount.Value < 1) {
			return game.AbilityContent{}, false
		}
		put := game.PutOnBattlefield{
			Source:      game.CardBattlefieldSource(sourceCard),
			EntryTapped: effect.EntersTapped,
		}
		if effect.CounterKindKnown {
			put.EntryCounters = []game.CounterPlacement{{Kind: counter.PlusOnePlusOne, Amount: effect.Amount.Value}}
		}
		return game.Mode{Sequence: []game.Instruction{{Primitive: put}}}.Ability(), true
	default:
		return game.AbilityContent{}, false
	}
}

func selfCardGraveyardReturnReferences(references []compiler.CompiledReference) bool {
	return referencesBindTo(references, compiler.ReferenceBindingSource, 0)
}

func lowerTargetedGraveyardReturn(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Targets) != 1 ||
		len(ctx.content.Effects) != 1 ||
		!ctx.content.Effects[0].Exact ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		ctx.content.Effects[0].FromZone != zone.Graveyard {
		return game.AbilityContent{}, false
	}
	targetSpec, ok := cardInZoneTargetSpec(ctx.content.Targets[0], zone.Graveyard)
	if !ok {
		return game.AbilityContent{}, false
	}
	sequence := make([]game.Instruction, 0, targetSpec.MaxTargets)
	switch ctx.content.Effects[0].ToZone {
	case zone.Hand:
		for i := range targetSpec.MaxTargets {
			sequence = append(sequence, game.Instruction{Primitive: game.MoveCard{
				Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: i},
				FromZone:    zone.Graveyard,
				Destination: zone.Hand,
			}})
		}
		return game.Mode{
			Targets:  []game.TargetSpec{targetSpec},
			Sequence: sequence,
		}.Ability(), true
	case zone.Library:
		if ctx.content.Effects[0].Destination != parser.EffectDestinationTop &&
			ctx.content.Effects[0].Destination != parser.EffectDestinationBottom {
			return game.AbilityContent{}, false
		}
		destinationBottom := ctx.content.Effects[0].Destination == parser.EffectDestinationBottom
		for i := range targetSpec.MaxTargets {
			sequence = append(sequence, game.Instruction{Primitive: game.MoveCard{
				Card:              game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: i},
				FromZone:          zone.Graveyard,
				Destination:       zone.Library,
				DestinationBottom: destinationBottom,
			}})
		}
		return game.Mode{
			Targets:  []game.TargetSpec{targetSpec},
			Sequence: sequence,
		}.Ability(), true
	case zone.Battlefield:
		for i := range targetSpec.MaxTargets {
			put := game.PutOnBattlefield{
				Source:      game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: i}),
				EntryTapped: ctx.content.Effects[0].EntersTapped,
			}
			if ctx.content.Effects[0].UnderYourControl {
				put.Recipient = opt.Val(game.ControllerReference())
			}
			sequence = append(sequence, game.Instruction{Primitive: put})
		}
		return game.Mode{
			Targets:  []game.TargetSpec{targetSpec},
			Sequence: sequence,
		}.Ability(), true
	default:
		return game.AbilityContent{}, false
	}
}

func cardInZoneTargetSpec(target compiler.CompiledTarget, targetZone zone.Type) (game.TargetSpec, bool) {
	if target.Cardinality.Min < 0 || target.Cardinality.Max < target.Cardinality.Min ||
		target.Cardinality.Max == 0 ||
		target.Selector.Zone != targetZone ||
		target.Selector.Other ||
		target.Selector.Attacking || target.Selector.Blocking ||
		target.Selector.Tapped || target.Selector.Untapped {
		return game.TargetSpec{}, false
	}
	selection, ok := cardSelectionForSelector(target.Selector)
	if !ok {
		return game.TargetSpec{}, false
	}
	return game.TargetSpec{
		MinTargets: target.Cardinality.Min,
		MaxTargets: target.Cardinality.Max,
		Constraint: lowerFirst(target.Text),
		Allow:      game.TargetAllowCard,
		TargetZone: targetZone,
		Selection:  opt.Val(selection),
	}, true
}

func cardSelectionForSelector(selector compiler.CompiledSelector) (game.Selection, bool) {
	selection := game.Selection{
		RequiredTypesAny: slices.Clone(selector.RequiredTypesAny()),
		ExcludedTypes:    slices.Clone(selector.ExcludedTypes()),
		Supertypes:       slices.Clone(selector.Supertypes()),
		ColorsAny:        slices.Clone(selector.ColorsAny()),
		ExcludedColors:   slices.Clone(selector.ExcludedColors()),
		SubtypesAny:      slices.Clone(selector.SubtypesAny()),
	}
	switch selector.Kind {
	case compiler.SelectorCard:
	case compiler.SelectorArtifact:
		selection.RequiredTypes = []types.Card{types.Artifact}
	case compiler.SelectorCreature:
		selection.RequiredTypes = []types.Card{types.Creature}
	case compiler.SelectorEnchantment:
		selection.RequiredTypes = []types.Card{types.Enchantment}
	case compiler.SelectorLand:
		selection.RequiredTypes = []types.Card{types.Land}
	case compiler.SelectorPlaneswalker:
		selection.RequiredTypes = []types.Card{types.Planeswalker}
	case compiler.SelectorPermanent:
		selection.RequiredTypesAny = []types.Card{types.Artifact, types.Creature, types.Enchantment, types.Land, types.Planeswalker, types.Battle}
	default:
		return game.Selection{}, false
	}
	switch selector.Controller {
	case compiler.ControllerAny:
	case compiler.ControllerYou:
		selection.Controller = game.ControllerYou
	case compiler.ControllerOpponent:
		selection.Controller = game.ControllerOpponent
	default:
		return game.Selection{}, false
	}
	if selector.Keyword != parser.KeywordUnknown {
		keyword, ok := runtimeKeyword(selector.Keyword)
		if !ok {
			return game.Selection{}, false
		}
		selection.Keyword = keyword
	}
	if selector.MatchManaValue {
		selection.ManaValue = opt.Val(selector.ManaValue)
	}
	if selector.MatchPower || selector.MatchToughness {
		return game.Selection{}, false
	}
	return selection, true
}

func lowerCounterPlacementSpell(
	ctx contentCtx,
) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if len(ctx.content.Targets) == 0 &&
		len(ctx.content.References) == 1 &&
		(ctx.content.References[0].Binding == compiler.ReferenceBindingSource ||
			ctx.content.References[0].Binding == compiler.ReferenceBindingTarget) {
		return lowerReferencedCounterPlacement(ctx)
	}
	if len(ctx.content.Targets) != 1 ||
		ctx.content.Targets[0].Cardinality.Min != 1 ||
		ctx.content.Targets[0].Cardinality.Max != 1 ||
		(effect.Amount.Known && effect.Amount.Value <= 0) ||
		!effect.Amount.Known && !effect.Amount.VariableX && effect.Amount.DynamicKind == compiler.DynamicAmountNone ||
		!effect.CounterKindKnown ||
		!compiler.CounterKindPlacementSupported(effect.CounterKind) ||
		!effect.Exact ||
		effect.Negated ||
		effect.Context != parser.EffectContextController ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, unsupportedCounterPlacementDiagnostic(ctx)
	}

	kind := effect.CounterKind
	var target game.TargetSpec
	var primitive game.Primitive
	if kind.PlayerOnly() {
		var ok bool
		target, ok = playerTargetSpec(ctx.content.Targets[0])
		if !ok {
			return game.AbilityContent{}, unsupportedCounterPlacementDiagnostic(ctx)
		}
	} else {
		var ok bool
		target, ok = permanentTargetSpec(ctx.content.Targets[0])
		if !ok {
			return game.AbilityContent{}, unsupportedCounterPlacementDiagnostic(ctx)
		}
	}

	amount := game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX})
	switch {
	case effect.Amount.Known:
		amount = game.Fixed(effect.Amount.Value)
		if len(ctx.content.References) != 0 {
			return game.AbilityContent{}, unsupportedCounterPlacementDiagnostic(ctx)
		}
	case effect.Amount.VariableX:
		if len(ctx.content.References) != 0 {
			return game.AbilityContent{}, unsupportedCounterPlacementDiagnostic(ctx)
		}
	case effect.Amount.DynamicKind != compiler.DynamicAmountNone:
		dynamic, supported := lowerDynamicAmount(effect.Amount, game.SourcePermanentReference())
		if !supported ||
			!exactDynamicAmountReference(effect.Amount, ctx.content.References) {
			return game.AbilityContent{}, unsupportedCounterPlacementDiagnostic(ctx)
		}
		amount = game.Dynamic(dynamic)
	default:
	}
	if kind.PlayerOnly() {
		primitive = game.AddPlayerCounter{
			Amount:      amount,
			Player:      game.TargetPlayerReference(0),
			CounterKind: kind,
		}
	} else {
		primitive = game.AddCounter{
			Amount:      amount,
			Object:      game.TargetPermanentReference(0),
			CounterKind: kind,
		}
	}
	return game.Mode{
		Targets: []game.TargetSpec{target},
		Sequence: []game.Instruction{{
			Primitive: primitive,
		}},
	}.Ability(), nil
}

// lowerReferencedCounterPlacement lowers an exact fixed counter placement whose
// object is a single referenced permanent: the source permanent itself ("Put a
// +1/+1 counter on this creature.") or a prior clause's target referenced by "it"
// in an ordered sequence ("… Put a +1/+1 counter on it."). The object lowers to
// game.SourcePermanentReference() or a target reference accordingly. Restricted to
// fixed positive amounts of a supported permanent counter kind.
func lowerReferencedCounterPlacement(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		!effect.Amount.Known || effect.Amount.Value <= 0 ||
		!effect.CounterKindKnown ||
		!compiler.CounterKindPlacementSupported(effect.CounterKind) ||
		effect.CounterKind.PlayerOnly() ||
		!effect.Exact ||
		effect.Negated ||
		effect.Context != parser.EffectContextController {
		return game.AbilityContent{}, unsupportedCounterPlacementDiagnostic(ctx)
	}
	object, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{
		AllowSource: true,
		AllowTarget: true,
	})
	if !ok {
		return game.AbilityContent{}, unsupportedCounterPlacementDiagnostic(ctx)
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.AddCounter{
				Amount:      game.Fixed(effect.Amount.Value),
				Object:      object,
				CounterKind: effect.CounterKind,
			},
		}},
	}.Ability(), nil
}

func unsupportedCounterPlacementDiagnostic(ctx contentCtx) *shared.Diagnostic {
	return contentDiagnostic(
		ctx,
		"unsupported counter placement",
		"the executable source backend supports exact recognized counter placement on one valid target",
	)
}

func exactDynamicAmountReference(
	amount compiler.CompiledAmount,
	references []compiler.CompiledReference,
) bool {
	if amount.DynamicKind != compiler.DynamicAmountSourcePower {
		return len(references) == 0
	}
	if len(references) != 1 || references[0].Span != amount.ReferenceSpan {
		return false
	}
	return references[0].Binding == compiler.ReferenceBindingSource
}

func textWithoutDelimited(text string, span shared.Span, groups []parser.Delimited) string {
	var result strings.Builder
	cursor := span.Start.Offset
	for _, group := range groups {
		if group.Span.Start.Offset < cursor ||
			group.Span.End.Offset > span.End.Offset {
			continue
		}
		start := group.Span.Start.Offset - span.Start.Offset
		end := cursor - span.Start.Offset
		_, _ = result.WriteString(text[end:start])
		cursor = group.Span.End.Offset
	}
	_, _ = result.WriteString(text[cursor-span.Start.Offset:])
	return strings.TrimSpace(result.String())
}

func lowerFightSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	if len(ctx.content.Targets) != 2 ||
		ctx.content.Targets[0].Cardinality != (compiler.TargetCardinality{Min: 1, Max: 1}) ||
		ctx.content.Targets[1].Cardinality != (compiler.TargetCardinality{Min: 1, Max: 1}) ||
		ctx.content.Effects[0].Negated ||
		!ctx.content.Effects[0].Exact ||
		ctx.content.Effects[0].Selector.Another ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 ||
		(ctx.content.Effects[0].Context != parser.EffectContextTarget &&
			ctx.content.Effects[0].Context != parser.EffectContextPriorSubject) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported fight spell",
			"the executable source backend supports only exact fights between two target creatures",
		)
	}
	first, firstOK := fightCreatureTargetSpec(ctx.content.Targets[0])
	second, secondOK := fightCreatureTargetSpec(ctx.content.Targets[1])
	if !firstOK || !secondOK {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported fight spell",
			"the executable source backend supports only exact fights between two target creatures",
		)
	}
	return game.Mode{
		Targets: []game.TargetSpec{first, second},
		Sequence: []game.Instruction{{
			Primitive: game.Fight{
				Object:        game.TargetPermanentReference(0),
				RelatedObject: game.TargetPermanentReference(1),
			},
		}},
	}.Ability(), nil
}

func fightCreatureTargetSpec(target compiler.CompiledTarget) (game.TargetSpec, bool) {
	if !targetCardinalityIsOne(target) ||
		target.Selector.Kind != compiler.SelectorCreature ||
		target.Selector.Another ||
		target.Selector.Other ||
		target.Selector.Attacking ||
		target.Selector.Blocking ||
		target.Selector.Tapped ||
		target.Selector.Untapped {
		return game.TargetSpec{}, false
	}
	spec := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Constraint: target.Text,
		Allow:      game.TargetAllowPermanent,
		Predicate: game.TargetPredicate{
			PermanentTypes: []types.Card{types.Creature},
		},
	}
	switch target.Selector.Controller {
	case compiler.ControllerAny:
	case compiler.ControllerYou:
		spec.Predicate.Controller = game.ControllerYou
	case compiler.ControllerOpponent:
		spec.Predicate.Controller = game.ControllerOpponent
	case compiler.ControllerNotYou:
		spec.Predicate.Controller = game.ControllerNotYou
	default:
		return game.TargetSpec{}, false
	}
	return spec, true
}

func lowerInvestigateSpell(
	ctx contentCtx,
	syntax *parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	return lowerExactPrimitiveSpell(
		ctx,
		syntax,
		"investigate",
		func(amount game.Quantity) game.Primitive {
			return game.Investigate{Amount: amount}
		},
	)
}

func lowerExploreSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	unsupportedExplore := contentDiagnostic(
		ctx,
		"unsupported explore spell",
		"the executable source backend supports only the source permanent pattern \"it explores\"",
	)
	if ctx.content.Effects[0].Negated ||
		!ctx.content.Effects[0].Exact ||
		ctx.content.Effects[0].Context != parser.EffectContextReferencedObject ||
		len(ctx.content.References) != 1 ||
		(ctx.content.References[0].Binding != compiler.ReferenceBindingSource &&
			ctx.content.References[0].Binding != compiler.ReferenceBindingEventPermanent) {
		return game.AbilityContent{}, unsupportedExplore
	}
	// Reference validated as "it" pronoun — clear before the fail-closed check.
	consumed := ctx
	consumed.content.References = nil
	if consumed.content.Unconsumed() {
		return game.AbilityContent{}, unsupportedExplore
	}
	object, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{
		AllowSource: true,
		AllowEvent:  true,
	})
	if !ok {
		return game.AbilityContent{}, unsupportedExplore
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.Explore{Creature: object},
	}}}.Ability(), nil
}

func lowerManifestSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if ctx.content.Effects[0].Negated ||
		!effect.Exact ||
		ctx.content.Unconsumed() ||
		len(ctx.content.References) != 0 {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported manifest spell",
			"the executable source backend supports only \"manifest the top card of your library\" and manifest dread",
		)
	}
	if effect.Kind == compiler.EffectManifestDread {
		return manifestDreadAbility(), nil
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.Manifest{},
	}}}.Ability(), nil
}

func manifestDreadAbility() game.AbilityContent {
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.Manifest{Dread: true},
	}}}.Ability()
}

func typedManifestDreadSequence(content compiler.AbilityContent) bool {
	if len(content.Effects) != 3 ||
		len(content.Targets) != 0 ||
		len(content.Conditions) != 0 ||
		len(content.Keywords) != 0 ||
		len(content.Modes) != 0 ||
		len(content.References) != 1 {
		return false
	}
	look := content.Effects[0]
	battlefield := content.Effects[1]
	graveyard := content.Effects[2]
	reference := content.References[0]
	return look.Kind == compiler.EffectManifestDread &&
		look.Amount.Known && look.Amount.Value == 2 &&
		battlefield.Kind == compiler.EffectPut &&
		battlefield.Amount.Known && battlefield.Amount.Value == 1 &&
		battlefield.ToZone == zone.Battlefield &&
		graveyard.Kind == compiler.EffectPut &&
		graveyard.Selector.Other &&
		graveyard.ToZone == zone.Graveyard &&
		reference.Binding == compiler.ReferenceBindingPriorInstructionResult &&
		reference.PriorInstruction == 0
}

func lowerExactPrimitiveSpell(
	ctx contentCtx,
	_ *parser.Ability,
	verb string,
	primitiveFactory func(game.Quantity) game.Primitive,
) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if effect.Negated ||
		!effect.Exact ||
		effect.Context != parser.EffectContextController ||
		!effect.Amount.Known ||
		effect.Amount.Value < 1 ||
		ctx.content.Unconsumed() ||
		len(ctx.content.References) != 0 {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported "+verb+" spell",
			"the executable source backend supports only exact "+verb,
		)
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: primitiveFactory(game.Fixed(effect.Amount.Value)),
	}}}.Ability(), nil
}
