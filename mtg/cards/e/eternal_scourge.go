package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// EternalScourge is the card definition for Eternal Scourge.
//
// Type: Creature — Eldrazi Horror
// Cost: {3}
//
// Oracle text:
//
//	You may cast this card from exile.
//	When this creature becomes the target of a spell or ability an opponent controls, exile this creature.
var EternalScourge = newEternalScourge

func newEternalScourge() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Eternal Scourge",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Eldrazi, types.Horror},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ZoneOfFunction: zone.Exile,
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCastFromZone,
							AffectedSource: true,
							AffectedPlayer: game.PlayerYou,
							CastFromZone:   zone.Exile,
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventObjectBecameTarget,
							Source:          game.TriggerSourceSelf,
							CauseController: game.TriggerControllerOpponent,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Exile{
									Object: game.SourceCardPermanentReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			You may cast this card from exile.
			When this creature becomes the target of a spell or ability an opponent controls, exile this creature.
		`,
		},
	}
}
