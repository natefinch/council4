package main

import (
	"cmp"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/natefinch/council4/internal/carddata"
)

// A family is unclaimed, unbatched queue cards that each add the same one rules
// paragraph beyond paragraphs already on supported cards (Key), or none at all
// (empty Key). Supporting that paragraph once makes the whole family authorable,
// so it is a natural card batch. Cards adding two or more new paragraphs are
// left to ordinary queue order.
type family struct {
	Key   string       `json:"key"`
	Decks int          `json:"deck_occurrences"`
	Cards []familyCard `json:"cards"`
}

type familyCard struct {
	OracleID string `json:"oracle_id"`
	Revision string `json:"revision"`
	Name     string `json:"name"`
	Decks    int    `json:"deck_occurrences,omitempty"`
}

var (
	reminderText  = regexp.MustCompile(`\([^)]*\)`)
	manaSymbols   = regexp.MustCompile(`(?:\{[^}]+\})+`)
	powerTough    = regexp.MustCompile(`[+-]?\d+/[+-]?\d+`)
	colorWords    = regexp.MustCompile(`(?i)\b(?:white|blue|black|red|green|colorless)\b`)
	quantityWords = regexp.MustCompile(`(?i)\b(?:\d+|x|one|two|three|four|five|six|seven|eight|nine|ten|a|an)\b`)
)

// cardFamilies groups the queue's unreviewed candidates by the single rules
// paragraph each adds, largest families first, then by deck occurrences.
// Singleton families are omitted except the no-new-paragraph family.
func cardFamilies(queue manifest, corpus []carddata.ScryfallCard) []family {
	byOracle := make(map[string]carddata.ScryfallCard, len(corpus))
	subtypes := make(map[string]bool)
	for _, card := range corpus {
		if _, ok := byOracle[card.OracleID]; !ok {
			byOracle[card.OracleID] = card
		}
		for _, face := range append([]carddata.ScryfallCardFace{{TypeLine: card.TypeLine}}, card.CardFaces...) {
			if _, after, ok := strings.Cut(face.TypeLine, "—"); ok {
				for _, subtype := range strings.Fields(after) {
					subtypes[subtype] = true
				}
			}
		}
	}
	paragraphs := func(oracleID string) []string {
		card, ok := byOracle[oracleID]
		if !ok {
			return nil
		}
		faces := card.CardFaces
		if len(faces) == 0 {
			faces = []carddata.ScryfallCardFace{{Name: card.Name, OracleText: card.OracleText}}
		}
		var out []string
		for _, face := range faces {
			for line := range strings.SplitSeq(face.OracleText, "\n") {
				if line = normalizeParagraph(face.Name, line, subtypes); line != "" && !slices.Contains(out, line) {
					out = append(out, line)
				}
			}
		}
		return out
	}
	supported := make(map[string]bool)
	for _, r := range queue.Records {
		if r.Status == "installed-compiled" || r.Status == "implemented-verified" {
			for _, p := range paragraphs(r.OracleID) {
				supported[p] = true
			}
		}
	}
	groups := make(map[string]*family)
	for _, r := range queue.Records {
		if r.Status != "unclaimed" || r.Batch != "" || isToken(byOracle[r.OracleID]) {
			continue
		}
		cardParagraphs := paragraphs(r.OracleID)
		if cardParagraphs == nil {
			continue
		}
		var added []string
		for _, p := range cardParagraphs {
			if !supported[p] {
				added = append(added, p)
			}
		}
		if len(added) > 1 {
			continue
		}
		key := strings.Join(added, "")
		group := groups[key]
		if group == nil {
			group = &family{Key: key}
			groups[key] = group
		}
		group.Decks += r.Decks
		group.Cards = append(group.Cards, familyCard{OracleID: r.OracleID, Revision: r.Revision, Name: r.Name, Decks: r.Decks})
	}
	var out []family
	for _, group := range groups {
		if len(group.Cards) > 1 || group.Key == "" {
			slices.SortStableFunc(group.Cards, func(a, b familyCard) int { return cmp.Compare(b.Decks, a.Decks) })
			out = append(out, *group)
		}
	}
	slices.SortFunc(out, func(a, b family) int {
		return cmp.Or(cmp.Compare(len(b.Cards), len(a.Cards)), cmp.Compare(b.Decks, a.Decks), cmp.Compare(a.Key, b.Key))
	})
	return out
}

// normalizeParagraph reduces one Oracle paragraph to its rules shape: the
// card's own name, mana symbols, power/toughness, colors, quantities and
// subtypes become placeholders and reminder text is dropped.
func normalizeParagraph(name, paragraph string, subtypes map[string]bool) string {
	paragraph = reminderText.ReplaceAllString(paragraph, "")
	short, _, _ := strings.Cut(name, ",")
	for _, n := range []string{name, short} {
		if n != "" {
			paragraph = strings.ReplaceAll(paragraph, n, "~")
		}
	}
	paragraph = manaSymbols.ReplaceAllString(paragraph, "{M}")
	paragraph = powerTough.ReplaceAllString(paragraph, "P/T")
	paragraph = colorWords.ReplaceAllString(paragraph, "COLOR")
	paragraph = quantityWords.ReplaceAllString(paragraph, "N")
	words := strings.Fields(paragraph)
	for i, word := range words {
		if subtypes[strings.Trim(word, ".,;:")] {
			words[i] = "TYPE"
		}
	}
	return strings.ToLower(strings.Join(words, " "))
}

// isToken reports printed token cards. They are not deck cards: a token is
// created by the card that makes it, and the deck-card registry cannot admit
// printed token identities, so they are never authored as queue batches.
func isToken(card carddata.ScryfallCard) bool {
	return card.Layout == "token" || card.Layout == "double_faced_token"
}

func rejectTokenBatch(plan reviewedPlan, corpus []carddata.ScryfallCard) error {
	tokens := make(map[string]string)
	for _, card := range corpus {
		if isToken(card) {
			tokens[card.OracleID] = card.Name
		}
	}
	for _, b := range plan.Batches {
		for _, id := range b.OracleIDs {
			if name, ok := tokens[id]; ok {
				return fmt.Errorf("batch %s includes printed token %s; tokens come from the cards that create them and cannot be registered as deck cards", b.ID, name)
			}
		}
	}
	return nil
}
