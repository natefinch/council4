package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// lowerExileFromHandContent lowers "exile a <filter> card from your hand" — the
// imprint effect of Chrome Mox ("you may exile a nonartifact, nonland card from
// your hand"), reached after its "you may" prefix is stripped to the mandatory
// core. It produces one game.ChooseFromZone instruction that has the controller
// exile one matching card from their own hand and publishes the imprint link by
// object identity, so a sibling "one mana of any of the exiled card's colors"
// mana ability on the same face can read the imprinted card's colors.
//
// It is card-name-blind and fails closed (ok=false) on any shape it does not
// fully model — a body-level optional, a modal or multi-effect body, references
// or targets, a non-hand or non-"your" zone, a selector qualifier it cannot
// express, or an amount other than exactly one card — so an unmodeled wording
// falls through to the generic exile path's diagnostic rather than lowering to a
// silently-wrong instruction.
func lowerExileFromHandContent(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 0 ||
		ctx.content.Unconsumed() {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectExile ||
		effect.Optional ||
		effect.Negated ||
		effect.Divided ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		effect.Context != parser.EffectContextController {
		return game.AbilityContent{}, false
	}
	selector := effect.Selector
	if selector.Zone != zone.Hand ||
		selector.Kind != compiler.SelectorCard ||
		selector.Controller != compiler.ControllerAny ||
		selector.All ||
		selector.Another ||
		selector.Other ||
		selector.Attacking ||
		selector.Blocking ||
		selector.Tapped ||
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
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.ExileFromHandChoice(
			game.ControllerReference(),
			selection,
			game.Fixed(1),
			game.LinkedKey(imprintLinkKey),
		),
	}}}.Ability(), true
}
