package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// FootMystic is the card definition for Foot Mystic.
//
// Type: Creature — Human Ninja Warlock
// Cost: {3}{B}
//
// Oracle text:
//
//	Lifelink
//	Disappear — When this creature enters, if a permanent left the battlefield under your control this turn, create a 1/1 black Ninja creature token.
var FootMystic = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name: "Foot Mystic",
		ManaCost: opt.Val(cost.Mana{
			cost.O(3),
			cost.B,
		}),
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Human, types.Ninja, types.Warlock},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 4}),
		StaticAbilities: []game.StaticAbility{
			game.LifelinkStaticBody,
		},
		TriggeredAbilities: []game.TriggeredAbility{
			game.TriggeredAbility{
				Trigger: game.TriggerCondition{
					Type: game.TriggerWhen,
					Pattern: game.TriggerPattern{
						Event:  game.EventPermanentEnteredBattlefield,
						Source: game.TriggerSourceSelf,
					},
					InterveningIf: "if a permanent left the battlefield under your control this turn",
					InterveningCondition: opt.Val(game.Condition{
						EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
							Event:         game.EventZoneChanged,
							Controller:    game.TriggerControllerYou,
							MatchFromZone: true,
							FromZone:      zone.Battlefield,
						}, Window: game.EventHistoryCurrentTurn}),
					}),
				},
				Content: game.Mode{
					Sequence: []game.Instruction{
						{
							Primitive: game.CreateToken{
								Amount: game.Fixed(1),
								Source: game.TokenDef(footMysticToken),
							},
						},
					},
				}.Ability(),
			},
		},
		OracleText: `
			Lifelink
			Disappear — When this creature enters, if a permanent left the battlefield under your control this turn, create a 1/1 black Ninja creature token.
		`,
	},
}

var footMysticToken = newFootMysticToken()

func newFootMysticToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Ninja",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Ninja},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
