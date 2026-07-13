package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CraterHellion is the card definition for Crater Hellion.
//
// Type: Creature — Hellion Beast
// Cost: {4}{R}{R}
//
// Oracle text:
//
//	Echo {4}{R}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	When this creature enters, it deals 4 damage to each other creature.
var CraterHellion = newCraterHellion

func newCraterHellion() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Crater Hellion",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Hellion, types.Beast},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 6}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.EchoTriggeredAbility(cost.Mana{cost.O(4), cost.R, cost.R}),
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:       game.Fixed(4),
									Recipient:    game.GroupDamageRecipient(game.BattlefieldGroupExcluding(game.Selection{RequiredTypes: []types.Card{types.Creature}}, game.SourcePermanentReference())),
									DamageSource: opt.Val(game.EventPermanentReference()),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Echo {4}{R}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
			When this creature enters, it deals 4 damage to each other creature.
		`,
		},
	}
}
