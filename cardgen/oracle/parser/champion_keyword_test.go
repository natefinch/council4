package parser

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
)

func TestParseChampionKeywordSubtype(t *testing.T) {
	t.Parallel()
	keywords := keywordsFor(t, "Champion a Goblin")
	if len(keywords) != 1 {
		t.Fatalf("keywords = %+v; want one", keywords)
	}
	if keywords[0].Kind != KeywordChampion || keywords[0].Parameter.Kind != KeywordParameterChampion {
		t.Fatalf("champion = %+v", keywords[0])
	}
	if !slices.Equal(keywords[0].Parameter.EnchantTarget().Subtypes, []types.Sub{types.Goblin}) {
		t.Fatalf("champion subtypes = %+v", keywords[0].Parameter.EnchantTarget().Subtypes)
	}
}

func TestParseChampionKeywordCreature(t *testing.T) {
	t.Parallel()
	keywords := keywordsFor(t, "Champion a creature")
	if len(keywords) != 1 || keywords[0].Kind != KeywordChampion ||
		keywords[0].Parameter.Kind != KeywordParameterChampion ||
		!slices.Equal(keywords[0].Parameter.EnchantTarget().CardTypes, []CardType{CardTypeCreature}) {
		t.Fatalf("champion = %+v", keywords)
	}
}
