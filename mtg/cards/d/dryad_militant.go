package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DryadMilitant is the card definition for Dryad Militant.
//
// Type: Creature — Dryad Soldier
// Cost: {G/W}
//
// Oracle text:
//
//	({G/W} can be paid with either {G} or {W}.)
//	If an instant or sorcery card would be put into a graveyard from anywhere, exile it instead.
var DryadMilitant = newDryadMilitant

func newDryadMilitant() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Dryad Militant",
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.G, mana.W),
			}),
			Colors:    []color.Color{color.Green, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dryad, types.Soldier},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ReplacementAbilities: []game.ReplacementAbility{
				game.GraveyardRedirectReplacement("If an instant or sorcery card would be put into a graveyard from anywhere, exile it instead.", game.TriggerControllerAny, game.TriggerControllerAny, false, types.Instant, types.Sorcery),
			},
			OracleText: `
			({G/W} can be paid with either {G} or {W}.)
			If an instant or sorcery card would be put into a graveyard from anywhere, exile it instead.
		`,
		},
	}
}
