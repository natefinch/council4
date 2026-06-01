package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

func TestValidateCardReportsOracleWithoutAbilities(t *testing.T) {
	card := &game.CardDef{
		Name:       "Unfinished Card",
		OracleText: "Draw a card.",
	}

	issues := ValidateCard(card, ValidationOptions{})

	if !hasIssue(issues, IssueOracleWithoutAbilities) {
		t.Fatalf("issues = %+v, want %s", issues, IssueOracleWithoutAbilities)
	}
}

func TestValidateCardAllowsOracleWithImplementationID(t *testing.T) {
	card := &game.CardDef{
		Name:             "Implemented Elsewhere",
		OracleText:       "Do something bespoke.",
		ImplementationID: "bespoke",
	}

	issues := ValidateCard(card, ValidationOptions{
		KnownImplementationIDs: map[string]bool{"bespoke": true},
	})

	if len(issues) != 0 {
		t.Fatalf("issues = %+v, want none", issues)
	}
}

func TestValidateCardReportsUnregisteredImplementationID(t *testing.T) {
	card := &game.CardDef{
		Name:             "Missing Implementation",
		OracleText:       "Do something bespoke.",
		ImplementationID: "missing",
	}

	issues := ValidateCard(card, ValidationOptions{
		KnownImplementationIDs: map[string]bool{"other": true},
	})

	if !hasIssue(issues, IssueUnregisteredImplementation) {
		t.Fatalf("issues = %+v, want %s", issues, IssueUnregisteredImplementation)
	}
}

func TestValidateCardReportsImplementationIDWhenRequested(t *testing.T) {
	card := &game.CardDef{
		Name:             "Implemented Elsewhere",
		OracleText:       "Do something bespoke.",
		ImplementationID: "bespoke",
	}

	issues := ValidateCard(card, ValidationOptions{ReportImplementationIDs: true})

	if !hasIssue(issues, IssueImplementationRequired) {
		t.Fatalf("issues = %+v, want %s", issues, IssueImplementationRequired)
	}
}

func TestValidateCardReportsUnexecutedEffect(t *testing.T) {
	card := &game.CardDef{
		Name:       "Unsupported Effect",
		OracleText: "Copy target spell.",
		Abilities: []game.AbilityDef{{
			Kind: game.SpellAbility,
			Targets: []game.TargetSpec{
				{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowStackObject},
			},
			Effects: []game.Effect{{Type: game.EffectCopy, TargetIndex: 0}},
		}},
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
			effect: game.Effect{Type: game.EffectSearch, TargetIndex: -1},
			code:   IssueMissingSearchSpec,
		},
		{
			name: "unsupported destination",
			effect: game.Effect{
				Type:        game.EffectSearch,
				TargetIndex: -1,
				Search: opt.Val(game.SearchSpec{
					SourceZone:  game.ZoneLibrary,
					Destination: game.ZoneExile,
				}),
			},
			code: IssueUnsupportedSearchSpec,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := &game.CardDef{
				Name:       "Search Card",
				OracleText: "Search your library.",
				Abilities: []game.AbilityDef{{
					Kind:    game.SpellAbility,
					Effects: []game.Effect{tt.effect},
				}},
			}

			issues := ValidateCard(card, ValidationOptions{})

			if !hasIssue(issues, tt.code) {
				t.Fatalf("issues = %+v, want %s", issues, tt.code)
			}
		})
	}
}

func TestValidateCardReportsTargetIndexOutOfRange(t *testing.T) {
	card := &game.CardDef{
		Name:       "Bad Target",
		OracleText: "Destroy target creature.",
		Abilities: []game.AbilityDef{{
			Kind:    game.SpellAbility,
			Effects: []game.Effect{{Type: game.EffectDestroy, TargetIndex: 0}},
		}},
	}

	issues := ValidateCard(card, ValidationOptions{})

	if !hasIssue(issues, IssueTargetIndexOutOfRange) {
		t.Fatalf("issues = %+v, want %s", issues, IssueTargetIndexOutOfRange)
	}
}

func TestValidateCardReportsInvalidTargetSpec(t *testing.T) {
	card := &game.CardDef{
		Name:       "Bad Target Spec",
		OracleText: "Destroy up to negative one target creature.",
		Abilities: []game.AbilityDef{{
			Kind: game.SpellAbility,
			Targets: []game.TargetSpec{
				{MinTargets: 2, MaxTargets: 1},
			},
			Effects: []game.Effect{{Type: game.EffectDestroy, TargetIndex: 0}},
		}},
	}

	issues := ValidateCard(card, ValidationOptions{})

	if !hasIssue(issues, IssueInvalidTargetSpec) {
		t.Fatalf("issues = %+v, want %s", issues, IssueInvalidTargetSpec)
	}
}

func TestValidateCardChecksFaces(t *testing.T) {
	card := &game.CardDef{
		Name: "Double Faced",
		Faces: []game.CardFace{{
			Name:       "Front",
			OracleText: "Draw a card.",
		}},
	}

	issues := ValidateCard(card, ValidationOptions{})

	if !hasIssue(issues, IssueOracleWithoutAbilities) {
		t.Fatalf("issues = %+v, want face oracle issue", issues)
	}
}

func TestValidateCardChecksDoubleFacedRootFieldsWithoutWalkingRootEffects(t *testing.T) {
	card := &game.CardDef{
		Name:             "Double Faced",
		OracleText:       "Root text.",
		ImplementationID: "missing",
		Abilities: []game.AbilityDef{{
			Kind:    game.SpellAbility,
			Effects: []game.Effect{{Type: game.EffectCopy}},
		}},
		Faces: []game.CardFace{{
			Name:       "Front",
			OracleText: "Draw a card.",
			Abilities: []game.AbilityDef{{
				Kind:    game.SpellAbility,
				Effects: []game.Effect{{Type: game.EffectDraw, TargetIndex: -1}},
			}},
		}},
	}

	issues := ValidateCard(card, ValidationOptions{
		KnownImplementationIDs: map[string]bool{"other": true},
	})

	if !hasIssue(issues, IssueUnregisteredImplementation) {
		t.Fatalf("issues = %+v, want root implementation issue", issues)
	}
	if hasIssue(issues, IssueUnexecutedEffect) {
		t.Fatalf("issues = %+v, did not want duplicate root ability walk for DFC", issues)
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
				Kind:          game.SpellAbility,
				KickerEffects: []game.Effect{{Type: game.EffectCopy}},
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
					Token: opt.Val(&game.CardDef{
						Name: "Unsupported Token",
						Abilities: []game.AbilityDef{{
							Kind:    game.ActivatedAbility,
							Effects: []game.Effect{{Type: game.EffectCopy}},
						}},
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
			card := &game.CardDef{
				Name:       "Nested Effects",
				OracleText: "Nested unsupported effect.",
				Abilities:  []game.AbilityDef{tt.ability},
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
			card := &game.CardDef{
				Name:       "Nested Target",
				OracleText: "Use target data.",
				Abilities: []game.AbilityDef{{
					Kind: game.SpellAbility,
					Targets: []game.TargetSpec{
						{MinTargets: 1, MaxTargets: 1},
					},
					Effects: []game.Effect{tt.effect},
				}},
			}

			issues := ValidateCard(card, ValidationOptions{})

			if !hasIssue(issues, IssueTargetIndexOutOfRange) {
				t.Fatalf("issues = %+v, want target index issue", issues)
			}
		})
	}
}

func TestValidateCardChecksEnchantTargetSpec(t *testing.T) {
	card := &game.CardDef{
		Name:       "Bad Aura",
		OracleText: "Enchant creature",
		Abilities: []game.AbilityDef{{
			Kind: game.StaticAbility,
			EnchantTarget: opt.Val(game.TargetSpec{
				MinTargets: 2,
				MaxTargets: 1,
			}),
		}},
	}

	issues := ValidateCard(card, ValidationOptions{})

	if !hasIssue(issues, IssueInvalidTargetSpec) {
		t.Fatalf("issues = %+v, want enchant target spec issue", issues)
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
