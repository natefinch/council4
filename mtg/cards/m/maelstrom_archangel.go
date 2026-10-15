package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MaelstromArchangel is the card definition for Maelstrom Archangel.
//
// Type: Creature — Angel
// Cost: {W}{U}{B}{R}{G}
//
// Oracle text:
//
//	Flying
//	Whenever this creature deals combat damage to a player, you may cast a spell from your hand without paying its mana cost.
var MaelstromArchangel = newMaelstromArchangel

func newMaelstromArchangel() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Black, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Maelstrom Archangel",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
				cost.U,
				cost.B,
				cost.R,
				cost.G,
			}),
			Colors:    []color.Color{color.Black, color.Green, color.Red, color.Blue, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Angel},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever this creature deals combat damage to a player, you may cast a spell from your hand without paying its mana cost.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:               game.EventDamageDealt,
							Source:              game.TriggerSourceSelf,
							Subject:             game.TriggerSubjectDamageSource,
							StackObjectKind:     game.StackSpell,
							DamageRecipient:     game.DamageRecipientPlayer,
							RequireCombatDamage: true,
							CounterKind:         counter.PlusOnePlusOne,
						},
					},
					Optional: true,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CastForFree{
											Player:    game.ControllerReference(),
											Selection: game.Selection{ExcludedTypes: []types.Card{types.Land}},
											Zone:      zone.Hand,
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
			Flying
			Whenever this creature deals combat damage to a player, you may cast a spell from your hand without paying its mana cost.
		`,
		},
	}
}
