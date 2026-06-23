package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestStackObjectSourceNameNamesAbilitySource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID:  cardID,
		Def: &game.CardDef{CardFace: game.CardFace{Name: "Llanowar Elves"}},
	}
	// An activated ability's SourceID is a permanent object ID, but SourceCardID
	// references the underlying card instance.
	obj := &game.StackObject{Kind: game.StackActivatedAbility, SourceID: 9999, SourceCardID: cardID}
	if name := stackObjectSourceName(g, obj); name != "Llanowar Elves" {
		t.Fatalf("stackObjectSourceName = %q, want Llanowar Elves", name)
	}
}

func TestStackObjectSourceNameNamesToken(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	obj := &game.StackObject{
		Kind:           game.StackTriggeredAbility,
		SourceTokenDef: &game.CardDef{CardFace: game.CardFace{Name: "Goblin"}},
	}
	if name := stackObjectSourceName(g, obj); name != "Goblin" {
		t.Fatalf("stackObjectSourceName = %q, want Goblin", name)
	}
}
