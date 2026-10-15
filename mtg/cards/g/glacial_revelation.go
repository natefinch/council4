package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GlacialRevelation is the card definition for Glacial Revelation.
//
// Type: Sorcery
// Cost: {2}{G}
//
// Oracle text:
//
//	Reveal the top six cards of your library. You may put any number of snow permanent cards from among them into your hand. Put the rest into your graveyard.
var GlacialRevelation = newGlacialRevelation

func newGlacialRevelation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Glacial Revelation",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Dig{
							Player:   game.ControllerReference(),
							Look:     game.Fixed(6),
							Take:     game.Fixed(6),
							Filter:   opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature, types.Enchantment, types.Land, types.Planeswalker, types.Battle}, Supertypes: []types.Super{types.Snow}}),
							TakeUpTo: true,
							Reveal:   true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Reveal the top six cards of your library. You may put any number of snow permanent cards from among them into your hand. Put the rest into your graveyard.
		`,
		},
	}
}
