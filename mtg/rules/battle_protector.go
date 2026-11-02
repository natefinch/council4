package rules

import (
	"fmt"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// battleProtector returns the player protecting battle (CR 310.9), if one has
// been designated.
func battleProtector(battle *game.Permanent) (game.PlayerID, bool) {
	if battle == nil || !battle.BattleProtector.Exists {
		return 0, false
	}
	return battle.BattleProtector.Val, true
}

// canProtectBattle reports whether player can be battle's protector under its
// current controller and battle types: only an opponent of a Siege's controller
// (CR 310.12a), and only the controller of a battle without battle types (CR
// 310.9a). A player who has left the game can't be a protector.
func canProtectBattle(g *game.Game, battle *game.Permanent, player game.PlayerID) bool {
	if !isPlayerAlive(g, player) {
		return false
	}
	controller := effectiveController(g, battle)
	if permanentHasSubtype(g, battle, types.Siege) {
		return player != controller
	}
	return player == controller
}

// battleProtectorCandidates lists, in player order, the players who can be
// battle's protector.
func battleProtectorCandidates(g *game.Game, battle *game.Permanent) []game.PlayerID {
	var candidates []game.PlayerID
	for player := range game.PlayerID(game.NumPlayers) {
		if canProtectBattle(g, battle, player) {
			candidates = append(candidates, player)
		}
	}
	return candidates
}

// chooseBattleProtector has battle's controller choose its protector from the
// appropriate players. With a single candidate it is chosen without a prompt;
// with none it reports false.
func chooseBattleProtector(e *Engine, g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog, battle *game.Permanent) (game.PlayerID, bool) {
	candidates := battleProtectorCandidates(g, battle)
	switch len(candidates) {
	case 0:
		return 0, false
	case 1:
		return candidates[0], true
	}
	options := make([]game.ChoiceOption, len(candidates))
	for i, candidate := range candidates {
		options[i] = game.ChoiceOption{Index: i, Label: fmt.Sprintf("Player %d", candidate+1)}
	}
	selected := e.chooseChoice(g, agents, game.ChoiceRequest{
		Kind:             game.ChoicePlayer,
		Player:           effectiveController(g, battle),
		Prompt:           fmt.Sprintf("Choose the protector of %s.", permanentChoiceLabel(g, battle)),
		Options:          options,
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{0},
	}, log)
	if len(selected) != 1 || selected[0] < 0 || selected[0] >= len(candidates) {
		return candidates[0], true
	}
	return candidates[selected[0]], true
}

// chooseEnteringBattleProtector has the controller of a battle entering the
// battlefield choose its protector as it enters (CR 310.9a, 310.12a). A
// permanent that isn't a battle as it enters gets no protector. If no player can
// be chosen, the battle enters without one and the CR 310.11 state-based action
// puts it into its owner's graveyard.
func chooseEnteringBattleProtector(ctx enterBattlefieldContext, g *game.Game, permanent *game.Permanent) {
	if !permanentHasType(g, permanent, types.Battle) {
		return
	}
	engine := ctx.engine
	if engine == nil {
		engine = NewEngine(nil)
	}
	if protector, ok := chooseBattleProtector(engine, g, ctx.agents, ctx.log, permanent); ok {
		permanent.BattleProtector = opt.Val(protector)
	}
}

// battleIsAttacked reports whether any attacking creature is currently attacking
// battleID.
func battleIsAttacked(g *game.Game, battleID id.ID) bool {
	if g.Combat == nil {
		return false
	}
	for _, attack := range g.Combat.Attackers {
		if !attack.Target.NoTarget && attack.Target.BattleID == battleID {
			return true
		}
	}
	return false
}

type stateBasedBattleProtector struct {
	battle    *game.Permanent
	protector game.PlayerID
}

// battleProtectorStateBasedActions applies the battle-protector state-based
// actions (CR 310.11, 704.5) for the battles player controls. A battle whose
// protector has left the game or was never designated gets a new protector only
// while no creature is attacking it; a battle whose protector is a player who
// can't protect it gets a new one regardless. The controller chooses among the
// appropriate players; if none can be chosen, the battle is put into its
// owner's graveyard. A permanent that has stopped being a battle keeps its
// protector (CR 310.9g).
func battleProtectorStateBasedActions(g *game.Game, player game.PlayerID) ([]stateBasedBattleProtector, []id.ID) {
	var assignments []stateBasedBattleProtector
	var departures []id.ID
	for _, permanent := range g.Battlefield {
		if !activeBattlefieldPermanent(permanent) || effectiveController(g, permanent) != player ||
			!permanentHasType(g, permanent, types.Battle) {
			continue
		}
		protector, designated := battleProtector(permanent)
		switch {
		case designated && canProtectBattle(g, permanent, protector):
			continue
		case (!designated || !isPlayerAlive(g, protector)) && battleIsAttacked(g, permanent.ObjectID):
			continue
		}
		var chosen game.PlayerID
		var ok bool
		if ctx, exists := replacementChoiceContextFor(g); exists {
			chosen, ok = chooseBattleProtector(ctx.engine, g, ctx.agents, ctx.log, permanent)
		} else if candidates := battleProtectorCandidates(g, permanent); len(candidates) > 0 {
			chosen, ok = candidates[0], true
		}
		if !ok {
			departures = append(departures, permanent.ObjectID)
			continue
		}
		assignments = append(assignments, stateBasedBattleProtector{battle: permanent, protector: chosen})
	}
	return assignments, departures
}
