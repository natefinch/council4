package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FangkeeperSFamiliar is the card definition for Fangkeeper's Familiar.
//
// Type: Creature — Snake
// Cost: {1}{B}{G}{U}
//
// Oracle text:
//
//	Flash
//	When this creature enters, choose one —
//	• You gain 3 life and surveil 3. (Look at the top three cards of your library, then put any number of them into your graveyard and the rest on top of your library in any order.)
//	• Destroy target enchantment.
//	• Counter target creature spell.
var FangkeeperSFamiliar = newFangkeeperSFamiliar

func newFangkeeperSFamiliar() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Fangkeeper's Familiar",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
				cost.G,
				cost.U,
			}),
			Colors:    []color.Color{color.Black, color.Green, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Snake},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
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
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "You gain 3 life and surveil 3. (Look at the top three cards of your library, then put any number of them into your graveyard and the rest on top of your library in any order.)",
								Sequence: []game.Instruction{
									{
										Primitive: game.GainLife{
											Amount: game.Fixed(3),
											Player: game.ControllerReference(),
										},
									},
									{
										Primitive: game.Surveil{
											Amount: game.Fixed(3),
											Player: game.ControllerReference(),
										},
									},
								},
							},
							game.Mode{
								Text: "Destroy target enchantment.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target enchantment",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Enchantment}}),
									},
								},
								Sequence: []game.Instruction{
									{
										Primitive: game.Destroy{
											Object: game.TargetPermanentReference(0),
										},
									},
								},
							},
							game.Mode{
								Text: "Counter target creature spell.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target creature spell",
										Allow:      game.TargetAllowStackObject,
										Predicate: game.TargetPredicate{
											SpellCardTypes:   []types.Card{types.Creature},
											StackObjectKinds: []game.StackObjectKind{game.StackSpell},
										},
									},
								},
								Sequence: []game.Instruction{
									{
										Primitive: game.CounterObject{
											Object: game.TargetStackObjectReference(0),
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
			Flash
			When this creature enters, choose one —
			• You gain 3 life and surveil 3. (Look at the top three cards of your library, then put any number of them into your graveyard and the rest on top of your library in any order.)
			• Destroy target enchantment.
			• Counter target creature spell.
		`,
		},
	}
}
