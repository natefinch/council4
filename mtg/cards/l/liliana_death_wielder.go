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

// LilianaDeathWielder is the card definition for Liliana, Death Wielder.
var LilianaDeathWielder = newLilianaDeathWielder()

func newLilianaDeathWielder() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Liliana, Death Wielder",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.B,
				cost.B,
			}),
			Colors:     []color.Color{color.Black},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Planeswalker},
			Subtypes:   []types.Sub{types.Liliana},
			Loyalty:    opt.Val(5),
			LoyaltyAbilities: []game.LoyaltyAbility{
				game.LoyaltyAbility{
					LoyaltyCost: 2,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 0,
								MaxTargets: 1,
								Constraint: "up to one target creature",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.TargetPermanentReference(0),
									CounterKind: counter.MinusOneMinusOne,
								},
							},
						},
					}.Ability(),
				},
				game.LoyaltyAbility{
					LoyaltyCost: -3,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature with a -1/-1 counter on it",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, MatchCounter: true, RequiredCounter: counter.MinusOneMinusOne}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Destroy{
									Object: game.TargetPermanentReference(0),
								},
							},
						},
					}.Ability(),
				},
				game.LoyaltyAbility{
					LoyaltyCost: -10,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.MassReturnFromGraveyard{
									Player:      game.ControllerReference(),
									Selection:   game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou},
									Destination: zone.Battlefield,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			+2: Put a -1/-1 counter on up to one target creature.
			−3: Destroy target creature with a -1/-1 counter on it.
			−10: Return all creature cards from your graveyard to the battlefield.
		`,
		},
	}
}
