package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BlasphemousEdict is the card definition for Blasphemous Edict.
//
// Type: Sorcery
// Cost: {3}{B}{B}
//
// Oracle text:
//
//	You may pay {B} rather than pay this spell's mana cost if there are thirteen or more creatures on the battlefield.
//	Each player sacrifices thirteen creatures of their choice.
var BlasphemousEdict = newBlasphemousEdict

func newBlasphemousEdict() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Blasphemous Edict",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label:                  "Pay {B}",
					ManaCost:               opt.Val(cost.Mana{cost.B}),
					Condition:              cost.AlternativeConditionPermanentsOnBattlefield,
					ConditionCount:         13,
					ConditionPermanentType: types.Creature,
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.SacrificePermanents{
							Amount:      game.Fixed(13),
							PlayerGroup: game.AllPlayersReference(),
							Selection:   game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
				},
			}.Ability()),
			OracleText: `
			You may pay {B} rather than pay this spell's mana cost if there are thirteen or more creatures on the battlefield.
			Each player sacrifices thirteen creatures of their choice.
		`,
		},
	}
}
