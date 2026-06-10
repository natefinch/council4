package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Llanowar Elves
//
// Type: Token Creature — Elf Druid
//
// Oracle text:
//   {T}: Add {G}.

// LlanowarElvesTokeneaddc7cf5d58422aba006708dee2d853 is the card definition for Llanowar Elves.
var LlanowarElvesTokeneaddc7cf5d58422aba006708dee2d853 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name:      "Llanowar Elves",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Elf, types.Druid},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		ManaAbilities: []game.ManaAbility{
			game.TapManaAbility(mana.G),
		},
		OracleText: `
			{T}: Add {G}.
		`,
	},
}
