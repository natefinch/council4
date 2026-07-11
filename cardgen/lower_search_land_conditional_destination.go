package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const searchLandConditionalKey = game.LinkedKey("search-land-conditional")

func lowerSearchLandConditionalDestination(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 5 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 1 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	search, reveal, landPut, handPut, shuffle := ctx.content.Effects[0], ctx.content.Effects[1], ctx.content.Effects[2], ctx.content.Effects[3], ctx.content.Effects[4]
	if !search.SearchLandElseHand ||
		search.Kind != compiler.EffectSearch ||
		reveal.Kind != compiler.EffectReveal ||
		landPut.Kind != compiler.EffectPut ||
		landPut.ToZone != zone.Battlefield ||
		!landPut.EntersTapped ||
		handPut.Kind != compiler.EffectPut ||
		handPut.ToZone != zone.Hand ||
		shuffle.Kind != compiler.EffectShuffle {
		return game.AbilityContent{}, false
	}
	spec, ok := searchSpecForSelector(search.Selector)
	if !ok {
		return game.AbilityContent{}, false
	}
	amount, ok := searchAmountQuantity(search)
	if !ok || amount.Value() != 1 {
		return game.AbilityContent{}, false
	}
	spec.SourceZone = zone.Library
	spec.Reveal = true
	spec.RevealOnly = true
	card := game.CardReference{Kind: game.CardReferenceLinked, LinkID: string(searchLandConditionalKey)}
	return game.Mode{Sequence: []game.Instruction{
		{Primitive: game.Search{
			Player:        game.ControllerReference(),
			Spec:          spec,
			Amount:        amount,
			PublishLinked: searchLandConditionalKey,
		}},
		{Primitive: game.ConditionalDestinationPlace{
			Card:     card,
			FromZone: zone.Library,
			CardCondition: opt.Val(game.CardSelection{
				Card:      card,
				Selection: game.Selection{RequiredTypes: []types.Card{types.Land}},
			}),
			EntryTapped:   true,
			ThenMandatory: true,
			Else:          zone.Hand,
		}},
		{Primitive: game.ShuffleLibrary{Player: game.ControllerReference()}},
	}}.Ability(), true
}
