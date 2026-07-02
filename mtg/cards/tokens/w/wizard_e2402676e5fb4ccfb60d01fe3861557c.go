package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Wizard
//
// Type: Token Creature — Wizard
//
// Oracle text:
//   {T}: Add {R}. Spend this mana only to cast a planeswalker spell.

// WizardTokene2402676e5fb4ccfb60d01fe3861557c is the card definition for Wizard.
var WizardTokene2402676e5fb4ccfb60d01fe3861557c = newWizardTokene2402676e5fb4ccfb60d01fe3861557c()

func newWizardTokene2402676e5fb4ccfb60d01fe3861557c() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:      "Wizard",
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wizard},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					AdditionalCosts: cost.Tap,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddMana{
									Amount:    game.Fixed(1),
									ManaColor: mana.R,
									SpendRider: opt.Val(game.ManaSpendRider{
										Condition:   game.ManaSpendCastPlaneswalkerSpell,
										Restriction: game.ManaSpendRestrictedToCondition,
									}),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{T}: Add {R}. Spend this mana only to cast a planeswalker spell.
		`,
		},
	}
}
