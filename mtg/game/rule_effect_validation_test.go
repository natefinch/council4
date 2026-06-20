package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/id"
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
	}
	for _, kind := range valid {
		if !kind.Valid() {
			t.Errorf("kind %d rejected", kind)
		}
	}

	invalid := []RuleEffectKind{
		RuleEffectNone,
		-1,
		RuleEffectPlayFromZone + 1,
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
		"future":       RuleEffectPlayFromZone + 1,
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
		"future":       RuleEffectPlayFromZone + 1,
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

func cardDefWithRuleEffect(effect *RuleEffect) *CardDef {
	return &CardDef{CardFace: CardFace{
		Name:            "Play Permission Tester",
		StaticAbilities: []StaticAbility{{RuleEffects: []RuleEffect{*effect}}},
	}}
}
