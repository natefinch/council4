package oracle

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

type cachedParserCard struct {
	Name       string       `json:"name"`
	OracleText string       `json:"oracle_text"`
	CardFaces  []cachedFace `json:"card_faces"`
	TypeLine   string       `json:"type_line"`
}

func TestParseAbilityKinds(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source  string
		context ParseContext
		want    AbilityKind
	}{
		"spell": {
			source:  "Destroy target creature.",
			context: ParseContext{InstantOrSorcery: true},
			want:    AbilitySpell,
		},
		"activated": {
			source: "{T}: Add {G}.",
			want:   AbilityActivated,
		},
		"loyalty": {
			source:  "−2: Target creature you control fights target creature you don't control.",
			context: ParseContext{Planeswalker: true},
			want:    AbilityLoyalty,
		},
		"variable loyalty": {
			source:  "+X: Draw X cards.",
			context: ParseContext{Planeswalker: true},
			want:    AbilityLoyalty,
		},
		"numeric activated": {
			source: "2: Draw a card.",
			want:   AbilityActivated,
		},
		"triggered": {
			source: "Whenever you attack, draw a card.",
			want:   AbilityTriggered,
		},
		"ability word trigger": {
			source: "Formidable — Whenever you attack, draw a card.",
			want:   AbilityTriggered,
		},
		"saga chapter": {
			source: "I, II — Draw a card.",
			context: ParseContext{
				Saga: true,
			},
			want: AbilityChapter,
		},
		"replacement": {
			source: "This land enters tapped.",
			want:   AbilityReplacement,
		},
		"static": {
			source: "Creatures you control have haste.",
			want:   AbilityStatic,
		},
		"reminder": {
			source: "(This creature can block creatures with flying.)",
			want:   AbilityReminder,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, test.context)
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if len(document.Abilities) != 1 {
				t.Fatalf("abilities = %d", len(document.Abilities))
			}
			if got := document.Abilities[0].Kind; got != test.want {
				t.Fatalf("kind = %s, want %s", got, test.want)
			}
		})
	}
}

func TestParseSagaChapterHeading(t *testing.T) {
	t.Parallel()
	source := "I, II, III — Draw a card."
	document, diagnostics := Parse(source, ParseContext{Saga: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := document.Abilities[0]
	if ability.AbilityWord != nil {
		t.Fatalf("ability word = %#v, want nil", ability.AbilityWord)
	}
	if !slices.Equal(ability.Chapters, []int{1, 2, 3}) {
		t.Fatalf("chapters = %v, want [1 2 3]", ability.Chapters)
	}
	assertTextSpan(t, "chapter heading", source, ability.ChapterSpan, "I, II, III")
}

func TestParseDoesNotTreatRomanNumeralsAsChaptersOutsideSagaContext(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse("I — Draw a card.", ParseContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := document.Abilities[0]
	if ability.Kind == AbilityChapter || ability.AbilityWord == nil || ability.AbilityWord.Text != "I" {
		t.Fatalf("ability = %#v, want ordinary ability-word syntax", ability)
	}
}

func TestParseStructures(t *testing.T) {
	t.Parallel()
	source := "Formidable — {1}{G}, {T}: Draw a card. Then discard a card. (Do this once.)"
	document, diagnostics := Parse(source, ParseContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := document.Abilities[0]
	if ability.AbilityWord == nil || ability.AbilityWord.Text != "Formidable" {
		t.Fatalf("ability word = %#v", ability.AbilityWord)
	}
	if ability.Cost == nil || ability.Cost.Text != "{1}{G}, {T}" {
		t.Fatalf("cost = %#v", ability.Cost)
	}
	if len(ability.Sentences) != 3 {
		t.Fatalf("sentences = %#v", ability.Sentences)
	}
	if len(ability.Reminders) != 1 || ability.Reminders[0].Text != "(Do this once.)" {
		t.Fatalf("reminders = %#v", ability.Reminders)
	}
}

func TestParseModalAbility(t *testing.T) {
	t.Parallel()
	source := "Choose one —\n• Draw a card.\n• Target creature fights another target creature. (They deal damage.)"
	document, diagnostics := Parse(source, ParseContext{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("abilities = %d", len(document.Abilities))
	}
	ability := document.Abilities[0]
	if ability.Kind != AbilitySpell || ability.Modal == nil {
		t.Fatalf("modal ability = %#v", ability)
	}
	if ability.Text != source {
		t.Fatalf("text = %q", ability.Text)
	}
	if len(ability.Modal.Options) != 2 {
		t.Fatalf("options = %#v", ability.Modal.Options)
	}
	if got := ability.Modal.Options[1].Text; got != "Target creature fights another target creature. (They deal damage.)" {
		t.Fatalf("second mode = %q", got)
	}
	if ability.Modal.Options[1].Span.Start != ability.Modal.Options[1].Tokens[0].Span.Start {
		t.Fatal("mode span includes syntax outside mode tokens")
	}
}

func TestParseInlineModalAbility(t *testing.T) {
	t.Parallel()
	source := "Choose one — Noxious Hydra Breath deals 5 damage to each player; or destroy each tapped non-Head creature."
	document, diagnostics := Parse(source, ParseContext{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("abilities = %#v", document.Abilities)
	}
	ability := document.Abilities[0]
	if ability.Modal == nil || len(ability.Modal.Options) != 2 {
		t.Fatalf("modal = %#v", ability.Modal)
	}
	if ability.Modal.Header.Text != "Choose one —" {
		t.Fatalf("header = %q", ability.Modal.Header.Text)
	}
	if got := ability.Modal.Options[0].Text; got != "Noxious Hydra Breath deals 5 damage to each player" {
		t.Fatalf("first mode = %q", got)
	}
	if got := ability.Modal.Options[1].Text; got != "destroy each tapped non-Head creature." {
		t.Fatalf("second mode = %q", got)
	}
}

func TestChooseSentenceBeforeVillainousChoiceIsNotModalHeader(t *testing.T) {
	t.Parallel()
	source := "Choose up to four target creatures you don't control. For each of them, that creature's controller faces a villainous choice — That creature becomes a 1/1 white Human creature and loses all abilities, or you create a token that's a copy of it."
	document, diagnostics := Parse(source, ParseContext{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("abilities = %#v", document.Abilities)
	}
	ability := document.Abilities[0]
	if ability.Kind != AbilitySpell || ability.Modal != nil || ability.Text != source {
		t.Fatalf("ability = %#v", ability)
	}
}

func TestParseQuotedAbility(t *testing.T) {
	t.Parallel()
	source := `Equipped creature has "{2}: This creature gets +1/+0 until end of turn."`
	document, diagnostics := Parse(source, ParseContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	quoted := document.Abilities[0].Quoted
	if len(quoted) != 1 || quoted[0].Text != `"{2}: This creature gets +1/+0 until end of turn."` {
		t.Fatalf("quoted = %#v", quoted)
	}
	if len(document.Abilities[0].Sentences) != 1 {
		t.Fatal("sentence split inside quote")
	}
}

func TestParseNestedDelimitedText(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		source        string
		wantReminders int
		wantQuoted    int
	}{
		"reminder inside quote": {
			source:        `Enchanted creature has "Flying (it can't be blocked)."`,
			wantReminders: 0,
			wantQuoted:    1,
		},
		"quote inside reminder": {
			source:        `Flying (This means "can't be blocked.")`,
			wantReminders: 1,
			wantQuoted:    0,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, ParseContext{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			ability := document.Abilities[0]
			if len(ability.Reminders) != test.wantReminders || len(ability.Quoted) != test.wantQuoted {
				t.Fatalf("reminders = %#v, quoted = %#v", ability.Reminders, ability.Quoted)
			}
		})
	}
}

func TestParseMultilineReminderOnlyAbility(t *testing.T) {
	t.Parallel()
	source := "(You can cover a face-down creature with this reminder card.\nA card with morph can be turned face up any time for its morph cost.)"
	document, diagnostics := Parse(source, ParseContext{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("abilities = %#v", document.Abilities)
	}
	ability := document.Abilities[0]
	if ability.Kind != AbilityReminder || ability.Text != source {
		t.Fatalf("ability = %#v", ability)
	}
	if len(ability.Reminders) != 1 || ability.Reminders[0].Text != source {
		t.Fatalf("reminders = %#v", ability.Reminders)
	}
}

func TestParseUnclosedMultilineReminderRecoversAtNewline(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse("(unclosed\nFlying", ParseContext{})
	if len(document.Abilities) != 2 {
		t.Fatalf("abilities = %#v", document.Abilities)
	}
	if len(diagnostics) != 1 || diagnostics[0].Summary != "unclosed parenthesis" {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
}

func TestParseEmbeddedParenthesisDoesNotJoinAbilities(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse("Flying (gains flying\nTrample)", ParseContext{})
	if len(document.Abilities) != 2 {
		t.Fatalf("abilities = %#v", document.Abilities)
	}
	if len(diagnostics) != 2 ||
		diagnostics[0].Summary != "unclosed parenthesis" ||
		diagnostics[1].Summary != "unmatched parenthesis" {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
}

func TestParseDiagnosticsAndRecovery(t *testing.T) {
	t.Parallel()
	source := "Flying)\nChoose one —\nHaste\n\"unclosed"
	document, diagnostics := Parse(source, ParseContext{})
	if len(document.Abilities) != 4 {
		t.Fatalf("abilities = %d", len(document.Abilities))
	}
	if len(diagnostics) != 3 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if diagnostics[0].Summary != "unmatched parenthesis" ||
		diagnostics[1].Summary != "modal ability has no options" ||
		diagnostics[2].Summary != "unclosed quote" {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
}

func TestParseScryfallCacheLosslessly(t *testing.T) {
	t.Parallel()
	cache := filepath.Join("..", "..", ".cardwork", "deck", "cache", "scryfall")
	paths, err := filepath.Glob(filepath.Join(cache, "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Skip("local Scryfall cache is not present")
	}

	var texts int
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var card cachedParserCard
		if err := json.Unmarshal(data, &card); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		check := func(name, typeLine, source string) {
			t.Helper()
			if source == "" {
				return
			}
			texts++
			context := ParseContext{
				CardName:         name,
				InstantOrSorcery: typeLine == "Instant" || typeLine == "Sorcery",
				Planeswalker:     typeLine == "Planeswalker" || typeLine == "Legendary Planeswalker",
			}
			document, diagnostics := Parse(source, context)
			if len(diagnostics) != 0 {
				t.Fatalf("%s: diagnostics = %#v", name, diagnostics)
			}
			if document.Source != source ||
				document.Span.Start.Offset != 0 ||
				document.Span.End.Offset != len(source) {
				t.Fatalf("%s: document is not lossless", name)
			}
			for _, ability := range document.Abilities {
				assertAbilitySpans(t, name, source, ability)
			}
		}
		check(card.Name, card.TypeLine, card.OracleText)
		for _, face := range card.CardFaces {
			check(face.Name, face.TypeLine, face.OracleText)
		}
	}
	if texts != 59 {
		t.Fatalf("checked %d non-empty Oracle texts, want 59", texts)
	}
}

func assertAbilitySpans(t *testing.T, name, source string, ability Ability) {
	t.Helper()
	assertTextSpan(t, name+" ability", source, ability.Span, ability.Text)
	assertTokensInSpan(t, name+" ability", ability.Span, ability.Tokens)
	for _, sentence := range ability.Sentences {
		assertTextSpan(t, name+" sentence", source, sentence.Span, sentence.Text)
		assertSpanContains(t, name+" sentence", ability.Span, sentence.Span)
	}
	for _, reminder := range ability.Reminders {
		assertTextSpan(t, name+" reminder", source, reminder.Span, reminder.Text)
		assertSpanContains(t, name+" reminder", ability.Span, reminder.Span)
	}
	for _, quoted := range ability.Quoted {
		assertTextSpan(t, name+" quote", source, quoted.Span, quoted.Text)
		assertSpanContains(t, name+" quote", ability.Span, quoted.Span)
	}
	assertDisjoint(t, name, ability.Reminders, ability.Quoted)
	if ability.AbilityWord != nil {
		assertTextSpan(t, name+" ability word", source, ability.AbilityWord.Span, ability.AbilityWord.Text)
	}
	if ability.Cost != nil {
		assertTextSpan(t, name+" cost", source, ability.Cost.Span, ability.Cost.Text)
	}
	if ability.Modal == nil {
		return
	}
	assertTextSpan(t, name+" modal header", source, ability.Modal.Header.Span, ability.Modal.Header.Text)
	for _, mode := range ability.Modal.Options {
		assertTextSpan(t, name+" mode", source, mode.Span, mode.Text)
		assertTokensInSpan(t, name+" mode", mode.Span, mode.Tokens)
		assertSpanContains(t, name+" mode", ability.Span, mode.Span)
		for _, sentence := range mode.Sentences {
			assertTextSpan(t, name+" mode sentence", source, sentence.Span, sentence.Text)
			assertSpanContains(t, name+" mode sentence", mode.Span, sentence.Span)
		}
	}
}

func assertTextSpan(t *testing.T, name, source string, span Span, text string) {
	t.Helper()
	if span.Start.Offset < 0 || span.End.Offset < span.Start.Offset || span.End.Offset > len(source) {
		t.Fatalf("%s has invalid span %#v", name, span)
	}
	if got := source[span.Start.Offset:span.End.Offset]; got != text {
		t.Fatalf("%s text = %q, source span = %q", name, text, got)
	}
}

func assertTokensInSpan(t *testing.T, name string, parent Span, tokens []Token) {
	t.Helper()
	for _, token := range tokens {
		assertSpanContains(t, name+" token", parent, token.Span)
	}
}

func assertSpanContains(t *testing.T, name string, parent, child Span) {
	t.Helper()
	if child.Start.Offset < parent.Start.Offset || child.End.Offset > parent.End.Offset {
		t.Fatalf("%s span %#v is outside parent %#v", name, child, parent)
	}
}

func assertDisjoint(t *testing.T, name string, reminders, quoted []Delimited) {
	t.Helper()
	for _, reminder := range reminders {
		for _, quote := range quoted {
			if reminder.Span.Start.Offset < quote.Span.End.Offset &&
				quote.Span.Start.Offset < reminder.Span.End.Offset {
				t.Fatalf("%s reminder %#v overlaps quote %#v", name, reminder.Span, quote.Span)
			}
		}
	}
}
