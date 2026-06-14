package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestValidateProtectionKeywordRejectsMixedPredicates verifies that a
// ProtectionKeyword with more than one predicate group is rejected.
func TestValidateProtectionKeywordRejectsMixedPredicates(t *testing.T) {
	t.Parallel()
	makeProtCard := func(kw ProtectionKeyword) *CardDef {
		return &CardDef{CardFace: CardFace{
			Name:  "Test Creature",
			Types: []types.Card{types.Creature},
			StaticAbilities: []StaticAbility{{
				KeywordAbilities: []KeywordAbility{kw},
			}},
		}}
	}

	t.Run("colors and types mixed", func(t *testing.T) {
		t.Parallel()

		issues := ValidateCardDef(makeProtCard(ProtectionKeyword{
			FromColors: []color.Color{color.Red},
			FromTypes:  []types.Card{types.Creature},
		}))
		if !hasCardDefIssue(issues, CardDefIssueInvalidKeywordAbility) {
			t.Fatalf("expected invalid-keyword-ability for mixed predicates, got %+v", issues)
		}
	})

	t.Run("colors and everything mixed", func(t *testing.T) {
		t.Parallel()

		issues := ValidateCardDef(makeProtCard(ProtectionKeyword{
			FromColors: []color.Color{color.Red},
			Everything: true,
		}))
		if !hasCardDefIssue(issues, CardDefIssueInvalidKeywordAbility) {
			t.Fatalf("expected invalid-keyword-ability for mixed predicates, got %+v", issues)
		}
	})

	t.Run("single color predicate is valid", func(t *testing.T) {
		t.Parallel()

		issues := ValidateCardDef(makeProtCard(ProtectionKeyword{
			FromColors: []color.Color{color.Blue},
		}))
		if hasCardDefIssue(issues, CardDefIssueInvalidKeywordAbility) {
			t.Fatalf("unexpected invalid-keyword-ability for single predicate: %+v", issues)
		}
	})
}

// TestValidateProtectionKeywordRejectsUnknownSubtype verifies that a
// ProtectionKeyword referencing an unknown subtype is rejected.
func TestValidateProtectionKeywordRejectsUnknownSubtype(t *testing.T) {
	t.Parallel()
	issues := ValidateCardDef(&CardDef{CardFace: CardFace{
		Name:  "Test Creature",
		Types: []types.Card{types.Creature},
		StaticAbilities: []StaticAbility{{
			KeywordAbilities: []KeywordAbility{
				ProtectionKeyword{FromSubtypes: []types.Sub{"NotARealSubtype"}},
			},
		}},
	}})
	if !hasCardDefIssue(issues, CardDefIssueInvalidKeywordAbility) {
		t.Fatalf("expected invalid-keyword-ability for unknown subtype, got %+v", issues)
	}
}

// TestValidateProtectionKeywordAcceptsKnownSubtype verifies that a
// ProtectionKeyword with a known creature subtype passes validation.
func TestValidateProtectionKeywordAcceptsKnownSubtype(t *testing.T) {
	t.Parallel()
	issues := ValidateCardDef(&CardDef{CardFace: CardFace{
		Name:  "Test Creature",
		Types: []types.Card{types.Creature},
		StaticAbilities: []StaticAbility{{
			KeywordAbilities: []KeywordAbility{
				ProtectionKeyword{FromSubtypes: []types.Sub{types.Dragon}},
			},
		}},
	}})
	if hasCardDefIssue(issues, CardDefIssueInvalidKeywordAbility) {
		t.Fatalf("unexpected invalid-keyword-ability for known subtype Dragon: %+v", issues)
	}
}

// TestValidateProtectionKeywordRejectsUnknownCardType verifies that a
// ProtectionKeyword with an unrecognised types.Card value is rejected.
func TestValidateProtectionKeywordRejectsUnknownCardType(t *testing.T) {
	t.Parallel()
	issues := ValidateCardDef(&CardDef{CardFace: CardFace{
		Name:  "Test Creature",
		Types: []types.Card{types.Creature},
		StaticAbilities: []StaticAbility{{
			KeywordAbilities: []KeywordAbility{
				ProtectionKeyword{FromTypes: []types.Card{"NotARealType"}},
			},
		}},
	}})
	if !hasCardDefIssue(issues, CardDefIssueInvalidKeywordAbility) {
		t.Fatalf("expected invalid-keyword-ability for unknown card type, got %+v", issues)
	}
}

// TestValidateProtectionKeywordAcceptsCanonicalCardTypes verifies that all
// card types supported by the renderer are accepted by validation.
func TestValidateProtectionKeywordAcceptsCanonicalCardTypes(t *testing.T) {
	t.Parallel()
	for _, cardType := range []types.Card{
		types.Creature, types.Artifact, types.Enchantment,
		types.Land, types.Instant, types.Sorcery, types.Planeswalker, types.Battle,
	} {
		issues := ValidateCardDef(&CardDef{CardFace: CardFace{
			Name:  "Test Creature",
			Types: []types.Card{types.Creature},
			StaticAbilities: []StaticAbility{{
				KeywordAbilities: []KeywordAbility{
					ProtectionKeyword{FromTypes: []types.Card{cardType}},
				},
			}},
		}})
		if hasCardDefIssue(issues, CardDefIssueInvalidKeywordAbility) {
			t.Fatalf("unexpected invalid-keyword-ability for canonical type %q: %+v", cardType, issues)
		}
	}
}

// TestValidateProtectionKeywordRejectsUnknownColor verifies that a
// ProtectionKeyword with an unrecognised color value is rejected.
func TestValidateProtectionKeywordRejectsUnknownColor(t *testing.T) {
	t.Parallel()
	issues := ValidateCardDef(&CardDef{CardFace: CardFace{
		Name:  "Test Creature",
		Types: []types.Card{types.Creature},
		StaticAbilities: []StaticAbility{{
			KeywordAbilities: []KeywordAbility{
				ProtectionKeyword{FromColors: []color.Color{"Purple"}},
			},
		}},
	}})
	if !hasCardDefIssue(issues, CardDefIssueInvalidKeywordAbility) {
		t.Fatalf("expected invalid-keyword-ability for unknown color, got %+v", issues)
	}
}

// TestValidateProtectionKeywordAcceptsAllFiveColors verifies that every
// canonical Magic color passes validation.
func TestValidateProtectionKeywordAcceptsAllFiveColors(t *testing.T) {
	t.Parallel()
	for _, c := range color.AllColors() {
		issues := ValidateCardDef(&CardDef{CardFace: CardFace{
			Name:  "Test Creature",
			Types: []types.Card{types.Creature},
			StaticAbilities: []StaticAbility{{
				KeywordAbilities: []KeywordAbility{
					ProtectionKeyword{FromColors: []color.Color{c}},
				},
			}},
		}})
		if hasCardDefIssue(issues, CardDefIssueInvalidKeywordAbility) {
			t.Fatalf("unexpected invalid-keyword-ability for canonical color %q: %+v", c, issues)
		}
	}
}

func TestValidateCardDefEventHistoryConditionRejectsUnknownEvent(t *testing.T) {
	t.Parallel()
	def := CardDef{
		CardFace: CardFace{
			Name:  "Test Bear",
			Types: []types.Card{types.Creature},
			Power: opt.Val(PT{Value: 2}), Toughness: opt.Val(PT{Value: 2}),
			TriggeredAbilities: []TriggeredAbility{{
				Text: "At the beginning of your upkeep, if you attacked this turn, draw a card.",
				Trigger: TriggerCondition{
					Type: TriggerWhenever,
					Pattern: TriggerPattern{
						Event: EventBeginningOfStep,
						Step:  StepUpkeep,
					},
					InterveningIf: "if you attacked this turn",
					InterveningCondition: opt.Val(Condition{
						EventHistory: opt.Val(EventHistoryCondition{
							Pattern: TriggerPattern{Event: EventUnknown},
							Window:  EventHistoryCurrentTurn,
						}),
					}),
				},
				Content: Mode{Sequence: []Instruction{{Primitive: Draw{
					Amount: Fixed(1), Player: ControllerReference(),
				}}}}.Ability(),
			}},
		},
	}
	issues := ValidateCardDef(&def)
	if !hasCardDefIssue(issues, CardDefIssueInvalidCondition) {
		t.Fatalf("issues = %v, want CardDefIssueInvalidCondition for EventUnknown", issues)
	}
}

func TestValidateCardDefEventHistoryConditionAcceptsValidPattern(t *testing.T) {
	t.Parallel()
	def := CardDef{
		CardFace: CardFace{
			Name:  "Test Bear",
			Types: []types.Card{types.Creature},
			Power: opt.Val(PT{Value: 2}), Toughness: opt.Val(PT{Value: 2}),
			TriggeredAbilities: []TriggeredAbility{{
				Text: "At the beginning of your upkeep, if you attacked this turn, draw a card.",
				Trigger: TriggerCondition{
					Type: TriggerWhen,
					Pattern: TriggerPattern{
						Event: EventBeginningOfStep,
						Step:  StepUpkeep,
					},
					InterveningIf: "if you attacked this turn",
					InterveningCondition: opt.Val(Condition{
						EventHistory: opt.Val(EventHistoryCondition{
							Pattern: TriggerPattern{
								Event:      EventAttackerDeclared,
								Controller: TriggerControllerYou,
							},
							Window: EventHistoryCurrentTurn,
						}),
					}),
				},
				Content: Mode{Sequence: []Instruction{{Primitive: Draw{
					Amount: Fixed(1), Player: ControllerReference(),
				}}}}.Ability(),
			}},
		},
	}
	issues := ValidateCardDef(&def)
	if hasCardDefIssue(issues, CardDefIssueInvalidCondition) {
		t.Fatalf("unexpected CardDefIssueInvalidCondition for valid EventHistory: %v", issues)
	}
}
