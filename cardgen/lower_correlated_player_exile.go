package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

const correlatedOpponentExileKey = game.LinkedKey("correlated-opponent-exile")

func lowerEachOpponentGreatestPowerExile(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectExile ||
		!effect.ExileEachOpponentChoosesGreatestPower ||
		!effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		effect.Context != parser.EffectContextEachOpponent {
		return game.AbilityContent{}, false
	}
	if effect.Selector.Kind != compiler.SelectorCreature ||
		!onlyThatPlayerReferences(ctx.content.References) {
		return game.AbilityContent{}, false
	}
	selection, ok := SelectionForSelector(effect.Selector)
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.ExileForEachOpponent{
			Chooser:      game.GroupOfferMemberReference(),
			Selection:    selection,
			LinkedKey:    correlatedOpponentExileKey,
			Required:     true,
			Extremum:     game.PermanentChoiceGreatestPower,
			Simultaneous: true,
		},
	}}}.Ability(), true
}

func lowerEachOpponentCorrelatedExiledPowerDamage(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectDealDamage ||
		!effect.DamageEachOpponentCorrelatedExiledPower ||
		!effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		!correlatedExiledPowerReferences(effect.References) {
		return game.AbilityContent{}, false
	}
	member := game.GroupOfferMemberReference()
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.Damage{
			Amount: game.Dynamic(game.DynamicAmount{
				Kind:   game.DynamicAmountObjectPower,
				Object: game.LinkedObjectReference(string(correlatedOpponentExileKey)),
				Player: &member,
			}),
			Recipient: game.PlayerDamageRecipient(member),
		},
		ForEachPlayerGroup: opt.Val(game.OpponentsReference()),
	}}}.Ability(), true
}

func onlyThatPlayerReferences(references []compiler.CompiledReference) bool {
	for _, reference := range references {
		if reference.Kind != compiler.ReferenceThatPlayer {
			return false
		}
	}
	return true
}

func correlatedExiledPowerReferences(references []compiler.CompiledReference) bool {
	hasSource := false
	hasThey := false
	for _, reference := range references {
		switch {
		case reference.Kind == compiler.ReferenceSelfName || reference.Kind == compiler.ReferenceThisObject:
			hasSource = true
		case reference.Kind == compiler.ReferencePronoun && reference.Pronoun == compiler.ReferencePronounThey:
			hasThey = true
		default:
			return false
		}
	}
	return hasSource && hasThey
}
