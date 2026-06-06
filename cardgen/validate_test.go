package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
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

func TestValidateCardReportsUnexecutedEffect(t *testing.T) {
	card := &game.CardDef{CardFace: game.CardFace{Name: "Unsupported Effect",
		OracleText: "Copy target spell.",
		Abilities: []game.AbilityDef{{
			Kind: game.SpellAbility,
			Targets: []game.TargetSpec{
				{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowStackObject},
			},
			Effects: []game.Effect{{Type: game.EffectCopy, TargetIndex: 0}},
		}}},
	}

	issues := ValidateCard(card, ValidationOptions{})

	if !hasIssue(issues, IssueUnexecutedEffect) {
		t.Fatalf("issues = %+v, want %s", issues, IssueUnexecutedEffect)
	}
}

func TestValidateCardReportsSearchSpecProblems(t *testing.T) {
	tests := []struct {
		name   string
		effect game.Effect
		code   ValidationCode
	}{
		{
			name:   "missing spec",
			effect: game.Effect{Type: game.EffectSearch, TargetIndex: game.TargetIndexController},
			code:   IssueMissingSearchSpec,
		},
		{
			name: "unsupported destination",
			effect: game.Effect{
				Type:        game.EffectSearch,
				TargetIndex: game.TargetIndexController,
				Search: opt.Val(game.SearchSpec{
					SourceZone:  game.ZoneLibrary,
					Destination: game.ZoneExile,
				}),
			},
			code: IssueUnsupportedSearchSpec,
		},
		{
			name: "missing supertype",
			effect: game.Effect{
				Type:        game.EffectSearch,
				TargetIndex: game.TargetIndexController,
				Search: opt.Val(game.SearchSpec{
					SourceZone:  game.ZoneLibrary,
					Destination: game.ZoneHand,
					Supertype:   opt.Val(types.Super("")),
				}),
			},
			code: IssueUnsupportedSearchSpec,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := &game.CardDef{CardFace: game.CardFace{Name: "Search Card",
				OracleText: "Search your library.",
				Abilities: []game.AbilityDef{{
					Kind:    game.SpellAbility,
					Effects: []game.Effect{tt.effect},
				}}},
			}

			issues := ValidateCard(card, ValidationOptions{})

			if !hasIssue(issues, tt.code) {
				t.Fatalf("issues = %+v, want %s", issues, tt.code)
			}
		})
	}
}

func TestValidateCardReportsTargetIndexOutOfRange(t *testing.T) {
	card := &game.CardDef{CardFace: game.CardFace{Name: "Bad Target",
		OracleText: "Destroy target creature.",
		Abilities: []game.AbilityDef{{
			Kind:    game.SpellAbility,
			Effects: []game.Effect{{Type: game.EffectDestroy, TargetIndex: 0}},
		}}},
	}

	issues := ValidateCard(card, ValidationOptions{})

	if !hasIssue(issues, IssueTargetIndexOutOfRange) {
		t.Fatalf("issues = %+v, want %s", issues, IssueTargetIndexOutOfRange)
	}
}

func TestValidateCardReportsTypedInstructionTargetIndexOutOfRange(t *testing.T) {
	card := &game.CardDef{CardFace: game.CardFace{
		Name:       "Bad Typed Target",
		OracleText: "Destroy target creature.",
		SpellAbility: opt.Val(game.SpellAbilityBody{
			Content: game.PlainAbilityContent{
				Targets: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1}},
				Sequence: []game.Instruction{{
					Primitive: game.Destroy{TargetIndex: 1},
				}},
			},
		}),
	}}

	issues := ValidateCard(card, ValidationOptions{})

	if !hasIssue(issues, IssueInvalidAbilityBody) {
		t.Fatalf("issues = %+v, want %s", issues, IssueInvalidAbilityBody)
	}
}

func TestValidateCardRejectsLegacyEffectsInCategorizedBody(t *testing.T) {
	card := &game.CardDef{CardFace: game.CardFace{
		Name:       "Legacy Body",
		OracleText: "Draw a card.",
		SpellAbility: opt.Val(game.SpellAbilityBody{
			Content: game.PlainAbilityContent{
				LegacyEffects: []game.Effect{{
					Type:        game.EffectDraw,
					Amount:      1,
					TargetIndex: game.TargetIndexController,
				}},
			},
		}),
	}}

	issues := ValidateCard(card, ValidationOptions{})

	if !hasIssue(issues, IssueLegacyEffectConfiguration) {
		t.Fatalf("issues = %+v, want %s", issues, IssueLegacyEffectConfiguration)
	}
}

func TestValidateCardReportsInvalidTargetSpec(t *testing.T) {
	card := &game.CardDef{CardFace: game.CardFace{Name: "Bad Target Spec",
		OracleText: "Destroy up to negative one target creature.",
		Abilities: []game.AbilityDef{{
			Kind: game.SpellAbility,
			Targets: []game.TargetSpec{
				{MinTargets: 2, MaxTargets: 1},
			},
			Effects: []game.Effect{{Type: game.EffectDestroy, TargetIndex: 0}},
		}}},
	}

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
				Abilities: []game.AbilityDef{{
					Kind:    game.SpellAbility,
					Targets: []game.TargetSpec{tt.spec},
					Effects: []game.Effect{{Type: game.EffectTap, TargetIndex: 0}},
				}}},
			}

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

func TestValidateCardChecksDoubleFacedRootFieldsAndBack(t *testing.T) {
	card := &game.CardDef{CardFace: game.CardFace{Name: "Double Faced",
		OracleText: "Root text.",
		Abilities: []game.AbilityDef{{
			Kind:    game.SpellAbility,
			Effects: []game.Effect{{Type: game.EffectCopy}},
		}}}, Back: opt.Val(game.CardFace{
		Name:       "Back",
		OracleText: "Draw a card.",
		Abilities: []game.AbilityDef{{
			Kind:    game.SpellAbility,
			Effects: []game.Effect{{Type: game.EffectDraw, TargetIndex: game.TargetIndexController}},
		}},
	}),
	}

	issues := ValidateCard(card, ValidationOptions{})

	if !hasIssue(issues, IssueUnexecutedEffect) {
		t.Fatalf("issues = %+v, want root ability walk for DFC front face", issues)
	}
}

func TestValidateCardWalksNestedEffects(t *testing.T) {
	tests := []struct {
		name    string
		ability game.AbilityDef
	}{
		{
			name: "kicker effects",
			ability: game.AbilityDef{
				Kind:             game.SpellAbility,
				KeywordAbilities: []game.KeywordAbility{game.KickerKeyword{Cost: cost.Mana{cost.G}, Bonus: []game.Effect{{Type: game.EffectCopy}}}},
			},
		},
		{
			name: "delayed trigger effects",
			ability: game.AbilityDef{
				Kind: game.SpellAbility,
				Effects: []game.Effect{{
					Type: game.EffectCreateDelayedTrigger,
					DelayedTrigger: opt.Val(game.DelayedTriggerDef{
						Effects: []game.Effect{{Type: game.EffectCopy}},
					}),
				}},
			},
		},
		{
			name: "token abilities",
			ability: game.AbilityDef{
				Kind: game.SpellAbility,
				Effects: []game.Effect{{
					Type: game.EffectCreateToken,
					Token: opt.Val(&game.CardDef{CardFace: game.CardFace{Name: "Unsupported Token",
						Abilities: []game.AbilityDef{{
							Kind:    game.ActivatedAbility,
							Effects: []game.Effect{{Type: game.EffectCopy}},
						}}},
					}),
				}},
			},
		},
		{
			name: "emblem abilities",
			ability: game.AbilityDef{
				Kind: game.SpellAbility,
				Effects: []game.Effect{{
					Type: game.EffectCreateEmblem,
					EmblemAbilities: []game.AbilityDef{{
						Kind:    game.StaticAbility,
						Effects: []game.Effect{{Type: game.EffectCopy}},
					}},
				}},
			},
		},
		{
			name: "continuous effect add abilities",
			ability: game.AbilityDef{
				Kind: game.SpellAbility,
				Effects: []game.Effect{{
					Type: game.EffectApplyContinuous,
					ContinuousEffects: []game.ContinuousEffect{{
						AddAbilities: []game.AbilityDef{{
							Kind:    game.StaticAbility,
							Effects: []game.Effect{{Type: game.EffectCopy}},
						}},
					}},
				}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := &game.CardDef{CardFace: game.CardFace{Name: "Nested Effects",
				OracleText: "Nested unsupported effect.",
				Abilities:  []game.AbilityDef{tt.ability}},
			}

			issues := ValidateCard(card, ValidationOptions{})

			if !hasIssue(issues, IssueUnexecutedEffect) {
				t.Fatalf("issues = %+v, want nested unexecuted effect", issues)
			}
		})
	}
}

func TestValidateCardChecksNestedTargetIndexes(t *testing.T) {
	tests := []struct {
		name   string
		effect game.Effect
	}{
		{
			name: "condition",
			effect: game.Effect{
				Type: game.EffectDestroy,
				Condition: opt.Val(game.EffectCondition{
					TargetIndex: 1,
				}),
			},
		},
		{
			name: "dynamic amount",
			effect: game.Effect{
				Type: game.EffectDamage,
				DynamicAmount: opt.Val(game.DynamicAmount{
					Kind:        game.DynamicAmountTargetPower,
					TargetIndex: 1,
				}),
			},
		},
		{
			name: "counter source",
			effect: game.Effect{
				Type: game.EffectMoveCounters,
				CounterSource: game.CounterSourceSpec{
					Kind:        game.CounterSourceTarget,
					TargetIndex: 1,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := &game.CardDef{CardFace: game.CardFace{Name: "Nested Target",
				OracleText: "Use target data.",
				Abilities: []game.AbilityDef{{
					Kind: game.SpellAbility,
					Targets: []game.TargetSpec{
						{MinTargets: 1, MaxTargets: 1},
					},
					Effects: []game.Effect{tt.effect},
				}}},
			}

			issues := ValidateCard(card, ValidationOptions{})

			if !hasIssue(issues, IssueTargetIndexOutOfRange) {
				t.Fatalf("issues = %+v, want target index issue", issues)
			}
		})
	}
}

func TestValidateCardChecksStructuredConditionObjectReferences(t *testing.T) {
	card := &game.CardDef{CardFace: game.CardFace{Name: "Bad Condition",
		OracleText: "Whenever a creature dies, if it was targeted, draw a card.",
		Abilities: []game.AbilityDef{{
			Kind: game.TriggeredAbility,
			Targets: []game.TargetSpec{
				{MinTargets: 1, MaxTargets: 1},
			},
			Trigger: opt.Val(game.TriggerCondition{
				Pattern: game.TriggerPattern{Event: game.EventPermanentDied},
				InterveningCondition: opt.Val(game.Condition{
					Object: opt.Val(game.ObjectReference{
						Kind:        game.ObjectReferenceTargetPermanent,
						TargetIndex: 1,
					}),
				}),
			}),
			Effects: []game.Effect{{Type: game.EffectDraw, TargetIndex: game.TargetIndexController}},
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
		Abilities: []game.AbilityDef{{
			Kind: game.StaticAbility,
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
		Abilities: []game.AbilityDef{{
			Kind: game.StaticAbility,
			Effects: []game.Effect{{
				Type: game.EffectApplyContinuous,
				ContinuousEffects: []game.ContinuousEffect{{
					Layer:       game.LayerAbility,
					Selector:    game.EffectSelectorCreaturesYouControl,
					AddKeywords: []game.Keyword{game.Haste},
				}},
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
			name: "kicker bonus effect",
			ability: game.KickerKeyword{
				Cost:  cost.Mana{cost.G},
				Bonus: []game.Effect{{Type: game.EffectCopy}},
			},
			code: IssueUnexecutedEffect,
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
				Abilities: []game.AbilityDef{{
					Kind:             game.StaticAbility,
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
		body game.AbilityBody
		code ValidationCode
	}{
		{
			name: "plain content target index",
			body: game.SpellAbilityBody{
				Content: game.PlainAbilityContent{
					LegacyEffects: []game.Effect{{Type: game.EffectDestroy, TargetIndex: 0}},
				},
			},
			code: IssueTargetIndexOutOfRange,
		},
		{
			name: "nested modal effect",
			body: game.SpellAbilityBody{
				Content: game.ModalAbilityContent{
					Modes: []game.Mode{{
						LegacyEffects: []game.Effect{{Type: game.EffectCopy}},
					}},
				},
			},
			code: IssueUnexecutedEffect,
		},
		{
			name: "static keyword",
			body: game.StaticAbilityBody{
				KeywordAbilities: []game.KeywordAbility{game.SimpleKeyword{}},
			},
			code: IssueInvalidKeywordAbility,
		},
		{
			name: "trigger intervening condition",
			body: game.TriggeredAbilityBody{
				Trigger: game.TriggerCondition{
					InterveningCondition: opt.Val(game.Condition{
						Object: opt.Val(game.ObjectReference{
							Kind:        game.ObjectReferenceTargetPermanent,
							TargetIndex: 1,
						}),
					}),
				},
				Content: game.PlainAbilityContent{},
			},
			code: IssueTargetIndexOutOfRange,
		},
		{
			name: "nil content",
			body: game.SpellAbilityBody{},
			code: IssueInvalidAbilityBody,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := &game.CardDef{CardFace: game.CardFace{Name: "Body Card",
				OracleText: "Body ability.",
				Abilities: []game.AbilityDef{{
					Body: tt.body,
				}}},
			}

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
