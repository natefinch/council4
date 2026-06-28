package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GwendlynDiCorci is the card definition for Gwendlyn Di Corci.
//
// Type: Legendary Creature — Human Rogue
// Cost: {U}{B}{B}{R}
//
// Oracle text:
//
//	{T}: Target player discards a card at random. Activate only during your turn.
var GwendlynDiCorci = newGwendlynDiCorci()

func newGwendlynDiCorci() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Gwendlyn Di Corci",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
				cost.B,
				cost.B,
				cost.R,
			}),
			Colors:     []color.Color{color.Black, color.Red, color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Rogue},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{T}: Target player discards a card at random. Activate only during your turn.",
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Timing:          game.DuringYourTurn,
					Content: game.Mode{
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
								Primitive: game.Discard{
									Amount:   game.Fixed(1),
									Player:   game.TargetPlayerReference(0),
									AtRandom: true,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{T}: Target player discards a card at random. Activate only during your turn.
		`,
		},
	}
}
