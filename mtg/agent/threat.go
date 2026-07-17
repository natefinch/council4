package agent

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/rules"
)

// Threat-model weights. A permanent's threat is dominated by an attacking
// creature's effective power, amplified by evasion and damage-multiplying
// keywords, with a small base so noncreature permanents can still be ranked and
// removed. Opponent threat aggregates the permanents a player controls plus a
// small staying-power term for life, so the strategy pressures and removes the
// most dangerous board rather than kingmaking a near-dead player.
const (
	threatPerPower     = 1.0
	threatCreatureBase = 1.0
	threatNoncreature  = 0.5
	threatTappedScale  = 0.5
	threatEvasionBonus = 2.0
	threatDeathtouch   = 2.0
	threatDoubleStrike = 2.0
	threatCommander    = 4.0
	threatLifePerPoint = 0.05
	threatScoreUnit    = 3.0
)

// permanentThreat scores how dangerous a single permanent is right now. It is a
// pure function of the permanent's public effective characteristics, discounting
// a tapped permanent as a less immediate threat this turn. Threat assessment
// (targeting, combat, blocks) uses it; position value (Evaluate) uses the
// tapped-agnostic permanentBoardValue instead, since a creature untaps every
// turn.
func permanentThreat(permanent rules.PermanentView) float64 {
	threat := permanentBoardValue(permanent)
	// A tapped permanent is a less immediate threat this turn.
	if permanent.Tapped {
		threat *= threatTappedScale
	}
	return threat
}

// permanentBoardValue is a permanent's value ignoring whether it is tapped: a
// creature by its power and damage-relevant keywords, a noncreature by a small
// base. It is the tapped-agnostic core of permanentThreat, used by position
// evaluation where a momentarily tapped creature (for example one that just
// attacked) untaps before it matters again, so tapping should not lower the
// position's value.
func permanentBoardValue(permanent rules.PermanentView) float64 {
	if !isCreaturePermanent(permanent) {
		return threatNoncreature
	}
	power := max(0, permanent.Power)
	threat := threatCreatureBase + threatPerPower*float64(power)
	if permanent.HasKeyword(game.Flying) || permanent.HasKeyword(game.Menace) || permanent.HasKeyword(game.Trample) {
		threat += threatEvasionBonus
	}
	if permanent.HasKeyword(game.DoubleStrike) {
		threat += threatDoubleStrike + threatPerPower*float64(power)
	}
	if permanent.HasKeyword(game.Deathtouch) {
		threat += threatDeathtouch
	}
	return threat
}

func isCreaturePermanent(permanent rules.PermanentView) bool {
	return slices.Contains(permanent.Types, types.Creature)
}

// ThreatModel ranks opponents by how dangerous their board is, computed once
// from an observation.
type ThreatModel struct {
	observer game.PlayerID
	byPlayer map[game.PlayerID]float64
}

// NewThreatModel builds a ThreatModel for the observing player.
func NewThreatModel(obs rules.PlayerObservation) ThreatModel {
	model := ThreatModel{observer: obs.Player, byPlayer: make(map[game.PlayerID]float64)}
	battlefield := obs.Battlefield()
	for i := range battlefield {
		// Phased-out permanents cannot attack, block, or be targeted, so they
		// do not contribute to a player's current threat.
		if battlefield[i].PhasedOut {
			continue
		}
		model.byPlayer[battlefield[i].Controller] += permanentThreat(battlefield[i])
	}
	players := obs.Players()
	for i := range players {
		player := players[i]
		if player.Eliminated {
			model.byPlayer[player.ID] = 0
			continue
		}
		model.byPlayer[player.ID] += threatLifePerPoint * float64(max(0, player.Life))
		if commanderOnBattlefield(obs, player) {
			model.byPlayer[player.ID] += threatCommander
		}
	}
	return model
}

// PlayerThreat returns the threat score for a player.
func (m ThreatModel) PlayerThreat(playerID game.PlayerID) float64 {
	return m.byPlayer[playerID]
}

// HighestThreatOpponent returns the most threatening opponent of the observing
// player, along with its threat. ok is false when there is no living opponent.
func (m ThreatModel) HighestThreatOpponent() (game.PlayerID, float64, bool) {
	best := game.PlayerID(0)
	bestThreat := 0.0
	found := false
	for playerID, threat := range m.byPlayer {
		if playerID == m.observer {
			continue
		}
		if !found || threat > bestThreat || (threat == bestThreat && playerID < best) {
			best = playerID
			bestThreat = threat
			found = true
		}
	}
	return best, bestThreat, found
}

func commanderOnBattlefield(obs rules.PlayerObservation, player rules.PlayerView) bool {
	if player.CommanderInstanceID == 0 {
		return false
	}
	battlefield := obs.Battlefield()
	for i := range battlefield {
		if slices.Contains(battlefield[i].CommanderInstanceIDs, player.CommanderInstanceID) {
			return true
		}
	}
	return false
}

// Archenemy-paint weights (playbook §3.1, §10.4). In a four-player game the
// scariest-looking player draws the other three's attacks and removal, so once
// the agent is clearly ahead it should stop painting a bigger target and let the
// threat baton pass rather than over-committing.
const (
	// archenemyMargin is how far ahead in aggregate threat the agent must be
	// before it is the table's clear archenemy — one large creature's worth of
	// lead in the threat currency.
	archenemyMargin = 6.0
	// archenemyPaintPenalty discourages deploying yet another creature while
	// already the archenemy. It is smaller than a creature's deploy value, so the
	// agent still develops; it just does so less eagerly once it is ahead.
	archenemyPaintPenalty = 12.0
)

// archenemyPaint penalizes deploying another creature when the agent is already
// the table's clear archenemy, so it stops painting a bigger target once it is
// ahead. It applies only to creatures and only past archenemyMargin of
// aggregate-threat lead over the strongest opponent, and is scaled by
// personality: an aggressive deck accepts the paint and piles on, a political
// deck avoids it more.
func archenemyPaint(obs rules.PlayerObservation, card rules.CardView, personality Personality) float64 {
	if !isCreature(card) {
		return 0
	}
	model := NewThreatModel(obs)
	_, opponentThreat, ok := model.HighestThreatOpponent()
	if !ok {
		return 0
	}
	if model.PlayerThreat(obs.Player)-opponentThreat <= archenemyMargin {
		return 0
	}
	return archenemyPaintPenalty * personality.paintScale()
}
