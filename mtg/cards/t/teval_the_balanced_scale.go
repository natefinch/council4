package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TevalTheBalancedScale is the card definition for Teval, the Balanced Scale.
//
// Type: Legendary Creature — Spirit Dragon
// Cost: {1}{B}{G}{U}
//
// Oracle text:
//
//	Flying
//	Whenever Teval attacks, mill three cards. Then you may return a land card from your graveyard to the battlefield tapped.
//	Whenever one or more cards leave your graveyard, create a 2/2 black Zombie Druid creature token.
var TevalTheBalancedScale = newTevalTheBalancedScale

func newTevalTheBalancedScale() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Teval, the Balanced Scale",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
				cost.G,
				cost.U,
			}),
			Colors:     []color.Color{color.Black, color.Green, color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Spirit, types.Dragon},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever Teval attacks, mill three cards. Then you may return a land card from your graveyard to the battlefield tapped.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventAttackerDeclared,
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
											Amount:      game.Fixed(3),
											Player:      game.ControllerReference(),
											Destination: zone.Graveyard,
										},
									},
									game.Instruction{
										Primitive: game.ChooseFromZone{
											Player:     game.ControllerReference(),
											SourceZone: zone.Graveyard,
											Filter:     game.Selection{RequiredTypes: []types.Card{types.Land}, Controller: game.ControllerYou},
											Quantity:   game.Fixed(1),
											Destination: game.ChooseDestination{
												Zone: zone.Battlefield,
											},
											Riders: game.ChooseRiders{
												EntersTapped: true,
											},
											Prompt: "Choose a card to return to the battlefield",
										},
										Optional: true,
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.TriggeredAbility{
					Text: "Whenever one or more cards leave your graveyard, create a 2/2 black Zombie Druid creature token.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventZoneChanged,
							Player:           game.TriggerPlayerYou,
							SubjectSelection: game.Selection{NonToken: true},
							MatchFromZone:    true,
							FromZone:         zone.Graveyard,
							StackObjectKind:  game.StackSpell,
							OneOrMore:        true,
							CounterKind:      counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(tevalTheBalancedScaleToken),
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
			Flying
			Whenever Teval attacks, mill three cards. Then you may return a land card from your graveyard to the battlefield tapped.
			Whenever one or more cards leave your graveyard, create a 2/2 black Zombie Druid creature token.
		`,
		},
	}
}

var tevalTheBalancedScaleToken = newTevalTheBalancedScaleToken()

func newTevalTheBalancedScaleToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Zombie Druid",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie, types.Druid},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
