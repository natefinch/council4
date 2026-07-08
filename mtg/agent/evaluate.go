package agent

import (
	"math"

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
	// the table each cycle), so raw card count matters — but a card in hand is
	// worth strictly LESS than the permanent it becomes once cast: it cannot yet
	// attack, block, tap for mana, or pressure the table, and it costs a turn of
	// tempo to deploy. Valuing it below an average creature's board value (a 3/3 is
	// 4.0) is what makes one-ply search deploy its hand and develop a board instead
	// of hoarding cards and passing every turn (which stalls games to a draw).
	evalCardInHand = 2.0
	// evalLifeAtFull and evalStartingLife anchor a CONCAVE life curve (lifeValue):
	// life is worth evalLifeAtFull at the Commander starting total of
	// evalStartingLife, and its marginal value rises as a player drops toward 0.
	// A linear "0.1 per point" is wrong for how a strong player treats life: at 40
	// a single point is almost free (you pay it for a fetch land, a painland, a
	// Phyrexian cost without a thought), but at 5 it is precious (one more point is
	// the difference between living and dying). Modeling life concavely makes the
	// agent (a) stop paying life for no board gain when it is already low — the
	// original "activates an ability that costs 2 life and does nothing until it
	// dies" complaint — and value lifegain then, and (b) race a low opponent: as
	// the strongest opponent's life falls, the concave curve subtracts less of
	// their power, so pushing a wounded leader toward death registers as progress.
	// Life is still a shallow resource (§6, W_LIFE low): a full 40 is about one
	// creature, so board and cards dominate a healthy game.
	evalLifeAtFull   = 4.0
	evalStartingLife = 40.0
	// evalManaSource values one mana-producing permanent — the tempo/development
	// axis (§4.2): more available mana means more and bigger plays sooner.
	evalManaSource = 1.5
	// evalPermanentBase is the value of a noncreature permanent (a land, a mana
	// rock, an enchantment, an artifact engine). A permanent is the realized form
	// of the card that made it, so it is worth about a card in hand
	// (evalCardInHand) — otherwise casting a noncreature would net a loss (a token
	// board value minus the card spent), and one-ply search would never develop
	// a ramp aura, an anthem, or an engine, only creatures. Its specific payoff is
	// left to the rollout policy's per-card score (the search prior); this is the
	// floor that keeps deploying it from looking bad.
	evalPermanentBase = 2.0
	// evalCommanderOnBoard rewards having the commander in play, the recurring
	// engine or threat most Commander decks are built around (§1.2).
	evalCommanderOnBoard = 4.0
	// evalOpponentEliminated rewards each opponent already removed from the game.
	// Eliminating a player is the largest swing in a free-for-all: their board,
	// blockers, interaction, and future threats vanish, and the position moves one
	// seat closer to the win (§2, the goal is to be the last player standing). It
	// is deliberately large — a strong board's worth — so search takes a line that
	// eliminates a player, including a lethal attack on someone who is NOT the
	// current leader, which the "my power minus the strongest opponent" core would
	// otherwise barely reward (killing a non-leader leaves the max unchanged). That
	// is what turns a durdling board stall into a closed-out game.
	evalOpponentEliminated = 20.0
	// evalClosingWeight scales a smooth "closing" reward for having an opponent
	// near death on ANY loss clock — low life (CR 104.3a: 0 or less loses), poison
	// (CR 704.5c: ten counters), or commander damage (CR 903.14a: 21 from one
	// commander). evalOpponentEliminated alone is a step function (nothing until a
	// kill, then +20), so a position one hit from ending a player looks no better
	// than a fresh table and four strong-eval agents durdle to a draw. Scaling a
	// fraction of the elimination reward by how far the MOST killable opponent has
	// been pushed toward a loss turns that step into a gradient: the agent commits
	// to finishing the target it is closest to killing (real players focus one kill
	// at a time), which produces decisive games instead of turn-limit stalls. It is
	// only a fraction of evalOpponentEliminated so the actual kill still dominates.
	evalClosingWeight = 0.4
	// evalPoisonToLose and evalCommanderDamageToLose are the alternate death
	// thresholds a player can be raced toward: ten poison counters (CR 704.5c) and
	// 21 combat damage from a single commander (CR 903.14a). They convert those
	// clocks into the same [0,1) kill-progress fraction as life (killProgress).
	evalPoisonToLose          = 10.0
	evalCommanderDamageToLose = 21.0
)

// Evaluate returns a position value for obs.Player: higher is a better position
// for that player. It is the search agent's leaf/rollout evaluation and a
// principled, position-level companion to the per-action heuristics — a
// replacement target for coarse action constants.
//
// The value is that player's power minus the strongest opponent's power, plus a
// reward for every opponent already eliminated, so it rises when the agent
// develops (board, mana, cards, life), when the table's leader is set back, and
// when a rival is removed from the game, and falls when an opponent pulls ahead.
// Winning the game is evalWin and being eliminated is evalLoss, dominating every
// heuristic term.
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
	eliminatedOpponents := 0
	maxKillProgress := 0.0
	players := obs.Players()
	for i := range players {
		if players[i].ID == me {
			continue
		}
		if players[i].Eliminated {
			eliminatedOpponents++
			continue
		}
		livingOpponents++
		if power := playerPower(obs, players[i].ID); power > strongestOpponent {
			strongestOpponent = power
		}
		if progress := killProgress(players[i]); progress > maxKillProgress {
			maxKillProgress = progress
		}
	}
	if livingOpponents == 0 {
		return evalWin
	}
	closing := evalClosingWeight * evalOpponentEliminated * maxKillProgress
	return playerPower(obs, me) - strongestOpponent +
		evalOpponentEliminated*float64(eliminatedOpponents) + closing
}

// killProgress rates how close one opponent is to losing, on its fastest loss
// clock, as a fraction in [0,1): 0 is a fresh player, approaching 1 is one hit
// from dead. It takes the max over the three clocks a Commander player can die
// to — life toward 0 (CR 104.3a), poison toward ten counters (CR 704.5c), and
// damage from a single commander toward 21 (CR 903.14a) — because a strong
// player finishes an opponent on whichever clock is furthest along, not just the
// life total. It never reaches 1 (the actual loss is evalOpponentEliminated); it
// is the gradient that makes the agent commit to the kill it is closest to.
func killProgress(opponent rules.PlayerView) float64 {
	progress := 1.0 - float64(max(0, opponent.Life))/evalStartingLife

	if poison := float64(opponent.PoisonCounters) / evalPoisonToLose; poison > progress {
		progress = poison
	}

	worstCommanderDamage := 0
	for _, dmg := range opponent.CommanderDamage {
		if dmg > worstCommanderDamage {
			worstCommanderDamage = dmg
		}
	}
	if cmdr := float64(worstCommanderDamage) / evalCommanderDamageToLose; cmdr > progress {
		progress = cmdr
	}

	if progress < 0 {
		return 0
	}
	if progress > 1 {
		return 1
	}
	return progress
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
	power += lifeValue(view.Life)
	if commanderOnBattlefield(obs, view) {
		power += evalCommanderOnBoard
	}
	return power
}

// lifeValue is the concave value of a life total: evalLifeAtFull at the starting
// total, tapering as life climbs and steepening as it falls toward 0, so a point
// of life is nearly free when a player is healthy and precious when they are low
// (see evalLifeAtFull). It uses a square root — a simple curve with exactly that
// shape — scaled so lifeValue(evalStartingLife) == evalLifeAtFull. Life at or
// below 0 is worthless (that player has lost); life above the starting total
// keeps rising, but ever more slowly.
func lifeValue(life int) float64 {
	if life <= 0 {
		return 0
	}
	return evalLifeAtFull * math.Sqrt(float64(life)/evalStartingLife)
}

// permanentPower values one permanent as board presence: a creature by its
// tapped-agnostic board value (evasion- and keyword-aware, via
// permanentBoardValue), plus a development bonus if it also produces mana (a mana
// dork); any noncreature permanent by evalPermanentBase, the value of a realized
// card. It ignores tapped state because a position's value looks across turns,
// where a creature that is momentarily tapped — for example one that just
// attacked — untaps before it matters again. A noncreature that produces mana (a
// land or rock) is already covered by evalPermanentBase, which equals a land's
// previous base-plus-mana value, so noncreature mana sources and other
// noncreatures are valued alike; their specific payoff is the search prior's job.
func permanentPower(permanent rules.PermanentView) float64 {
	if isCreaturePermanent(permanent) {
		power := permanentBoardValue(permanent)
		if permanent.ProducesMana {
			power += evalManaSource
		}
		return power
	}
	return evalPermanentBase
}
