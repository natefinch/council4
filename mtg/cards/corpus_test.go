package cards_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/cards/tokens"
	"github.com/natefinch/council4/mtg/game"
)

func TestRegisteredCardsValidate(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	if registry.Len() == 0 {
		t.Fatal("registry has no cards")
	}
	for _, card := range registeredCards() {
		t.Run(card.Name, func(t *testing.T) {
			issues := game.ValidateCardDef(card)
			if len(issues) != 0 {
				t.Fatalf("validation issues:\n%s", formatValidationIssues(issues))
			}
		})
	}
}

func TestTokenCardsValidateOutsideRegistry(t *testing.T) {
	registered := make(map[*game.CardDef]bool)
	for _, card := range registeredCards() {
		registered[card] = true
	}
	if len(tokens.Cards) == 0 {
		t.Fatal("token catalog has no cards")
	}
	for _, card := range tokens.Cards {
		t.Run(card.Name, func(t *testing.T) {
			if registered[card] {
				t.Fatal("token definition is included in the ordinary card registry")
			}
			if issues := game.ValidateCardDef(card); len(issues) != 0 {
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

func assertFaceAbilitiesHaveBodies(t *testing.T, faceName string, face *game.CardFace) {
	t.Helper()
	for abilityIndex := 0; abilityIndex < face.AbilityCount(); abilityIndex++ {
		body := face.BodyAt(abilityIndex)
		if body == nil {
			t.Fatalf("%s ability %d has nil body", faceName, abilityIndex)
		}
	}
}

func registeredCards() []*game.CardDef {
	var all []*game.CardDef
	for _, set := range cards.DefaultCardSets() {
		all = append(all, set...)
	}
	return all
}

func formatValidationIssues(issues []game.CardDefIssue) string {
	var out strings.Builder
	for _, issue := range issues {
		_, _ = fmt.Fprintf(&out, "- %s %s: %s\n", issue.Code, issue.Path, issue.Message)
	}
	return out.String()
}
