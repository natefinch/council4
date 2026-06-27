package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Illusion
//
// Type: Token Creature — Illusion
//
// Oracle text:
//   Whenever this creature blocks a creature, that creature doesn't untap during its controller's next untap step.

// IllusionToken349852d6169748cf86a4c20a2aabf0fc is the card definition for Illusion.
var IllusionToken349852d6169748cf86a4c20a2aabf0fc = newIllusionToken349852d6169748cf86a4c20a2aabf0fc()

func newIllusionToken349852d6169748cf86a4c20a2aabf0fc() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:      "Illusion",
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Illusion},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                   game.EventBlockerDeclared,
							Source:                  game.TriggerSourceSelf,
							RelatedSubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.SkipNextUntap{
									Object: game.EventRelatedPermanentReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever this creature blocks a creature, that creature doesn't untap during its controller's next untap step.
		`,
		},
	}
}
