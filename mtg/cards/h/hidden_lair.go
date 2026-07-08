package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HiddenLair is the card definition for Hidden Lair.
//
// Type: Land
//
// Oracle text:
//
//	{T}: Add {C}.
//	{T}: Add {U} or {B}. Activate only if this land entered this turn or if you control a basic land.
var HiddenLair = func() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name:  "Hidden Lair",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
				{
					AdditionalCosts: cost.Tap,
					ActivationCondition: opt.Val(game.Condition{
						LandEnteredThisTurnOrControlsBasicLand: true,
					}),
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Choose{
									Choice: game.ResolutionChoice{
										Kind:   game.ResolutionChoiceMana,
										Prompt: "Choose a color",
										Colors: []mana.Color{mana.U, mana.B},
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
			{T}: Add {U} or {B}. Activate only if this land entered this turn or if you control a basic land.
		`,
		},
	}
}
