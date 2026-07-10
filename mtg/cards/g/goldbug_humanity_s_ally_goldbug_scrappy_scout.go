package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GoldbugHumanitySAlly is the card definition for Goldbug, Humanity's Ally // Goldbug, Scrappy Scout.
//
// Type: Legendary Artifact Creature — Robot // Legendary Artifact — Vehicle
// Face: Goldbug, Scrappy Scout — Legendary Artifact — Vehicle
//
// Oracle text:
//
//	More Than Meets the Eye {W}{U} (You may cast this card converted for {W}{U}.)
//	Prevent all combat damage that would be dealt to attacking Humans you control.
//	Whenever you cast your second spell each turn, convert Goldbug.
var GoldbugHumanitySAlly = newGoldbugHumanitySAlly

func newGoldbugHumanitySAlly() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Goldbug, Humanity's Ally",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
				cost.U,
			}),
			Colors:     []color.Color{color.Blue, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact, types.Creature},
			Subtypes:   []types.Sub{types.Robot},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                      game.EventSpellCast,
							Controller:                 game.TriggerControllerYou,
							PlayerEventOrdinalThisTurn: 2,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Transform{
									Object: game.SourcePermanentReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.CombatDamagePreventionToGroupReplacement("Prevent all combat damage that would be dealt to attacking Humans you control.", game.Selection{SubtypesAny: []types.Sub{types.Sub("Human")}, Controller: game.ControllerYou, CombatState: game.CombatStateAttacking}),
			},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label:    "More Than Meets the Eye",
					ManaCost: opt.Val(cost.Mana{cost.W, cost.U}),
					Mechanic: cost.AlternativeMechanicMoreThanMeetsTheEye,
				},
			},
			OracleText: `
			More Than Meets the Eye {W}{U} (You may cast this card converted for {W}{U}.)
			Prevent all combat damage that would be dealt to attacking Humans you control.
			Whenever you cast your second spell each turn, convert Goldbug.
		`,
		},
		Layout: game.LayoutTransform,
		Back: opt.Val(game.CardFace{
			Name:       "Goldbug, Scrappy Scout",
			Colors:     []color.Color{color.Blue, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			Subtypes:   []types.Sub{types.Vehicle},
			Power:      opt.Val(game.PT{Value: 1}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.LivingMetalStaticBody,
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:               game.RuleEffectCantBeCountered,
							AffectedController: game.ControllerYou,
							SpellSubtypes:      []types.Sub{types.Human},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                     game.EventAttackerDeclared,
							Source:                    game.TriggerSourceSelf,
							AttacksAlongsideCount:     1,
							AttacksAlongsideSelection: game.Selection{SubtypesAny: []types.Sub{types.Sub("Human")}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
							{
								Primitive: game.Transform{
									Object: game.SourcePermanentReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Living metal (During your turn, this Vehicle is also a creature.)
			Human spells you control can't be countered.
			Whenever Goldbug and at least one Human attack, draw a card and convert Goldbug.
		`,
		}),
	}
}
