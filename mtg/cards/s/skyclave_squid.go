package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SkyclaveSquid is the card definition for Skyclave Squid.
//
// Type: Creature — Squid
// Cost: {1}{U}
//
// Oracle text:
//
//	Defender
//	Landfall — Whenever a land you control enters, this creature can attack this turn as though it didn't have defender.
var SkyclaveSquid = newSkyclaveSquid()

func newSkyclaveSquid() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Skyclave Squid",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Squid},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentEnteredBattlefield,
							Controller:       game.TriggerControllerYou,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Land}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyRule{
									Object: opt.Val(game.SourcePermanentReference()),
									RuleEffects: []game.RuleEffect{
										game.RuleEffect{
											Kind: game.RuleEffectCanAttackAsThoughDefender,
										},
									},
									Duration: game.DurationThisTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Defender
			Landfall — Whenever a land you control enters, this creature can attack this turn as though it didn't have defender.
		`,
		},
	}
}
