package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
)

// ActionObserver is an optional capability for PlayerAgents that want to be
// notified of the actions other players take, so they can maintain belief state
// about the table even when it is not their turn (see
// docs/research/card-game-ai-research.md §9.1, "Inform OTHER agents what
// happened"). Agents that do not implement it are never notified and incur no
// cost.
type ActionObserver interface {
	// ObserveAction is called after actor's action has been applied to the
	// game. obs is the notified player's own fog-of-war view of the resulting
	// state. The action is redacted for fog-of-war (see action.Action.Redacted),
	// so a face-down cast does not reveal the hidden card's identity. It is not
	// called for the observer's own actions, since an agent already knows the
	// action it chose.
	ObserveAction(actor game.PlayerID, act action.Action, obs PlayerObservation)
}

// notifyActionObservers informs every seat other than the actor whose agent
// implements ActionObserver that actor applied act. Each observer receives its
// own fog-of-war observation of the post-action state. A nil agent or an agent
// that does not implement ActionObserver is skipped.
func (*Engine) notifyActionObservers(g *game.Game, agents [game.NumPlayers]PlayerAgent, actor game.PlayerID, act action.Action) {
	for i := range agents {
		if game.PlayerID(i) == actor {
			continue
		}
		observer, ok := agents[i].(ActionObserver)
		if !ok {
			continue
		}
		observer.ObserveAction(actor, act.Redacted(), observe(g, game.PlayerID(i)))
	}
}
