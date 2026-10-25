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

// LaurineTheDiversion is the card definition for Laurine, the Diversion.
//
// Type: Legendary Creature — Human Rogue
// Cost: {2}{R}
//
// Oracle text:
//
//	Partner with Kamber, the Plunderer (When this creature enters, target player may put Kamber into their hand from their library, then shuffle.)
//	First strike
//	{2}, Sacrifice an artifact or creature: Goad target creature. (Until your next turn, that creature attacks each combat if able and attacks a player other than you if able.)
var LaurineTheDiversion = newLaurineTheDiversion

func newLaurineTheDiversion() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Laurine, the Diversion",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors:     []color.Color{color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Rogue},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.PartnerWithStaticBody,
				game.FirstStrikeStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.PartnerWithTriggeredAbility("Kamber, the Plunderer"),
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{2}, Sacrifice an artifact or creature: Goad target creature. (Until your next turn, that creature attacks each combat if able and attacks a player other than you if able.)",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:               cost.AdditionalSacrifice,
							Text:               "Sacrifice an artifact or creature",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Artifact,
							PermanentTypeAlt:   types.Creature,
							CounterKind:        counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
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
										Primitive: game.Goad{
											Object: game.TargetPermanentReference(0),
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
			Partner with Kamber, the Plunderer (When this creature enters, target player may put Kamber into their hand from their library, then shuffle.)
			First strike
			{2}, Sacrifice an artifact or creature: Goad target creature. (Until your next turn, that creature attacks each combat if able and attacks a player other than you if able.)
		`,
		},
	}
}
