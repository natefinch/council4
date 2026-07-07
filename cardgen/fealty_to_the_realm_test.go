package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// fealtyOracleText is the full Oracle text of Fealty to the Realm. The first two
// lines (Enchant creature, become the monarch) already lowered before this card
// was added; the third and fourth lines exercise the dynamic monarch-control
// static and the compound must-attack / can't-attack-you static.
const fealtyOracleText = "Enchant creature\n" +
	"When this Aura enters, you become the monarch.\n" +
	"The monarch controls enchanted creature.\n" +
	"Enchanted creature attacks each combat if able and can't attack you."

// staticWithControlToMonarch returns the lowered control continuous effect that
// makes the monarch control the enchanted creature, failing if none is present.
func staticWithControlToMonarch(t *testing.T, statics []loweredStaticAbility) game.ContinuousEffect {
	t.Helper()
	for _, static := range statics {
		for _, effect := range static.Body.ContinuousEffects {
			if effect.Layer == game.LayerControl && effect.NewControllerIsMonarch {
				return effect
			}
		}
	}
	t.Fatalf("no LayerControl NewControllerIsMonarch effect in %#v", statics)
	return game.ContinuousEffect{}
}

// TestGenerateFealtyToTheRealmDynamicMonarchControl proves ability 2 ("The
// monarch controls enchanted creature.") lowers to a LayerControl continuous
// effect flagged NewControllerIsMonarch and scoped to the enchanted (attached)
// creature.
func TestGenerateFealtyToTheRealmDynamicMonarchControl(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Fealty to the Realm",
		Layout:     "normal",
		TypeLine:   "Enchantment — Aura",
		ManaCost:   "{4}{U}",
		OracleText: fealtyOracleText,
	})

	effect := staticWithControlToMonarch(t, face.StaticAbilities)
	if effect.NewController.Exists || effect.NewControllerRef.Exists {
		t.Fatalf("monarch control effect also set a fixed controller: %#v", effect)
	}
	if effect.Group.Domain() != game.GroupDomainAttachedObject {
		t.Fatalf("control effect group domain = %v, want attached object", effect.Group.Domain())
	}
}

// TestGenerateFealtyToTheRealmCombatRuleEffects proves ability 3 ("Enchanted
// creature attacks each combat if able and can't attack you.") lowers to the
// compound rule effects: a must-attack and a direct-only can't-attack-you, both
// scoped to the enchanted (attached) creature with "you" being the Aura's
// controller.
func TestGenerateFealtyToTheRealmCombatRuleEffects(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Fealty to the Realm",
		Layout:     "normal",
		TypeLine:   "Enchantment — Aura",
		ManaCost:   "{4}{U}",
		OracleText: fealtyOracleText,
	})

	var mustAttack, cantAttack *game.RuleEffect
	for i := range face.StaticAbilities {
		for j := range face.StaticAbilities[i].Body.RuleEffects {
			effect := &face.StaticAbilities[i].Body.RuleEffects[j]
			if effect.Kind == game.RuleEffectMustAttack {
				mustAttack = effect
			}
			if effect.Kind == game.RuleEffectCantAttack {
				cantAttack = effect
			}
		}
	}

	if mustAttack == nil {
		t.Fatalf("no must-attack rule effect in %#v", face.StaticAbilities)
	}
	if !mustAttack.AffectedAttached {
		t.Fatalf("must-attack effect not scoped to attached creature: %#v", mustAttack)
	}

	if cantAttack == nil {
		t.Fatalf("no can't-attack rule effect in %#v", face.StaticAbilities)
	}
	if !cantAttack.AffectedAttached {
		t.Fatalf("can't-attack effect not scoped to attached creature: %#v", cantAttack)
	}
	if cantAttack.DefendingPlayer != game.PlayerYou {
		t.Fatalf("can't-attack defending player = %v, want PlayerYou (the Aura controller)", cantAttack.DefendingPlayer)
	}
	if !cantAttack.DefendingPlayerDirectOnly {
		t.Fatalf("can't-attack effect not direct-only: %#v", cantAttack)
	}
}

// TestGenerateFealtyToTheRealmBecomeMonarchTrigger keeps the pre-existing
// enters trigger covered alongside the new statics: "When this Aura enters, you
// become the monarch." lowers to a BecomeMonarch primitive on the controller.
func TestGenerateFealtyToTheRealmBecomeMonarchTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Fealty to the Realm",
		Layout:     "normal",
		TypeLine:   "Enchantment — Aura",
		ManaCost:   "{4}{U}",
		OracleText: fealtyOracleText,
	})

	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	prim := becomeMonarchPrimitive(t, face.TriggeredAbilities[0].Content)
	if prim.Player.Kind() != game.PlayerReferenceController {
		t.Fatalf("player reference = %v, want controller", prim.Player.Kind())
	}
}
