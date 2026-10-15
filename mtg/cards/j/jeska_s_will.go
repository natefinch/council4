package j

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// JeskaSWill is the card definition for Jeska's Will.
//
// Type: Sorcery
// Cost: {2}{R}
//
// Oracle text:
//
//	Choose one. If you control a commander as you cast this spell, you may choose both instead.
//	• Add {R} for each card in target opponent's hand.
//	• Exile the top three cards of your library. You may play them this turn.
var JeskaSWill = newJeskaSWill

func newJeskaSWill() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Jeska's Will",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Add {R} for each card in target opponent's hand.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target opponent's hand",
								Allow:      game.TargetAllowPlayer,
								Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.AddMana{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountCountCardsInZone,
										Multiplier: 1,
										Player:     func() *game.PlayerReference { ref := game.TargetPlayerReference(0); return &ref }(),
										CardZone:   zone.Hand,
										Selection:  &game.Selection{},
									}),
									ManaColor: mana.R,
								},
							},
						},
					},
					game.Mode{
						Text: "Exile the top three cards of your library. You may play them this turn.",
						Sequence: []game.Instruction{
							{
								Primitive: game.ImpulseExile{
									Player:   game.ControllerReference(),
									Amount:   game.Fixed(3),
									Duration: game.DurationThisTurn,
								},
							},
						},
					},
				},
				MinModes:        1,
				MaxModes:        1,
				ModeChoiceBonus: game.ModeChoiceBonus{Condition: game.ModeChoiceConditionControlsCommander, AdditionalMaxModes: 1},
			}),
			OracleText: `
			Choose one. If you control a commander as you cast this spell, you may choose both instead.
			• Add {R} for each card in target opponent's hand.
			• Exile the top three cards of your library. You may play them this turn.
		`,
		},
	}
}
