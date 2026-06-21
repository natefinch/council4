package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerModifiedCreaturesAnthem covers the "Modified creatures you control
// have <keyword>" anthem (Envoy of the Ancestors). The continuous effect must
// grant the keyword to the controller's creatures whose runtime Selection
// carries MatchModified so only modified permanents (a counter, Aura, or
// Equipment) are affected.
func TestLowerModifiedCreaturesAnthem(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Envoy of the Ancestors",
		Layout:     "normal",
		TypeLine:   "Creature — Human Cleric",
		OracleText: "Modified creatures you control have lifelink.",
	})
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %#v, want one", face.StaticAbilities)
	}
	effects := face.StaticAbilities[0].Body.ContinuousEffects
	if len(effects) != 1 {
		t.Fatalf("continuous effects = %#v, want one", effects)
	}
	effect := effects[0]
	if effect.Layer != game.LayerAbility {
		t.Fatalf("layer = %v, want LayerAbility", effect.Layer)
	}
	if effect.Group.Domain() != game.GroupDomainObjectControlled {
		t.Fatalf("group domain = %v, want GroupDomainObjectControlled", effect.Group.Domain())
	}
	selection := effect.Group.Selection()
	if !selection.MatchModified {
		t.Fatalf("selection = %#v, want MatchModified", selection)
	}
	if !slices.Equal(selection.RequiredTypes, []types.Card{types.Creature}) {
		t.Fatalf("required types = %v, want [Creature]", selection.RequiredTypes)
	}
	if !slices.Equal(effect.AddKeywords, []game.Keyword{game.Lifelink}) {
		t.Fatalf("keywords = %v, want [Lifelink]", effect.AddKeywords)
	}
}

// TestLowerCantAttackOrBlockUnlessLands covers the land-gated combat
// restriction "<this> can't attack or block unless you control N or more lands"
// (Topiary Stomper). The static ability prohibits both attacking and blocking
// for the source, gated by a negated controls-N-lands condition so the
// restriction lifts once the controller reaches the land count.
func TestLowerCantAttackOrBlockUnlessLands(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Topiary Stomper",
		Layout:     "normal",
		TypeLine:   "Creature — Plant Dinosaur",
		OracleText: "This creature can't attack or block unless you control seven or more lands.",
	})
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %#v, want one", face.StaticAbilities)
	}
	body := face.StaticAbilities[0].Body
	kinds := make([]game.RuleEffectKind, 0, len(body.RuleEffects))
	for _, effect := range body.RuleEffects {
		if !effect.AffectedSource {
			t.Fatalf("rule effect must affect the source: %#v", effect)
		}
		kinds = append(kinds, effect.Kind)
	}
	if !slices.Contains(kinds, game.RuleEffectCantAttack) ||
		!slices.Contains(kinds, game.RuleEffectCantBlock) {
		t.Fatalf("rule effect kinds = %v, want CantAttack and CantBlock", kinds)
	}
	if !body.Condition.Exists {
		t.Fatal("static ability must carry a guard condition")
	}
	condition := body.Condition.Val
	if !condition.Negate {
		t.Fatalf("condition = %#v, want negated", condition)
	}
	if !condition.ControlsMatching.Exists {
		t.Fatalf("condition = %#v, want ControlsMatching", condition)
	}
	controls := condition.ControlsMatching.Val
	if controls.MinCount != 7 {
		t.Fatalf("controls min count = %d, want 7", controls.MinCount)
	}
	if !slices.Equal(controls.Selection.RequiredTypes, []types.Card{types.Land}) {
		t.Fatalf("controls selection = %#v, want Land", controls.Selection)
	}
}

// TestLowerPreventAllCombatDamageActivated covers the activated ability "Prevent
// all combat damage that would be dealt this turn" (Spike Weaver). It lowers to
// a global combat-damage prevention shield, not one bound to a particular
// permanent or player.
func TestLowerPreventAllCombatDamageActivated(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Spike Weaver",
		Layout:     "normal",
		TypeLine:   "Creature — Spike",
		OracleText: "{1}, Remove a +1/+1 counter from this creature: Prevent all combat damage that would be dealt this turn.",
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %#v, want one", face.ActivatedAbilities)
	}
	modes := face.ActivatedAbilities[0].Content.Modes
	if len(modes) != 1 || len(modes[0].Sequence) != 1 {
		t.Fatalf("ability content = %#v, want one mode with one instruction", modes)
	}
	prevent, ok := modes[0].Sequence[0].Primitive.(game.PreventDamage)
	if !ok {
		t.Fatalf("primitive = %#v, want game.PreventDamage", modes[0].Sequence[0].Primitive)
	}
	if !prevent.Global || !prevent.All || !prevent.CombatOnly {
		t.Fatalf("prevent damage = %#v, want global all combat", prevent)
	}
}
