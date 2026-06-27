package cardgen

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestLowerCastTriggerAcceptsPlayerPhrases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		phrase     string
		controller game.TriggerControllerFilter
	}{
		{"you cast", game.TriggerControllerYou},
		{"a player casts", game.TriggerControllerAny},
		{"an opponent casts", game.TriggerControllerOpponent},
	}
	for _, tc := range tests {
		t.Run(tc.phrase, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Bear",
				Layout:     "normal",
				TypeLine:   "Creature — Bear",
				OracleText: "Whenever " + tc.phrase + " a spell, draw a card.",
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
			}
			ta := face.TriggeredAbilities[0]
			if ta.Trigger.Pattern.Event != game.EventSpellCast {
				t.Errorf("event = %v, want EventSpellCast", ta.Trigger.Pattern.Event)
			}
			if ta.Trigger.Pattern.Controller != tc.controller {
				t.Errorf("controller = %v, want %v", ta.Trigger.Pattern.Controller, tc.controller)
			}
			if !ta.Trigger.Pattern.CardSelection.Empty() {
				t.Errorf("CardSelection = %+v, want empty for 'a spell'", ta.Trigger.Pattern.CardSelection)
			}
		})
	}
}

func TestLowerCastTriggerAcceptsSpellTypePhrases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		phrase    string
		wantTypes []types.Card
		wantAny   []types.Card
		wantExcl  []types.Card
	}{
		{"a creature spell", []types.Card{types.Creature}, nil, nil},
		{"a noncreature spell", nil, nil, []types.Card{types.Creature}},
		{"an instant or sorcery spell", nil, []types.Card{types.Instant, types.Sorcery}, nil},
		{"an instant spell", []types.Card{types.Instant}, nil, nil},
		{"an instant", []types.Card{types.Instant}, nil, nil},
		{"a sorcery spell", []types.Card{types.Sorcery}, nil, nil},
		{"an artifact spell", []types.Card{types.Artifact}, nil, nil},
		{"an enchantment spell", []types.Card{types.Enchantment}, nil, nil},
		{"a land spell", []types.Card{types.Land}, nil, nil},
		{"a planeswalker spell", []types.Card{types.Planeswalker}, nil, nil},
		{"a noncreature, nonland spell", nil, nil, []types.Card{types.Creature, types.Land}},
	}
	for _, tc := range tests {
		t.Run(tc.phrase, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Bear",
				Layout:     "normal",
				TypeLine:   "Creature — Bear",
				OracleText: "Whenever you cast " + tc.phrase + ", draw a card.",
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
			}
			sel := face.TriggeredAbilities[0].Trigger.Pattern.CardSelection
			if !slices.Equal(sel.RequiredTypes, tc.wantTypes) {
				t.Errorf("RequiredTypes = %v, want %v", sel.RequiredTypes, tc.wantTypes)
			}
			if !slices.Equal(sel.RequiredTypesAny, tc.wantAny) {
				t.Errorf("RequiredTypesAny = %v, want %v", sel.RequiredTypesAny, tc.wantAny)
			}
			if !slices.Equal(sel.ExcludedTypes, tc.wantExcl) {
				t.Errorf("ExcludedTypes = %v, want %v", sel.ExcludedTypes, tc.wantExcl)
			}
		})
	}
}

func TestLowerCastTriggerAcceptsColorPhrases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		phrase    string
		wantColor color.Color
	}{
		{"a white spell", color.White},
		{"a blue spell", color.Blue},
		{"a black spell", color.Black},
		{"a red spell", color.Red},
		{"a green spell", color.Green},
	}
	for _, tc := range tests {
		t.Run(tc.phrase, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Bear",
				Layout:     "normal",
				TypeLine:   "Creature — Bear",
				OracleText: "Whenever you cast " + tc.phrase + ", draw a card.",
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
			}
			sel := face.TriggeredAbilities[0].Trigger.Pattern.CardSelection
			if len(sel.ColorsAny) != 1 || sel.ColorsAny[0] != tc.wantColor {
				t.Errorf("ColorsAny = %v, want [%v]", sel.ColorsAny, tc.wantColor)
			}
		})
	}
}

func TestLowerCastTriggerAcceptsColorCardinalityPhrases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		phrase           string
		wantColorless    bool
		wantMulticolored bool
	}{
		{"a colorless spell", true, false},
		{"a multicolored spell", false, true},
	}
	for _, tc := range tests {
		t.Run(tc.phrase, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Bear",
				Layout:     "normal",
				TypeLine:   "Creature — Bear",
				OracleText: "Whenever you cast " + tc.phrase + ", draw a card.",
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
			}
			sel := face.TriggeredAbilities[0].Trigger.Pattern.CardSelection
			if sel.Colorless != tc.wantColorless {
				t.Errorf("Colorless = %v, want %v", sel.Colorless, tc.wantColorless)
			}
			if sel.Multicolored != tc.wantMulticolored {
				t.Errorf("Multicolored = %v, want %v", sel.Multicolored, tc.wantMulticolored)
			}
		})
	}
}

func TestLowerCastTriggerAcceptsSubtypeAndHistoricPhrases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		phrase string
		assert func(t *testing.T, pattern game.TriggerPattern)
	}{
		{
			name:   "Spirit or Arcane",
			phrase: "a Spirit or Arcane spell",
			assert: func(t *testing.T, pattern game.TriggerPattern) {
				t.Helper()
				if !slices.Equal(pattern.CardSelection.SubtypesAny, []types.Sub{types.Spirit, types.Arcane}) {
					t.Fatalf("SubtypesAny = %v, want Spirit or Arcane", pattern.CardSelection.SubtypesAny)
				}
			},
		},
		{
			name:   "single creature subtype",
			phrase: "an Elf spell",
			assert: func(t *testing.T, pattern game.TriggerPattern) {
				t.Helper()
				if !slices.Equal(pattern.CardSelection.SubtypesAny, []types.Sub{types.Elf}) {
					t.Fatalf("SubtypesAny = %v, want [Elf]", pattern.CardSelection.SubtypesAny)
				}
				if len(pattern.CardSelection.RequiredTypes) != 0 {
					t.Fatalf("RequiredTypes = %v, want none", pattern.CardSelection.RequiredTypes)
				}
			},
		},
		{
			name:   "historic",
			phrase: "a historic spell",
			assert: func(t *testing.T, pattern game.TriggerPattern) {
				t.Helper()
				if !pattern.RequireHistoric {
					t.Fatal("RequireHistoric = false, want true")
				}
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Bear",
				Layout:     "normal",
				TypeLine:   "Creature — Bear",
				OracleText: "Whenever you cast " + tc.phrase + ", draw a card.",
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
			}
			tc.assert(t, face.TriggeredAbilities[0].Trigger.Pattern)
		})
	}
}

func TestLowerCastTriggerAcceptsNotFromHandProvenance(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Pilferer",
		Layout:     "normal",
		TypeLine:   "Creature — Spirit Rogue",
		ManaCost:   "{1}{U}",
		OracleText: "Whenever an opponent casts a spell from anywhere other than their hand, draw a card.",
		Power:      new("1"),
		Toughness:  new("2"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	pattern := face.TriggeredAbilities[0].Trigger.Pattern
	if !pattern.ExcludeFromZone || pattern.FromZone != zone.Hand {
		t.Fatalf("from-zone exclusion = (%v, %v), want exclude hand", pattern.ExcludeFromZone, pattern.FromZone)
	}
	if pattern.MatchFromZone {
		t.Fatal("MatchFromZone = true, want false")
	}
}

func TestLowerCastTriggerAcceptsManaValueKickedAndZonePhrases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		phrase string
		assert func(t *testing.T, pattern game.TriggerPattern)
	}{
		{
			name:   "mana value",
			phrase: "a spell with mana value 5 or greater",
			assert: func(t *testing.T, pattern game.TriggerPattern) {
				t.Helper()
				mv := pattern.CardSelection.ManaValue
				if !mv.Exists || mv.Val.Op != compare.GreaterOrEqual || mv.Val.Value != 5 {
					t.Fatalf("ManaValue = %+v, want >= 5", mv)
				}
			},
		},
		{
			name:   "kicked",
			phrase: "a kicked spell",
			assert: func(t *testing.T, pattern game.TriggerPattern) {
				t.Helper()
				if !pattern.RequireKickerPaid {
					t.Fatal("RequireKickerPaid = false, want true")
				}
			},
		},
		{
			name:   "graveyard",
			phrase: "a spell from your graveyard",
			assert: func(t *testing.T, pattern game.TriggerPattern) {
				t.Helper()
				if !pattern.MatchFromZone || pattern.FromZone != zone.Graveyard {
					t.Fatalf("from-zone filter = (%v, %v), want graveyard", pattern.MatchFromZone, pattern.FromZone)
				}
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Bear",
				Layout:     "normal",
				TypeLine:   "Creature — Bear",
				OracleText: "Whenever you cast " + tc.phrase + ", draw a card.",
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
			}
			tc.assert(t, face.TriggeredAbilities[0].Trigger.Pattern)
		})
	}
}

func TestLowerCastTriggerAcceptsTypedManaValuePhrases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		phrase    string
		wantValue int
		assert    func(t *testing.T, sel game.Selection)
	}{
		{
			name:      "creature spell mana value",
			phrase:    "a creature spell with mana value 6 or greater",
			wantValue: 6,
			assert: func(t *testing.T, sel game.Selection) {
				t.Helper()
				if !slices.Equal(sel.RequiredTypes, []types.Card{types.Creature}) {
					t.Fatalf("RequiredTypes = %v, want [Creature]", sel.RequiredTypes)
				}
			},
		},
		{
			name:      "artifact spell mana value",
			phrase:    "an artifact spell with mana value 5 or greater",
			wantValue: 5,
			assert: func(t *testing.T, sel game.Selection) {
				t.Helper()
				if !slices.Equal(sel.RequiredTypes, []types.Card{types.Artifact}) {
					t.Fatalf("RequiredTypes = %v, want [Artifact]", sel.RequiredTypes)
				}
			},
		},
		{
			name:      "colorless spell mana value",
			phrase:    "a colorless spell with mana value 7 or greater",
			wantValue: 7,
			assert: func(t *testing.T, sel game.Selection) {
				t.Helper()
				if !sel.Colorless {
					t.Fatal("Colorless = false, want true")
				}
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Bear",
				Layout:     "normal",
				TypeLine:   "Creature — Bear",
				OracleText: "Whenever you cast " + tc.phrase + ", draw a card.",
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
			}
			pattern := face.TriggeredAbilities[0].Trigger.Pattern
			if pattern.Event != game.EventSpellCast {
				t.Fatalf("event = %v, want EventSpellCast", pattern.Event)
			}
			mv := pattern.CardSelection.ManaValue
			if !mv.Exists || mv.Val.Op != compare.GreaterOrEqual || mv.Val.Value != tc.wantValue {
				t.Fatalf("ManaValue = %+v, want >= %d", mv, tc.wantValue)
			}
			tc.assert(t, pattern.CardSelection)
		})
	}
}

func TestLowerCastTriggerAcceptsManaValueAtMostPhrases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		phrase    string
		wantValue int
		assert    func(t *testing.T, sel game.Selection)
	}{
		{
			name:      "spell or less",
			phrase:    "a spell with mana value 5 or less",
			wantValue: 5,
			assert:    func(t *testing.T, _ game.Selection) { t.Helper() },
		},
		{
			name:      "spell or fewer",
			phrase:    "a spell with mana value 4 or fewer",
			wantValue: 4,
			assert:    func(t *testing.T, _ game.Selection) { t.Helper() },
		},
		{
			name:      "creature spell or less",
			phrase:    "a creature spell with mana value 3 or less",
			wantValue: 3,
			assert: func(t *testing.T, sel game.Selection) {
				t.Helper()
				if !slices.Equal(sel.RequiredTypes, []types.Card{types.Creature}) {
					t.Fatalf("RequiredTypes = %v, want [Creature]", sel.RequiredTypes)
				}
			},
		},
		{
			name:      "colorless spell or less",
			phrase:    "a colorless spell with mana value 2 or less",
			wantValue: 2,
			assert: func(t *testing.T, sel game.Selection) {
				t.Helper()
				if !sel.Colorless {
					t.Fatal("Colorless = false, want true")
				}
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Bear",
				Layout:     "normal",
				TypeLine:   "Creature — Bear",
				OracleText: "Whenever you cast " + tc.phrase + ", draw a card.",
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
			}
			pattern := face.TriggeredAbilities[0].Trigger.Pattern
			if pattern.Event != game.EventSpellCast {
				t.Fatalf("event = %v, want EventSpellCast", pattern.Event)
			}
			mv := pattern.CardSelection.ManaValue
			if !mv.Exists || mv.Val.Op != compare.LessOrEqual || mv.Val.Value != tc.wantValue {
				t.Fatalf("ManaValue = %+v, want <= %d", mv, tc.wantValue)
			}
			tc.assert(t, pattern.CardSelection)
		})
	}
}

func TestLowerCastTriggerRejectsUnsupportedForms(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		oracle string
	}{
		{"self-cast Whenever", "Whenever you cast this spell, draw a card."},
		{"general TriggerWhen", "When you cast a spell, draw a card."},
		{"unrecognized player", "Whenever each player casts a spell, draw a card."},
		{"copy without cast", "Whenever you copy an instant or sorcery spell, draw a card."},
		{"opponent cast or copy", "Whenever an opponent casts or copies an instant or sorcery spell, draw a card."},
		{"ordinal spell beyond supported word", "Whenever you cast your sixth spell each turn, draw a card."},
		{"ordinal spell opponent", "Whenever an opponent casts your second spell each turn, draw a card."},
		{"unsupported mana value comparison", "Whenever you cast a spell with mana value less than 5, draw a card."},
		{"unsupported zone-filtered spell", "Whenever you cast a spell from your library, draw a card."},
		{"any player your graveyard", "Whenever a player casts a spell from your graveyard, draw a card."},
		{"opponent your graveyard", "Whenever an opponent casts a spell from your graveyard, draw a card."},
		{"unknown spell subtype noun", "Whenever you cast a frobnicate spell, draw a card."},
		{"subtype with trailing qualifier", "Whenever you cast a Goblin spell you control, draw a card."},
		{"unsupported body", "Whenever you cast a spell, counter target spell or ability that targets a creature."},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Test Bear",
				Layout:     "normal",
				TypeLine:   "Creature — Bear",
				OracleText: tc.oracle,
				Power:      new("2"),
				Toughness:  new("2"),
			}
			faces, diagnostics := lowerExecutableFaces(card)
			if len(diagnostics) == 0 {
				t.Fatalf("expected unsupported diagnostic for %q", tc.oracle)
			}
			if len(faces) > 0 && len(faces[0].TriggeredAbilities) > 0 {
				t.Fatalf("unexpected triggered ability for unsupported form %q", tc.oracle)
			}
		})
	}
}

func TestLowerCastTriggerAcceptsSelfCastSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "When you cast this spell, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	ta := face.TriggeredAbilities[0]
	if ta.Trigger.Type != game.TriggerWhen {
		t.Errorf("type = %v, want TriggerWhen", ta.Trigger.Type)
	}
	pattern := ta.Trigger.Pattern
	if pattern.Event != game.EventSpellCast {
		t.Errorf("event = %v, want EventSpellCast", pattern.Event)
	}
	if pattern.Source != game.TriggerSourceSelf {
		t.Errorf("source = %v, want TriggerSourceSelf", pattern.Source)
	}
	if pattern.Controller != game.TriggerControllerYou {
		t.Errorf("controller = %v, want TriggerControllerYou", pattern.Controller)
	}
	if !pattern.SelfWasCast {
		t.Error("SelfWasCast = false, want true")
	}
}

func TestLowerCastTriggerAcceptsOrdinalSpell(t *testing.T) {
	t.Parallel()
	tests := []struct {
		phrase      string
		wantOrdinal int
	}{
		{"your first spell each turn", 1},
		{"your second spell each turn", 2},
		{"your third spell each turn", 3},
	}
	for _, tc := range tests {
		t.Run(tc.phrase, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Bear",
				Layout:     "normal",
				TypeLine:   "Creature — Bear",
				OracleText: "Whenever you cast " + tc.phrase + ", draw a card.",
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
			}
			pattern := face.TriggeredAbilities[0].Trigger.Pattern
			if pattern.Event != game.EventSpellCast {
				t.Errorf("event = %v, want EventSpellCast", pattern.Event)
			}
			if pattern.Controller != game.TriggerControllerYou {
				t.Errorf("controller = %v, want TriggerControllerYou", pattern.Controller)
			}
			if pattern.PlayerEventOrdinalThisTurn != tc.wantOrdinal {
				t.Errorf("PlayerEventOrdinalThisTurn = %d, want %d", pattern.PlayerEventOrdinalThisTurn, tc.wantOrdinal)
			}
			if pattern.MatchSpellCopy {
				t.Error("MatchSpellCopy = true, want false for ordinal cast trigger")
			}
			if !pattern.CardSelection.Empty() {
				t.Errorf("CardSelection = %+v, want empty", pattern.CardSelection)
			}
		})
	}
}

func TestLowerCastTriggerAcceptsNonControllerOrdinalSpell(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		oracle         string
		wantController game.TriggerControllerFilter
		wantOrdinal    int
		wantFiltered   bool
	}{
		{
			name:           "any player second spell (Lotho)",
			oracle:         "Whenever a player casts their second spell each turn, draw a card.",
			wantController: game.TriggerControllerAny,
			wantOrdinal:    2,
		},
		{
			name:           "opponent third spell",
			oracle:         "Whenever an opponent casts their third spell each turn, draw a card.",
			wantController: game.TriggerControllerOpponent,
			wantOrdinal:    3,
		},
		{
			name:           "any player second filtered spell",
			oracle:         "Whenever a player casts their second instant spell each turn, draw a card.",
			wantController: game.TriggerControllerAny,
			wantOrdinal:    2,
			wantFiltered:   true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Bear",
				Layout:     "normal",
				TypeLine:   "Creature — Bear",
				OracleText: tc.oracle,
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
			}
			pattern := face.TriggeredAbilities[0].Trigger.Pattern
			if pattern.Event != game.EventSpellCast {
				t.Errorf("event = %v, want EventSpellCast", pattern.Event)
			}
			if pattern.Controller != tc.wantController {
				t.Errorf("controller = %v, want %v", pattern.Controller, tc.wantController)
			}
			if pattern.PlayerEventOrdinalThisTurn != tc.wantOrdinal {
				t.Errorf("PlayerEventOrdinalThisTurn = %d, want %d", pattern.PlayerEventOrdinalThisTurn, tc.wantOrdinal)
			}
			if pattern.CardSelection.Empty() == tc.wantFiltered {
				t.Errorf("CardSelection empty = %v, want filtered = %v", pattern.CardSelection.Empty(), tc.wantFiltered)
			}
		})
	}
}

func TestLowerCastTriggerAcceptsCastOrCopy(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Mage",
		Layout:     "normal",
		TypeLine:   "Creature — Human Wizard",
		OracleText: "Magecraft — Whenever you cast or copy an instant or sorcery spell, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	pattern := face.TriggeredAbilities[0].Trigger.Pattern
	if pattern.Event != game.EventSpellCast {
		t.Errorf("event = %v, want EventSpellCast", pattern.Event)
	}
	if !pattern.MatchSpellCopy {
		t.Error("MatchSpellCopy = false, want true")
	}
	if pattern.Controller != game.TriggerControllerYou {
		t.Errorf("controller = %v, want TriggerControllerYou", pattern.Controller)
	}
	if !slices.Equal(pattern.CardSelection.RequiredTypesAny, []types.Card{types.Instant, types.Sorcery}) {
		t.Errorf("RequiredTypesAny = %v, want [Instant Sorcery]", pattern.CardSelection.RequiredTypesAny)
	}
}

func TestLowerCastTriggerOptionalBody(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "Whenever you cast a creature spell, you may draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	ta := face.TriggeredAbilities[0]
	if ta.Trigger.Pattern.Event != game.EventSpellCast {
		t.Errorf("event = %v, want EventSpellCast", ta.Trigger.Pattern.Event)
	}
	if !ta.Optional {
		t.Error("expected optional triggered ability")
	}
}

func TestLowerCastTriggerOptionalControllerBenefitUnlessEventPlayerPays(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		oracle    string
		manaCost  cost.Mana
		assertPay func(*testing.T, game.Primitive)
	}{
		{
			name:     "draw",
			oracle:   "Whenever an opponent casts a spell, you may draw a card unless that player pays {1}.",
			manaCost: cost.Mana{cost.O(1)},
			assertPay: func(t *testing.T, primitive game.Primitive) {
				t.Helper()
				draw, ok := primitive.(game.Draw)
				if !ok || draw.Player != game.ControllerReference() {
					t.Fatalf("benefit = %#v, want controller draw", primitive)
				}
			},
		},
		{
			name:     "gain life",
			oracle:   "Whenever an opponent casts a spell, you may gain 2 life unless that player pays {3}.",
			manaCost: cost.Mana{cost.O(3)},
			assertPay: func(t *testing.T, primitive game.Primitive) {
				t.Helper()
				gain, ok := primitive.(game.GainLife)
				if !ok || gain.Player != game.ControllerReference() {
					t.Fatalf("benefit = %#v, want controller life gain", primitive)
				}
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Tax Study",
				Layout:     "normal",
				TypeLine:   "Enchantment",
				OracleText: tc.oracle,
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
			}
			trigger := face.TriggeredAbilities[0]
			if trigger.Optional {
				t.Fatal("trigger Optional = true, want payment before the controller's optional benefit")
			}
			if trigger.Trigger.Pattern.Controller != game.TriggerControllerOpponent {
				t.Fatalf("trigger controller = %v, want opponent", trigger.Trigger.Pattern.Controller)
			}
			mode := trigger.Content.Modes[0]
			if len(mode.Targets) != 0 || len(mode.Sequence) != 2 {
				t.Fatalf("mode = %#v, want targetless two-instruction sequence", mode)
			}
			pay, ok := mode.Sequence[0].Primitive.(game.Pay)
			if !ok {
				t.Fatalf("instruction 0 = %T, want game.Pay", mode.Sequence[0].Primitive)
			}
			if payer, ok := pay.Payment.Payer.Val, pay.Payment.Payer.Exists; !ok || payer != game.EventPlayerReference() {
				t.Fatalf("payer = %#v, want event player", pay.Payment.Payer)
			}
			if !pay.Payment.ManaCost.Exists || !slices.Equal(pay.Payment.ManaCost.Val, tc.manaCost) {
				t.Fatalf("mana cost = %#v, want %v", pay.Payment.ManaCost, tc.manaCost)
			}
			if mode.Sequence[0].PublishResult != unlessPaidResultKey {
				t.Fatalf("payment result key = %q", mode.Sequence[0].PublishResult)
			}
			benefit := mode.Sequence[1]
			if !benefit.Optional ||
				!benefit.ResultGate.Exists ||
				benefit.ResultGate.Val.Key != unlessPaidResultKey ||
				benefit.ResultGate.Val.Succeeded != game.TriFalse {
				t.Fatalf("benefit envelope = %#v", benefit)
			}
			tc.assertPay(t, benefit.Primitive)
		})
	}
}

func TestLowerCastTriggerOptionalControllerBenefitUnlessEventPlayerPaysRejectsUnsafeForms(t *testing.T) {
	t.Parallel()
	tests := []string{
		"Whenever an opponent casts a spell, you may draw a card unless that player pays 2 life.",
		"Whenever an opponent casts a spell, you may draw a card unless you pay {1}.",
		"Whenever an opponent casts a spell, you may draw a card unless that player pays {X}.",
		"Whenever an opponent casts a spell, you may draw a card and gain 1 life unless that player pays {1}.",
		"Whenever an opponent casts a spell, target player may draw a card unless that player pays {1}.",
		"Whenever this creature attacks, you may draw a card unless that player pays {1}.",
	}
	for _, oracle := range tests {
		t.Run(oracle, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Unsafe Tax",
				Layout:     "normal",
				TypeLine:   "Enchantment",
				OracleText: oracle,
			}
			faces, diagnostics := lowerExecutableFaces(card)
			if len(diagnostics) == 0 {
				t.Fatalf("expected unsupported diagnostic for %q", oracle)
			}
			if len(faces) > 0 && len(faces[0].TriggeredAbilities) > 0 {
				t.Fatalf("unexpected supported trigger for %q", oracle)
			}
		})
	}
}

func TestLowerCastTriggerBindsThatPlayerToEventActor(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		oracle string
		assert func(*testing.T, game.Primitive)
	}{
		{
			name:   "discard",
			oracle: "Whenever a player casts a spell, that player discards a card.",
			assert: func(t *testing.T, primitive game.Primitive) {
				t.Helper()
				discard, ok := primitive.(game.Discard)
				if !ok || discard.Player != game.EventPlayerReference() {
					t.Fatalf("primitive = %#v, want event-player discard", primitive)
				}
			},
		},
		{
			name:   "mill",
			oracle: "Whenever an opponent casts a spell, that player mills two cards.",
			assert: func(t *testing.T, primitive game.Primitive) {
				t.Helper()
				mill, ok := primitive.(game.Mill)
				if !ok || mill.Player != game.EventPlayerReference() {
					t.Fatalf("primitive = %#v, want event-player mill", primitive)
				}
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Event Actor Study",
				Layout:     "normal",
				TypeLine:   "Enchantment",
				OracleText: tc.oracle,
			})
			sequence := face.TriggeredAbilities[0].Content.Modes[0].Sequence
			if len(sequence) != 1 {
				t.Fatalf("sequence = %#v, want one instruction", sequence)
			}
			tc.assert(t, sequence[0].Primitive)
		})
	}
}

// TestLowerCastTriggerTrailingOptionalLifeGain verifies a cast-trigger body with
// a mandatory lead effect and a trailing resolving optional ("draw a card. You
// may gain 1 life.") lowers with the trigger firing unconditionally and only the
// trailing gain-life instruction marked Optional. This exercises the now-exact
// optional life effect routed through the bare trailing-optional flow.
func TestLowerCastTriggerTrailingOptionalLifeGain(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "Whenever you cast a spell, draw a card. You may gain 1 life.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	ta := face.TriggeredAbilities[0]
	if ta.Optional {
		t.Error("triggered ability Optional = true, want false (trigger fires unconditionally)")
	}
	seq := ta.Content.Modes[0].Sequence
	if len(seq) != 2 {
		t.Fatalf("sequence length = %d, want 2", len(seq))
	}
	if _, ok := seq[0].Primitive.(game.Draw); !ok {
		t.Errorf("instruction 0 primitive = %T, want game.Draw", seq[0].Primitive)
	}
	if seq[0].Optional {
		t.Error("draw instruction Optional = true, want false (mandatory lead effect)")
	}
	if _, ok := seq[1].Primitive.(game.GainLife); !ok {
		t.Errorf("instruction 1 primitive = %T, want game.GainLife", seq[1].Primitive)
	}
	if !seq[1].Optional {
		t.Error("gain-life instruction Optional = false, want true (trailing optional)")
	}
}

func TestLowerCastTriggerSupportedInterveningCondition(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		oracle    string
		wantField string
	}{
		{
			name:      "if you control an artifact",
			oracle:    "Whenever you cast a spell, if you control an artifact, draw a card.",
			wantField: "ControlsMatching",
		},
		{
			name:      "if you have 5 or more life",
			oracle:    "Whenever you cast a creature spell, if you have 5 or more life, draw a card.",
			wantField: "Aggregates",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Bear",
				Layout:     "normal",
				TypeLine:   "Creature — Bear",
				OracleText: tc.oracle,
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
			}
			trigger := face.TriggeredAbilities[0].Trigger
			if trigger.Pattern.Event != game.EventSpellCast {
				t.Errorf("event = %v, want EventSpellCast", trigger.Pattern.Event)
			}
			if trigger.InterveningIf == "" || !trigger.InterveningCondition.Exists {
				t.Fatalf("trigger = %+v, want intervening condition", trigger)
			}
		})
	}
}

func TestLowerCastTriggerInterveningIfFailsClosedOnUnsupportedCondition(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "Whenever you cast a spell, if you have seven or more cards in hand, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	}, "c")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("cast trigger with unsupported intervening condition unexpectedly lowered")
	}
	if !strings.Contains(diagnostics[0].Detail, "condition") {
		t.Fatalf("diagnostic = %#v, want condition detail", diagnostics[0])
	}
}
