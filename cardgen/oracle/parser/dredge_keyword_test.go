package parser

import "testing"

func TestParseDredgeKeyword(t *testing.T) {
	t.Parallel()
	source := "Dredge 3 (If you would draw a card, you may mill three cards instead. " +
		"If you do, return this card from your graveyard to your hand.)"
	keywords := keywordsFor(t, source)
	if len(keywords) != 1 {
		t.Fatalf("keywords = %+v; want exactly one", keywords)
	}
	got := keywords[0]
	if got.Kind != KeywordDredge {
		t.Fatalf("kind = %v; want %v", got.Kind, KeywordDredge)
	}
	if got.Parameter.Kind != KeywordParameterInteger || got.Parameter.Integer() != 3 {
		t.Fatalf("parameter = %+v; want integer 3", got.Parameter)
	}
}

func TestParseDredgeKeywordVariants(t *testing.T) {
	t.Parallel()
	for n, source := range map[int]string{
		1: "Dredge 1", 2: "Dredge 2", 6: "Dredge 6",
	} {
		keywords := keywordsFor(t, source)
		if len(keywords) != 1 || keywords[0].Kind != KeywordDredge {
			t.Fatalf("%q keywords = %+v; want one Dredge", source, keywords)
		}
		if keywords[0].Parameter.Integer() != n {
			t.Fatalf("%q integer = %d; want %d", source, keywords[0].Parameter.Integer(), n)
		}
	}
}
