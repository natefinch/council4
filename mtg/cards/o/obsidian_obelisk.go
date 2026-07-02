package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ObsidianObelisk is the card definition for Obsidian Obelisk.
//
// Type: Artifact
// Cost: {2}
//
// Oracle text:
//
//	This artifact enters tapped.
//	{T}: Add {C}.
//	{T}: Add one mana of any color. Spend this mana only to cast a multicolored spell.
var ObsidianObelisk = newObsidianObelisk()

func newObsidianObelisk() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Obsidian Obelisk",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types: []types.Card{types.Artifact},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
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
									Amount:     game.Fixed(1),
									ChoiceFrom: game.ChoiceKey("oracle-mana-color"),
									SpendRider: opt.Val(game.ManaSpendRider{
										Condition:   game.ManaSpendCastMulticoloredSpell,
										Restriction: game.ManaSpendRestrictedToCondition,
									}),
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This artifact enters tapped."),
			},
			OracleText: `
			This artifact enters tapped.
			{T}: Add {C}.
			{T}: Add one mana of any color. Spend this mana only to cast a multicolored spell.
		`,
		},
	}
}
