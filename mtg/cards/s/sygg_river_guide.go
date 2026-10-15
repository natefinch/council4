package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SyggRiverGuide is the card definition for Sygg, River Guide.
//
// Type: Legendary Creature — Merfolk Wizard
// Cost: {W}{U}
//
// Oracle text:
//
//	Islandwalk (This creature can't be blocked as long as defending player controls an Island.)
//	{1}{W}: Target Merfolk you control gains protection from the color of your choice until end of turn.
var SyggRiverGuide = newSyggRiverGuide

func newSyggRiverGuide() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Sygg, River Guide",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
				cost.U,
			}),
			Colors:     []color.Color{color.Blue, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Merfolk, types.Wizard},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.IslandwalkStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}{W}: Target Merfolk you control gains protection from the color of your choice until end of turn.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
						cost.W,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target Merfolk you control",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Merfolk}, Controller: game.ControllerYou}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ApplyContinuous{
											Object: opt.Val(game.TargetPermanentReference(0)),
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Duration: game.DurationPermanent,
													Layer:    game.LayerAbility,
													AddAbilities: []game.Ability{
														new(game.ProtectionFromChosenColorStaticAbility()),
													},
												},
											},
											Duration: game.DurationUntilEndOfTurn,
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
			Islandwalk (This creature can't be blocked as long as defending player controls an Island.)
			{1}{W}: Target Merfolk you control gains protection from the color of your choice until end of turn.
		`,
		},
	}
}
