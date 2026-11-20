package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Source review: pinned Oracle revision a9c1c674-0742-45be-afd5-2dbcd6b499ae
// (Oracle ID 90ba0366-57ff-4dc4-a290-3925548129ce), verified against
// https://api.scryfall.com/cards/a9c1c674-0742-45be-afd5-2dbcd6b499ae on
// 2026-11-19: "Replicate {1}{U} (When you cast this spell, copy it for each
// time you paid its replicate cost.) Draw a card." Its rulings URL
// https://api.scryfall.com/cards/a9c1c674-0742-45be-afd5-2dbcd6b499ae/rulings
// returned an empty list.
// Official Magic: The Gathering Comprehensive Rules, effective 2026-11-06:
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt.
// Read 601.2b/f-h, 601.2i, 603.3, 702.56a, and 707.10. These require the
// additional Replicate payments to be made while casting, then create a cast
// trigger that puts one spell copy on the stack per payment when it resolves;
// each copy resolves the spell's draw effect but is not itself cast.
func TestAuthoredTrainOfThoughtReplicateDrawsForEachPayment(t *testing.T) {
	catalog, err := cardprogram.NewCatalog(authored.TrainOfThought.Programs)
	if err != nil {
		t.Fatal(err)
	}
	engine := NewEngineWithPrograms(nil, catalog)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1

	card := addCardToHand(g, game.Player1, authored.TrainOfThought.New())
	drawn := make([]id.ID, 3)
	for i := range drawn {
		drawn[i] = addCardToLibrary(g, game.Player1, &game.CardDef{
			CardFace: game.CardFace{Name: "Drawn card"},
		})
	}
	player := g.Players[game.Player1]
	player.ManaPool.Add(mana.C, 3)
	player.ManaPool.Add(mana.U, 3)

	cast := action.AnnounceSpell(action.CastSpellAction{
		CardID:          card,
		SourceZone:      zone.Hand,
		ReplicateCounts: []int{2},
	})
	if !engine.applyAction(g, game.Player1, cast) {
		t.Fatal("casting Train of Thought with two Replicate payments was rejected")
	}
	original, _ := g.Stack.Peek()
	if original.ManaSpentToCast != 6 || player.ManaPool.Total() != 0 ||
		len(original.ReplicateCounts) != 1 || original.ReplicateCounts[0] != 2 {
		t.Fatal("Train of Thought did not pay {1}{U} plus two {1}{U} Replicate costs")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 2 {
		t.Fatal("two Replicate payments did not produce the card's cast trigger")
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Stack.Size() != 3 {
		t.Fatal("Train of Thought did not create one spell copy per Replicate payment")
	}
	for range 3 {
		engine.resolveTopOfStack(g, &TurnLog{})
	}
	if !g.Stack.IsEmpty() || player.Hand.Size() != len(drawn) ||
		player.Library.Size() != 0 || !player.Graveyard.Contains(card) {
		t.Fatal("the original and both copies did not resolve Draw a card.")
	}
	for _, id := range drawn {
		if !player.Hand.Contains(id) {
			t.Fatalf("drawn library card %v did not reach the hand", id)
		}
	}
}
