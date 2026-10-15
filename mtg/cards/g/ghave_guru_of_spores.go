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

// GhaveGuruOfSpores is the card definition for Ghave, Guru of Spores.
//
// Type: Legendary Creature — Fungus Shaman
// Cost: {2}{W}{B}{G}
//
// Oracle text:
//
//	Ghave enters with five +1/+1 counters on it.
//	{1}, Remove a +1/+1 counter from a creature you control: Create a 1/1 green Saproling creature token.
//	{1}, Sacrifice a creature: Put a +1/+1 counter on target creature.
var GhaveGuruOfSpores = newGhaveGuruOfSpores

func newGhaveGuruOfSpores() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Ghave, Guru of Spores",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.B,
				cost.G,
			}),
			Colors:     []color.Color{color.Black, color.Green, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Fungus, types.Shaman},
			Power:      opt.Val(game.PT{Value: 0}),
			Toughness:  opt.Val(game.PT{Value: 0}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}, Remove a +1/+1 counter from a creature you control: Create a 1/1 green Saproling creature token.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:               cost.AdditionalRemoveCounterAmong,
							Text:               "Remove a +1/+1 counter from a creature you control",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Creature,
							CounterKind:        counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(ghaveGuruOfSporesToken),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.ActivatedAbility{
					Text: "{1}, Sacrifice a creature: Put a +1/+1 counter on target creature.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:               cost.AdditionalSacrifice,
							Text:               "Sacrifice a creature",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Creature,
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
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
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
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("Ghave enters with five +1/+1 counters on it.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 5}),
			},
			OracleText: `
			Ghave enters with five +1/+1 counters on it.
			{1}, Remove a +1/+1 counter from a creature you control: Create a 1/1 green Saproling creature token.
			{1}, Sacrifice a creature: Put a +1/+1 counter on target creature.
		`,
		},
	}
}

var ghaveGuruOfSporesToken = newGhaveGuruOfSporesToken()

func newGhaveGuruOfSporesToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Saproling",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Saproling},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
