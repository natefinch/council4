package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LightningBolt is the card definition for Lightning Bolt.
//
// Type: Instant
// Cost: {R}
//
// Oracle text:
//
//	Lightning Bolt deals 3 damage to any target.
var LightningBolt = &game.CardDef{
	Name: "Lightning Bolt",
	ManaCost: opt.Val(mana.Cost{
		mana.ColoredMana(mana.Red),
	}),
	ManaValue:     1,
	Colors:        []mana.Color{mana.Red},
	ColorIdentity: mana.NewColorIdentity(mana.Red),
	Types:         []types.Card{types.Instant},
	OracleText:    "Lightning Bolt deals 3 damage to any target.",
	Abilities: []game.AbilityDef{
		{
			Kind: game.SpellAbility,
			Text: "Lightning Bolt deals 3 damage to any target.",
			Targets: []game.TargetSpec{
				{MinTargets: 1, MaxTargets: 1, Constraint: "any target"},
			},
			Effects: []game.Effect{
				{Type: game.EffectDamage, Amount: 3, TargetIndex: 0},
			},
		},
	},
}
