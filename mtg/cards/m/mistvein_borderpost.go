package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MistveinBorderpost is the card definition for Mistvein Borderpost.
//
// Type: Artifact
// Cost: {1}{U}{B}
//
// Oracle text:
//
//	You may pay {1} and return a basic land you control to its owner's hand rather than pay this spell's mana cost.
//	This artifact enters tapped.
//	{T}: Add {U} or {B}.
var MistveinBorderpost = newMistveinBorderpost

func newMistveinBorderpost() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Mistvein Borderpost",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
				cost.B,
			}),
			Colors: []color.Color{color.Black, color.Blue},
			Types:  []types.Card{types.Artifact},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.U, mana.B),
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
			{T}: Add {U} or {B}.
		`,
		},
	}
}
