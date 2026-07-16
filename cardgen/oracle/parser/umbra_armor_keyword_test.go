package parser

import "testing"

func TestParseUmbraArmorKeywordNames(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Umbra armor (If enchanted creature would be destroyed, instead remove all damage from it and destroy this Aura.)",
		"Totem armor (If enchanted creature would be destroyed, instead remove all damage from it and destroy this Aura.)",
	} {
		keywords := keywordsFor(t, source)
		if len(keywords) != 1 {
			t.Fatalf("%q keywords = %+v, want one", source, keywords)
		}
		if got := keywords[0]; got.Kind != KeywordUmbraArmor ||
			got.Parameter.Kind != KeywordParameterNone {
			t.Fatalf("%q keyword = %+v, want Umbra armor with no parameter", source, got)
		}
	}
}
