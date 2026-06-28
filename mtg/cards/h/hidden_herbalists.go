package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// HiddenHerbalists is the card definition for Hidden Herbalists.
//
// Type: Creature — Human Druid
// Cost: {1}{G}
//
// Oracle text:
//
//	Revolt — When this creature enters, if a permanent left the battlefield under your control this turn, add {G}{G}.
var HiddenHerbalists = newHiddenHerbalists()

func newHiddenHerbalists() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Hidden Herbalists",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Druid},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
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
								Primitive: game.AddMana{
									Amount:    game.Fixed(1),
									ManaColor: mana.G,
								},
							},
							{
								Primitive: game.AddMana{
									Amount:    game.Fixed(1),
									ManaColor: mana.G,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Revolt — When this creature enters, if a permanent left the battlefield under your control this turn, add {G}{G}.
		`,
		},
	}
}
