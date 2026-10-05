package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestLocalProductSourceIncarnationAndScheduledCapture(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	observed := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	obj := &game.StackObject{
		ID: g.IDGen.Next(), Kind: game.StackActivatedAbility, Controller: game.Player1,
		SourceID: source.ObjectID, SourceCardID: source.CardInstanceID,
	}
	persistent := linkedObjectSourceKey(g, obj, "local-link")
	rememberLinkedObject(g, persistent, game.LinkedObjectRef{ObjectID: source.ObjectID})
	restore := enterLocalProductFrame(g, obj, []game.Instruction{{
		LocalProducts: game.LocalProducts{Links: []game.LinkedKey{"local-link"}},
	}})
	local := linkedObjectSourceKey(g, obj, "local-link")
	rememberLinkedObject(g, local, game.LinkedObjectRef{ObjectID: observed.ObjectID, CardID: observed.CardInstanceID})
	if !scheduleDelayedTrigger(g, obj, delayedSacrificeCapture("local-link")) {
		t.Fatal("could not freeze actual local object at schedule time")
	}
	obj.SourceID = g.IDGen.Next()
	obj.SourceCardID = g.IDGen.Next()
	if linkedObjectSourceKey(g, obj, "local-link") != local || linkedObjectByObjectKey(g, obj, "local-link") != local {
		t.Fatal("active local publication followed a new source incarnation")
	}
	restore()
	if _, exists := g.LinkedObjects[local]; exists || len(g.LinkedObjects[persistent]) != 1 {
		t.Fatal("frame leaked its actual publisher cell or rewrote persistent source data")
	}
	newPersistent := linkedObjectSourceKey(g, obj, "local-link")
	if newPersistent == persistent || len(g.LinkedObjects[newPersistent]) != 0 {
		t.Fatal("frame restored old data into a new source incarnation")
	}
	NewEngine(nil).runEndingPhase(g, [game.NumPlayers]PlayerAgent{})
	if _, exists := permanentByObjectID(g, observed.ObjectID); exists {
		t.Fatal("delayed frozen capture was lost when the local frame exited")
	}
	if _, exists := permanentByObjectID(g, source.ObjectID); !exists {
		t.Fatal("delayed capture acquired the persistent source instead")
	}
}
