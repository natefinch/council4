package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// KuraTheBoundlessSky is the card definition for Kura, the Boundless Sky.
//
// Type: Legendary Creature — Dragon Spirit
// Cost: {3}{G}{G}
//
// Oracle text:
//
//	Flying, deathtouch
//	When Kura dies, choose one —
//	• Search your library for up to three land cards, reveal them, put them into your hand, then shuffle.
//	• Create an X/X green Spirit creature token, where X is the number of lands you control.
var KuraTheBoundlessSky = newKuraTheBoundlessSky()

func newKuraTheBoundlessSky() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Kura, the Boundless Sky",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Dragon, types.Spirit},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.DeathtouchStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							Source:           game.TriggerSourceSelf,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "Search your library for up to three land cards, reveal them, put them into your hand, then shuffle.",
								Sequence: []game.Instruction{
									{
										Primitive: game.Search{
											Player: game.ControllerReference(),
											Spec: game.SearchSpec{
												SourceZone:  zone.Library,
												Destination: zone.Hand,
												Filter:      game.Selection{RequiredTypes: []types.Card{types.Land}},
												Reveal:      true,
											},
											Amount: game.Fixed(3),
										},
									},
								},
							},
							game.Mode{
								Text: "Create an X/X green Spirit creature token, where X is the number of lands you control.",
								Sequence: []game.Instruction{
									{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(kuraTheBoundlessSkyToken),
											Power: opt.Val(game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountCountSelector,
												Multiplier: 1,
												Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Land}, Controller: game.ControllerYou}),
											})),
											Toughness: opt.Val(game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountCountSelector,
												Multiplier: 1,
												Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Land}, Controller: game.ControllerYou}),
											})),
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
			Flying, deathtouch
			When Kura dies, choose one —
			• Search your library for up to three land cards, reveal them, put them into your hand, then shuffle.
			• Create an X/X green Spirit creature token, where X is the number of lands you control.
		`,
		},
	}
}

var kuraTheBoundlessSkyToken = newKuraTheBoundlessSkyToken()

func newKuraTheBoundlessSkyToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Spirit",
			Colors:   []color.Color{color.Green},
			Types:    []types.Card{types.Creature},
			Subtypes: []types.Sub{types.Spirit},
		},
	}
}
