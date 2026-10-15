package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// LumraBellowOfTheWoods is the card definition for Lumra, Bellow of the Woods.
//
// Type: Legendary Creature — Elemental Bear
// Cost: {4}{G}{G}
//
// Oracle text:
//
//	Reach, vigilance
//	Lumra's power and toughness are each equal to the number of lands you control.
//	When Lumra enters, mill four cards. Then return all land cards from your graveyard to the battlefield tapped.
var LumraBellowOfTheWoods = newLumraBellowOfTheWoods

func newLumraBellowOfTheWoods() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Lumra, Bellow of the Woods",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
				cost.G,
			}),
			Colors:           []color.Color{color.Green},
			Supertypes:       []types.Super{types.Legendary},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Elemental, types.Bear},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerLandCount}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerLandCount}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
				game.VigilanceStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When Lumra enters, mill four cards. Then return all land cards from your graveyard to the battlefield tapped.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MoveTopOfLibrary{
											Amount:      game.Fixed(4),
											Player:      game.ControllerReference(),
											Destination: zone.Graveyard,
										},
									},
									game.Instruction{
										Primitive: game.MassReturnFromGraveyard{
											Player:      game.ControllerReference(),
											Selection:   game.Selection{RequiredTypes: []types.Card{types.Land}, Controller: game.ControllerYou},
											Destination: zone.Battlefield,
											EntryTapped: true,
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
			Reach, vigilance
			Lumra's power and toughness are each equal to the number of lands you control.
			When Lumra enters, mill four cards. Then return all land cards from your graveyard to the battlefield tapped.
		`,
		},
	}
}
