package game

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestRuleEffectKindValid(t *testing.T) {
	t.Parallel()

	valid := []RuleEffectKind{
		RuleEffectCantGainLife,
		RuleEffectCantAttack,
		RuleEffectCantBlock,
		RuleEffectCostModifier,
		RuleEffectCastFromZone,
		RuleEffectCantBeCountered,
		RuleEffectCantBeBlocked,
		RuleEffectMustBeBlocked,
		RuleEffectMustAttack,
		RuleEffectGrantHandCardAbility,
		RuleEffectDoesntUntap,
		RuleEffectCantBeBlockedByMoreThanOne,
		RuleEffectNoMaximumHandSize,
		RuleEffectCantBeBlockedByCreaturesWith,
		RuleEffectCombatDamageCantBePrevented,
		RuleEffectPlayerProtection,
		RuleEffectAttackTax,
		RuleEffectLifeTotalCantChange,
		RuleEffectPlayFromZone,
		RuleEffectAdditionalTriggerForChosenCreatureType,
		RuleEffectAdditionalLandPlays,
		RuleEffectCantCastSpells,
		RuleEffectCantActivateAbilities,
		RuleEffectAdditionalTriggerForEnteringPermanent,
		RuleEffectUntapDuringOtherPlayersUntapStep,
		RuleEffectCastSpellsAsThoughFlash,
		RuleEffectPlayLandsFromZone,
		RuleEffectPlayWithTopCardRevealed,
		RuleEffectCastSpellsFromZone,
		RuleEffectCantCastFromZones,
		RuleEffectCantEnterFromZones,
		RuleEffectLookAtTopCardAnyTime,
		RuleEffectPayLifeForColoredMana,
		RuleEffectPayLifeForCommanderTax,
		RuleEffectDrawLimitPerTurn,
		RuleEffectCastLimitPerTurn,
		RuleEffectAdditionalTriggerForControlledPermanent,
		RuleEffectMustBeBlockedByAllAble,
		RuleEffectAssignCombatDamageAsThoughUnblocked,
		RuleEffectCantTransform,
		RuleEffectSuppressOpponentEnteringTriggers,
		RuleEffectAttackTaxPerCreature,
		RuleEffectManaProductionMultiplier,
		RuleEffectSkipDrawStep,
		RuleEffectCanBlockOnlyCreaturesWith,
		RuleEffectCantAttackAlone,
		RuleEffectCantBlockAlone,
		RuleEffectCanAttackAsThoughDefender,
		RuleEffectAssignCombatDamageUsingToughness,
		RuleEffectCantActivateAbilitiesOfPermanent,
		RuleEffectGoaded,
		RuleEffectPlayerHexproof,
		RuleEffectPlayerShroud,
		RuleEffectDamageDoesntCauseLifeLoss,
		RuleEffectRedirectDamageToSource,
		RuleEffectCantBeSacrificed,
		RuleEffectCastLinkedExileForFree,
		RuleEffectActivateAbilitiesAsThoughHaste,
		RuleEffectGrantSpellKeyword,
		RuleEffectAscend,
		RuleEffectCantBeTargetedByControllerOpponents,
		RuleEffectSkipExtraTurns,
		RuleEffectAdditionalTriggerForRoomAbility,
		RuleEffectControlOpponentSearches,
		RuleEffectExileOpponentSearchFinds,
		RuleEffectLegendRuleDoesNotApply,
		RuleEffectPlayFromGraveyardPerPermanentType,
		RuleEffectCascadeLandOption,
		RuleEffectDamageCantBePrevented,
	}
	for _, kind := range valid {
		if !kind.Valid() {
			t.Errorf("kind %d rejected", kind)
		}
	}

	invalid := []RuleEffectKind{
		RuleEffectNone,
		-1,
		RuleEffectDamageCantBePrevented + 1,
		RuleEffectKind(1 << 20),
	}
	for _, kind := range invalid {
		if kind.Valid() {
			t.Errorf("unsupported kind %d accepted", kind)
		}
	}
}

func TestValidateApplyRulePlayFromZone(t *testing.T) {
	t.Parallel()

	valid := playFromZoneTestEffect()
	if err := validateApplyRuleTestEffect(&valid); err != nil {
		t.Fatalf("valid play-from-zone effect rejected: %v", err)
	}

	for name, kind := range map[string]RuleEffectKind{
		"future":       RuleEffectDamageCantBePrevented + 1,
		"out of range": RuleEffectKind(1 << 20),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			effect := valid
			effect.Kind = kind
			if err := validateApplyRuleTestEffect(&effect); err == nil {
				t.Fatalf("unsupported kind %d accepted", kind)
			}
		})
	}
}

func TestValidateCardDefPlayFromZone(t *testing.T) {
	t.Parallel()

	valid := playFromZoneTestEffect()
	valid.AffectedCardID = 0
	if issues := ValidateCardDef(cardDefWithRuleEffect(&valid)); len(issues) != 0 {
		t.Fatalf("valid play-from-zone issues = %+v, want none", issues)
	}

	for name, kind := range map[string]RuleEffectKind{
		"future":       RuleEffectDamageCantBePrevented + 1,
		"out of range": RuleEffectKind(1 << 20),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			effect := valid
			effect.Kind = kind
			issues := ValidateCardDef(cardDefWithRuleEffect(&effect))
			if !hasCardDefIssue(issues, CardDefIssueInvalidRuleEffect) {
				t.Fatalf("issues = %+v, want %s", issues, CardDefIssueInvalidRuleEffect)
			}
		})
	}
}

func TestValidatePlayFromZoneStructure(t *testing.T) {
	t.Parallel()

	applyRuleTests := map[string]RuleEffect{
		"unknown affected player": func() RuleEffect {
			effect := playFromZoneTestEffect()
			effect.AffectedPlayer = PlayerRelation(99)
			return effect
		}(),
		"missing source zone": func() RuleEffect {
			effect := playFromZoneTestEffect()
			effect.CastFromZone = zone.None
			return effect
		}(),
		"unsupported source zone": func() RuleEffect {
			effect := playFromZoneTestEffect()
			effect.CastFromZone = zone.Graveyard
			return effect
		}(),
		"missing card": func() RuleEffect {
			effect := playFromZoneTestEffect()
			effect.AffectedCardID = 0
			return effect
		}(),
		"restricted face": func() RuleEffect {
			effect := playFromZoneTestEffect()
			effect.CastFace.Exists = true
			return effect
		}(),
		"permanent scoped": func() RuleEffect {
			effect := playFromZoneTestEffect()
			effect.AffectedSource = true
			return effect
		}(),
	}
	for name, effect := range applyRuleTests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if err := validateApplyRuleTestEffect(&effect); err == nil {
				t.Fatal("ApplyRule accepted invalid play-from-zone structure")
			}
		})
	}

	cardDefTests := map[string]RuleEffect{
		"unknown affected player": func() RuleEffect {
			effect := playFromZoneTestEffect()
			effect.AffectedCardID = 0
			effect.AffectedPlayer = PlayerRelation(99)
			return effect
		}(),
		"missing source zone": func() RuleEffect {
			effect := playFromZoneTestEffect()
			effect.AffectedCardID = 0
			effect.CastFromZone = zone.None
			return effect
		}(),
		"unsupported source zone": func() RuleEffect {
			effect := playFromZoneTestEffect()
			effect.AffectedCardID = 0
			effect.CastFromZone = zone.Graveyard
			return effect
		}(),
		"runtime card ID": playFromZoneTestEffect(),
		"restricted face": func() RuleEffect {
			effect := playFromZoneTestEffect()
			effect.AffectedCardID = 0
			effect.CastFace.Exists = true
			return effect
		}(),
		"permanent scoped": func() RuleEffect {
			effect := playFromZoneTestEffect()
			effect.AffectedCardID = 0
			effect.AffectedSource = true
			return effect
		}(),
	}
	for name, effect := range cardDefTests {
		t.Run("CardDef/"+name, func(t *testing.T) {
			t.Parallel()
			issues := ValidateCardDef(cardDefWithRuleEffect(&effect))
			if !hasCardDefIssue(issues, CardDefIssueInvalidRuleEffect) {
				t.Fatalf("CardDef issues = %+v, want %s", issues, CardDefIssueInvalidRuleEffect)
			}
		})
	}
}

func playFromZoneTestEffect() RuleEffect {
	return RuleEffect{
		Kind:           RuleEffectPlayFromZone,
		AffectedPlayer: PlayerYou,
		CastFromZone:   zone.Exile,
		AffectedCardID: id.ID(1),
	}
}

func validateApplyRuleTestEffect(effect *RuleEffect) error {
	return ValidateInstructionSequence([]Instruction{{
		Primitive: ApplyRule{RuleEffects: []RuleEffect{*effect}},
	}})
}

func TestValidateCardDefActionRestriction(t *testing.T) {
	t.Parallel()

	valid := []RuleEffect{
		{Kind: RuleEffectCantCastSpells, AffectedPlayer: PlayerOpponent, RestrictedDuringControllerTurn: true},
		{Kind: RuleEffectCantCastSpells, AffectedPlayer: PlayerAny},
		{Kind: RuleEffectCantActivateAbilities, AffectedPlayer: PlayerOpponent, PermanentTypes: []types.Card{types.Artifact, types.Creature, types.Enchantment}},
	}
	for _, effect := range valid {
		if issues := ValidateCardDef(cardDefWithRuleEffect(&effect)); len(issues) != 0 {
			t.Fatalf("valid action restriction %d issues = %+v, want none", effect.Kind, issues)
		}
	}

	invalid := map[string]RuleEffect{
		"cast prohibition affecting a permanent":  {Kind: RuleEffectCantCastSpells, AffectedPlayer: PlayerOpponent, AffectedSource: true},
		"cast prohibition with permanent types":   {Kind: RuleEffectCantCastSpells, AffectedPlayer: PlayerOpponent, PermanentTypes: []types.Card{types.Creature}},
		"activation prohibition with spell types": {Kind: RuleEffectCantActivateAbilities, AffectedPlayer: PlayerOpponent, SpellTypes: []types.Card{types.Instant}},
		"unknown affected player":                 {Kind: RuleEffectCantCastSpells, AffectedPlayer: PlayerRelation(99)},
	}
	for name, effect := range invalid {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if issues := ValidateCardDef(cardDefWithRuleEffect(&effect)); !hasCardDefIssue(issues, CardDefIssueInvalidRuleEffect) {
				t.Fatalf("issues = %+v, want %s", issues, CardDefIssueInvalidRuleEffect)
			}
		})
	}
}

func cardDefWithRuleEffect(effect *RuleEffect) *CardDef {
	return &CardDef{CardFace: CardFace{
		Name:            "Play Permission Tester",
		StaticAbilities: []StaticAbility{{RuleEffects: []RuleEffect{*effect}}},
	}}
}

func TestValidateGrantSpellKeywordRuleEffect(t *testing.T) {
	t.Parallel()

	valid := []RuleEffect{
		{Kind: RuleEffectGrantSpellKeyword, AffectedController: ControllerYou, GrantedKeyword: Improvise},
		{Kind: RuleEffectGrantSpellKeyword, AffectedController: ControllerYou, GrantedKeyword: Convoke, CardSelection: Selection{ExcludedTypes: []types.Card{types.Artifact}}},
		{Kind: RuleEffectGrantSpellKeyword, AffectedController: ControllerYou, GrantedKeyword: Improvise, AppliesToNextSpellOnly: true},
	}
	for _, effect := range valid {
		if issues := ValidateCardDef(cardDefWithRuleEffect(&effect)); len(issues) != 0 {
			t.Fatalf("valid spell keyword grant %+v issues = %+v, want none", effect, issues)
		}
	}

	invalid := map[string]RuleEffect{
		"unset controller":     {Kind: RuleEffectGrantSpellKeyword, GrantedKeyword: Improvise},
		"affects a permanent":  {Kind: RuleEffectGrantSpellKeyword, AffectedController: ControllerYou, GrantedKeyword: Improvise, AffectedSource: true},
		"no granted keyword":   {Kind: RuleEffectGrantSpellKeyword, AffectedController: ControllerYou},
		"unsupported keyword":  {Kind: RuleEffectGrantSpellKeyword, AffectedController: ControllerYou, GrantedKeyword: Flying},
		"nonprinted predicate": {Kind: RuleEffectGrantSpellKeyword, AffectedController: ControllerYou, GrantedKeyword: Improvise, CardSelection: Selection{NonToken: true}},
	}
	for name, effect := range invalid {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if issues := ValidateCardDef(cardDefWithRuleEffect(&effect)); !hasCardDefIssue(issues, CardDefIssueInvalidRuleEffect) && !hasCardDefIssue(issues, CardDefIssueInvalidSelection) {
				t.Fatalf("issues = %+v, want an invalid rule-effect or selection issue", issues)
			}
		})
	}
}

func TestValidatePayLifeForCommanderTaxRuleEffect(t *testing.T) {
	t.Parallel()

	valid := RuleEffect{
		Kind:           RuleEffectPayLifeForCommanderTax,
		AffectedPlayer: PlayerYou,
		AffectedSource: true,
	}
	if issues := ValidateCardDef(cardDefWithRuleEffect(&valid)); len(issues) != 0 {
		t.Fatalf("valid life-for-commander-tax issues = %+v, want none", issues)
	}

	for name, mutate := range map[string]func(*RuleEffect){
		"wrong player": func(e *RuleEffect) { e.AffectedPlayer = PlayerOpponent },
		"not self":     func(e *RuleEffect) { e.AffectedSource = false },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			effect := valid
			mutate(&effect)
			issues := ValidateCardDef(cardDefWithRuleEffect(&effect))
			if !hasCardDefIssue(issues, CardDefIssueInvalidRuleEffect) {
				t.Fatalf("issues = %+v, want %s", issues, CardDefIssueInvalidRuleEffect)
			}
		})
	}
}

func TestValidateDamageCantBePreventedRuleEffect(t *testing.T) {
	t.Parallel()

	creatures := Selection{RequiredTypes: []types.Card{types.Creature}}
	valid := []RuleEffect{
		{Kind: RuleEffectDamageCantBePrevented},
		{Kind: RuleEffectDamageCantBePrevented, AffectedSource: true},
		{Kind: RuleEffectDamageCantBePrevented, AffectedSelection: creatures},
		{Kind: RuleEffectDamageCantBePrevented, Duration: DurationUntilEndOfTurn},
	}
	for _, effect := range valid {
		if issues := ValidateCardDef(cardDefWithRuleEffect(&effect)); len(issues) != 0 {
			t.Errorf("valid static %+v issues = %+v, want none", effect, issues)
		}
		if err := validateApplyRuleTestEffect(&effect); err != nil {
			t.Errorf("valid instruction %+v rejected: %v", effect, err)
		}
	}
	scopedObject := ValidateInstructionSequence([]Instruction{{
		Primitive: ApplyRule{
			Object:      opt.Val(TargetPermanentReference(0)),
			RuleEffects: []RuleEffect{{Kind: RuleEffectDamageCantBePrevented}},
			Duration:    DurationUntilEndOfTurn,
		},
	}})
	if scopedObject != nil {
		t.Errorf("object-scoped instruction rejected: %v", scopedObject)
	}

	invalid := map[string]RuleEffect{
		"two scopes":      {Kind: RuleEffectDamageCantBePrevented, AffectedSource: true, AffectedSelection: creatures},
		"controller":      {Kind: RuleEffectDamageCantBePrevented, AffectedController: ControllerYou},
		"permanent types": {Kind: RuleEffectDamageCantBePrevented, PermanentTypes: []types.Card{types.Creature}},
		"affected player": {Kind: RuleEffectDamageCantBePrevented, AffectedPlayer: PlayerOpponent},
		"attached":        {Kind: RuleEffectDamageCantBePrevented, AffectedAttached: true},
	}
	for name, effect := range invalid {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if !hasCardDefIssue(ValidateCardDef(cardDefWithRuleEffect(&effect)), CardDefIssueInvalidRuleEffect) {
				t.Fatalf("static %+v accepted", effect)
			}
			if err := validateApplyRuleTestEffect(&effect); err == nil {
				t.Fatalf("instruction %+v accepted", effect)
			}
		})
	}
	sourceAndObject := ValidateInstructionSequence([]Instruction{{
		Primitive: ApplyRule{
			Object:      opt.Val(TargetPermanentReference(0)),
			RuleEffects: []RuleEffect{{Kind: RuleEffectDamageCantBePrevented, AffectedSource: true}},
		},
	}})
	if sourceAndObject == nil {
		t.Error("instruction with both source and object scopes accepted")
	}
}

func TestValidateCastEntryCountersRider(t *testing.T) {
	t.Parallel()

	finality := []CounterPlacement{{Kind: counter.Finality, Amount: 1}}
	for _, valid := range []RuleEffect{
		{Kind: RuleEffectCastFromZone, AffectedSource: true, Duration: DurationPermanent, AffectedPlayer: PlayerYou, CastFromZone: zone.Graveyard, CastEntryCounters: finality},
		{Kind: RuleEffectCastSpellsFromZone, Duration: DurationPermanent, AffectedPlayer: PlayerYou, CastFromZone: zone.Graveyard, SpellTypes: []types.Card{types.Creature}, CastEntryCounters: finality},
	} {
		if issues := ValidateCardDef(cardDefWithRuleEffect(&valid)); len(issues) != 0 {
			t.Fatalf("valid rider issues = %+v, want none", issues)
		}
	}
	base := RuleEffect{Kind: RuleEffectCastFromZone, AffectedSource: true, Duration: DurationPermanent, AffectedPlayer: PlayerYou, CastFromZone: zone.Graveyard}
	for name, mutate := range map[string]func(*RuleEffect){
		"wrong kind": func(e *RuleEffect) {
			e.Kind, e.CastFromZone = RuleEffectNoMaximumHandSize, 0
			e.CastEntryCounters = finality
		},
		"zero amount": func(e *RuleEffect) { e.CastEntryCounters = []CounterPlacement{{Kind: counter.Finality}} },
		"x amount": func(e *RuleEffect) {
			e.CastEntryCounters = []CounterPlacement{{Kind: counter.Finality, Amount: 1, AmountFromX: true}}
		},
		"player kind": func(e *RuleEffect) { e.CastEntryCounters = []CounterPlacement{{Kind: counter.Poison, Amount: 1}} },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			effect := base
			mutate(&effect)
			for _, issue := range ValidateCardDef(cardDefWithRuleEffect(&effect)) {
				if issue.Code == CardDefIssueInvalidRuleEffect && strings.Contains(issue.Path, "CastEntryCounters") {
					return
				}
			}
			t.Fatal("invalid entry-counter rider was not reported")
		})
	}
}
