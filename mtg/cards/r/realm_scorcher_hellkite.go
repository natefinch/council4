package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RealmScorcherHellkite is the card definition for Realm-Scorcher Hellkite.
//
// Type: Creature — Dragon
// Cost: {4}{R}{R}
//
// Oracle text:
//
//	Bargain (You may sacrifice an artifact, enchantment, or token as you cast this spell.)
//	Flying, haste
//	When this creature enters, if it was bargained, add four mana in any combination of colors.
//	{1}{R}: This creature deals 1 damage to any target.
var RealmScorcherHellkite = newRealmScorcherHellkite

func newRealmScorcherHellkite() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Realm-Scorcher Hellkite",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dragon},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.BargainStaticBody,
				game.FlyingStaticBody,
				game.HasteStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{1}{R}: This creature deals 1 damage to any target.",
					ManaCost:       opt.Val(cost.Mana{cost.O(1), cost.R}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "any target",
								Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:       game.Fixed(1),
									Recipient:    game.AnyTargetDamageRecipient(0),
									DamageSource: opt.Val(game.SourcePermanentReference()),
								},
							},
						},
					}.Ability(),
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
						InterveningIf:                           "if it was bargained",
						InterveningIfEventPermanentWasBargained: true,
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddMana{
									Amount:            game.Fixed(4),
									CombinationColors: []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G},
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Bargain (You may sacrifice an artifact, enchantment, or token as you cast this spell.)
			Flying, haste
			When this creature enters, if it was bargained, add four mana in any combination of colors.
			{1}{R}: This creature deals 1 damage to any target.
		`,
		},
	}
}
