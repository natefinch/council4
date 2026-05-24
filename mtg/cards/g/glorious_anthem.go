package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
)

// Glorious Anthem
//
// Type: Enchantment
// Cost: {1}{W}{W}
//
// Oracle text:
//   Creatures you control get +1/+1.
//
// TODO: Fill in Abilities from oracle text.

var GloriousAnthem = &game.CardDef{
	Name: "Glorious Anthem",
	ManaCost: &mana.Cost{
			mana.GenericMana(1),
			mana.ColoredMana(mana.White),
			mana.ColoredMana(mana.White),
		},
	ManaValue: 3,
	Colors: []mana.Color{mana.White},
	ColorIdentity: mana.NewColorIdentity(mana.White),
	Types: []game.CardType{game.TypeEnchantment},
	OracleText: "Creatures you control get +1/+1.",
	// Abilities: filled in by LLM from oracle text.
	Abilities: []game.AbilityDef{},
}
