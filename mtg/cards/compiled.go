package cards

import (
	"slices"

	"github.com/natefinch/council4/mtg/cards/cardset"
)

// CompilerExclusion identifies a checked-in generated definition with a proven
// runtime blocker. Compiler acceptance is not permission to admit partial cards.
type CompilerExclusion struct {
	OracleID string
	Name     string
	Blocker  string
}

// CompilerExclusions is separate from generated sources and Engine behavior.
// Remove an exclusion only after its owning runtime regression passes.
func CompilerExclusions() []CompilerExclusion {
	return []CompilerExclusion{
		{OracleID: "82004860-e589-4e38-8d61-8c0210e4ea39", Name: "Demonic Tutor", Blocker: "https://github.com/natefinch/council4/issues/3289"},
		{OracleID: "14589b6b-1814-46f9-a364-83cc15dacac2", Name: "Diabolic Tutor", Blocker: "https://github.com/natefinch/council4/issues/3289"},
	}
}

func installedCardSets() [][]cardset.Entry {
	sets := defaultCardSets()
	exclusions := CompilerExclusions()
	for index, set := range sets {
		sets[index] = slices.DeleteFunc(slices.Clone(set), func(entry cardset.Entry) bool {
			return slices.ContainsFunc(exclusions, func(exclusion CompilerExclusion) bool {
				return entry.Name == exclusion.Name
			})
		})
	}
	return sets
}
