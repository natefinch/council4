package cards_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/cards/a"
	"github.com/natefinch/council4/mtg/cards/b"
	"github.com/natefinch/council4/mtg/cards/c"
	"github.com/natefinch/council4/mtg/cards/d"
	"github.com/natefinch/council4/mtg/cards/e"
	"github.com/natefinch/council4/mtg/cards/f"
	"github.com/natefinch/council4/mtg/cards/g"
	"github.com/natefinch/council4/mtg/cards/h"
	"github.com/natefinch/council4/mtg/cards/i"
	"github.com/natefinch/council4/mtg/cards/k"
	"github.com/natefinch/council4/mtg/cards/l"
	"github.com/natefinch/council4/mtg/cards/m"
	"github.com/natefinch/council4/mtg/cards/n"
	"github.com/natefinch/council4/mtg/cards/p"
	"github.com/natefinch/council4/mtg/cards/r"
	"github.com/natefinch/council4/mtg/cards/s"
	"github.com/natefinch/council4/mtg/game"
)

func TestRegisteredCardsValidate(t *testing.T) {
	registry := cards.NewRegistry(registeredCardSets()...)
	if registry.Len() == 0 {
		t.Fatal("registry has no cards")
	}
	for _, card := range registeredCards() {
		t.Run(card.Name, func(t *testing.T) {
			issues := cardgen.ValidateCard(card, cardgen.ValidationOptions{})
			if len(issues) != 0 {
				t.Fatalf("validation issues:\n%s", formatValidationIssues(issues))
			}
		})
	}
}

func TestRegisteredCardAbilitiesHaveBodies(t *testing.T) {
	for _, card := range registeredCards() {
		t.Run(card.Name, func(t *testing.T) {
			assertFaceAbilitiesHaveBodies(t, card.Name, &card.CardFace)
			if card.Back.Exists {
				assertFaceAbilitiesHaveBodies(t, card.Name+" back", &card.Back.Val)
			}
		})
	}
}

// TestRegisteredCardFacesHaveNoLegacyAbilities asserts that registered card
// source definitions use only the categorized CardFace fields and leave the
// legacy Abilities slice empty. Nested AbilityDef values inside effect data
// (e.g. ContinuousEffect.AddAbilities, EmblemAbilities) are not card-face
// source abilities and are not checked here.
func TestRegisteredCardFacesHaveNoLegacyAbilities(t *testing.T) {
	for _, card := range registeredCards() {
		t.Run(card.Name, func(t *testing.T) {
			if n := len(card.Abilities); n != 0 {
				t.Errorf("front face has %d legacy Abilities entries; use categorized fields instead", n)
			}
			if card.Back.Exists {
				if n := len(card.Back.Val.Abilities); n != 0 {
					t.Errorf("back face has %d legacy Abilities entries; use categorized fields instead", n)
				}
			}
		})
	}
}

func assertFaceAbilitiesHaveBodies(t *testing.T, faceName string, face *game.CardFace) {
	t.Helper()
	abilities := face.AbilityDefs()
	for abilityIndex := range abilities {
		if abilities[abilityIndex].Body == nil {
			t.Fatalf("%s ability %d has nil Body: %+v", faceName, abilityIndex, abilities[abilityIndex])
		}
	}
}

func registeredCards() []*game.CardDef {
	var all []*game.CardDef
	for _, set := range registeredCardSets() {
		all = append(all, set...)
	}
	return all
}

func registeredCardSets() [][]*game.CardDef {
	return [][]*game.CardDef{
		a.Cards,
		b.Cards,
		c.Cards,
		d.Cards,
		e.Cards,
		f.Cards,
		g.Cards,
		h.Cards,
		i.Cards,
		k.Cards,
		l.Cards,
		m.Cards,
		n.Cards,
		p.Cards,
		r.Cards,
		s.Cards,
	}
}

func formatValidationIssues(issues []cardgen.ValidationIssue) string {
	var out strings.Builder
	for _, issue := range issues {
		_, _ = fmt.Fprintf(&out, "- %s %s: %s\n", issue.Code, issue.Path, issue.Message)
	}
	return out.String()
}
