package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MoxAmber is the card definition for Mox Amber.
//
// Type: Legendary Artifact
// Cost: {0}
//
// Oracle text:
//
//	{T}: Add one mana of any color among legendary creatures and planeswalkers you control.
var MoxAmber = newMoxAmber

func newMoxAmber() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Mox Amber",
			ManaCost: opt.Val(cost.Mana{
				cost.O(0),
			}),
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAmongControlledColorsAbility("{T}: Add one mana of any color among legendary creatures and planeswalkers you control.", game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker}, Supertypes: []types.Super{types.Legendary}, Controller: game.ControllerYou}),
			},
			OracleText: `
			{T}: Add one mana of any color among legendary creatures and planeswalkers you control.
		`,
		},
	}
}
