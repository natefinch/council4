package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
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

func TestValidateCardDefReportsInvalidModalChoiceRange(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name: "Invalid Modal",
		TriggeredAbilities: []TriggeredAbility{{
			Content: AbilityContent{
				MinModes: 0,
				MaxModes: 3,
				Modes:    []Mode{{}, {}},
			},
		}},
	}}

	issues := ValidateCardDef(card)
	if !hasCardDefIssue(issues, CardDefIssueInvalidAbilityBody) {
		t.Fatalf("issues = %+v, want %s for invalid modal choice range", issues, CardDefIssueInvalidAbilityBody)
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

func TestValidateCardDefValidatesSourceAbilityCostModifiers(t *testing.T) {
	t.Parallel()

	valid := CostModifier{
		Kind:               CostModifierAbility,
		PerObjectReduction: 1,
		CountSelection: Selection{
			RequiredTypes: []types.Card{types.Creature},
			Supertypes:    []types.Super{types.Legendary},
			Controller:    ControllerYou,
		},
	}
	card := &CardDef{CardFace: CardFace{
		Name: "Source Ability Modifier",
		ActivatedAbilities: []ActivatedAbility{{
			CostModifiers: []CostModifier{valid},
			Content:       Mode{Sequence: []Instruction{{Primitive: Draw{Amount: Fixed(1), Player: ControllerReference()}}}}.Ability(),
		}},
	}}
	if issues := ValidateCardDef(card); len(issues) != 0 {
		t.Fatalf("valid source ability modifier issues = %+v, want none", issues)
	}

	card.ActivatedAbilities[0].CostModifiers[0].CountSelection = Selection{}
	issues := ValidateCardDef(card)
	if !hasCardDefIssue(issues, CardDefIssueInvalidRuleEffect) {
		t.Fatalf("missing count selection issues = %+v, want %s", issues, CardDefIssueInvalidRuleEffect)
	}
}

func TestValidateCardDefValidatesDynamicSpellCostReduction(t *testing.T) {
	t.Parallel()

	makeCard := func(modifier CostModifier) *CardDef {
		return &CardDef{CardFace: CardFace{
			Name:  "Dynamic Reducer",
			Types: []types.Card{types.Sorcery},
			StaticAbilities: []StaticAbility{{
				RuleEffects: []RuleEffect{{
					Kind:           RuleEffectCostModifier,
					AffectedSource: true,
					CostModifier:   modifier,
				}},
			}},
		}}
	}

	valid := CostModifier{
		Kind: CostModifierSpell,
		DynamicReduction: &DynamicAmount{
			Kind:  DynamicAmountGreatestPowerInGroup,
			Group: BattlefieldGroup(Selection{RequiredTypes: []types.Card{types.Creature}, Controller: ControllerYou}),
		},
	}
	if issues := ValidateCardDef(makeCard(valid)); len(issues) != 0 {
		t.Fatalf("valid dynamic spell cost reduction issues = %+v, want none", issues)
	}

	withPerObject := valid
	withPerObject.PerObjectReduction = 1
	withPerObject.CountSelection = Selection{RequiredTypes: []types.Card{types.Creature}}
	if issues := ValidateCardDef(makeCard(withPerObject)); !hasCardDefIssue(issues, CardDefIssueInvalidRuleEffect) {
		t.Fatalf("dynamic+per-object reduction issues = %+v, want %s", issues, CardDefIssueInvalidRuleEffect)
	}

	unsupportedKind := CostModifier{
		Kind: CostModifierSpell,
		DynamicReduction: &DynamicAmount{
			Kind: DynamicAmountTargetPower,
		},
	}
	if issues := ValidateCardDef(makeCard(unsupportedKind)); !hasCardDefIssue(issues, CardDefIssueInvalidRuleEffect) {
		t.Fatalf("unsupported dynamic reduction kind issues = %+v, want %s", issues, CardDefIssueInvalidRuleEffect)
	}
}

func TestValidateCardDefChosenSubtypeCostModifierRequiresEntryChoice(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name: "Chosen Type Reducer",
		StaticAbilities: []StaticAbility{{
			RuleEffects: []RuleEffect{{
				Kind: RuleEffectCostModifier,
				CostModifier: CostModifier{
					Kind:                         CostModifierSpell,
					MatchCardType:                true,
					CardType:                     types.Creature,
					ChosenSubtypeFromEntryChoice: true,
					GenericReduction:             1,
				},
			}},
		}},
	}}

	issues := ValidateCardDef(card)

	if !hasCardDefIssue(issues, CardDefIssueInvalidAbilityBody) {
		t.Fatalf("issues = %+v, want %s", issues, CardDefIssueInvalidAbilityBody)
	}
}

func TestValidateCardDefChosenSubtypeCostModifierRequiresCreatureSpells(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name: "Chosen Type Reducer",
		ReplacementAbilities: []ReplacementAbility{{
			Replacement: ReplacementEffect{EntryTypeChoice: true},
		}},
		StaticAbilities: []StaticAbility{{
			RuleEffects: []RuleEffect{{
				Kind: RuleEffectCostModifier,
				CostModifier: CostModifier{
					Kind:                         CostModifierSpell,
					MatchCardType:                true,
					CardType:                     types.Artifact,
					ChosenSubtypeFromEntryChoice: true,
					GenericReduction:             1,
				},
			}},
		}},
	}}

	issues := ValidateCardDef(card)

	if !hasCardDefIssue(issues, CardDefIssueInvalidRuleEffect) {
		t.Fatalf("issues = %+v, want %s", issues, CardDefIssueInvalidRuleEffect)
	}
}

func TestValidateCardDefAllowsVanillaPermanentWithAdditionalCost(t *testing.T) {
	// A permanent spell whose only Oracle text is an additional cost to cast
	// (e.g. Makeshift Mauler) has no abilities but is still a complete card.
	card := &CardDef{CardFace: CardFace{
		Name:       "Makeshift Mauler",
		OracleText: "As an additional cost to cast this spell, exile a creature card from your graveyard.",
		Types:      []types.Card{types.Creature},
		AdditionalCosts: []cost.Additional{{
			Kind:          cost.AdditionalExile,
			Amount:        1,
			MatchCardType: true,
			CardType:      types.Creature,
			Source:        zone.Graveyard,
		}},
	}}

	issues := ValidateCardDef(card)

	if len(issues) != 0 {
		t.Fatalf("issues = %+v, want none", issues)
	}
}

func TestValidateCardDefAllowsAlternativeCostOnlyOracleText(t *testing.T) {
	t.Parallel()
	card := &CardDef{CardFace: CardFace{
		Name:             "Conditional Free Spell",
		OracleText:       "If you control a commander, you may cast this spell without paying its mana cost.",
		AlternativeCosts: []cost.Alternative{{Condition: cost.AlternativeConditionControlsCommander}},
	}}
	if issues := ValidateCardDef(card); len(issues) != 0 {
		t.Fatalf("issues = %+v, want none", issues)
	}
}

func TestValidateCardDefValidatesEnterBattlefieldResolutionPaymentContext(t *testing.T) {
	t.Parallel()
	sourceCounters := DynamicAmount{
		Kind:        DynamicAmountObjectCounters,
		Object:      SourcePermanentReference(),
		CounterKind: counter.Age,
	}
	card := &CardDef{CardFace: CardFace{
		Name: "Dynamic ETB Payment",
		ReplacementAbilities: []ReplacementAbility{
			EntersTappedUnlessPaidReplacement("Dynamic ETB Payment enters tapped unless its cost is paid.", ResolutionPayment{
				ManaCost:           opt.Val(cost.Mana{cost.O(1)}),
				ManaCostMultiplier: opt.Val(&sourceCounters),
			}),
		},
	}}
	if issues := ValidateCardDef(card); len(issues) != 0 {
		t.Fatalf("source-counter payment issues = %+v, want none", issues)
	}

	unsafeSource := sourceCounters
	unsafeSource.Object = SourceCardPermanentReference()
	card.ReplacementAbilities[0].UnlessPaid.Val.ManaCostMultiplier = opt.Val(&unsafeSource)
	if issues := ValidateCardDef(card); !hasCardDefIssue(issues, CardDefIssueInvalidAbilityBody) {
		t.Fatalf("unsafe source payment issues = %+v, want %s", issues, CardDefIssueInvalidAbilityBody)
	}

	eventAmount := DynamicAmount{Kind: DynamicAmountEventDamage}
	card.ReplacementAbilities[0].UnlessPaid.Val.ManaCostMultiplier = opt.Val(&eventAmount)
	if issues := ValidateCardDef(card); !hasCardDefIssue(issues, CardDefIssueInvalidAbilityBody) {
		t.Fatalf("event payment issues = %+v, want %s", issues, CardDefIssueInvalidAbilityBody)
	}

	sourcePower := DynamicAmount{
		Kind:   DynamicAmountObjectPower,
		Object: SourcePermanentReference(),
	}
	card.ReplacementAbilities[0].UnlessPaid.Val.ManaCostMultiplier = opt.Val(&sourcePower)
	if issues := ValidateCardDef(card); !hasCardDefIssue(issues, CardDefIssueInvalidAbilityBody) {
		t.Fatalf("source-power payment issues = %+v, want %s", issues, CardDefIssueInvalidAbilityBody)
	}
}

func TestValidateCardDefRejectsUnknownAlternativeCostCondition(t *testing.T) {
	t.Parallel()
	card := &CardDef{CardFace: CardFace{
		Name:             "Invalid Alternate",
		AlternativeCosts: []cost.Alternative{{Condition: cost.AlternativeCondition(99)}},
	}}
	issues := ValidateCardDef(card)
	if !hasCardDefIssue(issues, CardDefIssueInvalidAlternativeCost) {
		t.Fatalf("issues = %+v, want %s", issues, CardDefIssueInvalidAlternativeCost)
	}
}

func TestValidateCardDefRejectsTargetedOverloadAbility(t *testing.T) {
	t.Parallel()
	normal := Mode{Sequence: []Instruction{{Primitive: Draw{
		Amount: Fixed(1),
		Player: ControllerReference(),
	}}}}.Ability()
	targeted := Mode{
		Targets: []TargetSpec{{MinTargets: 1, MaxTargets: 1}},
		Sequence: []Instruction{{Primitive: Destroy{
			Object: TargetPermanentReference(0),
		}}},
	}.Ability()
	card := &CardDef{CardFace: CardFace{
		Name:         "Invalid Overload",
		SpellAbility: opt.Val(normal),
		Overload: opt.Val(OverloadAbility{
			Cost:         cost.Mana{cost.U},
			SpellAbility: targeted,
		}),
	}}
	issues := ValidateCardDef(card)
	if !hasCardDefIssue(issues, CardDefIssueInvalidAlternativeCost) {
		t.Fatalf("issues = %+v, want %s", issues, CardDefIssueInvalidAlternativeCost)
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
		{
			name: "library without top position",
			spec: SearchSpec{SourceZone: zone.Library, Destination: zone.Library},
		},
		{
			name: "multiple cards to library top",
			spec: SearchSpec{
				SourceZone:          zone.Library,
				Destination:         zone.Library,
				DestinationPosition: SearchPositionTop,
			},
		},
		{
			name: "single-item card type union",
			spec: SearchSpec{
				SourceZone:   zone.Library,
				Destination:  zone.Hand,
				CardTypesAny: []types.Card{types.Artifact},
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
							Amount: func() Quantity {
								if tt.name == "multiple cards to library top" {
									return Fixed(2)
								}
								return Fixed(1)
							}(),
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

func TestValidateCardDefRejectsInvalidRequiredSearchPolicies(t *testing.T) {
	tests := []struct {
		name   string
		amount Quantity
		spec   SearchSpec
	}{
		{
			name:   "unknown policy",
			amount: Fixed(1),
			spec: SearchSpec{
				SourceZone:       zone.Library,
				Destination:      zone.Hand,
				FailToFindPolicy: SearchFailToFindPolicy(255),
			},
		},
		{
			name:   "qualified search",
			amount: Fixed(1),
			spec: SearchSpec{
				SourceZone:       zone.Library,
				Destination:      zone.Hand,
				FailToFindPolicy: SearchMustFindIfAvailable,
				CardType:         opt.Val(types.Creature),
			},
		},
		{
			name:   "explicit fail on singular unrestricted search",
			amount: Fixed(1),
			spec: SearchSpec{
				SourceZone:       zone.Library,
				Destination:      zone.Hand,
				FailToFindPolicy: SearchMayFailToFind,
			},
		},
		{
			name:   "multiple cards",
			amount: Fixed(2),
			spec: SearchSpec{
				SourceZone:       zone.Library,
				Destination:      zone.Hand,
				FailToFindPolicy: SearchMustFindIfAvailable,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			card := &CardDef{CardFace: CardFace{
				Name:       "Bad Required Search",
				OracleText: "Search your library.",
				SpellAbility: opt.Val(Mode{Sequence: []Instruction{{
					Primitive: Search{
						Amount: test.amount,
						Player: ControllerReference(),
						Spec:   test.spec,
					},
				}}}.Ability()),
			}}

			issues := ValidateCardDef(card)
			if !hasCardDefIssue(issues, CardDefIssueInvalidAbilityBody) {
				t.Fatalf("issues = %+v, want %s", issues, CardDefIssueInvalidAbilityBody)
			}
		})
	}
}

func TestValidateCardDefChecksNestedEmblemAbility(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name:       "Bad Emblem",
		OracleText: "You get an emblem.",
		SpellAbility: opt.Val(Mode{Sequence: []Instruction{{
			Primitive: CreateEmblem{EmblemAbilities: []Ability{&StaticAbility{
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

func TestValidateCardDefRejectsUnknownEntryChoiceSubtypeReference(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name:  "Invalid Choice Reader",
		Types: []types.Card{types.Creature},
		StaticAbilities: []StaticAbility{{
			ContinuousEffects: []ContinuousEffect{{
				Layer:                     LayerType,
				AffectedSource:            true,
				AddSubtypeFromEntryChoice: ChoiceKey("unknown-choice"),
			}},
		}},
	}}

	issues := ValidateCardDef(card)
	if !hasCardDefIssue(issues, CardDefIssueInvalidReference) {
		t.Fatalf("issues = %+v, want %s", issues, CardDefIssueInvalidReference)
	}
}

func TestValidateCardDefGrantedManaAbility(t *testing.T) {
	cardWithEffect := func(layer ContinuousLayer, ability ManaAbility) *CardDef {
		return &CardDef{CardFace: CardFace{
			Name:       "Mana Grant",
			OracleText: "Lands you control have a mana ability.",
			StaticAbilities: []StaticAbility{{
				ContinuousEffects: []ContinuousEffect{{
					Layer: layer,
					Group: ObjectControlledGroup(
						SourcePermanentReference(),
						Selection{RequiredTypes: []types.Card{types.Land}},
					),
					AddAbilities: []Ability{&ability},
				}},
			}},
		}}
	}

	canonical := TapAnyColorManaAbility()
	if issues := ValidateCardDef(cardWithEffect(LayerAbility, canonical)); len(issues) != 0 {
		t.Fatalf("canonical granted mana ability issues = %+v, want none", issues)
	}
	tests := []struct {
		name   string
		mutate func(*ManaAbility)
	}{
		{
			name: "fixed color",
			mutate: func(ability *ManaAbility) {
				*ability = TapManaAbility(mana.G)
			},
		},
		{
			name: "wrong cost",
			mutate: func(ability *ManaAbility) {
				ability.AdditionalCosts = nil
				ability.ManaCost = opt.Val(cost.Mana{cost.O(1)})
			},
		},
		{
			name: "mutated tap cost",
			mutate: func(ability *ManaAbility) {
				ability.AdditionalCosts[0].Kind = cost.AdditionalUntap
			},
		},
		{
			name: "nonbattlefield zone",
			mutate: func(ability *ManaAbility) {
				ability.ZoneOfFunction = zone.Hand
			},
		},
		{
			name: "activation condition",
			mutate: func(ability *ManaAbility) {
				ability.ActivationCondition = opt.Val(Condition{ControllerLifeAtLeast: 1})
			},
		},
		{
			name: "commander identity color source with WUBRG colors",
			mutate: func(ability *ManaAbility) {
				choose, ok := ability.Content.Modes[0].Sequence[0].Primitive.(Choose)
				if !ok {
					panic("TapAnyColorManaAbility choice instruction is not Choose")
				}
				choose.Choice.ColorSource = ResolutionChoiceColorSourceCommanderIdentity
				ability.Content.Modes[0].Sequence[0].Primitive = choose
			},
		},
		{
			name: "condition-gated choice",
			mutate: func(ability *ManaAbility) {
				ability.Content.Modes[0].Sequence[0].Condition = opt.Val(EffectCondition{Text: "condition"})
			},
		},
		{
			name: "result-gated AddMana",
			mutate: func(ability *ManaAbility) {
				ability.Content.Modes[0].Sequence[1].ResultGate = opt.Val(InstructionResultGate{
					Key:       "prior",
					Succeeded: TriTrue,
				})
			},
		},
		{
			name: "optional AddMana",
			mutate: func(ability *ManaAbility) {
				ability.Content.Modes[0].Sequence[1].Optional = true
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ability := TapAnyColorManaAbility()
			test.mutate(&ability)
			issues := ValidateCardDef(cardWithEffect(LayerAbility, ability))
			if !hasCardDefIssue(issues, CardDefIssueInvalidAbilityBody) {
				t.Fatalf("issues = %+v, want %s", issues, CardDefIssueInvalidAbilityBody)
			}
		})
	}
	if issues := ValidateCardDef(cardWithEffect(LayerType, TapAnyColorManaAbility())); !hasCardDefIssue(issues, CardDefIssueInvalidAbilityBody) {
		t.Fatalf("wrong-layer granted mana issues = %+v, want %s", issues, CardDefIssueInvalidAbilityBody)
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
			name:    "cumulative upkeep without explicit cost",
			ability: CumulativeUpkeepKeyword{},
			code:    CardDefIssueInvalidKeywordAbility,
		},
		{
			name:    "flashback without explicit cost",
			ability: FlashbackKeyword{},
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

func TestValidateCardDefAcceptsCumulativeUpkeepTemplate(t *testing.T) {
	t.Parallel()
	card := &CardDef{CardFace: CardFace{
		Name: "Cumulative Upkeep Card",
		TriggeredAbilities: []TriggeredAbility{
			CumulativeUpkeepTriggeredAbility(cost.Mana{cost.O(1), cost.U}),
		},
	}}
	if issues := ValidateCardDef(card); len(issues) != 0 {
		t.Fatalf("issues = %+v; want none", issues)
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

func TestValidateCardDefAllowsNoMaximumHandSizeRuleEffect(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name:            "Reliquary Tester",
		StaticAbilities: []StaticAbility{NoMaximumHandSizeStaticBody},
	}}
	if issues := ValidateCardDef(card); len(issues) != 0 {
		t.Fatalf("issues = %+v, want none", issues)
	}
}

func TestValidateCardDefRejectsInvalidNoMaximumHandSizeRuleEffect(t *testing.T) {
	tests := []struct {
		name   string
		effect RuleEffect
	}{
		{
			name:   "affects any player",
			effect: RuleEffect{Kind: RuleEffectNoMaximumHandSize, AffectedPlayer: PlayerAny},
		},
		{
			name:   "scoped to source",
			effect: RuleEffect{Kind: RuleEffectNoMaximumHandSize, AffectedPlayer: PlayerYou, AffectedSource: true},
		},
		{
			name:   "scoped to attached",
			effect: RuleEffect{Kind: RuleEffectNoMaximumHandSize, AffectedPlayer: PlayerYou, AffectedAttached: true},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			card := &CardDef{CardFace: CardFace{
				Name:            "Reliquary Tester",
				StaticAbilities: []StaticAbility{{RuleEffects: []RuleEffect{tc.effect}}},
			}}
			issues := ValidateCardDef(card)
			if len(issues) == 0 {
				t.Fatalf("issues = none, want %v", CardDefIssueInvalidRuleEffect)
			}
			for _, issue := range issues {
				if issue.Code != CardDefIssueInvalidRuleEffect {
					t.Fatalf("issue code = %v, want %v", issue.Code, CardDefIssueInvalidRuleEffect)
				}
			}
		})
	}
}

func TestValidateCardDefRejectsChosenTypeTriggerMultiplierPayload(t *testing.T) {
	card := &CardDef{CardFace: CardFace{
		Name:  "Invalid Trigger Multiplier",
		Types: []types.Card{types.Creature},
		StaticAbilities: []StaticAbility{{
			RuleEffects: []RuleEffect{{
				Kind:           RuleEffectAdditionalTriggerForChosenCreatureType,
				AffectedPlayer: PlayerYou,
			}},
		}},
	}}

	issues := ValidateCardDef(card)
	if !hasCardDefIssue(issues, CardDefIssueInvalidRuleEffect) {
		t.Fatalf("issues = %+v, want %s", issues, CardDefIssueInvalidRuleEffect)
	}
}

func TestValidateCardDefAttackTaxRuleEffect(t *testing.T) {
	valid := RuleEffect{
		Kind:             RuleEffectAttackTax,
		AffectedPlayer:   PlayerYou,
		AttackTaxGeneric: 2,
	}
	card := func(effect RuleEffect) *CardDef {
		return &CardDef{CardFace: CardFace{
			Name:            "Attack Tax Tester",
			StaticAbilities: []StaticAbility{{RuleEffects: []RuleEffect{effect}}},
		}}
	}
	if issues := ValidateCardDef(card(valid)); len(issues) != 0 {
		t.Fatalf("valid issues = %+v, want none", issues)
	}

	tests := map[string]RuleEffect{
		"missing affected player": {Kind: RuleEffectAttackTax, AttackTaxGeneric: 2},
		"unknown affected player": {Kind: RuleEffectAttackTax, AffectedPlayer: PlayerRelation(99), AttackTaxGeneric: 2},
		"zero amount":             {Kind: RuleEffectAttackTax, AffectedPlayer: PlayerYou},
		"negative amount":         {Kind: RuleEffectAttackTax, AffectedPlayer: PlayerYou, AttackTaxGeneric: -1},
		"permanent scoped":        {Kind: RuleEffectAttackTax, AffectedPlayer: PlayerYou, AttackTaxGeneric: 2, AffectedSource: true},
	}
	for name, effect := range tests {
		t.Run(name, func(t *testing.T) {
			issues := ValidateCardDef(card(effect))
			if !hasCardDefIssue(issues, CardDefIssueInvalidRuleEffect) {
				t.Fatalf("issues = %+v, want %s", issues, CardDefIssueInvalidRuleEffect)
			}
		})
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
