package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PillarOfTheParuns is the card definition for Pillar of the Paruns.
//
// Type: Land
//
// Oracle text:
//
//	{T}: Add one mana of any color. Spend this mana only to cast a multicolored spell.
var PillarOfTheParuns = newPillarOfTheParuns()

func newPillarOfTheParuns() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Pillar of the Paruns",
			Types: []types.Card{types.Land},
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
			OracleText: `
			{T}: Add one mana of any color. Spend this mana only to cast a multicolored spell.
		`,
		},
	}
}
