package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CurseOfMaritLage is the card definition for Curse of Marit Lage.
//
// Type: Enchantment
// Cost: {3}{R}{R}
//
// Oracle text:
//
//	When this enchantment enters, tap all Islands.
//	Islands don't untap during their controllers' untap steps.
var CurseOfMaritLage = newCurseOfMaritLage

func newCurseOfMaritLage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Curse of Marit Lage",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:              game.RuleEffectDoesntUntap,
							AffectedSelection: game.Selection{SubtypesAny: []types.Sub{types.Sub("Island")}},
						},
					},
				},
			},
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
						Sequence: []game.Instruction{
							{
								Primitive: game.Tap{
									Group: game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Island")}}),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this enchantment enters, tap all Islands.
			Islands don't untap during their controllers' untap steps.
		`,
		},
	}
}
