package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GorgonRecluse is the card definition for Gorgon Recluse.
//
// Type: Creature — Gorgon
// Cost: {3}{B}{B}
//
// Oracle text:
//
//	Whenever this creature blocks or becomes blocked by a nonblack creature, destroy that creature at end of combat.
//	Madness {B}{B} (If you discard this card, discard it into exile. When you do, cast it for its madness cost or put it into your graveyard.)
var GorgonRecluse = newGorgonRecluse

func newGorgonRecluse() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Gorgon Recluse",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Gorgon},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.MadnessKeyword{Cost: cost.Mana{cost.B, cost.B}},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                   game.EventBlockerDeclared,
							Source:                  game.TriggerSourceSelf,
							UnionEvent:              game.EventAttackerBecameBlocked,
							RelatedSubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludedColors: []color.Color{color.Black}},
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
			Whenever this creature blocks or becomes blocked by a nonblack creature, destroy that creature at end of combat.
			Madness {B}{B} (If you discard this card, discard it into exile. When you do, cast it for its madness cost or put it into your graveyard.)
		`,
		},
	}
}
