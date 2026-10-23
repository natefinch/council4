// Package authored owns non-generated Card Implementations.
// Legacy corpus generation only writes letter packages, never this directory.
package authored

import "github.com/natefinch/council4/mtg/cards/cardset"

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
	entries = append(entries, GhostVacuum, NecromanticSelection, ToothCollector, MarchesaDealerOfDeath)
	entries = append(entries, MyrCustodian)
	entries = append(entries, Mundungu, SlyInstigator, HalfElfMonk)
	entries = append(entries, EagerConstruct, SeaGodsRevenge, DevoutDecree)
	for i := range entries {
		entries[i].ReplaceLegacy = replacesCompiledEntry(entries[i].Name)
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
