package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AvalancheRiders is the card definition for Avalanche Riders.
//
// Type: Creature — Human Nomad
// Cost: {3}{R}
//
// Oracle text:
//
//	Haste
//	Echo {3}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	When this creature enters, destroy target land.
var AvalancheRiders = newAvalancheRiders

func newAvalancheRiders() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Avalanche Riders",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Nomad},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.EchoTriggeredAbility(cost.Mana{cost.O(3), cost.R}),
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target land",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Land}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Destroy{
									Object: game.TargetPermanentReference(0),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Haste
			Echo {3}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
			When this creature enters, destroy target land.
		`,
		},
	}
}
