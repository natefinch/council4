package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
)

// Sol Ring
//
// Type: Artifact
// Cost: {1}
//
// Oracle text:
//   {T}: Add {C}{C}.
//
// TODO: Fill in Abilities from oracle text.

var SolRing = &game.CardDef{
	Name: "Sol Ring",
	ManaCost: &mana.Cost{
			mana.GenericMana(1),
		},
	ManaValue: 1,
	Types: []game.CardType{game.TypeArtifact},
	OracleText: "{T}: Add {C}{C}.",
	// Abilities: filled in by LLM from oracle text.
	Abilities: []game.AbilityDef{},
}
