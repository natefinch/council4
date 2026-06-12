package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestValidateCardDefReportsNilCard(t *testing.T) {
	issues := ValidateCardDef(nil)
	if !hasCardDefIssue(issues, CardDefIssueNilCard) {
		t.Fatalf("issues = %+v, want %s", issues, CardDefIssueNilCard)
	}
}

func TestValidateCardDefReportsMissingName(t *testing.T) {
	card := &CardDef{CardFace: CardFace{Name: "   "}}
	issues := ValidateCardDef(card)
	if !hasCardDefIssue(issues, CardDefIssueMissingName) {
		t.Fatalf("issues = %+v, want %s", issues, CardDefIssueMissingName)
	}
}

func TestValidateCardDefReportsOracleWithoutAbilities(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name:       "Unfinished Card",
		OracleText: "Draw a card.",
	}}

	issues := ValidateCardDef(card)

	if !hasCardDefIssue(issues, CardDefIssueOracleWithoutAbilities) {
		t.Fatalf("issues = %+v, want %s", issues, CardDefIssueOracleWithoutAbilities)
	}
}

func TestValidateCardDefAllowsOracleWithImplementationID(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name:             "Implemented Elsewhere",
		OracleText:       "Do something bespoke.",
		ImplementationID: "bespoke",
	}}

	issues := ValidateCardDef(card)

	if len(issues) != 0 {
		t.Fatalf("issues = %+v, want none", issues)
	}
}

func TestValidateCardDefReportsTypedInstructionTargetIndexOutOfRange(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name:       "Bad Typed Target",
		OracleText: "Destroy target creature.",
		SpellAbility: opt.Val(Mode{
			Targets: []TargetSpec{{MinTargets: 1, MaxTargets: 1}},
			Sequence: []Instruction{{
				Primitive: Destroy{Object: TargetPermanentReference(1)},
			}},
		}.Ability()),
	}}

	issues := ValidateCardDef(card)

	if !hasCardDefIssue(issues, CardDefIssueInvalidAbilityBody) {
		t.Fatalf("issues = %+v, want %s", issues, CardDefIssueInvalidAbilityBody)
	}
}

func TestValidateCardDefReportsPutOnBattlefieldTargetCardWithoutCardTargetSpec(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name:       "Bad Reanimation",
		OracleText: "Return target creature card from your graveyard to the battlefield.",
		SpellAbility: opt.Val(Mode{
			Sequence: []Instruction{{
				Primitive: PutOnBattlefield{
					Source: CardBattlefieldSource(CardReference{Kind: CardReferenceTarget}),
				},
			}},
		}.Ability()),
	}}

	issues := ValidateCardDef(card)

	if !hasCardDefIssue(issues, CardDefIssueInvalidAbilityBody) {
		t.Fatalf("issues = %+v, want %s", issues, CardDefIssueInvalidAbilityBody)
	}
}

func TestValidateCardDefAllowsPutOnBattlefieldTargetCardWithCardTargetSpec(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name:       "Good Reanimation",
		OracleText: "Return target creature card from your graveyard to the battlefield.",
		SpellAbility: opt.Val(Mode{
			Targets: []TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      TargetAllowCard,
				TargetZone: zone.Graveyard,
				Selection:  opt.Val(Selection{RequiredTypes: []types.Card{types.Creature}, Controller: ControllerYou}),
			}},
			Sequence: []Instruction{{
				Primitive: PutOnBattlefield{
					Source: CardBattlefieldSource(CardReference{Kind: CardReferenceTarget}),
				},
			}},
		}.Ability()),
	}}

	issues := ValidateCardDef(card)

	if hasCardDefIssue(issues, CardDefIssueInvalidAbilityBody) {
		t.Fatalf("issues = %+v, want no %s", issues, CardDefIssueInvalidAbilityBody)
	}
}

func TestValidateInstructionSequenceAttachedPermanentReferenceBounds(t *testing.T) {
	targets := []TargetSpec{{MinTargets: 1, MaxTargets: 1}}

	inRange := []Instruction{{
		Primitive: ModifyPT{Object: TargetAttachedPermanentReference(0), PowerDelta: Fixed(1)},
	}}
	if err := ValidateInstructionSequence(inRange, targets); err != nil {
		t.Fatalf("in-range attached reference: ValidateInstructionSequence() = %v, want nil", err)
	}

	sourceDerived := []Instruction{{
		Primitive: ModifyPT{Object: SourceAttachedPermanentReference(), PowerDelta: Fixed(1)},
	}}
	if err := ValidateInstructionSequence(sourceDerived, targets); err != nil {
		t.Fatalf("source-attached reference: ValidateInstructionSequence() = %v, want nil", err)
	}

	outOfRange := []Instruction{{
		Primitive: ModifyPT{Object: TargetAttachedPermanentReference(5), PowerDelta: Fixed(1)},
	}}
	if err := ValidateInstructionSequence(outOfRange, targets); err == nil {
		t.Fatal("out-of-range attached reference: ValidateInstructionSequence() = nil, want error")
	}

	arbitraryNegative := []Instruction{{
		Primitive: ModifyPT{
			Object:     objectReferenceForTest(ObjectReferenceTargetAttachedPermanent, -5, ""),
			PowerDelta: Fixed(1),
		},
	}}
	if err := ValidateInstructionSequence(arbitraryNegative, targets); err == nil {
		t.Fatal("arbitrary-negative attached reference: ValidateInstructionSequence() = nil, want error")
	}
}

func TestValidateCardDefAttachedPermanentReferenceBounds(t *testing.T) {
	makeCard := func(targetIndex int) *CardDef {
		return &CardDef{CardFace: CardFace{
			Name:       "Attached Modifier",
			OracleText: "The creature an Aura is attached to gets +1/+1.",
			SpellAbility: opt.Val(Mode{
				Targets: []TargetSpec{{MinTargets: 1, MaxTargets: 1}},
				Sequence: []Instruction{{
					Primitive: ModifyPT{
						Object:     TargetAttachedPermanentReference(targetIndex),
						PowerDelta: Fixed(1),
					},
				}},
			}.Ability()),
		}}
	}

	if issues := ValidateCardDef(makeCard(0)); hasCardDefIssue(issues, CardDefIssueInvalidAbilityBody) {
		t.Fatalf("in-range attached reference: issues = %+v, want no ability-body issue", issues)
	}
	if issues := ValidateCardDef(makeCard(5)); !hasCardDefIssue(issues, CardDefIssueInvalidAbilityBody) {
		t.Fatalf("out-of-range attached reference: issues = %+v, want %s", issues, CardDefIssueInvalidAbilityBody)
	}
}

func TestValidateCardDefReportsTypedSearchProblems(t *testing.T) {
	tests := []struct {
		name string
		spec SearchSpec
	}{
		{name: "missing zones"},
		{
			name: "unsupported destination",
			spec: SearchSpec{SourceZone: zone.Library, Destination: zone.Exile},
		},
		{
			name: "empty supertype",
			spec: SearchSpec{
				SourceZone:  zone.Library,
				Destination: zone.Hand,
				Supertype:   opt.Val(types.Super("")),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := &CardDef{CardFace: CardFace{
				Name:       "Bad Search",
				OracleText: "Search your library.",
				SpellAbility: opt.Val(Mode{
					Sequence: []Instruction{{
						Primitive: Search{
							Amount: Fixed(1),
							Player: ControllerReference(),
							Spec:   tt.spec,
						},
					}},
				}.Ability()),
			}}

			issues := ValidateCardDef(card)
			if !hasCardDefIssue(issues, CardDefIssueInvalidAbilityBody) {
				t.Fatalf("issues = %+v, want %s", issues, CardDefIssueInvalidAbilityBody)
			}
		})
	}
}

func TestValidateCardDefChecksDelayedTriggerContent(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name:       "Bad Delayed Trigger",
		OracleText: "At the beginning of the next end step, destroy target creature.",
		SpellAbility: opt.Val(Mode{
			Sequence: []Instruction{{
				Primitive: CreateDelayedTrigger{Trigger: DelayedTriggerDef{
					Timing: DelayedAtBeginningOfNextEndStep,
					Content: Mode{
						Sequence: []Instruction{{
							Primitive: Destroy{Object: TargetPermanentReference(0)},
						}},
					}.Ability(),
				}},
			}},
		}.Ability()),
	}}

	issues := ValidateCardDef(card)
	if !hasCardDefIssue(issues, CardDefIssueInvalidAbilityBody) {
		t.Fatalf("issues = %+v, want %s", issues, CardDefIssueInvalidAbilityBody)
	}
}

func TestValidateCardDefAllowsDelayedTriggerUsingEnclosingLinkedObject(t *testing.T) {
	const key LinkedKey = "delayed-blink"
	card := &CardDef{CardFace: CardFace{
		Name:       "Delayed Blink",
		OracleText: "Exile target creature. Return it to the battlefield at the beginning of the next end step.",
		SpellAbility: opt.Val(Mode{
			Targets: []TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      TargetAllowPermanent,
			}},
			Sequence: []Instruction{
				{Primitive: Exile{
					Object:         TargetPermanentReference(0),
					ExileLinkedKey: key,
				}},
				{Primitive: CreateDelayedTrigger{Trigger: DelayedTriggerDef{
					Timing: DelayedAtBeginningOfNextEndStep,
					Content: Mode{
						Sequence: []Instruction{{Primitive: PutOnBattlefield{
							Source: LinkedBattlefieldSource(key),
						}}},
					}.Ability(),
				}}},
			},
		}.Ability()),
	}}

	if issues := ValidateCardDef(card); len(issues) != 0 {
		t.Fatalf("issues = %+v, want none", issues)
	}
}

func TestValidateCardDefRejectsDelayedTriggerUsingUnpublishedLinkedObject(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name:       "Bad Delayed Blink",
		OracleText: "Return it to the battlefield at the beginning of the next end step.",
		SpellAbility: opt.Val(Mode{
			Sequence: []Instruction{{Primitive: CreateDelayedTrigger{Trigger: DelayedTriggerDef{
				Timing: DelayedAtBeginningOfNextEndStep,
				Content: Mode{
					Sequence: []Instruction{{Primitive: PutOnBattlefield{
						Source: LinkedBattlefieldSource("missing"),
					}}},
				}.Ability(),
			}}}},
		}.Ability()),
	}}

	issues := ValidateCardDef(card)
	if !hasCardDefIssue(issues, CardDefIssueInvalidAbilityBody) {
		t.Fatalf("issues = %+v, want %s", issues, CardDefIssueInvalidAbilityBody)
	}
}

func TestValidateCardDefChecksDelayedTriggerInstructionCondition(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name:       "Bad Delayed Trigger Condition",
		OracleText: "At the beginning of the next end step, draw a card.",
		SpellAbility: opt.Val(Mode{
			Sequence: []Instruction{{
				Primitive: CreateDelayedTrigger{Trigger: DelayedTriggerDef{
					Timing: DelayedAtBeginningOfNextEndStep,
					Content: Mode{
						Sequence: []Instruction{{
							Primitive: Draw{Amount: Fixed(1), Player: ControllerReference()},
							Condition: opt.Val(EffectCondition{Condition: opt.Val(Condition{
								ControllerLifeAtLeast: -1,
							})}),
						}},
					}.Ability(),
				}},
			}},
		}.Ability()),
	}}

	issues := ValidateCardDef(card)
	if !hasCardDefIssue(issues, CardDefIssueInvalidCondition) {
		t.Fatalf("issues = %+v, want %s", issues, CardDefIssueInvalidCondition)
	}
}

func TestValidateCardDefRejectsDelayedTriggerConditionUsingEnclosingTarget(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name:       "Unavailable Delayed Trigger Target",
		OracleText: "Target creature gets +1/+1. At the beginning of the next end step, draw a card.",
		SpellAbility: opt.Val(Mode{
			Targets: []TargetSpec{{MinTargets: 1, MaxTargets: 1}},
			Sequence: []Instruction{{
				Primitive: CreateDelayedTrigger{Trigger: DelayedTriggerDef{
					Timing: DelayedAtBeginningOfNextEndStep,
					Content: Mode{
						Sequence: []Instruction{{
							Primitive: Draw{Amount: Fixed(1), Player: ControllerReference()},
							Condition: opt.Val(EffectCondition{Condition: opt.Val(Condition{
								Object: opt.Val(TargetPermanentReference(0)),
							})}),
						}},
					}.Ability(),
				}},
			}},
		}.Ability()),
	}}

	issues := ValidateCardDef(card)
	if !hasCardDefIssue(issues, CardDefIssueTargetIndexOutOfRange) {
		t.Fatalf("issues = %+v, want %s", issues, CardDefIssueTargetIndexOutOfRange)
	}
}

func TestValidateCardDefChecksNestedEmblemAbility(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name:       "Bad Emblem",
		OracleText: "You get an emblem.",
		SpellAbility: opt.Val(Mode{Sequence: []Instruction{{
			Primitive: CreateEmblem{EmblemAbilities: []Ability{StaticAbility{
				Condition: opt.Val(Condition{ControllerLifeAtLeast: -1}),
			}}},
		}}}.Ability()),
	}}

	issues := ValidateCardDef(card)
	if !hasCardDefIssue(issues, CardDefIssueInvalidCondition) {
		t.Fatalf("issues = %+v, want %s", issues, CardDefIssueInvalidCondition)
	}
}

func TestValidateCardDefChecksNestedReplacementCondition(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name:       "Bad Replacement",
		OracleText: "Create a replacement effect.",
		SpellAbility: opt.Val(Mode{Sequence: []Instruction{{
			Primitive: CreateReplacement{Replacement: &ReplacementEffect{
				MatchEvent: EventPermanentEnteredBattlefield,
				Condition:  opt.Val(Condition{ControllerLifeAtLeast: -1}),
			}},
		}}}.Ability()),
	}}

	issues := ValidateCardDef(card)
	if !hasCardDefIssue(issues, CardDefIssueInvalidCondition) {
		t.Fatalf("issues = %+v, want %s", issues, CardDefIssueInvalidCondition)
	}
}

func TestValidateCardDefReportsInvalidTargetSpec(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name:       "Bad Target Spec",
		OracleText: "Destroy up to negative one target creature.",
		SpellAbility: opt.Val(Mode{
			Targets: []TargetSpec{
				{MinTargets: 2, MaxTargets: 1},
			},
		}.Ability()),
	}}

	issues := ValidateCardDef(card)

	if !hasCardDefIssue(issues, CardDefIssueInvalidTargetSpec) {
		t.Fatalf("issues = %+v, want %s", issues, CardDefIssueInvalidTargetSpec)
	}
}

func TestValidateCardDefStackObjectTargetKinds(t *testing.T) {
	tests := []struct {
		name      string
		spec      TargetSpec
		wantIssue bool
	}{
		{
			name:      "stack target without kinds",
			spec:      TargetSpec{MinTargets: 1, MaxTargets: 1, Allow: TargetAllowStackObject},
			wantIssue: true,
		},
		{
			name: "kinds without stack target",
			spec: TargetSpec{
				MinTargets: 1,
				MaxTargets: 1,
				Predicate:  TargetPredicate{StackObjectKinds: []StackObjectKind{StackActivatedAbility}},
			},
			wantIssue: true,
		},
		{
			name: "duplicate kind",
			spec: TargetSpec{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      TargetAllowStackObject,
				Predicate:  TargetPredicate{StackObjectKinds: []StackObjectKind{StackSpell, StackSpell}},
			},
			wantIssue: true,
		},
		{
			name: "unknown kind",
			spec: TargetSpec{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      TargetAllowStackObject,
				Predicate:  TargetPredicate{StackObjectKinds: []StackObjectKind{StackObjectKind(99)}},
			},
			wantIssue: true,
		},
		{
			name: "spell type without spell kind",
			spec: TargetSpec{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      TargetAllowStackObject,
				Predicate: TargetPredicate{
					SpellCardTypes:   []types.Card{types.Creature},
					StackObjectKinds: []StackObjectKind{StackActivatedAbility},
				},
			},
			wantIssue: true,
		},
		{
			name: "spell type with mixed kinds",
			spec: TargetSpec{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      TargetAllowStackObject,
				Predicate: TargetPredicate{
					SpellCardTypes:   []types.Card{types.Creature},
					StackObjectKinds: []StackObjectKind{StackSpell, StackActivatedAbility},
				},
			},
			wantIssue: true,
		},
		{
			name: "stack target with unsupported predicate",
			spec: TargetSpec{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      TargetAllowStackObject,
				Predicate: TargetPredicate{
					StackObjectKinds: []StackObjectKind{StackActivatedAbility},
					Controller:       ControllerOpponent,
				},
			},
			wantIssue: true,
		},
		{
			name: "mixed stack target with unsupported predicate",
			spec: TargetSpec{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      TargetAllowPermanent | TargetAllowStackObject,
				Predicate: TargetPredicate{
					StackObjectKinds: []StackObjectKind{StackActivatedAbility},
					Controller:       ControllerOpponent,
				},
			},
			wantIssue: true,
		},
		{
			name: "mixed stack target with selection",
			spec: TargetSpec{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      TargetAllowPermanent | TargetAllowStackObject,
				Predicate: TargetPredicate{
					StackObjectKinds: []StackObjectKind{StackActivatedAbility},
				},
				Selection: opt.Val(Selection{
					Controller: ControllerOpponent,
				}),
			},
			wantIssue: true,
		},
		{
			name: "stack target with unknown allow bit",
			spec: TargetSpec{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      TargetAllowStackObject | TargetAllow(1<<30),
				Predicate: TargetPredicate{
					StackObjectKinds: []StackObjectKind{StackActivatedAbility},
					Controller:       ControllerOpponent,
				},
			},
			wantIssue: true,
		},
		{
			name: "valid composite",
			spec: TargetSpec{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      TargetAllowStackObject,
				Predicate: TargetPredicate{
					StackObjectKinds: []StackObjectKind{StackSpell, StackActivatedAbility, StackTriggeredAbility},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			card := &CardDef{CardFace: CardFace{
				Name: "Stack Target",
				SpellAbility: opt.Val(Mode{
					Targets: []TargetSpec{test.spec},
				}.Ability()),
			}}
			got := hasCardDefIssue(ValidateCardDef(card), CardDefIssueInvalidTargetSpec)
			if got != test.wantIssue {
				t.Fatalf("invalid target issue = %v, want %v", got, test.wantIssue)
			}
		})
	}
}

func TestValidateCardDefReportsInvalidTargetChooserSpec(t *testing.T) {
	tests := []struct {
		name string
		spec TargetSpec
	}{
		{
			name: "opponent chooser with optional count",
			spec: TargetSpec{MinTargets: 0, MaxTargets: 1, Chooser: TargetChooserOpponent},
		},
		{
			name: "opponent chooser with opponent-relative controller predicate",
			spec: TargetSpec{
				MinTargets: 1,
				MaxTargets: 1,
				Chooser:    TargetChooserOpponent,
				Predicate:  TargetPredicate{Controller: ControllerOpponent},
			},
		},
		{
			name: "unknown chooser",
			spec: TargetSpec{MinTargets: 1, MaxTargets: 1, Chooser: TargetChooser(99)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := &CardDef{CardFace: CardFace{
				Name:       "Bad Target Chooser",
				OracleText: "Tap target creature.",
				SpellAbility: opt.Val(Mode{
					Targets: []TargetSpec{tt.spec},
				}.Ability()),
			}}

			issues := ValidateCardDef(card)

			if !hasCardDefIssue(issues, CardDefIssueInvalidTargetSpec) {
				t.Fatalf("issues = %+v, want %s", issues, CardDefIssueInvalidTargetSpec)
			}
		})
	}
}

func TestValidateCardDefChecksFaces(t *testing.T) {
	card := &CardDef{
		CardFace: CardFace{Name: "Front"},
		Back: opt.Val(CardFace{
			Name:       "Back Face",
			OracleText: "Draw a card.",
		}),
	}

	issues := ValidateCardDef(card)

	if !hasCardDefIssue(issues, CardDefIssueOracleWithoutAbilities) {
		t.Fatalf("issues = %+v, want face oracle issue", issues)
	}
}

func TestValidateCardDefChecksAlternateFace(t *testing.T) {
	card := &CardDef{
		CardFace: CardFace{Name: "Main Spell"},
		Alternate: opt.Val(CardFace{
			Name:       "Alternate Spell",
			OracleText: "Draw a card.",
		}),
	}

	issues := ValidateCardDef(card)

	if !hasCardDefIssue(issues, CardDefIssueOracleWithoutAbilities) {
		t.Fatalf("issues = %+v, want alternate face oracle issue", issues)
	}
}

func TestValidateCardDefChecksDoubleFacedRootFieldsAndBack(t *testing.T) {
	card := &CardDef{
		CardFace: CardFace{
			Name:       "Double Faced",
			OracleText: "Root text.",
			SpellAbility: opt.Val(Mode{
				Sequence: []Instruction{{Primitive: Destroy{Object: TargetPermanentReference(1)}}},
			}.Ability()),
		},
		Back: opt.Val(CardFace{
			Name:       "Back",
			OracleText: "Draw a card.",
			SpellAbility: opt.Val(Mode{
				Sequence: []Instruction{{Primitive: Draw{Amount: Fixed(1), Player: ControllerReference()}}},
			}.Ability()),
		}),
	}

	issues := ValidateCardDef(card)

	if !hasCardDefIssue(issues, CardDefIssueInvalidAbilityBody) {
		t.Fatalf("issues = %+v, want root ability walk for DFC front face", issues)
	}
}

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

func TestValidateCardDefChecksEnchantTargetSpec(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name:       "Bad Aura",
		OracleText: "Enchant creature",
		StaticAbilities: []StaticAbility{{
			KeywordAbilities: []KeywordAbility{EnchantKeyword{Target: TargetSpec{
				MinTargets: 2,
				MaxTargets: 1,
			}}},
		}},
	}}

	issues := ValidateCardDef(card)

	if !hasCardDefIssue(issues, CardDefIssueInvalidTargetSpec) {
		t.Fatalf("issues = %+v, want enchant target spec issue", issues)
	}
}

func TestValidateCardDefAllowsSelectorOnlyContinuousEffects(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name:       "Static Haste",
		OracleText: "Creatures you control have haste.",
		StaticAbilities: []StaticAbility{{
			ContinuousEffects: []ContinuousEffect{{
				Layer: LayerAbility,
				Group: BattlefieldGroup(Selection{
					RequiredTypes: []types.Card{types.Creature},
					Controller:    ControllerYou,
				}),
				AddKeywords: []Keyword{Haste},
			}},
		}},
	}}

	issues := ValidateCardDef(card)

	if len(issues) != 0 {
		t.Fatalf("issues = %+v, want none", issues)
	}
}

func TestValidateCardDefChecksKeywordAbilities(t *testing.T) {
	tests := []struct {
		name    string
		ability KeywordAbility
		code    CardDefIssueCode
	}{
		{
			name: "nil keyword",
			code: CardDefIssueInvalidKeywordAbility,
		},
		{
			name:    "simple keyword without kind",
			ability: SimpleKeyword{},
			code:    CardDefIssueInvalidKeywordAbility,
		},
		{
			name:    "mana keyword without explicit cost",
			ability: WardKeyword{},
			code:    CardDefIssueInvalidKeywordAbility,
		},
		{
			name: "kicker bonus content",
			ability: KickerKeyword{
				Cost: cost.Mana{cost.G},
				BonusContent: Mode{
					Sequence: []Instruction{{Primitive: Destroy{Object: TargetPermanentReference(0)}}},
				}.Ability(),
			},
			code: CardDefIssueInvalidAbilityBody,
		},
		{
			name: "enchant target spec",
			ability: EnchantKeyword{Target: TargetSpec{
				MinTargets: 2,
				MaxTargets: 1,
			}},
			code: CardDefIssueInvalidTargetSpec,
		},
		{
			name:    "suspend counters",
			ability: SuspendKeyword{Cost: cost.Mana{cost.G}},
			code:    CardDefIssueInvalidKeywordAbility,
		},
		{
			name:    "toxic amount",
			ability: ToxicKeyword{},
			code:    CardDefIssueInvalidKeywordAbility,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := &CardDef{CardFace: CardFace{
				Name:       "Keyword Card",
				OracleText: "Keyword ability.",
				StaticAbilities: []StaticAbility{{
					KeywordAbilities: []KeywordAbility{tt.ability},
				}},
			}}

			issues := ValidateCardDef(card)

			if !hasCardDefIssue(issues, tt.code) {
				t.Fatalf("issues = %+v, want %s", issues, tt.code)
			}
		})
	}
}

func TestValidateCardDefChecksAbilityBodies(t *testing.T) {
	tests := []struct {
		name string
		face CardFace
		code CardDefIssueCode
	}{
		{
			name: "plain content target index",
			face: CardFace{SpellAbility: opt.Val(Mode{
				Sequence: []Instruction{{Primitive: Destroy{Object: TargetPermanentReference(0)}}},
			}.Ability())},
			code: CardDefIssueInvalidAbilityBody,
		},
		{
			name: "nested modal effect",
			face: CardFace{SpellAbility: opt.Val(AbilityContent{
				Modes: []Mode{{
					Sequence: []Instruction{{Primitive: Destroy{Object: TargetPermanentReference(0)}}},
				}},
			})},
			code: CardDefIssueInvalidAbilityBody,
		},
		{
			name: "static keyword",
			face: CardFace{StaticAbilities: []StaticAbility{{
				KeywordAbilities: []KeywordAbility{SimpleKeyword{}},
			}}},
			code: CardDefIssueInvalidKeywordAbility,
		},
		{
			name: "activated keyword",
			face: CardFace{ActivatedAbilities: []ActivatedAbility{{
				KeywordAbilities: []KeywordAbility{SimpleKeyword{}},
				Content:          Mode{}.Ability(),
			}}},
			code: CardDefIssueInvalidKeywordAbility,
		},
		{
			name: "triggered keyword",
			face: CardFace{TriggeredAbilities: []TriggeredAbility{{
				KeywordAbilities: []KeywordAbility{WardKeyword{}},
				Content:          Mode{}.Ability(),
			}}},
			code: CardDefIssueInvalidKeywordAbility,
		},
		{
			name: "trigger intervening condition",
			face: CardFace{TriggeredAbilities: []TriggeredAbility{{
				Trigger: TriggerCondition{
					InterveningCondition: opt.Val(Condition{
						Object: opt.Val(TargetPermanentReference(1)),
					}),
				},
				Content: Mode{}.Ability(),
			}}},
			code: CardDefIssueTargetIndexOutOfRange,
		},
		{
			name: "nil content",
			face: CardFace{SpellAbility: opt.Val(AbilityContent{})},
			code: CardDefIssueInvalidAbilityBody,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			face := tt.face
			face.Name = "Body Card"
			face.OracleText = "Body ability."
			card := &CardDef{CardFace: face}

			issues := ValidateCardDef(card)

			if !hasCardDefIssue(issues, tt.code) {
				t.Fatalf("issues = %+v, want %s", issues, tt.code)
			}
		})
	}
}

func TestValidateCardDefIssueHasFaceName(t *testing.T) {
	card := &CardDef{
		CardFace: CardFace{Name: "Front"},
		Back: opt.Val(CardFace{
			Name:       "BackFace",
			OracleText: "Draw a card.",
		}),
	}

	issues := ValidateCardDef(card)

	for _, issue := range issues {
		if issue.Code == CardDefIssueOracleWithoutAbilities && issue.FaceName == "BackFace" {
			return
		}
	}
	t.Fatalf("issues = %+v, want oracle issue on BackFace", issues)
}

func TestValidateCardDefIssueHasPath(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name:       "Path Card",
		OracleText: "Enchant creature",
		StaticAbilities: []StaticAbility{{
			KeywordAbilities: []KeywordAbility{EnchantKeyword{Target: TargetSpec{
				MinTargets: 2, MaxTargets: 1,
			}}},
		}},
	}}

	issues := ValidateCardDef(card)

	for _, issue := range issues {
		if issue.Code == CardDefIssueInvalidTargetSpec && issue.Path != "" {
			return
		}
	}
	t.Fatalf("issues = %+v, want invalid-target-spec issue with non-empty path", issues)
}

func TestValidateCardDefBackFaceDefaultName(t *testing.T) {
	card := &CardDef{
		CardFace: CardFace{Name: "Front"},
		Back: opt.Val(CardFace{
			OracleText: "Draw a card.",
		}),
	}

	issues := ValidateCardDef(card)

	for _, issue := range issues {
		if issue.Code == CardDefIssueOracleWithoutAbilities && issue.FaceName == "back face" {
			return
		}
	}
	t.Fatalf("issues = %+v, want oracle issue with face name 'back face'", issues)
}

func TestValidateCardDefAlternateFaceDefaultName(t *testing.T) {
	card := &CardDef{
		CardFace:  CardFace{Name: "Front"},
		Alternate: opt.Val(CardFace{OracleText: "Draw a card."}),
	}

	issues := ValidateCardDef(card)

	for _, issue := range issues {
		if issue.Code == CardDefIssueOracleWithoutAbilities && issue.FaceName == "alternate face" {
			return
		}
	}
	t.Fatalf("issues = %+v, want oracle issue with face name 'alternate face'", issues)
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

func TestValidateCardDefReportsTargetSpecDualSpecification(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name:       "Dual Target Spec",
		OracleText: "Destroy target creature.",
		SpellAbility: opt.Val(Mode{
			Targets: []TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      TargetAllowPermanent,
				Predicate:  TargetPredicate{PermanentTypes: []types.Card{types.Creature}},
				Selection:  opt.Val(Selection{RequiredTypesAny: []types.Card{types.Creature}}),
			}},
		}.Ability()),
	}}

	issues := ValidateCardDef(card)

	if !hasCardDefIssue(issues, CardDefIssueInvalidSelection) {
		t.Fatalf("issues = %+v, want %s", issues, CardDefIssueInvalidSelection)
	}
}

func TestValidateCardDefReportsConditionDualSpecification(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name:       "Dual Condition",
		OracleText: "As long as you control a creature, this gets +1/+1.",
		StaticAbilities: []StaticAbility{{
			Condition: opt.Val(Condition{
				ControllerControls: PermanentFilter{Types: []types.Card{types.Creature}},
				ControlsMatching: opt.Val(SelectionCount{
					Selection: Selection{RequiredTypes: []types.Card{types.Creature}},
				}),
			}),
		}},
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
				ControllerControls: PermanentFilter{
					ColorsAny:      []color.Color{color.Red},
					ExcludedColors: []color.Color{color.Red},
				},
			}),
		}},
	}}

	issues := ValidateCardDef(card)

	if !hasCardDefIssue(issues, CardDefIssueInvalidSelection) {
		t.Fatalf("issues = %+v, want %s", issues, CardDefIssueInvalidSelection)
	}
}

func TestValidateCardDefReportsNegativeConditionThresholds(t *testing.T) {
	tests := map[string]Condition{
		"controller life":                     {ControllerLifeAtLeast: -1},
		"any player life":                     {AnyPlayerLifeAtMost: -1},
		"opponent count":                      {OpponentCountAtLeast: -1},
		"controller graveyard cards":          {ControllerGraveyardCardCountAtLeast: -1},
		"controller graveyard card types":     {ControllerGraveyardCardTypeCountAtLeast: -1},
		"controller basic land types":         {ControllerBasicLandTypeCountAtLeast: -1},
		"controller creature power diversity": {ControllerCreaturePowerDiversityAtLeast: -1},
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
				ControllerLifeAtLeast: -1,
			})}),
		}}}.Ability()),
	}}

	issues := ValidateCardDef(card)

	if !hasCardDefIssue(issues, CardDefIssueInvalidCondition) {
		t.Fatalf("issues = %+v, want %s", issues, CardDefIssueInvalidCondition)
	}
}

func TestValidateCardDefReportsTriggerPatternDualSpecification(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name:       "Dual Trigger Pattern",
		OracleText: "Whenever a nontoken creature you control dies, draw a card.",
		TriggeredAbilities: []TriggeredAbility{{
			Content: Mode{}.Ability(),
			Trigger: TriggerCondition{
				Pattern: TriggerPattern{
					Event:                 EventPermanentDied,
					RequirePermanentTypes: []types.Card{types.Creature},
					SubjectSelection:      Selection{RequiredTypes: []types.Card{types.Creature}},
				},
			},
		}},
	}}

	issues := ValidateCardDef(card)

	if !hasCardDefIssue(issues, CardDefIssueInvalidSelection) {
		t.Fatalf("issues = %+v, want %s", issues, CardDefIssueInvalidSelection)
	}
}

func TestValidateCardDefRejectsInvalidIndependentTriggerPatternFields(t *testing.T) {
	invalidSelection := Selection{
		RequiredTypes: []types.Card{types.Creature},
		ExcludedTypes: []types.Card{types.Creature},
	}
	tests := []struct {
		name    string
		pattern TriggerPattern
	}{
		{
			name: "damage source selection",
			pattern: TriggerPattern{
				Event:                 EventDamageDealt,
				DamageSourceSelection: invalidSelection,
			},
		},
		{
			name: "damage recipient is source",
			pattern: TriggerPattern{
				Event:                   EventDamageDealt,
				DamageRecipient:         DamageRecipientPlayer,
				DamageRecipientIsSource: true,
			},
		},
		{
			name: "attack recipient selection",
			pattern: TriggerPattern{
				Event:                    EventAttackerDeclared,
				AttackRecipientSelection: invalidSelection,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := &CardDef{CardFace: CardFace{
				Name: "Invalid Trigger Pattern",
				TriggeredAbilities: []TriggeredAbility{{
					Content: Mode{}.Ability(),
					Trigger: TriggerCondition{Pattern: tt.pattern},
				}},
			}}

			issues := ValidateCardDef(card)

			if len(issues) != 1 || issues[0].Code != CardDefIssueInvalidSelection {
				t.Fatalf("issues = %+v, want one %s", issues, CardDefIssueInvalidSelection)
			}
		})
	}
}

func TestValidateCardDefRequiresNonManaAbilityActivatedTrigger(t *testing.T) {
	t.Parallel()
	card := &CardDef{CardFace: CardFace{
		Name: "Activation Watcher",
		TriggeredAbilities: []TriggeredAbility{{
			Content: Mode{}.Ability(),
			Trigger: TriggerCondition{Pattern: TriggerPattern{
				Event: EventAbilityActivated,
			}},
		}},
	}}

	issues := ValidateCardDef(card)
	if len(issues) != 1 ||
		issues[0].Code != CardDefIssueInvalidSelection ||
		issues[0].Message != "unrestricted ability-activated triggers are unavailable because the runtime event stream omits payment-time mana abilities" {
		t.Fatalf("issues = %+v", issues)
	}

	card.TriggeredAbilities[0].Trigger.Pattern.ExcludeManaAbility = true
	if issues := ValidateCardDef(card); len(issues) != 0 {
		t.Fatalf("non-mana ability-activated trigger issues = %+v", issues)
	}
}

func TestValidateCardDefReportsOneInvalidOneOrMorePerAttackTargetIssue(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name: "Invalid Per-Target Attack Trigger",
		TriggeredAbilities: []TriggeredAbility{{
			Content: Mode{}.Ability(),
			Trigger: TriggerCondition{Pattern: TriggerPattern{
				Event:                    EventAttackerDeclared,
				OneOrMorePerAttackTarget: true,
			}},
		}},
	}}

	issues := ValidateCardDef(card)

	if len(issues) != 1 || issues[0].Code != CardDefIssueInvalidSelection {
		t.Fatalf("issues = %+v, want one %s", issues, CardDefIssueInvalidSelection)
	}
}

func TestValidateCardDefAllowsTokenOnlyTriggerSubject(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name:       "Token Watcher",
		OracleText: "Whenever a token dies, draw a card.",
		TriggeredAbilities: []TriggeredAbility{{
			Content: Mode{}.Ability(),
			Trigger: TriggerCondition{Pattern: TriggerPattern{
				Event:            EventPermanentDied,
				SubjectSelection: Selection{TokenOnly: true},
			}},
		}},
	}}

	issues := ValidateCardDef(card)
	if hasCardDefIssue(issues, CardDefIssueInvalidSelection) {
		t.Fatalf("issues = %+v, want token-only trigger selection accepted", issues)
	}
}

func TestValidateCardDefAllowsColorFilteredSpellCastTrigger(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name:       "Blue Spell Watcher",
		OracleText: "Whenever you cast a blue spell, draw a card.",
		TriggeredAbilities: []TriggeredAbility{{
			Content: Mode{}.Ability(),
			Trigger: TriggerCondition{Pattern: TriggerPattern{
				Event:         EventSpellCast,
				CardSelection: Selection{ColorsAny: []color.Color{color.Blue}},
			}},
		}},
	}}

	issues := ValidateCardDef(card)
	if hasCardDefIssue(issues, CardDefIssueInvalidSelection) {
		t.Fatalf("issues = %+v, want color-filtered spell-cast trigger accepted", issues)
	}
}

func TestValidateCardDefAllowsColorCardinalitySpellCastTrigger(t *testing.T) {
	tests := []struct {
		name      string
		selection Selection
	}{
		{"colorless", Selection{Colorless: true}},
		{"multicolored", Selection{Multicolored: true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := &CardDef{CardFace: CardFace{
				Name:       "Spell Watcher",
				OracleText: "Whenever you cast a spell, draw a card.",
				TriggeredAbilities: []TriggeredAbility{{
					Content: Mode{}.Ability(),
					Trigger: TriggerCondition{Pattern: TriggerPattern{
						Event:         EventSpellCast,
						CardSelection: tt.selection,
					}},
				}},
			}}

			issues := ValidateCardDef(card)
			if hasCardDefIssue(issues, CardDefIssueInvalidSelection) {
				t.Fatalf("issues = %+v, want color-cardinality spell-cast trigger accepted", issues)
			}
		})
	}
}

func TestValidateCardDefAllowsSubtypeSupertypeAndHistoricSpellCastTrigger(t *testing.T) {
	tests := []struct {
		name    string
		pattern TriggerPattern
	}{
		{
			name: "subtypes",
			pattern: TriggerPattern{
				Event:         EventSpellCast,
				CardSelection: Selection{SubtypesAny: []types.Sub{types.Spirit, types.Arcane}},
			},
		},
		{
			name: "supertype",
			pattern: TriggerPattern{
				Event:         EventSpellCast,
				CardSelection: Selection{Supertypes: []types.Super{types.Legendary}},
			},
		},
		{
			name: "historic",
			pattern: TriggerPattern{
				Event:           EventSpellCast,
				RequireHistoric: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := &CardDef{CardFace: CardFace{
				Name:       "Historic Watcher",
				OracleText: "Whenever you cast a historic spell, draw a card.",
				TriggeredAbilities: []TriggeredAbility{{
					Content: Mode{}.Ability(),
					Trigger: TriggerCondition{Pattern: tt.pattern},
				}},
			}}

			issues := ValidateCardDef(card)
			if hasCardDefIssue(issues, CardDefIssueInvalidSelection) {
				t.Fatalf("issues = %+v, want spell-cast trigger predicate accepted", issues)
			}
		})
	}
}

func TestValidateCardDefAllowsManaValueKickedAndZoneSpellCastTrigger(t *testing.T) {
	tests := []struct {
		name    string
		pattern TriggerPattern
	}{
		{
			name: "mana value",
			pattern: TriggerPattern{
				Event: EventSpellCast,
				CardSelection: Selection{
					ManaValue: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 5}),
				},
			},
		},
		{
			name: "kicked",
			pattern: TriggerPattern{
				Event:             EventSpellCast,
				RequireKickerPaid: true,
			},
		},
		{
			name: "graveyard",
			pattern: TriggerPattern{
				Event:         EventSpellCast,
				MatchFromZone: true,
				FromZone:      zone.Graveyard,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := &CardDef{CardFace: CardFace{
				Name:       "Spell Watcher",
				OracleText: "Whenever you cast a spell, draw a card.",
				TriggeredAbilities: []TriggeredAbility{{
					Content: Mode{}.Ability(),
					Trigger: TriggerCondition{Pattern: tt.pattern},
				}},
			}}

			issues := ValidateCardDef(card)
			if hasCardDefIssue(issues, CardDefIssueInvalidSelection) {
				t.Fatalf("issues = %+v, want spell-cast trigger predicate accepted", issues)
			}
		})
	}
}

func TestValidateCardDefRejectsKickerFilterOutsideSpellCastTrigger(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name:       "Invalid Kicker Watcher",
		OracleText: "Whenever a creature enters, draw a card.",
		TriggeredAbilities: []TriggeredAbility{{
			Content: Mode{}.Ability(),
			Trigger: TriggerCondition{Pattern: TriggerPattern{
				Event:             EventPermanentEnteredBattlefield,
				RequireKickerPaid: true,
			}},
		}},
	}}

	issues := ValidateCardDef(card)
	if !hasCardDefIssue(issues, CardDefIssueInvalidSelection) {
		t.Fatalf("issues = %+v, want invalid-selection issue", issues)
	}
}

func TestValidateCardDefRejectsHistoricFilterOutsideSpellCastTrigger(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name:       "Invalid Historic Watcher",
		OracleText: "Whenever a creature enters, draw a card.",
		TriggeredAbilities: []TriggeredAbility{{
			Content: Mode{}.Ability(),
			Trigger: TriggerCondition{Pattern: TriggerPattern{
				Event:           EventPermanentEnteredBattlefield,
				RequireHistoric: true,
			}},
		}},
	}}

	issues := ValidateCardDef(card)
	if !hasCardDefIssue(issues, CardDefIssueInvalidSelection) {
		t.Fatalf("issues = %+v, want invalid-selection issue", issues)
	}
}

func TestValidateCardDefRejectsContradictoryColorCardinalitySelection(t *testing.T) {
	tests := []struct {
		name      string
		selection Selection
	}{
		{"colorless multicolored", Selection{Colorless: true, Multicolored: true}},
		{"colorless colored", Selection{Colorless: true, ColorsAny: []color.Color{color.Blue}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := &CardDef{CardFace: CardFace{
				Name:       "Impossible Watcher",
				OracleText: "Whenever you cast a spell, draw a card.",
				TriggeredAbilities: []TriggeredAbility{{
					Content: Mode{}.Ability(),
					Trigger: TriggerCondition{Pattern: TriggerPattern{
						Event:         EventSpellCast,
						CardSelection: tt.selection,
					}},
				}},
			}}

			issues := ValidateCardDef(card)
			if !hasCardDefIssue(issues, CardDefIssueInvalidSelection) {
				t.Fatalf("issues = %+v, want invalid-selection issue", issues)
			}
		})
	}
}

func TestValidateCardDefAllowsSelectionOnlyTargetSpec(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name:       "Selection Target Spec",
		OracleText: "Destroy target creature.",
		SpellAbility: opt.Val(Mode{
			Targets: []TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      TargetAllowPermanent,
				Selection:  opt.Val(Selection{RequiredTypesAny: []types.Card{types.Creature}}),
			}},
		}.Ability()),
	}}

	issues := ValidateCardDef(card)

	if hasCardDefIssue(issues, CardDefIssueInvalidSelection) {
		t.Fatalf("issues = %+v, want no invalid-selection issue", issues)
	}
}

func TestValidateCardDefRejectsSelectionTargetWithoutAllow(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name:       "Selection Without Valence",
		OracleText: "Destroy target creature.",
		SpellAbility: opt.Val(Mode{
			Targets: []TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Selection:  opt.Val(Selection{RequiredTypes: []types.Card{types.Creature}}),
			}},
		}.Ability()),
	}}

	issues := ValidateCardDef(card)

	if !hasCardDefIssue(issues, CardDefIssueInvalidTargetSpec) {
		t.Fatalf("issues = %+v, want %s", issues, CardDefIssueInvalidTargetSpec)
	}
}

func TestValidateCardDefRejectsSelectionFieldsUnavailableInContext(t *testing.T) {
	tests := []struct {
		name string
		face CardFace
	}{
		{
			name: "player target with permanent predicate",
			face: CardFace{SpellAbility: opt.Val(Mode{
				Targets: []TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Allow:      TargetAllowPlayer,
					Selection:  opt.Val(Selection{RequiredTypes: []types.Card{types.Creature}}),
				}},
			}.Ability())},
		},
		{
			name: "player target with token predicate",
			face: CardFace{SpellAbility: opt.Val(Mode{
				Targets: []TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Allow:      TargetAllowPlayer,
					Selection:  opt.Val(Selection{TokenOnly: true}),
				}},
			}.Ability())},
		},
		{
			name: "controlled permanents with player relation",
			face: CardFace{StaticAbilities: []StaticAbility{{
				Condition: opt.Val(Condition{ControlsMatching: opt.Val(SelectionCount{
					Selection: Selection{Player: PlayerOpponent},
				})}),
			}}},
		},
		{
			name: "trigger card with power",
			face: CardFace{TriggeredAbilities: []TriggeredAbility{{
				Content: Mode{}.Ability(),
				Trigger: TriggerCondition{Pattern: TriggerPattern{
					Event:         EventSpellCast,
					CardSelection: Selection{Power: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 2})},
				}},
			}}},
		},
		{
			name: "non-cast trigger card with color",
			face: CardFace{TriggeredAbilities: []TriggeredAbility{{
				Content: Mode{}.Ability(),
				Trigger: TriggerCondition{Pattern: TriggerPattern{
					Event:         EventCardDrawn,
					CardSelection: Selection{ColorsAny: []color.Color{color.Blue}},
				}},
			}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			face := tt.face
			face.Name = "Invalid Selection Context"
			face.OracleText = "Ability."

			issues := ValidateCardDef(&CardDef{CardFace: face})

			if !hasCardDefIssue(issues, CardDefIssueInvalidSelection) {
				t.Fatalf("issues = %+v, want %s", issues, CardDefIssueInvalidSelection)
			}
		})
	}
}

func TestValidateCardDefAllowsTriggerSubjectLastKnownSelectionFields(t *testing.T) {
	def := &CardDef{CardFace: CardFace{
		Name:       "LKI Watcher",
		OracleText: "Whenever a legendary green Dragon with flying dies, draw a card.",
		TriggeredAbilities: []TriggeredAbility{{
			Trigger: TriggerCondition{Pattern: TriggerPattern{
				Event: EventPermanentDied,
				SubjectSelection: Selection{
					Supertypes:  []types.Super{types.Legendary},
					SubtypesAny: []types.Sub{types.Dragon},
					ColorsAny:   []color.Color{color.Green},
					Tapped:      TriTrue,
					Keyword:     Flying,
					ManaValue:   opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4}),
					Power:       opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4}),
					Toughness:   opt.Val(compare.Int{Op: compare.Equal, Value: 4}),
				},
			}},
			Content: Mode{}.Ability(),
		}},
	}}

	if issues := ValidateCardDef(def); len(issues) != 0 {
		t.Fatalf("issues = %+v, want none", issues)
	}
}

func TestValidateCardDefAllowsHandCyclingGrantRuleEffect(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name: "Cycling Granter",
		StaticAbilities: []StaticAbility{{
			RuleEffects: []RuleEffect{{
				Kind:           RuleEffectGrantHandCardAbility,
				AffectedPlayer: PlayerYou,
				CardSelection: Selection{
					RequiredTypes: []types.Card{types.Land},
				},
				GrantedAbility: CyclingActivatedAbility(cost.Mana{cost.R}),
			}},
		}},
	}}
	if issues := ValidateCardDef(card); len(issues) != 0 {
		t.Fatalf("issues = %+v, want none", issues)
	}
}

func TestValidateCardDefRejectsInvalidHandCyclingGrantRuleEffect(t *testing.T) {
	tests := []struct {
		name   string
		effect RuleEffect
		code   CardDefIssueCode
	}{
		{
			name: "missing affected player",
			effect: RuleEffect{
				Kind: RuleEffectGrantHandCardAbility,
				CardSelection: Selection{
					RequiredTypes: []types.Card{types.Land},
				},
				GrantedAbility: CyclingActivatedAbility(cost.Mana{cost.R}),
			},
			code: CardDefIssueInvalidRuleEffect,
		},
		{
			name: "empty card selection",
			effect: RuleEffect{
				Kind:           RuleEffectGrantHandCardAbility,
				AffectedPlayer: PlayerYou,
				GrantedAbility: CyclingActivatedAbility(cost.Mana{cost.R}),
			},
			code: CardDefIssueInvalidSelection,
		},
		{
			name: "unsupported card predicate",
			effect: RuleEffect{
				Kind:           RuleEffectGrantHandCardAbility,
				AffectedPlayer: PlayerYou,
				CardSelection: Selection{
					RequiredTypes: []types.Card{types.Land},
					Tapped:        TriTrue,
				},
				GrantedAbility: CyclingActivatedAbility(cost.Mana{cost.R}),
			},
			code: CardDefIssueInvalidSelection,
		},
		{
			name: "non-cycling grant",
			effect: RuleEffect{
				Kind:           RuleEffectGrantHandCardAbility,
				AffectedPlayer: PlayerYou,
				CardSelection: Selection{
					RequiredTypes: []types.Card{types.Land},
				},
				GrantedAbility: ActivatedAbility{},
			},
			code: CardDefIssueInvalidRuleEffect,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := &CardDef{CardFace: CardFace{
				Name: "Bad Cycling Granter",
				StaticAbilities: []StaticAbility{{
					RuleEffects: []RuleEffect{tt.effect},
				}},
			}}
			issues := ValidateCardDef(card)
			if !hasCardDefIssue(issues, tt.code) {
				t.Fatalf("issues = %+v, want %s", issues, tt.code)
			}
		})
	}
}

func hasCardDefIssue(issues []CardDefIssue, code CardDefIssueCode) bool {
	for _, issue := range issues {
		if issue.Code == code {
			return true
		}
	}
	return false
}

func TestValidateCardDefGroupReferenceValidation(t *testing.T) {
	// makeStaticWithContinuous wraps a ContinuousEffect in a StaticAbility with nil targets.
	makeStaticWithContinuous := func(group GroupReference) *CardDef {
		return &CardDef{CardFace: CardFace{
			Name:       "Static Haste Card",
			OracleText: "Creatures you control have haste.",
			StaticAbilities: []StaticAbility{{
				ContinuousEffects: []ContinuousEffect{{
					Layer:       LayerAbility,
					Group:       group,
					AddKeywords: []Keyword{Haste},
				}},
			}},
		}}
	}

	t.Run("valid source-reference BattlefieldGroup", func(t *testing.T) {
		def := makeStaticWithContinuous(BattlefieldGroup(Selection{
			RequiredTypes: []types.Card{types.Creature},
			Controller:    ControllerYou,
		}))
		issues := ValidateCardDef(def)
		if len(issues) != 0 {
			t.Fatalf("expected no issues, got %+v", issues)
		}
	})

	t.Run("group with contradictory token selection", func(t *testing.T) {
		def := makeStaticWithContinuous(BattlefieldGroup(Selection{
			NonToken:  true,
			TokenOnly: true,
		}))
		issues := ValidateCardDef(def)
		if !hasCardDefIssue(issues, CardDefIssueInvalidReference) {
			t.Fatalf("expected %s for contradictory token selection, got %+v", CardDefIssueInvalidReference, issues)
		}
	})

	t.Run("BattlefieldGroupExcluding with out-of-range exclusion target", func(t *testing.T) {
		// TargetPermanentReference(5) with nil targets means index 5 >= len(nil)=0.
		def := makeStaticWithContinuous(BattlefieldGroupExcluding(
			Selection{RequiredTypes: []types.Card{types.Creature}},
			TargetPermanentReference(5),
		))
		issues := ValidateCardDef(def)
		if !hasCardDefIssue(issues, CardDefIssueTargetIndexOutOfRange) {
			t.Fatalf("expected %s for out-of-range exclusion target, got %+v", CardDefIssueTargetIndexOutOfRange, issues)
		}
	})

	t.Run("ObjectControlledGroup with out-of-range anchor target", func(t *testing.T) {
		def := makeStaticWithContinuous(ObjectControlledGroup(
			TargetPermanentReference(3),
			Selection{RequiredTypes: []types.Card{types.Creature}},
		))
		issues := ValidateCardDef(def)
		if !hasCardDefIssue(issues, CardDefIssueTargetIndexOutOfRange) {
			t.Fatalf("expected %s for out-of-range anchor target, got %+v", CardDefIssueTargetIndexOutOfRange, issues)
		}
	})

	t.Run("valid source-anchor ObjectControlledGroup", func(t *testing.T) {
		def := makeStaticWithContinuous(ObjectControlledGroup(
			SourcePermanentReference(),
			Selection{RequiredTypes: []types.Card{types.Creature}},
		))
		issues := ValidateCardDef(def)
		if len(issues) != 0 {
			t.Fatalf("expected no issues for source-anchored group, got %+v", issues)
		}
	})
}

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
