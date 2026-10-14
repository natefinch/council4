package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"
)

func TestPermanentTypePlayPermissionValidation(t *testing.T) {
	valid := RuleEffect{
		Kind: RuleEffectPlayFromGraveyardPerPermanentType, PermissionKey: "types",
		AffectedPlayer: PlayerYou, CastFromZone: zone.Graveyard, RestrictedDuringControllerTurn: true,
	}
	if issues := ValidateCardDef(cardDefWithRuleEffect(&valid)); len(issues) != 0 {
		t.Fatalf("valid static declaration: %+v", issues)
	}
	for _, edit := range []func(*RuleEffect){
		func(e *RuleEffect) { e.PermissionKey = "" },
		func(e *RuleEffect) { e.AffectedPlayer = PlayerAny },
		func(e *RuleEffect) { e.CastFromZone = zone.Hand },
		func(e *RuleEffect) { e.RestrictedDuringControllerTurn = false },
		func(e *RuleEffect) { e.WithoutPayingManaCost = true },
		func(e *RuleEffect) { e.AffectedSource = true },
		func(e *RuleEffect) { e.AffectedCardID = 42 },
	} {
		effect := valid
		edit(&effect)
		if issues := ValidateCardDef(cardDefWithRuleEffect(&effect)); !hasCardDefIssue(issues, CardDefIssueInvalidRuleEffect) {
			t.Fatalf("malformed declaration accepted: %+v", effect)
		}
	}
	if err := validateApplyRuleTestEffect(&valid); err == nil {
		t.Fatal("static-only permission accepted as an inert resolving instruction")
	}
}
