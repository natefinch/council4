package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DwynenGiltLeafDaen is the card definition for Dwynen, Gilt-Leaf Daen.
//
// Type: Legendary Creature — Elf Warrior
// Cost: {2}{G}{G}
//
// Oracle text:
//
//	Reach (This creature can block creatures with flying.)
//	Other Elf creatures you control get +1/+1.
//	Whenever Dwynen attacks, you gain 1 life for each attacking Elf you control.
var DwynenGiltLeafDaen = newDwynenGiltLeafDaen

func newDwynenGiltLeafDaen() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Dwynen, Gilt-Leaf Daen",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Elf, types.Warrior},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Elf")}}, game.SourcePermanentReference()),
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventAttackerDeclared,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.GainLife{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountCountSelector,
										Multiplier: 1,
										Group:      game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Elf")}, Controller: game.ControllerYou, CombatState: game.CombatStateAttacking}),
									}),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Reach (This creature can block creatures with flying.)
			Other Elf creatures you control get +1/+1.
			Whenever Dwynen attacks, you gain 1 life for each attacking Elf you control.
		`,
		},
	}
}
