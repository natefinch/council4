package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SubterraneanShambler is the card definition for Subterranean Shambler.
//
// Type: Creature — Elemental
// Cost: {3}{R}
//
// Oracle text:
//
//	Echo {3}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	When this creature enters or leaves the battlefield, it deals 1 damage to each creature without flying.
var SubterraneanShambler = newSubterraneanShambler

func newSubterraneanShambler() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Subterranean Shambler",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.EchoTriggeredAbility(cost.Mana{cost.O(3), cost.R}),
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
									Amount:       game.Fixed(1),
									Recipient:    game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludedKeyword: game.Flying})),
									DamageSource: opt.Val(game.EventPermanentReference()),
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:         game.EventZoneChanged,
							Source:        game.TriggerSourceSelf,
							MatchFromZone: true,
							FromZone:      zone.Battlefield,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:       game.Fixed(1),
									Recipient:    game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludedKeyword: game.Flying})),
									DamageSource: opt.Val(game.EventPermanentReference()),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Echo {3}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
			When this creature enters or leaves the battlefield, it deals 1 damage to each creature without flying.
		`,
		},
	}
}
