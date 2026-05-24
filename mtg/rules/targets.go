package rules

import (
	"strings"

	"github.com/natefinch/council4/mtg/game"
)

func targetChoicesForSpell(g *game.Game, card *game.CardDef) [][]game.Target {
	specs := spellTargetSpecs(card)
	if len(specs) == 0 {
		return [][]game.Target{nil}
	}
	if len(specs) != 1 || !isSinglePlayerTargetSpec(specs[0]) {
		return nil
	}

	var choices [][]game.Target
	for playerID := game.Player1; playerID < game.NumPlayers; playerID++ {
		if !isPlayerAlive(g, playerID) {
			continue
		}
		choices = append(choices, []game.Target{game.PlayerTarget(playerID)})
	}
	return choices
}

func targetsValidForSpell(g *game.Game, card *game.CardDef, targets []game.Target) bool {
	specs := spellTargetSpecs(card)
	if len(specs) == 0 {
		return len(targets) == 0
	}
	if len(specs) != 1 || !isSinglePlayerTargetSpec(specs[0]) || len(targets) != 1 {
		return false
	}
	target := targets[0]
	return target.Kind == game.TargetPlayer && isPlayerAlive(g, target.PlayerID)
}

func spellTargetSpecs(card *game.CardDef) []game.TargetSpec {
	ability := firstSpellAbility(card)
	if ability == nil {
		return nil
	}
	return ability.Targets
}

func isSinglePlayerTargetSpec(spec game.TargetSpec) bool {
	return spec.MinTargets == 1 &&
		spec.MaxTargets == 1 &&
		targetSpecAllowsPlayers(spec)
}

func targetSpecAllowsPlayers(spec game.TargetSpec) bool {
	switch strings.ToLower(strings.TrimSpace(spec.Constraint)) {
	case "player", "target player", "any target":
		return true
	default:
		return false
	}
}

func isPlayerAlive(g *game.Game, playerID game.PlayerID) bool {
	if g == nil || playerID < 0 || int(playerID) >= len(g.Players) {
		return false
	}
	player := g.Players[playerID]
	return player != nil && !player.Eliminated && !g.TurnOrder.IsEliminated(playerID)
}
