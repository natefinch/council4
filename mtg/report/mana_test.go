package report

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/rules"
	"github.com/natefinch/council4/mtg/sim"
	"github.com/natefinch/council4/opt"
)

const (
	manaTestLandID  id.ID = 10
	manaTestSpellID id.ID = 20
	manaTestBombID  id.ID = 30
)

func manaTestCards() map[id.ID]rules.CardInfo {
	return map[id.ID]rules.CardInfo{
		manaTestLandID:  {Name: "Forest", Owner: game.Player1, ManaValue: 0, Types: []types.Card{types.Land}},
		manaTestSpellID: {Name: "Spell", Owner: game.Player1, ManaValue: 2, Types: []types.Card{types.Sorcery}},
		manaTestBombID:  {Name: "Bomb", Owner: game.Player1, ManaValue: 6, Types: []types.Card{types.Sorcery}},
	}
}

// landTurn is one of Player1's turns in which it plays `played` lands.
func landTurn(played int) rules.TurnLog {
	turn := rules.TurnLog{ActivePlayer: game.Player1}
	for range played {
		turn.Actions = append(turn.Actions, rules.ActionLog{Player: game.Player1, Action: action.PlayLand(manaTestLandID)})
	}
	return turn
}

func floodedGame() rules.GameResult {
	g := rules.GameResult{HasWinner: true, Winner: game.Player1, Cards: manaTestCards()}
	for range 7 {
		g.Turns = append(g.Turns, landTurn(1))
	}
	g.Events = []game.Event{
		{Kind: game.EventSpellCast, CardID: manaTestSpellID, ManaValue: opt.Val(2)},
		{Kind: game.EventCardDrawn, CardID: manaTestBombID},
	}
	return g
}

func screwedGame() rules.GameResult {
	g := rules.GameResult{HasWinner: true, Winner: game.Player2, Cards: manaTestCards()}
	g.Turns = append(g.Turns, landTurn(1))
	for range 4 {
		g.Turns = append(g.Turns, landTurn(0))
	}
	g.Events = []game.Event{
		{Kind: game.EventCardDrawn, CardID: manaTestBombID},
	}
	return g
}

func makeSimulation(games []rules.GameResult, failures []sim.GameFailure) sim.SimulationResult {
	seeds := make([]uint64, len(games))
	for i := range games {
		seeds[i] = uint64(i + 1)
	}
	return sim.SimulationResult{
		Games:      games,
		Seeds:      seeds,
		GameCount:  len(games),
		MasterSeed: 1,
		Failures:   failures,
	}
}

func TestComputeManaCurveFloodScrewRot(t *testing.T) {
	result := makeSimulation([]rules.GameResult{floodedGame(), screwedGame()}, nil)
	cards := computeCardMetrics(result, game.Player1)
	metrics := computeManaCurve(result, game.Player1, cards)

	// 8 lands over 12 tested turns.
	if !approxEqual(metrics.LandsPerTurn, 8.0/12.0) {
		t.Errorf("LandsPerTurn = %v, want %v", metrics.LandsPerTurn, 8.0/12.0)
	}
	if !approxEqual(metrics.FloodRate, 0.5) {
		t.Errorf("FloodRate = %v, want 0.5", metrics.FloodRate)
	}
	if !approxEqual(metrics.ScrewRate, 0.5) {
		t.Errorf("ScrewRate = %v, want 0.5", metrics.ScrewRate)
	}
	// 4 no-land-drop turns out of 12.
	if !approxEqual(metrics.NoLandDropRate, 4.0/12.0) {
		t.Errorf("NoLandDropRate = %v, want %v", metrics.NoLandDropRate, 4.0/12.0)
	}
	// One 2-mana spell cast across two games.
	if !approxEqual(metrics.ManaSpentPerGame, 1.0) || !approxEqual(metrics.SpellsPerGame, 0.5) {
		t.Errorf("ManaSpent/Spells per game = %v/%v, want 1.0/0.5", metrics.ManaSpentPerGame, metrics.SpellsPerGame)
	}
	// Bomb (mv 6) drawn twice, never cast: rot = 6*2 / 2 games = 6.
	if len(metrics.RottedCards) != 1 || metrics.RottedCards[0].Name != "Bomb" {
		t.Fatalf("RottedCards = %+v, want a single Bomb", metrics.RottedCards)
	}
	if metrics.RottedCards[0].Draws != 2 || metrics.RottedCards[0].ManaValue != 6 {
		t.Errorf("rotted Bomb = %+v, want draws 2 mv 6", metrics.RottedCards[0])
	}
	if !approxEqual(metrics.RotMVPerGame, 6.0) {
		t.Errorf("RotMVPerGame = %v, want 6.0", metrics.RotMVPerGame)
	}
}

func TestManaCurveExcludesLandsFromRot(t *testing.T) {
	// A land drawn but never cast must not be reported as rot.
	game0 := rules.GameResult{
		HasWinner: true, Winner: game.Player1, Cards: manaTestCards(),
		Events: []game.Event{{Kind: game.EventCardDrawn, CardID: manaTestLandID}},
		Turns:  []rules.TurnLog{landTurn(0)},
	}
	result := makeSimulation([]rules.GameResult{game0}, nil)
	cards := computeCardMetrics(result, game.Player1)
	metrics := computeManaCurve(result, game.Player1, cards)
	for _, card := range metrics.RottedCards {
		if card.Name == "Forest" {
			t.Error("a land was reported as expensive rot")
		}
	}
}
