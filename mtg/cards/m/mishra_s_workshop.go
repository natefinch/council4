package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MishraSWorkshop is the card definition for Mishra's Workshop.
//
// Type: Land
//
// Oracle text:
//
//	{T}: Add {C}{C}{C}. Spend this mana only to cast artifact spells.
var MishraSWorkshop = newMishraSWorkshop()

func newMishraSWorkshop() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Mishra's Workshop",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					AdditionalCosts: cost.Tap,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddMana{
									Amount:    game.Fixed(1),
									ManaColor: mana.C,
									SpendRider: opt.Val(game.ManaSpendRider{
										Condition:   game.ManaSpendCastArtifactSpellOnly,
										Restriction: game.ManaSpendRestrictedToCondition,
									}),
								},
							},
							{
								Primitive: game.AddMana{
									Amount:    game.Fixed(1),
									ManaColor: mana.C,
									SpendRider: opt.Val(game.ManaSpendRider{
										Condition:   game.ManaSpendCastArtifactSpellOnly,
										Restriction: game.ManaSpendRestrictedToCondition,
									}),
								},
							},
							{
								Primitive: game.AddMana{
									Amount:    game.Fixed(1),
									ManaColor: mana.C,
									SpendRider: opt.Val(game.ManaSpendRider{
										Condition:   game.ManaSpendCastArtifactSpellOnly,
										Restriction: game.ManaSpendRestrictedToCondition,
									}),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{T}: Add {C}{C}{C}. Spend this mana only to cast artifact spells.
		`,
		},
	}
}
