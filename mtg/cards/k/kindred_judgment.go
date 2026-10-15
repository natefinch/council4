package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KindredJudgment is the card definition for Kindred Judgment.
//
// Type: Sorcery
// Cost: {5}{W}{W}
//
// Oracle text:
//
//	Choose a creature type. Destroy all creatures that aren't of the chosen type.
var KindredJudgment = newKindredJudgment

func newKindredJudgment() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Kindred Judgment",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Choose{
							Choice: game.ResolutionChoice{
								Kind:          game.ResolutionChoiceSubtype,
								Prompt:        "Choose a creature type",
								SubtypeOfType: types.Creature,
							},
							PublishChoice: game.SpellChosenTypeChoiceKey,
						},
					},
					{
						Primitive: game.Destroy{
							Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, SubtypeChoice: game.SubtypeChoiceResolutionExcluded}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Choose a creature type. Destroy all creatures that aren't of the chosen type.
		`,
		},
	}
}
