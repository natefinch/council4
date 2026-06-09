package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// DeathcapGlade is the card definition for Deathcap Glade.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped unless you control two or more other lands.
//	{T}: Add {B} or {G}.
var DeathcapGlade = func() *game.CardDef {
	card := &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name:  "Deathcap Glade",
			Types: []types.Card{types.Land},
			OracleText: `
				This land enters tapped unless you control two or more other lands.
				{T}: Add {B} or {G}.
			`,
		},
	}
	card.ReplacementAbilities = append(card.ReplacementAbilities,
		game.EntersTappedIfReplacement("This land enters tapped unless you control two or more other lands.", &game.Condition{
			Negate: true,
			ControllerControls: game.PermanentFilter{
				Types:    []types.Card{types.Land},
				MinCount: 2,
			},
		}),
	)
	card.ManaAbilities = append(card.ManaAbilities, game.TapManaChoiceAbility(mana.B, mana.G))
	return card
}()
