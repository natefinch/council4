package cardgen

import (
	goparser "go/parser"
	"go/token"
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
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
		name             string
		oracleText       string
		matchCounterKind bool
		counterKind      counter.Kind
	}{
		{
			name:             "specific plus one counters",
			oracleText:       "If one or more +1/+1 counters would be put on a creature you control, twice that many +1/+1 counters are put on that creature instead.",
			matchCounterKind: true,
			counterKind:      counter.PlusOnePlusOne,
		},
		{
			name:       "any counters",
			oracleText: "If you would put one or more counters on a permanent or player, put twice that many of each of those kinds of counters on that permanent or player instead.",
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
				replacement.CounterMultiplier != 2 ||
				replacement.MatchCounterKind != test.matchCounterKind ||
				replacement.CounterKindFilter != test.counterKind ||
				replacement.Duration != game.DurationPermanent {
				t.Fatalf("replacement = %+v, want counter placement doubler", replacement)
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

func TestLowerEntersWithCountersRejectsUnsupportedForms(t *testing.T) {
	t.Parallel()
	tests := map[string]string{
		"conditional": "If a creature died this turn, this creature enters with a +1/+1 counter on it.",
		"dynamic":     "This creature enters with X +1/+1 counters on it.",
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
