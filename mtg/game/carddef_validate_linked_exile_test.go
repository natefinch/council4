package game

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
)

const linkedExileColorTestLink = "imprint"

func linkedExileColorIssuePresent(issues []CardDefIssue) bool {
	for _, issue := range issues {
		if issue.Code == CardDefIssueInvalidAbilityBody &&
			strings.Contains(issue.Message, "linked-exile-color mana ability requires an exile-from-hand effect") {
			return true
		}
	}
	return false
}

func imprintExileTriggered() TriggeredAbility {
	return TriggeredAbility{
		Trigger: TriggerCondition{Pattern: TriggerPattern{Event: EventPermanentEnteredBattlefield}},
		Content: Mode{Sequence: []Instruction{{
			Optional: true,
			Primitive: ExileFromHandChoice(
				ControllerReference(),
				Selection{ExcludedTypes: []types.Card{types.Artifact, types.Land}},
				Fixed(1),
				linkedExileColorTestLink,
			),
		}}}.Ability(),
	}
}

// TestValidateCardDefAllowsImprintWithExileFromHand verifies a face that pairs
// an exile-from-hand imprint with the imprint-color mana ability raises no
// linked-exile dependency issue.
func TestValidateCardDefAllowsImprintWithExileFromHand(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name:               "Chrome Mox",
		Types:              []types.Card{types.Artifact},
		TriggeredAbilities: []TriggeredAbility{imprintExileTriggered()},
		ManaAbilities:      []ManaAbility{TapLinkedExileColorManaAbility(linkedExileColorTestLink)},
	}}

	if linkedExileColorIssuePresent(ValidateCardDef(card)) {
		t.Fatal("paired imprint exile and mana ability reported a dependency issue")
	}
}

// TestValidateCardDefRejectsImprintManaWithoutExile verifies the imprint-color
// mana ability is rejected when no exile-from-hand effect publishes its link.
func TestValidateCardDefRejectsImprintManaWithoutExile(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name:          "Dangling Mox",
		Types:         []types.Card{types.Artifact},
		ManaAbilities: []ManaAbility{TapLinkedExileColorManaAbility(linkedExileColorTestLink)},
	}}

	if !linkedExileColorIssuePresent(ValidateCardDef(card)) {
		t.Fatal("imprint mana ability without an exile-from-hand effect was not rejected")
	}
}

// TestValidateCardDefRejectsImprintManaWithMismatchedLink verifies the
// dependency is link-specific: a different published link does not satisfy it.
func TestValidateCardDefRejectsImprintManaWithMismatchedLink(t *testing.T) {
	exile := imprintExileTriggered()
	exilePrim, ok := exile.Content.Modes[0].Sequence[0].Primitive.(ChooseFromZone)
	if !ok {
		t.Fatal("imprint exile primitive is not a ChooseFromZone")
	}
	exilePrim.Riders.PublishLinked = "other"
	exile.Content.Modes[0].Sequence[0].Primitive = exilePrim

	card := &CardDef{CardFace: CardFace{
		Name:               "Mismatched Mox",
		Types:              []types.Card{types.Artifact},
		TriggeredAbilities: []TriggeredAbility{exile},
		ManaAbilities:      []ManaAbility{TapLinkedExileColorManaAbility(linkedExileColorTestLink)},
	}}

	if !linkedExileColorIssuePresent(ValidateCardDef(card)) {
		t.Fatal("imprint mana ability with mismatched link was not rejected")
	}
}
