package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// StrengthOfSolidarity is the card definition for Strength of Solidarity.
//
// Type: Sorcery
// Cost: {G}
//
// Oracle text:
//
//	Choose target creature you control. Put a +1/+1 counter on it for each creature in your party. (Your party consists of up to one each of Cleric, Rogue, Warrior, and Wizard.)
var StrengthOfSolidarity = newStrengthOfSolidarity

func newStrengthOfSolidarity() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Strength of Solidarity",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature you control",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.AddCounter{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountPartySize,
								Multiplier: 1,
							}),
							Object:      game.TargetPermanentReference(0),
							CounterKind: counter.PlusOnePlusOne,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Choose target creature you control. Put a +1/+1 counter on it for each creature in your party. (Your party consists of up to one each of Cleric, Rogue, Warrior, and Wizard.)
		`,
		},
	}
}
