package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// startingTownLand builds the Starting Town enters-tapped replacement land: it
// enters tapped unless it is the controller's first, second, or third turn of
// the game. It mirrors the generated CardDef so the runtime test exercises the
// same ControllerTurnOfGameAtMost predicate the compiler emits.
func startingTownLand() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Starting Town",
		Types: []types.Card{types.Land},
		ReplacementAbilities: []game.ReplacementAbility{
			game.EntersTappedIfReplacement(
				"This land enters tapped unless it's your first, second, or third turn of the game.",
				&game.Condition{Negate: true, ControllerTurnOfGameAtMost: 3},
			),
		},
	}}
}

// TestConditionControllerTurnOfGameAtMost exercises the per-player turn-ordinal
// predicate directly: it holds only on the controller's own turn and only while
// the controller's turn count is within the bound.
func TestConditionControllerTurnOfGameAtMost(t *testing.T) {
	condition := opt.Val(game.Condition{ControllerTurnOfGameAtMost: 3})

	for _, test := range []struct {
		name       string
		turnsTaken int
		active     game.PlayerID
		want       bool
	}{
		{name: "first turn", turnsTaken: 1, active: game.Player1, want: true},
		{name: "second turn", turnsTaken: 2, active: game.Player1, want: true},
		{name: "third turn", turnsTaken: 3, active: game.Player1, want: true},
		{name: "fourth turn", turnsTaken: 4, active: game.Player1, want: false},
		{name: "before first turn", turnsTaken: 0, active: game.Player1, want: false},
		{name: "opponent's turn", turnsTaken: 1, active: game.Player2, want: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			g.Turn.ActivePlayer = test.active
			g.Players[game.Player1].TurnsTaken = test.turnsTaken
			got := conditionSatisfied(g, conditionContext{controller: game.Player1}, condition)
			if got != test.want {
				t.Fatalf("conditionSatisfied = %v, want %v", got, test.want)
			}
		})
	}
}

// TestConditionControllerTurnOfGameAtMostNegated exercises the "unless" form the
// enters-tapped replacement uses: the negated predicate holds (the land enters
// tapped) exactly when it is not the controller's first three turns of the game.
func TestConditionControllerTurnOfGameAtMostNegated(t *testing.T) {
	condition := opt.Val(game.Condition{Negate: true, ControllerTurnOfGameAtMost: 3})

	for _, test := range []struct {
		name       string
		turnsTaken int
		wantTapped bool
	}{
		{name: "third turn stays untapped", turnsTaken: 3, wantTapped: false},
		{name: "fourth turn enters tapped", turnsTaken: 4, wantTapped: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			g.Turn.ActivePlayer = game.Player1
			g.Players[game.Player1].TurnsTaken = test.turnsTaken
			got := conditionSatisfied(g, conditionContext{controller: game.Player1}, condition)
			if got != test.wantTapped {
				t.Fatalf("negated conditionSatisfied = %v, want %v", got, test.wantTapped)
			}
		})
	}
}

// TestStartingTownEntersTappedByTurnOfGame plays the Starting Town land and
// confirms it enters untapped on the controller's first three turns of the game
// and tapped from the fourth turn on, driven through the real enters-tapped
// replacement rather than the predicate in isolation.
func TestStartingTownEntersTappedByTurnOfGame(t *testing.T) {
	for _, test := range []struct {
		name       string
		turnsTaken int
		wantTapped bool
	}{
		{name: "first turn", turnsTaken: 1, wantTapped: false},
		{name: "second turn", turnsTaken: 2, wantTapped: false},
		{name: "third turn", turnsTaken: 3, wantTapped: false},
		{name: "fourth turn", turnsTaken: 4, wantTapped: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			setSorcerySpeedTurn(g, game.Player1)
			g.Players[game.Player1].TurnsTaken = test.turnsTaken
			cardID := addCardToHand(g, game.Player1, startingTownLand())
			if !NewEngine(nil).applyPlayLand(g, game.Player1, cardID) {
				t.Fatal("applyPlayLand() = false")
			}
			if got := g.Battlefield[len(g.Battlefield)-1].Tapped; got != test.wantTapped {
				t.Fatalf("Tapped = %v, want %v", got, test.wantTapped)
			}
		})
	}
}

// TestTurnsTakenCountsEachPlayersOwnTurns runs two full rounds of a four-player
// game and confirms each player's TurnsTaken counts their own turns, so a later
// seat's first turn is turn 1 for that player rather than the global turn
// number.
func TestTurnsTakenCountsEachPlayersOwnTurns(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	for i := range g.Players {
		stockLibrary(g, game.PlayerID(i), 40)
	}
	agents := allPassAgents()
	for turn := 0; turn < 2*game.NumPlayers && !g.IsGameOver(); turn++ {
		engine.runTurn(g, agents)
	}
	for i := range g.Players {
		if got := g.Players[i].TurnsTaken; got != 2 {
			t.Fatalf("player %d TurnsTaken = %d, want 2", i, got)
		}
	}
}

// TestTurnsTakenCountsExtraTurns confirms an extra turn a player takes counts
// toward that player's own turn tally, so "your third turn of the game" reaches
// its bound one turn sooner for a player who took an extra turn.
func TestTurnsTakenCountsExtraTurns(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	for i := range g.Players {
		stockLibrary(g, game.PlayerID(i), 40)
	}
	agents := allPassAgents()

	engine.runTurn(g, agents) // Player1's first turn; advances to Player2.
	if g.IsGameOver() {
		t.Fatal("game ended before extra turn setup")
	}
	// Queue an extra turn for Player1 to run after Player2's turn ends.
	g.Turn.ExtraTurns = append(g.Turn.ExtraTurns, game.Player1)
	engine.runTurn(g, agents) // Player2's first turn; advances to Player1's extra turn.
	engine.runTurn(g, agents) // Player1's extra (second) turn.

	if got := g.Players[game.Player1].TurnsTaken; got != 2 {
		t.Fatalf("Player1 TurnsTaken = %d, want 2 (turn plus extra turn)", got)
	}
	if got := g.Players[game.Player2].TurnsTaken; got != 1 {
		t.Fatalf("Player2 TurnsTaken = %d, want 1", got)
	}
}
