package game

import "github.com/natefinch/council4/mtg/game/id"

// DecisionKind distinguishes rules decisions from personal/tournament decisions
// that a controller cannot make for another player (CR 723.5b, 723.6).
type DecisionKind uint8

// Decision kinds classify rules-owned and personal decisions.
const (
	DecisionRules DecisionKind = iota
	DecisionPersonal
)

// PlayerControlEffect is a game-owned, one-turn control effect (CR 723.1).
// Turn is zero while pending; CreatedTurn prevents consuming it in the turn
// during which it resolved. ID orders effects by creation, not source entry.
type PlayerControlEffect struct {
	ID             id.ID
	SourceCardID   id.ID
	SourceObjectID id.ID
	Controller     PlayerID
	Player         PlayerID
	CreatedTurn    int
	Turn           int
}

// SearchControlScope lasts only while Player searches their library. It takes
// precedence over turn control; all of its data can be copied into a search
// branch without capturing a live agent, Engine, or observation.
type SearchControlScope struct {
	Player     PlayerID
	Controller PlayerID
	Override   bool
	BeforeFind bool
}

// SearchCastingAllowed reports a real own-library search's before-find window.
func (g *Game) SearchCastingAllowed(player PlayerID) bool {
	for i := len(g.SearchControlScopes) - 1; i >= 0; i-- {
		if scope := g.SearchControlScopes[i]; scope.Player == player {
			return scope.BeforeFind && g.controlPlayerAlive(player)
		}
	}
	return false
}

// ManaPaymentAllowed permits only the actor currently paying a mana cost.
func (g *Game) ManaPaymentAllowed(player PlayerID) bool {
	return len(g.ManaPaymentScopes) > 0 && g.ManaPaymentScopes[len(g.ManaPaymentScopes)-1] == player
}

// SchedulePlayerControlEffect overwrites older pending effects for the same next
// turn (CR 723.1a), including if the new controller leaves before it begins.
// Control already applying in the current turn has a different duration.
func (g *Game) SchedulePlayerControlEffect(effect PlayerControlEffect) {
	kept := g.PlayerControlEffects[:0]
	for _, previous := range g.PlayerControlEffects {
		if previous.Player != effect.Player || previous.Turn != 0 {
			kept = append(kept, previous)
		}
	}
	kept = append(kept, effect)
	g.PlayerControlEffects = kept
}

func (g *Game) controlPlayerAlive(player PlayerID) bool {
	return player >= 0 && int(player) < len(g.Players) &&
		g.Players[player] != nil && !g.Players[player].Eliminated
}

func (g *Game) immediateDecisionController(player PlayerID) PlayerID {
	for i := len(g.SearchControlScopes) - 1; i >= 0; i-- {
		scope := g.SearchControlScopes[i]
		if scope.Override && scope.Player == player && g.controlPlayerAlive(scope.Controller) {
			return scope.Controller
		}
	}
	var newest id.ID
	controller := player
	for _, effect := range g.PlayerControlEffects {
		if effect.Player == player && effect.Turn == g.Turn.TurnNumber &&
			effect.Turn > 0 && effect.ID > newest && g.controlPlayerAlive(effect.Controller) {
			newest, controller = effect.ID, effect.Controller
		}
	}
	return controller
}

// DecisionMaker redirects a rules decision, never the action's actor or any
// object's controller. A controller's own decisions may themselves be controlled.
func (g *Game) DecisionMaker(player PlayerID, kind DecisionKind) PlayerID {
	if kind != DecisionRules || !g.controlPlayerAlive(player) {
		return player
	}
	var seen [NumPlayers]bool
	current := player
	for !seen[current] {
		seen[current] = true
		next := g.immediateDecisionController(current)
		if next == current {
			return current
		}
		current = next
	}
	// Circular control grants no additional decision maker.
	return player
}

// CanSeePlayerInformation implements CR 723.4: information is shared with the
// controller, not in the reverse direction, including through nested control.
func (g *Game) CanSeePlayerInformation(viewer, player PlayerID) bool {
	if viewer == player {
		return true
	}
	if !g.controlPlayerAlive(viewer) || !g.controlPlayerAlive(player) {
		return false
	}
	var seen [NumPlayers]bool
	current := player
	for !seen[current] {
		seen[current] = true
		current = g.immediateDecisionController(current)
		if current == viewer {
			return true
		}
	}
	return false
}

// LookingAtLibrary reports the scoped search permission, not permission to see
// a library merely because its owner is controlled during a turn.
func (g *Game) LookingAtLibrary(viewer, player PlayerID) bool {
	for _, scope := range g.SearchControlScopes {
		if scope.Player == player && g.CanSeePlayerInformation(viewer, player) {
			return true
		}
	}
	return false
}

// BeginPlayerControlTurn consumes every competing pending effect for the active
// player's next actual turn. Losing effects must not reappear if the winner's
// controller subsequently leaves. Repeated calls in the same turn are inert.
func (g *Game) BeginPlayerControlTurn() {
	var winner PlayerControlEffect
	kept := g.PlayerControlEffects[:0]
	for _, effect := range g.PlayerControlEffects {
		if !g.controlPlayerAlive(effect.Player) || !g.controlPlayerAlive(effect.Controller) ||
			(effect.Turn > 0 && effect.Turn != g.Turn.TurnNumber) {
			continue
		}
		if effect.Turn == 0 && effect.Player == g.Turn.ActivePlayer && effect.CreatedTurn < g.Turn.TurnNumber {
			if effect.ID > winner.ID {
				winner = effect
			}
			continue
		}
		kept = append(kept, effect)
	}
	if winner.ID != 0 {
		winner.Turn = g.Turn.TurnNumber
		kept = append(kept, winner)
	}
	g.PlayerControlEffects = kept
}

// EndPlayerControls removes effects involving a player who leaves (CR 800.4a).
func (g *Game) EndPlayerControls(player PlayerID) {
	for i := range g.SearchControlScopes {
		scope := &g.SearchControlScopes[i]
		if scope.Player == player || scope.Controller == player {
			scope.Override = false
		}
	}
	kept := g.PlayerControlEffects[:0]
	for _, effect := range g.PlayerControlEffects {
		if effect.Player != player && effect.Controller != player {
			kept = append(kept, effect)
		}
	}
	g.PlayerControlEffects = kept
}
