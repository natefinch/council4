package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BaldurSGate is the card definition for Baldur's Gate.
//
// Type: Legendary Land — Gate
//
// Oracle text:
//
//	{T}: Add {C}.
//	{2}, {T}: Add X mana of any one color, where X is the number of other Gates you control.
var BaldurSGate = newBaldurSGate

func newBaldurSGate() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:       "Baldur's Gate",
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Land},
			Subtypes:   []types.Sub{types.Gate},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
				game.ManaAbility{
					ManaCost:        opt.Val(cost.Mana{cost.O(2)}),
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
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountCountSelector,
										Multiplier: 1,
										Group:      game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Gate")}, Controller: game.ControllerYou, ExcludeSource: true}),
									}),
									ChoiceFrom: game.ChoiceKey("oracle-mana-color"),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{T}: Add {C}.
			{2}, {T}: Add X mana of any one color, where X is the number of other Gates you control.
		`,
		},
	}
}
