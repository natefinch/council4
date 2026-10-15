package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EidolonOfTheGreatRevel is the card definition for Eidolon of the Great Revel.
//
// Type: Enchantment Creature — Spirit
// Cost: {R}{R}
//
// Oracle text:
//
//	Whenever a player casts a spell with mana value 3 or less, this creature deals 2 damage to that player.
var EidolonOfTheGreatRevel = newEidolonOfTheGreatRevel

func newEidolonOfTheGreatRevel() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Eidolon of the Great Revel",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Enchantment, types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever a player casts a spell with mana value 3 or less, this creature deals 2 damage to that player.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event: game.EventSpellCast,
							CardSelection: game.Selection{ManaValue: opt.Val(compare.Int{
								Op:    compare.LessOrEqual,
								Value: 3,
							})},
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
											Amount:       game.Fixed(2),
											Recipient:    game.PlayerDamageRecipient(game.EventPlayerReference()),
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
			Whenever a player casts a spell with mana value 3 or less, this creature deals 2 damage to that player.
		`,
		},
	}
}
