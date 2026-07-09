package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Terrarion is the card definition for Terrarion.
//
// Type: Artifact
// Cost: {1}
//
// Oracle text:
//
//	This artifact enters tapped.
//	{2}, {T}, Sacrifice this artifact: Add two mana in any combination of colors.
//	When this artifact is put into a graveyard from the battlefield, draw a card.
var Terrarion = newTerrarion

func newTerrarion() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Terrarion",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
			}),
			Types: []types.Card{types.Artifact},
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					ManaCost: opt.Val(cost.Mana{cost.O(2)}),
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:   cost.AdditionalSacrificeSource,
							Text:   "Sacrifice this artifact",
							Amount: 1,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddMana{
									Amount:            game.Fixed(2),
									CombinationColors: []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G},
								},
							},
						},
					}.Ability(),
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:         game.EventZoneChanged,
							Source:        game.TriggerSourceSelf,
							MatchFromZone: true,
							FromZone:      zone.Battlefield,
							MatchToZone:   true,
							ToZone:        zone.Graveyard,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This artifact enters tapped."),
			},
			OracleText: `
			This artifact enters tapped.
			{2}, {T}, Sacrifice this artifact: Add two mana in any combination of colors.
			When this artifact is put into a graveyard from the battlefield, draw a card.
		`,
		},
	}
}
