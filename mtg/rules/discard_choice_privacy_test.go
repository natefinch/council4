package rules

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
)

// Source review: Raffine's Informant SNC 26 printing
// 4e64ff87-2099-4360-94f6-164277b7b514 (oracle 838c4831-7d5c-4025-92b3-46eee59d6b34),
// Oracle "When this creature enters, it connives. (Draw a card, then discard a
// card. If you discarded a nonland card, put a +1/+1 counter on this creature.)"
// from the pinned corpus; https://api.scryfall.com/cards/4e64ff87-2099-4360-94f6-164277b7b514/rulings
// returned three rulings dated 2022-04-29 (no actions mid-connive; no counter
// when nothing is discarded; a departed creature still connives). CR 400.2
// (hand is a hidden zone), CR 514.1 (cleanup discard to maximum hand size),
// CR 701.9a/701.9b (discard; the affected player chooses) and CR 701.50a
// (connive draws, then discards) from the rules effective September 25, 2026.

// pickingDiscardAgent picks the named cards from a discard request and records
// the card names it was offered.
type pickingDiscardAgent struct {
	picks   []string
	offered *[][]string
}

func (*pickingDiscardAgent) ChooseAction(_ PlayerObservation, _ []action.Action) action.Action {
	return actionBuild.pass()
}

func (a *pickingDiscardAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	var names []string
	var selected []int
	for _, option := range request.Options {
		if !option.Card.Exists {
			continue
		}
		names = append(names, option.Card.Val.Name)
		if slices.Contains(a.picks, option.Card.Val.Name) {
			selected = append(selected, option.Index)
		}
	}
	*a.offered = append(*a.offered, names)
	return selected
}

// assertPrivateDiscardLog checks that every logged choice of kind is redacted
// and that the marshalled game log names none of the hidden cards.
func assertPrivateDiscardLog(t *testing.T, turnLog *TurnLog, kind game.ChoiceKind, hidden []string) {
	t.Helper()
	logged := 0
	for _, decision := range turnLog.Choices {
		if decision.Request.Kind != kind {
			continue
		}
		logged++
		request := decision.Request
		if !request.Private || request.Prompt != "Private decision" {
			t.Fatalf("logged discard request is not private: %+v", request)
		}
		for _, option := range request.Options {
			if option.Card.Exists || option.Label != "Private option" {
				t.Fatalf("logged discard option retains hand information: %+v", option)
			}
		}
	}
	if logged != 1 {
		t.Fatalf("logged discard choices = %d, want 1", logged)
	}
	data, err := json.Marshal(GameResult{Turns: []TurnLog{*turnLog}})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range hidden {
		if strings.Contains(string(data), name) {
			t.Fatalf("returned game log reveals undiscarded hand card %q", name)
		}
	}
}

// TestConniveDiscardChoiceKeepsHandOutOfLog proves the production Raffine's
// Informant connive lets its controller choose the discard from their whole
// hand (CR 701.50a, CR 701.9b), while the returned game log redacts the request
// so neither the kept cards nor the drawn card are revealed (CR 400.2).
func TestConniveDiscardChoiceKeepsHandOutOfLog(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.ActivePlayer = game.Player1
	for _, name := range []string{"Connive Keep", "Connive Toss", "Connive Hold"} {
		addCardToHand(g, game.Player1, discardableCard(name))
	}
	addCardToLibrary(g, game.Player1, discardableCard("Connive Drawn"))
	var offered [][]string
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = &pickingDiscardAgent{picks: []string{"Connive Toss"}, offered: &offered}
	informant := enterBattlefieldFromHand(t, engine, g, game.Player1, registryCardDef(t, "Raffine's Informant"))
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, nil) || g.Stack.Size() != 1 {
		t.Fatalf("Raffine's Informant trigger count = %d, want 1", g.Stack.Size())
	}
	turnLog := &TurnLog{}

	engine.resolveTopOfStackWithChoices(g, agents, turnLog)

	if len(offered) != 1 {
		t.Fatalf("connive discard requests = %d, want 1", len(offered))
	}
	got := slices.Clone(offered[0])
	slices.Sort(got)
	want := []string{"Connive Drawn", "Connive Hold", "Connive Keep", "Connive Toss"}
	if !slices.Equal(got, want) {
		t.Fatalf("connive discard options = %v, want the whole hand %v", got, want)
	}
	if !graveyardHasCardNamed(g, game.Player1, "Connive Toss") {
		t.Fatal("connive did not discard the chosen card")
	}
	hand := handCardNames(g, game.Player1)
	slices.Sort(hand)
	if wantHand := []string{"Connive Drawn", "Connive Hold", "Connive Keep"}; !slices.Equal(hand, wantHand) {
		t.Fatalf("hand after connive = %v, want %v", hand, wantHand)
	}
	if got := informant.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("+1/+1 counters = %d, want 1 after discarding a nonland card", got)
	}
	assertPrivateDiscardLog(t, turnLog, game.ChoiceResolution, []string{"Connive Keep", "Connive Hold", "Connive Drawn"})
}

// TestCleanupHandSizeDiscardChoiceKeepsHandOutOfLog proves the cleanup
// turn-based discard to maximum hand size (CR 514.1) lets the active player
// choose which cards to discard, while the returned game log redacts the
// request so the cards kept in hand are not revealed (CR 400.2).
func TestCleanupHandSizeDiscardChoiceKeepsHandOutOfLog(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.ActivePlayer = game.Player1
	var names []string
	for i := range maximumHandSize + 2 {
		name := fmt.Sprintf("Cleanup Card %c", 'A'+i)
		names = append(names, name)
		addCardToHand(g, game.Player1, discardableCard(name))
	}
	picks := []string{names[1], names[4]}
	var offered [][]string
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = &pickingDiscardAgent{picks: picks, offered: &offered}
	turnLog := &TurnLog{}

	engine.runCleanupStep(g, agents, turnLog)

	if len(offered) != 1 {
		t.Fatalf("cleanup discard requests = %d, want 1", len(offered))
	}
	if got := slices.Sorted(slices.Values(offered[0])); !slices.Equal(got, names) {
		t.Fatalf("cleanup discard options = %v, want the whole hand %v", got, names)
	}
	for _, name := range picks {
		if !graveyardHasCardNamed(g, game.Player1, name) {
			t.Fatalf("cleanup did not discard chosen card %q", name)
		}
	}
	var kept []string
	for _, name := range names {
		if !slices.Contains(picks, name) {
			kept = append(kept, name)
		}
	}
	if hand := slices.Sorted(slices.Values(handCardNames(g, game.Player1))); !slices.Equal(hand, kept) {
		t.Fatalf("hand after cleanup = %v, want %v", hand, kept)
	}
	assertPrivateDiscardLog(t, turnLog, game.ChoiceZoneSelection, kept)
}
