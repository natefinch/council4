package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BrilliantRestoration is the card definition for Brilliant Restoration.
//
// Type: Sorcery
// Cost: {3}{W}{W}{W}{W}
//
// Oracle text:
//
//	Return all artifact and enchantment cards from your graveyard to the battlefield.
var BrilliantRestoration = newBrilliantRestoration

func newBrilliantRestoration() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Brilliant Restoration",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
				cost.W,
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
							Selection:   game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment}, Controller: game.ControllerYou},
							Destination: zone.Battlefield,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Return all artifact and enchantment cards from your graveyard to the battlefield.
		`,
		},
	}
}
