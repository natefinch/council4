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
	entries = append(entries, Bedevil, Terminate, Vindicate, Disenchant, Naturalize)
	entries = append(entries, AnOfferYouCantRefuse, Counterspell, Negate, EssenceScatter, Disallow)
	entries = append(entries, EladamrisCall, SteelshapersGift, Fabricate, IdyllicTutor, SylvanScrying, DemonicTutor, DiabolicTutor)
	entries = append(entries, DeathDenied, WildestDreams, RatsFeast, ShatteredCrypt, ReturnToTheRanks, ChampionOfStraySouls)
	entries = append(entries, WarpWorld, PanglacialWurm)
	for i := range entries {
		entries[i].ReplaceLegacy = replacesCompiledEntry(entries[i].Name)
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

func replacesCompiledEntry(name string) bool {
	// These exact authored identities intentionally supersede their newly
	// installed compiler definitions; the generated sources remain available.
	switch name {
	case "An Offer You Can't Refuse", "Bedevil", "Cackling Counterpart",
		"Champion of Stray Souls", "Counterspell", "Death Denied", "Demonic Tutor", "Diabolic Tutor", "Disallow", "Disenchant", "Dualcaster Mage", "Eladamri's Call",
		"Essence Scatter", "Fabricate", "Fated Infatuation", "Generous Gift",
		"Idyllic Tutor", "Naturalize", "Negate", "Pongify", "Rapid Hybridization",
		"Rats' Feast", "Resculpt", "Return to the Ranks", "Reverberate", "Shattered Crypt", "Steelshaper's Gift", "Sylvan Scrying", "Terminate",
		"Twincast", "Vindicate", "Wildest Dreams":
		return true
	default:
		return false
	}
}
