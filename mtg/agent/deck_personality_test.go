package agent

import "testing"

func TestDeckPersonalityMatchesArchetype(t *testing.T) {
	cases := []struct {
		archetype  Archetype
		aggressive bool
		political  bool
	}{
		{ArchetypeAggro, true, false},
		{ArchetypeTokens, true, false},
		{ArchetypeAristocrats, true, false},
		{ArchetypeControl, false, true},
		{ArchetypeRamp, false, true},
		{ArchetypeMidrange, false, false},
	}
	for _, c := range cases {
		got := DeckPersonality(DeckProfile{Archetype: c.archetype})
		if (got.Aggression > 0) != c.aggressive {
			t.Errorf("archetype %v aggression = %v, want aggressive=%v", c.archetype, got.Aggression, c.aggressive)
		}
		if (got.PoliticsWeight > 0) != c.political {
			t.Errorf("archetype %v politics = %v, want political=%v", c.archetype, got.PoliticsWeight, c.political)
		}
	}
	if DeckPersonality(DeckProfile{Archetype: ArchetypeMidrange}) != (Personality{}) {
		t.Error("midrange should map to the neutral zero personality")
	}
}
