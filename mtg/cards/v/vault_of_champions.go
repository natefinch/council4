package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// VaultOfChampions is the card definition for Vault of Champions.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped unless you have two or more opponents.
//	{T}: Add {W} or {B}.
var VaultOfChampions = newVaultOfChampions

func newVaultOfChampions() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name:  "Vault of Champions",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.W, mana.B),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedIfReplacement("This land enters tapped unless you have two or more opponents.", &game.Condition{
					Negate:     true,
					Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateOpponentCount, Op: compare.GreaterOrEqual, Value: 2}},
				}),
			},
			OracleText: `
			This land enters tapped unless you have two or more opponents.
			{T}: Add {W} or {B}.
		`,
		},
	}
}
