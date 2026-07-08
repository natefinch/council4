package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VultureSchemingScavenger is the card definition for Vulture, Scheming Scavenger.
//
// Type: Legendary Creature — Human Artificer Villain
// Cost: {5}{U/B}
//
// Oracle text:
//
//	Flying
//	Whenever Vulture attacks, other Villains you control gain flying until end of turn.
var VultureSchemingScavenger = newVultureSchemingScavenger

func newVultureSchemingScavenger() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Vulture, Scheming Scavenger",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.HybridMana(mana.U, mana.B),
			}),
			Colors:     []color.Color{color.Black, color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Artificer, types.Villain},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventAttackerDeclared,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											Group: game.BattlefieldGroupExcluding(game.Selection{SubtypesAny: []types.Sub{types.Sub("Villain")}, Controller: game.ControllerYou}, game.SourcePermanentReference()),
											AddKeywords: []game.Keyword{
												game.Flying,
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying
			Whenever Vulture attacks, other Villains you control gain flying until end of turn.
		`,
		},
	}
}
