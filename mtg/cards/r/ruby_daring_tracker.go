package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RubyDaringTracker is the card definition for Ruby, Daring Tracker.
//
// Type: Legendary Creature — Human Scout
// Cost: {R}{G}
//
// Oracle text:
//
//	Haste (This creature can attack and {T} as soon as it comes under your control.)
//	Whenever Ruby attacks while you control a creature with power 4 or greater, Ruby gets +2/+2 until end of turn.
//	{T}: Add {R} or {G}.
var RubyDaringTracker = newRubyDaringTracker

func newRubyDaringTracker() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Ruby, Daring Tracker",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
				cost.G,
			}),
			Colors:     []color.Color{color.Green, color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Scout},
			Power:      opt.Val(game.PT{Value: 1}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.R, mana.G),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventAttackerDeclared,
							Source: game.TriggerSourceSelf,
						},
						InterveningIf: "while you control a creature with power 4 or greater",
						InterveningCondition: opt.Val(game.Condition{
							ControlsMatching: opt.Val(game.SelectionCount{
								Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Power: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4})},
							}),
						}),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ModifyPT{
									Object:         game.SourcePermanentReference(),
									PowerDelta:     game.Fixed(2),
									ToughnessDelta: game.Fixed(2),
									Duration:       game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Haste (This creature can attack and {T} as soon as it comes under your control.)
			Whenever Ruby attacks while you control a creature with power 4 or greater, Ruby gets +2/+2 until end of turn.
			{T}: Add {R} or {G}.
		`,
		},
	}
}
