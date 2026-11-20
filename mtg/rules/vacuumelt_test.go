package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Source review: pinned Oracle revision 32dfe9aa-8303-460b-a4c3-e5533af2e1fd
// (Oracle ID eb421a1b-147f-422b-bb8d-483a182a4096), exact text:
// "Replicate {2}{U} (When you cast this spell, copy it for each time you paid
// its replicate cost. You may choose new targets for the copies.)"
// "Return target creature to its owner's hand." The pinned rulings URI,
// https://api.scryfall.com/cards/32dfe9aa-8303-460b-a4c3-e5533af2e1fd/rulings,
// was retrieved on 2026-11-20 and returned an empty list.
//
// Official Magic: The Gathering Comprehensive Rules, effective 2026-11-06:
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt.
// Read 601.2b/c/f-h, 603.2-3, 608.2b, 702.56a-b, and 707.10. These require
// announcing the additional Replicate payment and creature target while casting,
// paying the additional cost before the spell is cast, triggering one copy per
// payment, and checking the spell's target as it resolves. Replicate copies may
// retain or change the original target.
func TestVacuumeltReplicateReturnsTargetToOwnerHand(t *testing.T) {
	engine := NewEngine(nil)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1

	cardID := addCardToHand(g, game.Player1, authored.Vacuumelt.New())
	target := addCreaturePermanent(g, game.Player2)
	player := g.Players[game.Player1]
	player.ManaPool.Add(mana.C, 4)
	player.ManaPool.Add(mana.U, 2)

	cast := action.AnnounceSpell(action.CastSpellAction{
		CardID:          cardID,
		SourceZone:      zone.Hand,
		Targets:         []game.Target{game.PermanentTarget(target.ObjectID)},
		ReplicateCounts: []int{1},
	})
	if !engine.applyAction(g, game.Player1, cast) {
		t.Fatal("casting Vacuumelt with one paid Replicate cost was rejected")
	}
	original, ok := g.Stack.Peek()
	if !ok || original.ManaSpentToCast != 6 || len(original.ReplicateCounts) != 1 ||
		original.ReplicateCounts[0] != 1 || player.ManaPool.Total() != 0 {
		t.Fatalf("Vacuumelt did not pay {2}{U} plus Replicate {2}{U}: %+v", original)
	}
	if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 2 {
		t.Fatal("paid Replicate did not put its cast trigger above Vacuumelt")
	}

	log := &TurnLog{}
	engine.resolveTopOfStack(g, log)
	if g.Stack.Size() != 2 {
		t.Fatal("resolving Vacuumelt's Replicate trigger did not create its copy")
	}
	engine.resolveTopOfStack(g, log)
	if !g.Players[game.Player2].Hand.Contains(target.CardInstanceID) {
		t.Fatal("the replicated Vacuumelt did not return the target creature to its owner's hand")
	}
	engine.resolveTopOfStack(g, log)
	if !g.Stack.IsEmpty() || !player.Graveyard.Contains(cardID) {
		t.Fatal("Vacuumelt did not finish resolving and leave the stack")
	}
}
