package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TurretOgre is the card definition for Turret Ogre.
//
// Type: Creature — Ogre Warrior
// Cost: {3}{R}
//
// Oracle text:
//
//	Reach (This creature can block creatures with flying.)
//	When this creature enters, if you control another creature with power 4 or greater, this creature deals 2 damage to each opponent.
var TurretOgre = newTurretOgre

func newTurretOgre() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Turret Ogre",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Ogre, types.Warrior},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this creature enters, if you control another creature with power 4 or greater, this creature deals 2 damage to each opponent.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
						InterveningIf: "if you control another creature with power 4 or greater",
						InterveningCondition: opt.Val(game.Condition{
							Text: "if you control another creature with power 4 or greater",
							ControlsMatching: opt.Val(game.SelectionCount{
								Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludeSource: true, Power: opt.Val(compare.Int{
									Op:    compare.GreaterOrEqual,
									Value: 4,
								})},
							}),
							SourceCounterKind: counter.PlusOnePlusOne,
						}),
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Damage{
											Amount:       game.Fixed(2),
											Recipient:    game.PlayerGroupDamageRecipient(game.OpponentsReference()),
											DamageSource: opt.Val(game.SourcePermanentReference()),
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
			Reach (This creature can block creatures with flying.)
			When this creature enters, if you control another creature with power 4 or greater, this creature deals 2 damage to each opponent.
		`,
		},
	}
}
