package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
)

// Serra Angel
//
// Type: Creature — Angel
// Cost: {3}{W}{W}
//
// Oracle text:
//   Flying
//   Vigilance (Attacking doesn't cause this creature to tap.)
//
// TODO: Fill in Abilities from oracle text.

var SerraAngel = &game.CardDef{
	Name: "Serra Angel",
	ManaCost: &mana.Cost{
			mana.GenericMana(3),
			mana.ColoredMana(mana.White),
			mana.ColoredMana(mana.White),
		},
	ManaValue: 5,
	Colors: []mana.Color{mana.White},
	ColorIdentity: mana.NewColorIdentity(mana.White),
	Types: []game.CardType{game.TypeCreature},
	Subtypes: []string{"Angel"},
	Power: &game.PT{Value: 4},
	Toughness: &game.PT{Value: 4},
	OracleText: "Flying\nVigilance (Attacking doesn't cause this creature to tap.)",
	// Abilities: filled in by LLM from oracle text.
	Abilities: []game.AbilityDef{},
}
