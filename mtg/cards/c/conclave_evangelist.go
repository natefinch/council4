package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ConclaveEvangelist is the card definition for Conclave Evangelist.
//
// Type: Creature — Elephant Cleric
// Cost: {3}{G/W}{G/W}
//
// Oracle text:
//
//	Myriad (Whenever this creature attacks, for each opponent other than defending player, you may create a token copy that's tapped and attacking that player or a planeswalker they control. Exile the tokens at end of combat.)
//	Whenever this creature deals combat damage to a player, create a token that's a copy of this creature.
var ConclaveEvangelist = newConclaveEvangelist

func newConclaveEvangelist() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Conclave Evangelist",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.HybridMana(mana.G, mana.W),
				cost.HybridMana(mana.G, mana.W),
			}),
			Colors:    []color.Color{color.Green, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elephant, types.Cleric},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.MyriadTriggeredBody,
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:               game.EventDamageDealt,
							Source:              game.TriggerSourceSelf,
							Subject:             game.TriggerSubjectDamageSource,
							RequireCombatDamage: true,
							DamageRecipient:     game.DamageRecipientPlayer,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenCopyOf(game.TokenCopySpec{
										Source: game.TokenCopySourceObject,
										Object: game.SourcePermanentReference(),
									}),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Myriad (Whenever this creature attacks, for each opponent other than defending player, you may create a token copy that's tapped and attacking that player or a planeswalker they control. Exile the tokens at end of combat.)
			Whenever this creature deals combat damage to a player, create a token that's a copy of this creature.
		`,
		},
	}
}
