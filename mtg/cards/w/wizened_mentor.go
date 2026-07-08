package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WizenedMentor is the card definition for Wizened Mentor.
//
// Type: Creature — Zombie Cleric
// Cost: {1}{W}
//
// Oracle text:
//
//	Whenever an opponent activates an ability of a permanent that isn't a mana ability, you create a 1/1 white Zombie creature token. This ability triggers only once each turn.
var WizenedMentor = newWizenedMentor

func newWizenedMentor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Wizened Mentor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie, types.Cleric},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:              game.EventAbilityActivated,
							Player:             game.TriggerPlayerOpponent,
							ExcludeManaAbility: true,
						},
					},
					MaxTriggersPerTurn: 1,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(wizenedMentorToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever an opponent activates an ability of a permanent that isn't a mana ability, you create a 1/1 white Zombie creature token. This ability triggers only once each turn.
		`,
		},
	}
}

var wizenedMentorToken = newWizenedMentorToken()

func newWizenedMentorToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Zombie",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
