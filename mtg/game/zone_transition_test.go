package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"
)

func TestZoneTransitionDeclarationValidation(t *testing.T) {
	t.Parallel()
	def := &CardDef{CardFace: CardFace{Name: "Counter retention", StaticAbilities: []StaticAbility{{
		ZoneTransitionPolicy: ZoneTransitionRetainCountersExceptHandLibrary,
	}}}}
	if issues := ValidateCardDef(def); len(issues) != 0 {
		t.Fatalf("valid typed retention: %+v", issues)
	}
	def.StaticAbilities[0].ZoneTransitionPolicy = ZoneTransitionPolicy(99)
	if issues := ValidateCardDef(def); len(issues) == 0 {
		t.Fatal("unknown policy was silently accepted")
	}
}

func TestCardZoneAbilityEffectValidation(t *testing.T) {
	t.Parallel()
	def := &CardDef{CardFace: CardFace{Name: "Graveyard ability loss", StaticAbilities: []StaticAbility{{
		ContinuousEffects: []ContinuousEffect{{Layer: LayerAbility, CardZone: zone.Graveyard, RemoveAllAbilities: true}},
	}}}}
	if issues := ValidateCardDef(def); len(issues) != 0 {
		t.Fatalf("valid source-zone ability loss: %+v", issues)
	}
	effect := &def.StaticAbilities[0].ContinuousEffects[0]
	effect.Layer = LayerPowerToughnessModify
	if issues := ValidateCardDef(def); len(issues) == 0 {
		t.Fatal("unsupported off-battlefield layer silently accepted")
	}
	effect.Layer = LayerAbility
	effect.CardZone = zone.Battlefield
	if issues := ValidateCardDef(def); len(issues) == 0 {
		t.Fatal("card-zone declaration bypassed permanent ability layers")
	}
}
