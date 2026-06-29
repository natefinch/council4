package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GleamingBastion is the card definition for Gleaming Bastion.
//
// Type: Land
//
// Oracle text:
//
//	{T}: Add {C}.
//	{T}: Add {W} or {U}. Activate only if this land entered this turn or if you control a basic land.
var GleamingBastion = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White, color.Blue),
	CardFace: game.CardFace{
		Name:  "Gleaming Bastion",
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
									Colors: []mana.Color{mana.W, mana.U},
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
			{T}: Add {W} or {U}. Activate only if this land entered this turn or if you control a basic land.
		`,
	},
}
