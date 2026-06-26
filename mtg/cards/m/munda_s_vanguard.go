package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MundaSVanguard is the card definition for Munda's Vanguard.
//
// Type: Creature — Kor Knight Ally
// Cost: {4}{W}
//
// Oracle text:
//
//	Cohort — {T}, Tap an untapped Ally you control: Put a +1/+1 counter on each creature you control.
var MundaSVanguard = newMundaSVanguard()

func newMundaSVanguard() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Munda's Vanguard",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Kor, types.Knight, types.Ally},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Cohort — {T}, Tap an untapped Ally you control: Put a +1/+1 counter on each creature you control.",
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:        cost.AdditionalTapPermanents,
							Text:        "Tap an untapped Ally you control",
							Amount:      1,
							SubtypesAny: cost.SubtypeSet{types.Ally},
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Cohort — {T}, Tap an untapped Ally you control: Put a +1/+1 counter on each creature you control.
		`,
		},
	}
}
