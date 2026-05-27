package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/opt"
)

// Soul Warden
//
// Type: Creature — Human Cleric
// Cost: {W}
//
// Oracle text:
//   Whenever another creature enters, you gain 1 life.
//
// TODO: Fill in Abilities from oracle text.

var SoulWarden = &game.CardDef{
	Name: "Soul Warden",
	ManaCost: opt.Val(mana.Cost{
		mana.ColoredMana(mana.White),
	}),
	ManaValue:     1,
	Colors:        []mana.Color{mana.White},
	ColorIdentity: mana.NewColorIdentity(mana.White),
	Types:         []game.CardType{game.TypeCreature},
	Subtypes:      []string{"Human", "Cleric"},
	Power:         opt.Val(game.PT{Value: 1}),
	Toughness:     opt.Val(game.PT{Value: 1}),
	OracleText:    "Whenever another creature enters, you gain 1 life.",
	// Abilities: filled in by LLM from oracle text.
	Abilities: []game.AbilityDef{},
}
