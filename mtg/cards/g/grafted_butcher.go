package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GraftedButcher is the card definition for Grafted Butcher.
//
// Type: Creature — Phyrexian Samurai
// Cost: {1}{B}
//
// Oracle text:
//
//	When this creature enters, Phyrexians you control gain menace until end of turn.
//	Other Phyrexians you control get +1/+1.
//	{3}{B}, Sacrifice an artifact or creature: Return this card from your graveyard to the battlefield. Activate only as a sorcery.
var GraftedButcher = newGraftedButcher

func newGraftedButcher() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Grafted Butcher",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Samurai},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Phyrexian")}}, game.SourcePermanentReference()),
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{3}{B}, Sacrifice an artifact or creature: Return this card from your graveyard to the battlefield. Activate only as a sorcery.",
					ManaCost: opt.Val(cost.Mana{cost.O(3), cost.B}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:               cost.AdditionalSacrifice,
							Text:               "Sacrifice an artifact or creature",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Artifact,
							PermanentTypeAlt:   types.Creature,
						},
					},
					ZoneOfFunction: zone.Graveyard,
					Timing:         game.SorceryOnly,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.PutOnBattlefield{
									Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceSource}),
								},
							},
						},
					}.Ability(),
				},
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
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											Group: game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Phyrexian")}, Controller: game.ControllerYou}),
											AddKeywords: []game.Keyword{
												game.Menace,
											},
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
			When this creature enters, Phyrexians you control gain menace until end of turn.
			Other Phyrexians you control get +1/+1.
			{3}{B}, Sacrifice an artifact or creature: Return this card from your graveyard to the battlefield. Activate only as a sorcery.
		`,
		},
	}
}
