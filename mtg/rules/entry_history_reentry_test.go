package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: Éowyn, Shieldmaiden (LTC 1 printing
// dd4a00ff-2206-4e12-a0ab-61ed82c9e6c5, Oracle
// 98a2a642-a30a-43e7-b47f-4776c22d1cb0) "if another Human entered the
// battlefield under your control this turn"; its only Scryfall ruling
// (2023-06-16) requires that Human to have entered before beginning of combat.
// CR effective 2026-09-25 (verified official TXT retained under .cardwork):
// 400.7 an object that moves between zones becomes a new object with no
// relation to its previous existence; 201.5 text naming the object it is on
// means that particular object; 603.6a/603.10a entry triggers look at the
// entering permanent while leaves-the-battlefield triggers look back to the
// object that left. So a card's earlier battlefield incarnation is "another"
// permanent to the object it later became, while its own current entry is not.

func reenterEntryHistoryCard(t *testing.T, g *game.Game, permanent *game.Permanent) *game.Permanent {
	t.Helper()
	if !movePermanentToZone(g, permanent, zone.Graveyard) {
		t.Fatalf("permanent %d did not leave the battlefield", permanent.ObjectID)
	}
	card, ok := g.GetCardInstance(permanent.CardInstanceID)
	if !ok || !g.Players[card.Owner].Graveyard.Contains(card.ID) {
		t.Fatalf("card %d is not in its owner's graveyard", permanent.CardInstanceID)
	}
	g.Players[card.Owner].Graveyard.Remove(card.ID)
	returned, ok := createCardPermanent(g, card, card.Owner, zone.Graveyard)
	if !ok {
		t.Fatalf("card %d did not return to the battlefield", card.ID)
	}
	if returned.ObjectID == permanent.ObjectID || returned.CardInstanceID != permanent.CardInstanceID {
		t.Fatalf("returned object %d/card %d, want a new object for card %d",
			returned.ObjectID, returned.CardInstanceID, permanent.CardInstanceID)
	}
	return returned
}

func entryHistoryReentryGame() *game.Game {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	g.Turn.Phase = game.PhasePrecombatMain
	return g
}

func TestEntryHistoryEarlierIncarnationIsAnotherPermanent(t *testing.T) {
	t.Parallel()
	t.Run("own entry alone does not count", func(t *testing.T) {
		t.Parallel()
		g := entryHistoryReentryGame()
		eowyn := enterEntryHistoryCard(t, g, game.Player1, entryHistoryRegistryCard(t, "Éowyn, Shieldmaiden"))
		if eowynTriggersAtCombat(t, g, eowyn) {
			t.Fatal("Éowyn's own entry satisfied its another-Human history")
		}
	})
	t.Run("earlier incarnation counts after reentry", func(t *testing.T) {
		t.Parallel()
		g := entryHistoryReentryGame()
		first := enterEntryHistoryCard(t, g, game.Player1, entryHistoryRegistryCard(t, "Éowyn, Shieldmaiden"))
		eowyn := reenterEntryHistoryCard(t, g, first)
		if !eowynTriggersAtCombat(t, g, eowyn) {
			t.Fatal("Éowyn's earlier battlefield incarnation was not another Human that entered this turn")
		}
	})
	t.Run("earlier incarnation under another controller does not count", func(t *testing.T) {
		t.Parallel()
		g := entryHistoryReentryGame()
		def := entryHistoryRegistryCard(t, "Éowyn, Shieldmaiden")
		cardID := addCardToHand(g, game.Player1, def)
		card, _ := g.GetCardInstance(cardID)
		g.Players[game.Player1].Hand.Remove(cardID)
		first, ok := createCardPermanent(g, card, game.Player2, zone.Hand)
		if !ok {
			t.Fatal("Éowyn did not enter under Player2's control")
		}
		eowyn := reenterEntryHistoryCard(t, g, first)
		if eowyn.Controller != game.Player1 {
			t.Fatalf("returned Éowyn controller = %d, want owner Player1", eowyn.Controller)
		}
		if eowynTriggersAtCombat(t, g, eowyn) {
			t.Fatal("earlier incarnation that entered under Player2's control counted for Player1's Éowyn")
		}
	})
	t.Run("same-named different card counts", func(t *testing.T) {
		t.Parallel()
		g := entryHistoryReentryGame()
		eowyn := enterEntryHistoryCard(t, g, game.Player1, entryHistoryRegistryCard(t, "Éowyn, Shieldmaiden"))
		enterEntryHistoryCard(t, g, game.Player1, entryHistoryRegistryCard(t, "Éowyn, Shieldmaiden"))
		if !eowynTriggersAtCombat(t, g, eowyn) {
			t.Fatal("a different Éowyn card was not another Human that entered this turn")
		}
	})
	t.Run("different Human counts", func(t *testing.T) {
		t.Parallel()
		g := entryHistoryReentryGame()
		first := enterEntryHistoryCard(t, g, game.Player1, entryHistoryRegistryCard(t, "Éowyn, Shieldmaiden"))
		eowyn := reenterEntryHistoryCard(t, g, first)
		human := enterEntryHistoryCard(t, g, game.Player1, entryHistoryCreature("Entry History Squire", types.Human))
		if !movePermanentToZone(g, human, zone.Graveyard) || !eowynTriggersAtCombat(t, g, eowyn) {
			t.Fatal("a different Human that entered and left did not trigger the reentered Éowyn")
		}
	})
}

func entryHistorySelfEntries(g *game.Game, source *game.Permanent, minCount int) bool {
	return conditionEventHistorySatisfied(g, conditionContext{controller: game.Player1, source: source}, &game.EventHistoryCondition{
		Pattern:  game.TriggerPattern{Event: game.EventPermanentEnteredBattlefield, Source: game.TriggerSourceSelf},
		Window:   game.EventHistoryCurrentTurn,
		MinCount: minCount,
	})
}

func TestEntryHistorySelfEntryNamesOnlyCurrentObject(t *testing.T) {
	t.Parallel()
	g := entryHistoryReentryGame()
	first := enterEntryHistoryCard(t, g, game.Player1, entryHistoryCreature("Entry History Squire", types.Human))
	if !entryHistorySelfEntries(g, first, 1) {
		t.Fatal("permanent did not see its own entry this turn")
	}
	returned := reenterEntryHistoryCard(t, g, first)
	if !entryHistorySelfEntries(g, returned, 1) {
		t.Fatal("reentered permanent did not see its own current entry this turn")
	}
	if entryHistorySelfEntries(g, returned, 2) {
		t.Fatal("reentered permanent counted its earlier incarnation's entry as its own")
	}
	if !entryHistorySelfEntries(g, first, 1) || entryHistorySelfEntries(g, first, 2) {
		t.Fatal("earlier incarnation's last-known object did not see exactly its own entry")
	}
}

func entryHistoryReentryWatcher() *game.CardDef {
	pt := game.PT{Value: 1}
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Entry History Reentry Watcher",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Human},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
		TriggeredAbilities: []game.TriggeredAbility{
			{
				Text: "When this creature enters, draw a card.",
				Trigger: game.TriggerCondition{Type: game.TriggerWhen, Pattern: game.TriggerPattern{
					Event: game.EventPermanentEnteredBattlefield, Source: game.TriggerSourceSelf,
				}},
			},
			{
				Text: "Whenever another creature enters, draw a card.",
				Trigger: game.TriggerCondition{Type: game.TriggerWhenever, Pattern: game.TriggerPattern{
					Event: game.EventPermanentEnteredBattlefield, ExcludeSelf: true,
				}},
			},
			{
				Text: "When this creature dies, draw a card.",
				Trigger: game.TriggerCondition{Type: game.TriggerWhen, Pattern: game.TriggerPattern{
					Event: game.EventPermanentDied, Source: game.TriggerSourceSelf,
				}},
			},
		},
	}}
}

func entryHistoryTriggeredIndexes(event game.Event, sourceID id.ID) []int {
	var indexes []int
	for _, triggered := range event.TriggeredAbilities {
		if triggered.SourceID == sourceID {
			indexes = append(indexes, triggered.AbilityIndex)
		}
	}
	return indexes
}

func lastEntryHistoryEvent(t *testing.T, g *game.Game, kind game.EventKind) game.Event {
	t.Helper()
	for i := len(g.Events) - 1; i >= 0; i-- {
		if g.Events[i].Kind == kind {
			if !g.Events[i].TriggeredAbilitiesCaptured {
				t.Fatalf("event %+v did not capture triggered abilities", g.Events[i])
			}
			return g.Events[i]
		}
	}
	t.Fatalf("no %v event", kind)
	return game.Event{}
}

func TestEntryHistoryReentryKeepsLiveSelfTriggers(t *testing.T) {
	t.Parallel()
	g := entryHistoryReentryGame()
	first := enterEntryHistoryCard(t, g, game.Player1, entryHistoryReentryWatcher())
	returned := reenterEntryHistoryCard(t, g, first)
	entry := lastEntryHistoryEvent(t, g, game.EventPermanentEnteredBattlefield)
	if entry.PermanentID != returned.ObjectID {
		t.Fatalf("last entry event names %d, want returned object %d", entry.PermanentID, returned.ObjectID)
	}
	if got := entryHistoryTriggeredIndexes(entry, returned.ObjectID); len(got) != 1 || got[0] != 0 {
		t.Fatalf("returned watcher entry triggered abilities %v, want only its self-entry trigger 0", got)
	}

	other := enterEntryHistoryCard(t, g, game.Player1, entryHistoryCreature("Entry History Squire", types.Human))
	entry = lastEntryHistoryEvent(t, g, game.EventPermanentEnteredBattlefield)
	if entry.PermanentID != other.ObjectID {
		t.Fatalf("last entry event names %d, want other creature %d", entry.PermanentID, other.ObjectID)
	}
	if got := entryHistoryTriggeredIndexes(entry, returned.ObjectID); len(got) != 1 || got[0] != 1 {
		t.Fatalf("another creature's entry triggered watcher abilities %v, want only another-enters trigger 1", got)
	}

	if !movePermanentToZone(g, returned, zone.Graveyard) {
		t.Fatal("returned watcher did not die")
	}
	died := lastEntryHistoryEvent(t, g, game.EventPermanentDied)
	if got := entryHistoryTriggeredIndexes(died, returned.ObjectID); len(got) != 1 || got[0] != 2 {
		t.Fatalf("returned watcher death triggered abilities %v, want only its dies trigger 2", got)
	}
}
