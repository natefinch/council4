package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Cease is the card definition for Cease // Desist.
//
// Type: Instant // Sorcery
// Cost: {1}{B/G} // {4}{G/W}{G/W}
// Face: Desist — Sorcery ({4}{G/W}{G/W})
//
// Oracle text:
//
//	Exile up to two target cards from a single graveyard. Target player gains 2 life and draws a card.
var Cease = newCease()

func newCease() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Cease",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.HybridMana(mana.B, mana.G),
			}),
			Colors: []color.Color{color.Black, color.Green},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 0,
						MaxTargets: 2,
						Constraint: "up to two target cards from a single graveyard",
						Allow:      game.TargetAllowCard,
						TargetZone: zone.Graveyard,
						Selection:  opt.Val(game.Selection{}),
					},
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "Target player",
						Allow:      game.TargetAllowPlayer,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget},
							FromZone:    zone.Graveyard,
							Destination: zone.Exile,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 1},
							FromZone:    zone.Graveyard,
							Destination: zone.Exile,
						},
					},
					{
						Primitive: game.GainLife{
							Amount: game.Fixed(2),
							Player: game.TargetPlayerReference(1),
						},
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(1),
							Player: game.TargetPlayerReference(1),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Exile up to two target cards from a single graveyard. Target player gains 2 life and draws a card.
		`,
		},
		Layout: game.LayoutSplit,
		Alternate: opt.Val(game.CardFace{
			Name: "Desist",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.HybridMana(mana.G, mana.W),
				cost.HybridMana(mana.G, mana.W),
			}),
			Colors: []color.Color{color.White, color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Group: game.BattlefieldGroup(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment}}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy all artifacts and enchantments.
		`,
		}),
	}
}
