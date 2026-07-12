package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CastleGarenbrig is the card definition for Castle Garenbrig.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped unless you control a Forest.
//	{T}: Add {G}.
//	{2}{G}{G}, {T}: Add six {G}. Spend this mana only to cast creature spells or activate abilities of creatures.
var CastleGarenbrig = newCastleGarenbrig

func newCastleGarenbrig() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:  "Castle Garenbrig",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.G),
				game.ManaAbility{
					ManaCost:        opt.Val(cost.Mana{cost.O(2), cost.G, cost.G}),
					AdditionalCosts: cost.Tap,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddMana{
									Amount:    game.Fixed(1),
									ManaColor: mana.G,
									SpendRider: opt.Val(game.ManaSpendRider{
										Condition:   game.ManaSpendCastOrActivateCreature,
										Restriction: game.ManaSpendRestrictedToCondition,
									}),
								},
							},
							{
								Primitive: game.AddMana{
									Amount:    game.Fixed(1),
									ManaColor: mana.G,
									SpendRider: opt.Val(game.ManaSpendRider{
										Condition:   game.ManaSpendCastOrActivateCreature,
										Restriction: game.ManaSpendRestrictedToCondition,
									}),
								},
							},
							{
								Primitive: game.AddMana{
									Amount:    game.Fixed(1),
									ManaColor: mana.G,
									SpendRider: opt.Val(game.ManaSpendRider{
										Condition:   game.ManaSpendCastOrActivateCreature,
										Restriction: game.ManaSpendRestrictedToCondition,
									}),
								},
							},
							{
								Primitive: game.AddMana{
									Amount:    game.Fixed(1),
									ManaColor: mana.G,
									SpendRider: opt.Val(game.ManaSpendRider{
										Condition:   game.ManaSpendCastOrActivateCreature,
										Restriction: game.ManaSpendRestrictedToCondition,
									}),
								},
							},
							{
								Primitive: game.AddMana{
									Amount:    game.Fixed(1),
									ManaColor: mana.G,
									SpendRider: opt.Val(game.ManaSpendRider{
										Condition:   game.ManaSpendCastOrActivateCreature,
										Restriction: game.ManaSpendRestrictedToCondition,
									}),
								},
							},
							{
								Primitive: game.AddMana{
									Amount:    game.Fixed(1),
									ManaColor: mana.G,
									SpendRider: opt.Val(game.ManaSpendRider{
										Condition:   game.ManaSpendCastOrActivateCreature,
										Restriction: game.ManaSpendRestrictedToCondition,
									}),
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedIfReplacement("This land enters tapped unless you control a Forest.", &game.Condition{
					Negate: true,
					ControlsMatching: opt.Val(game.SelectionCount{
						Selection: game.Selection{SubtypesAny: []types.Sub{types.Sub("Forest")}},
					}),
				}),
			},
			OracleText: `
			This land enters tapped unless you control a Forest.
			{T}: Add {G}.
			{2}{G}{G}, {T}: Add six {G}. Spend this mana only to cast creature spells or activate abilities of creatures.
		`,
		},
	}
}
