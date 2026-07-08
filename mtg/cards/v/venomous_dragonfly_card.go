package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VenomousDragonfly is the card definition for Venomous Dragonfly.
//
// Type: Creature — Insect
// Cost: {3}{G}
//
// Oracle text:
//
//	Flying
//	Whenever this creature blocks or becomes blocked by a creature, destroy that creature at end of combat.
var VenomousDragonfly = newVenomousDragonfly

func newVenomousDragonfly() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Venomous Dragonfly",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Insect},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                   game.EventBlockerDeclared,
							Source:                  game.TriggerSourceSelf,
							UnionEvent:              game.EventAttackerBecameBlocked,
							RelatedSubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateDelayedTrigger{
									Trigger: game.DelayedTriggerDef{
										Timing:         game.DelayedAtEndOfCombat,
										CapturedObject: opt.Val(game.EventRelatedPermanentReference()),
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
			Flying
			Whenever this creature blocks or becomes blocked by a creature, destroy that creature at end of combat.
		`,
		},
	}
}
