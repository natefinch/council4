package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// FruitOfTizerus is the card definition for Fruit of Tizerus.
//
// Type: Sorcery
// Cost: {B}
//
// Oracle text:
//
//	Target player loses 2 life.
//	Escape—{3}{B}, Exile three other cards from your graveyard. (You may cast this card from your graveyard for its escape cost.)
var FruitOfTizerus = newFruitOfTizerus

func newFruitOfTizerus() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Fruit of Tizerus",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.SimpleKeyword{
							Kind: game.Escape,
						},
					},
				},
			},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label:    "Escape",
					ManaCost: opt.Val(cost.Mana{cost.O(3), cost.B}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:          cost.AdditionalExile,
							Text:          "Exile three other cards from your graveyard",
							Amount:        3,
							Source:        zone.Graveyard,
							ExcludeSource: true,
						},
					},
					Mechanic: cost.AlternativeMechanicEscape,
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "Target player",
						Allow:      game.TargetAllowPlayer,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.LoseLife{
							Amount: game.Fixed(2),
							Player: game.TargetPlayerReference(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target player loses 2 life.
			Escape—{3}{B}, Exile three other cards from your graveyard. (You may cast this card from your graveyard for its escape cost.)
		`,
		},
	}
}
