package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WrithingChrysalis is the card definition for Writhing Chrysalis.
//
// Type: Creature — Eldrazi Drone
// Cost: {2}{R}{G}
//
// Oracle text:
//
//	Devoid (This card has no color.)
//	When you cast this spell, create two 0/1 colorless Eldrazi Spawn creature tokens with "Sacrifice this token: Add {C}."
//	Reach
//	Whenever you sacrifice another Eldrazi, put a +1/+1 counter on this creature.
var WrithingChrysalis = newWrithingChrysalis

func newWrithingChrysalis() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Writhing Chrysalis",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
				cost.G,
			}),
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Eldrazi, types.Drone},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.DevoidStaticBody,
				game.ReachStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:       game.EventSpellCast,
							Source:      game.TriggerSourceSelf,
							Controller:  game.TriggerControllerYou,
							SelfWasCast: true,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(2),
									Source: game.TokenDef(writhingChrysalisToken),
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentSacrificed,
							ExcludeSelf:      true,
							Player:           game.TriggerPlayerYou,
							SubjectSelection: game.Selection{SubtypesAny: []types.Sub{types.Sub("Eldrazi")}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Devoid (This card has no color.)
			When you cast this spell, create two 0/1 colorless Eldrazi Spawn creature tokens with "Sacrifice this token: Add {C}."
			Reach
			Whenever you sacrifice another Eldrazi, put a +1/+1 counter on this creature.
		`,
		},
	}
}

var writhingChrysalisToken = newWrithingChrysalisToken()

func newWrithingChrysalisToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Eldrazi Spawn",
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Eldrazi, types.Spawn},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalSacrificeSource,
							Text:   "Sacrifice this token",
							Amount: 1,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddMana{
									Amount:    game.Fixed(1),
									ManaColor: mana.C,
								},
							},
						},
					}.Ability(),
				},
			},
		},
	}
}
