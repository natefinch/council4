package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TyvarThePummeler is the card definition for Tyvar, the Pummeler.
//
// Type: Legendary Creature — Elf Warrior
// Cost: {1}{G}{G}
//
// Oracle text:
//
//	Tap another untapped creature you control: Tyvar gains indestructible until end of turn. Tap it.
//	{3}{G}{G}: Creatures you control get +X/+X until end of turn, where X is the greatest power among creatures you control.
var TyvarThePummeler = newTyvarThePummeler()

func newTyvarThePummeler() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Tyvar, the Pummeler",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Elf, types.Warrior},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Tap another untapped creature you control: Tyvar gains indestructible until end of turn. Tap it.",
					AdditionalCosts: []cost.Additional{
						{
							Kind:               cost.AdditionalTapPermanents,
							Text:               "Tap another untapped creature you control",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Creature,
							ExcludeSource:      true,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.SourceCardPermanentReference()),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											AddKeywords: []game.Keyword{
												game.Indestructible,
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
							{
								Primitive: game.Tap{
									Object: game.SourcePermanentReference(),
								},
							},
						},
					}.Ability(),
				},
				game.ActivatedAbility{
					Text:           "{3}{G}{G}: Creatures you control get +X/+X until end of turn, where X is the greatest power among creatures you control.",
					ManaCost:       opt.Val(cost.Mana{cost.O(3), cost.G, cost.G}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerPowerToughnessModify,
											Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
											PowerDeltaDynamic: opt.Val(game.DynamicAmount{
												Kind:       game.DynamicAmountGreatestPowerInGroup,
												Multiplier: 1,
												Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
											}),
											ToughnessDeltaDynamic: opt.Val(game.DynamicAmount{
												Kind:       game.DynamicAmountGreatestPowerInGroup,
												Multiplier: 1,
												Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
											}),
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Tap another untapped creature you control: Tyvar gains indestructible until end of turn. Tap it.
			{3}{G}{G}: Creatures you control get +X/+X until end of turn, where X is the greatest power among creatures you control.
		`,
		},
	}
}
