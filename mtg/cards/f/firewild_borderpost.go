package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FirewildBorderpost is the card definition for Firewild Borderpost.
//
// Type: Artifact
// Cost: {1}{R}{G}
//
// Oracle text:
//
//	You may pay {1} and return a basic land you control to its owner's hand rather than pay this spell's mana cost.
//	This artifact enters tapped.
//	{T}: Add {R} or {G}.
var FirewildBorderpost = newFirewildBorderpost

func newFirewildBorderpost() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Firewild Borderpost",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
				cost.G,
			}),
			Colors: []color.Color{color.Green, color.Red},
			Types:  []types.Card{types.Artifact},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.R, mana.G),
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
			{T}: Add {R} or {G}.
		`,
		},
	}
}
