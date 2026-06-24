package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

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

func TestValidateCardDefDelayedCapturedTargetControllerBounds(t *testing.T) {
	for _, test := range []struct {
		name      string
		index     int
		allow     TargetAllow
		wantIssue bool
	}{
		{name: "valid", index: 0, allow: TargetAllowStackObject},
		{name: "out of range", index: 1, allow: TargetAllowStackObject, wantIssue: true},
		{name: "non-stack target", index: 0, allow: TargetAllowPlayer, wantIssue: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			target := TargetSpec{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      test.allow,
			}
			if test.allow == TargetAllowStackObject {
				target.Predicate.StackObjectKinds = []StackObjectKind{StackSpell}
			}
			card := &CardDef{CardFace: CardFace{
				Name:       "Delayed Captured Controller",
				OracleText: "Counter target spell. Its controller draws a card at the beginning of the next upkeep.",
				SpellAbility: opt.Val(Mode{
					Targets: []TargetSpec{target},
					Sequence: []Instruction{{Primitive: CreateDelayedTrigger{Trigger: DelayedTriggerDef{
						Timing: DelayedAtBeginningOfNextUpkeep,
						Content: Mode{Sequence: []Instruction{{Primitive: Draw{
							Amount: Fixed(1),
							Player: CapturedTargetControllerReference(test.index),
						}}}}.Ability(),
					}}}},
				}.Ability()),
			}}

			issues := ValidateCardDef(card)
			if got := hasCardDefIssue(issues, CardDefIssueInvalidAbilityBody); got != test.wantIssue {
				t.Fatalf("issues = %+v, invalid ability body = %v, want %v", issues, got, test.wantIssue)
			}
		})
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
			name: "spell type union",
			spec: TargetSpec{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      TargetAllowStackObject,
				Predicate: TargetPredicate{
					SpellCardTypesAny: []types.Card{types.Instant, types.Sorcery},
					StackObjectKinds:  []StackObjectKind{StackSpell},
				},
			},
		},
		{
			name: "single member spell type union",
			spec: TargetSpec{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      TargetAllowStackObject,
				Predicate: TargetPredicate{
					SpellCardTypesAny: []types.Card{types.Instant},
					StackObjectKinds:  []StackObjectKind{StackSpell},
				},
			},
			wantIssue: true,
		},
		{
			name: "duplicate spell type union member",
			spec: TargetSpec{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      TargetAllowStackObject,
				Predicate: TargetPredicate{
					SpellCardTypesAny: []types.Card{types.Instant, types.Instant},
					StackObjectKinds:  []StackObjectKind{StackSpell},
				},
			},
			wantIssue: true,
		},
		{
			name: "mixed all and any spell types",
			spec: TargetSpec{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      TargetAllowStackObject,
				Predicate: TargetPredicate{
					SpellCardTypes:    []types.Card{types.Instant},
					SpellCardTypesAny: []types.Card{types.Instant, types.Sorcery},
					StackObjectKinds:  []StackObjectKind{StackSpell},
				},
			},
			wantIssue: true,
		},
		{
			name: "stack target with permanent-only predicate",
			spec: TargetSpec{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      TargetAllowStackObject,
				Predicate: TargetPredicate{
					StackObjectKinds: []StackObjectKind{StackActivatedAbility},
					PermanentTypes:   []types.Card{types.Creature},
				},
			},
			wantIssue: true,
		},
		{
			name: "stack target with controller restriction",
			spec: TargetSpec{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      TargetAllowStackObject,
				Predicate: TargetPredicate{
					StackObjectKinds: []StackObjectKind{StackActivatedAbility, StackTriggeredAbility},
					Controller:       ControllerNotYou,
				},
			},
		},
		{
			name: "stack target with source-type restriction",
			spec: TargetSpec{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      TargetAllowStackObject,
				Predicate: TargetPredicate{
					StackObjectKinds:       []StackObjectKind{StackActivatedAbility},
					StackObjectSourceTypes: []types.Card{types.Artifact},
				},
			},
		},
		{
			name: "mixed stack target with spell supertype",
			spec: TargetSpec{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      TargetAllowStackObject,
				Predicate: TargetPredicate{
					StackObjectKinds: []StackObjectKind{StackSpell, StackActivatedAbility, StackTriggeredAbility},
					SpellSupertypes:  []types.Super{types.Legendary},
				},
			},
		},
		{
			name: "spell supertype without spell kind",
			spec: TargetSpec{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      TargetAllowStackObject,
				Predicate: TargetPredicate{
					StackObjectKinds: []StackObjectKind{StackActivatedAbility},
					SpellColorless:   true,
				},
			},
			wantIssue: true,
		},
		{
			name: "mixed stack target with controller restriction",
			spec: TargetSpec{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      TargetAllowPermanent | TargetAllowStackObject,
				Predicate: TargetPredicate{
					StackObjectKinds: []StackObjectKind{StackActivatedAbility},
					Controller:       ControllerOpponent,
				},
			},
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
			name: "non-cast trigger card with power",
			face: CardFace{TriggeredAbilities: []TriggeredAbility{{
				Content: Mode{}.Ability(),
				Trigger: TriggerCondition{Pattern: TriggerPattern{
					Event:         EventCardDrawn,
					CardSelection: Selection{Power: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 2})},
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

// TestValidateCardDefMultiTargetSpecAdmitsPerSlotReferences proves a single
// multi-target spec admits object references for every slot it can hold
// (TargetPermanentReference(0)..(MaxTargets-1)), the shape the cardgen backend
// emits for "Exile up to N target permanents.", while still rejecting a
// reference beyond the spec's capacity.
func TestValidateCardDefMultiTargetSpecAdmitsPerSlotReferences(t *testing.T) {
	multiExile := func(refIndexes ...int) *CardDef {
		sequence := make([]Instruction, 0, len(refIndexes))
		for _, index := range refIndexes {
			sequence = append(sequence, Instruction{
				Primitive: Exile{Object: TargetPermanentReference(index)},
			})
		}
		return &CardDef{CardFace: CardFace{
			Name: "Multi Exile",
			SpellAbility: opt.Val(Mode{
				Targets: []TargetSpec{{
					MinTargets: 0,
					MaxTargets: 3,
					Allow:      TargetAllowPermanent,
					Predicate:  TargetPredicate{PermanentTypes: []types.Card{types.Enchantment}},
				}},
				Sequence: sequence,
			}.Ability()),
		}}
	}

	if issues := ValidateCardDef(multiExile(0, 1, 2)); len(issues) != 0 {
		t.Fatalf("issues = %+v, want none for in-capacity per-slot references", issues)
	}

	if issues := ValidateCardDef(multiExile(0, 3)); !hasCardDefIssue(issues, CardDefIssueInvalidAbilityBody) {
		t.Fatalf("issues = %+v, want %s for a reference beyond capacity", issues, CardDefIssueInvalidAbilityBody)
	}
}
