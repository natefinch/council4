package cardgen

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func damageableSelection() game.Selection {
	return game.Selection{RequiredTypesAny: []types.Card{
		types.Creature, types.Planeswalker, types.Battle,
	}}
}
