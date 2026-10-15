package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WildGrowth is the card definition for Wild Growth.
//
// Type: Enchantment — Aura
// Cost: {G}
//
// Oracle text:
//
//	Enchant land
//	Whenever enchanted land is tapped for mana, its controller adds an additional {G}.
var WildGrowth = newWildGrowth

func newWildGrowth() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Wild Growth",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
			}),
			Colors:   []color.Color{color.Green},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "land",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Land}}),
				}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever enchanted land is tapped for mana, its controller adds an additional {G}.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                game.EventPermanentTapped,
							Source:               game.TriggerSourceAttachedPermanent,
							SubjectSelection:     game.Selection{RequiredTypes: []types.Card{types.Land}},
							StackObjectKind:      game.StackSpell,
							CounterKind:          counter.PlusOnePlusOne,
							RequireTappedForMana: true,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddMana{
											Amount:    game.Fixed(1),
											ManaColor: mana.G,
											Player:    opt.Val(game.ObjectControllerReference(game.EventPermanentReference())),
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
			Enchant land
			Whenever enchanted land is tapped for mana, its controller adds an additional {G}.
		`,
		},
	}
}
