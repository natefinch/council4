package cardgen

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
)

func lowerFixedManaAction(ctx contentCtx, effect compiler.CompiledEffect) (game.AbilityContent, *shared.Diagnostic) {
	colors := effect.Mana.Colors
	if !ctx.singleAction {
		return manaFixedContent(colors), nil
	}
	if !slices.ContainsFunc(colors, func(c mana.Color) bool { return c != colors[0] }) {
		return game.Mode{Sequence: []game.Instruction{{
			Primitive: game.AddMana{Amount: game.Fixed(len(colors)), ManaColor: colors[0]},
		}}}.Ability(), nil
	}
	return manaFixedContent(colors), nil
}
