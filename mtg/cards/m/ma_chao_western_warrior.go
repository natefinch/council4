package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MaChaoWesternWarrior is the card definition for Ma Chao, Western Warrior.
//
// Type: Legendary Creature — Human Soldier Warrior
// Cost: {3}{R}{R}
//
// Oracle text:
//
//	Horsemanship (This creature can't be blocked except by creatures with horsemanship.)
//	Whenever Ma Chao attacks alone, it can't be blocked this combat.
var MaChaoWesternWarrior = newMaChaoWesternWarrior

func newMaChaoWesternWarrior() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Ma Chao, Western Warrior",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.R,
			}),
			Colors:     []color.Color{color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Soldier, types.Warrior},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.HorsemanshipStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:       game.EventAttackerDeclared,
							Source:      game.TriggerSourceSelf,
							AttackAlone: true,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyRule{
									Object: opt.Val(game.EventPermanentReference()),
									RuleEffects: []game.RuleEffect{
										game.RuleEffect{
											Kind: game.RuleEffectCantBeBlocked,
										},
									},
									Duration: game.DurationUntilEndOfCombat,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Horsemanship (This creature can't be blocked except by creatures with horsemanship.)
			Whenever Ma Chao attacks alone, it can't be blocked this combat.
		`,
		},
	}
}
