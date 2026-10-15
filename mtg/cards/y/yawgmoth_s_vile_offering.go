package y

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// YawgmothSVileOffering is the card definition for Yawgmoth's Vile Offering.
//
// Type: Legendary Sorcery
// Cost: {4}{B}
//
// Oracle text:
//
//	(You may cast a legendary sorcery only if you control a legendary creature or planeswalker.)
//	Put up to one target creature or planeswalker card from a graveyard onto the battlefield under your control. Destroy up to one target creature or planeswalker. Exile Yawgmoth's Vile Offering.
var YawgmothSVileOffering = newYawgmothSVileOffering

func newYawgmothSVileOffering() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Yawgmoth's Vile Offering",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
			}),
			Colors:     []color.Color{color.Black},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 0,
						MaxTargets: 1,
						Constraint: "up to one target creature or planeswalker card from a graveyard",
						Allow:      game.TargetAllowCard,
						TargetZone: zone.Graveyard,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker}}),
					},
					game.TargetSpec{
						MinTargets: 0,
						MaxTargets: 1,
						Constraint: "up to one target creature or planeswalker",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.PutOnBattlefield{
							Source:    game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
							Recipient: opt.Val(game.ControllerReference()),
						},
					},
					{
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(1),
						},
					},
					{
						Primitive: game.MoveResolvingSpell{
							Destination: zone.Exile,
						},
					},
				},
			}.Ability()),
			OracleText: `
			(You may cast a legendary sorcery only if you control a legendary creature or planeswalker.)
			Put up to one target creature or planeswalker card from a graveyard onto the battlefield under your control. Destroy up to one target creature or planeswalker. Exile Yawgmoth's Vile Offering.
		`,
		},
	}
}
