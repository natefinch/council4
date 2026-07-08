package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BlackOakOfOdunos is the card definition for Black Oak of Odunos.
//
// Type: Creature — Zombie Treefolk
// Cost: {2}{B}
//
// Oracle text:
//
//	Defender
//	{B}, Tap another untapped creature you control: This creature gets +1/+1 until end of turn.
var BlackOakOfOdunos = newBlackOakOfOdunos

func newBlackOakOfOdunos() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Black Oak of Odunos",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie, types.Treefolk},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{B}, Tap another untapped creature you control: This creature gets +1/+1 until end of turn.",
					ManaCost: opt.Val(cost.Mana{cost.B}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:               cost.AdditionalTapPermanents,
							Text:               "Tap another untapped creature you control",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Creature,
							ExcludeSource:      true,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ModifyPT{
									Object:         game.SourcePermanentReference(),
									PowerDelta:     game.Fixed(1),
									ToughnessDelta: game.Fixed(1),
									Duration:       game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Defender
			{B}, Tap another untapped creature you control: This creature gets +1/+1 until end of turn.
		`,
		},
	}
}
