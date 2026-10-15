package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VaultSkirge is the card definition for Vault Skirge.
//
// Type: Artifact Creature — Phyrexian Imp
// Cost: {1}{B/P}
//
// Oracle text:
//
//	({B/P} can be paid with either {B} or 2 life.)
//	Flying
//	Lifelink (Damage dealt by this creature also causes you to gain that much life.)
var VaultSkirge = newVaultSkirge

func newVaultSkirge() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Vault Skirge",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.PhyrexianMana(mana.B),
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Imp},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.LifelinkStaticBody,
			},
			OracleText: `
			({B/P} can be paid with either {B} or 2 life.)
			Flying
			Lifelink (Damage dealt by this creature also causes you to gain that much life.)
		`,
		},
	}
}
