package report

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/rules"
)

const (
	tempoCardID      id.ID = 50
	tempoPermanentID id.ID = 100
)

func tempoCards() map[id.ID]rules.CardInfo {
	return map[id.ID]rules.CardInfo{
		tempoCardID: {Name: "Threat", Owner: game.Player1},
	}
}

func tempoWinGame() rules.GameResult {
	g := rules.GameResult{HasWinner: true, Winner: game.Player1, Cards: tempoCards()}
	g.Turns = []rules.TurnLog{
		{TurnNumber: 1, ActivePlayer: game.Player1, Actions: []rules.ActionLog{
			{Player: game.Player1, Action: action.PlayLand(tempoCardID)},
		}},
		{TurnNumber: 2, ActivePlayer: game.Player1,
			Actions:      []rules.ActionLog{{Player: game.Player1, Action: action.CastSpell(tempoCardID, nil, 0, nil)}},
			CombatDamage: []rules.CombatDamageLog{{Controller: game.Player1, DefendingPlayer: game.Player2, Damage: 5}}},
		{TurnNumber: 3, ActivePlayer: game.Player1,
			CombatDamage: []rules.CombatDamageLog{{Controller: game.Player1, DefendingPlayer: game.Player2, Damage: 3}}},
	}
	g.Events = []game.Event{
		{Kind: game.EventPermanentEnteredBattlefield, PermanentID: tempoPermanentID, CardID: tempoCardID},
		{Kind: game.EventObjectBecameTarget, Controller: game.Player2, Target: game.PermanentTarget(tempoPermanentID)},
		{Kind: game.EventObjectBecameTarget, Controller: game.Player2, Target: game.PlayerTarget(game.Player1)},
		{Kind: game.EventObjectBecameTarget, Controller: game.Player1, Target: game.PlayerTarget(game.Player2)}, // own, ignored
	}
	g.EndState.Players[game.Player1].CommanderCasts = 1
	return g
}

func tempoLossGame() rules.GameResult {
	g := rules.GameResult{HasWinner: true, Winner: game.Player2, Cards: tempoCards()}
	g.Turns = []rules.TurnLog{
		{TurnNumber: 1, ActivePlayer: game.Player1, Actions: []rules.ActionLog{
			{Player: game.Player1, Action: action.PlayLand(tempoCardID)},
		}},
		{TurnNumber: 2, ActivePlayer: game.Player1, Actions: []rules.ActionLog{
			{Player: game.Player1, Action: action.CastSpell(tempoCardID, nil, 0, nil)},
		}},
	}
	g.EndState.Players[game.Player1].CommanderCasts = 0
	return g
}

func TestComputeTempo(t *testing.T) {
	result := makeSimulation([]rules.GameResult{tempoWinGame(), tempoLossGame()}, nil)
	tempo := computeTempo(result, game.Player1)

	if tempo.OnlineGames != 2 || !approxEqual(tempo.ComeOnlineTurn, 2.0) {
		t.Errorf("come online = %.2f over %d games, want 2.0 over 2", tempo.ComeOnlineTurn, tempo.OnlineGames)
	}
	if !approxEqual(tempo.DamageDealtPerGame, 4.0) {
		t.Errorf("DamageDealtPerGame = %v, want 4.0 (8 damage / 2 games)", tempo.DamageDealtPerGame)
	}
	// 8 damage over 5 tested turns (3 + 2).
	if !approxEqual(tempo.DamagePerActiveTurn, 8.0/5.0) {
		t.Errorf("DamagePerActiveTurn = %v, want %v", tempo.DamagePerActiveTurn, 8.0/5.0)
	}
}

func TestComputeCommander(t *testing.T) {
	result := makeSimulation([]rules.GameResult{tempoWinGame(), tempoLossGame()}, nil)
	commander := computeCommander(result, game.Player1)

	if !approxEqual(commander.AverageCasts, 0.5) {
		t.Errorf("AverageCasts = %v, want 0.5", commander.AverageCasts)
	}
	if commander.CastDistribution[1] != 1 || commander.CastDistribution[0] != 1 {
		t.Errorf("CastDistribution = %v, want {0:1, 1:1}", commander.CastDistribution)
	}
	// The only tested win cast the commander.
	if commander.WinsWithCommander != 1 || commander.WinsWithoutCommander != 0 {
		t.Errorf("wins with/without = %d/%d, want 1/0", commander.WinsWithCommander, commander.WinsWithoutCommander)
	}
	if !approxEqual(commander.DependencyRate, 1.0) {
		t.Errorf("DependencyRate = %v, want 1.0", commander.DependencyRate)
	}
}

func TestComputeInteraction(t *testing.T) {
	result := makeSimulation([]rules.GameResult{tempoWinGame(), tempoLossGame()}, nil)
	interaction := computeInteraction(result, game.Player1)

	// Opponent targeted the tested permanent once and the tested player once; the
	// tested deck's own targeting of an opponent is excluded.
	if interaction.TargetedByOpponents != 2 {
		t.Errorf("TargetedByOpponents = %d, want 2", interaction.TargetedByOpponents)
	}
	if !approxEqual(interaction.TargetedByOpponentsPerGame, 1.0) {
		t.Errorf("TargetedByOpponentsPerGame = %v, want 1.0", interaction.TargetedByOpponentsPerGame)
	}
}

func TestTempoMetricsInReport(t *testing.T) {
	result := makeSimulation([]rules.GameResult{tempoWinGame(), tempoLossGame()}, nil)
	report := Generate(result, Options{
		TestedSeat: game.Player1,
		DeckNames:  [game.NumPlayers]string{"Mine", "A", "B", "C"},
	})
	if report.Tempo.OnlineGames != 2 || report.Commander.WinsWithCommander != 1 || report.Interaction.TargetedByOpponents != 2 {
		t.Errorf("report tempo/commander/interaction not wired: %+v %+v %+v",
			report.Tempo, report.Commander, report.Interaction)
	}
}
