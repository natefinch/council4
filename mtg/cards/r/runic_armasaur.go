package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RunicArmasaur is the card definition for Runic Armasaur.
//
// Type: Creature — Dinosaur
// Cost: {1}{G}{G}
//
// Oracle text:
//
//	Whenever an opponent activates an ability of a creature or land that isn't a mana ability, you may draw a card.
var RunicArmasaur = newRunicArmasaur

func newRunicArmasaur() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Runic Armasaur",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dinosaur},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 5}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever an opponent activates an ability of a creature or land that isn't a mana ability, you may draw a card.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:              game.EventAbilityActivated,
							Player:             game.TriggerPlayerOpponent,
							SubjectSelection:   game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Land}},
							StackObjectKind:    game.StackSpell,
							ExcludeManaAbility: true,
							CounterKind:        counter.PlusOnePlusOne,
						},
					},
					Optional: true,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			Whenever an opponent activates an ability of a creature or land that isn't a mana ability, you may draw a card.
		`,
		},
	}
}
