package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AetherChanneler is the card definition for Aether Channeler.
//
// Type: Creature — Human Wizard
// Cost: {2}{U}
//
// Oracle text:
//
//	When this creature enters, choose one —
//	• Create a 1/1 white Bird creature token with flying.
//	• Return another target nonland permanent to its owner's hand.
//	• Draw a card.
var AetherChanneler = newAetherChanneler()

func newAetherChanneler() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Aether Channeler",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Wizard},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "Create a 1/1 white Bird creature token with flying.",
								Sequence: []game.Instruction{
									{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(aetherChannelerToken),
										},
									},
								},
							},
							game.Mode{
								Text: "Return another target nonland permanent to its owner's hand.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "another target nonland permanent",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{ExcludedTypes: []types.Card{types.Land}, ExcludeSource: true}),
									},
								},
								Sequence: []game.Instruction{
									{
										Primitive: game.Bounce{
											Object: game.TargetPermanentReference(0),
										},
									},
								},
							},
							game.Mode{
								Text: "Draw a card.",
								Sequence: []game.Instruction{
									{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
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
			When this creature enters, choose one —
			• Create a 1/1 white Bird creature token with flying.
			• Return another target nonland permanent to its owner's hand.
			• Draw a card.
		`,
		},
	}
}

var aetherChannelerToken = newAetherChannelerToken()

func newAetherChannelerToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Bird",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bird},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}
