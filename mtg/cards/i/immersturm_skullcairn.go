package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ImmersturmSkullcairn is the card definition for Immersturm Skullcairn.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped.
//	{T}: Add {B}.
//	{1}{B}{R}{R}, {T}, Sacrifice this land: It deals 3 damage to target player. That player discards a card. Activate only as a sorcery.
var ImmersturmSkullcairn = newImmersturmSkullcairn

func newImmersturmSkullcairn() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name:  "Immersturm Skullcairn",
			Types: []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{1}{B}{R}{R}, {T}, Sacrifice this land: It deals 3 damage to target player. That player discards a card. Activate only as a sorcery.",
					ManaCost: opt.Val(cost.Mana{cost.O(1), cost.B, cost.R, cost.R}),
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:   cost.AdditionalSacrificeSource,
							Text:   "Sacrifice this land",
							Amount: 1,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Timing:         game.SorceryOnly,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target player",
								Allow:      game.TargetAllowPlayer,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:       game.Fixed(3),
									Recipient:    game.AnyTargetDamageRecipient(0),
									DamageSource: opt.Val(game.SourcePermanentReference()),
								},
							},
							{
								Primitive: game.Discard{
									Amount: game.Fixed(1),
									Player: game.TargetPlayerReference(0),
								},
							},
						},
					}.Ability(),
				},
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.B),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			This land enters tapped.
			{T}: Add {B}.
			{1}{B}{R}{R}, {T}, Sacrifice this land: It deals 3 damage to target player. That player discards a card. Activate only as a sorcery.
		`,
		},
	}
}
