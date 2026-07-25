package cardgen

// The renderer's literal tables are generated from the mtg/game type graph
// rather than written by hand, so a constant added to mtg/game is renderable
// without a matching renderer edit. See cardgen/cmd/genrender.
//go:generate go run github.com/natefinch/council4/cardgen/cmd/genrender -root=..
