package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// EarthquakeDragon is the card definition for Earthquake Dragon.
//
// Type: Creature — Elemental Dragon
// Cost: {14}{G}
//
// Oracle text:
//
//	This spell costs {X} less to cast, where X is the total mana value of Dragons you control.
//	Flying, trample
//	{2}{G}, Sacrifice a land: Return this card from your graveyard to your hand.
var EarthquakeDragon = newEarthquakeDragon

func newEarthquakeDragon() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Earthquake Dragon",
			ManaCost: opt.Val(cost.Mana{
				cost.O(14),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental, types.Dragon},
			Power:     opt.Val(game.PT{Value: 10}),
			Toughness: opt.Val(game.PT{Value: 10}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This spell costs {X} less to cast, where X is the total mana value of Dragons you control.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
							CostModifier: game.CostModifier{
								Kind: game.CostModifierSpell,
								DynamicReduction: &game.DynamicAmount{
									Kind:        game.DynamicAmountTotalManaValueInGroup,
									Multiplier:  1,
									CounterKind: counter.PlusOnePlusOne,
									Group:       game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Dragon")}, Controller: game.ControllerYou}),
								},
							},
						},
					},
				},
				game.FlyingStaticBody,
				game.TrampleStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{2}{G}, Sacrifice a land: Return this card from your graveyard to your hand.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
						cost.G,
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:               cost.AdditionalSacrifice,
							Text:               "Sacrifice a land",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Land,
							CounterKind:        counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Graveyard,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MoveCard{
											Card: game.CardReference{
												Kind: game.CardReferenceSource,
											},
											FromZone:    zone.Graveyard,
											Destination: zone.Hand,
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
			This spell costs {X} less to cast, where X is the total mana value of Dragons you control.
			Flying, trample
			{2}{G}, Sacrifice a land: Return this card from your graveyard to your hand.
		`,
		},
	}
}
