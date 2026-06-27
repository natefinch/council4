package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AnowonTheRuinSage is the card definition for Anowon, the Ruin Sage.
//
// Type: Legendary Creature — Vampire Shaman
// Cost: {3}{B}{B}
//
// Oracle text:
//
//	At the beginning of your upkeep, each player sacrifices a non-Vampire creature of their choice.
var AnowonTheRuinSage = newAnowonTheRuinSage()

func newAnowonTheRuinSage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Anowon, the Ruin Sage",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
				cost.B,
			}),
			Colors:     []color.Color{color.Black},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Vampire, types.Shaman},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepUpkeep,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.SacrificePermanents{
									Amount:      game.Fixed(1),
									PlayerGroup: game.AllPlayersReference(),
									Selection:   game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludedSubtype: types.Sub("Vampire")},
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			At the beginning of your upkeep, each player sacrifices a non-Vampire creature of their choice.
		`,
		},
	}
}
