package rules

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: Rotting Rats HOP 39 printing
// 3895f0ce-ef97-4188-8b86-d0cc7067aacd (oracle 1f361ae9-98a2-40ff-a579-fa5b9f510df4),
// Oracle "When this creature enters, each player discards a card."; its
// https://api.scryfall.com/cards/3895f0ce-ef97-4188-8b86-d0cc7067aacd/rulings
// returned one ruling dated 2009-02-01: "... first the player whose turn it is
// chooses a card to discard, then each other player in turn order does the same,
// then all cards are discarded at the same time."
// Tinybones, Pocket Nuisance FRA 237 printing
// 2f47ddf7-35b6-4205-8045-f057914c5f64 (oracle 97dcf9fb-3f2f-4aa6-959a-e3c4889c1673),
// Oracle "When Tinybones enters, each opponent discards a card. / Whenever a
// player discards one or more cards, Tinybones deals 1 damage to each opponent.";
// its rulings endpoint returned no rulings.
// Marauding Mako DFT 138 printing 9efbfd67-e0f5-43e0-9fff-1eb4a2bed0d8
// (oracle e349be42-5f14-44a9-9608-281985c10e2d), Oracle "Whenever you discard
// one or more cards, put that many +1/+1 counters on this creature."; its
// rulings endpoint returned no rulings. All Oracle text matches the pinned
// corpus. Rest in Peace (oracle 087f9ad7-e74f-40e2-8102-1ed2925d0418): "If a
// card or token would be put into a graveyard from anywhere, exile it
// instead." Progenitus (oracle 2d7e00b6-12f0-4b03-82a6-50e3d1b5395e): "If
// Progenitus would be put into a graveyard from anywhere, reveal Progenitus and
// shuffle it into its owner's library instead." CR 101.4 (APNAP choices, then
// the actions happen simultaneously), CR 101.4d (APNAP order restarts for
// outstanding choices), CR 603.2c ("An ability triggers only once each time its
// trigger event occurs. However, it can trigger repeatedly if one event contains
// multiple occurrences."), CR 616.1 (the affected object's owner chooses among
// replacement effects; simultaneous choices are made in APNAP order) and
// CR 701.9a (discard) from the rules effective November 14, 2025.

// cohortDiscardEvents returns the discard events recorded since start.
func cohortDiscardEvents(g *game.Game, start int) []game.Event {
	var discards []game.Event
	for _, event := range g.Events[start:] {
		if event.Kind == game.EventCardDiscarded {
			discards = append(discards, event)
		}
	}
	return discards
}

// assertOneDiscardCohort checks that every discard and its hand-to-graveyard
// zone change since start share one nonzero simultaneous cohort, and that each
// player discarded exactly perPlayer[player] cards in it. It returns the cohort.
func assertOneDiscardCohort(t *testing.T, g *game.Game, start int, perPlayer map[game.PlayerID]int) id.ID {
	t.Helper()
	discards := cohortDiscardEvents(g, start)
	if len(discards) == 0 {
		t.Fatal("no discard events recorded")
	}
	cohort := discards[0].SimultaneousID
	if cohort == 0 {
		t.Fatal("group discard has no simultaneous cohort")
	}
	counts := make(map[game.PlayerID]int)
	for _, event := range discards {
		if event.SimultaneousID != cohort {
			t.Fatalf("player %d discard of card %d in batch %d, want shared cohort %d", event.Player, event.CardID, event.SimultaneousID, cohort)
		}
		if event.PlayerEventOrdinalThisTurn != 1 {
			t.Fatalf("player %d discard ordinal = %d, want 1 (each player's first discard occurrence)", event.Player, event.PlayerEventOrdinalThisTurn)
		}
		counts[event.Player]++
	}
	for _, event := range g.Events[start:] {
		if event.Kind == game.EventZoneChanged && event.FromZone == zone.Hand && event.SimultaneousID != cohort {
			t.Fatalf("zone change of card %d in batch %d, want shared cohort %d", event.CardID, event.SimultaneousID, cohort)
		}
	}
	for _, playerID := range []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4} {
		if counts[playerID] != perPlayer[playerID] {
			t.Fatalf("player %d discarded %d cards, want %d", playerID, counts[playerID], perPlayer[playerID])
		}
	}
	return cohort
}

// pendingTriggersFrom returns the stack objects whose source is sourceID.
func pendingTriggersFrom(g *game.Game, sourceID id.ID) []*game.StackObject {
	var objects []*game.StackObject
	for _, obj := range g.Stack.Objects() {
		if obj.SourceID == sourceID {
			objects = append(objects, obj)
		}
	}
	return objects
}

func resolveWholeStack(t *testing.T, engine *Engine, g *game.Game) {
	t.Helper()
	var agents [game.NumPlayers]PlayerAgent
	for range 64 {
		engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
		if g.Stack.IsEmpty() {
			return
		}
		engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	}
	t.Fatal("stack did not empty")
}

func fillHands(g *game.Game, count int, skip ...game.PlayerID) {
	for _, playerID := range []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4} {
		skipped := false
		for _, s := range skip {
			skipped = skipped || s == playerID
		}
		if skipped {
			continue
		}
		for i := range count {
			addCardToHand(g, playerID, discardableCard(fmt.Sprintf("P%d Card %d", playerID, i)))
		}
	}
}

// assertTinybonesTriggeredPerDiscarder checks Tinybones has one trigger on the
// stack per discarding player, each retaining that player's discard event.
func assertTinybonesTriggeredPerDiscarder(t *testing.T, g *game.Game, tinybones *game.Permanent, cohort id.ID, discarders []game.PlayerID) {
	t.Helper()
	triggers := pendingTriggersFrom(g, tinybones.ObjectID)
	if len(triggers) != len(discarders) {
		t.Fatalf("Tinybones triggers = %d, want %d (one per discarding player)", len(triggers), len(discarders))
	}
	seen := make(map[game.PlayerID]bool)
	for _, obj := range triggers {
		event := obj.TriggerEvent
		if event.Kind != game.EventCardDiscarded || event.SimultaneousID != cohort {
			t.Fatalf("Tinybones trigger event = %+v, want a discard in cohort %d", event, cohort)
		}
		if seen[event.Player] {
			t.Fatalf("Tinybones triggered twice for player %d's discard", event.Player)
		}
		seen[event.Player] = true
	}
	for _, playerID := range discarders {
		if !seen[playerID] {
			t.Fatalf("Tinybones did not trigger for player %d's discard", playerID)
		}
	}
}

// TestRottingRatsGroupDiscardIsOneCohortTinybonesTriggersPerPlayer proves the
// production Rotting Rats trigger discards every player's chosen card at the
// same time as one simultaneous cohort, while production Tinybones ("Whenever a
// player discards one or more cards") still triggers once per discarding player
// (CR 603.2c), dealing 1 damage to each of its controller's opponents per trigger.
func TestRottingRatsGroupDiscardIsOneCohortTinybonesTriggersPerPlayer(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.ActivePlayer = game.Player1
	tinybones := addCombatPermanent(g, game.Player1, registryCardDef(t, "Tinybones, Pocket Nuisance"))
	fillHands(g, 2)
	enterBattlefieldFromHand(t, engine, g, game.Player1, registryCardDef(t, "Rotting Rats"))
	var agents [game.NumPlayers]PlayerAgent
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) || g.Stack.Size() != 1 {
		t.Fatalf("Rotting Rats trigger count = %d, want 1", g.Stack.Size())
	}
	start := len(g.Events)
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	cohort := assertOneDiscardCohort(t, g, start, map[game.PlayerID]int{
		game.Player1: 1, game.Player2: 1, game.Player3: 1, game.Player4: 1,
	})
	engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
	assertTinybonesTriggeredPerDiscarder(t, g, tinybones, cohort,
		[]game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4})

	resolveWholeStack(t, engine, g)
	if got := g.Players[game.Player1].Life; got != 40 {
		t.Fatalf("Tinybones controller life = %d, want 40", got)
	}
	for _, playerID := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		if got := g.Players[playerID].Life; got != 36 {
			t.Fatalf("player %d life = %d, want 36 (four Tinybones triggers)", playerID, got)
		}
		if got := g.Players[playerID].Hand.Size(); got != 1 {
			t.Fatalf("player %d hand = %d, want 1", playerID, got)
		}
	}
}

// TestTinybonesOpponentGroupDiscardTriggersPerOpponent proves production
// Tinybones' own "each opponent discards a card" is one simultaneous cohort and
// its discard trigger fires once for each opponent who discarded.
func TestTinybonesOpponentGroupDiscardTriggersPerOpponent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.ActivePlayer = game.Player1
	fillHands(g, 1)
	tinybones := enterBattlefieldFromHand(t, engine, g, game.Player1, registryCardDef(t, "Tinybones, Pocket Nuisance"))
	var agents [game.NumPlayers]PlayerAgent
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) || g.Stack.Size() != 1 {
		t.Fatalf("Tinybones enters trigger count = %d, want 1", g.Stack.Size())
	}
	start := len(g.Events)
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	cohort := assertOneDiscardCohort(t, g, start, map[game.PlayerID]int{
		game.Player2: 1, game.Player3: 1, game.Player4: 1,
	})
	engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
	assertTinybonesTriggeredPerDiscarder(t, g, tinybones, cohort,
		[]game.PlayerID{game.Player2, game.Player3, game.Player4})

	resolveWholeStack(t, engine, g)
	for _, playerID := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		if got := g.Players[playerID].Life; got != 37 {
			t.Fatalf("player %d life = %d, want 37 (three Tinybones triggers)", playerID, got)
		}
	}
}

// TestGroupMultiCardDiscardCountsEachPlayersOccurrence proves a group discard of
// two cards per player discards all eight cards in one cohort, yet each player's
// two cards remain that player's own occurrence: Tinybones triggers once per
// player, and each player's Marauding Mako counts only its controller's two
// discarded cards ("that many"), not the whole cohort.
func TestGroupMultiCardDiscardCountsEachPlayersOccurrence(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.ActivePlayer = game.Player1
	tinybones := addCombatPermanent(g, game.Player1, registryCardDef(t, "Tinybones, Pocket Nuisance"))
	makoOne := addCombatPermanent(g, game.Player1, registryCardDef(t, "Marauding Mako"))
	makoTwo := addCombatPermanent(g, game.Player2, registryCardDef(t, "Marauding Mako"))
	fillHands(g, 3)
	addEffectSpellToStack(g, game.Player1, game.Discard{
		Amount:      game.Fixed(2),
		PlayerGroup: game.AllPlayersReference(),
	}, nil)
	start := len(g.Events)
	var agents [game.NumPlayers]PlayerAgent
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	cohort := assertOneDiscardCohort(t, g, start, map[game.PlayerID]int{
		game.Player1: 2, game.Player2: 2, game.Player3: 2, game.Player4: 2,
	})
	engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
	assertTinybonesTriggeredPerDiscarder(t, g, tinybones, cohort,
		[]game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4})
	for _, mako := range []*game.Permanent{makoOne, makoTwo} {
		triggers := pendingTriggersFrom(g, mako.ObjectID)
		if len(triggers) != 1 || triggers[0].TriggerEvent.Player != mako.Controller {
			t.Fatalf("player %d Mako triggers = %d, want one for its controller's discard", mako.Controller, len(triggers))
		}
	}

	resolveWholeStack(t, engine, g)
	for _, mako := range []*game.Permanent{makoOne, makoTwo} {
		if got := mako.Counters.Get(counter.PlusOnePlusOne); got != 2 {
			t.Fatalf("player %d Mako +1/+1 counters = %d, want 2 (its controller's discards only)", mako.Controller, got)
		}
	}
	for _, playerID := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		if got := g.Players[playerID].Life; got != 36 {
			t.Fatalf("player %d life = %d, want 36", playerID, got)
		}
	}
}

// TestGroupDiscardRoutesShareOneCohort proves the at-random and entire-hand
// player-group discards also discard simultaneously as one cohort, and that a
// player with no card to discard has no discard occurrence, so Tinybones does
// not trigger for them.
func TestGroupDiscardRoutesShareOneCohort(t *testing.T) {
	for _, tc := range []struct {
		name      string
		discard   game.Discard
		perPlayer int
	}{
		{
			name:      "at random",
			discard:   game.Discard{Amount: game.Fixed(1), AtRandom: true, PlayerGroup: game.AllPlayersReference()},
			perPlayer: 1,
		},
		{
			name:      "entire hand",
			discard:   game.Discard{EntireHand: true, PlayerGroup: game.AllPlayersReference()},
			perPlayer: 2,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			g.Turn.ActivePlayer = game.Player1
			tinybones := addCombatPermanent(g, game.Player1, registryCardDef(t, "Tinybones, Pocket Nuisance"))
			fillHands(g, 2, game.Player4)
			addEffectSpellToStack(g, game.Player1, tc.discard, nil)
			start := len(g.Events)
			var agents [game.NumPlayers]PlayerAgent
			engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

			cohort := assertOneDiscardCohort(t, g, start, map[game.PlayerID]int{
				game.Player1: tc.perPlayer, game.Player2: tc.perPlayer, game.Player3: tc.perPlayer,
			})
			engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
			assertTinybonesTriggeredPerDiscarder(t, g, tinybones, cohort,
				[]game.PlayerID{game.Player1, game.Player2, game.Player3})
		})
	}
}

// discardOccurrenceScopes are the two explicit one-or-more discard scopes: the
// per-player wording of "Whenever a player discards one or more cards" and a
// cohort wording that treats one simultaneous discard as a single occurrence
// ("Whenever one or more cards are discarded").
var discardOccurrenceScopes = []struct {
	name      string
	perPlayer bool
	want      int
}{
	{name: "per player", perPlayer: true, want: 4},
	{name: "cohort", perPlayer: false, want: 1},
}

// assertScopedDiscardTriggers checks the stack holds want triggers whose events
// all belong to cohort and, for a per-player scope, name distinct discarders.
func assertScopedDiscardTriggers(t *testing.T, g *game.Game, cohort id.ID, perPlayer bool, want int) {
	t.Helper()
	seen := make(map[game.PlayerID]bool)
	count := 0
	for _, obj := range g.Stack.Objects() {
		if obj.TriggerEvent.Kind != game.EventCardDiscarded || obj.TriggerEvent.SimultaneousID != cohort {
			t.Fatalf("trigger event = %+v, want a discard in cohort %d", obj.TriggerEvent, cohort)
		}
		if perPlayer && seen[obj.TriggerEvent.Player] {
			t.Fatalf("per-player trigger fired twice for player %d", obj.TriggerEvent.Player)
		}
		seen[obj.TriggerEvent.Player] = true
		count++
	}
	if count != want {
		t.Fatalf("discard triggers = %d, want %d", count, want)
	}
}

// TestOneOrMoreDiscardTriggerOccurrenceScope proves the explicit occurrence
// scope of an ordinary one-or-more discard trigger: with OneOrMorePerPlayer a
// simultaneous four-player discard of two cards each is four occurrences, while
// the cohort scope sees that one simultaneous discard as a single occurrence.
func TestOneOrMoreDiscardTriggerOccurrenceScope(t *testing.T) {
	for _, tc := range discardOccurrenceScopes {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			g.Turn.ActivePlayer = game.Player1
			fillHands(g, 2)
			addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
				Event: game.EventCardDiscarded, OneOrMore: true, OneOrMorePerPlayer: tc.perPlayer,
			}, []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
			addEffectSpellToStack(g, game.Player2, game.Discard{
				Amount:      game.Fixed(2),
				PlayerGroup: game.AllPlayersReference(),
			}, nil)
			start := len(g.Events)
			var agents [game.NumPlayers]PlayerAgent
			engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

			cohort := assertOneDiscardCohort(t, g, start, map[game.PlayerID]int{
				game.Player1: 2, game.Player2: 2, game.Player3: 2, game.Player4: 2,
			})
			engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
			assertScopedDiscardTriggers(t, g, cohort, tc.perPlayer, tc.want)
			resolveWholeStack(t, engine, g)
			if got := g.Players[game.Player1].Life; got != 40+tc.want {
				t.Fatalf("trigger controller life = %d, want %d", got, 40+tc.want)
			}
		})
	}
}

// TestDelayedOneOrMoreDiscardTriggerOccurrenceScope proves the event-based
// delayed trigger route applies the same explicit occurrence scope: a
// per-player "whenever a player discards one or more cards this turn" fires once
// for each player of a simultaneous group discard, and a cohort-scoped one fires
// once for the whole simultaneous discard.
func TestDelayedOneOrMoreDiscardTriggerOccurrenceScope(t *testing.T) {
	for _, tc := range discardOccurrenceScopes {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			g.Turn.ActivePlayer = game.Player1
			fillHands(g, 2)
			g.DelayedTriggers = append(g.DelayedTriggers, game.DelayedTrigger{
				ID:          g.IDGen.Next(),
				Controller:  game.Player1,
				CreatedTurn: g.Turn.TurnNumber,
				Ability: game.TriggeredAbility{
					Content: game.Mode{Sequence: []game.Instruction{{
						Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()},
					}}}.Ability(),
				},
				EventPattern: opt.Val(game.TriggerPattern{
					Event: game.EventCardDiscarded, OneOrMore: true, OneOrMorePerPlayer: tc.perPlayer,
				}),
				Window: game.DelayedWindowThisTurn,
			})
			addEffectSpellToStack(g, game.Player2, game.Discard{
				Amount:      game.Fixed(2),
				PlayerGroup: game.AllPlayersReference(),
			}, nil)
			start := len(g.Events)
			var agents [game.NumPlayers]PlayerAgent
			engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

			cohort := assertOneDiscardCohort(t, g, start, map[game.PlayerID]int{
				game.Player1: 2, game.Player2: 2, game.Player3: 2, game.Player4: 2,
			})
			engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
			assertScopedDiscardTriggers(t, g, cohort, tc.perPlayer, tc.want)
			resolveWholeStack(t, engine, g)
			if got := g.Players[game.Player1].Life; got != 40+tc.want {
				t.Fatalf("delayed trigger controller life = %d, want %d", got, 40+tc.want)
			}
		})
	}
}

// TestOneOrMorePerPlayerRequiresDiscardPattern proves card validation rejects a
// per-player occurrence scope outside a one-or-more discard pattern.
func TestOneOrMorePerPlayerRequiresDiscardPattern(t *testing.T) {
	for _, tc := range []struct {
		name    string
		pattern game.TriggerPattern
		valid   bool
	}{
		{name: "discard", pattern: game.TriggerPattern{Event: game.EventCardDiscarded, OneOrMore: true, OneOrMorePerPlayer: true}, valid: true},
		{name: "not one-or-more", pattern: game.TriggerPattern{Event: game.EventCardDiscarded, OneOrMorePerPlayer: true}},
		{name: "other event", pattern: game.TriggerPattern{Event: game.EventCardDrawn, OneOrMore: true, OneOrMorePerPlayer: true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pattern := tc.pattern
			card := triggeredCreature(&pattern, []game.Instruction{{
				Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()},
			}}, nil)
			issues := game.ValidateCardDef(card)
			if got := len(issues) == 0; got != tc.valid {
				t.Fatalf("validation issues = %v, want valid %v", issues, tc.valid)
			}
		})
	}
}

// replacementChoiceRecorder answers every replacement choice for its player by
// picking the option whose label contains pick, and checks before answering that every card
// of the simultaneous discard is still in its owner's hand and that no player
// can observe any of them in exile, a graveyard or a library.
type replacementChoiceRecorder struct {
	t       *testing.T
	g       *game.Game
	pending map[id.ID]game.PlayerID
	pick    string
	asked   *[]game.PlayerID
}

func (r replacementChoiceRecorder) agent(player game.PlayerID) PlayerAgent {
	return authoredChoicePolicy{choose: func(obs PlayerObservation, req game.ChoiceRequest) []int {
		if req.Kind != game.ChoiceReplacement {
			return req.DefaultSelection
		}
		*r.asked = append(*r.asked, player)
		for cardID, owner := range r.pending {
			if !r.g.Players[owner].Hand.Contains(cardID) {
				r.t.Errorf("player %d chose a replacement after card %d left player %d's hand", player, cardID, owner)
			}
			for _, other := range []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4} {
				for _, view := range append(obs.Exile(other), obs.Graveyard(other)...) {
					if view.CardInstanceID == cardID {
						r.t.Errorf("player %d sees discarded card %q before choosing its replacement", player, view.Name)
					}
				}
			}
		}
		for _, option := range req.Options {
			if strings.Contains(option.Label, r.pick) {
				return []int{option.Index}
			}
		}
		r.t.Fatalf("replacement options %+v do not name %q", req.Options, r.pick)
		return nil
	}}
}

// TestSimultaneousDiscardDecidesReplacementsBeforeAnyCardMoves proves every
// replacement decision of a simultaneous discard is made against the
// pre-discard state, in APNAP order, before any card leaves a hand or any
// discard publishes. It covers the default-registry Rotting Rats group discard
// (one card each) and the player-group entire-hand discard (several cards per
// player), with default-registry Rest in Peace and Progenitus competing for each
// Progenitus while player 1's ordinary card goes to exile without a choice.
func TestSimultaneousDiscardDecidesReplacementsBeforeAnyCardMoves(t *testing.T) {
	for _, tc := range []struct {
		name       string
		perPlayer  int
		pick       string
		wantZone   zone.Type
		rottingRat bool
	}{
		{name: "Rotting Rats keeps Progenitus", perPlayer: 1, pick: "Progenitus", wantZone: zone.Library, rottingRat: true},
		{name: "Rotting Rats chooses Rest in Peace", perPlayer: 1, pick: "exile it instead", wantZone: zone.Exile, rottingRat: true},
		{name: "entire hands keep Progenitus", perPlayer: 2, pick: "Progenitus", wantZone: zone.Library},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			g.Turn.ActivePlayer = game.Player1
			engine := NewEngine(nil)
			set := hiddenReplacementRegistry(t)
			resolveHiddenReplacementPermanent(t, g, engine, game.Player1, set.restInPeace)
			resolveWholeStack(t, engine, g)
			pending := make(map[id.ID]game.PlayerID)
			var ordinary []id.ID
			for range tc.perPlayer {
				cardID := addCardToHand(g, game.Player1, discardableCard("First secret"))
				pending[cardID] = game.Player1
				ordinary = append(ordinary, cardID)
				for _, player := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
					pending[addCardToHand(g, player, set.progenitus)] = player
				}
			}
			var asked []game.PlayerID
			recorder := replacementChoiceRecorder{t: t, g: g, pending: pending, pick: tc.pick, asked: &asked}
			var agents [game.NumPlayers]PlayerAgent
			for _, player := range []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4} {
				agents[player] = recorder.agent(player)
			}
			if tc.rottingRat {
				enterBattlefieldFromHand(t, engine, g, game.Player1, registryCardDef(t, "Rotting Rats"))
				if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) || g.Stack.Size() != 1 {
					t.Fatalf("Rotting Rats trigger count = %d, want 1", g.Stack.Size())
				}
			} else {
				addEffectSpellToStack(g, game.Player1, game.Discard{EntireHand: true, PlayerGroup: game.AllPlayersReference()}, nil)
			}
			start := len(g.Events)
			engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

			var wantAsked []game.PlayerID
			for _, player := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
				for range tc.perPlayer {
					wantAsked = append(wantAsked, player)
				}
			}
			if !slices.Equal(asked, wantAsked) {
				t.Fatalf("replacement choosers = %v, want %v (APNAP, owners only)", asked, wantAsked)
			}
			assertOneDiscardCohort(t, g, start, map[game.PlayerID]int{
				game.Player1: tc.perPlayer, game.Player2: tc.perPlayer, game.Player3: tc.perPlayer, game.Player4: tc.perPlayer,
			})
			for cardID, owner := range pending {
				want := tc.wantZone
				if slices.Contains(ordinary, cardID) {
					want = zone.Exile
				}
				if got := discardedCardZone(g, owner, cardID); got != want {
					t.Fatalf("player %d card %d in %v, want %v", owner, cardID, got, want)
				}
			}
		})
	}
}

// discardedCardZone returns which of owner's hand, graveyard, exile or library
// holds cardID.
func discardedCardZone(g *game.Game, owner game.PlayerID, cardID id.ID) zone.Type {
	player := g.Players[owner]
	switch {
	case player.Hand.Contains(cardID):
		return zone.Hand
	case player.Graveyard.Contains(cardID):
		return zone.Graveyard
	case player.Exile.Contains(cardID):
		return zone.Exile
	case player.Library.Contains(cardID):
		return zone.Library
	}
	return zone.Command
}
