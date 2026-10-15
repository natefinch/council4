package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// VilespawnSpider is the card definition for Vilespawn Spider.
//
// Type: Creature — Spider
// Cost: {G}{U}
//
// Oracle text:
//
//	Reach
//	At the beginning of your upkeep, mill a card. (Put the top card of your library into your graveyard.)
//	{2}{G}{U}, {T}, Sacrifice this creature: Create a 1/1 green Insect creature token for each creature card in your graveyard. Activate only as a sorcery.
var VilespawnSpider = newVilespawnSpider

func newVilespawnSpider() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Vilespawn Spider",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
				cost.U,
			}),
			Colors:    []color.Color{color.Green, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spider},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{2}{G}{U}, {T}, Sacrifice this creature: Create a 1/1 green Insect creature token for each creature card in your graveyard. Activate only as a sorcery.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
						cost.G,
						cost.U,
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:        cost.AdditionalSacrificeSource,
							Text:        "Sacrifice this creature",
							Amount:      1,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Timing:         game.SorceryOnly,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountCountCardsInZone,
												Multiplier: 1,
												Player:     func() *game.PlayerReference { ref := game.ControllerReference(); return &ref }(),
												CardZone:   zone.Graveyard,
												Selection:  &game.Selection{RequiredTypes: []types.Card{types.Creature}},
											}),
											Source: game.TokenDef(vilespawnSpiderToken),
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
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "At the beginning of your upkeep, mill a card. (Put the top card of your library into your graveyard.)",
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:           game.EventBeginningOfStep,
							Controller:      game.TriggerControllerYou,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
							Step:            game.StepUpkeep,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MoveTopOfLibrary{
											Amount:      game.Fixed(1),
											Player:      game.ControllerReference(),
											Destination: zone.Graveyard,
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
			Reach
			At the beginning of your upkeep, mill a card. (Put the top card of your library into your graveyard.)
			{2}{G}{U}, {T}, Sacrifice this creature: Create a 1/1 green Insect creature token for each creature card in your graveyard. Activate only as a sorcery.
		`,
		},
	}
}

var vilespawnSpiderToken = newVilespawnSpiderToken()

func newVilespawnSpiderToken() *game.CardDef {
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
