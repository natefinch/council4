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

// lowerChosenCardGraveyardReturn lowers the non-target "Return a <filter> card
// from your graveyard to your hand" recursion wording, where the returned card
// is chosen from the controller's own graveyard at resolution rather than
// targeted (Takenuma's "creature or planeswalker card", Grapple with the Past,
// ...). The targeted form lowers through lowerTargetedGraveyardReturn instead.
// It produces one game.ReturnFromGraveyard instruction whose Selection carries
// the same card filter the targeted and search paths reconstruct. It is
// card-name-blind and fails closed on any shape it does not fully model — a
// reference or target, a non-graveyard source, a non-hand destination, an
// "enters tapped"/counter/control rider, a selector qualifier it cannot
// express, or an amount other than exactly one card.
func lowerChosenCardGraveyardReturn(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Effects) != 1 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	battlefield := effect.ToZone == zone.Battlefield
	if effect.Kind != compiler.EffectReturn ||
		!effect.Exact ||
		effect.Negated ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		effect.FromZone != zone.Graveyard ||
		(effect.ToZone != zone.Hand && effect.ToZone != zone.Battlefield) ||
		(effect.EntersTapped && !battlefield) ||
		(effect.UnderYourControl && !battlefield) ||
		effect.CounterKindKnown {
		return game.AbilityContent{}, false
	}
	selector := effect.Selector
	if selector.Zone != zone.Graveyard ||
		selector.Controller != compiler.ControllerYou ||
		selector.All ||
		selector.Another ||
		selector.Other ||
		selector.Attacking ||
		selector.Blocking ||
		(selector.Tapped && (!battlefield || !effect.EntersTapped)) ||
		selector.Untapped {
		return game.AbilityContent{}, false
	}
	if !effect.Amount.Known ||
		effect.Amount.RangeKnown ||
		effect.Amount.VariableX ||
		effect.Amount.DynamicKind != 0 ||
		effect.Amount.Value != 1 {
		return game.AbilityContent{}, false
	}
	selection, ok := cardSelectionForSelector(selector)
	if !ok {
		return game.AbilityContent{}, false
	}
	primitive := game.ReturnFromGraveyard{
		Player:    game.ControllerReference(),
		Selection: selection,
		Amount:    game.Fixed(1),
	}
	if battlefield {
		primitive.Destination = zone.Battlefield
		primitive.EntryTapped = effect.EntersTapped
	}
	return game.Mode{Sequence: []game.Instruction{{Primitive: primitive}}}.Ability(), true
}

// lowerMassGraveyardReturn lowers the non-target mass recursion wording "Return
// all <filter> cards from your graveyard to the battlefield" (Brilliant
// Restoration) or "... to your hand", and the all-graveyards reanimation "Put/
// Return all <filter> cards from all graveyards onto the battlefield under your
// control" / "... under their owners' control" (Rise of the Dark Realms, Open
// the Vaults, Planar Birth). The compiler models it as a single non-target
// EffectReturn or EffectPut whose graveyard selector has All set; every matching
// card moves at once with no choice. It is card-name-blind and fails closed on
// any shape it does not fully model — a target or reference, a non-graveyard
// source, a destination other than hand or battlefield, a counter/amount rider,
// or a selector qualifier it cannot express.
//
// The selector's controller scope picks the source graveyards: "your graveyard"
// (ControllerYou) scans only the controller's, carries no ownership rider, and
// reaches both hand and battlefield; the all-graveyards form (ControllerAny)
// scans every player's, requires the battlefield destination and exactly one of
// the "under your control" / "under their owners' control" riders, and chooses
// the entering controller accordingly.
func lowerMassGraveyardReturn(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Targets) != 0 ||
		len(ctx.content.Effects) != 1 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if (effect.Kind != compiler.EffectReturn && effect.Kind != compiler.EffectPut) ||
		effect.Negated ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		effect.FromZone != zone.Graveyard ||
		effect.CounterKindKnown ||
		effect.Amount.Known {
		return game.AbilityContent{}, false
	}
	if effect.ToZone != zone.Hand && effect.ToZone != zone.Battlefield {
		return game.AbilityContent{}, false
	}
	if effect.EntersTapped && effect.ToZone != zone.Battlefield {
		return game.AbilityContent{}, false
	}
	selector := effect.Selector
	// A battlefield entry-tapped rider ("... to the battlefield tapped ...")
	// leaves the entry word inside the selector span, setting selector.Tapped;
	// graveyard cards are never tapped, so that filter is vacuous and is ignored
	// when it coincides with the entry-tapped destination. A genuine tapped
	// filter without entry-tapped still fails closed.
	if !selector.All ||
		selector.Zone != zone.Graveyard ||
		selector.Another ||
		selector.Other ||
		selector.Attacking ||
		selector.Blocking ||
		(selector.Tapped && !effect.EntersTapped) ||
		selector.Untapped {
		return game.AbilityContent{}, false
	}
	var sourceGroup game.PlayerGroupReference
	controlledByOwner := false
	switch selector.Controller {
	case compiler.ControllerYou:
		if len(ctx.content.References) != 0 ||
			effect.UnderYourControl ||
			effect.UnderOwnersControl {
			return game.AbilityContent{}, false
		}
	case compiler.ControllerAny:
		// The all-graveyards form needs an explicit destination-controller rider
		// ("under your control" vs "under their owners' control"); the latter
		// leaves only an ownership pronoun reference ("their"), which the
		// UnderOwnersControl flag already captures, so permit pronoun references.
		if effect.ToZone != zone.Battlefield ||
			effect.UnderYourControl == effect.UnderOwnersControl ||
			!massGraveyardReferencesAllPronoun(ctx.content.References) {
			return game.AbilityContent{}, false
		}
		sourceGroup = game.AllPlayersReference()
		controlledByOwner = effect.UnderOwnersControl
	default:
		return game.AbilityContent{}, false
	}
	selection, ok := cardSelectionForSelector(selector)
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.MassReturnFromGraveyard{
			Player:            game.ControllerReference(),
			Selection:         selection,
			Destination:       effect.ToZone,
			EntryTapped:       effect.EntersTapped,
			SourceGroup:       sourceGroup,
			ControlledByOwner: controlledByOwner,
		},
	}}}.Ability(), true
}

// massGraveyardReferencesAllPronoun reports whether every reference in an
// all-graveyards mass-recursion clause is a grammatical pronoun (the "their" of
// "under their owners' control"), which carries no semantics the primitive needs.
func massGraveyardReferencesAllPronoun(references []compiler.CompiledReference) bool {
	for _, reference := range references {
		if reference.Kind != compiler.ReferencePronoun {
			return false
		}
	}
	return true
}

// lowerMassReanimationExchangeSpell lowers the single typed
// EffectMassReanimationExchange the parser collapses the symmetric
// mass-reanimation sentence into ("Each player exiles all <type> cards from
// their graveyard, then sacrifices all <type> they control, then puts all cards
// they exiled this way onto the battlefield." — Living Death, Living End, Scrap
// Mastery). The effect's selector carries only the card-type filter; the
// symmetric per-player behavior lives entirely in the runtime primitive, so the
// lowering simply forwards the type filter (with no controller narrowing, since
// every player acts on their own cards). It fails closed on any extra target,
// mode, condition, keyword, or a selector whose type filter is not a single
// creature/artifact card type.
func lowerMassReanimationExchangeSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := contentDiagnostic(
		ctx,
		"unsupported mass reanimation exchange",
		"the executable source backend supports only the symmetric exile-sacrifice-return reanimation of one creature or artifact card type",
	)
	if len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 {
		return game.AbilityContent{}, unsupported
	}
	selection, ok := cardSelectionForSelector(ctx.content.Effects[0].Selector)
	if !ok || len(selection.RequiredTypes) != 1 {
		return game.AbilityContent{}, unsupported
	}
	selection.Controller = game.ControllerAny
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.MassReanimationExchange{Selection: selection},
	}}}.Ability(), nil
}

func lowerTargetedGraveyardReturn(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Targets) != 1 ||
		len(ctx.content.Effects) != 1 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		ctx.content.Effects[0].FromZone != zone.Graveyard {
		return game.AbilityContent{}, false
	}
	// The plain return clause is byte-exact. The only inexact form accepted here
	// is a return-to-battlefield carrying a recognized counter entry rider ("...
	// with a +1/+1 counter on it", "... with two additional +1/+1 counters on
	// it"), whose supported counter kind and fixed count the compiler captures;
	// every other inexact return fails closed.
	counterRider := ctx.content.Effects[0].ToZone == zone.Battlefield &&
		ctx.content.Effects[0].CounterKindKnown
	if !ctx.content.Effects[0].Exact && !counterRider {
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
		var entryCounters []game.CounterPlacement
		if counterRider {
			// The counter count rides in the effect amount; a single-target return
			// keeps that amount unambiguous (multi-target cardinality would also
			// land in the amount), so only the fixed positive single-target form is
			// modeled.
			if targetSpec.MaxTargets != 1 ||
				!ctx.content.Effects[0].Amount.Known ||
				ctx.content.Effects[0].Amount.Value < 1 {
				return game.AbilityContent{}, false
			}
			entryCounters = []game.CounterPlacement{{
				Kind:   ctx.content.Effects[0].CounterKind,
				Amount: ctx.content.Effects[0].Amount.Value,
			}}
		}
		for i := range targetSpec.MaxTargets {
			put := game.PutOnBattlefield{
				Source:        game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: i}),
				EntryTapped:   ctx.content.Effects[0].EntersTapped,
				EntryCounters: entryCounters,
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

// lowerManaValueDynamicBound maps a compiled dynamic mana-value bound kind to a
// runtime ManaValueDynamicBound. Only the turn-event life totals are modeled
// (matching the runtime predicate); any other kind fails closed.
func lowerManaValueDynamicBound(kind compiler.DynamicAmountKind) (game.ManaValueDynamicBound, bool) {
	switch kind {
	case compiler.DynamicAmountLifeLostThisTurn:
		return game.ManaValueDynamicBound{Kind: game.DynamicAmountLifeLostThisTurn, Multiplier: 1}, true
	case compiler.DynamicAmountLifeGainedThisTurn:
		return game.ManaValueDynamicBound{Kind: game.DynamicAmountLifeGainedThisTurn, Multiplier: 1}, true
	default:
		return game.ManaValueDynamicBound{}, false
	}
}

// DO-NOT-COPY(filter): projects card-zone (graveyard/hand/library/exile)
// selections, including the bare "a card" (SelectorCard) noun and a zone
// filter, which the battlefield-only canonical projector fails closed on by
// design; prefer SelectionForSelectorMasked for new code. (retire: #1393)
func cardSelectionForSelector(selector compiler.CompiledSelector) (game.Selection, bool) {
	if selector.PowerLessThanSource || selector.PowerGreaterThanSource {
		// A source-relative power comparison applies only to a targeted
		// permanent (Mentor); a card-zone selection has no source to compare
		// against, so reject it rather than silently dropping the filter.
		return game.Selection{}, false
	}
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
	// A type union (RequiredTypesAny) carries the full disjunctive set of card
	// types, including the selector Kind's own type as its first member. The
	// single-Kind RequiredTypes above is a conjunctive (AND) requirement, so
	// leaving it set alongside a union would intersect the union down to the
	// Kind's type alone ("creature or enchantment card" matching creatures
	// only). Drop it so the union's OR semantics stand, mirroring the permanent
	// target path's union overwrite in permanentTargetSpecWithCardinality.
	if len(selection.RequiredTypesAny) > 0 {
		selection.RequiredTypes = nil
	}
	if selector.Historic {
		// A historic card is an artifact, a legendary, or a Saga (CR 702.61b).
		// That spans a card type, a supertype, and a subtype, which the flat
		// type/supertype/subtype fields cannot OR together, so it lowers to an
		// AnyOf disjunction. AnyOf is conjunctive with the selection's other
		// fields, so a card-type Kind or controller filter still applies on top.
		selection.AnyOf = append(selection.AnyOf,
			game.Selection{RequiredTypes: []types.Card{types.Artifact}},
			game.Selection{Supertypes: []types.Super{types.Legendary}},
			game.Selection{SubtypesAny: []types.Sub{types.Saga}},
		)
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
		// game.Selection's mana-value bound is a fixed comparison; it cannot
		// express the spell's chosen {X} ("with mana value X or less"), so an
		// X-derived bound fails closed here rather than lowering to a wrong fixed
		// bound. Only the dedicated library-search path models the X bound.
		if selector.ManaValueX {
			return game.Selection{}, false
		}
		selection.ManaValue = opt.Val(selector.ManaValue)
	}
	if selector.ManaValueDynamic != compiler.DynamicAmountNone {
		bound, ok := lowerManaValueDynamicBound(selector.ManaValueDynamic)
		if !ok {
			return game.Selection{}, false
		}
		selection.ManaValueDynamic = opt.Val(bound)
	}
	// A total (set-sum) mana-value bound is not a per-card filter; game.Selection
	// cannot express it. Fail closed here so no path silently drops the
	// constraint; the dedicated total-mana-value reanimation lowering clears this
	// flag before building the selection and models the cap on the primitive.
	if selector.MatchTotalManaValue {
		return game.Selection{}, false
	}
	selection.Colorless = selector.Colorless
	selection.Multicolored = selector.Multicolored
	if selector.MatchPower || selector.MatchToughness {
		return game.Selection{}, false
	}
	return selection, true
}

func lowerCounterPlacementSpell(
	ctx contentCtx,
) (game.AbilityContent, *shared.Diagnostic) {
	if content, ok := lowerAttachedCounterPlacement(ctx); ok {
		return content, nil
	}
	effect := ctx.content.Effects[0]
	if content, ok := lowerGroupCounterPlacement(ctx); ok {
		return content, nil
	}
	if content, ok := lowerSingleChoiceCounterPlacement(ctx); ok {
		return content, nil
	}
	if len(ctx.content.Targets) == 0 &&
		len(ctx.content.References) == 1 &&
		(ctx.content.References[0].Binding == compiler.ReferenceBindingSource ||
			ctx.content.References[0].Binding == compiler.ReferenceBindingTarget ||
			ctx.content.References[0].Binding == compiler.ReferenceBindingEventPermanent) {
		return lowerReferencedCounterPlacement(ctx)
	}
	if content, ok := lowerMultiTargetCounterPlacement(ctx); ok {
		return content, nil
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

	if !singleTargetCounterReferencesOK(effect.Amount, ctx.content.References) {
		return game.AbilityContent{}, unsupportedCounterPlacementDiagnostic(ctx)
	}
	amount := game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX})
	switch {
	case effect.Amount.Known:
		amount = game.Fixed(effect.Amount.Value)
	case effect.Amount.VariableX:
	case effect.Amount.DynamicKind != compiler.DynamicAmountNone:
		dynamic, supported := lowerDynamicAmount(effect.Amount, game.SourcePermanentReference())
		if !supported {
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

// lowerAttachedCounterPlacement lowers an exact fixed counter placement on the
// permanent the source Aura is attached to ("At the beginning of your upkeep,
// put a +1/+1 counter on enchanted creature."). The runtime resolves the
// recipient through its source attached-permanent reference, the same reference
// Aura stat buffs use, and no-ops when the source is unattached, so the
// placement needs no target. It is restricted to fixed positive amounts of a
// supported permanent counter kind, failing closed for player counters, dynamic
// or variable amounts, and any referenced, targeted, conditional, or modal
// shape.
func lowerAttachedCounterPlacement(ctx contentCtx) (game.AbilityContent, bool) {
	effect := ctx.content.Effects[0]
	if !effect.CounterRecipientAttached ||
		!effect.Exact ||
		effect.Negated ||
		effect.Context != parser.EffectContextController ||
		!effect.Amount.Known ||
		effect.Amount.Value <= 0 ||
		!effect.CounterKindKnown ||
		!compiler.CounterKindPlacementSupported(effect.CounterKind) ||
		effect.CounterKind.PlayerOnly() ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.AddCounter{
				Amount:      game.Fixed(effect.Amount.Value),
				Object:      game.SourceAttachedPermanentReference(),
				CounterKind: effect.CounterKind,
			},
		}},
	}.Ability(), true
}

// lowerMultiTargetCounterPlacement lowers an exact counter placement that puts
// counters on each of several targets ("Put a +1/+1 counter on each of up to two
// target creatures.") or on an optional single target ("Put a number of +1/+1
// counters on up to one other target creature you control equal to the amount of
// life you gained this turn." — Betor, Ancestor's Voice). The runtime models
// this as a single target spec with one AddCounter instruction per target index,
// mirroring the per-target instruction fan-out of multi-target graveyard return;
// resolution skips instructions whose optional target was not chosen. The
// per-target count is a fixed positive amount or a life-changed-this-turn
// dynamic amount (the amount of life you gained/lost this turn), applied
// identically to every chosen target. It is restricted to a supported permanent
// counter kind on a plain permanent target, failing closed for player counters
// and any referenced or conditional shape. The plain single target ("Put a
// +1/+1 counter on target creature") is handled by the dedicated single-target
// branch in lowerCounterPlacementSpell, not here.
func lowerMultiTargetCounterPlacement(ctx contentCtx) (game.AbilityContent, bool) {
	effect := ctx.content.Effects[0]
	if len(ctx.content.Targets) != 1 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		!effect.Exact ||
		effect.Negated ||
		effect.Context != parser.EffectContextController ||
		!effect.CounterKindKnown ||
		!compiler.CounterKindPlacementSupported(effect.CounterKind) ||
		effect.CounterKind.PlayerOnly() {
		return game.AbilityContent{}, false
	}
	amount, ok := multiTargetCounterAmount(effect.Amount)
	if !ok {
		return game.AbilityContent{}, false
	}
	target := ctx.content.Targets[0]
	// Handle plural ("two target creatures") and optional ("up to one/two
	// target creatures") cardinalities via per-target fan-out. The plain single
	// target (exactly one) is handled by the dedicated single-target branch.
	if target.Cardinality.Max < 1 ||
		(target.Cardinality.Min == 1 && target.Cardinality.Max == 1) {
		return game.AbilityContent{}, false
	}
	spec, ok := permanentTargetSpecWithCardinality(target)
	if !ok {
		return game.AbilityContent{}, false
	}
	sequence := make([]game.Instruction, 0, spec.MaxTargets)
	for i := range spec.MaxTargets {
		sequence = append(sequence, game.Instruction{Primitive: game.AddCounter{
			Amount:      amount,
			Object:      game.TargetPermanentReference(i),
			CounterKind: effect.CounterKind,
		}})
	}
	return game.Mode{
		Targets:  []game.TargetSpec{spec},
		Sequence: sequence,
	}.Ability(), true
}

// multiTargetCounterAmount resolves the per-target count for a fan-out counter
// placement, accepting a fixed positive amount or a life-changed-this-turn
// dynamic amount ("a number of +1/+1 counters ... equal to the amount of life
// you gained this turn" — Betor, Ancestor's Voice). Those turn-event totals are
// computed once for the controller and apply identically to every chosen target.
// It fails closed for non-positive fixed amounts, the variable X, and every
// other dynamic amount, whose per-target or distribution semantics this fan-out
// does not model.
func multiTargetCounterAmount(amount compiler.CompiledAmount) (game.Quantity, bool) {
	switch {
	case amount.Known:
		if amount.Value < 1 {
			return game.Quantity{}, false
		}
		return game.Fixed(amount.Value), true
	case amount.DynamicKind == compiler.DynamicAmountLifeGainedThisTurn ||
		amount.DynamicKind == compiler.DynamicAmountLifeLostThisTurn:
		dynamic, ok := lowerDynamicAmount(amount, game.SourcePermanentReference())
		if !ok {
			return game.Quantity{}, false
		}
		return game.Dynamic(dynamic), true
	default:
		return game.Quantity{}, false
	}
}

// counterPlacementKeywordsBenign reports whether every semantic keyword on a
// counter-placement ability is a benign artifact of naming a keyword counter:
// the parser records a spurious keyword for wordings like "flying counter"
// (the keyword name inside the counter name). Such keywords match the placed
// keyword counter's granted keyword exactly. Any keyword that does not match is
// a genuine ability the group lowerer must not silently drop, so it fails closed.
func counterPlacementKeywordsBenign(keywords []compiler.CompiledKeyword, kind counter.Kind) bool {
	placed, ok := keywordCounterKeyword(kind)
	if !ok {
		return len(keywords) == 0
	}
	for i := range keywords {
		keyword, ok := runtimeKeyword(keywords[i].Kind)
		if !ok || keyword != placed {
			return false
		}
	}
	return true
}

// keywordCounterKeyword maps a keyword counter kind to the keyword it grants,
// mirroring the runtime keywordCounters mapping. It reports false for counter
// kinds that grant no keyword.
func keywordCounterKeyword(kind counter.Kind) (game.Keyword, bool) {
	switch kind {
	case counter.Deathtouch:
		return game.Deathtouch, true
	case counter.FirstStrike:
		return game.FirstStrike, true
	case counter.Flying:
		return game.Flying, true
	case counter.Hexproof:
		return game.Hexproof, true
	case counter.Indestructible:
		return game.Indestructible, true
	case counter.Lifelink:
		return game.Lifelink, true
	case counter.Menace:
		return game.Menace, true
	case counter.Reach:
		return game.Reach, true
	case counter.Trample:
		return game.Trample, true
	case counter.Vigilance:
		return game.Vigilance, true
	default:
		return game.KeywordNone, false
	}
}

// lowerGroupCounterPlacement lowers an exact counter placement on every
// permanent in a filtered battlefield group ("Put a +1/+1 counter on each
// creature you control."). It reuses the group recipient reconstruction shared
// with group damage so the exactness gate and the executable backend accept the
// same filtered groups, and supports a fixed positive count or a recognized
// dynamic count (such as "X +1/+1 counters … where X is the number of +1/+1
// counters on this creature") of a supported permanent counter kind.
func lowerGroupCounterPlacement(ctx contentCtx) (game.AbilityContent, bool) {
	effect := ctx.content.Effects[0]
	if len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		!counterPlacementKeywordsBenign(ctx.content.Keywords, effect.CounterKind) ||
		!effect.Exact ||
		effect.Negated ||
		effect.Context != parser.EffectContextController ||
		!effect.CounterKindKnown ||
		!compiler.CounterKindPlacementSupported(effect.CounterKind) ||
		effect.CounterKind.PlayerOnly() ||
		effect.CounterRecipientSingleChoice {
		return game.AbilityContent{}, false
	}
	// A "with a <kind> counter on it/them" group filter carries a trailing
	// pronoun referent ("it"/"them") that names the filtered permanent, not a
	// placement count subject. The runtime represents that filter through the
	// group selection's counter requirement, so drop the qualifier pronoun
	// before validating the placement amount; otherwise its stray reference
	// makes the fixed-amount shape look reference-bearing and the group fails to
	// reconstruct.
	references := ctx.content.References
	if effect.Selector.MatchCounter || effect.Selector.MatchAnyCounter {
		references = counterQualifierFilteredReferences(references)
	}
	amount, ok := groupCounterPlacementAmount(effect.Amount, references)
	if !ok {
		return game.AbilityContent{}, false
	}
	group, ok := groupCounterRecipient(effect.Selector)
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.AddCounter{
				Amount:      amount,
				Group:       group,
				CounterKind: effect.CounterKind,
			},
		}},
	}.Ability(), true
}

// lowerSingleChoiceCounterPlacement lowers an exact fixed counter placement whose
// non-target recipient is a single chosen member of a battlefield group ("put a
// vigilance counter on a creature you control", Ajani Fells the Godsire chapter
// II; "another creature you control"). The resolving controller chooses one
// matching permanent at resolution, so the placement carries the group and the
// ChooseOne flag rather than a target. It reuses the group recipient projection,
// so every filter the group form supports is supported here too, and is
// restricted to fixed positive amounts of a supported permanent counter kind
// with no references, targets, conditions, or modes.
func lowerSingleChoiceCounterPlacement(ctx contentCtx) (game.AbilityContent, bool) {
	effect := ctx.content.Effects[0]
	if len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		!counterPlacementKeywordsBenign(ctx.content.Keywords, effect.CounterKind) ||
		!effect.Exact ||
		effect.Negated ||
		effect.Context != parser.EffectContextController ||
		!effect.CounterKindKnown ||
		!compiler.CounterKindPlacementSupported(effect.CounterKind) ||
		effect.CounterKind.PlayerOnly() ||
		!effect.CounterRecipientSingleChoice ||
		!effect.Amount.Known || effect.Amount.Value < 1 {
		return game.AbilityContent{}, false
	}
	group, ok := groupCounterRecipient(effect.Selector)
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.AddCounter{
				Amount:      game.Fixed(effect.Amount.Value),
				Group:       group,
				CounterKind: effect.CounterKind,
				ChooseOne:   true,
			},
		}},
	}.Ability(), true
}

// counter placement and reconstructs its count: a fixed positive amount accepts
// no references, while a recognized dynamic amount accepts either no references
// or source-bound referents (such as "this creature" in "where X is the number
// of +1/+1 counters on this creature"). It fails closed for any other shape.
func groupCounterPlacementAmount(
	amount compiler.CompiledAmount,
	references []compiler.CompiledReference,
) (game.Quantity, bool) {
	if amount.Known {
		if amount.Value < 1 || len(references) != 0 {
			return game.Quantity{}, false
		}
		return game.Fixed(amount.Value), true
	}
	if amount.DynamicKind == compiler.DynamicAmountNone {
		return game.Quantity{}, false
	}
	if len(references) != 0 &&
		!referencesBindTo(references, compiler.ReferenceBindingSource, 0) {
		return game.Quantity{}, false
	}
	dynamic, supported := lowerDynamicAmount(amount, game.SourcePermanentReference())
	if !supported {
		return game.Quantity{}, false
	}
	return game.Dynamic(dynamic), true
}

// counterQualifierFilteredReferences drops the pronoun referent a
// "with a <kind> counter on it/them" group filter introduces ("it", "them"),
// leaving the references that genuinely bind a placement count. The caller
// applies it only when the recipient selector carries a counter requirement, so
// the only pronoun present names the filtered permanent rather than a placement
// recipient or count subject.
func counterQualifierFilteredReferences(references []compiler.CompiledReference) []compiler.CompiledReference {
	filtered := make([]compiler.CompiledReference, 0, len(references))
	for _, reference := range references {
		if reference.Kind == compiler.ReferencePronoun && counterQualifierPronoun(reference.Pronoun) {
			continue
		}
		filtered = append(filtered, reference)
	}
	return filtered
}

func counterQualifierPronoun(pronoun compiler.ReferencePronounKind) bool {
	switch pronoun {
	case compiler.ReferencePronounIt,
		compiler.ReferencePronounIts,
		compiler.ReferencePronounThem,
		compiler.ReferencePronounThose,
		compiler.ReferencePronounThey,
		compiler.ReferencePronounTheir:
		return true
	default:
		return false
	}
}

// groupCounterQualifierClause reports whether an ordered-sequence effect is an
// exact counter placement on a filtered battlefield group whose filter requires
// group members to carry a counter ("each creature you control with a +1/+1
// counter on it"). Such a clause carries a qualifier pronoun ("it"/"them")
// naming each filtered member rather than a prior clause's target, so the
// sequence lowerer drops that pronoun before antecedent target binding.
func groupCounterQualifierClause(effect *compiler.CompiledEffect) bool {
	if !effect.Exact ||
		!effect.CounterKindKnown ||
		effect.Context != parser.EffectContextController ||
		(!effect.Selector.MatchCounter && !effect.Selector.MatchAnyCounter) {
		return false
	}
	_, ok := groupCounterRecipient(effect.Selector)
	return ok
}

// groupCounterRecipient reconstructs the battlefield group a counter placement
// targets. It first reuses the group-damage recipient reconstruction so groups
// already accepted for group damage lower identically, then falls back to the
// broader mass-group selection (which represents supertype filters such as
// "each legendary creature you control") so dynamic-count placements reach the
// same groups other mass effects already support.
func groupCounterRecipient(sel compiler.CompiledSelector) (game.GroupReference, bool) {
	if group, ok := damageGroupRecipient(sel); ok {
		return group, true
	}
	selection, ok := massGroupSelection(sel)
	if !ok {
		return game.GroupReference{}, false
	}
	if sel.Another || sel.Other {
		return game.BattlefieldGroupExcluding(selection, game.SourcePermanentReference()), true
	}
	return game.BattlefieldGroup(selection), true
}

// lowerReferencedCounterPlacement lowers an exact fixed counter placement whose
// object is a single referenced permanent: the source permanent itself ("Put a
// +1/+1 counter on this creature."), a prior clause's target referenced by "it"
// in an ordered sequence ("… Put a +1/+1 counter on it."), or the permanent
// involved in the triggering event referenced by "it"/"that creature" ("Whenever
// a creature you control enters, put a +1/+1 counter on that creature."). The
// object lowers to game.SourcePermanentReference(), a target reference, or
// game.EventPermanentReference() accordingly. The EventPermanent binding is
// accepted only for standalone (non-sequence) effects, since within a sequence
// the compiler binds a pronoun whose antecedent is a prior instruction's product
// to the triggering event permanent. Restricted to fixed positive amounts of a
// supported permanent counter kind.
func lowerReferencedCounterPlacement(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		!effect.Amount.Known || effect.Amount.Value <= 0 ||
		!effect.Exact ||
		effect.Negated ||
		effect.Context != parser.EffectContextController {
		return game.AbilityContent{}, unsupportedCounterPlacementDiagnostic(ctx)
	}
	kindChoices, ok := referencedCounterKindChoices(effect)
	if !ok {
		return game.AbilityContent{}, unsupportedCounterPlacementDiagnostic(ctx)
	}
	object, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{
		AllowSource: true,
		AllowTarget: true,
		AllowEvent:  !ctx.sequenceClause || ctx.allowEventPronoun,
	})
	if !ok {
		return game.AbilityContent{}, unsupportedCounterPlacementDiagnostic(ctx)
	}
	add := game.AddCounter{
		Amount: game.Fixed(effect.Amount.Value),
		Object: object,
	}
	if len(kindChoices) != 0 {
		add.KindChoices = kindChoices
	} else {
		add.CounterKind = effect.CounterKind
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: add,
		}},
	}.Ability(), nil
}

// referencedCounterKindChoices resolves the counter kind(s) a referenced counter
// placement applies. A single recognized, placeable kind yields no choice list; a
// two-or-more-kind choice ("a +1/+1 counter or a loyalty counter", Elspeth
// Conquers Death chapter III) yields the list of placeable kinds. Any
// unrecognized, player-only, or otherwise unplaceable kind fails closed.
func referencedCounterKindChoices(effect compiler.CompiledEffect) ([]counter.Kind, bool) {
	if len(effect.CounterKindChoices) >= 2 {
		for _, kind := range effect.CounterKindChoices {
			if !compiler.CounterKindPlacementSupported(kind) || kind.PlayerOnly() {
				return nil, false
			}
		}
		return append([]counter.Kind(nil), effect.CounterKindChoices...), true
	}
	if effect.CounterKindKnown &&
		compiler.CounterKindPlacementSupported(effect.CounterKind) &&
		!effect.CounterKind.PlayerOnly() {
		return nil, true
	}
	return nil, false
}

// lowerMoveCountersSpell lowers the counter-movement family ("Move a +1/+1
// counter from this creature onto target creature.", "Move all counters from
// this permanent onto target creature.") into a single MoveCounters instruction
// that reads counters from the ability's own source permanent
// (CounterSourceSelf) and places them on the single target. The specific-kind
// form moves one counter of the recognized kind; the kind-agnostic "all
// counters" form moves every counter regardless of kind. It fails closed for any
// shape the parser did not recognize as exact, any non-controller or negated
// effect, a missing or non-source self reference, a non-single target, and any
// conditional or modal content.
func lowerMoveCountersSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if effect.MoveCountersDistribute {
		return lowerMoveCountersDistributeSpell(ctx)
	}
	if effect.MoveCountersFromTarget {
		return lowerMoveCountersFromTargetSpell(ctx)
	}
	if content, ok := lowerMoveCountersOntoEventPermanent(ctx); ok {
		return content, nil
	}
	if !effect.Exact ||
		effect.Negated ||
		effect.Context != parser.EffectContextController ||
		len(ctx.content.Targets) != 1 ||
		ctx.content.Targets[0].Cardinality.Min != 1 ||
		ctx.content.Targets[0].Cardinality.Max != 1 ||
		len(ctx.content.References) != 1 ||
		ctx.content.References[0].Binding != compiler.ReferenceBindingSource ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, unsupportedCounterPlacementDiagnostic(ctx)
	}
	target, ok := permanentTargetSpec(ctx.content.Targets[0])
	if !ok {
		return game.AbilityContent{}, unsupportedCounterPlacementDiagnostic(ctx)
	}
	move := game.MoveCounters{
		Object: game.TargetPermanentReference(0),
		Source: game.CounterSourceSpec{Kind: game.CounterSourceSelf},
	}
	if effect.MoveCountersAll {
		move.AllKinds = true
	} else {
		if !effect.CounterKindKnown ||
			!compiler.CounterKindPlacementSupported(effect.CounterKind) ||
			effect.CounterKind.PlayerOnly() ||
			!effect.Amount.Known ||
			effect.Amount.Value != 1 {
			return game.AbilityContent{}, unsupportedCounterPlacementDiagnostic(ctx)
		}
		move.Amount = game.Fixed(effect.Amount.Value)
		move.CounterKind = effect.CounterKind
	}
	return game.Mode{
		Targets: []game.TargetSpec{target},
		Sequence: []game.Instruction{{
			Primitive: move,
		}},
	}.Ability(), nil
}

// lowerRemoveCounterSpell lowers the "remove a counter from target permanent"
// family (Ferropede, Cemetery Desecrator, Mutated Cultist, Thrull Parasite,
// Medicine Runner) into a single RemoveCounter instruction acting on one target
// permanent. The kind-unspecified "a counter" wording leaves the kind to the
// resolving controller (ChooseKind); a named, placeable kind removes that kind
// directly. It fails closed for any non-controller or negated effect, a wrong
// target count or cardinality, a non-permanent target, a non-positive or
// dynamic amount, a non-placeable named kind, and any reference, conditional, or
// modal content.
func lowerRemoveCounterSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if !effect.Exact ||
		effect.Negated ||
		effect.Context != parser.EffectContextController ||
		len(ctx.content.Targets) != 1 ||
		ctx.content.Targets[0].Cardinality.Max != 1 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		!effect.Amount.Known ||
		effect.Amount.Value < 1 {
		return game.AbilityContent{}, unsupportedCounterPlacementDiagnostic(ctx)
	}
	target, ok := permanentTargetSpecWithCardinality(ctx.content.Targets[0])
	if !ok {
		return game.AbilityContent{}, unsupportedCounterPlacementDiagnostic(ctx)
	}
	remove := game.RemoveCounter{
		Amount: game.Fixed(effect.Amount.Value),
		Object: game.TargetPermanentReference(0),
	}
	if effect.CounterKindKnown {
		if !compiler.CounterKindPlacementSupported(effect.CounterKind) ||
			effect.CounterKind.PlayerOnly() {
			return game.AbilityContent{}, unsupportedCounterPlacementDiagnostic(ctx)
		}
		remove.CounterKind = effect.CounterKind
	} else {
		// The kind-unspecified form removes one counter of a single
		// controller-chosen kind; a plural unspecified count has no
		// single-choice resolution, so fail closed rather than removing the
		// whole amount from one kind.
		if effect.Amount.Value != 1 {
			return game.AbilityContent{}, unsupportedCounterPlacementDiagnostic(ctx)
		}
		remove.ChooseKind = true
	}
	return game.Mode{
		Targets: []game.TargetSpec{target},
		Sequence: []game.Instruction{{
			Primitive: remove,
		}},
	}.Ability(), nil
}

// lowerMoveCountersOntoEventPermanent lowers the Graft-style move "move a +1/+1
// counter from this creature onto that creature." where the destination "that
// creature" is the permanent of the triggering enters event (CR 702.57). It
// reads one counter of the named kind from the ability's own source
// (CounterSourceSelf) and places it on the event permanent
// (EventPermanentReference). It fails closed (ok=false) for any shape outside
// exactly: a controller-context, non-negated, single named +1/+1 move of one
// counter; no targets; exactly two references — one source-bound counter source
// and one event-permanent-bound destination — and no conditions or modes.
func lowerMoveCountersOntoEventPermanent(ctx contentCtx) (game.AbilityContent, bool) {
	effect := ctx.content.Effects[0]
	if effect.Negated ||
		effect.Context != parser.EffectContextController ||
		effect.MoveCountersAll ||
		!effect.CounterKindKnown ||
		!compiler.CounterKindPlacementSupported(effect.CounterKind) ||
		effect.CounterKind.PlayerOnly() ||
		!effect.Amount.Known ||
		effect.Amount.Value != 1 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 2 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	sources, destinations := 0, 0
	for _, reference := range ctx.content.References {
		switch reference.Binding {
		case compiler.ReferenceBindingSource:
			sources++
		case compiler.ReferenceBindingEventPermanent:
			destinations++
		default:
			return game.AbilityContent{}, false
		}
	}
	if sources != 1 || destinations != 1 {
		return game.AbilityContent{}, false
	}
	move := game.MoveCounters{
		Object:      game.EventPermanentReference(),
		Source:      game.CounterSourceSpec{Kind: game.CounterSourceSelf},
		Amount:      game.Fixed(effect.Amount.Value),
		CounterKind: effect.CounterKind,
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: move,
		}},
	}.Ability(), true
}

// lowerPutThoseCountersSpell lowers the counter-salvage form "put those counters
// on <destination>" (The Ozolith, Iron Apprentice). The counters are read from
// the triggering event permanent's last-known information
// (CounterSourceEventPermanent) and placed, regardless of kind, on either a
// single/optional target permanent or the source permanent itself. It applies
// only inside a triggered ability, where a triggering event permanent exists,
// and fails closed for any non-controller or negated effect, an unrepresentable
// target, and any conditional or modal content.
func lowerPutThoseCountersSpell(ctx contentCtx) (game.AbilityContent, bool) {
	effect := ctx.content.Effects[0]
	if !effect.MoveThoseCounters ||
		!effect.Exact ||
		effect.Negated ||
		effect.Context != parser.EffectContextController ||
		ctx.enclosingKind != compiler.AbilityTriggered ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	move := game.MoveCounters{
		Source:   game.CounterSourceSpec{Kind: game.CounterSourceEventPermanent},
		AllKinds: true,
	}
	if len(ctx.content.Targets) == 0 {
		move.Object = game.SourcePermanentReference()
		return game.Mode{
			Sequence: []game.Instruction{{Primitive: move}},
		}.Ability(), true
	}
	if len(ctx.content.Targets) != 1 {
		return game.AbilityContent{}, false
	}
	target, ok := permanentTargetSpecWithCardinality(ctx.content.Targets[0])
	if !ok {
		return game.AbilityContent{}, false
	}
	move.Object = game.TargetPermanentReference(0)
	return game.Mode{
		Targets:  []game.TargetSpec{target},
		Sequence: []game.Instruction{{Primitive: move}},
	}.Ability(), true
}

// lowerMoveCountersDistributeSpell lowers the "move any number of <kind>
// counters from this permanent onto other creatures" form (Forgotten Ancient)
// into a MoveCounters instruction that reads counters from the ability's own
// source (CounterSourceSelf) and distributes them, by the controller, among the
// "other creatures" group rather than a single target. It fails closed for any
// non-controller or negated effect, a missing or non-source self reference, a
// non-placeable or kind-unknown counter, an unrepresentable destination group,
// and any target, conditional, or modal content.
func lowerMoveCountersDistributeSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if !effect.Exact ||
		effect.Negated ||
		effect.Context != parser.EffectContextController ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 1 ||
		ctx.content.References[0].Binding != compiler.ReferenceBindingSource ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, unsupportedCounterPlacementDiagnostic(ctx)
	}
	if !effect.CounterKindKnown ||
		!compiler.CounterKindPlacementSupported(effect.CounterKind) ||
		effect.CounterKind.PlayerOnly() {
		return game.AbilityContent{}, unsupportedCounterPlacementDiagnostic(ctx)
	}
	selection, ok := massGroupSelection(effect.Selector)
	if !ok {
		return game.AbilityContent{}, unsupportedCounterPlacementDiagnostic(ctx)
	}
	move := game.MoveCounters{
		CounterKind: effect.CounterKind,
		Source:      game.CounterSourceSpec{Kind: game.CounterSourceSelf},
		Group:       game.GroupRef(game.BattlefieldGroupExcluding(selection, game.SourcePermanentReference())),
		Distribute:  true,
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: move,
		}},
	}.Ability(), nil
}

// lowerMoveCountersFromTargetSpell lowers the two-target counter-move form ("Move
// a counter from target permanent you control onto a second target permanent." —
// Nesting Grounds, "Move a +1/+1 counter from target creature onto a second
// target creature." — Daghatar, "Move all counters from target creature onto
// another target creature." — Fate Transfer) into a single MoveCounters
// instruction that reads counters from the first target (CounterSourceTarget)
// and places them on the second target. The kind-agnostic "all counters" form
// moves every counter regardless of kind; the named-kind form moves one counter
// of the recognized kind; the "a counter" form moves one counter of a kind the
// controller chooses. It fails closed for any non-controller or negated effect,
// a wrong target count, a non-permanent target, a non-placeable named kind, and
// any conditional or modal content.
func lowerMoveCountersFromTargetSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if !effect.Exact ||
		effect.Negated ||
		effect.Context != parser.EffectContextController ||
		len(ctx.content.Targets) != 2 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, unsupportedCounterPlacementDiagnostic(ctx)
	}
	sourceTarget, ok := permanentTargetSpec(ctx.content.Targets[0])
	if !ok {
		return game.AbilityContent{}, unsupportedCounterPlacementDiagnostic(ctx)
	}
	destTarget, ok := permanentTargetSpec(ctx.content.Targets[1])
	if !ok {
		return game.AbilityContent{}, unsupportedCounterPlacementDiagnostic(ctx)
	}
	move := game.MoveCounters{
		Object: game.TargetPermanentReference(1),
		Source: game.CounterSourceSpec{
			Kind:   game.CounterSourceTarget,
			Object: game.TargetPermanentReference(0),
		},
	}
	switch {
	case effect.MoveCountersAll:
		move.AllKinds = true
	case effect.MoveCountersAnyKind:
		move.Amount = game.Fixed(1)
		move.ChooseKind = true
	default:
		if !effect.CounterKindKnown ||
			!compiler.CounterKindPlacementSupported(effect.CounterKind) ||
			effect.CounterKind.PlayerOnly() {
			return game.AbilityContent{}, unsupportedCounterPlacementDiagnostic(ctx)
		}
		move.Amount = game.Fixed(1)
		move.CounterKind = effect.CounterKind
	}
	return game.Mode{
		Targets: []game.TargetSpec{sourceTarget, destTarget},
		Sequence: []game.Instruction{{
			Primitive: move,
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

func unsupportedLibraryPlacementDiagnostic(ctx contentCtx) *shared.Diagnostic {
	return contentDiagnostic(
		ctx,
		"unsupported library placement",
		"the executable source backend supports only exact target graveyard-to-library placement",
	)
}

func singleTargetCounterReferencesOK(
	amount compiler.CompiledAmount,
	references []compiler.CompiledReference,
) bool {
	if len(references) == 0 {
		return true
	}
	if amount.ReferenceSpan == (shared.Span{}) &&
		referencesBindTo(references, compiler.ReferenceBindingTarget, 0) {
		return true
	}
	return exactDynamicAmountReference(amount, references)
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
		ctx.content.Targets[1].Cardinality.Max != 1 ||
		ctx.content.Targets[1].Cardinality.Min < 0 ||
		ctx.content.Targets[1].Cardinality.Min > 1 ||
		ctx.content.Effects[0].Negated ||
		!ctx.content.Effects[0].Exact ||
		ctx.content.Targets[0].Selector.Another ||
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
	first, firstOK := fightCreatureTargetSpec(ctx.content.Targets[0], false)
	second, secondOK := fightCreatureTargetSpec(ctx.content.Targets[1], true)
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

// lowerSourceFightSpell lowers the source-referenced fight family, where the
// fighting permanent is the source pronoun rather than a second target:
// "it fights target creature" (the event permanent of an enters trigger) and
// "this creature fights up to one target creature an opponent controls" (the
// source permanent). The parser leaves a single subject reference bound to the
// event or source permanent and a single creature target; this emits a Fight
// with the resolved source object against the lone target, reusing the shared
// fight target spec so controller and "up to one" optionality carry through.
func lowerSourceFightSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := contentDiagnostic(
		ctx,
		"unsupported fight spell",
		"the executable source backend supports only a source permanent fighting one target creature",
	)
	effect := ctx.content.Effects[0]
	if effect.Negated ||
		effect.Optional ||
		ctx.optional ||
		effect.Selector.Another ||
		(effect.Context != parser.EffectContextReferencedObject &&
			effect.Context != parser.EffectContextSource) ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.References) != 1 ||
		(ctx.content.References[0].Binding != compiler.ReferenceBindingSource &&
			ctx.content.References[0].Binding != compiler.ReferenceBindingEventPermanent) ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, unsupported
	}
	target, ok := fightCreatureTargetSpec(ctx.content.Targets[0], false)
	if !ok {
		return game.AbilityContent{}, unsupported
	}
	object, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{
		AllowSource: true,
		AllowEvent:  true,
	})
	if !ok {
		return game.AbilityContent{}, unsupported
	}
	return game.Mode{
		Targets: []game.TargetSpec{target},
		Sequence: []game.Instruction{{
			Primitive: game.Fight{
				Object:        object,
				RelatedObject: game.TargetPermanentReference(0),
			},
		}},
	}.Ability(), nil
}

// lowerLookAtHandSpell lowers "Look at target player's hand." to a single
// player-targeted LookAtHand primitive. The parser retypes the possessive hand
// phrase into a clean "target player"/"target opponent" target, so this reads
// only the typed target and rejects any other shape (riders, conditions,
// keywords, modes, references).
func lowerLookAtHandSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported look-at-hand spell",
			"the executable source backend supports only looking at one target player's hand",
		)
	}
	if len(ctx.content.Targets) != 1 ||
		!effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		ctx.optional ||
		effect.Context != parser.EffectContextTarget ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 {
		return unsupported()
	}
	targetSpec, ok := playerTargetSpec(ctx.content.Targets[0])
	if !ok {
		return unsupported()
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{{
			Primitive: game.LookAtHand{Player: game.TargetPlayerReference(0)},
		}},
	}.Ability(), nil
}

// fightCreatureTargetSpec lowers one fight target. allowAnother permits the
// "another target creature" determiner on the second target, lowering it to a
// DistinctFromPriorTargets spec so the chosen creature must differ from the
// first fighter; "the other" (Selector.Other) and the directional combat
// qualifiers remain unsupported.
func fightCreatureTargetSpec(target compiler.CompiledTarget, allowAnother bool) (game.TargetSpec, bool) {
	if target.Cardinality.Max != 1 ||
		target.Cardinality.Min < 0 ||
		target.Cardinality.Min > 1 ||
		!fightTargetSelectsCreature(target.Selector) ||
		(target.Selector.Another && !allowAnother) ||
		target.Selector.Other ||
		target.Selector.Attacking ||
		target.Selector.Blocking ||
		target.Selector.Tapped ||
		target.Selector.Untapped {
		return game.TargetSpec{}, false
	}
	spec := game.TargetSpec{
		MinTargets:               target.Cardinality.Min,
		MaxTargets:               1,
		Constraint:               target.Text,
		Allow:                    game.TargetAllowPermanent,
		DistinctFromPriorTargets: target.Selector.Another,
		Predicate: game.TargetPredicate{
			PermanentTypes: []types.Card{types.Creature},
			Subtypes:       slices.Clone(target.Selector.SubtypesAny()),
			RequiredName:   target.Selector.RequiredName,
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

// fightTargetSelectsCreature reports whether a fight target's selector denotes a
// creature. The plain "target creature" selection compiles to SelectorCreature;
// a bare creature-subtype selection ("Target Mutant", The Curse of Fenric III)
// carries no card-type word, so it compiles to SelectorUnknown with only the
// subtype recorded. A subtype defined for the creature type (CR 205.3m) denotes
// a creature, so that subtype-only shape is accepted while every other Unknown
// selection, and any selector carrying additional type, color, or supertype
// filters the fight spec cannot represent, fails closed.
func fightTargetSelectsCreature(selector compiler.CompiledSelector) bool {
	if selector.Kind == compiler.SelectorCreature {
		return true
	}
	if selector.Kind != compiler.SelectorUnknown || len(selector.SubtypesAny()) == 0 {
		return false
	}
	if len(selector.RequiredTypesAny()) != 0 ||
		len(selector.ExcludedTypes()) != 0 ||
		len(selector.ColorsAny()) != 0 ||
		len(selector.ExcludedColors()) != 0 ||
		len(selector.Supertypes()) != 0 ||
		len(selector.ExcludedSupertypes()) != 0 ||
		len(selector.ExcludedSubtypes()) != 0 {
		return false
	}
	for _, subtype := range selector.SubtypesAny() {
		if !parser.SubtypeMatchesCardType(subtype, parser.CardTypeCreature) {
			return false
		}
	}
	return true
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

// lowerGainPlayerCounterSpell lowers "You get {E}…{E}." / "You get <N> <kind>
// counter(s)." and the broader "<recipient> gets <N> <kind> counter(s)." to a
// player-counter placement of the fixed count. The energy symbol form carries no
// named counter kind, so it defaults to energy; the named word form carries the
// recognized player counter. The recipient is resolved from the same typed
// context the fixed-life lowering uses: controller, defending player, the
// triggering "that player", a lone targeted player, or a player group ("each
// opponent" / "each player").
func lowerGainPlayerCounterSpell(
	ctx contentCtx,
	_ *parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	kind := counter.Energy
	if effect.CounterKindKnown {
		kind = effect.CounterKind
	}
	unsupported := contentDiagnostic(
		ctx,
		"unsupported gain player counter spell",
		"the executable source backend supports only exact gain player counter",
	)
	if effect.Negated || ctx.optional || !effect.Exact ||
		!effect.Amount.Known || effect.Amount.Value < 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, unsupported
	}
	amount := game.Fixed(effect.Amount.Value)
	if len(ctx.content.Targets) == 0 {
		var group game.PlayerGroupReference
		switch effect.Context {
		case parser.EffectContextEachOpponent, parser.EffectContextEachOtherPlayer:
			group = game.OpponentsReference()
		case parser.EffectContextEachPlayer:
			group = game.AllPlayersReference()
		default:
		}
		if group.Kind != game.PlayerGroupReferenceNone {
			return game.Mode{
				Sequence: []game.Instruction{{
					Primitive: game.AddPlayerCounter{
						Amount:      amount,
						PlayerGroup: group,
						CounterKind: kind,
					},
				}},
			}.Ability(), nil
		}
	}
	playerRef, targets, ok := gainPlayerCounterRecipient(ctx, effect)
	if !ok {
		return game.AbilityContent{}, unsupported
	}
	return game.Mode{
		Targets: targets,
		Sequence: []game.Instruction{{
			Primitive: game.AddPlayerCounter{
				Amount:      amount,
				Player:      playerRef,
				CounterKind: kind,
			},
		}},
	}.Ability(), nil
}

// gainPlayerCounterRecipient resolves the player who receives the counters from
// the effect's typed context, returning any target specs the reference needs.
// It mirrors the single-player recipients lowerFixedLifeSpell supports, plus the
// referenced-object-controller recipient ("its controller gets a counter"): the
// controller of an inherited sequence target ("Destroy target creature. Its
// controller gets a poison counter.") or of the triggering event permanent
// ("Whenever enchanted artifact becomes tapped, its controller gets a poison
// counter.").
func gainPlayerCounterRecipient(
	ctx contentCtx,
	effect compiler.CompiledEffect,
) (game.PlayerReference, []game.TargetSpec, bool) {
	switch {
	case len(ctx.content.Targets) == 0 && len(ctx.content.References) == 0 &&
		effect.Context == parser.EffectContextController:
		return game.ControllerReference(), nil, true
	case len(ctx.content.Targets) == 0 && len(ctx.content.References) == 0 &&
		effect.Context == parser.EffectContextDefendingPlayer:
		return game.DefendingPlayerReference(), nil, true
	case len(ctx.content.Targets) == 0 && len(ctx.content.References) == 1 &&
		(effect.Context == parser.EffectContextEventPlayer &&
			ctx.content.References[0].Kind == compiler.ReferencePronoun &&
			ctx.content.References[0].Pronoun == compiler.ReferencePronounThey ||
			effect.Context == parser.EffectContextReferencedPlayer &&
				ctx.content.References[0].Kind == compiler.ReferenceThatPlayer &&
				ctx.content.References[0].Binding != compiler.ReferenceBindingTarget):
		return game.EventPlayerReference(), nil, true
	case len(ctx.content.Targets) == 1 &&
		(effect.Context == parser.EffectContextTarget ||
			effect.Context == parser.EffectContextPriorSubject):
		targetSpec, ok := playerTargetSpec(ctx.content.Targets[0])
		if !ok {
			return game.PlayerReference{}, nil, false
		}
		return game.TargetPlayerReference(0), []game.TargetSpec{targetSpec}, true
	case len(ctx.content.Targets) == 1 &&
		effect.Context == parser.EffectContextReferencedObjectController:
		ref, ok := referencedControllerPlayerRef(ctx)
		return ref, nil, ok
	case len(ctx.content.Targets) == 0 &&
		len(ctx.content.References) == 1 &&
		effect.Context == parser.EffectContextReferencedObjectController &&
		ctx.content.References[0].Binding == compiler.ReferenceBindingEventPermanent:
		return game.ObjectControllerReference(game.EventPermanentReference()), nil, true
	default:
		return game.PlayerReference{}, nil, false
	}
}

// lowerAmassContent lowers a single amass keyword-action effect ("Amass Orcs N"
// / "Amass Zombies N" / "Amass N") to a game.Amass primitive carrying the fixed
// count and the named Army subtype recognized by the parser.
func lowerAmassContent(
	ctx contentCtx,
	syntax *parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	subtype := ctx.content.Effects[0].AmassSubtype
	return lowerExactPrimitiveSpell(
		ctx,
		syntax,
		"amass",
		func(amount game.Quantity) game.Primitive {
			return game.Amass{Amount: amount, Subtype: subtype}
		},
	)
}

// lowerRenownContent lowers the renown keyword action ("renown N") that the
// printed "Renown N" keyword expands to. It produces a game.Renown primitive
// targeting the source permanent and carrying the fixed counter count. The
// runtime guard applies the counters and the renowned mark only once.
func lowerRenownContent(
	ctx contentCtx,
	syntax *parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	return lowerExactPrimitiveSpell(
		ctx,
		syntax,
		"renown",
		func(amount game.Quantity) game.Primitive {
			return game.Renown{Object: game.SourcePermanentReference(), Amount: amount}
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
	// The look clause is classified EffectDig ("Look at the top two cards of
	// your library."); the manifest-dread long form is disambiguated from a plain
	// impulse dig by its second sentence putting a card onto the battlefield face
	// down rather than into hand.
	return look.Kind == compiler.EffectDig &&
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

// lowerWinGameSpell lowers the exact controller effect "You win the game."
// (Felidar Sovereign, Thassa's Oracle) to a single PlayerWinsGame instruction
// scoped to the ability's controller. It mirrors lowerExactPrimitiveSpell but
// carries no amount, since winning the game takes no count.
func lowerWinGameSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if effect.Negated ||
		!effect.Exact ||
		effect.Context != parser.EffectContextController ||
		ctx.content.Unconsumed() ||
		len(ctx.content.References) != 0 {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported win-game effect",
			"the executable source backend supports only the exact controller \"You win the game.\"",
		)
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.PlayerWinsGame{Player: game.ControllerReference()},
	}}}.Ability(), nil
}

// lowerLoseGameSpell lowers an exact EffectLoseGame body to a single
// PlayerLosesGame instruction scoped to the losing player. It mirrors
// lowerWinGameSpell but resolves the player from the effect context, supporting
// the controller ("You lose the game."), the referenced triggering player
// ("Whenever this creature deals combat damage to a player, that player loses
// the game."), and a single targeted player ("Target player loses the game.").
// Negated, optional, durational, conditional, and group forms fail closed.
func lowerLoseGameSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported lose-game effect",
			"the executable source backend supports only the exact controller, referenced, or target \"loses the game\" effect",
		)
	}
	effect := ctx.content.Effects[0]
	if effect.Negated ||
		!effect.Exact ||
		ctx.optional ||
		effect.Optional ||
		effect.Duration != compiler.DurationNone ||
		effect.DelayedTiming != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return unsupported()
	}
	player, targets, ok := loseGamePlayer(ctx, effect)
	if !ok {
		return unsupported()
	}
	return game.Mode{
		Targets: targets,
		Sequence: []game.Instruction{{
			Primitive: game.PlayerLosesGame{Player: player},
		}},
	}.Ability(), nil
}

// loseGamePlayer resolves the player who loses the game from an EffectLoseGame
// body's typed context. It accepts the controller, the referenced triggering
// player ("that player"), and a single targeted player, returning any target
// spec the reference requires. Every other recipient shape fails closed.
func loseGamePlayer(
	ctx contentCtx,
	effect compiler.CompiledEffect,
) (game.PlayerReference, []game.TargetSpec, bool) {
	switch {
	case len(ctx.content.Targets) == 0 && len(ctx.content.References) == 0 &&
		effect.Context == parser.EffectContextController:
		return game.ControllerReference(), nil, true
	case len(ctx.content.Targets) == 0 && len(ctx.content.References) == 1 &&
		effect.Context == parser.EffectContextReferencedPlayer &&
		ctx.content.References[0].Kind == compiler.ReferenceThatPlayer &&
		ctx.content.References[0].Binding != compiler.ReferenceBindingTarget:
		return game.EventPlayerReference(), nil, true
	case len(ctx.content.Targets) == 1 && len(ctx.content.References) == 0 &&
		effect.Context == parser.EffectContextTarget:
		targetSpec, ok := playerTargetSpec(ctx.content.Targets[0])
		if !ok {
			return game.PlayerReference{}, nil, false
		}
		return game.TargetPlayerReference(0), []game.TargetSpec{targetSpec}, true
	default:
		return game.PlayerReference{}, nil, false
	}
}

// lowerPreventDamageSpell lowers an EffectPreventDamage clause into one or two
// PreventDamage prevention shields (one per prevented direction) that prevent
// all combat damage to and/or from a single permanent for the turn. The
// permanent is named either by the clause's lone target (with a redundant
// "that creature" back-reference, as in Maze of Ith's untap sequence) or by a
// lone source/event back-reference ("it"/"this creature", as in Goblin
// Snowman and Moonlight Geist). The global form ("Prevent all combat damage
// that would be dealt this turn." — Spike Weaver) lowers to a single object-less
// combat-only shield that prevents every combat damage event for the turn.
func lowerPreventDamageSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported prevent-damage effect",
			"the executable source backend supports only preventing all combat damage to and/or from one referenced permanent this turn",
		)
	}
	if effect.Negated ||
		effect.Optional ||
		!effect.Exact ||
		effect.Context != parser.EffectContextController ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return unsupported()
	}
	if effect.PreventDamageGlobal {
		if effect.PreventDamageTo ||
			effect.PreventDamageBy ||
			len(ctx.content.Targets) != 0 ||
			len(ctx.content.References) != 0 {
			return unsupported()
		}
		mode := game.Mode{Sequence: []game.Instruction{{Primitive: game.PreventDamage{
			All:        true,
			CombatOnly: true,
			Global:     true,
		}}}}
		return mode.Ability(), nil
	}
	if !effect.PreventDamageTo && !effect.PreventDamageBy {
		return unsupported()
	}
	object, targetSpec, ok := preventDamageObject(ctx)
	if !ok {
		return unsupported()
	}
	var sequence []game.Instruction
	if effect.PreventDamageTo {
		sequence = append(sequence, game.Instruction{Primitive: game.PreventDamage{
			Object:     object,
			All:        true,
			CombatOnly: true,
		}})
	}
	if effect.PreventDamageBy {
		sequence = append(sequence, game.Instruction{Primitive: game.PreventDamage{
			Object:     object,
			All:        true,
			CombatOnly: true,
			BySource:   true,
		}})
	}
	mode := game.Mode{Sequence: sequence}
	if targetSpec != nil {
		mode.Targets = []game.TargetSpec{*targetSpec}
	}
	return mode.Ability(), nil
}

// preventDamageObject resolves the permanent an EffectPreventDamage clause
// shields, returning the runtime object reference and, for the targeted form, a
// TargetSpec to attach to the mode.
func preventDamageObject(ctx contentCtx) (game.ObjectReference, *game.TargetSpec, bool) {
	switch {
	case len(ctx.content.Targets) == 1:
		if !targetCardinalityIsOne(ctx.content.Targets[0]) ||
			!referencesAreRedundantSoleTargetBackReferences(ctx.content.References) {
			return game.ObjectReference{}, nil, false
		}
		targetSpec, ok := permanentTargetSpec(ctx.content.Targets[0])
		if !ok {
			return game.ObjectReference{}, nil, false
		}
		return game.TargetPermanentReference(0), &targetSpec, true
	case len(ctx.content.Targets) == 0 && len(ctx.content.References) == 1:
		object, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{
			AllowEvent:  true,
			AllowSource: true,
			AllowTarget: true,
		})
		if !ok {
			return game.ObjectReference{}, nil, false
		}
		return object, nil, true
	default:
		return game.ObjectReference{}, nil, false
	}
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
