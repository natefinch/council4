package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SamiteHealer is the card definition for Samite Healer.
//
// Type: Creature — Human Cleric
// Cost: {1}{W}
//
// Oracle text:
//
//	{T}: Prevent the next 1 damage that would be dealt to any target this turn.
var SamiteHealer = newSamiteHealer()

func newSamiteHealer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Samite Healer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Cleric},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{T}: Prevent the next 1 damage that would be dealt to any target this turn.",
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "any target",
								Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.PreventDamage{
									AnyTarget: game.AnyTargetDamageRecipient(0),
									Amount:    game.Fixed(1),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{T}: Prevent the next 1 damage that would be dealt to any target this turn.
		`,
		},
	}
}
