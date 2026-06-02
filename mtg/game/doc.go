// Package game provides the core data structures for a 4-player Commander
// Magic: The Gathering game engine.
//
// This package contains all the interrelated game types — cards, permanents,
// zones, abilities, players, the stack, turn structure, and combat state.
// It is designed as a pure in-memory representation with no display or
// physical card concerns.
//
// Leaf packages with no game dependencies live in sub-packages:
//   - [github.com/natefinch/council4/mtg/game/id] — unique object identifiers
//   - [github.com/natefinch/council4/mtg/game/color] — card colors and color identity
//   - [github.com/natefinch/council4/mtg/game/cost] — printed mana costs
//   - [github.com/natefinch/council4/mtg/game/mana] — produced mana and mana pools
//   - [github.com/natefinch/council4/mtg/game/counter] — counter types and tracking
//   - [github.com/natefinch/council4/mtg/game/action] — player action types
package game
