package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SarkhanDragonsoul is the card definition for Sarkhan, Dragonsoul.
//
// Type: Legendary Planeswalker — Sarkhan
// Cost: {4}{R}{R}
//
// Oracle text:
//
//	+2: Sarkhan deals 1 damage to each opponent and each creature your opponents control.
//	−3: Sarkhan deals 4 damage to target player or planeswalker.
//	−9: Search your library for any number of Dragon creature cards, put them onto the battlefield, then shuffle.
var SarkhanDragonsoul = newSarkhanDragonsoul

func newSarkhanDragonsoul() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Sarkhan, Dragonsoul",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.R,
				cost.R,
			}),
			Colors:     []color.Color{color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Planeswalker},
			Subtypes:   []types.Sub{types.Sarkhan},
			Loyalty:    opt.Val(5),
			LoyaltyAbilities: []game.LoyaltyAbility{
				game.LoyaltyAbility{
					LoyaltyCost: 2,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:    game.Fixed(1),
									Recipient: game.PlayerGroupDamageRecipient(game.OpponentsReference()),
								},
							},
							{
								Primitive: game.Damage{
									Amount:    game.Fixed(1),
									Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerOpponent})),
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
								Constraint: "target player or planeswalker",
								Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Planeswalker}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:    game.Fixed(4),
									Recipient: game.AnyTargetDamageRecipient(0),
								},
							},
						},
					}.Ability(),
				},
				game.LoyaltyAbility{
					LoyaltyCost: -9,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Search{
									Player: game.ControllerReference(),
									Spec: game.SearchSpec{
										SourceZone:  zone.Library,
										Destination: zone.Battlefield,
										Filter:      game.Selection{RequiredTypes: []types.Card{types.Creature}, SubtypesAny: []types.Sub{types.Sub("Dragon")}},
										AnyNumber:   true,
									},
									Amount: game.Fixed(0),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			+2: Sarkhan deals 1 damage to each opponent and each creature your opponents control.
			−3: Sarkhan deals 4 damage to target player or planeswalker.
			−9: Search your library for any number of Dragon creature cards, put them onto the battlefield, then shuffle.
		`,
		},
	}
}
