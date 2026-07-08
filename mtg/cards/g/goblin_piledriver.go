package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GoblinPiledriver is the card definition for Goblin Piledriver.
//
// Type: Creature — Goblin Warrior
// Cost: {1}{R}
//
// Oracle text:
//
//	Protection from blue (This creature can't be blocked, targeted, dealt damage, or enchanted by anything blue.)
//	Whenever this creature attacks, it gets +2/+0 until end of turn for each other attacking Goblin.
var GoblinPiledriver = newGoblinPiledriver

func newGoblinPiledriver() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Goblin Piledriver",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin, types.Warrior},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.ProtectionFromColorsStaticAbility(color.Blue),
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
								Primitive: game.ModifyPT{
									Object: game.EventPermanentReference(),
									PowerDelta: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountCountSelector,
										Multiplier: 2,
										Group:      game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Goblin")}, CombatState: game.CombatStateAttacking, ExcludeSource: true}),
									}),
									ToughnessDelta: game.Fixed(0),
									Duration:       game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Protection from blue (This creature can't be blocked, targeted, dealt damage, or enchanted by anything blue.)
			Whenever this creature attacks, it gets +2/+0 until end of turn for each other attacking Goblin.
		`,
		},
	}
}
