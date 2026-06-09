package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// CinderGlade is the card definition for Cinder Glade.
//
// Type: Land — Mountain Forest
//
// Oracle text:
//
//	({T}: Add {R} or {G}.)
//	This land enters tapped unless you control two or more basic lands.
//
// The parenthetical mana ability is reminder text for the Mountain and Forest
// subtypes. It is modelled explicitly because council4 does not auto-derive
// subtype mana abilities at runtime.
var CinderGlade = func() *game.CardDef {
	card := &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green, color.Red),
		CardFace: game.CardFace{
			Name:     "Cinder Glade",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Mountain, types.Forest},
			OracleText: `
				({T}: Add {R} or {G}.)
				This land enters tapped unless you control two or more basic lands.
			`,
		},
	}
	card.ManaAbilities = append(card.ManaAbilities, game.TapManaChoiceAbility(mana.R, mana.G))
	card.ReplacementAbilities = append(card.ReplacementAbilities,
		game.EntersTappedIfReplacement("This land enters tapped unless you control two or more basic lands.", &game.Condition{
			Negate: true,
			ControllerControls: game.PermanentFilter{
				Types:      []types.Card{types.Land},
				Supertypes: []types.Super{types.Basic},
				MinCount:   2,
			},
		}),
	)
	return card
}()
