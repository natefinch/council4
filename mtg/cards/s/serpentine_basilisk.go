package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SerpentineBasilisk is the card definition for Serpentine Basilisk.
//
// Type: Creature — Basilisk
// Cost: {2}{G}{G}
//
// Oracle text:
//
//	Whenever this creature deals combat damage to a creature, destroy that creature at end of combat.
//	Morph {1}{G}{G} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its morph cost.)
var SerpentineBasilisk = newSerpentineBasilisk

func newSerpentineBasilisk() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Serpentine Basilisk",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Basilisk},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.MorphKeyword{Cost: cost.Mana{cost.O(1), cost.G, cost.G}},
					},
				},
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
								Primitive: game.CreateDelayedTrigger{
									Trigger: game.DelayedTriggerDef{
										Timing:         game.DelayedAtEndOfCombat,
										CapturedObject: opt.Val(game.EventPermanentReference()),
										Content: game.Mode{
											Sequence: []game.Instruction{
												{
													Primitive: game.Destroy{
														Object: game.CapturedObjectReference(),
													},
												},
											},
										}.Ability(),
									},
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever this creature deals combat damage to a creature, destroy that creature at end of combat.
			Morph {1}{G}{G} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its morph cost.)
		`,
		},
	}
}
