package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ChitteringDispatcher is the card definition for Chittering Dispatcher.
//
// Type: Creature — Eldrazi Drone
// Cost: {2}{G}
//
// Oracle text:
//
//	Devoid (This card has no color.)
//	Myriad
//	When this creature leaves the battlefield, create a 0/1 colorless Eldrazi Spawn creature token with "Sacrifice this token: Add {C}."
var ChitteringDispatcher = newChitteringDispatcher

func newChitteringDispatcher() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Chittering Dispatcher",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Eldrazi, types.Drone},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.DevoidStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.MyriadTriggeredBody,
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:         game.EventZoneChanged,
							Source:        game.TriggerSourceSelf,
							MatchFromZone: true,
							FromZone:      zone.Battlefield,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(chitteringDispatcherToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Devoid (This card has no color.)
			Myriad
			When this creature leaves the battlefield, create a 0/1 colorless Eldrazi Spawn creature token with "Sacrifice this token: Add {C}."
		`,
		},
	}
}

var chitteringDispatcherToken = newChitteringDispatcherToken()

func newChitteringDispatcherToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Eldrazi Spawn",
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Eldrazi, types.Spawn},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalSacrificeSource,
							Text:   "Sacrifice this token",
							Amount: 1,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddMana{
									Amount:    game.Fixed(1),
									ManaColor: mana.C,
								},
							},
						},
					}.Ability(),
				},
			},
		},
	}
}
