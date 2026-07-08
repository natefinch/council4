package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MysticMight is the card definition for Mystic Might.
//
// Type: Enchantment — Aura
// Cost: {U}
//
// Oracle text:
//
//	Enchant land you control
//	Cumulative upkeep {1}{U} (At the beginning of your upkeep, put an age counter on this permanent, then sacrifice it unless you pay its upkeep cost for each age counter on it.)
//	Enchanted land has "{T}: Target creature gets +2/+2 until end of turn."
var MysticMight = newMysticMight

func newMysticMight() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Mystic Might",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
			}),
			Colors:   []color.Color{color.Blue},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "land you control",
					Allow:      game.TargetAllowPermanent,
					Selection: opt.Val(game.Selection{
						RequiredTypesAny: []types.Card{types.Land},
						Controller:       game.ControllerYou,
					}),
				}),
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							AddAbilities: []game.Ability{
								new(game.ActivatedAbility{
									Text:            "{T}: Target creature gets +2/+2 until end of turn.",
									AdditionalCosts: cost.Tap,
									ZoneOfFunction:  zone.Battlefield,
									Content: game.Mode{
										Targets: []game.TargetSpec{
											game.TargetSpec{
												MinTargets: 1,
												MaxTargets: 1,
												Constraint: "target creature",
												Allow:      game.TargetAllowPermanent,
												Selection: opt.Val(game.Selection{
													RequiredTypesAny: []types.Card{types.Creature},
												}),
											},
										},
										Sequence: []game.Instruction{
											{
												Primitive: game.ModifyPT{
													Object:         game.TargetPermanentReference(0),
													PowerDelta:     game.Fixed(2),
													ToughnessDelta: game.Fixed(2),
													Duration:       game.DurationUntilEndOfTurn,
												},
											},
										},
									}.Ability(),
								}),
							},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.CumulativeUpkeepTriggeredAbility(cost.Mana{cost.O(1), cost.U}),
			},
			OracleText: `
			Enchant land you control
			Cumulative upkeep {1}{U} (At the beginning of your upkeep, put an age counter on this permanent, then sacrifice it unless you pay its upkeep cost for each age counter on it.)
			Enchanted land has "{T}: Target creature gets +2/+2 until end of turn."
		`,
		},
	}
}
