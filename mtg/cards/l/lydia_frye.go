package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LydiaFrye is the card definition for Lydia Frye.
//
// Type: Legendary Creature — Human Assassin
// Cost: {2}{U/B}
//
// Oracle text:
//
//	Lydia Frye can't be blocked by creatures with power 3 or greater.
//	At the beginning of your end step, surveil X, where X is the number of tapped Assassins you control. (Look at the top X cards of your library, then put any number of them into your graveyard and the rest on top of your library in any order.)
var LydiaFrye = newLydiaFrye()

func newLydiaFrye() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Lydia Frye",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.HybridMana(mana.U, mana.B),
			}),
			Colors:     []color.Color{color.Black, color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Assassin},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCantBeBlockedByCreaturesWith,
							AffectedSource: true,
							BlockerRestriction: game.BlockerRestriction{
								Kind:  game.BlockerRestrictionPowerGreaterOrEqual,
								Power: 3,
							},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepEnd,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Surveil{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountCountSelector,
										Multiplier: 1,
										Group:      game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Assassin")}, Controller: game.ControllerYou, Tapped: game.TriTrue}),
									}),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Lydia Frye can't be blocked by creatures with power 3 or greater.
			At the beginning of your end step, surveil X, where X is the number of tapped Assassins you control. (Look at the top X cards of your library, then put any number of them into your graveyard and the rest on top of your library in any order.)
		`,
		},
	}
}
