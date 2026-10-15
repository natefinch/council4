package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// FiendArtisan is the card definition for Fiend Artisan.
//
// Type: Creature — Nightmare
// Cost: {B/G}{B/G}
//
// Oracle text:
//
//	This creature gets +1/+1 for each creature card in your graveyard.
//	{X}{B/G}, {T}, Sacrifice another creature: Search your library for a creature card with mana value X or less, put it onto the battlefield, then shuffle. Activate only as a sorcery.
var FiendArtisan = newFiendArtisan

func newFiendArtisan() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Fiend Artisan",
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.B, mana.G),
				cost.HybridMana(mana.B, mana.G),
			}),
			Colors:    []color.Color{color.Black, color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Nightmare},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This creature gets +1/+1 for each creature card in your graveyard.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessModify,
							PowerDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:        game.DynamicAmountCountCardsInZone,
								Multiplier:  1,
								CounterKind: counter.PlusOnePlusOne,
								Player:      func() *game.PlayerReference { ref := game.ControllerReference(); return &ref }(),
								CardZone:    zone.Graveyard,
								Selection:   &game.Selection{RequiredTypes: []types.Card{types.Creature}},
							}),
							ToughnessDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:        game.DynamicAmountCountCardsInZone,
								Multiplier:  1,
								CounterKind: counter.PlusOnePlusOne,
								Player:      func() *game.PlayerReference { ref := game.ControllerReference(); return &ref }(),
								CardZone:    zone.Graveyard,
								Selection:   &game.Selection{RequiredTypes: []types.Card{types.Creature}},
							}),
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{X}{B/G}, {T}, Sacrifice another creature: Search your library for a creature card with mana value X or less, put it onto the battlefield, then shuffle. Activate only as a sorcery.",
					ManaCost: opt.Val(cost.Mana{
						cost.X,
						cost.HybridMana(mana.B, mana.G),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:               cost.AdditionalSacrifice,
							Text:               "Sacrifice another creature",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Creature,
							CounterKind:        counter.PlusOnePlusOne,
							ExcludeSource:      true,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Timing:         game.SorceryOnly,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Search{
											Player: game.ControllerReference(),
											Spec: game.SearchSpec{
												SourceZone:        zone.Library,
												Destination:       zone.Battlefield,
												Filter:            game.Selection{RequiredTypes: []types.Card{types.Creature}},
												MaxManaValueFromX: true,
											},
											Amount: game.Fixed(1),
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
			OracleText: `
			This creature gets +1/+1 for each creature card in your graveyard.
			{X}{B/G}, {T}, Sacrifice another creature: Search your library for a creature card with mana value X or less, put it onto the battlefield, then shuffle. Activate only as a sorcery.
		`,
		},
	}
}
