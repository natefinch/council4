package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WildfieldBorderpost is the card definition for Wildfield Borderpost.
//
// Type: Artifact
// Cost: {1}{G}{W}
//
// Oracle text:
//
//	You may pay {1} and return a basic land you control to its owner's hand rather than pay this spell's mana cost.
//	This artifact enters tapped.
//	{T}: Add {G} or {W}.
var WildfieldBorderpost = newWildfieldBorderpost

func newWildfieldBorderpost() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Wildfield Borderpost",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
				cost.W,
			}),
			Colors: []color.Color{color.Green, color.White},
			Types:  []types.Card{types.Artifact},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.G, mana.W),
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
			{T}: Add {G} or {W}.
		`,
		},
	}
}
