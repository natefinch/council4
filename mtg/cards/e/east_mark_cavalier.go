package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EastMarkCavalier is the card definition for East-Mark Cavalier.
//
// Type: Creature — Human Knight
// Cost: {1}{W}
//
// Oracle text:
//
//	Vigilance
//	Whenever this creature deals damage to a Goblin or Orc, destroy that creature.
var EastMarkCavalier = newEastMarkCavalier

func newEastMarkCavalier() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "East-Mark Cavalier",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Knight},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever this creature deals damage to a Goblin or Orc, destroy that creature.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                    game.EventDamageDealt,
							Source:                   game.TriggerSourceSelf,
							Subject:                  game.TriggerSubjectDamageSource,
							StackObjectKind:          game.StackSpell,
							DamageRecipient:          game.DamageRecipientPermanent,
							DamageRecipientSelection: game.Selection{SubtypesAny: []types.Sub{types.Goblin, types.Orc}},
							CounterKind:              counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Destroy{
											Object: game.EventPermanentReference(),
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
			Vigilance
			Whenever this creature deals damage to a Goblin or Orc, destroy that creature.
		`,
		},
	}
}
