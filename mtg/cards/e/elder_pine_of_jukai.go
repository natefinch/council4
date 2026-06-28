package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ElderPineOfJukai is the card definition for Elder Pine of Jukai.
//
// Type: Creature — Spirit
// Cost: {2}{G}
//
// Oracle text:
//
//	Whenever you cast a Spirit or Arcane spell, reveal the top three cards of your library. Put all land cards revealed this way into your hand and the rest on the bottom of your library in any order.
//	Soulshift 2 (When this creature dies, you may return target Spirit card with mana value 2 or less from your graveyard to your hand.)
var ElderPineOfJukai = newElderPineOfJukai()

func newElderPineOfJukai() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Elder Pine of Jukai",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:         game.EventSpellCast,
							Controller:    game.TriggerControllerYou,
							CardSelection: game.Selection{SubtypesAny: []types.Sub{types.Sub("Spirit"), types.Sub("Arcane")}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.RevealTopPartition{
									Player:    game.ControllerReference(),
									Amount:    game.Fixed(3),
									Selection: game.Selection{RequiredTypes: []types.Card{types.Land}},
									Remainder: game.DigRemainderLibraryBottom,
								},
							},
						},
					}.Ability(),
				},
				game.SoulshiftTriggeredAbility(2),
			},
			OracleText: `
			Whenever you cast a Spirit or Arcane spell, reveal the top three cards of your library. Put all land cards revealed this way into your hand and the rest on the bottom of your library in any order.
			Soulshift 2 (When this creature dies, you may return target Spirit card with mana value 2 or less from your graveyard to your hand.)
		`,
		},
	}
}
