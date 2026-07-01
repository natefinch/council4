package cardgen

import (
	"fmt"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// lowerTotalManaValueGraveyardReanimation lowers the non-target, set-sum-capped
// graveyard reanimation "Return up to N <filter> cards with total mana value X
// or less from your graveyard to the battlefield" (Lively Dirge). The player
// chooses up to N cards from their own graveyard whose combined mana value does
// not exceed X and puts them onto the battlefield, tapped when the wording says
// so. Unlike a per-card "with mana value X or less" filter, the constraint binds
// the chosen set as a whole, so it lowers to the ChooseFromZone
// Riders.MaxTotalManaValue cap rather than a Selection mana-value bound.
//
// It is card-name-blind and fails closed on any shape it does not fully model —
// a target or reference, a non-graveyard source, a destination other than the
// battlefield, a counter/control/duration rider, a per-card mana-value bound, a
// non-"or less" total comparison, an unknown or non-positive count, or a
// selector qualifier the chosen-card selection cannot express.
func lowerTotalManaValueGraveyardReanimation(ctx contentCtx) (game.AbilityContent, bool) {
	// lowerReturnSpell — this function's only caller — is reached solely through
	// the EffectReturn arm of lowerImmediateSingleEffectSpell, whose content is
	// always single-effect, so an effect count other than one or an effect kind
	// other than EffectReturn is a dispatch bug rather than an unsupported card.
	if len(ctx.content.Effects) != 1 {
		panic(fmt.Sprintf(
			"lowerTotalManaValueGraveyardReanimation: reached with %d effects; lowerReturnSpell dispatches only single-effect content",
			len(ctx.content.Effects)))
	}
	if ctx.optional ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectReturn {
		panic(fmt.Sprintf(
			"lowerTotalManaValueGraveyardReanimation: reached with effect kind %v; lowerReturnSpell dispatches only EffectReturn content",
			effect.Kind))
	}
	if effect.Negated ||
		effect.Optional ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		effect.FromZone != zone.Graveyard ||
		effect.ToZone != zone.Battlefield ||
		effect.UnderYourControl ||
		effect.UnderOwnersControl ||
		effect.CounterKindKnown {
		return game.AbilityContent{}, false
	}
	selector := effect.Selector
	if !selector.MatchTotalManaValue ||
		selector.MatchManaValue ||
		selector.TotalManaValue.Op != compare.LessOrEqual ||
		selector.TotalManaValue.Value < 0 ||
		selector.Zone != zone.Graveyard ||
		selector.Controller != compiler.ControllerYou ||
		selector.All ||
		selector.Another ||
		selector.Other ||
		selector.Attacking ||
		selector.Blocking ||
		(selector.Tapped && !effect.EntersTapped) ||
		selector.Untapped {
		return game.AbilityContent{}, false
	}
	if !effect.Amount.Known || effect.Amount.Value < 1 {
		return game.AbilityContent{}, false
	}
	plain := selector
	plain.MatchTotalManaValue = false
	plain.TotalManaValue = compare.Int{}
	selection, ok := cardSelectionForSelector(plain)
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.ReturnFromGraveyardChoice(
			game.ControllerReference(),
			selection,
			game.Fixed(effect.Amount.Value),
			zone.Battlefield,
			effect.EntersTapped,
			opt.Val(selector.TotalManaValue.Value),
			false,
			"",
		),
	}}}.Ability(), true
}
