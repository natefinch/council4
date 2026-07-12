package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ShredMemory is the card definition for Shred Memory.
//
// Type: Instant
// Cost: {1}{B}
//
// Oracle text:
//
//	Exile up to four target cards from a single graveyard.
//	Transmute {1}{B}{B} ({1}{B}{B}, Discard this card: Search your library for a card with the same mana value as this card, reveal it, put it into your hand, then shuffle. Transmute only as a sorcery.)
var ShredMemory = newShredMemory

func newShredMemory() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Shred Memory",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Instant},
			ActivatedAbilities: []game.ActivatedAbility{
				game.TransmuteActivatedAbility(cost.Mana{cost.O(1), cost.B, cost.B}, 2),
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets:    0,
						MaxTargets:    4,
						Constraint:    "up to four target cards from a single graveyard",
						Allow:         game.TargetAllowCard,
						TargetZone:    zone.Graveyard,
						Selection:     opt.Val(game.Selection{}),
						SameGraveyard: true,
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
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 2},
							FromZone:    zone.Graveyard,
							Destination: zone.Exile,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 3},
							FromZone:    zone.Graveyard,
							Destination: zone.Exile,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Exile up to four target cards from a single graveyard.
			Transmute {1}{B}{B} ({1}{B}{B}, Discard this card: Search your library for a card with the same mana value as this card, reveal it, put it into your hand, then shuffle. Transmute only as a sorcery.)
		`,
		},
	}
}
