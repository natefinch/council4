// Package authored owns non-generated Card Implementations.
// Legacy corpus generation only writes letter packages, never this directory.
package authored

import "github.com/natefinch/council4/mtg/cards/cardset"

// Cards is extended by completed, behavior-verified authored card slices.
// The program infrastructure itself does not promote a synthetic parity fixture.
var Cards = []cardset.Entry{Terastodon, HiveMind, GiftsUngiven, GildedDrake, WorldgorgerDragon, NimDeathmantle, Abundance, BloodMoon, Muldrotha}
