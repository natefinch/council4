package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// EiganjoCastle is the card definition for Eiganjo Castle.
//
// Type: Legendary Land
//
// Oracle text:
//
//	{T}: Add {W}.
//	{W}, {T}: Prevent the next 2 damage that would be dealt to target legendary creature this turn.
var EiganjoCastle = newEiganjoCastle()

func newEiganjoCastle() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:       "Eiganjo Castle",
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{W}, {T}: Prevent the next 2 damage that would be dealt to target legendary creature this turn.",
					ManaCost:        opt.Val(cost.Mana{cost.W}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target legendary creature",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Supertypes: []types.Super{types.Legendary}}),
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
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.W),
			},
			OracleText: `
			{T}: Add {W}.
			{W}, {T}: Prevent the next 2 damage that would be dealt to target legendary creature this turn.
		`,
		},
	}
}
