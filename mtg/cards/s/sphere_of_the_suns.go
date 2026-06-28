package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SphereOfTheSuns is the card definition for Sphere of the Suns.
//
// Type: Artifact
// Cost: {2}
//
// Oracle text:
//
//	This artifact enters tapped and with three charge counters on it.
//	{T}, Remove a charge counter from this artifact: Add one mana of any color.
var SphereOfTheSuns = newSphereOfTheSuns()

func newSphereOfTheSuns() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Sphere of the Suns",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types: []types.Card{types.Artifact},
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:        cost.AdditionalRemoveCounter,
							Text:        "Remove a charge counter from this artifact",
							Amount:      1,
							CounterKind: counter.Charge,
						},
					},
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
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedWithCountersReplacement("This artifact enters tapped and with three charge counters on it.", game.CounterPlacement{Kind: counter.Charge, Amount: 3}),
			},
			OracleText: `
			This artifact enters tapped and with three charge counters on it.
			{T}, Remove a charge counter from this artifact: Add one mana of any color.
		`,
		},
	}
}
