package mana

import "github.com/natefinch/council4/mtg/game/color"

// Unit is a spendable unit of mana. Color records the mana color or colorless;
// Snow records whether the mana was produced by a snow source for costs such as
// {S}.
type Unit struct {
	Color color.Color
	Snow  bool
}
