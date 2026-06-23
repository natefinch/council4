package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestPutLinkedExiledCardsInLibraryMovesToBottomAndClearsLink exercises the
// disposal that backs Trial of a Time Lord's guilty verdict: after an
// exile-until-leaves clause links a creature to its source, the disposal moves
// each linked exiled card to the bottom of its owner's library and clears the
// link so the synthesized return finds nothing.
func TestPutLinkedExiledCardsInLibraryMovesToBottomAndClearsLink(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	victim := addCombatCreaturePermanent(g, game.Player2)
	source := addCombatPermanent(g, game.Player1, oringCardDef())

	obj := linkedSourceObject(source)
	obj.Targets = []game.Target{game.PermanentTarget(victim.ObjectID)}
	resolveInstruction(engine, g, obj, game.Exile{
		Object:         game.TargetPermanentReference(0),
		ExileLinkedKey: game.LinkedKey("exile-until-leaves"),
	}, nil)
	if !g.Players[game.Player2].Exile.Contains(victim.CardInstanceID) {
		t.Fatal("victim did not reach its owner's exile zone")
	}

	resolveInstruction(engine, g, obj, game.PutLinkedExiledCardsInLibrary{
		LinkedKey: game.LinkedKey("exile-until-leaves"),
		Bottom:    true,
	}, nil)

	if g.Players[game.Player2].Exile.Contains(victim.CardInstanceID) {
		t.Fatal("victim remained in exile after the linked disposal")
	}
	if bottom, ok := g.Players[game.Player2].Library.Bottom(); !ok || bottom != victim.CardInstanceID {
		t.Fatalf("library bottom = %v (ok=%v), want victim %v on the bottom", bottom, ok, victim.CardInstanceID)
	}
	if remaining := linkedObjects(g, linkedObjectSourceKey(g, obj, "exile-until-leaves")); len(remaining) != 0 {
		t.Fatalf("linked objects = %+v, want the link cleared after disposal", remaining)
	}
}

// TestPutLinkedExiledCardsInLibraryLeavesNothingToReturn verifies the disposal
// clears the link so a later leaves-the-battlefield return trigger brings the
// disposed card back from neither the library nor anywhere else.
func TestPutLinkedExiledCardsInLibraryLeavesNothingToReturn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	victim := addCombatCreaturePermanent(g, game.Player2)
	source := addCombatPermanent(g, game.Player1, oringCardDef())

	obj := linkedSourceObject(source)
	obj.Targets = []game.Target{game.PermanentTarget(victim.ObjectID)}
	resolveInstruction(engine, g, obj, game.Exile{
		Object:         game.TargetPermanentReference(0),
		ExileLinkedKey: game.LinkedKey("exile-until-leaves"),
	}, nil)
	resolveInstruction(engine, g, obj, game.PutLinkedExiledCardsInLibrary{
		LinkedKey: game.LinkedKey("exile-until-leaves"),
		Bottom:    true,
	}, nil)

	movePermanentToZone(g, source, zone.Graveyard)
	if engine.putTriggeredAbilitiesOnStack(g) {
		engine.resolveTopOfStack(g, &TurnLog{})
	}

	if permanentByCardID(g, victim.CardInstanceID) != nil {
		t.Fatal("disposed card returned to the battlefield after the source left")
	}
	if bottom, ok := g.Players[game.Player2].Library.Bottom(); !ok || bottom != victim.CardInstanceID {
		t.Fatalf("library bottom = %v (ok=%v), want disposed card to stay on the bottom", bottom, ok)
	}
}
