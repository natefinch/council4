package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TyranidHarridan is the card definition for Tyranid Harridan.
//
// Type: Creature — Tyranid
// Cost: {4}{G}{U}
//
// Oracle text:
//
//	Flying, ward {4}
//	Shrieking Gargoyles — Whenever this creature or another Tyranid you control deals combat damage to a player, create a 1/1 blue Tyranid Gargoyle creature token with flying.
var TyranidHarridan = newTyranidHarridan

func newTyranidHarridan() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Tyranid Harridan",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
				cost.U,
			}),
			Colors:    []color.Color{color.Green, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Tyranid},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.WardStaticAbility(cost.Mana{cost.O(4)}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                       game.EventDamageDealt,
							Controller:                  game.TriggerControllerYou,
							Subject:                     game.TriggerSubjectDamageSource,
							RequireCombatDamage:         true,
							DamageRecipient:             game.DamageRecipientPlayer,
							DamageSourceSelection:       game.Selection{SubtypesAny: []types.Sub{types.Sub("Tyranid")}},
							DamageSourceSelectionOrSelf: true,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(tyranidHarridanToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying, ward {4}
			Shrieking Gargoyles — Whenever this creature or another Tyranid you control deals combat damage to a player, create a 1/1 blue Tyranid Gargoyle creature token with flying.
		`,
		},
	}
}

var tyranidHarridanToken = newTyranidHarridanToken()

func newTyranidHarridanToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Tyranid Gargoyle",
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Tyranid, types.Gargoyle},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}
