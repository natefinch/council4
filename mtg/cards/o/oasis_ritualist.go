package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// OasisRitualist is the card definition for Oasis Ritualist.
//
// Type: Creature — Snake Druid
// Cost: {3}{G}
//
// Oracle text:
//
//	{T}: Add one mana of any color.
//	{T}, Exert this creature: Add two mana of any one color. (An exerted creature won't untap during your next untap step.)
var OasisRitualist = newOasisRitualist

func newOasisRitualist() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Oasis Ritualist",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Snake, types.Druid},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 4}),
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.W, mana.U, mana.B, mana.R, mana.G),
				game.ManaAbility{
					Text: "{T}, Exert this creature: Add two mana of any one color. (An exerted creature won't untap during your next untap step.)",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:        cost.AdditionalExert,
							Text:        "Exert this creature",
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Choose{
											Choice: game.ResolutionChoice{
												Kind:   game.ResolutionChoiceMana,
												Prompt: "Choose a color",
												Colors: []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G},
											},
											PublishChoice: game.ChoiceKey("oracle-mana-color"),
										},
									},
									game.Instruction{
										Primitive: game.AddMana{
											Amount:     game.Fixed(2),
											ChoiceFrom: game.ChoiceKey("oracle-mana-color"),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			{T}: Add one mana of any color.
			{T}, Exert this creature: Add two mana of any one color. (An exerted creature won't untap during your next untap step.)
		`,
		},
	}
}
