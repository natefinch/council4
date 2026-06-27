package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VertigoSpawn is the card definition for Vertigo Spawn.
//
// Type: Creature — Illusion
// Cost: {1}{U}
//
// Oracle text:
//
//	Defender
//	Whenever this creature blocks a creature, tap that creature. That creature doesn't untap during its controller's next untap step.
var VertigoSpawn = newVertigoSpawn()

func newVertigoSpawn() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Vertigo Spawn",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Illusion},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
			},
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
								Primitive: game.Tap{
									Object: game.EventRelatedPermanentReference(),
								},
							},
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
			Defender
			Whenever this creature blocks a creature, tap that creature. That creature doesn't untap during its controller's next untap step.
		`,
		},
	}
}
