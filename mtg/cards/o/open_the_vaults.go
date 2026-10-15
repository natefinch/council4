package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// OpenTheVaults is the card definition for Open the Vaults.
//
// Type: Sorcery
// Cost: {4}{W}{W}
//
// Oracle text:
//
//	Return all artifact and enchantment cards from all graveyards to the battlefield under their owners' control. (Auras with nothing to enchant remain in graveyards.)
var OpenTheVaults = newOpenTheVaults

func newOpenTheVaults() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Open the Vaults",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.MassReturnFromGraveyard{
							Player:            game.ControllerReference(),
							Selection:         game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment}},
							Destination:       zone.Battlefield,
							SourceGroup:       game.AllPlayersReference(),
							ControlledByOwner: true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Return all artifact and enchantment cards from all graveyards to the battlefield under their owners' control. (Auras with nothing to enchant remain in graveyards.)
		`,
		},
	}
}
