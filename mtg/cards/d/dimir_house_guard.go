package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// DimirHouseGuard is the card definition for Dimir House Guard.
//
// Type: Creature — Skeleton
// Cost: {3}{B}
//
// Oracle text:
//
//	Fear (This creature can't be blocked except by artifact creatures and/or black creatures.)
//	Sacrifice a creature: Regenerate this creature.
//	Transmute {1}{B}{B} ({1}{B}{B}, Discard this card: Search your library for a card with the same mana value as this card, reveal it, put it into your hand, then shuffle. Transmute only as a sorcery.)
var DimirHouseGuard = newDimirHouseGuard

func newDimirHouseGuard() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Dimir House Guard",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Skeleton},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FearStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Sacrifice a creature: Regenerate this creature.",
					AdditionalCosts: []cost.Additional{
						{
							Kind:               cost.AdditionalSacrifice,
							Text:               "Sacrifice a creature",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Creature,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Regenerate{
									Object: game.SourcePermanentReference(),
								},
							},
						},
					}.Ability(),
				},
				game.TransmuteActivatedAbility(cost.Mana{cost.O(1), cost.B, cost.B}, 4),
			},
			OracleText: `
			Fear (This creature can't be blocked except by artifact creatures and/or black creatures.)
			Sacrifice a creature: Regenerate this creature.
			Transmute {1}{B}{B} ({1}{B}{B}, Discard this card: Search your library for a card with the same mana value as this card, reveal it, put it into your hand, then shuffle. Transmute only as a sorcery.)
		`,
		},
	}
}
