package game

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func temptingOfferTokenDef() *CardDef {
	return &CardDef{CardFace: CardFace{
		Name:      "Elemental",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Sub("Elemental")},
		Power:     opt.Val(PT{Value: 1}),
		Toughness: opt.Val(PT{Value: 1}),
	}}
}

func temptingOfferSpell(optional bool, group opt.V[PlayerGroupReference], tempting bool) *CardDef {
	return &CardDef{CardFace: CardFace{
		Name:  "Tempt Test",
		Types: []types.Card{types.Sorcery},
		SpellAbility: opt.Val(Mode{
			Sequence: []Instruction{{
				Primitive: CreateToken{
					Amount:    Fixed(1),
					Source:    TokenDef(temptingOfferTokenDef()),
					Recipient: opt.Val(GroupOfferMemberReference()),
				},
				Optional:           optional,
				OptionalActorGroup: group,
				TemptingOffer:      tempting,
			}},
		}.Ability()),
	}}
}

func issueMessagePresent(issues []CardDefIssue, substr string) bool {
	for _, issue := range issues {
		if strings.Contains(issue.Message, substr) {
			return true
		}
	}
	return false
}

// A well-formed Tempting-offer instruction (Optional with an OptionalActorGroup)
// raises no TemptingOffer validation issue.
func TestValidateCardDefAllowsWellFormedTemptingOffer(t *testing.T) {
	card := temptingOfferSpell(true, opt.Val(OpponentsReference()), true)
	issues := ValidateCardDef(card)
	if issueMessagePresent(issues, "TemptingOffer") {
		t.Fatalf("issues = %+v, want no TemptingOffer issue for a well-formed offer", issues)
	}
}

// A Tempting-offer instruction that is not optional is invalid: the offer is
// inherently an optional per-member choice.
func TestValidateCardDefReportsNonOptionalTemptingOffer(t *testing.T) {
	card := temptingOfferSpell(false, opt.Val(OpponentsReference()), true)
	issues := ValidateCardDef(card)
	if !hasCardDefIssue(issues, CardDefIssueInvalidAbilityBody) ||
		!issueMessagePresent(issues, "TemptingOffer set on a non-optional instruction") {
		t.Fatalf("issues = %+v, want TemptingOffer non-optional issue", issues)
	}
}

// A Tempting-offer instruction without an OptionalActorGroup is invalid: there is
// no group to offer the effect to.
func TestValidateCardDefReportsTemptingOfferWithoutGroup(t *testing.T) {
	card := temptingOfferSpell(true, opt.V[PlayerGroupReference]{}, true)
	issues := ValidateCardDef(card)
	if !hasCardDefIssue(issues, CardDefIssueInvalidAbilityBody) ||
		!issueMessagePresent(issues, "TemptingOffer requires OptionalActorGroup") {
		t.Fatalf("issues = %+v, want TemptingOffer missing-group issue", issues)
	}
}
