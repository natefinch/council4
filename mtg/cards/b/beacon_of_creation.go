package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BeaconOfCreation is the card definition for Beacon of Creation.
//
// Type: Sorcery
// Cost: {3}{G}
//
// Oracle text:
//
//	Create a 1/1 green Insect creature token for each Forest you control. Shuffle Beacon of Creation into its owner's library.
var BeaconOfCreation = newBeaconOfCreation

func newBeaconOfCreation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Beacon of Creation",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountCountSelector,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Forest")}, Controller: game.ControllerYou}),
							}),
							Source: game.TokenDef(beaconOfCreationToken),
						},
					},
					{
						Primitive: game.MoveResolvingSpell{
							Destination: zone.Library,
							Shuffle:     true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create a 1/1 green Insect creature token for each Forest you control. Shuffle Beacon of Creation into its owner's library.
		`,
		},
	}
}

var beaconOfCreationToken = newBeaconOfCreationToken()

func newBeaconOfCreationToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Insect",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Insect},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
