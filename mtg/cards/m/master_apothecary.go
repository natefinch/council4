package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MasterApothecary is the card definition for Master Apothecary.
//
// Type: Creature — Human Cleric
// Cost: {W}{W}{W}
//
// Oracle text:
//
//	Tap an untapped Cleric you control: Prevent the next 2 damage that would be dealt to any target this turn.
var MasterApothecary = newMasterApothecary

func newMasterApothecary() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Master Apothecary",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Cleric},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Tap an untapped Cleric you control: Prevent the next 2 damage that would be dealt to any target this turn.",
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalTapPermanents,
							Text:        "Tap an untapped Cleric you control",
							Amount:      1,
							SubtypesAny: cost.SubtypeSet{types.Cleric},
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "any target",
								Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.PreventDamage{
									AnyTarget: game.AnyTargetDamageRecipient(0),
									Amount:    game.Fixed(2),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Tap an untapped Cleric you control: Prevent the next 2 damage that would be dealt to any target this turn.
		`,
		},
	}
}
