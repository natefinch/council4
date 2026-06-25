package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// aesirExiledCardKey is the linked key binding the permanent card a Saga's first
// chapter exiles from its controller's graveyard to the source permanent that
// exiled it. The runtime keys linked objects by source card-instance id plus
// this string, so the set survives across the Saga's chapters: chapter I
// publishes the exiled card, chapter II reads its mana value, and chapter III
// returns it to hand (The Aesir Escape Valhalla).
const aesirExiledCardKey = game.LinkedKey("exile-graveyard-card")

// lowerAesirExileGraveyardScaledGain lowers a Saga's first chapter "Exile a
// <filter> card from your graveyard. You gain life equal to its mana value."
// (The Aesir Escape Valhalla chapter I) into a linked exile-from-graveyard
// ChooseFromZone followed by a controller life gain scaled by the exiled card's
// mana value.
// The exile publishes the chosen card under aesirExiledCardKey, keyed by the
// source permanent, so the life gain reads its mana value through that link and
// later chapters can reference the same card. The "its" pronoun naming the
// exiled card is consumed by the scaled amount rather than as a target.
//
// It returns ok=false for any shape it does not fully consume: a target, a
// reference other than the single exiled-card pronoun, a condition, mode, or
// keyword rider, an optional or negated effect, a non-controller context, a
// non-"your"-graveyard or qualified exile selector, a non-fixed exile amount, or
// a life gain that is not exactly the exiled card's mana value, so an unmodeled
// wording fails closed.
func lowerAesirExileGraveyardScaledGain(ctx contentCtx) (game.AbilityContent, bool) {
	content := ctx.content
	if ctx.optional ||
		len(content.Effects) != 2 ||
		len(content.Targets) != 0 ||
		len(content.References) != 1 ||
		len(content.Conditions) != 0 ||
		len(content.Keywords) != 0 ||
		len(content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	selection, ok := aesirGraveyardExileSelection(&content.Effects[0])
	if !ok {
		return game.AbilityContent{}, false
	}
	amount, ok := aesirExiledCardManaValueAmount(&content.Effects[1])
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{Sequence: []game.Instruction{
		{Primitive: game.ExileFromGraveyardChoice(
			game.ControllerReference(),
			selection,
			game.Fixed(content.Effects[0].Amount.Value),
			false,
			aesirExiledCardKey,
		)},
		{Primitive: game.GainLife{
			Player: game.ControllerReference(),
			Amount: amount,
		}},
	}}.Ability(), true
}

// aesirGraveyardExileSelection reconstructs the card selection for the chapter I
// exile clause "Exile a <filter> card from your graveyard", reusing the same
// selector reconstruction as the standalone graveyard-choice exile so the filter
// stays card-name-blind. It fails closed for any non-"your"-graveyard scope, a
// qualified selector it cannot express, or a non-fixed positive amount.
func aesirGraveyardExileSelection(effect *compiler.CompiledEffect) (game.Selection, bool) {
	if effect.Kind != compiler.EffectExile ||
		effect.Negated ||
		effect.Divided ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		effect.Context != parser.EffectContextController ||
		effect.FromZone != zone.Graveyard {
		return game.Selection{}, false
	}
	selector := effect.Selector
	if selector.Zone != zone.Graveyard ||
		selector.Controller != compiler.ControllerYou ||
		selector.All ||
		selector.Another ||
		selector.Other ||
		selector.Attacking ||
		selector.Blocking ||
		selector.Tapped ||
		selector.Untapped {
		return game.Selection{}, false
	}
	if !effect.Amount.Known ||
		effect.Amount.RangeKnown ||
		effect.Amount.VariableX ||
		effect.Amount.DynamicKind != compiler.DynamicAmountNone ||
		effect.Amount.Value < 1 {
		return game.Selection{}, false
	}
	return cardSelectionForSelector(selector)
}

// aesirExiledCardManaValueAmount lowers the chapter I reward "You gain life
// equal to its mana value." into a controller life gain whose amount reads the
// exiled card's mana value through the linked object chapter I published. It
// fails closed for any gain that is not the controller's life total scaled by
// exactly the exiled card's mana value.
func aesirExiledCardManaValueAmount(effect *compiler.CompiledEffect) (game.Quantity, bool) {
	if effect.Kind != compiler.EffectGain ||
		!effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		effect.Context != parser.EffectContextController ||
		!effect.LifeObject ||
		effect.Amount.DynamicKind != compiler.DynamicAmountSourceManaValue {
		return game.Quantity{}, false
	}
	dynamic, ok := objectCharacteristicAmount(
		effect.Amount.DynamicKind,
		game.LinkedObjectReference(string(aesirExiledCardKey)),
	)
	if !ok {
		return game.Quantity{}, false
	}
	return game.Dynamic(dynamic), true
}

// lowerAesirCounterFromExiledCard lowers the chapter II clause "Put a number of
// +1/+1 counters on target creature you control equal to the mana value of the
// exiled card." (The Aesir Escape Valhalla) into a counter placement on the
// target whose amount reads the linked exiled card's mana value through the link
// chapter I published. The parser drops the dynamic amount span (the count lives
// behind the source link, not a printed number), so this consumes the exact
// recognized flag and rebuilds the amount from the link.
//
// It returns ok=false for any shape it does not fully consume: a missing or
// multi-target recipient, a reference, condition, mode, or keyword rider, an
// optional or negated effect, a non-controller context, or an unsupported
// counter kind, so an unmodeled wording fails closed.
func lowerAesirCounterFromExiledCard(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectPut ||
		!effect.CounterExiledCardManaValue ||
		!effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		effect.Context != parser.EffectContextController ||
		!effect.CounterKindKnown ||
		effect.CounterKind.PlayerOnly() ||
		!compiler.CounterKindPlacementSupported(effect.CounterKind) {
		return game.AbilityContent{}, false
	}
	target, ok := permanentTargetSpec(ctx.content.Targets[0])
	if !ok {
		return game.AbilityContent{}, false
	}
	amount := game.Dynamic(game.DynamicAmount{
		Kind:       game.DynamicAmountObjectManaValue,
		Multiplier: 1,
		Object:     game.LinkedObjectReference(string(aesirExiledCardKey)),
	})
	return game.Mode{
		Targets: []game.TargetSpec{target},
		Sequence: []game.Instruction{{
			Primitive: game.AddCounter{
				Amount:      amount,
				Object:      game.TargetPermanentReference(0),
				CounterKind: effect.CounterKind,
			},
		}},
	}.Ability(), true
}

// lowerAesirReturnSourceAndExiledCard lowers the chapter III clause "Return this
// Saga and the exiled card to their owner's hand." (The Aesir Escape Valhalla)
// into a linked return of the exiled card to hand paired with a bounce of the
// source permanent. The exiled card is the one chapter I published under
// aesirExiledCardKey, returned before the source bounce so the source-keyed link
// resolves while the Saga is still on the battlefield. The "this Saga" and
// owner-pronoun references are consumed in place of a target.
//
// It returns ok=false for any shape it does not fully consume: a target, a
// condition, mode, or keyword rider, an optional or negated effect, or a
// non-controller context, so an unmodeled wording fails closed.
func lowerAesirReturnSourceAndExiledCard(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectReturn ||
		!effect.ReturnSourceAndExiledCardToHand ||
		!effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		effect.Context != parser.EffectContextController {
		return game.AbilityContent{}, false
	}
	return game.Mode{Sequence: []game.Instruction{
		{Primitive: game.ReturnExiledCardsToHand{LinkedKey: aesirExiledCardKey}},
		{Primitive: game.Bounce{Object: game.SourcePermanentReference()}},
	}}.Ability(), true
}
