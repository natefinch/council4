package mana

// Color represents a color of mana in the game.
type Color string

// Colors of mana, to be used in mana pools etc.
const (
	W = Color("W") // white mana
	U = Color("U") // blue mana
	B = Color("B") // black mana
	R = Color("R") // red mana
	G = Color("G") // green mana
	C = Color("◇") // colorless mana
)
