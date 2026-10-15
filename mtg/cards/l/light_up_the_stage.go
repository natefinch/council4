package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LightUpTheStage is the card definition for Light Up the Stage.
//
// Type: Sorcery
// Cost: {2}{R}
//
// Oracle text:
//
//	Spectacle {R} (You may cast this spell for its spectacle cost rather than its mana cost if an opponent lost life this turn.)
//	Exile the top two cards of your library. Until the end of your next turn, you may play those cards.
var LightUpTheStage = newLightUpTheStage

func newLightUpTheStage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Light Up the Stage",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label:     "Spectacle",
					ManaCost:  opt.Val(cost.Mana{cost.R}),
					Condition: cost.AlternativeConditionOpponentLostLifeThisTurn,
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.ImpulseExile{
							Player:   game.ControllerReference(),
							Amount:   game.Fixed(2),
							Duration: game.DurationUntilEndOfYourNextTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Spectacle {R} (You may cast this spell for its spectacle cost rather than its mana cost if an opponent lost life this turn.)
			Exile the top two cards of your library. Until the end of your next turn, you may play those cards.
		`,
		},
	}
}
