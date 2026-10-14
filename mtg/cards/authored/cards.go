// Package authored owns non-generated Card Implementations.
// Legacy corpus generation only writes letter packages, never this directory.
package authored

import (
	"slices"

	"github.com/natefinch/council4/mtg/cards/cardset"
)

//go:generate go run ../../../cmd/sourcebindings

// Cards is extended by completed, behavior-verified authored card slices.
// The program infrastructure itself does not promote a synthetic parity fixture.
var Cards = verifiedEntries()

func verifiedEntries() []cardset.Entry {
	entries := []cardset.Entry{Terastodon, HiveMind, GiftsUngiven, GildedDrake, WorldgorgerDragon, NimDeathmantle, Abundance, BloodMoon, Muldrotha, Skullbriar, Mindslaver, GenerousGift, Pongify, RapidHybridization, Resculpt, Twincast, Reverberate, DualcasterMage, CacklingCounterpart, FatedInfatuation, Intuition, RealmsUncharted, ChainsOfMephistopheles, Humility, Opalescence}
	for i := range entries {
		if entries[i].Verification == "" {
			entries[i].Verification = "mtg/rules named-card scenarios; mixed production acceptance (#3278)"
		}
		entries[i].SourceRevision = SourceRevision
		entries[i].RulesRevision = RulesRevision
		entries[i].Programs = slices.Clone(entries[i].Programs)
		for j := range entries[i].Programs {
			entries[i].Programs[j].SourceRevision = SourceRevision
			entries[i].Programs[j].RulesRevision = RulesRevision
		}
	}
	return entries
}
