package j

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// JaredCarthalionTrueHeir is the card definition for Jared Carthalion, True Heir.
//
// Type: Legendary Creature — Human Warrior
// Cost: {R}{G}{W}
//
// Oracle text:
//
//	When Jared Carthalion enters, target opponent becomes the monarch. You can't become the monarch this turn.
//	If damage would be dealt to Jared Carthalion while you're the monarch, prevent that damage and put that many +1/+1 counters on it.
var JaredCarthalionTrueHeir = newJaredCarthalionTrueHeir()

func newJaredCarthalionTrueHeir() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Jared Carthalion, True Heir",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
				cost.G,
				cost.W,
			}),
			Colors:     []color.Color{color.Green, color.Red, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Warrior},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
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
								Constraint: "target opponent",
								Allow:      game.TargetAllowPlayer,
								Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.BecomeMonarch{
									Player: game.TargetPlayerReference(0),
								},
							},
							{
								Primitive: game.CantBecomeMonarch{
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.DamagePreventionToPlusOneCountersReplacement("If damage would be dealt to Jared Carthalion while you're the monarch, prevent that damage and put that many +1/+1 counters on it.", false, opt.Val(game.Condition{
					ControllerIsMonarch: true,
				})),
			},
			OracleText: `
			When Jared Carthalion enters, target opponent becomes the monarch. You can't become the monarch this turn.
			If damage would be dealt to Jared Carthalion while you're the monarch, prevent that damage and put that many +1/+1 counters on it.
		`,
		},
	}
}
