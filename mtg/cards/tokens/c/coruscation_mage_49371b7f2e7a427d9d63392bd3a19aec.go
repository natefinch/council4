package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CoruscationMage is the card definition for Coruscation Mage.
//
// Type: Token Creature — Otter Wizard
// Cost: {1}{R}
//
// Oracle text:
//   Whenever you cast a noncreature spell, this creature deals 1 damage to each opponent.
//   (This token's mana cost is {1}{R}.)

// CoruscationMageToken49371b7f2e7a427d9d63392bd3a19aec is the card definition for Coruscation Mage.
var CoruscationMageToken49371b7f2e7a427d9d63392bd3a19aec = newCoruscationMageToken49371b7f2e7a427d9d63392bd3a19aec()

func newCoruscationMageToken49371b7f2e7a427d9d63392bd3a19aec() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Coruscation Mage",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Otter, types.Wizard},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever you cast a noncreature spell, this creature deals 1 damage to each opponent.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventSpellCast,
							Controller:      game.TriggerControllerYou,
							CardSelection:   game.Selection{ExcludedTypes: []types.Card{types.Creature}},
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Damage{
											Amount:       game.Fixed(1),
											Recipient:    game.PlayerGroupDamageRecipient(game.OpponentsReference()),
											DamageSource: opt.Val(game.SourcePermanentReference()),
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
			Whenever you cast a noncreature spell, this creature deals 1 damage to each opponent.
			(This token's mana cost is {1}{R}.)
		`,
		},
	}
}
