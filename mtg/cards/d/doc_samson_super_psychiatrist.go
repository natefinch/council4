package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DocSamsonSuperPsychiatrist is the card definition for Doc Samson, Super Psychiatrist.
//
// Type: Legendary Creature — Gamma Doctor Hero
// Cost: {4}{G}
//
// Oracle text:
//
//	If you would put one or more counters on a permanent you control, put that many plus one of each of those kinds of counters on that permanent instead.
//	{T}: Add X mana of any one color, where X is Doc Samson's power.
var DocSamsonSuperPsychiatrist = newDocSamsonSuperPsychiatrist()

func newDocSamsonSuperPsychiatrist() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Doc Samson, Super Psychiatrist",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Sub("Gamma"), types.Doctor, types.Hero},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 6}),
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
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
										Kind:       game.DynamicAmountObjectPower,
										Multiplier: 1,
										Object:     game.SourcePermanentReference(),
									}),
									ChoiceFrom: game.ChoiceKey("oracle-mana-color"),
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.ControlledPermanentCounterPlacementReplacement("If you would put one or more counters on a permanent you control, put that many plus one of each of those kinds of counters on that permanent instead.", 0, 1, game.TriggerControllerYou),
			},
			OracleText: `
			If you would put one or more counters on a permanent you control, put that many plus one of each of those kinds of counters on that permanent instead.
			{T}: Add X mana of any one color, where X is Doc Samson's power.
		`,
		},
	}
}
