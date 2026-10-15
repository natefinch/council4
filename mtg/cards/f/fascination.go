package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Fascination is the card definition for Fascination.
//
// Type: Sorcery
// Cost: {X}{U}{U}
//
// Oracle text:
//
//	Choose one —
//	• Each player draws X cards.
//	• Each player mills X cards.
var Fascination = newFascination

func newFascination() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Fascination",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Each player draws X cards.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind: game.DynamicAmountX,
									}),
									PlayerGroup: game.AllPlayersReference(),
								},
							},
						},
					},
					game.Mode{
						Text: "Each player mills X cards.",
						Sequence: []game.Instruction{
							{
								Primitive: game.MoveTopOfLibrary{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind: game.DynamicAmountX,
									}),
									PlayerGroup: game.AllPlayersReference(),
									Destination: zone.Graveyard,
								},
							},
						},
					},
				},
				MinModes: 1,
				MaxModes: 1,
			}),
			OracleText: `
			Choose one —
			• Each player draws X cards.
			• Each player mills X cards.
		`,
		},
	}
}
