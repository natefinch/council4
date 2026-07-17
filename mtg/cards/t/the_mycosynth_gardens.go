package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TheMycosynthGardens is the card definition for The Mycosynth Gardens.
//
// Type: Land — Sphere
//
// Oracle text:
//
//	{T}: Add {C}.
//	{1}, {T}: Add one mana of any color.
//	{X}, {T}: This land becomes a copy of target nontoken artifact you control with mana value X.
var TheMycosynthGardens = newTheMycosynthGardens

func newTheMycosynthGardens() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "The Mycosynth Gardens",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Sphere},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{X}, {T}: This land becomes a copy of target nontoken artifact you control with mana value X.",
					ManaCost:        opt.Val(cost.Mana{cost.X}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets:       1,
								MaxTargets:       1,
								Constraint:       "target nontoken artifact you control with mana value X",
								Allow:            game.TargetAllowPermanent,
								Selection:        opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact}, Controller: game.ControllerYou, NonToken: true}),
								ManaValueEqualsX: true,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.BecomeCopy{
									Object: game.TargetPermanentReference(0),
								},
							},
						},
					}.Ability(),
				},
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
				game.ManaAbility{
					ManaCost:        opt.Val(cost.Mana{cost.O(1)}),
					AdditionalCosts: cost.Tap,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Choose{
									Choice: game.ResolutionChoice{
										Kind:   game.ResolutionChoiceMana,
										Prompt: "Choose a color",
										Colors: []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G},
									},
									PublishChoice: game.ChoiceKey("oracle-mana-color"),
								},
							},
							{
								Primitive: game.AddMana{
									Amount:     game.Fixed(1),
									ChoiceFrom: game.ChoiceKey("oracle-mana-color"),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{T}: Add {C}.
			{1}, {T}: Add one mana of any color.
			{X}, {T}: This land becomes a copy of target nontoken artifact you control with mana value X.
		`,
		},
	}
}
