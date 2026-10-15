package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VeinfireBorderpost is the card definition for Veinfire Borderpost.
//
// Type: Artifact
// Cost: {1}{B}{R}
//
// Oracle text:
//
//	You may pay {1} and return a basic land you control to its owner's hand rather than pay this spell's mana cost.
//	This artifact enters tapped.
//	{T}: Add {B} or {R}.
var VeinfireBorderpost = newVeinfireBorderpost

func newVeinfireBorderpost() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Veinfire Borderpost",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
				cost.R,
			}),
			Colors: []color.Color{color.Black, color.Red},
			Types:  []types.Card{types.Artifact},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.B, mana.R),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This artifact enters tapped."),
			},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label:    "Pay {1} and return a basic land",
					ManaCost: opt.Val(cost.Mana{cost.O(1)}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:               cost.AdditionalReturnToHand,
							Text:               "return a basic land you control to its owner's hand",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Land,
							RequireSupertype:   types.Basic,
						},
					},
				},
			},
			OracleText: `
			You may pay {1} and return a basic land you control to its owner's hand rather than pay this spell's mana cost.
			This artifact enters tapped.
			{T}: Add {B} or {R}.
		`,
		},
	}
}
