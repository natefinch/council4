package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Mindmelter is the card definition for Mindmelter.
//
// Type: Creature — Eldrazi Drone
// Cost: {1}{U}{B}
//
// Oracle text:
//
//	Devoid (This card has no color.)
//	This creature can't be blocked.
//	{3}{C}: Target opponent exiles a card from their hand. Activate only as a sorcery. ({C} represents colorless mana.)
var Mindmelter = newMindmelter

func newMindmelter() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Mindmelter",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
				cost.B,
			}),
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Eldrazi, types.Drone},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.DevoidStaticBody,
				game.CantBeBlockedStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{3}{C}: Target opponent exiles a card from their hand. Activate only as a sorcery. ({C} represents colorless mana.)",
					ManaCost: opt.Val(cost.Mana{
						cost.O(3),
						cost.C,
					}),
					ZoneOfFunction: zone.Battlefield,
					Timing:         game.SorceryOnly,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "Target opponent",
										Allow:      game.TargetAllowPlayer,
										Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ChooseFromZone{
											Player:     game.TargetPlayerReference(0),
											SourceZone: zone.Hand,
											Quantity:   game.Fixed(1),
											Destination: game.ChooseDestination{
												Zone: zone.Exile,
											},
											Riders: game.ChooseRiders{
												PublishObjectScoped: true,
											},
											Prompt: "Choose a card to exile",
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			Devoid (This card has no color.)
			This creature can't be blocked.
			{3}{C}: Target opponent exiles a card from their hand. Activate only as a sorcery. ({C} represents colorless mana.)
		`,
		},
	}
}
