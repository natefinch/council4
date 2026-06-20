package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/cost"
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

	if issues := ValidateCardDef(cardWithEffect(LayerAbility, TapAnyColorManaAbility())); len(issues) != 0 {
		t.Fatalf("canonical granted mana ability issues = %+v, want none", issues)
	}
	if issues := ValidateCardDef(cardWithEffect(LayerAbility, TapManaAbility(mana.G))); !hasCardDefIssue(issues, CardDefIssueInvalidAbilityBody) {
		t.Fatalf("fixed-color granted mana issues = %+v, want %s", issues, CardDefIssueInvalidAbilityBody)
	}
	wrongCost := TapAnyColorManaAbility()
	wrongCost.AdditionalCosts = nil
	wrongCost.ManaCost = opt.Val(cost.Mana{cost.O(1)})
	if issues := ValidateCardDef(cardWithEffect(LayerAbility, wrongCost)); !hasCardDefIssue(issues, CardDefIssueInvalidAbilityBody) {
		t.Fatalf("wrong-cost granted mana issues = %+v, want %s", issues, CardDefIssueInvalidAbilityBody)
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
