package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// LittjaraGladeWarden is the card definition for Littjara Glade-Warden.
//
// Type: Creature — Shapeshifter
// Cost: {3}{G}
//
// Oracle text:
//
//	Changeling (This card is every creature type.)
//	{2}{G}, {T}, Exile a creature card from your graveyard: Put two +1/+1 counters on target creature. Activate only as a sorcery.
var LittjaraGladeWarden = newLittjaraGladeWarden

func newLittjaraGladeWarden() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Littjara Glade-Warden",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Shapeshifter},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.ChangelingStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{2}{G}, {T}, Exile a creature card from your graveyard: Put two +1/+1 counters on target creature. Activate only as a sorcery.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
						cost.G,
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:          cost.AdditionalExile,
							Text:          "Exile a creature card from your graveyard",
							Amount:        1,
							MatchCardType: true,
							CardType:      types.Creature,
							Source:        zone.Graveyard,
							CounterKind:   counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Timing:         game.SorceryOnly,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target creature",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(2),
											Object:      game.TargetPermanentReference(0),
											CounterKind: counter.PlusOnePlusOne,
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
			Changeling (This card is every creature type.)
			{2}{G}, {T}, Exile a creature card from your graveyard: Put two +1/+1 counters on target creature. Activate only as a sorcery.
		`,
		},
	}
}
