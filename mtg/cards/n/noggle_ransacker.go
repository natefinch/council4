package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NoggleRansacker is the card definition for Noggle Ransacker.
//
// Type: Creature — Noggle Rogue
// Cost: {2}{U/R}
//
// Oracle text:
//
//	When this creature enters, each player draws two cards, then discards a card at random.
var NoggleRansacker = newNoggleRansacker

func newNoggleRansacker() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Noggle Ransacker",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.HybridMana(mana.U, mana.R),
			}),
			Colors:    []color.Color{color.Red, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Noggle, types.Rogue},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			TriggeredAbilities: []game.TriggeredAbility{
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
								Primitive: game.Draw{
									Amount:      game.Fixed(2),
									PlayerGroup: game.AllPlayersReference(),
								},
							},
							{
								Primitive: game.Discard{
									Amount:      game.Fixed(1),
									PlayerGroup: game.AllPlayersReference(),
									AtRandom:    true,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this creature enters, each player draws two cards, then discards a card at random.
		`,
		},
	}
}
