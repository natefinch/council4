package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Scarecrone is the card definition for Scarecrone.
//
// Type: Artifact Creature — Scarecrow
// Cost: {3}
//
// Oracle text:
//
//	{1}, Sacrifice a Scarecrow: Draw a card.
//	{4}, {T}: Return target artifact creature card from your graveyard to the battlefield.
var Scarecrone = newScarecrone

func newScarecrone() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Scarecrone",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Scarecrow},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 2}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{1}, Sacrifice a Scarecrow: Draw a card.",
					ManaCost: opt.Val(cost.Mana{cost.O(1)}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalSacrifice,
							Text:        "Sacrifice a Scarecrow",
							Amount:      1,
							SubtypesAny: cost.SubtypeSet{types.Scarecrow},
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
				game.ActivatedAbility{
					Text:            "{4}, {T}: Return target artifact creature card from your graveyard to the battlefield.",
					ManaCost:        opt.Val(cost.Mana{cost.O(4)}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target artifact creature card from your graveyard",
								Allow:      game.TargetAllowCard,
								TargetZone: zone.Graveyard,
								Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Artifact, types.Creature}, Controller: game.ControllerYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.PutOnBattlefield{
									Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{1}, Sacrifice a Scarecrow: Draw a card.
			{4}, {T}: Return target artifact creature card from your graveyard to the battlefield.
		`,
		},
	}
}
