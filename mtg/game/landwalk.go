package game

import "github.com/natefinch/council4/mtg/game/types"

// Reusable StaticAbilityBody templates for the landwalk evasion family (CR
// 702.14). Each typed variant keys off the defending player controlling a land
// of the named subtype; the generic variant keys off any land. Treat these
// values as immutable.
var (
	// PlainswalkStaticBody is the reusable StaticAbility for plainswalk.
	PlainswalkStaticBody = landwalkStaticBody("Plainswalk", types.Plains)

	// IslandwalkStaticBody is the reusable StaticAbility for islandwalk.
	IslandwalkStaticBody = landwalkStaticBody("Islandwalk", types.Island)

	// SwampwalkStaticBody is the reusable StaticAbility for swampwalk.
	SwampwalkStaticBody = landwalkStaticBody("Swampwalk", types.Swamp)

	// MountainwalkStaticBody is the reusable StaticAbility for mountainwalk.
	MountainwalkStaticBody = landwalkStaticBody("Mountainwalk", types.Mountain)

	// ForestwalkStaticBody is the reusable StaticAbility for forestwalk.
	ForestwalkStaticBody = landwalkStaticBody("Forestwalk", types.Forest)

	// DesertwalkStaticBody is the reusable StaticAbility for desertwalk.
	DesertwalkStaticBody = landwalkStaticBody("Desertwalk", types.Desert)

	// LandwalkStaticBody is the reusable StaticAbility for generic landwalk,
	// which keys off the defending player controlling any land.
	LandwalkStaticBody = StaticAbility{
		Text:             "Landwalk",
		KeywordAbilities: []KeywordAbility{LandwalkKeyword{AnyLand: true}},
	}

	// NonbasicLandwalkStaticBody is the reusable StaticAbility for nonbasic
	// landwalk, which keys off the defending player controlling a nonbasic land.
	NonbasicLandwalkStaticBody = StaticAbility{
		Text:             "Nonbasic landwalk",
		KeywordAbilities: []KeywordAbility{LandwalkKeyword{Nonbasic: true}},
	}
)

func landwalkStaticBody(text string, subtype types.Sub) StaticAbility {
	return StaticAbility{
		Text:             text,
		KeywordAbilities: []KeywordAbility{LandwalkKeyword{Subtype: subtype}},
	}
}
