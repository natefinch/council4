package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Nevermaker is the card definition for Nevermaker.
var Nevermaker = newNevermaker()

func newNevermaker() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Nevermaker",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:         game.EventZoneChanged,
							Source:        game.TriggerSourceSelf,
							MatchFromZone: true,
							FromZone:      zone.Battlefield,
						},
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target nonland permanent",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{ExcludedTypes: []types.Card{types.Land}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.PutPermanentOnLibrary{
									Object: game.TargetPermanentReference(0),
								},
							},
						},
					}.Ability(),
				},
				game.EvokeSacrificeTriggeredAbility(),
			},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label:    "Evoke",
					ManaCost: opt.Val(cost.Mana{cost.O(3), cost.U}),
					Mechanic: cost.AlternativeMechanicEvoke,
				},
			},
			OracleText: `
			Flying
			When this creature leaves the battlefield, put target nonland permanent on top of its owner's library.
			Evoke {3}{U} (You may cast this spell for its evoke cost. If you do, it's sacrificed when it enters.)
		`,
		},
	}
}
