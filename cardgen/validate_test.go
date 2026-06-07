package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestValidateCardReportsOracleWithoutAbilities(t *testing.T) {
	card := &game.CardDef{CardFace: game.CardFace{Name: "Unfinished Card",
		OracleText: "Draw a card."},
	}

	issues := ValidateCard(card, ValidationOptions{})

	if !hasIssue(issues, IssueOracleWithoutAbilities) {
		t.Fatalf("issues = %+v, want %s", issues, IssueOracleWithoutAbilities)
	}
}

func TestValidateCardAllowsOracleWithImplementationID(t *testing.T) {
	card := &game.CardDef{CardFace: game.CardFace{Name: "Implemented Elsewhere",
		OracleText:       "Do something bespoke.",
		ImplementationID: "bespoke"},
	}

	issues := ValidateCard(card, ValidationOptions{
		KnownImplementationIDs: map[string]bool{"bespoke": true},
	})

	if len(issues) != 0 {
		t.Fatalf("issues = %+v, want none", issues)
	}
}

func TestValidateCardReportsUnregisteredImplementationID(t *testing.T) {
	card := &game.CardDef{CardFace: game.CardFace{Name: "Missing Implementation",
		OracleText:       "Do something bespoke.",
		ImplementationID: "missing"},
	}

	issues := ValidateCard(card, ValidationOptions{
		KnownImplementationIDs: map[string]bool{"other": true},
	})

	if !hasIssue(issues, IssueUnregisteredImplementation) {
		t.Fatalf("issues = %+v, want %s", issues, IssueUnregisteredImplementation)
	}
}

func TestValidateCardReportsImplementationIDWhenRequested(t *testing.T) {
	card := &game.CardDef{CardFace: game.CardFace{Name: "Implemented Elsewhere",
		OracleText:       "Do something bespoke.",
		ImplementationID: "bespoke"},
	}

	issues := ValidateCard(card, ValidationOptions{ReportImplementationIDs: true})

	if !hasIssue(issues, IssueImplementationRequired) {
		t.Fatalf("issues = %+v, want %s", issues, IssueImplementationRequired)
	}
}

func TestValidateCardReportsTypedInstructionTargetIndexOutOfRange(t *testing.T) {
	card := &game.CardDef{CardFace: game.CardFace{
		Name:       "Bad Typed Target",
		OracleText: "Destroy target creature.",
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1}},
			Sequence: []game.Instruction{{
				Primitive: game.Destroy{TargetIndex: 1},
			}},
		}.Ability()),
	}}

	issues := ValidateCard(card, ValidationOptions{})

	if !hasIssue(issues, IssueInvalidAbilityBody) {
		t.Fatalf("issues = %+v, want %s", issues, IssueInvalidAbilityBody)
	}
}

func TestValidateCardReportsTypedSearchProblems(t *testing.T) {
	tests := []struct {
		name string
		spec game.SearchSpec
	}{
		{name: "missing zones"},
		{
			name: "unsupported destination",
			spec: game.SearchSpec{SourceZone: zone.Library, Destination: zone.Exile},
		},
		{
			name: "empty supertype",
			spec: game.SearchSpec{
				SourceZone:  zone.Library,
				Destination: zone.Hand,
				Supertype:   opt.Val(types.Super("")),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := &game.CardDef{CardFace: game.CardFace{
				Name:       "Bad Search",
				OracleText: "Search your library.",
				SpellAbility: opt.Val(game.Mode{
					Sequence: []game.Instruction{{
						Primitive: game.Search{
							Amount:      game.Fixed(1),
							TargetIndex: game.TargetIndexController,
							Spec:        tt.spec,
						},
					}},
				}.Ability()),
			}}

			issues := ValidateCard(card, ValidationOptions{})
			if !hasIssue(issues, IssueInvalidAbilityBody) {
				t.Fatalf("issues = %+v, want %s", issues, IssueInvalidAbilityBody)
			}
		})
	}
}

func TestValidateCardChecksDelayedTriggerContent(t *testing.T) {
	card := &game.CardDef{CardFace: game.CardFace{
		Name:       "Bad Delayed Trigger",
		OracleText: "At the beginning of the next end step, destroy target creature.",
		SpellAbility: opt.Val(game.Mode{
			Sequence: []game.Instruction{{
				Primitive: game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
					Timing: game.DelayedAtBeginningOfNextEndStep,
					Content: game.Mode{
						Sequence: []game.Instruction{{
							Primitive: game.Destroy{TargetIndex: 0},
						}},
					}.Ability(),
				}},
			}},
		}.Ability()),
	}}

	issues := ValidateCard(card, ValidationOptions{})
	if !hasIssue(issues, IssueInvalidAbilityBody) {
		t.Fatalf("issues = %+v, want %s", issues, IssueInvalidAbilityBody)
	}
}

func TestValidateCardReportsInvalidTargetSpec(t *testing.T) {
	card := &game.CardDef{CardFace: game.CardFace{Name: "Bad Target Spec",
		OracleText: "Destroy up to negative one target creature.",
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{
				{MinTargets: 2, MaxTargets: 1},
			},
		}.Ability()),
	}}

	issues := ValidateCard(card, ValidationOptions{})

	if !hasIssue(issues, IssueInvalidTargetSpec) {
		t.Fatalf("issues = %+v, want %s", issues, IssueInvalidTargetSpec)
	}
}

func TestValidateCardReportsInvalidTargetChooserSpec(t *testing.T) {
	tests := []struct {
		name string
		spec game.TargetSpec
	}{
		{
			name: "opponent chooser with optional count",
			spec: game.TargetSpec{MinTargets: 0, MaxTargets: 1, Chooser: game.TargetChooserOpponent},
		},
		{
			name: "opponent chooser with opponent-relative controller predicate",
			spec: game.TargetSpec{
				MinTargets: 1,
				MaxTargets: 1,
				Chooser:    game.TargetChooserOpponent,
				Predicate:  game.TargetPredicate{Controller: game.ControllerOpponent},
			},
		},
		{
			name: "unknown chooser",
			spec: game.TargetSpec{MinTargets: 1, MaxTargets: 1, Chooser: game.TargetChooser(99)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := &game.CardDef{CardFace: game.CardFace{Name: "Bad Target Chooser",
				OracleText: "Tap target creature.",
				SpellAbility: opt.Val(game.Mode{
					Targets: []game.TargetSpec{tt.spec},
				}.Ability()),
			}}

			issues := ValidateCard(card, ValidationOptions{})

			if !hasIssue(issues, IssueInvalidTargetSpec) {
				t.Fatalf("issues = %+v, want %s", issues, IssueInvalidTargetSpec)
			}
		})
	}
}

func TestValidateCardChecksFaces(t *testing.T) {
	card := &game.CardDef{CardFace: game.CardFace{Name: "Front"}, Back: opt.Val(game.CardFace{
		Name:       "Front",
		OracleText: "Draw a card.",
	}),
	}

	issues := ValidateCard(card, ValidationOptions{})

	if !hasIssue(issues, IssueOracleWithoutAbilities) {
		t.Fatalf("issues = %+v, want face oracle issue", issues)
	}
}

func TestValidateCardChecksAlternateFace(t *testing.T) {
	card := &game.CardDef{
		CardFace: game.CardFace{Name: "Main Spell"},
		Alternate: opt.Val(game.CardFace{
			Name:       "Alternate Spell",
			OracleText: "Draw a card.",
		}),
	}

	issues := ValidateCard(card, ValidationOptions{})

	if !hasIssue(issues, IssueOracleWithoutAbilities) {
		t.Fatalf("issues = %+v, want alternate face oracle issue", issues)
	}
}

func TestValidateCardChecksDoubleFacedRootFieldsAndBack(t *testing.T) {
	card := &game.CardDef{CardFace: game.CardFace{Name: "Double Faced",
		OracleText: "Root text.",
		SpellAbility: opt.Val(game.Mode{
			Sequence: []game.Instruction{{Primitive: game.Destroy{TargetIndex: 1}}},
		}.Ability())}, Back: opt.Val(game.CardFace{
		Name:       "Back",
		OracleText: "Draw a card.",
		SpellAbility: opt.Val(game.Mode{
			Sequence: []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), TargetIndex: game.TargetIndexController}}},
		}.Ability()),
	}),
	}

	issues := ValidateCard(card, ValidationOptions{})

	if !hasIssue(issues, IssueInvalidAbilityBody) {
		t.Fatalf("issues = %+v, want root ability walk for DFC front face", issues)
	}
}

func TestValidateCardChecksStructuredConditionObjectReferences(t *testing.T) {
	card := &game.CardDef{CardFace: game.CardFace{Name: "Bad Condition",
		OracleText: "Whenever a creature dies, if it was targeted, draw a card.",
		TriggeredAbilities: []game.TriggeredAbilityBody{{
			Content: game.Mode{
				Targets: []game.TargetSpec{
					{MinTargets: 1, MaxTargets: 1},
				},
			}.Ability(),

			Trigger: game.TriggerCondition{
				Pattern: game.TriggerPattern{Event: game.EventPermanentDied},
				InterveningCondition: opt.Val(game.Condition{
					Object: opt.Val(game.ObjectReference{
						Kind:        game.ObjectReferenceTargetPermanent,
						TargetIndex: 1,
					}),
				}),
			},
		}}},
	}

	issues := ValidateCard(card, ValidationOptions{})

	if !hasIssue(issues, IssueTargetIndexOutOfRange) {
		t.Fatalf("issues = %+v, want target index issue from structured condition object", issues)
	}
}

func TestValidateCardChecksEnchantTargetSpec(t *testing.T) {
	card := &game.CardDef{CardFace: game.CardFace{Name: "Bad Aura",
		OracleText: "Enchant creature",
		StaticAbilities: []game.StaticAbilityBody{{
			KeywordAbilities: []game.KeywordAbility{game.EnchantKeyword{Target: game.TargetSpec{
				MinTargets: 2,
				MaxTargets: 1,
			}}},
		}}},
	}

	issues := ValidateCard(card, ValidationOptions{})

	if !hasIssue(issues, IssueInvalidTargetSpec) {
		t.Fatalf("issues = %+v, want enchant target spec issue", issues)
	}
}

func TestValidateCardAllowsSelectorOnlyContinuousEffects(t *testing.T) {
	card := &game.CardDef{CardFace: game.CardFace{Name: "Static Haste",
		OracleText: "Creatures you control have haste.",
		StaticAbilities: []game.StaticAbilityBody{{
			ContinuousEffects: []game.ContinuousEffect{{
				Layer:       game.LayerAbility,
				Selector:    game.EffectSelectorCreaturesYouControl,
				AddKeywords: []game.Keyword{game.Haste},
			}},
		}}},
	}

	issues := ValidateCard(card, ValidationOptions{})

	if len(issues) != 0 {
		t.Fatalf("issues = %+v, want none", issues)
	}
}

func TestValidateCardChecksKeywordAbilities(t *testing.T) {
	tests := []struct {
		name    string
		ability game.KeywordAbility
		code    ValidationCode
	}{
		{
			name: "nil keyword",
			code: IssueInvalidKeywordAbility,
		},
		{
			name:    "simple keyword without kind",
			ability: game.SimpleKeyword{},
			code:    IssueInvalidKeywordAbility,
		},
		{
			name:    "mana keyword without explicit cost",
			ability: game.WardKeyword{},
			code:    IssueInvalidKeywordAbility,
		},
		{
			name: "kicker bonus content",
			ability: game.KickerKeyword{
				Cost: cost.Mana{cost.G},
				BonusContent: game.Mode{
					Sequence: []game.Instruction{{Primitive: game.Destroy{TargetIndex: 0}}},
				}.Ability(),
			},
			code: IssueInvalidAbilityBody,
		},
		{
			name: "enchant target spec",
			ability: game.EnchantKeyword{Target: game.TargetSpec{
				MinTargets: 2,
				MaxTargets: 1,
			}},
			code: IssueInvalidTargetSpec,
		},
		{
			name:    "suspend counters",
			ability: game.SuspendKeyword{Cost: cost.Mana{cost.G}},
			code:    IssueInvalidKeywordAbility,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := &game.CardDef{CardFace: game.CardFace{Name: "Keyword Card",
				OracleText: "Keyword ability.",
				StaticAbilities: []game.StaticAbilityBody{{
					KeywordAbilities: []game.KeywordAbility{tt.ability},
				}}},
			}

			issues := ValidateCard(card, ValidationOptions{})

			if !hasIssue(issues, tt.code) {
				t.Fatalf("issues = %+v, want %s", issues, tt.code)
			}
		})
	}
}

func TestValidateCardChecksAbilityBodies(t *testing.T) {
	tests := []struct {
		name string
		face game.CardFace
		code ValidationCode
	}{
		{
			name: "plain content target index",
			face: game.CardFace{SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{{Primitive: game.Destroy{TargetIndex: 0}}},
			}.Ability())},
			code: IssueInvalidAbilityBody,
		},
		{
			name: "nested modal effect",
			face: game.CardFace{SpellAbility: opt.Val(game.ModalAbilityContent{
				Modes: []game.Mode{{
					Sequence: []game.Instruction{{Primitive: game.Destroy{TargetIndex: 0}}},
				}},
			})},
			code: IssueInvalidAbilityBody,
		},
		{
			name: "static keyword",
			face: game.CardFace{StaticAbilities: []game.StaticAbilityBody{{
				KeywordAbilities: []game.KeywordAbility{game.SimpleKeyword{}},
			}}},
			code: IssueInvalidKeywordAbility,
		},
		{
			name: "activated keyword",
			face: game.CardFace{ActivatedAbilities: []game.ActivatedAbilityBody{{
				KeywordAbilities: []game.KeywordAbility{game.SimpleKeyword{}},
				Content:          game.Mode{}.Ability(),
			}}},
			code: IssueInvalidKeywordAbility,
		},
		{
			name: "triggered keyword",
			face: game.CardFace{TriggeredAbilities: []game.TriggeredAbilityBody{{
				KeywordAbilities: []game.KeywordAbility{game.WardKeyword{}},
				Content:          game.Mode{}.Ability(),
			}}},
			code: IssueInvalidKeywordAbility,
		},
		{
			name: "trigger intervening condition",
			face: game.CardFace{TriggeredAbilities: []game.TriggeredAbilityBody{{
				Trigger: game.TriggerCondition{
					InterveningCondition: opt.Val(game.Condition{
						Object: opt.Val(game.ObjectReference{
							Kind:        game.ObjectReferenceTargetPermanent,
							TargetIndex: 1,
						}),
					}),
				},
				Content: game.Mode{}.Ability(),
			}}},
			code: IssueTargetIndexOutOfRange,
		},
		{
			name: "nil content",
			face: game.CardFace{SpellAbility: opt.Val(game.ModalAbilityContent{})},
			code: IssueInvalidAbilityBody,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			face := tt.face
			face.Name = "Body Card"
			face.OracleText = "Body ability."
			card := &game.CardDef{CardFace: face}

			issues := ValidateCard(card, ValidationOptions{})

			if !hasIssue(issues, tt.code) {
				t.Fatalf("issues = %+v, want %s", issues, tt.code)
			}
		})
	}
}

func hasIssue(issues []ValidationIssue, code ValidationCode) bool {
	for _, issue := range issues {
		if issue.Code == code {
			return true
		}
	}
	return false
}
