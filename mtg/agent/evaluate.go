package agent

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/rules"
)

// Position-evaluation weights (see docs/research/COMMANDER-AGENT-PLAYBOOK.md §6).
// Evaluate scores a whole position rather than a single action, so it is the leaf
// and rollout evaluator the search agent needs (ADR 0011) and a principled basis
// the action heuristics approximate. Every term is expressed in the threat
// currency of threat.go (a 3/3's board value ~ 4), so board, card, life, and
// development contributions are comparable.
const (
	// evalWin and evalLoss dominate every heuristic term, so a decided position
	// is always valued above or below any contested one. They are finite (not
	// infinity) so search can still compare "win in 1" against "win in 3" by the
	// heuristic terms of intermediate states without NaNs.
	evalWin  = 1_000_000.0
	evalLoss = -1_000_000.0

	// evalCardInHand values a card in hand. Card advantage is the dominant
	// long-game resource in multiplayer Commander (§4.1: you fall ~3 cards behind
	// the table each cycle), so a held card is worth about a small creature.
	evalCardInHand = 4.0
	// evalLifePerPoint values one life point. Life is a real but shallow resource
	// (§6, W_LIFE low), so a full 40 life contributes about one creature.
	evalLifePerPoint = 0.1
	// evalManaSource values one mana-producing permanent — the tempo/development
	// axis (§4.2): more available mana means more and bigger plays sooner.
	evalManaSource = 1.5
	// evalPermanentBase is the floor value of a noncreature, non-mana permanent
	// (an enchantment, an artifact engine), so board development still registers
	// even when the permanent's payoff is not modeled.
	evalPermanentBase = 0.5
	// evalCommanderOnBoard rewards having the commander in play, the recurring
	// engine or threat most Commander decks are built around (§1.2).
	evalCommanderOnBoard = 4.0
)

// Evaluate returns a position value for obs.Player: higher is a better position
// for that player. It is the search agent's leaf/rollout evaluation and a
// principled, position-level companion to the per-action heuristics — a
// replacement target for coarse action constants.
//
// The value is that player's power minus the strongest opponent's power, so it
// rises when the agent develops (board, mana, cards, life) and when the table's
// leader is set back, and falls when an opponent pulls ahead. Winning the game is
// evalWin and being eliminated is evalLoss, dominating every heuristic term.
//
// Evaluate reads only the observation, so it respects fog of war: opponents'
// board, life, and hand sizes are public, and their hidden hand contents are
// never consulted. That makes it valid both on the live observation and on a
// determinized world a search agent samples.
func Evaluate(obs rules.PlayerObservation) float64 {
	me := obs.Player
	if obs.PlayerState(me).Eliminated {
		return evalLoss
	}

	strongestOpponent := 0.0
	livingOpponents := 0
	players := obs.Players()
	for i := range players {
		if players[i].ID == me || players[i].Eliminated {
			continue
		}
		livingOpponents++
		if power := playerPower(obs, players[i].ID); power > strongestOpponent {
			strongestOpponent = power
		}
	}
	if livingOpponents == 0 {
		return evalWin
	}
	return playerPower(obs, me) - strongestOpponent
}

// playerPower scores how well positioned one player is, aggregating board
// presence, card advantage, life, and having the commander in play — all in the
// threat currency of threat.go. An eliminated player has no power.
func playerPower(obs rules.PlayerObservation, playerID game.PlayerID) float64 {
	view := obs.PlayerState(playerID)
	if view.Eliminated {
		return 0
	}

	power := 0.0
	battlefield := obs.Battlefield()
	for i := range battlefield {
		permanent := battlefield[i]
		if permanent.Controller != playerID || permanent.PhasedOut {
			continue
		}
		power += permanentPower(permanent)
	}
	power += float64(view.HandSize) * evalCardInHand
	power += float64(max(0, view.Life)) * evalLifePerPoint
	if commanderOnBattlefield(obs, view) {
		power += evalCommanderOnBoard
	}
	return power
}

// permanentPower values one permanent as board presence: a creature by its
// tapped-agnostic board value (evasion- and keyword-aware, via
// permanentBoardValue), any other permanent by a small base, plus a development
// bonus for a mana source (a rock, dork, or land). It ignores tapped state
// because a position's value looks across turns, where a creature that is
// momentarily tapped — for example one that just attacked — untaps before it
// matters again.
func permanentPower(permanent rules.PermanentView) float64 {
	power := evalPermanentBase
	if isCreaturePermanent(permanent) {
		power = permanentBoardValue(permanent)
	}
	if permanent.ProducesMana {
		power += evalManaSource
	}
	return power
}
