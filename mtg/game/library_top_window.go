package game

import "github.com/natefinch/council4/mtg/game/id"

// LibraryTopWindow is each player's library top card, by PlayerID, when an
// action began. A zero ID records an empty library.
type LibraryTopWindow [NumPlayers]id.ID

// OpenLibraryTopWindow begins CR 401.5's window for an action such as casting
// a spell, and returns the function that ends it. While any window is open, a
// library's top card that differs from the outermost window's snapshot is
// neither revealed nor lookable; nested windows keep that outermost snapshot.
func (g *Game) OpenLibraryTopWindow() (closeWindow func()) {
	var window LibraryTopWindow
	for i, player := range g.Players {
		if player == nil {
			continue
		}
		if top, ok := player.Library.Top(); ok {
			window[i] = top
		}
	}
	g.LibraryTopWindows = append(g.LibraryTopWindows, window)
	depth := len(g.LibraryTopWindows)
	return func() { g.LibraryTopWindows = g.LibraryTopWindows[:depth-1] }
}

// LibraryTopHidden reports that player's library top changed during an open
// CR 401.5 window, so the new top card can't be revealed or looked at until
// the action finishes. Outside any window it is always false.
func (g *Game) LibraryTopHidden(player PlayerID) bool {
	if len(g.LibraryTopWindows) == 0 || int(player) < 0 || int(player) >= NumPlayers || g.Players[player] == nil {
		return false
	}
	top, _ := g.Players[player].Library.Top()
	return top != g.LibraryTopWindows[0][player]
}
