package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: Fathom Mage (NCC 339 printing
// 22be6c55-f95e-498e-b6d2-65b7ed57d2d3, Oracle
// 93d0e129-e3b5-4aff-9e50-f34771ed00ff, from the pinned corpus), 1/1 "Evolve
// (Whenever a creature you control enters, if that creature has greater power
// or toughness than this creature, put a +1/+1 counter on this creature.)".
// Its rulings_uri returned nine rulings; the 2013-04-15 rulings say to check the
// entering creature's power and toughness against the evolve creature's when it
// enters, and evolve does not trigger at all if neither is greater. Grand Ball
// Guest (WOE 134 printing d6e75228-16af-42b0-8441-ed253a660cc9, Oracle
// 0b0fdb28-f7b1-4c30-9f2c-14a9074da22d), 2/2 Celebration "+1/+1 and has
// trample as long as two or more nonland permanents entered the battlefield
// under your control this turn"; its five 2023-09-01 rulings say Celebration
// checks past events. Official Comprehensive Rules TXT effective 2026-09-25
// (https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt),
// sections read: 603.4 (an intervening "if" is checked when the trigger event
// occurs and again on resolution), 603.6a (each event putting permanents onto
// the battlefield checks all permanents, newcomers included), 603.10 (objects
// and continuous effects immediately after the event determine triggers and
// what event objects look like), 611.3a (static effects apply continuously),
// 613.1 (layers) and 702.100a (evolve's intervening-if comparison). Immediately
// after an entry, that entry and every simultaneous entry are completed this-turn
// history, so live trigger capture must evaluate Celebration with them exactly
// once, as the entering permanent's snapshot does.

const entryCaptureCreaturesYouControlEnter = "Whenever a creature you control enters, if two or more nonland permanents entered the battlefield under your control this turn, draw a card."

func entryCaptureFathomMage(t *testing.T, g *game.Game) *game.Permanent {
	t.Helper()
	mage := addCombatPermanent(g, game.Player1, entryHistoryRegistryCard(t, "Fathom Mage"))
	mage.Counters.Add(counter.PlusOnePlusOne, 1)
	if values := effectivePermanentValues(g, mage); values.power != 2 || values.toughness != 2 {
		t.Fatalf("Fathom Mage with one +1/+1 counter is %d/%d, want 2/2", values.power, values.toughness)
	}
	return mage
}

func entryCaptureTriggeredBy(t *testing.T, g *game.Game, source, entered *game.Permanent) bool {
	t.Helper()
	for _, trigger := range entryHistoryEnteredEvent(t, g, entered).TriggeredAbilities {
		if trigger.SourceID == source.ObjectID {
			return true
		}
	}
	return false
}

func entryCaptureAssertNoPendingOverlay(t *testing.T, g *game.Game) {
	t.Helper()
	if len(g.PendingEntryEvents()) != 0 {
		t.Fatalf("pending entry overlay leaked after entry: %d events", len(g.PendingEntryEvents()))
	}
}

// Fathom Mage's evolve compares against a Grand Ball Guest–equivalent creature
// with the continuous effects that exist immediately after it entered: as the
// second nonland permanent it is already a 3/3 (CR 603.10, 702.100a).
func TestEntryTriggerCaptureEvolveSeesCompletedEntryHistory(t *testing.T) {
	t.Parallel()
	t.Run("second nonland entry triggers evolve", func(t *testing.T) {
		t.Parallel()
		g, _ := entryHistoryCelebrationGame(t)
		mage := entryCaptureFathomMage(t, g)
		enterEntryHistoryCard(t, g, game.Player1, entryHistoryArtifact())
		guest := enterEntryHistoryCard(t, g, game.Player1, entryHistoryCelebrationGuest("Entry Capture Guest"))
		entryCaptureAssertNoPendingOverlay(t, g)
		if got := entryHistoryEnteredPower(t, g, guest); got != 3 {
			t.Fatalf("Guest entering as the second nonland permanent had entry power %d, want 3", got)
		}
		if !entryCaptureTriggeredBy(t, g, mage, guest) {
			t.Fatal("Fathom Mage's evolve did not trigger for a 3/3 Guest entering beside the 2/2 Mage")
		}
		engine := NewEngine(nil)
		if !engine.putTriggeredAbilitiesOnStack(g) {
			t.Fatal("evolve trigger was not put on the stack")
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		if got := mage.Counters.Get(counter.PlusOnePlusOne); got != 2 {
			t.Fatalf("Fathom Mage has %d +1/+1 counters after evolve resolved, want 2", got)
		}
	})
	t.Run("first nonland entry does not trigger evolve", func(t *testing.T) {
		t.Parallel()
		g, _ := entryHistoryCelebrationGame(t)
		mage := entryCaptureFathomMage(t, g)
		enterEntryHistoryCard(t, g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name: "Entry Capture Wastes", Types: []types.Card{types.Land},
		}})
		guest := enterEntryHistoryCard(t, g, game.Player1, entryHistoryCelebrationGuest("Entry Capture Guest"))
		entryCaptureAssertNoPendingOverlay(t, g)
		if got := entryHistoryEnteredPower(t, g, guest); got != 2 {
			t.Fatalf("Guest entering after only a land had entry power %d, want 2", got)
		}
		if entryCaptureTriggeredBy(t, g, mage, guest) {
			t.Fatal("Fathom Mage's evolve triggered for a 2/2 Guest entering beside the 2/2 Mage")
		}
	})
	t.Run("simultaneous Guest tokens each trigger evolve", func(t *testing.T) {
		t.Parallel()
		g, _ := entryHistoryCelebrationGame(t)
		mage := entryCaptureFathomMage(t, g)
		guests, ok := createTokenPermanentsCollectingWithChoices(NewEngine(nil), g, game.Player1,
			entryHistoryCelebrationGuest("Entry Capture Guest Token"), 2, false, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
		if !ok || len(guests) != 2 {
			t.Fatalf("created %d Guest tokens, want 2", len(guests))
		}
		entryCaptureAssertNoPendingOverlay(t, g)
		for i, guest := range guests {
			if got := entryHistoryEnteredPower(t, g, guest); got != 3 {
				t.Errorf("simultaneous Guest token %d had entry power %d, want 3", i, got)
			}
			if !entryCaptureTriggeredBy(t, g, mage, guest) {
				t.Errorf("Fathom Mage's evolve did not trigger for simultaneous 3/3 Guest token %d", i)
			}
		}
	})
	t.Run("simultaneous card entries each trigger evolve", func(t *testing.T) {
		t.Parallel()
		g, _ := entryHistoryCelebrationGame(t)
		mage := entryCaptureFathomMage(t, g)
		e := NewEngine(nil)
		var entries []preparedCardPermanentEntry
		for _, def := range []*game.CardDef{entryHistoryCelebrationGuest("Entry Capture Guest"), entryHistoryArtifact()} {
			cardID := addCardToHand(g, game.Player1, def)
			card := g.CardInstances[cardID]
			g.Players[game.Player1].Hand.Remove(cardID)
			entry, ok := prepareCardPermanentFaceForSimultaneousEntry(e, g, card, game.Player1, zone.Hand, game.FaceFront, nil,
				permanentCreationOptions{SimultaneousID: g.IDGen.Next()}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			if !ok {
				t.Fatalf("%s was not prepared for simultaneous entry", def.Name)
			}
			entries = append(entries, entry)
		}
		commitSimultaneousCardPermanentEntries(g, entries)
		entryCaptureAssertNoPendingOverlay(t, g)
		guest := entries[0].permanent
		if got := entryHistoryEnteredPower(t, g, guest); got != 3 {
			t.Fatalf("Guest entering simultaneously with an artifact had entry power %d, want 3", got)
		}
		if !entryCaptureTriggeredBy(t, g, mage, guest) {
			t.Fatal("Fathom Mage's evolve did not trigger for a 3/3 Guest entering simultaneously with an artifact")
		}
	})
}

func entryCaptureCelebrationWatcher(g *game.Game) *game.Permanent {
	return addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Entry Capture Watcher",
		Types: []types.Card{types.Enchantment},
		TriggeredAbilities: []game.TriggeredAbility{{
			Text: entryCaptureCreaturesYouControlEnter,
			Trigger: game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event:            game.EventPermanentEnteredBattlefield,
					Controller:       game.TriggerControllerYou,
					SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
				},
				InterveningIf: "two or more nonland permanents entered the battlefield under your control this turn",
				InterveningCondition: opt.Val(game.Condition{EventHistory: opt.Val(game.EventHistoryCondition{
					Pattern: game.TriggerPattern{
						Event:            game.EventPermanentEnteredBattlefield,
						Controller:       game.TriggerControllerYou,
						SubjectSelection: game.Selection{ExcludedTypes: []types.Card{types.Land}},
					},
					Window:   game.EventHistoryCurrentTurn,
					MinCount: 2,
				})}),
			},
		}},
	}})
}

// An intervening-if history condition checked as an entry triggers counts the
// triggering entry and its simultaneous entries exactly once (CR 603.4, 603.6a,
// 603.10).
func TestEntryTriggerCaptureInterveningIfCountsCompletedEntriesOnce(t *testing.T) {
	t.Parallel()
	t.Run("only entry this turn", func(t *testing.T) {
		t.Parallel()
		g, _ := entryHistoryCelebrationGame(t)
		watcher := entryCaptureCelebrationWatcher(g)
		creature := enterEntryHistoryCard(t, g, game.Player1, entryHistoryCreature("Entry Capture Squire"))
		entryCaptureAssertNoPendingOverlay(t, g)
		if entryCaptureTriggeredBy(t, g, watcher, creature) {
			t.Fatal("the only nonland entry this turn satisfied a two-or-more intervening-if")
		}
	})
	t.Run("second nonland entry", func(t *testing.T) {
		t.Parallel()
		g, _ := entryHistoryCelebrationGame(t)
		watcher := entryCaptureCelebrationWatcher(g)
		enterEntryHistoryCard(t, g, game.Player1, entryHistoryArtifact())
		creature := enterEntryHistoryCard(t, g, game.Player1, entryHistoryCreature("Entry Capture Squire"))
		entryCaptureAssertNoPendingOverlay(t, g)
		if !entryCaptureTriggeredBy(t, g, watcher, creature) {
			t.Fatal("the triggering entry was not counted as the second nonland entry this turn")
		}
	})
	t.Run("simultaneous entries", func(t *testing.T) {
		t.Parallel()
		g, _ := entryHistoryCelebrationGame(t)
		watcher := entryCaptureCelebrationWatcher(g)
		creatures, ok := createTokenPermanentsCollectingWithChoices(NewEngine(nil), g, game.Player1,
			entryHistoryCreature("Entry Capture Squire Token"), 2, false, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
		if !ok || len(creatures) != 2 {
			t.Fatalf("created %d creature tokens, want 2", len(creatures))
		}
		entryCaptureAssertNoPendingOverlay(t, g)
		for i, creature := range creatures {
			if !entryCaptureTriggeredBy(t, g, watcher, creature) {
				t.Errorf("simultaneous token %d did not see both simultaneous entries", i)
			}
		}
	})
}
