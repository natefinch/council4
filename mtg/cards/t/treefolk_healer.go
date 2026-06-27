package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TreefolkHealer is the card definition for Treefolk Healer.
//
// Type: Creature — Treefolk Cleric
// Cost: {4}{G}
//
// Oracle text:
//
//	{2}{W}, {T}: Prevent the next 2 damage that would be dealt to any target this turn.
var TreefolkHealer = newTreefolkHealer()

func newTreefolkHealer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Treefolk Healer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Treefolk, types.Cleric},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{2}{W}, {T}: Prevent the next 2 damage that would be dealt to any target this turn.",
					ManaCost:        opt.Val(cost.Mana{cost.O(2), cost.W}),
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
									Amount:    game.Fixed(2),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{2}{W}, {T}: Prevent the next 2 damage that would be dealt to any target this turn.
		`,
		},
	}
}
