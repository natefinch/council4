package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// FuryOfTheHorde is the card definition for Fury of the Horde.
//
// Type: Sorcery
// Cost: {5}{R}{R}
//
// Oracle text:
//
//	You may exile two red cards from your hand rather than pay this spell's mana cost.
//	Untap all creatures that attacked this turn. After this main phase, there is an additional combat phase followed by an additional main phase.
var FuryOfTheHorde = newFuryOfTheHorde

func newFuryOfTheHorde() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Fury of the Horde",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label: "Exile 2 red cards",
					AdditionalCosts: []cost.Additional{
						{
							Kind:           cost.AdditionalExile,
							Amount:         2,
							Source:         zone.Hand,
							MatchCardColor: true,
							CardColor:      color.Red,
						},
					},
				},
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
			You may exile two red cards from your hand rather than pay this spell's mana cost.
			Untap all creatures that attacked this turn. After this main phase, there is an additional combat phase followed by an additional main phase.
		`,
		},
	}
}
