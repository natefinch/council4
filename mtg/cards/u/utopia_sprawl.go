package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// UtopiaSprawl is the card definition for Utopia Sprawl.
//
// Type: Enchantment — Aura
// Cost: {G}
//
// Oracle text:
//
//	Enchant Forest
//	As this Aura enters, choose a color.
//	Whenever enchanted Forest is tapped for mana, its controller adds an additional one mana of the chosen color.
var UtopiaSprawl = newUtopiaSprawl

func newUtopiaSprawl() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Utopia Sprawl",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
			}),
			Colors:   []color.Color{color.Green},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "forest",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Sub("Forest")}}),
				}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever enchanted Forest is tapped for mana, its controller adds an additional one mana of the chosen color.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                game.EventPermanentTapped,
							Source:               game.TriggerSourceAttachedPermanent,
							SubjectSelection:     game.Selection{SubtypesAny: []types.Sub{types.Forest}},
							StackObjectKind:      game.StackSpell,
							CounterKind:          counter.PlusOnePlusOne,
							RequireTappedForMana: true,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddMana{
											Amount:          game.Fixed(1),
											EntryChoiceFrom: game.EntryColorChoiceKey,
											Player:          opt.Val(game.ObjectControllerReference(game.EventPermanentReference())),
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
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntryColorChoiceReplacement("As this Aura enters, choose a color."),
			},
			OracleText: `
			Enchant Forest
			As this Aura enters, choose a color.
			Whenever enchanted Forest is tapped for mana, its controller adds an additional one mana of the chosen color.
		`,
		},
	}
}
