package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestLowerCastFromTopPayLifeRider verifies the Bolas's Citadel static that lets
// you play lands and cast spells from the top of your library, where the
// trailing "If you cast a spell this way, pay life equal to its mana value …"
// rider marks the cast-from-zone effect to charge life instead of mana.
func TestLowerCastFromTopPayLifeRider(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Test Citadel",
		Layout:   "normal",
		TypeLine: "Artifact",
		OracleText: "You may play lands and cast spells from the top of your library. " +
			"If you cast a spell this way, pay life equal to its mana value rather than pay its mana cost.",
	})
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %d, want 1", len(face.StaticAbilities))
	}
	effects := face.StaticAbilities[0].Body.RuleEffects
	if len(effects) != 2 {
		t.Fatalf("rule effects = %#v, want two", effects)
	}
	cast := effects[0]
	if cast.Kind != game.RuleEffectCastSpellsFromZone ||
		cast.AffectedPlayer != game.PlayerYou ||
		cast.CastFromZone != zone.Library ||
		!cast.TopCardOnly ||
		!cast.PayLifeEqualToManaValue {
		t.Fatalf("cast effect = %#v, want pay-life cast-from-library-top", cast)
	}
	play := effects[1]
	if play.Kind != game.RuleEffectPlayLandsFromZone ||
		play.AffectedPlayer != game.PlayerYou ||
		play.CastFromZone != zone.Library ||
		!play.TopCardOnly {
		t.Fatalf("play-lands effect = %#v, want play-lands-from-library-top", play)
	}
}

// TestLowerCastFromTopWithoutRiderKeepsManaCost proves the plain Future Sight
// wording (no pay-life rider) still lowers to a cast-from-top static that pays
// the spell's ordinary mana cost.
func TestLowerCastFromTopWithoutRiderKeepsManaCost(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Future Sight",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "You may play lands and cast spells from the top of your library.",
	})
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %d, want 1", len(face.StaticAbilities))
	}
	effects := face.StaticAbilities[0].Body.RuleEffects
	if len(effects) != 2 {
		t.Fatalf("rule effects = %#v, want two", effects)
	}
	if effects[0].Kind != game.RuleEffectCastSpellsFromZone || effects[0].PayLifeEqualToManaValue {
		t.Fatalf("cast effect = %#v, want cast-from-top without pay-life", effects[0])
	}
}
