package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestPutPermanentOnLibraryTopMovesSourceToTop verifies the PutPermanentOnLibrary
// primitive moves the source permanent to the top of its owner's library without
// shuffling, modeling Sensei's Divining Top's "put this artifact on top of its
// owner's library".
func TestPutPermanentOnLibraryTopMovesSourceToTop(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	existing := g.IDGen.Next()
	g.Players[game.Player1].Library.Add(existing)

	permanent := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Source Artifact",
		Types: []types.Card{types.Artifact},
	}})
	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackActivatedAbility,
		SourceID:   permanent.ObjectID,
		Controller: game.Player1,
	}
	log := TurnLog{}
	instr := game.Instruction{Primitive: game.PutPermanentOnLibrary{Object: game.SourcePermanentReference()}}
	engine.resolveInstructionWithChoices(g, obj, &instr, [game.NumPlayers]PlayerAgent{}, &log)

	if _, ok := permanentByObjectID(g, permanent.ObjectID); ok {
		t.Fatal("permanent remained on battlefield")
	}
	top, ok := g.Players[game.Player1].Library.Top()
	if !ok || top != permanent.CardInstanceID {
		t.Fatalf("library top = %v (ok=%v), want source card %v", top, ok, permanent.CardInstanceID)
	}
}

// TestPutPermanentOnLibraryBottomMovesSourceToBottom verifies the Bottom flag
// places the source permanent on the bottom of its owner's library.
func TestPutPermanentOnLibraryBottomMovesSourceToBottom(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	existing := g.IDGen.Next()
	g.Players[game.Player1].Library.Add(existing)

	permanent := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Source Artifact",
		Types: []types.Card{types.Artifact},
	}})
	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackActivatedAbility,
		SourceID:   permanent.ObjectID,
		Controller: game.Player1,
	}
	log := TurnLog{}
	instr := game.Instruction{Primitive: game.PutPermanentOnLibrary{Object: game.SourcePermanentReference(), Bottom: true}}
	engine.resolveInstructionWithChoices(g, obj, &instr, [game.NumPlayers]PlayerAgent{}, &log)

	if _, ok := permanentByObjectID(g, permanent.ObjectID); ok {
		t.Fatal("permanent remained on battlefield")
	}
	bottom, ok := g.Players[game.Player1].Library.Bottom()
	if !ok || bottom != permanent.CardInstanceID {
		t.Fatalf("library bottom = %v (ok=%v), want source card %v", bottom, ok, permanent.CardInstanceID)
	}
	top, ok := g.Players[game.Player1].Library.Top()
	if !ok || top != existing {
		t.Fatalf("library top = %v (ok=%v), want pre-existing card %v", top, ok, existing)
	}
}
