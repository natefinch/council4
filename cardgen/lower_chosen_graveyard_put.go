package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// lowerChosenCardGraveyardPut lowers the non-target single-card reanimation
// "Put a <filter> card from [a/your] graveyard onto the battlefield under your
// control[, tapped]" (Extract from Darkness's resolving clause, Exhume-family
// controller wording). It is the EffectPut counterpart of
// lowerChosenCardGraveyardReturn: the controller chooses one matching card at
// resolution and puts it onto the battlefield under their control rather than
// targeting it. "from your graveyard" (ControllerYou) scans only the
// controller's graveyard; "from a graveyard" (ControllerAny) scans every
// player's, which the chosen card enters under the resolving controller's
// control regardless of owner.
//
// It is card-name-blind and fails closed on any shape it does not fully model —
// a target or reference, a non-graveyard source, a destination other than the
// battlefield, a missing "under your control" rider, an owners'-control rider, a
// counter rider, an amount other than exactly one card, or a selector qualifier
// the chosen-card selection cannot express.
func lowerChosenCardGraveyardPut(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Effects) != 1 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectPut ||
		effect.Negated ||
		effect.Divided ||
		effect.Optional ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		effect.FromZone != zone.Graveyard ||
		effect.ToZone != zone.Battlefield ||
		!effect.UnderYourControl ||
		effect.UnderOwnersControl ||
		effect.CounterKindKnown {
		return game.AbilityContent{}, false
	}
	if !effect.Amount.Known ||
		effect.Amount.RangeKnown ||
		effect.Amount.VariableX ||
		effect.Amount.DynamicKind != 0 ||
		effect.Amount.Value != 1 {
		return game.AbilityContent{}, false
	}
	selector := effect.Selector
	if selector.Zone != zone.Graveyard ||
		selector.All ||
		selector.Another ||
		selector.Other ||
		selector.Attacking ||
		selector.Blocking ||
		(selector.Tapped && !effect.EntersTapped) ||
		selector.Untapped {
		return game.AbilityContent{}, false
	}
	var allOwners bool
	switch selector.Controller {
	case compiler.ControllerYou:
		allOwners = false
	case compiler.ControllerAny:
		allOwners = true
	default:
		return game.AbilityContent{}, false
	}
	selection, ok := cardSelectionForSelector(selector)
	if !ok {
		return game.AbilityContent{}, false
	}
	primitive := game.ReturnFromGraveyardChoice(
		game.ControllerReference(),
		selection,
		game.Fixed(1),
		zone.Battlefield,
		effect.EntersTapped,
		opt.V[int]{},
		false,
		"",
	)
	primitive.AllOwners = allOwners
	return game.Mode{Sequence: []game.Instruction{{Primitive: primitive}}}.Ability(), true
}
