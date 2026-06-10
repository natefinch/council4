package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Elf Druid
//
// Type: Token Creature — Elf Druid
//
// Oracle text:
//   {T}: Add {G}.

// ElfDruidTokenc89c4a7131784243a9fb5eda544591ea is the card definition for Elf Druid.
var ElfDruidTokenc89c4a7131784243a9fb5eda544591ea = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name:      "Elf Druid",
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
