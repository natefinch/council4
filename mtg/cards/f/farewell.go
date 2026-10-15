package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Farewell is the card definition for Farewell.
//
// Type: Sorcery
// Cost: {4}{W}{W}
//
// Oracle text:
//
//	Choose one or more —
//	• Exile all artifacts.
//	• Exile all creatures.
//	• Exile all enchantments.
//	• Exile all graveyards.
var Farewell = newFarewell

func newFarewell() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Farewell",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Exile all artifacts.",
						Sequence: []game.Instruction{
							{
								Primitive: game.MovePermanent{
									Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Artifact}}),
									Destination: zone.Exile,
								},
							},
						},
					},
					game.Mode{
						Text: "Exile all creatures.",
						Sequence: []game.Instruction{
							{
								Primitive: game.MovePermanent{
									Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
									Destination: zone.Exile,
								},
							},
						},
					},
					game.Mode{
						Text: "Exile all enchantments.",
						Sequence: []game.Instruction{
							{
								Primitive: game.MovePermanent{
									Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Enchantment}}),
									Destination: zone.Exile,
								},
							},
						},
					},
					game.Mode{
						Text: "Exile all graveyards.",
						Sequence: []game.Instruction{
							{
								Primitive: game.MoveCard{
									PlayerGroup: game.AllPlayersReference(),
									FromZone:    zone.Graveyard,
									Destination: zone.Exile,
								},
							},
						},
					},
				},
				MinModes: 1,
				MaxModes: 4,
			}),
			OracleText: `
			Choose one or more —
			• Exile all artifacts.
			• Exile all creatures.
			• Exile all enchantments.
			• Exile all graveyards.
		`,
		},
	}
}
