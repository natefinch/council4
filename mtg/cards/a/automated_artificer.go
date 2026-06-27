package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AutomatedArtificer is the card definition for Automated Artificer.
//
// Type: Artifact Creature — Artificer
// Cost: {2}
//
// Oracle text:
//
//	{T}: Add {C}. Spend this mana only to activate an ability or cast an artifact spell.
var AutomatedArtificer = newAutomatedArtificer()

func newAutomatedArtificer() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Automated Artificer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Artificer},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 3}),
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
										Condition:   game.ManaSpendCastArtifactOrActivateAbility,
										Restriction: game.ManaSpendRestrictedToCondition,
									}),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{T}: Add {C}. Spend this mana only to activate an ability or cast an artifact spell.
		`,
		},
	}
}
