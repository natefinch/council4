package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MomentSPeace is the card definition for Moment's Peace.
//
// Type: Instant
// Cost: {1}{G}
//
// Oracle text:
//
//	Prevent all combat damage that would be dealt this turn.
//	Flashback {2}{G} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
var MomentSPeace = newMomentSPeace

func newMomentSPeace() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Moment's Peace",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Flashback {2}{G}",
					KeywordAbilities: []game.KeywordAbility{
						game.FlashbackKeyword{
							Cost: cost.Mana{
								cost.O(2),
								cost.G,
							},
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.PreventDamage{
							All:        true,
							CombatOnly: true,
							Global:     true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Prevent all combat damage that would be dealt this turn.
			Flashback {2}{G} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
		`,
		},
	}
}
