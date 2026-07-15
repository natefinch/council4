package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// castSpells emits n spell-cast events for controller in the current turn
// window, mirroring the EventSpellCast the cast machinery records once per
// spell actually cast (CR 601).
func castSpells(g *game.Game, controller game.PlayerID, n int) {
	for range n {
		emitEvent(g, game.Event{Kind: game.EventSpellCast, Controller: controller})
	}
}

// TestAnyOpponentCastSpellsThisTurnIsPerOpponent proves the Mindbreak Trap
// condition is evaluated per opponent and never sums casts across opponents:
// one opponent must reach the threshold alone. It also confirms the boundary is
// "N or more" (>=), so exactly N satisfies it.
func TestAnyOpponentCastSpellsThisTurnIsPerOpponent(t *testing.T) {
	t.Run("single opponent below, at, and above threshold", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		castSpells(g, game.Player2, 2)
		if anyOpponentCastSpellsThisTurn(g, game.Player1, 3) {
			t.Fatal("threshold 3 satisfied when the only opponent cast 2 spells")
		}

		castSpells(g, game.Player2, 1) // now 3 total
		if !anyOpponentCastSpellsThisTurn(g, game.Player1, 3) {
			t.Fatal("threshold 3 not satisfied when an opponent cast exactly 3 spells")
		}

		castSpells(g, game.Player2, 1) // now 4 total
		if !anyOpponentCastSpellsThisTurn(g, game.Player1, 3) {
			t.Fatal("threshold 3 not satisfied when an opponent cast 4 spells")
		}
	})

	t.Run("casts are never summed across opponents", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		// Two opponents cast 2 and 1: five short of nobody, but neither alone
		// reaches 3, so the condition is false.
		castSpells(g, game.Player2, 2)
		castSpells(g, game.Player3, 1)
		if anyOpponentCastSpellsThisTurn(g, game.Player1, 3) {
			t.Fatal("threshold 3 satisfied by summing 2+1 across two opponents (must be per opponent)")
		}
		// A third opponent reaching 3 alone flips it true.
		castSpells(g, game.Player4, 3)
		if !anyOpponentCastSpellsThisTurn(g, game.Player1, 3) {
			t.Fatal("threshold 3 not satisfied when a single opponent reached 3")
		}
	})
}

// TestAnyOpponentCastSpellsThisTurnIgnoresOwnSpells proves the caster's own
// spells never count toward the opponent condition: only opponents are
// considered, so a player who cast many spells themselves does not satisfy their
// own Mindbreak Trap.
func TestAnyOpponentCastSpellsThisTurnIgnoresOwnSpells(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	castSpells(g, game.Player1, 5)
	if anyOpponentCastSpellsThisTurn(g, game.Player1, 1) {
		t.Fatal("condition satisfied by the caster's own spells")
	}
	// From an opponent's point of view, Player1's five casts are an opponent's
	// casts and do satisfy the threshold.
	if !anyOpponentCastSpellsThisTurn(g, game.Player2, 3) {
		t.Fatal("condition not satisfied for a player whose opponent cast five spells")
	}
}

// TestAnyOpponentCastSpellsThisTurnCountsOnlyCastEvents proves only spells that
// were actually cast count: spell copies, activated/triggered abilities, and
// played lands are excluded because none of them emit EventSpellCast.
func TestAnyOpponentCastSpellsThisTurnCountsOnlyCastEvents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	castSpells(g, game.Player2, 2)
	// Noise that must not count toward the spells-cast total.
	emitEvent(g, game.Event{Kind: game.EventSpellCopied, Controller: game.Player2})
	emitEvent(g, game.Event{Kind: game.EventSpellCopied, Controller: game.Player2})
	emitEvent(g, game.Event{Kind: game.EventAbilityActivated, Controller: game.Player2})
	emitEvent(g, game.Event{Kind: game.EventLandPlayed, Controller: game.Player2})

	if anyOpponentCastSpellsThisTurn(g, game.Player1, 3) {
		t.Fatal("copies, abilities, and lands were counted as cast spells")
	}
	castSpells(g, game.Player2, 1) // a genuine third cast tips it over
	if !anyOpponentCastSpellsThisTurn(g, game.Player1, 3) {
		t.Fatal("three genuine casts did not satisfy the threshold")
	}
}

// TestAnyOpponentCastSpellsThisTurnCountsCounteredSpells proves a spell that was
// countered still counts: its EventSpellCast was recorded when it was cast (CR
// 601.2i) and countering it emits no offsetting event, so the cast total is
// unaffected.
func TestAnyOpponentCastSpellsThisTurnCountsCounteredSpells(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	// Three spells were cast (recording three EventSpellCast) and subsequently
	// countered; countering does not retract those cast events.
	castSpells(g, game.Player2, 3)

	if !anyOpponentCastSpellsThisTurn(g, game.Player1, 3) {
		t.Fatal("countered spells stopped counting toward the spells-cast total")
	}
}

// TestAnyOpponentCastSpellsThisTurnResetsEachTurn proves only the current turn's
// casts count: three casts last turn do not satisfy the condition once a new
// turn has begun.
func TestAnyOpponentCastSpellsThisTurnResetsEachTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	castSpells(g, game.Player2, 3)
	if !anyOpponentCastSpellsThisTurn(g, game.Player1, 3) {
		t.Fatal("condition not satisfied on the turn the spells were cast")
	}

	// Advance to the next turn: the previous turn's casts fall out of the
	// current-turn window.
	g.EventTurnStarts = append(g.EventTurnStarts, len(g.Events))
	g.Turn.TurnNumber++

	if anyOpponentCastSpellsThisTurn(g, game.Player1, 3) {
		t.Fatal("last turn's casts still counted on the following turn")
	}
	// Fresh casts this turn satisfy it again.
	castSpells(g, game.Player2, 3)
	if !anyOpponentCastSpellsThisTurn(g, game.Player1, 3) {
		t.Fatal("this turn's casts did not satisfy the threshold after the reset")
	}
}

// TestAnyOpponentCastSpellsThisTurnSkipsEliminatedOpponents proves an eliminated
// player is not treated as an opponent: their casts never satisfy the condition,
// while a player still in the game whose opponent reached the threshold does.
func TestAnyOpponentCastSpellsThisTurnSkipsEliminatedOpponents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Players[game.Player2].Eliminated = true
	castSpells(g, game.Player2, 3)
	if anyOpponentCastSpellsThisTurn(g, game.Player1, 3) {
		t.Fatal("an eliminated player's casts satisfied the condition")
	}

	// A living opponent reaching the threshold satisfies it.
	castSpells(g, game.Player3, 3)
	if !anyOpponentCastSpellsThisTurn(g, game.Player1, 3) {
		t.Fatal("a living opponent's three casts did not satisfy the condition")
	}
}

// TestAnyOpponentCastSpellsThisTurnNonPositiveCount documents the defensive
// lower bound: a threshold below one is always satisfied. Card validation
// forbids a non-positive ConditionCount, so this only guards a malformed
// definition.
func TestAnyOpponentCastSpellsThisTurnNonPositiveCount(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	if !anyOpponentCastSpellsThisTurn(g, game.Player1, 0) {
		t.Fatal("threshold 0 was not trivially satisfied")
	}
}
