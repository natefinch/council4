package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ReconCraftTheta is the card definition for Recon Craft Theta.
//
// Type: Artifact — Vehicle
// Cost: {4}
//
// Oracle text:
//
//	Flying
//	When this Vehicle enters, create a 0/0 blue Alien creature token. Put a +1/+1 counter on it.
//	Whenever this Vehicle attacks, proliferate. (Choose any number of permanents and/or players, then give each another counter of each kind already there.)
//	Crew 2
var ReconCraftTheta = newReconCraftTheta()

func newReconCraftTheta() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Recon Craft Theta",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types:     []types.Card{types.Artifact},
			Subtypes:  []types.Sub{types.Vehicle},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CrewActivatedAbility(2),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount:        game.Fixed(1),
									Source:        game.TokenDef(reconCraftThetaToken),
									PublishLinked: game.LinkedKey("created-token"),
								},
							},
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.LinkedObjectReference("created-token"),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventAttackerDeclared,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Proliferate{
									Amount: game.Fixed(1),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying
			When this Vehicle enters, create a 0/0 blue Alien creature token. Put a +1/+1 counter on it.
			Whenever this Vehicle attacks, proliferate. (Choose any number of permanents and/or players, then give each another counter of each kind already there.)
			Crew 2
		`,
		},
	}
}

var reconCraftThetaToken = newReconCraftThetaToken()

func newReconCraftThetaToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Alien",
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Alien},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 0}),
		},
	}
}
