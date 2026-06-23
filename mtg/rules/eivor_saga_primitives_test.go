package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestShuffleGraveyardIntoLibraryMovesEveryCard resolves a
// ShuffleGraveyardIntoLibrary effect and asserts every graveyard card moves into
// the controller's library and the graveyard is emptied (The Mending of
// Dominaria chapter III).
func TestShuffleGraveyardIntoLibraryMovesEveryCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	first := addCfzGraveyardCard(g, game.Player1, chooseFromZoneCardDef("Bear", 2, []types.Card{types.Creature}, nil))
	second := addCfzGraveyardCard(g, game.Player1, chooseFromZoneCardDef("Bolt", 1, []types.Card{types.Instant}, nil))
	libraryBefore := g.Players[game.Player1].Library.Size()

	addEffectSpellToStack(g, game.Player1, game.ShuffleGraveyardIntoLibrary{
		Player: game.ControllerReference(),
	}, nil)
	NewEngine(nil).resolveTopOfStack(g, &TurnLog{})

	if g.Players[game.Player1].Graveyard.Contains(first) || g.Players[game.Player1].Graveyard.Contains(second) {
		t.Fatal("graveyard cards were not removed from the graveyard")
	}
	if got := g.Players[game.Player1].Library.Size(); got != libraryBefore+2 {
		t.Fatalf("library size = %d, want %d", got, libraryBefore+2)
	}
	if !g.Players[game.Player1].Library.Contains(first) || !g.Players[game.Player1].Library.Contains(second) {
		t.Fatal("graveyard cards were not moved into the library")
	}
}

// TestAddCounterChooseOnePlacesOnChosenMember resolves an AddCounter with
// ChooseOne set over a battlefield group and asserts the counter lands only on
// the single permanent the controller chooses, not on every group member (Ajani
// Fells the Godsire chapter II).
func TestAddCounterChooseOnePlacesOnChosenMember(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	first := addCombatPermanent(g, game.Player1, greenCreatureDef("First"))
	second := addCombatPermanent(g, game.Player1, greenCreatureDef("Second"))

	agent := &sequencedChoiceAgent{choices: [][]int{{1}}}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}
	addEffectSpellToStack(g, game.Player1, game.AddCounter{
		Amount:      game.Fixed(1),
		CounterKind: counter.Vigilance,
		ChooseOne:   true,
		Group: game.BattlefieldGroup(game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			Controller:    game.ControllerYou,
		}),
	}, nil)
	NewEngine(nil).resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := first.Counters.Get(counter.Vigilance); got != 0 {
		t.Fatalf("unchosen creature got %d vigilance counters, want 0", got)
	}
	if got := second.Counters.Get(counter.Vigilance); got != 1 {
		t.Fatalf("chosen creature got %d vigilance counters, want 1", got)
	}
}
