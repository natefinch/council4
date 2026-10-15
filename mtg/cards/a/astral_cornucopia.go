package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AstralCornucopia is the card definition for Astral Cornucopia.
//
// Type: Artifact
// Cost: {X}{X}{X}
//
// Oracle text:
//
//	This artifact enters with X charge counters on it.
//	{T}: Choose a color. Add one mana of that color for each charge counter on this artifact.
var AstralCornucopia = newAstralCornucopia

func newAstralCornucopia() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Astral Cornucopia",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.X,
				cost.X,
			}),
			Types: []types.Card{types.Artifact},
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					Text: "{T}: Choose a color. Add one mana of that color for each charge counter on this artifact.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
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
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:        game.DynamicAmountObjectCounters,
												Multiplier:  1,
												CounterKind: counter.Charge,
												Object:      game.SourcePermanentReference(),
											}),
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
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This artifact enters with X charge counters on it.", game.CounterPlacement{Kind: counter.Charge, AmountFromX: true}),
			},
			OracleText: `
			This artifact enters with X charge counters on it.
			{T}: Choose a color. Add one mana of that color for each charge counter on this artifact.
		`,
		},
	}
}
