package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ThousandYearElixir is the card definition for Thousand-Year Elixir.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	You may activate abilities of creatures you control as though those creatures had haste.
//	{1}, {T}: Untap target creature.
var ThousandYearElixir = newThousandYearElixir

func newThousandYearElixir() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Thousand-Year Elixir",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.ActivateAbilitiesAsThoughHasteStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{1}, {T}: Untap target creature.",
					ManaCost:        opt.Val(cost.Mana{cost.O(1)}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Untap{
									Object: game.TargetPermanentReference(0),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			You may activate abilities of creatures you control as though those creatures had haste.
			{1}, {T}: Untap target creature.
		`,
		},
	}
}
