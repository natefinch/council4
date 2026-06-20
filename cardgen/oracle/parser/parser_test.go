package parser

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/zone"
)

type cachedParserCard struct {
	Name       string       `json:"name"`
	OracleText string       `json:"oracle_text"`
	CardFaces  []cachedFace `json:"card_faces"`
	TypeLine   string       `json:"type_line"`
}

type cachedFace struct {
	Name       string `json:"name"`
	TypeLine   string `json:"type_line"`
	OracleText string `json:"oracle_text"`
}

func TestParseSagaChapterHeading(t *testing.T) {
	t.Parallel()
	source := "I, II, III — Draw a card."
	document, diagnostics := Parse(source, Context{Saga: true})
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
	document, diagnostics := Parse("I — Draw a card.", Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := document.Abilities[0]
	if ability.Kind == AbilityChapter || ability.AbilityWord == nil || ability.AbilityWord.Label != "I" {
		t.Fatalf("ability = %#v, want ordinary ability-word syntax", ability)
	}
}

func TestParseSpellAdditionalCost(t *testing.T) {
	t.Parallel()
	source := "As an additional cost to cast this spell, sacrifice a creature.\nDestroy target creature."
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 2 {
		t.Fatalf("abilities = %d, want 2", len(document.Abilities))
	}
	cost := document.Abilities[0]
	if cost.Kind != AbilitySpellAdditionalCost {
		t.Fatalf("ability[0] kind = %v, want AbilitySpellAdditionalCost", cost.Kind)
	}
	if len(cost.Sentences) != 0 {
		t.Fatalf("ability[0] sentences = %d, want 0", len(cost.Sentences))
	}
	if cost.CostSyntax == nil || len(cost.CostSyntax.Components) != 1 ||
		cost.CostSyntax.Components[0].Kind != CostComponentSacrifice {
		t.Fatalf("ability[0] cost = %#v, want one sacrifice component", cost.CostSyntax)
	}
	if document.Abilities[1].Kind != AbilitySpell {
		t.Fatalf("ability[1] kind = %v, want AbilitySpell", document.Abilities[1].Kind)
	}
}

func TestParseSpellAdditionalCostOnPermanentSpell(t *testing.T) {
	t.Parallel()
	// A permanent spell (creature) prints the same additional-cost prefix; the
	// paragraph is recognized regardless of the card's permanent type.
	source := "As an additional cost to cast this spell, exile a creature card from your graveyard.\nFlying"
	document, diagnostics := Parse(source, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 2 {
		t.Fatalf("abilities = %d, want 2", len(document.Abilities))
	}
	cost := document.Abilities[0]
	if cost.Kind != AbilitySpellAdditionalCost {
		t.Fatalf("ability[0] kind = %v, want AbilitySpellAdditionalCost", cost.Kind)
	}
	if cost.CostSyntax == nil || len(cost.CostSyntax.Components) != 1 ||
		cost.CostSyntax.Components[0].Kind != CostComponentExile {
		t.Fatalf("ability[0] cost = %#v, want one exile component", cost.CostSyntax)
	}
	component := cost.CostSyntax.Components[0]
	if !component.AmountKnown || component.AmountValue != 1 ||
		component.SourceZone != zone.Graveyard || component.ObjectNoun != ObjectNounCreature {
		t.Fatalf("exile component = %#v", component)
	}
}

func TestParseSpellAdditionalCostExileXCards(t *testing.T) {
	t.Parallel()
	source := "As an additional cost to cast this spell, exile X cards from your graveyard.\nDraw a card."
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	cost := document.Abilities[0]
	if cost.Kind != AbilitySpellAdditionalCost || cost.CostSyntax == nil ||
		len(cost.CostSyntax.Components) != 1 {
		t.Fatalf("ability[0] = %#v", cost)
	}
	component := cost.CostSyntax.Components[0]
	if component.Kind != CostComponentExile || !component.AmountFromX ||
		component.AmountKnown || component.SourceZone != zone.Graveyard {
		t.Fatalf("exile-X component = %#v", component)
	}
}

func TestParseSpellAdditionalCostPayXLife(t *testing.T) {
	t.Parallel()
	source := "As an additional cost to cast this spell, pay X life.\nAll creatures get -X/-X until end of turn."
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	cost := document.Abilities[0]
	if cost.Kind != AbilitySpellAdditionalCost || cost.CostSyntax == nil ||
		len(cost.CostSyntax.Components) != 1 {
		t.Fatalf("ability[0] = %#v", cost)
	}
	component := cost.CostSyntax.Components[0]
	if component.Kind != CostComponentPayLife || !component.AmountFromX || component.AmountKnown {
		t.Fatalf("pay-X-life component = %#v", component)
	}
}

func TestParseStructures(t *testing.T) {
	t.Parallel()
	source := "Formidable — {1}{G}, {T}: Draw a card. Then discard a card. (Do this once.)"
	document, diagnostics := Parse(source, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := document.Abilities[0]
	if ability.AbilityWord == nil || ability.AbilityWord.Label != "Formidable" {
		t.Fatalf("ability word = %#v", ability.AbilityWord)
	}
	if ability.costPhrase == nil || ability.costPhrase.Text != "{1}{G}, {T}" {
		t.Fatalf("cost = %#v", ability.costPhrase)
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
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
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

func TestParseModalActivatedAbility(t *testing.T) {
	t.Parallel()
	source := "{1}, Discard a card: Choose one —\n• Draw a card.\n• You gain 3 life."
	document, diagnostics := Parse(source, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("abilities = %#v, want one", document.Abilities)
	}
	ability := document.Abilities[0]
	if ability.Kind != AbilityActivated || ability.costPhrase == nil || ability.Modal == nil {
		t.Fatalf("ability = %#v, want modal activated ability", ability)
	}
	if ability.costPhrase.Text != "{1}, Discard a card" || ability.Modal.header.Text != "Choose one —" || len(ability.Modal.Options) != 2 {
		t.Fatalf("cost/header/options = %q/%q/%d", ability.costPhrase.Text, ability.Modal.header.Text, len(ability.Modal.Options))
	}

	withWord, diagnostics := Parse("Hellbent — "+source, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("ability-word diagnostics = %#v", diagnostics)
	}
	ability = withWord.Abilities[0]
	if ability.AbilityWord == nil || ability.AbilityWord.Label != "Hellbent" ||
		ability.costPhrase == nil || ability.costPhrase.Text != "{1}, Discard a card" ||
		ability.Modal == nil {
		t.Fatalf("ability-word modal activated ability = %#v", ability)
	}
}

func TestParseInlineModalAbility(t *testing.T) {
	t.Parallel()
	source := "Choose one — Noxious Hydra Breath deals 5 damage to each player; or destroy each tapped non-Head creature."
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
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
	if ability.Modal.header.Text != "Choose one —" {
		t.Fatalf("header = %q", ability.Modal.header.Text)
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
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
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
	document, diagnostics := Parse(source, Context{})
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
			document, diagnostics := Parse(test.source, Context{})
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
	document, diagnostics := Parse(source, Context{})
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
	document, diagnostics := Parse("(unclosed\nFlying", Context{})
	if len(document.Abilities) != 2 {
		t.Fatalf("abilities = %#v", document.Abilities)
	}
	if len(diagnostics) != 1 || diagnostics[0].Summary != "unclosed parenthesis" {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
}

func TestParseEmbeddedParenthesisDoesNotJoinAbilities(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse("Flying (gains flying\nTrample)", Context{})
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
	document, diagnostics := Parse(source, Context{})
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
			context := Context{
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
			for i := range document.Abilities {
				assertAbilitySpans(t, name, source, &document.Abilities[i])
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

func assertAbilitySpans(t *testing.T, name, source string, ability *Ability) {
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
		assertTextSpan(t, name+" ability word", source, ability.AbilityWord.Span, ability.AbilityWord.Label)
	}
	if ability.costPhrase != nil {
		assertTextSpan(t, name+" cost", source, ability.costPhrase.Span, ability.costPhrase.Text)
	}
	if ability.Trigger != nil {
		assertTextSpan(t, name+" trigger", source, ability.Trigger.Span, ability.Trigger.Text)
		assertTokensInSpan(t, name+" trigger", ability.Trigger.Span, ability.Trigger.Tokens)
		assertSpanContains(t, name+" trigger introduction", ability.Trigger.Span, ability.Trigger.Introduction.Span)
		if ability.Trigger.Event != "" {
			assertTextSpan(t, name+" trigger event", source, ability.Trigger.EventSpan, ability.Trigger.Event)
			assertSpanContains(t, name+" trigger event", ability.Trigger.Span, ability.Trigger.EventSpan)
		}
		if phaseStep := ability.Trigger.PhaseStep; phaseStep != nil {
			assertSpanContains(t, name+" phase/step", ability.Trigger.EventSpan, phaseStep.Span)
			assertSpanContains(t, name+" phase/step name", phaseStep.Span, phaseStep.Name.Span)
			if phaseStep.Quantifier.Span != (shared.Span{}) {
				assertSpanContains(t, name+" phase/step quantifier", phaseStep.Span, phaseStep.Quantifier.Span)
			}
			if phaseStep.Player.Span != (shared.Span{}) {
				assertSpanContains(t, name+" phase/step player", phaseStep.Span, phaseStep.Player.Span)
			}
			if phaseStep.Player.AttachedSubject.Span != (shared.Span{}) {
				assertSpanContains(t, name+" phase/step attached subject", phaseStep.Player.Span, phaseStep.Player.AttachedSubject.Span)
			}
		}
	}
	if ability.Modal == nil {
		return
	}
	assertTextSpan(t, name+" modal header", source, ability.Modal.header.Span, ability.Modal.header.Text)
	for i := range ability.Modal.Options {
		mode := &ability.Modal.Options[i]
		assertTextSpan(t, name+" mode", source, mode.Span, mode.Text)
		assertTokensInSpan(t, name+" mode", mode.Span, mode.Tokens)
		assertSpanContains(t, name+" mode", ability.Span, mode.Span)
		for _, sentence := range mode.Sentences {
			assertTextSpan(t, name+" mode sentence", source, sentence.Span, sentence.Text)
			assertSpanContains(t, name+" mode sentence", mode.Span, sentence.Span)
		}
	}
}

func assertTextSpan(t *testing.T, name, source string, span shared.Span, text string) {
	t.Helper()
	if span.Start.Offset < 0 || span.End.Offset < span.Start.Offset || span.End.Offset > len(source) {
		t.Fatalf("%s has invalid span %#v", name, span)
	}
	if got := source[span.Start.Offset:span.End.Offset]; got != text {
		t.Fatalf("%s text = %q, source span = %q", name, text, got)
	}
}

func assertTokensInSpan(t *testing.T, name string, parent shared.Span, tokens []shared.Token) {
	t.Helper()
	for _, token := range tokens {
		assertSpanContains(t, name+" token", parent, token.Span)
	}
}

func assertSpanContains(t *testing.T, name string, parent, child shared.Span) {
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
