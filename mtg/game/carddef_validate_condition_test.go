package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestValidateCardDefChecksStructuredConditionObjectReferences(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name:       "Bad Condition",
		OracleText: "Whenever a creature dies, if it was targeted, draw a card.",
		TriggeredAbilities: []TriggeredAbility{{
			Content: Mode{
				Targets: []TargetSpec{
					{MinTargets: 1, MaxTargets: 1},
				},
			}.Ability(),
			Trigger: TriggerCondition{
				Pattern: TriggerPattern{Event: EventPermanentDied},
				InterveningCondition: opt.Val(Condition{
					Object: opt.Val(TargetPermanentReference(1)),
				}),
			},
		}},
	}}

	issues := ValidateCardDef(card)

	if !hasCardDefIssue(issues, CardDefIssueTargetIndexOutOfRange) {
		t.Fatalf("issues = %+v, want target index issue from structured condition object", issues)
	}

}

func TestValidateCardDefChecksConditionObjectMatches(t *testing.T) {
	makeCard := func(condition Condition) *CardDef {
		return &CardDef{CardFace: CardFace{
			Name:       "Object Condition",
			OracleText: "Whenever a creature dies, if it was a Human, draw a card.",
			TriggeredAbilities: []TriggeredAbility{{
				Content: Mode{}.Ability(),
				Trigger: TriggerCondition{
					Pattern:              TriggerPattern{Event: EventPermanentDied},
					InterveningCondition: opt.Val(condition),
				},
			}},
		}}
	}

	valid := Condition{
		Object: opt.Val(EventPermanentReference()),
		ObjectMatches: opt.Val(Selection{
			RequiredTypes: []types.Card{types.Creature},
			SubtypesAny:   []types.Sub{types.Human},
		}),
	}
	if issues := ValidateCardDef(makeCard(valid)); len(issues) != 0 {
		t.Fatalf("valid object condition issues = %+v", issues)
	}

	missingObject := valid
	missingObject.Object = opt.V[ObjectReference]{}
	if issues := ValidateCardDef(makeCard(missingObject)); !hasCardDefIssue(issues, CardDefIssueInvalidCondition) {
		t.Fatalf("missing-object issues = %+v, want %s", issues, CardDefIssueInvalidCondition)
	}

	dual := valid
	dual.Types = []types.Card{types.Creature}
	if issues := ValidateCardDef(makeCard(dual)); !hasCardDefIssue(issues, CardDefIssueInvalidSelection) {
		t.Fatalf("dual-selection issues = %+v, want %s", issues, CardDefIssueInvalidSelection)
	}

	invalid := valid
	invalid.ObjectMatches = opt.Val(Selection{
		RequiredTypes: []types.Card{types.Creature},
		ExcludedTypes: []types.Card{types.Creature},
	})
	if issues := ValidateCardDef(makeCard(invalid)); !hasCardDefIssue(issues, CardDefIssueInvalidSelection) {
		t.Fatalf("invalid-selection issues = %+v, want %s", issues, CardDefIssueInvalidSelection)
	}
}

func TestValidateCardDefReportsStructurallyInvalidReference(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name:       "Bad Reference",
		OracleText: "Whenever a creature dies, draw a card.",
		TriggeredAbilities: []TriggeredAbility{{
			Content: Mode{}.Ability(),
			Trigger: TriggerCondition{
				Pattern: TriggerPattern{Event: EventPermanentDied},
				InterveningCondition: opt.Val(Condition{
					Object: opt.Val(objectReferenceForTest(ObjectReferenceLinkedObject, 0, "")),
				}),
			},
		}},
	}}

	issues := ValidateCardDef(card)

	if !hasCardDefIssue(issues, CardDefIssueInvalidReference) {
		t.Fatalf("issues = %+v, want invalid-reference issue from structurally invalid object reference", issues)
	}
}

func TestValidateCardDefReportsContradictorySelection(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name:       "Contradictory Selection",
		OracleText: "Destroy target creature.",
		SpellAbility: opt.Val(Mode{
			Targets: []TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      TargetAllowPermanent,
				Selection: opt.Val(Selection{
					RequiredTypes: []types.Card{types.Creature},
					ExcludedTypes: []types.Card{types.Creature},
				}),
			}},
		}.Ability()),
	}}

	issues := ValidateCardDef(card)

	if !hasCardDefIssue(issues, CardDefIssueInvalidSelection) {
		t.Fatalf("issues = %+v, want %s", issues, CardDefIssueInvalidSelection)
	}
}

func TestValidateCardDefReportsInvalidControllerControlsSelection(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name: "Invalid Condition",
		StaticAbilities: []StaticAbility{{
			Condition: opt.Val(Condition{
				ControlsMatching: opt.Val(SelectionCount{
					Selection: Selection{
						ColorsAny:      []color.Color{color.Red},
						ExcludedColors: []color.Color{color.Red},
					},
				}),
			}),
		}},
	}}

	issues := ValidateCardDef(card)

	if !hasCardDefIssue(issues, CardDefIssueInvalidSelection) {
		t.Fatalf("issues = %+v, want %s", issues, CardDefIssueInvalidSelection)
	}
}

func TestValidateCardDefAcceptsConditionalSelfCastFromGraveyard(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name:  "Conditional Vehicle",
		Types: []types.Card{types.Artifact},
		StaticAbilities: []StaticAbility{{
			ZoneOfFunction: zone.Graveyard,
			Condition: opt.Val(Condition{
				ControlsMatching: opt.Val(SelectionCount{
					Selection: Selection{
						SubtypesAny: []types.Sub{types.Pirate, types.Vehicle},
						Tapped:      TriTrue,
					},
					MinCount: 3,
				}),
			}),
			RuleEffects: []RuleEffect{{
				Kind:           RuleEffectCastFromZone,
				AffectedPlayer: PlayerYou,
				CastFromZone:   zone.Graveyard,
				AffectedSource: true,
			}},
		}},
	}}

	if issues := ValidateCardDef(card); len(issues) != 0 {
		t.Fatalf("issues = %+v, want valid conditional graveyard cast permission", issues)
	}
}

func TestValidateCardDefReportsNegativeConditionThresholds(t *testing.T) {
	tests := map[string]Condition{
		"controller life":                     {Aggregates: []AggregateComparison{{Aggregate: AggregateControllerLife, Op: compare.GreaterOrEqual, Value: -1}}},
		"controller life at most":             {Aggregates: []AggregateComparison{{Aggregate: AggregateControllerLife, Op: compare.LessOrEqual, Value: -1}}},
		"controller life above starting":      {Aggregates: []AggregateComparison{{Aggregate: AggregateControllerLifeAboveStarting, Op: compare.GreaterOrEqual, Value: -1}}},
		"any player life":                     {AnyPlayerLifeAtMost: -1},
		"opponent count":                      {Aggregates: []AggregateComparison{{Aggregate: AggregateOpponentCount, Op: compare.GreaterOrEqual, Value: -1}}},
		"controller graveyard cards":          {Aggregates: []AggregateComparison{{Aggregate: AggregateControllerGraveyardCardCount, Op: compare.GreaterOrEqual, Value: -1}}},
		"controller graveyard card types":     {Aggregates: []AggregateComparison{{Aggregate: AggregateControllerGraveyardCardTypeCount, Op: compare.GreaterOrEqual, Value: -1}}},
		"controller basic land types":         {Aggregates: []AggregateComparison{{Aggregate: AggregateControllerBasicLandTypeCount, Op: compare.GreaterOrEqual, Value: -1}}},
		"controller creature power diversity": {Aggregates: []AggregateComparison{{Aggregate: AggregateControllerCreaturePowerDiversity, Op: compare.GreaterOrEqual, Value: -1}}},
	}
	for name, condition := range tests {
		t.Run(name, func(t *testing.T) {
			card := &CardDef{CardFace: CardFace{
				Name:       "Invalid Condition",
				OracleText: "Invalid condition.",
				StaticAbilities: []StaticAbility{{
					Condition: opt.Val(condition),
				}},
			}}

			issues := ValidateCardDef(card)

			if !hasCardDefIssue(issues, CardDefIssueInvalidCondition) {
				t.Fatalf("issues = %+v, want %s", issues, CardDefIssueInvalidCondition)
			}
		})
	}
}

func TestValidateCardDefReportsNegativeConditionPermanentCount(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name:       "Invalid Permanent Count",
		OracleText: "Invalid condition.",
		StaticAbilities: []StaticAbility{{
			Condition: opt.Val(Condition{
				AnyOpponentControls: opt.Val(SelectionCount{MinCount: -1}),
			}),
		}},
	}}

	issues := ValidateCardDef(card)

	if !hasCardDefIssue(issues, CardDefIssueInvalidCondition) {
		t.Fatalf("issues = %+v, want %s", issues, CardDefIssueInvalidCondition)
	}
}

func TestValidateCardDefChecksInstructionSharedCondition(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name:       "Invalid Instruction Condition",
		OracleText: "Draw a card.",
		SpellAbility: opt.Val(Mode{Sequence: []Instruction{{
			Primitive: Draw{Amount: Fixed(1), Player: ControllerReference()},
			Condition: opt.Val(EffectCondition{Condition: opt.Val(Condition{
				Aggregates: []AggregateComparison{{Aggregate: AggregateControllerLife, Op: compare.GreaterOrEqual, Value: -1}},
			})}),
		}}}.Ability()),
	}}

	issues := ValidateCardDef(card)

	if !hasCardDefIssue(issues, CardDefIssueInvalidCondition) {
		t.Fatalf("issues = %+v, want %s", issues, CardDefIssueInvalidCondition)
	}
}
