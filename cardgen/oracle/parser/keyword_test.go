package parser

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

func keywordsFor(t *testing.T, source string) []Keyword {
	t.Helper()
	return atomsFor(t, source, "").Keywords()
}

func TestParseKeywordVocabularyMeaning(t *testing.T) {
	t.Parallel()
	tests := map[string]KeywordKind{
		"Affinity": KeywordAffinity, "Annihilator": KeywordAnnihilator, "Cascade": KeywordCascade,
		"Companion": KeywordCompanion, "Convoke": KeywordConvoke, "Cumulative upkeep": KeywordCumulativeUpkeep, "Cycling": KeywordCycling,
		"Deathtouch": KeywordDeathtouch, "Defender": KeywordDefender, "Delve": KeywordDelve,
		"Devoid": KeywordDevoid, "Disguise": KeywordDisguise, "Double strike": KeywordDoubleStrike,
		"Emerge": KeywordEmerge, "Enchant": KeywordEnchant, "Equip": KeywordEquip, "Escape": KeywordEscape,
		"Eternalize": KeywordEternalize, "Embalm": KeywordEmbalm, "Evolve": KeywordEvolve, "Exalted": KeywordExalted, "Fear": KeywordFear, "First strike": KeywordFirstStrike,
		"Flash": KeywordFlash, "Flashback": KeywordFlashback, "Flying": KeywordFlying, "Foretell": KeywordForetell,
		"Haste": KeywordHaste, "Hexproof": KeywordHexproof, "Improvise": KeywordImprovise,
		"Horsemanship":   KeywordHorsemanship,
		"Indestructible": KeywordIndestructible, "Infect": KeywordInfect, "Intimidate": KeywordIntimidate, "Kicker": KeywordKicker,
		"Lifelink": KeywordLifelink, "Madness": KeywordMadness, "Menace": KeywordMenace, "Morph": KeywordMorph,
		"Mutate": KeywordMutate, "Ninjutsu": KeywordNinjutsu, "Persist": KeywordPersist,
		"Protection": KeywordProtection, "Prowess": KeywordProwess, "Read ahead": KeywordReadAhead,
		"Reach": KeywordReach, "Shroud": KeywordShroud, "Skulk": KeywordSkulk, "Split second": KeywordSplitSecond, "Storm": KeywordStorm,
		"Suspend": KeywordSuspend, "Toxic": KeywordToxic, "Trample": KeywordTrample, "Undying": KeywordUndying,
		"Vigilance": KeywordVigilance, "Ward": KeywordWard, "Wither": KeywordWither, "Riot": KeywordRiot,
		"Landcycling": KeywordLandcycling, "Basic landcycling": KeywordBasicLandcycling,
		"Plainscycling": KeywordPlainscycling, "Islandcycling": KeywordIslandcycling,
		"Swampcycling": KeywordSwampcycling, "Mountaincycling": KeywordMountaincycling,
		"Forestcycling": KeywordForestcycling,
	}
	for source, want := range tests {
		keywords := keywordsFor(t, source)
		if len(keywords) != 1 || keywords[0].Kind != want || keywords[0].Kind.String() != source {
			t.Errorf("%q keywords = %+v; want %v", source, keywords, want)
		}
	}
}

func TestParseKeywordParameterComposition(t *testing.T) {
	t.Parallel()
	keywords := keywordsFor(t, "Flying, ward {2}{U}, toxic 3, enchant creature, protection from black and from red, cycling {X}")
	if len(keywords) != 6 {
		t.Fatalf("keywords = %+v; want six", keywords)
	}
	if keywords[0].Kind != KeywordFlying || keywords[0].Parameter.Kind != KeywordParameterNone {
		t.Fatalf("flying = %+v", keywords[0])
	}
	if keywords[1].Kind != KeywordWard ||
		keywords[1].Parameter.Kind != KeywordParameterManaCost ||
		!slices.Equal(keywords[1].Parameter.ManaCost(), cost.Mana{cost.O(2), cost.U}) {
		t.Fatalf("ward = %+v, mana=%+v", keywords[1], keywords[1].Parameter.ManaCost())
	}
	if keywords[2].Kind != KeywordToxic ||
		keywords[2].Parameter.Kind != KeywordParameterInteger ||
		keywords[2].Parameter.Integer() != 3 {
		t.Fatalf("toxic = %+v", keywords[2])
	}
	if keywords[3].Kind != KeywordEnchant ||
		keywords[3].Parameter.Kind != KeywordParameterEnchantTarget ||
		!slices.Equal(keywords[3].Parameter.EnchantTarget().CardTypes, []CardType{CardTypeCreature}) {
		t.Fatalf("enchant = %+v", keywords[3])
	}
	protection := keywords[4].Parameter.Protection()
	if keywords[4].Kind != KeywordProtection ||
		keywords[4].Parameter.Kind != KeywordParameterProtection ||
		!slices.Equal(protection.FromColors, []Color{ColorBlack, ColorRed}) {
		t.Fatalf("protection = %+v, predicate=%+v", keywords[4], protection)
	}
	if keywords[5].Kind != KeywordCycling ||
		keywords[5].Parameter.Kind != KeywordParameterManaCost ||
		!slices.Equal(keywords[5].Parameter.ManaCost(), cost.Mana{cost.X}) {
		t.Fatalf("cycling = %+v", keywords[5])
	}
}

func TestParseFlashbackManaCost(t *testing.T) {
	t.Parallel()
	keywords := keywordsFor(t, "Flashback {2}{R}")
	if len(keywords) != 1 ||
		keywords[0].Kind != KeywordFlashback ||
		keywords[0].Parameter.Kind != KeywordParameterManaCost ||
		!slices.Equal(keywords[0].Parameter.ManaCost(), cost.Mana{cost.O(2), cost.R}) {
		t.Fatalf("flashback keywords = %+v", keywords)
	}
}

func TestParseCumulativeUpkeepManaCostAndSpans(t *testing.T) {
	t.Parallel()
	keywords := keywordsFor(t, "Cumulative upkeep {1}{U}")
	if len(keywords) != 1 {
		t.Fatalf("keywords = %+v; want one", keywords)
	}
	keyword := keywords[0]
	if keyword.Kind != KeywordCumulativeUpkeep ||
		keyword.Kind.String() != "Cumulative upkeep" ||
		keyword.Parameter.Kind != KeywordParameterManaCost ||
		!slices.Equal(keyword.Parameter.ManaCost(), cost.Mana{cost.O(1), cost.U}) {
		t.Fatalf("cumulative upkeep = %+v, mana=%+v", keyword, keyword.Parameter.ManaCost())
	}
	if got, want := keyword.NameSpan, (shared.Span{
		Start: shared.Position{Offset: 0, Line: 1, Column: 1},
		End:   shared.Position{Offset: 17, Line: 1, Column: 18},
	}); got != want {
		t.Fatalf("name span = %+v; want %+v", got, want)
	}
	if got, want := keyword.Parameter.Span, (shared.Span{
		Start: shared.Position{Offset: 18, Line: 1, Column: 19},
		End:   shared.Position{Offset: 24, Line: 1, Column: 25},
	}); got != want {
		t.Fatalf("parameter span = %+v; want %+v", got, want)
	}
	if got, want := keyword.Span, (shared.Span{
		Start: shared.Position{Offset: 0, Line: 1, Column: 1},
		End:   shared.Position{Offset: 24, Line: 1, Column: 25},
	}); got != want {
		t.Fatalf("keyword span = %+v; want %+v", got, want)
	}
}

func TestParseProtectionParameterFamilies(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		check  func(ProtectionParameter) bool
	}{
		{source: "Protection from everything", check: func(p ProtectionParameter) bool { return p.Everything }},
		{source: "Protection from each color", check: func(p ProtectionParameter) bool { return p.EachColor }},
		{source: "Protection from all colors", check: func(p ProtectionParameter) bool { return p.EachColor }},
		{source: "Protection from multicolored", check: func(p ProtectionParameter) bool { return p.Multicolored }},
		{source: "Protection from monocolored", check: func(p ProtectionParameter) bool { return p.Monocolored }},
		{source: "Protection from artifacts", check: func(p ProtectionParameter) bool {
			return slices.Equal(p.FromTypes, []CardType{CardTypeArtifact})
		}},
		{source: "Protection from Dragons", check: func(p ProtectionParameter) bool {
			return slices.Equal(p.FromSubtypes, []types.Sub{types.Dragon})
		}},
	}
	for _, test := range tests {
		keywords := keywordsFor(t, test.source)
		if len(keywords) != 1 ||
			keywords[0].Parameter.Kind != KeywordParameterProtection ||
			!test.check(keywords[0].Parameter.Protection()) {
			t.Errorf("%q keywords = %+v, protection=%+v", test.source, keywords, keywords[0].Parameter.Protection())
		}
	}
}

func TestParseEquipRestriction(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source     string
		supertypes []Supertype
		subtypes   []types.Sub
	}{
		{source: "Equip legendary creature {3}", supertypes: []Supertype{SupertypeLegendary}},
		{source: "Equip Knight {1}", subtypes: []types.Sub{types.Knight}},
		{source: "Equip Shaman, Warlock, or Wizard {2}", subtypes: []types.Sub{types.Shaman, types.Warlock, types.Wizard}},
	}
	for _, test := range tests {
		keywords := keywordsFor(t, test.source)
		if len(keywords) != 1 || keywords[0].Kind != KeywordEquip {
			t.Fatalf("%q keywords = %+v", test.source, keywords)
		}
		restriction := keywords[0].EquipRestriction
		if restriction == nil {
			t.Fatalf("%q has nil EquipRestriction", test.source)
		}
		if !slices.Equal(restriction.Supertypes, test.supertypes) {
			t.Errorf("%q supertypes = %v; want %v", test.source, restriction.Supertypes, test.supertypes)
		}
		if !slices.Equal(restriction.Subtypes, test.subtypes) {
			t.Errorf("%q subtypes = %v; want %v", test.source, restriction.Subtypes, test.subtypes)
		}
	}
}

func TestParseEquipRestrictionFailsClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Equip commander {3}",
		"Equip planeswalker {5}",
		"Equip {2}",
	} {
		keywords := keywordsFor(t, source)
		if len(keywords) != 1 || keywords[0].Kind != KeywordEquip {
			t.Fatalf("%q keywords = %+v", source, keywords)
		}
		if keywords[0].EquipRestriction != nil {
			t.Errorf("%q EquipRestriction = %+v; want nil", source, keywords[0].EquipRestriction)
		}
	}
}

func TestParseKeywordSelectorsCompose(t *testing.T) {
	t.Parallel()
	atoms := atomsFor(t, "cards with cycling, creatures with a flying ability, and creatures without shadow", "")
	selectors := atoms.KeywordSelectors()
	if len(selectors) != 3 {
		t.Fatalf("selectors = %+v; want three", selectors)
	}
	if selectors[0].Keyword != KeywordCycling || selectors[0].Form != KeywordSelectorFormDirect || selectors[0].Excluded ||
		selectors[1].Keyword != KeywordFlying || selectors[1].Form != KeywordSelectorFormAbility || selectors[1].Excluded ||
		selectors[2].Keyword != KeywordShadow || selectors[2].Form != KeywordSelectorFormDirect || !selectors[2].Excluded {
		t.Fatalf("selectors = %+v", selectors)
	}
}

func TestParseKeywordNearMissesFailClosed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source        string
		wantKind      KeywordKind
		wantParameter KeywordParameterKind
	}{
		{source: "First striker"},
		{source: "Read a head"},
		{source: "nonflying"},
		{source: "Enchant creatures", wantKind: KeywordEnchant},
		{source: "Ward {T}", wantKind: KeywordWard},
		{source: "Protection against red", wantKind: KeywordProtection},
		{source: "Protection from each colors", wantKind: KeywordProtection},
		{source: "Protection from red and from artifacts", wantKind: KeywordProtection},
	}
	for _, test := range tests {
		keywords := keywordsFor(t, test.source)
		if test.wantKind == KeywordUnknown {
			if len(keywords) != 0 {
				t.Errorf("%q keywords = %+v; want none", test.source, keywords)
			}
			continue
		}
		if len(keywords) != 1 ||
			keywords[0].Kind != test.wantKind ||
			keywords[0].Parameter.Kind != test.wantParameter {
			t.Errorf("%q keywords = %+v; want %v with no parameter", test.source, keywords, test.wantKind)
		}
	}
	if selectors := atomsFor(t, "creature with cyclings", "").KeywordSelectors(); len(selectors) != 0 {
		t.Errorf("near-miss selectors = %+v; want none", selectors)
	}
}

func TestKeywordManaSymbolFamilies(t *testing.T) {
	t.Parallel()
	keywords := keywordsFor(t, "Ward {W}{U/B}{2/R}{G/P}{C}{S}{10}")
	want := cost.Mana{
		cost.W,
		cost.HybridMana(mana.U, mana.B),
		cost.Twobrid(mana.R),
		cost.PhyrexianMana(mana.G),
		cost.C,
		cost.S,
		cost.O(10),
	}
	if len(keywords) != 1 || !slices.Equal(keywords[0].Parameter.ManaCost(), want) {
		t.Fatalf("mana = %+v; want %+v", keywords[0].Parameter.ManaCost(), want)
	}
}

func TestParseLandcyclingKeywords(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		want   KeywordKind
	}{
		{"Basic landcycling {1}", KeywordBasicLandcycling},
		{"Landcycling {2}", KeywordLandcycling},
		{"Plainscycling {1}{W}", KeywordPlainscycling},
		{"Forestcycling {2}", KeywordForestcycling},
	}
	for _, test := range tests {
		keywords := keywordsFor(t, test.source)
		if len(keywords) != 1 {
			t.Fatalf("%q keywords = %+v; want one", test.source, keywords)
		}
		if keywords[0].Kind != test.want {
			t.Errorf("%q kind = %v; want %v", test.source, keywords[0].Kind, test.want)
		}
		if keywords[0].Parameter.Kind != KeywordParameterManaCost ||
			len(keywords[0].Parameter.ManaCost()) == 0 {
			t.Errorf("%q parameter = %+v; want mana cost", test.source, keywords[0].Parameter)
		}
	}
}

func TestExpandBushidoKeyword(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		source string
		rank   int
	}{
		{"jade avenger", "Bushido 2 (Whenever this creature blocks or becomes blocked, it gets +2/+2 until end of turn.)", 2},
		{"nezumi ronin", "Bushido 1 (Whenever this creature blocks or becomes blocked, it gets +1/+1 until end of turn.)", 1},
		{"bare keyword", "Bushido 3", 3},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			trigger := parseTriggerEventFromSource(t, test.source, "Jade Avenger")
			if trigger == nil {
				t.Fatal("trigger = nil")
			}
			if trigger.Kind != TriggerEventKindBlock ||
				trigger.UnionKind != TriggerEventKindBecameBlocked ||
				trigger.Subject.Kind != TriggerEventSubjectSelf {
				t.Fatalf("trigger = %#v", trigger)
			}
		})
	}
}

func TestExpandBushidoKeywordLeavesOtherTextAlone(t *testing.T) {
	t.Parallel()
	// A line that only mentions the word elsewhere must not be rewritten.
	if got := expandBushidoKeyword("Whenever Bushido blocks, draw a card."); got != "Whenever Bushido blocks, draw a card." {
		t.Fatalf("rewrote unrelated line: %q", got)
	}
	if got := expandBushidoKeyword("Bushido"); got != "Bushido" {
		t.Fatalf("rewrote rankless keyword: %q", got)
	}
}
