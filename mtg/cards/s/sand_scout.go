package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SandScout is the card definition for Sand Scout.
//
// Type: Creature — Human Scout
// Cost: {1}{W}
//
// Oracle text:
//
//	When this creature enters, if an opponent controls more lands than you, search your library for a Desert card, put it onto the battlefield tapped, then shuffle.
//	Whenever one or more land cards are put into your graveyard from anywhere, create a 1/1 red, green, and white Sand Warrior creature token. This ability triggers only once each turn.
var SandScout = newSandScout()

func newSandScout() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Sand Scout",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Scout},
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
						InterveningIf: "if an opponent controls more lands than you",
						InterveningCondition: opt.Val(game.Condition{
							ControlComparison: opt.Val(game.ControlCountComparison{
								Selection: game.Selection{RequiredTypes: []types.Card{types.Land}},
								Left:      game.ControlPlayerAnyOpponent,
								Right:     game.ControlPlayerController,
								Op:        compare.GreaterThan,
							}),
						}),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Search{
									Player: game.ControllerReference(),
									Spec: game.SearchSpec{
										SourceZone:   zone.Library,
										Destination:  zone.Battlefield,
										Filter:       game.Selection{SubtypesAny: []types.Sub{types.Sub("Desert")}},
										EntersTapped: true,
									},
									Amount: game.Fixed(1),
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventZoneChanged,
							Player:           game.TriggerPlayerYou,
							MatchToZone:      true,
							ToZone:           zone.Graveyard,
							OneOrMore:        true,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Land}},
						},
					},
					MaxTriggersPerTurn: 1,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(sandScoutToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this creature enters, if an opponent controls more lands than you, search your library for a Desert card, put it onto the battlefield tapped, then shuffle.
			Whenever one or more land cards are put into your graveyard from anywhere, create a 1/1 red, green, and white Sand Warrior creature token. This ability triggers only once each turn.
		`,
		},
	}
}

var sandScoutToken = newSandScoutToken()

func newSandScoutToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Sand Warrior",
			Colors:    []color.Color{color.Red, color.Green, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Sand, types.Warrior},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
