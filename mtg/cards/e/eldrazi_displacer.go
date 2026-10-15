package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// EldraziDisplacer is the card definition for Eldrazi Displacer.
//
// Type: Creature — Eldrazi
// Cost: {2}{W}
//
// Oracle text:
//
//	Devoid (This card has no color.)
//	{2}{C}: Exile another target creature, then return it to the battlefield tapped under its owner's control. ({C} represents colorless mana.)
var EldraziDisplacer = newEldraziDisplacer

func newEldraziDisplacer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Eldrazi Displacer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Eldrazi},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.DevoidStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{2}{C}: Exile another target creature, then return it to the battlefield tapped under its owner's control. ({C} represents colorless mana.)",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
						cost.C,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "another target creature",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, ExcludeSource: true}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MovePermanent{
											Object:        game.TargetPermanentReference(0),
											Destination:   zone.Exile,
											PublishLinked: game.LinkedKey("blink-1"),
										},
									},
									game.Instruction{
										Primitive: game.PutOnBattlefield{
											Source:      game.LinkedBattlefieldSource(game.LinkedKey("blink-1")),
											EntryTapped: true,
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
			Devoid (This card has no color.)
			{2}{C}: Exile another target creature, then return it to the battlefield tapped under its owner's control. ({C} represents colorless mana.)
		`,
		},
	}
}
