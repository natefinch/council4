package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TemurDevotee is the card definition for Temur Devotee.
//
// Type: Creature — Human Druid
// Cost: {1}{U}
//
// Oracle text:
//
//	Defender
//	{1}: Add {G}, {U}, or {R}. Activate only once each turn.
var TemurDevotee = newTemurDevotee

func newTemurDevotee() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Temur Devotee",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Druid},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
			},
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					Text: "{1}: Add {G}, {U}, or {R}. Activate only once each turn.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
					}),
					Timing: game.OncePerTurn,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Choose{
											Choice: game.ResolutionChoice{
												Kind:   game.ResolutionChoiceMana,
												Prompt: "Choose a color",
												Colors: []mana.Color{mana.G, mana.U, mana.R},
											},
											PublishChoice: game.ChoiceKey("oracle-mana-color"),
										},
									},
									game.Instruction{
										Primitive: game.AddMana{
											Amount:     game.Fixed(1),
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
			Defender
			{1}: Add {G}, {U}, or {R}. Activate only once each turn.
		`,
		},
	}
}
