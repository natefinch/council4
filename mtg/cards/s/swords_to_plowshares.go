package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/opt"
)

// Swords to Plowshares
//
// Type: Instant
// Cost: {W}
//
// Oracle text:
//   Exile target creature. Its controller gains life equal to its power.
//
// TODO: Fill in Abilities from oracle text.

var SwordsToPlowshares = &game.CardDef{
	Name: "Swords to Plowshares",
	ManaCost: opt.Val(mana.Cost{
		mana.ColoredMana(mana.White),
	}),
	ManaValue:     1,
	Colors:        []mana.Color{mana.White},
	ColorIdentity: mana.NewColorIdentity(mana.White),
	Types:         []game.CardType{game.TypeInstant},
	OracleText:    "Exile target creature. Its controller gains life equal to its power.",
	// Abilities: filled in by LLM from oracle text.
	Abilities: []game.AbilityDef{},
}
