package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KomaWorldEater is the card definition for Koma, World-Eater.
//
// Type: Legendary Creature — Serpent
// Cost: {3}{G}{G}{U}{U}
//
// Oracle text:
//
//	This spell can't be countered.
//	Trample, ward {4}
//	Whenever Koma deals combat damage to a player, create four 3/3 blue Serpent creature tokens named Koma's Coil.
var KomaWorldEater = newKomaWorldEater

func newKomaWorldEater() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Koma, World-Eater",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
				cost.G,
				cost.U,
				cost.U,
			}),
			Colors:     []color.Color{color.Green, color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Serpent},
			Power:      opt.Val(game.PT{Value: 8}),
			Toughness:  opt.Val(game.PT{Value: 12}),
			StaticAbilities: []game.StaticAbility{
				game.CantBeCounteredStaticBody,
				game.TrampleStaticBody,
				game.WardStaticAbility(cost.Mana{cost.O(4)}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever Koma deals combat damage to a player, create four 3/3 blue Serpent creature tokens named Koma's Coil.",
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
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(4),
											Source: game.TokenDef(komaWorldEaterToken),
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
			This spell can't be countered.
			Trample, ward {4}
			Whenever Koma deals combat damage to a player, create four 3/3 blue Serpent creature tokens named Koma's Coil.
		`,
		},
	}
}

var komaWorldEaterToken = newKomaWorldEaterToken()

func newKomaWorldEaterToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Koma's Coil",
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Serpent},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
		},
	}
}
