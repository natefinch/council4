package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MercurialKite is the card definition for Mercurial Kite.
//
// Type: Creature — Bird
// Cost: {3}{U}
//
// Oracle text:
//
//	Flying
//	Whenever this creature deals combat damage to a creature, tap that creature. That creature doesn't untap during its controller's next untap step.
var MercurialKite = newMercurialKite()

func newMercurialKite() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Mercurial Kite",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bird},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
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
			Flying
			Whenever this creature deals combat damage to a creature, tap that creature. That creature doesn't untap during its controller's next untap step.
		`,
		},
	}
}
