package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ReturnToTheLightRealms is the card definition for Return to the Light Realms.
//
// Type: Sorcery
// Cost: {7}{W}{W}
//
// Oracle text:
//
//	Return all nonland permanent cards from your graveyard to the battlefield.
var ReturnToTheLightRealms = newReturnToTheLightRealms

func newReturnToTheLightRealms() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Return to the Light Realms",
			ManaCost: opt.Val(cost.Mana{
				cost.O(7),
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.MassReturnFromGraveyard{
							Player:      game.ControllerReference(),
							Selection:   game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature, types.Enchantment, types.Land, types.Planeswalker, types.Battle}, ExcludedTypes: []types.Card{types.Land}, Controller: game.ControllerYou},
							Destination: zone.Battlefield,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Return all nonland permanent cards from your graveyard to the battlefield.
		`,
		},
	}
}
