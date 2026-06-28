package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// CorruptOfficial is the card definition for Corrupt Official.
//
// Type: Creature — Human Minion
// Cost: {4}{B}
//
// Oracle text:
//
//	{2}{B}: Regenerate this creature.
//	Whenever this creature becomes blocked, defending player discards a card at random.
var CorruptOfficial = newCorruptOfficial()

func newCorruptOfficial() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Corrupt Official",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Minion},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{2}{B}: Regenerate this creature.",
					ManaCost:       opt.Val(cost.Mana{cost.O(2), cost.B}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Regenerate{
									Object: game.SourcePermanentReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventAttackerBecameBlocked,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Discard{
									Amount:   game.Fixed(1),
									Player:   game.DefendingPlayerReference(),
									AtRandom: true,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{2}{B}: Regenerate this creature.
			Whenever this creature becomes blocked, defending player discards a card at random.
		`,
		},
	}
}
