package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
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
	}
	for _, kind := range valid {
		if !kind.Valid() {
			t.Errorf("kind %d rejected", kind)
		}
	}

	invalid := []RuleEffectKind{
		RuleEffectNone,
		-1,
		RuleEffectCantBeSacrificed + 1,
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
		"future":       RuleEffectCantBeSacrificed + 1,
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
		"future":       RuleEffectCantBeSacrificed + 1,
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
