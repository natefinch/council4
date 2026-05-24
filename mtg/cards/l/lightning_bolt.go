package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
)

// Lightning Bolt
//
// Type: Instant
// Cost: {R}
//
// Oracle text:
//   Lightning Bolt deals 3 damage to any target.
//
// TODO: Fill in Abilities from oracle text.

var LightningBolt = &game.CardDef{
	Name: "Lightning Bolt",
	ManaCost: &mana.Cost{
			mana.ColoredMana(mana.Red),
		},
	ManaValue: 1,
	Colors: []mana.Color{mana.Red},
	ColorIdentity: mana.NewColorIdentity(mana.Red),
	Types: []game.CardType{game.TypeInstant},
	OracleText: "Lightning Bolt deals 3 damage to any target.",
	// Abilities: filled in by LLM from oracle text.
	Abilities: []game.AbilityDef{},
}
