package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerCastAsThoughFlashOneShot proves that "You may cast spells this turn as
// though they had flash." lowers to an ApplyRule carrying the controller-scoped
// RuleEffectCastSpellsAsThoughFlash for the turn (Borne Upon a Wind).
func TestLowerCastAsThoughFlashOneShot(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Borne Upon a Wind",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "You may cast spells this turn as though they had flash.\nDraw a card.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("expected a spell ability")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want two primitives", mode.Sequence)
	}
	apply, ok := mode.Sequence[0].Primitive.(game.ApplyRule)
	if !ok {
		t.Fatalf("first primitive = %T, want game.ApplyRule", mode.Sequence[0].Primitive)
	}
	if apply.Duration != game.DurationThisTurn {
		t.Fatalf("duration = %v, want DurationThisTurn", apply.Duration)
	}
	if len(apply.RuleEffects) != 1 {
		t.Fatalf("rule effects = %#v, want one", apply.RuleEffects)
	}
	effect := apply.RuleEffects[0]
	if effect.Kind != game.RuleEffectCastSpellsAsThoughFlash {
		t.Fatalf("kind = %v, want RuleEffectCastSpellsAsThoughFlash", effect.Kind)
	}
	if effect.AffectedPlayer != game.PlayerYou {
		t.Fatalf("affected player = %v, want PlayerYou", effect.AffectedPlayer)
	}
	if _, ok := mode.Sequence[1].Primitive.(game.Draw); !ok {
		t.Fatalf("second primitive = %T, want game.Draw", mode.Sequence[1].Primitive)
	}
}

// TestLowerCastAsThoughFlashInActivatedAbility proves the same permission lowers
// inside an activated ability body (Emergence Zone).
func TestLowerCastAsThoughFlashInActivatedAbility(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Emergence Zone",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "{T}: Add {C}.\n{1}, {T}, Sacrifice this land: You may cast spells this turn as though they had flash.",
	})
	var found bool
	for _, ability := range face.ActivatedAbilities {
		for _, ins := range ability.Content.Modes[0].Sequence {
			apply, ok := ins.Primitive.(game.ApplyRule)
			if ok && len(apply.RuleEffects) == 1 &&
				apply.RuleEffects[0].Kind == game.RuleEffectCastSpellsAsThoughFlash {
				found = true
			}
		}
	}
	if !found {
		t.Fatal("expected an activated ability granting RuleEffectCastSpellsAsThoughFlash")
	}
}

// flashGrantFromActivatedAbilities returns the single RuleEffectCastSpellsAsThoughFlash
// carried by an ApplyRule in one of face's activated abilities, requiring exactly
// one such grant.
func flashGrantFromActivatedAbilities(t *testing.T, face loweredFaceAbilities) game.RuleEffect {
	t.Helper()
	var grants []game.RuleEffect
	for _, ability := range face.ActivatedAbilities {
		for _, ins := range ability.Content.Modes[0].Sequence {
			apply, ok := ins.Primitive.(game.ApplyRule)
			if !ok {
				continue
			}
			for _, effect := range apply.RuleEffects {
				if effect.Kind == game.RuleEffectCastSpellsAsThoughFlash {
					grants = append(grants, effect)
				}
			}
		}
	}
	if len(grants) != 1 {
		t.Fatalf("cast-as-though-flash grants = %d, want 1", len(grants))
	}
	return grants[0]
}

// TestLowerCastAsThoughFlashActivatedManaTapCost proves the activated,
// mana-and-tap-cost this-turn permission lowers fully (Alchemist's Refuge:
// "{G}{U}, {T}: You may cast spells this turn as though they had flash."). Unlike
// Emergence Zone's "Sacrifice this land" cost, this ability has no source
// antecedent in its body, so the internal "they" pronoun must be consumed by the
// exact effect rather than rejected as an unbound activation reference.
func TestLowerCastAsThoughFlashActivatedManaTapCost(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Alchemist's Refuge",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "{T}: Add {C}.\n{G}{U}, {T}: You may cast spells this turn as though they had flash.",
	})
	grant := flashGrantFromActivatedAbilities(t, face)
	if grant.AffectedPlayer != game.PlayerYou {
		t.Fatalf("affected player = %v, want PlayerYou", grant.AffectedPlayer)
	}
	if len(grant.SpellTypes) != 0 || len(grant.SpellSubtypes) != 0 {
		t.Fatalf("unfiltered grant carried filters: types=%#v subtypes=%#v", grant.SpellTypes, grant.SpellSubtypes)
	}
	if len(face.ManaAbilities) != 1 {
		t.Fatalf("mana abilities = %d, want 1", len(face.ManaAbilities))
	}
}

// TestLowerCastAsThoughFlashActivatedTypeFilter proves the activated card-type
// filtered form lowers with the matching SpellTypes narrowing (Winding Canyons:
// "{2}, {T}: You may cast creature spells this turn as though they had flash.").
func TestLowerCastAsThoughFlashActivatedTypeFilter(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Winding Canyons",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "{T}: Add {C}.\n{2}, {T}: You may cast creature spells this turn as though they had flash.",
	})
	grant := flashGrantFromActivatedAbilities(t, face)
	if !slices.Equal(grant.SpellTypes, []types.Card{types.Creature}) {
		t.Fatalf("spell types = %#v, want [Creature]", grant.SpellTypes)
	}
	if len(grant.SpellSubtypes) != 0 {
		t.Fatalf("spell subtypes = %#v, want none", grant.SpellSubtypes)
	}
}

// TestLowerCastAsThoughFlashOneShotSubtypeFilter proves the one-shot subtype
// filtered form carries the matching SpellSubtypes narrowing ("You may cast Aura
// and Equipment spells this turn as though they had flash.").
func TestLowerCastAsThoughFlashOneShotSubtypeFilter(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Subtype Flash",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "You may cast Aura and Equipment spells this turn as though they had flash.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("expected a spell ability")
	}
	apply, ok := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.ApplyRule)
	if !ok {
		t.Fatalf("first primitive = %T, want game.ApplyRule", face.SpellAbility.Val.Modes[0].Sequence[0].Primitive)
	}
	if len(apply.RuleEffects) != 1 {
		t.Fatalf("rule effects = %#v, want one", apply.RuleEffects)
	}
	effect := apply.RuleEffects[0]
	if len(effect.SpellTypes) != 0 {
		t.Fatalf("spell types = %#v, want none", effect.SpellTypes)
	}
	if !slices.Equal(effect.SpellSubtypes, []types.Sub{types.Sub("Aura"), types.Sub("Equipment")}) {
		t.Fatalf("spell subtypes = %#v, want [Aura Equipment]", effect.SpellSubtypes)
	}
}
