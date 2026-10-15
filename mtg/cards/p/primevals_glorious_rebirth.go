package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// PrimevalsGloriousRebirth is the card definition for Primevals' Glorious Rebirth.
//
// Type: Legendary Sorcery
// Cost: {5}{W}{B}
//
// Oracle text:
//
//	(You may cast a legendary sorcery only if you control a legendary creature or planeswalker.)
//	Return all legendary permanent cards from your graveyard to the battlefield.
var PrimevalsGloriousRebirth = newPrimevalsGloriousRebirth

func newPrimevalsGloriousRebirth() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name: "Primevals' Glorious Rebirth",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.W,
				cost.B,
			}),
			Colors:     []color.Color{color.Black, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.MassReturnFromGraveyard{
							Player:      game.ControllerReference(),
							Selection:   game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature, types.Enchantment, types.Land, types.Planeswalker, types.Battle}, Supertypes: []types.Super{types.Legendary}, Controller: game.ControllerYou},
							Destination: zone.Battlefield,
						},
					},
				},
			}.Ability()),
			OracleText: `
			(You may cast a legendary sorcery only if you control a legendary creature or planeswalker.)
			Return all legendary permanent cards from your graveyard to the battlefield.
		`,
		},
	}
}
