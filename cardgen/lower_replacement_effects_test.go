package cardgen

import (
	goparser "go/parser"
	"go/token"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestLowerEntersTappedReplacement(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Land",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "This land enters tapped.",
	})
	if len(face.ReplacementAbilities) != 1 {
		t.Fatalf("got %d replacement abilities, want 1", len(face.ReplacementAbilities))
	}
	if !face.ReplacementAbilities[0].Replacement.EntersTapped {
		t.Fatal("replacement is not an enters-tapped replacement")
	}
}

func TestLowerGroupEntersTappedReplacement(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracle     string
		controller game.TriggerControllerFilter
		cardTypes  []types.Card
	}{
		{
			name:       "opponent creatures",
			oracle:     "Creatures your opponents control enter tapped.",
			controller: game.TriggerControllerOpponent,
			cardTypes:  []types.Card{types.Creature},
		},
		{
			name:       "multi-type opponent",
			oracle:     "Artifacts, creatures, and lands your opponents control enter the battlefield tapped.",
			controller: game.TriggerControllerOpponent,
			cardTypes:  []types.Card{types.Artifact, types.Creature, types.Land},
		},
		{
			name:       "all permanents",
			oracle:     "Permanents enter the battlefield tapped.",
			controller: game.TriggerControllerAny,
			cardTypes:  nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Enchantment",
				Layout:     "normal",
				TypeLine:   "Enchantment",
				OracleText: test.oracle,
			})
			if len(face.ReplacementAbilities) != 1 {
				t.Fatalf("got %d replacement abilities, want 1", len(face.ReplacementAbilities))
			}
			replacement := face.ReplacementAbilities[0].Replacement
			if !replacement.EntersTapped || !replacement.EntersTappedOthers {
				t.Fatalf("replacement is not a group enters-tapped replacement: %#v", replacement)
			}
			if replacement.ControllerFilter != test.controller {
				t.Fatalf("controller filter = %v, want %v", replacement.ControllerFilter, test.controller)
			}
			if !slices.Equal(replacement.EntersTappedTypes, test.cardTypes) {
				t.Fatalf("types = %v, want %v", replacement.EntersTappedTypes, test.cardTypes)
			}
		})
	}
}

func TestLowerEntersTappedReplacementCardNamePhrasing(t *testing.T) {
	t.Parallel()
	// Card-name entry phrasing ("<name> enters tapped.") must lower through the
	// typed EntersTappedSelf flag, not a fixed whitelist of subject nouns.
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Timeless Lotus",
		Layout:     "normal",
		TypeLine:   "Legendary Artifact",
		OracleText: "Timeless Lotus enters tapped.",
	})
	if len(face.ReplacementAbilities) != 1 || !face.ReplacementAbilities[0].Replacement.EntersTapped {
		t.Fatalf("expected enters-tapped replacement, got %#v", face.ReplacementAbilities)
	}
}

func TestLowerAsEntersChoiceIsNotEntersTapped(t *testing.T) {
	t.Parallel()
	// "As ~ enters, choose ..." shares the enters verb with a plain tapped entry
	// but is a different construct; lowering must fail closed rather than mistake
	// it for an enters-tapped replacement.
	faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Siege",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "As this enchantment enters, choose Khans or Dragons.",
	})
	if len(diagnostics) == 0 {
		t.Fatal("expected fail-closed diagnostics, got none")
	}
	for _, face := range faces {
		for _, replacement := range face.ReplacementAbilities {
			if replacement.Replacement.EntersTapped {
				t.Fatal("As-enters-choose was mistakenly lowered as enters-tapped")
			}
		}
	}
}

func TestLowerEntryColorChoiceReplacement(t *testing.T) {
	t.Parallel()
	// "As this <permanent> enters, choose a color." plus "{T}: Add one mana of
	// the chosen color." must lower to an entry-time color-choice replacement and
	// a mana ability that reads the stored choice.
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Sol Grail",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "As this artifact enters, choose a color.\n{T}: Add one mana of the chosen color.",
	})
	if len(face.ReplacementAbilities) != 1 {
		t.Fatalf("got %d replacement abilities, want 1", len(face.ReplacementAbilities))
	}
	replacement := face.ReplacementAbilities[0].Replacement
	if !replacement.EntryColorChoice {
		t.Fatalf("replacement is not an entry color-choice replacement: %+v", replacement)
	}
	if replacement.EntersTapped {
		t.Fatalf("standalone choose-a-color must not enter tapped: %+v", replacement)
	}
	if len(face.ManaAbilities) != 1 {
		t.Fatalf("got %d mana abilities, want 1", len(face.ManaAbilities))
	}
	if !manaAbilityReadsEntryColorChoice(&face.ManaAbilities[0]) {
		t.Fatalf("mana ability does not read the entry color choice: %#v", face.ManaAbilities[0])
	}
}

func TestLowerEntersTappedColorChoiceReplacement(t *testing.T) {
	t.Parallel()
	// The combined "This land enters tapped. As it enters, choose a color." parses
	// as a single replacement ability with two enters effects; it must lower to a
	// replacement that both taps the permanent and records the color choice.
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Uncharted Haven",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "This land enters tapped. As it enters, choose a color.\n{T}: Add one mana of the chosen color.",
	})
	if len(face.ReplacementAbilities) != 1 {
		t.Fatalf("got %d replacement abilities, want 1", len(face.ReplacementAbilities))
	}
	replacement := face.ReplacementAbilities[0].Replacement
	if !replacement.EntersTapped || !replacement.EntryColorChoice {
		t.Fatalf("replacement must both enter tapped and record a color choice: %+v", replacement)
	}
	if len(face.ManaAbilities) != 1 || !manaAbilityReadsEntryColorChoice(&face.ManaAbilities[0]) {
		t.Fatalf("mana ability does not read the entry color choice: %#v", face.ManaAbilities)
	}
}

func TestLowerEntryColorChoiceForbiddenColor(t *testing.T) {
	t.Parallel()
	// "This land enters tapped. As it enters, choose a color other than white."
	// parses as a single replacement ability with two enters effects; it must
	// lower to a replacement that taps, records the color choice, and excludes the
	// forbidden color. The composite "Add {W} or one mana of the chosen color."
	// body lowers to a fixed-or-chosen mana ability.
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Thriving Heath",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "This land enters tapped. As it enters, choose a color other than white.\n{T}: Add {W} or one mana of the chosen color.",
	})
	if len(face.ReplacementAbilities) != 1 {
		t.Fatalf("got %d replacement abilities, want 1", len(face.ReplacementAbilities))
	}
	replacement := face.ReplacementAbilities[0].Replacement
	if !replacement.EntersTapped || !replacement.EntryColorChoice {
		t.Fatalf("replacement must both enter tapped and record a color choice: %+v", replacement)
	}
	if replacement.EntryColorChoiceExclude != mana.W {
		t.Fatalf("forbidden color = %q, want white", replacement.EntryColorChoiceExclude)
	}
	if len(face.ManaAbilities) != 1 {
		t.Fatalf("got %d mana abilities, want 1", len(face.ManaAbilities))
	}
	if !reflect.DeepEqual(face.ManaAbilities[0], game.TapFixedOrChosenColorManaAbility(
		"{T}: Add {W} or one mana of the chosen color.", mana.W)) {
		t.Fatalf("mana ability is not the fixed-or-chosen composite: %#v", face.ManaAbilities[0])
	}
}

func TestLowerEntryColorChoiceForbiddenColorSeparateSentences(t *testing.T) {
	t.Parallel()
	// The Gate sub-cycle prints the enters-tapped and color-choice clauses as
	// separate sentences, lowering to two replacement abilities.
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Citadel Gate",
		Layout:     "normal",
		TypeLine:   "Land — Gate",
		OracleText: "This land enters tapped.\nAs this land enters, choose a color other than white.\n{T}: Add {W} or one mana of the chosen color.",
	})
	if len(face.ReplacementAbilities) != 2 {
		t.Fatalf("got %d replacement abilities, want 2", len(face.ReplacementAbilities))
	}
	var choice *game.ReplacementEffect
	for i := range face.ReplacementAbilities {
		if face.ReplacementAbilities[i].Replacement.EntryColorChoice {
			choice = &face.ReplacementAbilities[i].Replacement
		}
	}
	if choice == nil {
		t.Fatal("no entry color-choice replacement lowered")
	}
	if choice.EntersTapped {
		t.Fatal("standalone color-choice replacement must not also enter tapped")
	}
	if choice.EntryColorChoiceExclude != mana.W {
		t.Fatalf("forbidden color = %q, want white", choice.EntryColorChoiceExclude)
	}
}

func TestLowerEntryTypeChoiceReplacement(t *testing.T) {
	t.Parallel()
	// "As this <permanent> enters, choose a creature type." must lower to an
	// entry-time creature-type-choice replacement that records the chosen type on
	// the permanent (CR 614.12). #554 groundwork.
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Banner",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "As this artifact enters, choose a creature type.",
	})
	if len(face.ReplacementAbilities) != 1 {
		t.Fatalf("got %d replacement abilities, want 1", len(face.ReplacementAbilities))
	}
	replacement := face.ReplacementAbilities[0].Replacement
	if !replacement.EntryTypeChoice {
		t.Fatalf("replacement is not an entry type-choice replacement: %+v", replacement)
	}
	if replacement.EntersTapped || replacement.EntryColorChoice {
		t.Fatalf("type-choice replacement must not tap or record a color: %+v", replacement)
	}
}

func TestLowerEntryTypeChoiceWithReferencingAbilityFailsClosed(t *testing.T) {
	t.Parallel()
	// A full creature-type-choice card (Metallic Mimic) also references "the
	// chosen type" in abilities the runtime cannot yet model; the card must fail
	// closed rather than generate a partial face. #554 stays fail-closed for
	// referencing abilities.
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Metallic Mimic",
		Layout:     "normal",
		TypeLine:   "Artifact Creature — Shapeshifter",
		OracleText: "As this creature enters, choose a creature type.\nThis creature is the chosen type in addition to its other types.\nEach other creature you control of the chosen type enters with an additional +1/+1 counter on it.",
	})
	if len(diagnostics) == 0 {
		t.Fatal("expected fail-closed diagnostics for referencing abilities, got none")
	}
}

func manaAbilityReadsEntryColorChoice(ability *game.ManaAbility) bool {
	for i := range ability.Content.Modes {
		mode := &ability.Content.Modes[i]
		for j := range mode.Sequence {
			addMana, ok := mode.Sequence[j].Primitive.(game.AddMana)
			if ok && addMana.EntryChoiceFrom == game.EntryColorChoiceKey {
				return true
			}
		}
	}
	return false
}

func TestLowerNamedTokenSetReplacement(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Academy Manufactor",
		Layout:     "normal",
		TypeLine:   "Artifact Creature — Construct",
		OracleText: "If you would create a Clue, Food, or Treasure token, instead create one of each.",
	})
	if len(face.ReplacementAbilities) != 1 {
		t.Fatalf("got %d replacement abilities, want 1", len(face.ReplacementAbilities))
	}
	replacement := face.ReplacementAbilities[0].Replacement
	if replacement.MatchEvent != game.EventTokenCreated ||
		replacement.ControllerFilter != game.TriggerControllerYou ||
		replacement.Duration != game.DurationPermanent ||
		len(replacement.CreateOneOfEachTokens) != 3 {
		t.Fatalf("replacement = %+v, want one-of-each token replacement with 3 defs", replacement)
	}
	wantNames := map[string]bool{"Clue": false, "Food": false, "Treasure": false}
	for _, def := range replacement.CreateOneOfEachTokens {
		if _, ok := wantNames[def.Name]; !ok {
			t.Fatalf("unexpected token def %q", def.Name)
		}
		wantNames[def.Name] = true
	}
	for name, seen := range wantNames {
		if !seen {
			t.Fatalf("missing token def %q", name)
		}
	}
}

func TestGenerateNamedTokenSetReplacementSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Academy Manufactor",
		Layout:     "normal",
		TypeLine:   "Artifact Creature — Construct",
		OracleText: "If you would create a Clue, Food, or Treasure token, instead create one of each.",
	}, "p")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.NamedTokenSetReplacement",
		"game.TriggerControllerYou",
		"[]*game.CardDef{",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if _, err := goparser.ParseFile(token.NewFileSet(), "generated.go", source, goparser.AllErrors); err != nil {
		t.Fatalf("generated source does not parse: %v\n%s", err, source)
	}
}

func TestLowerTokenCreationReplacement(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Anointed Procession",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "If an effect would create one or more tokens under your control, it creates twice that many of those tokens instead.",
	})
	if len(face.ReplacementAbilities) != 1 {
		t.Fatalf("got %d replacement abilities, want 1", len(face.ReplacementAbilities))
	}
	replacement := face.ReplacementAbilities[0].Replacement
	if replacement.MatchEvent != game.EventTokenCreated ||
		replacement.ControllerFilter != game.TriggerControllerYou ||
		replacement.TokenMultiplier != 2 ||
		replacement.Duration != game.DurationPermanent {
		t.Fatalf("replacement = %+v, want token creation doubler", replacement)
	}
}

func TestLowerPassiveTokenCreationReplacement(t *testing.T) {
	t.Parallel()
	ptr := func(s string) *string { return &s }
	tests := map[string]*ScryfallCard{
		"Mondrak": {
			Name:       "Mondrak, Glory Dominus",
			Layout:     "normal",
			TypeLine:   "Legendary Creature — Phyrexian Horror",
			ManaCost:   "{2}{W}{W}",
			Power:      ptr("3"),
			Toughness:  ptr("3"),
			OracleText: "If one or more tokens would be created under your control, twice that many of those tokens are created instead.\n{1}{W/P}{W/P}, Sacrifice two other artifacts and/or creatures: Put an indestructible counter on Mondrak.",
		},
		"Adrix and Nev": {
			Name:       "Adrix and Nev, Twincasters",
			Layout:     "normal",
			TypeLine:   "Legendary Creature — Merfolk Wizard",
			ManaCost:   "{2}{G}{U}",
			Power:      ptr("2"),
			Toughness:  ptr("3"),
			OracleText: "Ward {2}\nIf one or more tokens would be created under your control, twice that many of those tokens are created instead.",
		},
	}
	for name, card := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, card)
			var doubler game.ReplacementEffect
			found := false
			for _, ability := range face.ReplacementAbilities {
				if ability.Replacement.MatchEvent == game.EventTokenCreated {
					doubler = ability.Replacement
					found = true
				}
			}
			if !found {
				t.Fatalf("face %+v, want a token-creation replacement", face.ReplacementAbilities)
			}
			if doubler.ControllerFilter != game.TriggerControllerYou ||
				doubler.TokenMultiplier != 2 ||
				doubler.Duration != game.DurationPermanent {
				t.Fatalf("replacement = %+v, want token creation doubler", doubler)
			}
		})
	}
}

func TestLowerRejectsOptionalReplacementEffect(t *testing.T) {
	t.Parallel()
	faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Optional Procession",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "If an effect would create one or more tokens under your control, it may create twice that many of those tokens instead.",
	})
	if len(diagnostics) == 0 {
		t.Fatalf("faces = %#v, want unsupported optional replacement diagnostic", faces)
	}
}

func TestLowerDamageReplacement(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		oracleText   string
		multiplier   int
		addend       int
		sourceColors []color.Color
	}{
		{
			name:         "red additive damage",
			oracleText:   "If another red source you control would deal damage to a permanent or player, it deals that much damage plus 1 to that permanent or player instead.",
			addend:       1,
			sourceColors: []color.Color{color.Red},
		},
		{
			name:       "double damage",
			oracleText: "If a source you control would deal damage to a permanent or player, it deals double that damage to that permanent or player instead.",
			multiplier: 2,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Damage Replacer",
				Layout:     "normal",
				TypeLine:   "Creature",
				OracleText: test.oracleText,
				Power:      new("4"),
				Toughness:  new("5"),
			})
			if len(face.ReplacementAbilities) != 1 {
				t.Fatalf("got %d replacement abilities, want 1", len(face.ReplacementAbilities))
			}
			replacement := face.ReplacementAbilities[0].Replacement
			if replacement.MatchEvent != game.EventDamageDealt ||
				replacement.ControllerFilter != game.TriggerControllerYou ||
				replacement.DamageMultiplier != test.multiplier ||
				replacement.DamageAddend != test.addend ||
				!slices.Equal(replacement.DamageSourceColors, test.sourceColors) ||
				replacement.DamageExcludeSource != (test.name == "red additive damage") ||
				replacement.Duration != game.DurationPermanent {
				t.Fatalf("replacement = %+v, want damage replacement", replacement)
			}
		})
	}
}

func TestLowerCounterPlacementReplacement(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                  string
		oracleText            string
		matchCounterKind      bool
		counterKind           counter.Kind
		multiplier            int
		addend                int
		recipientAnyPermanent bool
	}{
		{
			name:             "specific plus one counters",
			oracleText:       "If one or more +1/+1 counters would be put on a creature you control, twice that many +1/+1 counters are put on that creature instead.",
			matchCounterKind: true,
			counterKind:      counter.PlusOnePlusOne,
			multiplier:       2,
		},
		{
			name:       "any counters",
			oracleText: "If you would put one or more counters on a permanent or player, put twice that many of each of those kinds of counters on that permanent or player instead.",
			multiplier: 2,
		},
		{
			name:                  "controlled permanent any counters doubling",
			oracleText:            "If an effect would put one or more counters on a permanent you control, it puts twice that many of those counters on that permanent instead.",
			multiplier:            2,
			recipientAnyPermanent: true,
		},
		{
			name:                  "controlled permanent plus one counters additive",
			oracleText:            "If one or more +1/+1 counters would be put on a permanent you control, that many plus one +1/+1 counters are put on that permanent instead.",
			matchCounterKind:      true,
			counterKind:           counter.PlusOnePlusOne,
			addend:                1,
			recipientAnyPermanent: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Counter Doubler",
				Layout:     "normal",
				TypeLine:   "Enchantment",
				OracleText: test.oracleText,
			})
			if len(face.ReplacementAbilities) != 1 {
				t.Fatalf("got %d replacement abilities, want 1", len(face.ReplacementAbilities))
			}
			replacement := face.ReplacementAbilities[0].Replacement
			if replacement.MatchEvent != game.EventCountersAdded ||
				replacement.ControllerFilter != game.TriggerControllerYou ||
				replacement.CounterMultiplier != test.multiplier ||
				replacement.CounterAddend != test.addend ||
				replacement.MatchCounterKind != test.matchCounterKind ||
				replacement.CounterKindFilter != test.counterKind ||
				replacement.CounterRecipientAnyPermanent != test.recipientAnyPermanent ||
				replacement.Duration != game.DurationPermanent {
				t.Fatalf("replacement = %+v, want counter placement modifier", replacement)
			}
		})
	}
}

func TestGenerateTokenCreationReplacementSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Parallel Lives",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "If an effect would create one or more tokens under your control, it creates twice that many of those tokens instead.",
	}, "p")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.TokenCreationReplacement",
		"game.TriggerControllerYou",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if _, err := goparser.ParseFile(token.NewFileSet(), "generated.go", source, goparser.AllErrors); err != nil {
		t.Fatalf("generated source does not parse: %v\n%s", err, source)
	}
}

func TestGenerateDamageReplacementSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Embermaw Hellion",
		Layout:     "normal",
		TypeLine:   "Creature — Hellion",
		OracleText: "If another red source you control would deal damage to a permanent or player, it deals that much damage plus 1 to that permanent or player instead.",
		Power:      new("4"),
		Toughness:  new("5"),
	}, "e")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.DamageReplacementExcludingSource",
		"color.Red",
		"game.TriggerControllerYou",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if _, err := goparser.ParseFile(token.NewFileSet(), "generated.go", source, goparser.AllErrors); err != nil {
		t.Fatalf("generated source does not parse: %v\n%s", err, source)
	}
}

func TestGenerateCounterPlacementReplacementSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Branching Evolution",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "If one or more +1/+1 counters would be put on a creature you control, twice that many +1/+1 counters are put on that creature instead.",
	}, "b")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.CounterPlacementReplacement",
		"counter.PlusOnePlusOne",
		"game.TriggerControllerYou",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if _, err := goparser.ParseFile(token.NewFileSet(), "generated.go", source, goparser.AllErrors); err != nil {
		t.Fatalf("generated source does not parse: %v\n%s", err, source)
	}
}

func TestGenerateAdditiveCounterPlacementReplacementSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Hardened Scales",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "If one or more +1/+1 counters would be put on a creature you control, that many plus one +1/+1 counters are put on it instead.",
	}, "h")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.CounterPlacementReplacement",
		"counter.PlusOnePlusOne",
		"game.TriggerControllerYou",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if _, err := goparser.ParseFile(token.NewFileSet(), "generated.go", source, goparser.AllErrors); err != nil {
		t.Fatalf("generated source does not parse: %v\n%s", err, source)
	}
}

func TestGenerateDoublingSeasonSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Doubling Season",
		Layout:   "normal",
		TypeLine: "Enchantment",
		OracleText: "If an effect would create one or more tokens under your control, it creates twice that many of those tokens instead.\n" +
			"If an effect would put one or more counters on a permanent you control, it puts twice that many of those counters on that permanent instead.",
	}, "d")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.TokenCreationReplacement",
		"game.ControlledPermanentCounterPlacementReplacement",
		"game.TriggerControllerYou",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if _, err := goparser.ParseFile(token.NewFileSet(), "generated.go", source, goparser.AllErrors); err != nil {
		t.Fatalf("generated source does not parse: %v\n%s", err, source)
	}
}

func TestGenerateControlledPermanentCounterKindReplacementSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Permanent Scales",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "If one or more +1/+1 counters would be put on a permanent you control, that many plus one +1/+1 counters are put on that permanent instead.",
	}, "p")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.ControlledPermanentCounterKindPlacementReplacement",
		"counter.PlusOnePlusOne",
		"game.TriggerControllerYou",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if _, err := goparser.ParseFile(token.NewFileSet(), "generated.go", source, goparser.AllErrors); err != nil {
		t.Fatalf("generated source does not parse: %v\n%s", err, source)
	}
}

func TestLowerEntersWithCountersReplacement(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		typeLine   string
		oracleText string
		kind       counter.Kind
		amount     int
	}{
		{
			name:       "plus one counters",
			typeLine:   "Creature — Shapeshifter",
			oracleText: "This creature enters with three +1/+1 counters on it.",
			kind:       counter.PlusOnePlusOne,
			amount:     3,
		},
		{
			name:       "shield counter",
			typeLine:   "Creature — Human Knight",
			oracleText: "This creature enters with a shield counter on it.",
			kind:       counter.Shield,
			amount:     1,
		},
		{
			name:       "charge counters",
			typeLine:   "Artifact",
			oracleText: "This artifact enters with two charge counters on it.",
			kind:       counter.Charge,
			amount:     2,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Permanent",
				Layout:     "normal",
				TypeLine:   test.typeLine,
				OracleText: test.oracleText,
			})
			if len(face.ReplacementAbilities) != 1 {
				t.Fatalf("got %d replacement abilities, want 1", len(face.ReplacementAbilities))
			}
			replacement := face.ReplacementAbilities[0].Replacement
			if replacement.EntersTapped {
				t.Fatal("replacement unexpectedly enters tapped")
			}
			if len(replacement.EntersWithCounters) != 1 {
				t.Fatalf("counter placements = %#v, want one", replacement.EntersWithCounters)
			}
			placement := replacement.EntersWithCounters[0]
			if placement.Kind != test.kind || placement.Amount != test.amount {
				t.Fatalf("placement = %#v, want %v x%d", placement, test.kind, test.amount)
			}
		})
	}
}

func TestGenerateEntersWithCountersReplacementSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Creature",
		Layout:     "normal",
		TypeLine:   "Creature — Shapeshifter",
		OracleText: "This creature enters with three +1/+1 counters on it.",
		Power:      new("0"),
		Toughness:  new("0"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	for _, wanted := range []string{
		`game.EntersWithCountersReplacement("This creature enters with three +1/+1 counters on it."`,
		"game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 3}",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if _, err := goparser.ParseFile(token.NewFileSet(), "generated.go", source, goparser.AllErrors); err != nil {
		t.Fatalf("generated source does not parse: %v\n%s", err, source)
	}
}

func TestLowerEntersWithXCountersReplacement(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Walking Ballista",
		Layout:     "normal",
		TypeLine:   "Artifact Creature — Construct",
		OracleText: "This creature enters with X +1/+1 counters on it.",
		Power:      new("0"),
		Toughness:  new("0"),
	})
	if len(face.ReplacementAbilities) != 1 {
		t.Fatalf("got %d replacement abilities, want 1", len(face.ReplacementAbilities))
	}
	replacement := face.ReplacementAbilities[0].Replacement
	if len(replacement.EntersWithCounters) != 1 {
		t.Fatalf("counter placements = %#v, want one", replacement.EntersWithCounters)
	}
	placement := replacement.EntersWithCounters[0]
	if placement.Kind != counter.PlusOnePlusOne || !placement.AmountFromX || placement.Amount != 0 {
		t.Fatalf("placement = %#v, want +1/+1 from X", placement)
	}
}

func TestGenerateEntersWithXCountersReplacementSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Walking Ballista",
		Layout:     "normal",
		TypeLine:   "Artifact Creature — Construct",
		OracleText: "This creature enters with X +1/+1 counters on it.",
		Power:      new("0"),
		Toughness:  new("0"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	if !strings.Contains(source, "game.CounterPlacement{Kind: counter.PlusOnePlusOne, AmountFromX: true}") {
		t.Fatalf("source missing X counter placement:\n%s", source)
	}
	if _, err := goparser.ParseFile(token.NewFileSet(), "generated.go", source, goparser.AllErrors); err != nil {
		t.Fatalf("generated source does not parse: %v\n%s", err, source)
	}
}

func TestLowerEntersWithCountersRejectsUnsupportedForms(t *testing.T) {
	t.Parallel()
	tests := map[string]string{
		"conditional": "If a creature died this turn, this creature enters with a +1/+1 counter on it.",
	}
	for name, oracleText := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, diagnostics := lowerExecutableFaces(&ScryfallCard{
				Name:       "Test Creature",
				Layout:     "normal",
				TypeLine:   "Creature",
				OracleText: oracleText,
				Power:      new("1"),
				Toughness:  new("1"),
			})
			if len(diagnostics) == 0 {
				t.Fatal("expected diagnostic")
			}
			if diagnostics[0].Summary != "unsupported enters-with-counters replacement" {
				t.Fatalf("summary = %q, want unsupported enters-with-counters replacement", diagnostics[0].Summary)
			}
		})
	}
}

func TestLowerEntersTappedWithCountersReplacement(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Vivid Marsh",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "This land enters tapped with two charge counters on it.\n{T}: Add {B}.",
	})
	if len(face.ReplacementAbilities) != 1 {
		t.Fatalf("got %d replacement abilities, want 1", len(face.ReplacementAbilities))
	}
	replacement := face.ReplacementAbilities[0].Replacement
	if !replacement.EntersTapped {
		t.Fatal("replacement does not enter tapped")
	}
	if replacement.Condition.Exists {
		t.Fatal("replacement unexpectedly has a condition")
	}
	if len(replacement.EntersWithCounters) != 1 {
		t.Fatalf("counter placements = %#v, want one", replacement.EntersWithCounters)
	}
	placement := replacement.EntersWithCounters[0]
	if placement.Kind != counter.Charge || placement.Amount != 2 {
		t.Fatalf("placement = %#v, want charge x2", placement)
	}
}

func TestGenerateEntersTappedWithCountersReplacementSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Vivid Marsh",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "This land enters tapped with two charge counters on it.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	for _, wanted := range []string{
		`game.EntersTappedWithCountersReplacement("This land enters tapped with two charge counters on it."`,
		"game.CounterPlacement{Kind: counter.Charge, Amount: 2}",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if _, err := goparser.ParseFile(token.NewFileSet(), "generated.go", source, goparser.AllErrors); err != nil {
		t.Fatalf("generated source does not parse: %v\n%s", err, source)
	}
}

func TestLowerConditionalEntersWithCountersReplacement(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		typeLine   string
		oracleText string
		amount     int
		checkCond  func(t *testing.T, cond game.Condition)
	}{
		"morbid": {
			typeLine:   "Creature — Boar",
			oracleText: "Morbid — This creature enters with two +1/+1 counters on it if a creature died this turn.",
			amount:     2,
			checkCond: func(t *testing.T, cond game.Condition) {
				if !cond.EventHistory.Exists {
					t.Fatalf("condition = %#v, want EventHistory", cond)
				}
			},
		},
		"raid": {
			typeLine:   "Creature — Human Pirate",
			oracleText: "Raid — This creature enters with a +1/+1 counter on it if you attacked this turn.",
			amount:     1,
			checkCond: func(t *testing.T, cond game.Condition) {
				if !cond.EventHistory.Exists {
					t.Fatalf("condition = %#v, want EventHistory", cond)
				}
			},
		},
		"ferocious controls": {
			typeLine:   "Creature — Elephant",
			oracleText: "Ferocious — This creature enters with a +1/+1 counter on it if you control a creature with power 4 or greater.",
			amount:     1,
			checkCond: func(t *testing.T, cond game.Condition) {
				if !cond.ControlsMatching.Exists {
					t.Fatalf("condition = %#v, want ControlsMatching", cond)
				}
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Creature",
				Layout:     "normal",
				TypeLine:   test.typeLine,
				OracleText: test.oracleText,
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(face.ReplacementAbilities) != 1 {
				t.Fatalf("got %d replacement abilities, want 1", len(face.ReplacementAbilities))
			}
			replacement := face.ReplacementAbilities[0].Replacement
			if replacement.EntersTapped {
				t.Fatal("conditional enters-with-counters must not enter tapped")
			}
			if !replacement.Condition.Exists {
				t.Fatal("replacement missing condition")
			}
			test.checkCond(t, replacement.Condition.Val)
			if len(replacement.EntersWithCounters) != 1 ||
				replacement.EntersWithCounters[0].Kind != counter.PlusOnePlusOne ||
				replacement.EntersWithCounters[0].Amount != test.amount {
				t.Fatalf("placements = %#v, want +1/+1 x%d", replacement.EntersWithCounters, test.amount)
			}
		})
	}
}

func TestGenerateConditionalEntersWithCountersReplacementSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Festerhide Boar",
		Layout:     "normal",
		TypeLine:   "Creature — Boar",
		OracleText: "Trample\nMorbid — This creature enters with two +1/+1 counters on it if a creature died this turn.",
		Power:      new("4"),
		Toughness:  new("3"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.EntersWithCountersIfReplacement(",
		"EventHistory: opt.Val(",
		"game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 2}",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if _, err := goparser.ParseFile(token.NewFileSet(), "generated.go", source, goparser.AllErrors); err != nil {
		t.Fatalf("generated source does not parse: %v\n%s", err, source)
	}
}

func TestLowerConditionalEntersWithCountersFailsClosed(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		typeLine   string
		oracleText string
	}{
		// "a permanent left the battlefield under your control this turn" is not a
		// modeled condition predicate.
		"unmodeled revolt predicate": {
			typeLine:   "Creature — Elf Warrior",
			oracleText: "Revolt — This creature enters with two +1/+1 counters on it if a permanent left the battlefield under your control this turn.",
		},
		// Dynamic "for each" counter quantity is not modeled.
		"dynamic for each": {
			typeLine:   "Creature — Plant",
			oracleText: "Converge — This creature enters with two +1/+1 counters on it for each color of mana spent to cast it.",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, diagnostics := lowerExecutableFaces(&ScryfallCard{
				Name:       "Test Creature",
				Layout:     "normal",
				TypeLine:   test.typeLine,
				OracleText: test.oracleText,
				Power:      new("1"),
				Toughness:  new("1"),
			})
			if len(diagnostics) == 0 {
				t.Fatal("expected diagnostic")
			}
			if diagnostics[0].Summary != "unsupported enters-with-counters replacement" {
				t.Fatalf("summary = %q, want unsupported enters-with-counters replacement", diagnostics[0].Summary)
			}
		})
	}
}

func TestLowerSelfZoneDestinationReplacement(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		cardName      string
		typeLine      string
		oracleText    string
		matchFromZone bool
		fromZone      zone.Type
		replaceToZone zone.Type
	}{
		{
			name:          "from anywhere into library",
			cardName:      "Darksteel Colossus",
			typeLine:      "Artifact Creature — Golem",
			oracleText:    "If Darksteel Colossus would be put into a graveyard from anywhere, reveal Darksteel Colossus and shuffle it into its owner's library instead.",
			replaceToZone: zone.Library,
		},
		{
			name:          "dies into exile",
			cardName:      "Test Phoenix",
			typeLine:      "Creature — Phoenix",
			oracleText:    "If this creature would die, exile it instead.",
			matchFromZone: true,
			fromZone:      zone.Battlefield,
			replaceToZone: zone.Exile,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       test.cardName,
				Layout:     "normal",
				TypeLine:   test.typeLine,
				OracleText: test.oracleText,
				Power:      new("11"),
				Toughness:  new("11"),
			})
			if len(face.ReplacementAbilities) != 1 {
				t.Fatalf("got %d replacement abilities, want 1", len(face.ReplacementAbilities))
			}
			replacement := face.ReplacementAbilities[0].Replacement
			if replacement.MatchEvent != game.EventZoneChanged ||
				replacement.MatchFromZone != test.matchFromZone ||
				replacement.FromZone != test.fromZone ||
				!replacement.MatchToZone ||
				replacement.ToZone != zone.Graveyard ||
				replacement.ReplaceToZone != test.replaceToZone ||
				replacement.ShuffleIntoLibrary != (test.replaceToZone == zone.Library) ||
				replacement.RevealSource != (test.replaceToZone == zone.Library) {
				t.Fatalf("replacement = %+v, want self zone-destination replacement", replacement)
			}
		})
	}
}

func TestGenerateSelfZoneDestinationReplacementSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Darksteel Colossus",
		Layout:     "normal",
		TypeLine:   "Artifact Creature — Golem",
		OracleText: "If Darksteel Colossus would be put into a graveyard from anywhere, reveal Darksteel Colossus and shuffle it into its owner's library instead.",
		Power:      new("11"),
		Toughness:  new("11"),
	}, "d")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.EventZoneChanged",
		"MatchToZone:",
		"ToZone:",
		"zone.Graveyard",
		"ReplaceToZone:",
		"zone.Library",
		"ShuffleIntoLibrary:",
		"RevealSource:",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if _, err := goparser.ParseFile(token.NewFileSet(), "generated.go", source, goparser.AllErrors); err != nil {
		t.Fatalf("generated source does not parse: %v\n%s", err, source)
	}
}

func TestGenerateEquippedCreaturePTBuff(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Equipment",
		Layout:     "normal",
		TypeLine:   "Artifact — Equipment",
		OracleText: "Equipped creature gets +2/+0.\nEquip {2}",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	if !strings.Contains(source, "LayerPowerToughnessModify") {
		t.Fatalf("source does not contain static PT effect:\n%s", source)
	}
	if !strings.Contains(source, "AttachedObjectGroup") {
		t.Fatalf("source does not contain AttachedObjectGroup:\n%s", source)
	}
	if _, err := goparser.ParseFile(token.NewFileSet(), "generated.go", source, goparser.AllErrors); err != nil {
		t.Fatalf("generated source does not parse: %v\n%s", err, source)
	}
}

func TestGenerateEquippedCreaturePTBuffWithKeywords(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Equipment",
		Layout:     "normal",
		TypeLine:   "Artifact — Equipment",
		OracleText: "Equipped creature gets +2/+2 and has trample and lifelink.\nEquip {3}",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.LayerPowerToughnessModify",
		"game.LayerAbility",
		"AddKeywords: []game.Keyword",
		"game.Trample",
		"game.Lifelink",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateControlledCreaturesPTBuffWithKeyword(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Anthem",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "Creatures you control get +1/+1 and have vigilance.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	if !strings.Contains(source, "game.Vigilance") {
		t.Fatalf("source missing vigilance:\n%s", source)
	}
}

func TestLowerGraveyardRedirectReplacement(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                string
		typeLine            string
		oracle              string
		ownerFilter         game.TriggerControllerFilter
		cardTypes           []types.Card
		fromBattlefieldOnly bool
	}{
		{
			name:        "any graveyard from anywhere",
			typeLine:    "Enchantment",
			oracle:      "If a card would be put into a graveyard from anywhere, exile it instead.",
			ownerFilter: game.TriggerControllerAny,
		},
		{
			name:        "opponent graveyard from anywhere",
			typeLine:    "Creature — Human",
			oracle:      "If a card would be put into an opponent's graveyard from anywhere, exile it instead.",
			ownerFilter: game.TriggerControllerOpponent,
		},
		{
			name:        "typed card filter",
			typeLine:    "Creature — Human Soldier",
			oracle:      "If an instant or sorcery card would be put into a graveyard from anywhere, exile it instead.",
			ownerFilter: game.TriggerControllerAny,
			cardTypes:   []types.Card{types.Instant, types.Sorcery},
		},
		{
			name:                "permanent from battlefield",
			typeLine:            "Creature — Kor",
			oracle:              "If a permanent would be put into a graveyard, exile it instead.",
			ownerFilter:         game.TriggerControllerAny,
			fromBattlefieldOnly: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Card",
				Layout:     "normal",
				TypeLine:   test.typeLine,
				OracleText: test.oracle,
			})
			if len(face.ReplacementAbilities) != 1 {
				t.Fatalf("got %d replacement abilities, want 1", len(face.ReplacementAbilities))
			}
			replacement := face.ReplacementAbilities[0].Replacement
			if !replacement.ContinuousZoneRedirect {
				t.Fatalf("replacement is not a continuous graveyard redirect: %#v", replacement)
			}
			if replacement.ReplaceToZone != zone.Exile || replacement.ToZone != zone.Graveyard || !replacement.MatchToZone {
				t.Fatalf("replacement zones = %#v, want exile-instead-of-graveyard", replacement)
			}
			if replacement.RedirectOwnerFilter != test.ownerFilter {
				t.Fatalf("owner filter = %v, want %v", replacement.RedirectOwnerFilter, test.ownerFilter)
			}
			if !slices.Equal(replacement.RedirectTypeFilter, test.cardTypes) {
				t.Fatalf("type filter = %v, want %v", replacement.RedirectTypeFilter, test.cardTypes)
			}
			gotBattlefieldOnly := replacement.MatchFromZone && replacement.FromZone == zone.Battlefield
			if gotBattlefieldOnly != test.fromBattlefieldOnly {
				t.Fatalf("from-battlefield-only = %v, want %v", gotBattlefieldOnly, test.fromBattlefieldOnly)
			}
		})
	}
}
