package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RamsesOverdark is the card definition for Ramses Overdark.
var RamsesOverdark = newRamsesOverdark()

func newRamsesOverdark() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Ramses Overdark",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
				cost.U,
				cost.B,
				cost.B,
			}),
			Colors:     []color.Color{color.Black, color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Assassin},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{T}: Destroy target enchanted creature.",
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target enchanted creature",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, MatchEnchanted: true}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Destroy{
									Object: game.TargetPermanentReference(0),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{T}: Destroy target enchanted creature.
		`,
		},
	}
}
