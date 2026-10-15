package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// HammerOfBogardan is the card definition for Hammer of Bogardan.
//
// Type: Sorcery
// Cost: {1}{R}{R}
//
// Oracle text:
//
//	Hammer of Bogardan deals 3 damage to any target.
//	{2}{R}{R}{R}: Return this card from your graveyard to your hand. Activate only during your upkeep.
var HammerOfBogardan = newHammerOfBogardan

func newHammerOfBogardan() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Hammer of Bogardan",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{2}{R}{R}{R}: Return this card from your graveyard to your hand. Activate only during your upkeep.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
						cost.R,
						cost.R,
						cost.R,
					}),
					ZoneOfFunction: zone.Graveyard,
					Timing:         game.DuringUpkeep,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MoveCard{
											Card: game.CardReference{
												Kind: game.CardReferenceSource,
											},
											FromZone:    zone.Graveyard,
											Destination: zone.Hand,
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
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "any target",
						Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker, types.Battle}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(3),
							Recipient: game.AnyTargetDamageRecipient(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Hammer of Bogardan deals 3 damage to any target.
			{2}{R}{R}{R}: Return this card from your graveyard to your hand. Activate only during your upkeep.
		`,
		},
	}
}
