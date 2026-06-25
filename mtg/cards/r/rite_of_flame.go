package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RiteOfFlame is the card definition for Rite of Flame.
//
// Type: Sorcery
// Cost: {R}
//
// Oracle text:
//
//	Add {R}{R}, then add {R} for each card named Rite of Flame in each graveyard.
var RiteOfFlame = newRiteOfFlame()

func newRiteOfFlame() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Rite of Flame",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.AddMana{
							Amount:    game.Fixed(1),
							ManaColor: mana.R,
						},
					},
					{
						Primitive: game.AddMana{
							Amount:    game.Fixed(1),
							ManaColor: mana.R,
						},
					},
					{
						Primitive: game.AddMana{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountCardsNamedSourceInGraveyards,
								Multiplier: 1,
							}),
							ManaColor: mana.R,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Add {R}{R}, then add {R} for each card named Rite of Flame in each graveyard.
		`,
		},
	}
}
