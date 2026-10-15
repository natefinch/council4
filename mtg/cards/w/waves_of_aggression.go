package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WavesOfAggression is the card definition for Waves of Aggression.
//
// Type: Sorcery
// Cost: {3}{R/W}{R/W}
//
// Oracle text:
//
//	Untap all creatures that attacked this turn. After this main phase, there is an additional combat phase followed by an additional main phase.
//	Retrace (You may cast this card from your graveyard by discarding a land card in addition to paying its other costs.)
var WavesOfAggression = newWavesOfAggression

func newWavesOfAggression() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name: "Waves of Aggression",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.HybridMana(mana.R, mana.W),
				cost.HybridMana(mana.R, mana.W),
			}),
			Colors: []color.Color{color.Red, color.White},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.RetraceStaticBody,
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Untap{
							Group: game.AttackedThisTurnGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
						},
					},
					{
						Primitive: game.AddExtraPhases{
							Combat: true,
							Main:   true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Untap all creatures that attacked this turn. After this main phase, there is an additional combat phase followed by an additional main phase.
			Retrace (You may cast this card from your graveyard by discarding a land card in addition to paying its other costs.)
		`,
		},
	}
}
