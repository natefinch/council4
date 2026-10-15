package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BotanicalSanctum is the card definition for Botanical Sanctum.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped unless you control two or fewer other lands.
//	{T}: Add {G} or {U}.
var BotanicalSanctum = newBotanicalSanctum

func newBotanicalSanctum() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name:  "Botanical Sanctum",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.G, mana.U),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedIfReplacement("This land enters tapped unless you control two or fewer other lands.", &game.Condition{
					ControlsMatching: opt.Val(game.SelectionCount{
						Selection: game.Selection{RequiredTypes: []types.Card{types.Land}, ExcludeSource: true},
						MinCount:  3,
					}),
				}),
			},
			OracleText: `
			This land enters tapped unless you control two or fewer other lands.
			{T}: Add {G} or {U}.
		`,
		},
	}
}
