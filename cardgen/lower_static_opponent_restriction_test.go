package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestLowerOpponentActionRestrictionGrandAbolisher proves the Grand Abolisher
// static "During your turn, your opponents can't cast spells or activate
// abilities of artifacts, creatures, or enchantments." lowers to a cast
// prohibition and a type-scoped activation prohibition, both scoped to the
// controller's turn and to the controller's opponents.
func TestLowerOpponentActionRestrictionGrandAbolisher(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Grand Abolisher",
		Layout:     "normal",
		TypeLine:   "Creature — Human Cleric",
		OracleText: "During your turn, your opponents can't cast spells or activate abilities of artifacts, creatures, or enchantments.",
	})
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %d, want 1", len(face.StaticAbilities))
	}
	effects := face.StaticAbilities[0].Body.RuleEffects
	if len(effects) != 2 {
		t.Fatalf("rule effects = %#v, want two", effects)
	}
	cast := effects[0]
	if cast.Kind != game.RuleEffectCantCastSpells ||
		cast.AffectedPlayer != game.PlayerOpponent ||
		!cast.RestrictedDuringControllerTurn ||
		len(cast.SpellTypes) != 0 {
		t.Fatalf("cast prohibition = %#v", cast)
	}
	activate := effects[1]
	if activate.Kind != game.RuleEffectCantActivateAbilities ||
		activate.AffectedPlayer != game.PlayerOpponent ||
		!activate.RestrictedDuringControllerTurn ||
		!reflect.DeepEqual(activate.PermanentTypes, []types.Card{types.Artifact, types.Creature, types.Enchantment}) {
		t.Fatalf("activation prohibition = %#v", activate)
	}
}

// TestLowerOpponentActionRestrictionVariants proves the wording family lowers to
// the expected affected-player relation, turn scope, and prohibited actions.
func TestLowerOpponentActionRestrictionVariants(t *testing.T) {
	t.Parallel()
	cases := []struct {
		text         string
		affected     game.PlayerRelation
		duringTurn   bool
		restrictCast bool
	}{
		{"Your opponents can't cast spells.", game.PlayerOpponent, false, true},
		{"Players can't cast spells during your turn.", game.PlayerAny, true, true},
		{"Spells can't be cast during your turn.", game.PlayerAny, true, true},
	}
	for _, tc := range cases {
		face := lowerSingleFace(t, &ScryfallCard{
			Name:       "Test Restriction",
			Layout:     "normal",
			TypeLine:   "Enchantment",
			OracleText: tc.text,
		})
		if len(face.StaticAbilities) != 1 {
			t.Fatalf("%q: static abilities = %d, want 1", tc.text, len(face.StaticAbilities))
		}
		effects := face.StaticAbilities[0].Body.RuleEffects
		if len(effects) != 1 {
			t.Fatalf("%q: rule effects = %#v, want one", tc.text, effects)
		}
		effect := effects[0]
		if effect.Kind != game.RuleEffectCantCastSpells ||
			effect.AffectedPlayer != tc.affected ||
			effect.RestrictedDuringControllerTurn != tc.duringTurn {
			t.Fatalf("%q: effect = %#v", tc.text, effect)
		}
	}
}

// TestLowerOpponentActionRestrictionFailsClosed proves an unrecognized inner
// action keeps the static ability unsupported rather than lowering a partial
// prohibition.
func TestLowerOpponentActionRestrictionFailsClosed(t *testing.T) {
	t.Parallel()
	face := lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Test Restriction",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "Your opponents can't gain life.",
	})
	if len(face.StaticAbilities) != 0 {
		t.Fatalf("unexpected static abilities: %#v", face.StaticAbilities)
	}
}

// TestLowerCastZoneRestrictionDrannith proves Drannith Magistrate's "Your
// opponents can't cast spells from anywhere other than their hands." lowers to a
// cast-zone restriction forbidding every non-hand cast zone for the controller's
// opponents.
func TestLowerCastZoneRestrictionDrannith(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Drannith Magistrate",
		Layout:     "normal",
		TypeLine:   "Creature — Human Wizard",
		OracleText: "Your opponents can't cast spells from anywhere other than their hands.",
	})
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %d, want 1", len(face.StaticAbilities))
	}
	effects := face.StaticAbilities[0].Body.RuleEffects
	if len(effects) != 1 {
		t.Fatalf("rule effects = %#v, want one", effects)
	}
	effect := effects[0]
	if effect.Kind != game.RuleEffectCantCastFromZones ||
		effect.AffectedPlayer != game.PlayerOpponent ||
		!reflect.DeepEqual(effect.CantCastFromZones, []zone.Type{zone.Graveyard, zone.Exile, zone.Library, zone.Command}) {
		t.Fatalf("cast-zone restriction = %#v", effect)
	}
}

// TestLowerCastZoneRestrictionExplicitZones proves an explicit zone list lowers
// to exactly those zones for every player.
func TestLowerCastZoneRestrictionExplicitZones(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Restriction",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "Players can't cast spells from graveyards or libraries.",
	})
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %d, want 1", len(face.StaticAbilities))
	}
	effects := face.StaticAbilities[0].Body.RuleEffects
	if len(effects) != 1 {
		t.Fatalf("rule effects = %#v, want one", effects)
	}
	effect := effects[0]
	if effect.Kind != game.RuleEffectCantCastFromZones ||
		effect.AffectedPlayer != game.PlayerAny ||
		!reflect.DeepEqual(effect.CantCastFromZones, []zone.Type{zone.Graveyard, zone.Library}) {
		t.Fatalf("cast-zone restriction = %#v", effect)
	}
}
