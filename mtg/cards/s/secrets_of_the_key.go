package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SecretsOfTheKey is the card definition for Secrets of the Key.
//
// Type: Instant
// Cost: {U}
//
// Oracle text:
//
//	Investigate. If this spell was cast from a graveyard, investigate twice instead. (Create a Clue token. It's an artifact with "{2}, Sacrifice this token: Draw a card.")
//	Flashback {3}{U}
var SecretsOfTheKey = newSecretsOfTheKey

func newSecretsOfTheKey() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Secrets of the Key",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.FlashbackKeyword{Cost: cost.Mana{cost.O(3), cost.U}},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Investigate{
							Amount: game.Fixed(1),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Negate:       true,
								CastFromZone: opt.Val(zone.Graveyard),
							}),
						}),
					},
					{
						Primitive: game.Investigate{
							Amount: game.Fixed(2),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								CastFromZone: opt.Val(zone.Graveyard),
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Investigate. If this spell was cast from a graveyard, investigate twice instead. (Create a Clue token. It's an artifact with "{2}, Sacrifice this token: Draw a card.")
			Flashback {3}{U}
		`,
		},
	}
}
