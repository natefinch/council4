package q

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// QueenOfIce is the card definition for Queen of Ice // Rage of Winter.
//
// Type: Creature — Human Noble Wizard // Sorcery — Adventure
// Cost: {2}{U} // {1}{U}
// Face: Rage of Winter — Sorcery — Adventure ({1}{U})
//
// Oracle text:
//
//	Whenever this creature deals combat damage to a creature, tap that creature. It doesn't untap during its controller's next untap step.
var QueenOfIce = newQueenOfIce()

func newQueenOfIce() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Queen of Ice",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Noble, types.Wizard},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                game.EventDamageDealt,
							Source:               game.TriggerSourceSelf,
							Subject:              game.TriggerSubjectDamageSource,
							RequireCombatDamage:  true,
							DamageRecipient:      game.DamageRecipientPermanent,
							DamageRecipientTypes: []types.Card{types.Creature},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Tap{
									Object: game.EventPermanentReference(),
								},
							},
							{
								Primitive: game.SkipNextUntap{
									Object: game.EventPermanentReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever this creature deals combat damage to a creature, tap that creature. It doesn't untap during its controller's next untap step.
		`,
		},
		Layout: game.LayoutAdventure,
		Alternate: opt.Val(game.CardFace{
			Name: "Rage of Winter",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:   []color.Color{color.Blue},
			Types:    []types.Card{types.Sorcery},
			Subtypes: []types.Sub{types.Adventure},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Tap{
							Object: game.TargetPermanentReference(0),
						},
					},
					{
						Primitive: game.SkipNextUntap{
							Object: game.TargetPermanentReference(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Tap target creature. It doesn't untap during its controller's next untap step. (Then exile this card. You may cast the creature later from exile.)
		`,
		}),
	}
}
