package rules

import (
	"github.com/natefinch/council4/mtg/game"
)

// handleVote resolves the Vote primitive: the "Starting with you, each player
// votes for <A> or <B>." voting interaction (CR 701.32). Each non-eliminated
// player, starting with the resolving controller and proceeding in turn order,
// casts a single vote for one of the named options. The result's amount is the
// signed margin (votes for option 0 minus votes for option 1) so the arm
// instructions gate on its sign through their result-gate amount range.
func handleVote(r *effectResolver, prim game.Vote) effectResolved {
	if len(prim.Options) != 2 {
		return effectResolved{accepted: true, succeeded: false}
	}
	tally := make([]int, len(prim.Options))
	for _, voter := range votersStartingWith(r.game, stackObjectController(r.obj)) {
		choice := r.engine.castVote(r.game, r.agents, voter, prim.Options, r.log)
		tally[choice]++
	}
	return effectResolved{
		accepted:  true,
		succeeded: true,
		amount:    tally[0] - tally[1],
	}
}

// votersStartingWith returns the non-eliminated players who vote, ordered to
// begin with start and proceed in turn order (CR 701.32: "Starting with you,
// each player votes ..."). When start itself is eliminated the order begins with
// the next active player. The loop terminates on returning to the first active
// voter, so it cannot spin even if start is eliminated.
func votersStartingWith(g *game.Game, start game.PlayerID) []game.PlayerID {
	first := start
	if g.TurnOrder.IsEliminated(start) {
		first = g.TurnOrder.NextActivePlayer(start)
	}
	voters := make([]game.PlayerID, 0, game.NumPlayers)
	for p := first; ; {
		voters = append(voters, p)
		p = g.TurnOrder.NextActivePlayer(p)
		if p == first {
			break
		}
	}
	return voters
}

// castVote asks one voter to choose a single option among the named choices and
// returns the chosen option index.
func (e *Engine) castVote(
	g *game.Game,
	agents [game.NumPlayers]PlayerAgent,
	voter game.PlayerID,
	options []string,
	log *TurnLog,
) int {
	choiceOptions := make([]game.ChoiceOption, len(options))
	for i, label := range options {
		choiceOptions[i] = game.ChoiceOption{Index: i, Label: label}
	}
	selected := e.chooseChoice(g, agents, game.ChoiceRequest{
		Kind:             game.ChoiceVote,
		Player:           voter,
		Prompt:           "Vote",
		Options:          choiceOptions,
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{0},
	}, log)
	if len(selected) == 1 && selected[0] >= 0 && selected[0] < len(options) {
		return selected[0]
	}
	return 0
}
