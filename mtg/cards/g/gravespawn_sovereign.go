package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GravespawnSovereign is the card definition for Gravespawn Sovereign.
//
// Type: Creature — Zombie
// Cost: {4}{B}{B}
//
// Oracle text:
//
//	Tap five untapped Zombies you control: Put target creature card from a graveyard onto the battlefield under your control.
var GravespawnSovereign = newGravespawnSovereign

func newGravespawnSovereign() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Gravespawn Sovereign",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Tap five untapped Zombies you control: Put target creature card from a graveyard onto the battlefield under your control.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTapPermanents,
							Text:        "Tap five untapped Zombies you control",
							Amount:      5,
							SubtypesAny: cost.SubtypeSet{0: types.Zombie},
							CounterKind: counter.PlusOnePlusOne,
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
										Constraint: "target creature card from a graveyard",
										Allow:      game.TargetAllowCard,
										Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
										TargetZone: zone.Graveyard,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.PutOnBattlefield{
											Source:    game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
											Recipient: opt.Val(game.ControllerReference()),
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
			Tap five untapped Zombies you control: Put target creature card from a graveyard onto the battlefield under your control.
		`,
		},
	}
}
