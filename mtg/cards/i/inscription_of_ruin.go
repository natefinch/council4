package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// InscriptionOfRuin is the card definition for Inscription of Ruin.
//
// Type: Sorcery
// Cost: {2}{B}
//
// Oracle text:
//
//	Kicker {2}{B}{B}
//	Choose one. If this spell was kicked, choose any number instead.
//	• Target opponent discards two cards.
//	• Return target creature card with mana value 2 or less from your graveyard to the battlefield.
//	• Destroy target creature with mana value 3 or less.
var InscriptionOfRuin = newInscriptionOfRuin

func newInscriptionOfRuin() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Inscription of Ruin",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.KickerKeyword{Cost: cost.Mana{cost.O(2), cost.B, cost.B}},
					},
				},
			},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Target opponent discards two cards.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "Target opponent",
								Allow:      game.TargetAllowPlayer,
								Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Discard{
									Amount: game.Fixed(2),
									Player: game.TargetPlayerReference(0),
								},
							},
						},
					},
					game.Mode{
						Text: "Return target creature card with mana value 2 or less from your graveyard to the battlefield.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature card with mana value 2 or less from your graveyard",
								Allow:      game.TargetAllowCard,
								TargetZone: zone.Graveyard,
								Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, ManaValue: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 2})}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.PutOnBattlefield{
									Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
								},
							},
						},
					},
					game.Mode{
						Text: "Destroy target creature with mana value 3 or less.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature with mana value 3 or less",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, ManaValue: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 3})}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Destroy{
									Object: game.TargetPermanentReference(0),
								},
							},
						},
					},
				},
				MinModes:        1,
				MaxModes:        1,
				ModeChoiceBonus: game.ModeChoiceBonus{Condition: game.ModeChoiceConditionSpellKicked, ReplaceRange: true, MinModes: 0, MaxModes: 3},
			}),
			OracleText: `
			Kicker {2}{B}{B}
			Choose one. If this spell was kicked, choose any number instead.
			• Target opponent discards two cards.
			• Return target creature card with mana value 2 or less from your graveyard to the battlefield.
			• Destroy target creature with mana value 3 or less.
		`,
		},
	}
}
