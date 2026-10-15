package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AkromaSVengeance is the card definition for Akroma's Vengeance.
//
// Type: Sorcery
// Cost: {4}{W}{W}
//
// Oracle text:
//
//	Destroy all artifacts, creatures, and enchantments.
//	Cycling {3} ({3}, Discard this card: Draw a card.)
var AkromaSVengeance = newAkromaSVengeance

func newAkromaSVengeance() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Akroma's Vengeance",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(3)}),
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Group: game.BattlefieldGroup(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature, types.Enchantment}}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy all artifacts, creatures, and enchantments.
			Cycling {3} ({3}, Discard this card: Draw a card.)
		`,
		},
	}
}
