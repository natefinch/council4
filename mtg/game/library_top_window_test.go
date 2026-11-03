package game

import "testing"

// TestLibraryTopWindowNestsAndCloses checks CR 401.5's window bookkeeping: a
// top card is hidden only while a window is open and differs from the
// outermost window's snapshot (an empty library included), nested windows
// keep that snapshot, and each closer ends only its own window.
func TestLibraryTopWindowNestsAndCloses(t *testing.T) {
	g := NewGame([NumPlayers]PlayerConfig{})
	library := &g.Players[Player1].Library
	first, second, third := g.IDGen.Next(), g.IDGen.Next(), g.IDGen.Next()
	library.Add(first)
	library.Add(second)

	if g.LibraryTopHidden(Player1) {
		t.Fatal("a top card is hidden with no window open")
	}
	closeOuter := g.OpenLibraryTopWindow()
	if g.LibraryTopHidden(Player1) || g.LibraryTopHidden(Player2) {
		t.Fatal("an unchanged top card (or an unchanged empty library) is hidden")
	}
	library.Remove(second)
	closeInner := g.OpenLibraryTopWindow()
	if !g.LibraryTopHidden(Player1) {
		t.Fatal("a nested window replaced the outermost snapshot")
	}
	library.Add(second)
	if g.LibraryTopHidden(Player1) {
		t.Fatal("the original top card, back on top, is hidden")
	}
	g.Players[Player2].Library.Add(third)
	if !g.LibraryTopHidden(Player2) {
		t.Fatal("a card put into an empty library isn't hidden")
	}
	clone := g.Clone()
	closeInner()
	if len(g.LibraryTopWindows) != 1 || !g.LibraryTopHidden(Player2) {
		t.Fatalf("windows after closing the inner one = %v, want the outer one open", g.LibraryTopWindows)
	}
	closeOuter()
	if len(g.LibraryTopWindows) != 0 || g.LibraryTopHidden(Player2) {
		t.Fatal("the top card is still hidden after every window closed")
	}
	if len(clone.LibraryTopWindows) != 2 || !clone.LibraryTopHidden(Player2) {
		t.Fatalf("clone windows = %v, want both windows kept independently", clone.LibraryTopWindows)
	}
	if g.LibraryTopHidden(PlayerID(-1)) || g.LibraryTopHidden(PlayerID(NumPlayers)) {
		t.Fatal("an out-of-range player is hidden")
	}
}
