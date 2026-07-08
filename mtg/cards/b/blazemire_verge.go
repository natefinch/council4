package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BlazemireVerge is the card definition for Blazemire Verge.
//
// Type: Land
//
// Oracle text:
//
//	{T}: Add {B}.
//	{T}: Add {R}. Activate only if you control a Swamp or a Mountain.
var BlazemireVerge = func() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name:  "Blazemire Verge",
			Types: []types.Card{types.Land},
			OracleText: `
			{T}: Add {B}.
			{T}: Add {R}. Activate only if you control a Swamp or a Mountain.
		`,
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.B),
				{
					Text: `
					{T}: Add {R}. Activate only if you control a Swamp or a Mountain.
				`,
					AdditionalCosts: cost.Tap,
					ActivationCondition: opt.Val(game.Condition{
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{
								SubtypesAny: []types.Sub{
									types.Swamp,
									types.Mountain,
								},
							},
						}),
					}),
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddMana{
									Amount:    game.Fixed(1),
									ManaColor: mana.R,
								},
							},
						},
					}.Ability(),
				},
			},
		},
	}
}
