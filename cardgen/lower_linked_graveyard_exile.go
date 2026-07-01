package cardgen

import (
	"fmt"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// exiledWithSourceKey is the linked key binding every card a source's ability
// exiles from a graveyard to the source permanent that exiled it. The runtime
// keys linked objects by the source permanent's card-instance identity plus this
// string, so a later static ability on the same source can read the set of
// "cards exiled with this creature" (Cemetery Prowler's exile-linked cost
// reduction reads the distinct card types among them).
const exiledWithSourceKey = game.LinkedKey("exiled-with-source")

// lowerLinkedAnyGraveyardChoiceExile lowers the non-target, any-graveyard exile
// "exile a <filter> card from a graveyard" (Cemetery Prowler's enter-or-attack
// trigger), where the controller chooses one card from any player's graveyard at
// resolution. It produces one game.ChooseFromZone whose AllOwners flag widens
// the candidate pool to every player's graveyard and whose PublishLinked records
// each exiled card under exiledWithSourceKey, so a sibling static on the same
// source can read the cards exiled with it.
//
// It is the any-graveyard sibling of lowerControllerGraveyardChoiceExile, which
// covers the "from your graveyard" controller-scoped form. It is card-name-blind
// and fails closed (ok=false) on any shape it does not fully model — a reference
// or target, a non-graveyard source, a non-any controller scope, a selector
// qualifier it cannot express, or a non-fixed positive amount — so an unmodeled
// wording falls through to the generic exile path's diagnostic rather than
// lowering to a silently-wrong instruction.
func lowerLinkedAnyGraveyardChoiceExile(ctx contentCtx) (game.AbilityContent, bool) {
	// lowerFixedExileSpell reaches this only through lowerImmediateSingleEffectSpell's
	// EffectExile arm, which guarantees single-effect content whose sole effect is an
	// EffectExile; a different count or kind is a dispatch bug, not an unsupported card.
	if len(ctx.content.Effects) != 1 {
		panic(fmt.Sprintf("lowerLinkedAnyGraveyardChoiceExile: reached with %d effects; the EffectExile dispatch is single-effect", len(ctx.content.Effects)))
	}
	if len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectExile {
		panic(fmt.Sprintf("lowerLinkedAnyGraveyardChoiceExile: reached with effect kind %v; the EffectExile dispatch guarantees EffectExile", effect.Kind))
	}
	if effect.Negated ||
		effect.Divided ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		effect.Context != parser.EffectContextController ||
		effect.FromZone != zone.Graveyard {
		return game.AbilityContent{}, false
	}
	selector := effect.Selector
	if selector.Zone != zone.Graveyard ||
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
		effect.Amount.DynamicKind != compiler.DynamicAmountNone ||
		effect.Amount.Value < 1 {
		return game.AbilityContent{}, false
	}
	selection, ok := cardSelectionForSelector(selector)
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.ExileFromGraveyardChoice(
			game.ControllerReference(),
			selection,
			game.Fixed(effect.Amount.Value),
			true,
			exiledWithSourceKey,
		),
	}}}.Ability(), true
}
