package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestValidateCardDefAllowsCrossAbilityLinkedExileReturn(t *testing.T) {
	const key LinkedKey = "exile-until-leaves"
	card := &CardDef{CardFace: CardFace{
		Name:       "Banisher",
		OracleText: "When this creature enters, exile target creature until this creature leaves the battlefield.",
		TriggeredAbilities: []TriggeredAbility{
			{
				Trigger: TriggerCondition{Pattern: TriggerPattern{
					Event:  EventPermanentEnteredBattlefield,
					Source: TriggerSourceSelf,
				}},
				Content: Mode{
					Targets: []TargetSpec{{MinTargets: 1, MaxTargets: 1, Allow: TargetAllowPermanent}},
					Sequence: []Instruction{{Primitive: Exile{
						Object:         TargetPermanentReference(0),
						ExileLinkedKey: key,
					}}},
				}.Ability(),
			},
			{
				Trigger: TriggerCondition{Pattern: TriggerPattern{
					Event:         EventZoneChanged,
					Source:        TriggerSourceSelf,
					MatchFromZone: true,
					FromZone:      zone.Battlefield,
				}},
				Content: Mode{
					Sequence: []Instruction{{Primitive: PutOnBattlefield{
						Source: LinkedBattlefieldSource(key),
					}}},
				}.Ability(),
			},
		},
	}}

	if issues := ValidateCardDef(card); len(issues) != 0 {
		t.Fatalf("issues = %+v, want none", issues)
	}
}

func TestValidateCardDefRejectsCrossAbilityLinkedReturnWithoutPublisher(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name:       "Bad Banisher",
		OracleText: "When this creature leaves the battlefield, return the exiled card.",
		TriggeredAbilities: []TriggeredAbility{{
			Trigger: TriggerCondition{Pattern: TriggerPattern{
				Event:         EventZoneChanged,
				Source:        TriggerSourceSelf,
				MatchFromZone: true,
				FromZone:      zone.Battlefield,
			}},
			Content: Mode{
				Sequence: []Instruction{{Primitive: PutOnBattlefield{
					Source: LinkedBattlefieldSource("missing"),
				}}},
			}.Ability(),
		}},
	}}

	issues := ValidateCardDef(card)
	if !hasCardDefIssue(issues, CardDefIssueInvalidAbilityBody) {
		t.Fatalf("issues = %+v, want %s", issues, CardDefIssueInvalidAbilityBody)
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
								Aggregates: []AggregateComparison{{Aggregate: AggregateControllerLife, Op: compare.GreaterOrEqual, Value: -1}},
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

func TestValidateCardDefAllowsSubjectSelectionOrSelfTrigger(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name:       "Ally Watcher",
		OracleText: "Whenever this creature or another Ally you control enters, draw a card.",
		TriggeredAbilities: []TriggeredAbility{{
			Content: Mode{}.Ability(),
			Trigger: TriggerCondition{Pattern: TriggerPattern{
				Event:                  EventPermanentEnteredBattlefield,
				Controller:             TriggerControllerYou,
				SubjectSelectionOrSelf: true,
				SubjectSelection:       Selection{SubtypesAny: []types.Sub{types.Sub("Ally")}},
			}},
		}},
	}}

	issues := ValidateCardDef(card)
	if hasCardDefIssue(issues, CardDefIssueInvalidSelection) {
		t.Fatalf("issues = %+v, want self-or-another trigger accepted", issues)
	}
}

func TestValidateCardDefRejectsInvalidSubjectSelectionOrSelfTrigger(t *testing.T) {
	tests := []struct {
		name    string
		pattern TriggerPattern
	}{
		{
			name: "missing subject selection",
			pattern: TriggerPattern{
				Event:                  EventPermanentEnteredBattlefield,
				SubjectSelectionOrSelf: true,
			},
		},
		{
			name: "combined with source filter",
			pattern: TriggerPattern{
				Event:                  EventPermanentEnteredBattlefield,
				Source:                 TriggerSourceSelf,
				SubjectSelectionOrSelf: true,
				SubjectSelection:       Selection{RequiredTypes: []types.Card{types.Creature}},
			},
		},
		{
			name: "combined with exclude self",
			pattern: TriggerPattern{
				Event:                  EventPermanentEnteredBattlefield,
				ExcludeSelf:            true,
				SubjectSelectionOrSelf: true,
				SubjectSelection:       Selection{RequiredTypes: []types.Card{types.Creature}},
			},
		},
		{
			name: "unsupported event",
			pattern: TriggerPattern{
				Event:                  EventSpellCast,
				SubjectSelectionOrSelf: true,
				SubjectSelection:       Selection{RequiredTypes: []types.Card{types.Creature}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := &CardDef{CardFace: CardFace{
				Name: "Invalid Self Or Another",
				TriggeredAbilities: []TriggeredAbility{{
					Content: Mode{}.Ability(),
					Trigger: TriggerCondition{Pattern: tt.pattern},
				}},
			}}
			issues := ValidateCardDef(card)
			if !hasCardDefIssue(issues, CardDefIssueInvalidSelection) {
				t.Fatalf("issues = %+v, want %s", issues, CardDefIssueInvalidSelection)
			}
		})
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
