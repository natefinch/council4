package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// oneManaSpellDef returns a castable {1} spell of the given card types and
// colors, usable as a cast target for spell-type mana-spend restriction tests.
func oneManaSpellDef(name string, cardTypes []types.Card, colors ...color.Color) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		Types:    cardTypes,
		Colors:   colors,
		ManaCost: opt.Val(cost.Mana{cost.O(1)}),
	}}
}

// instantSpellDef, sorcerySpellDef, planeswalkerSpellDef, and
// multicoloredSpellDef return castable spells covering each modeled spell-type
// restriction, mirroring Vodalian Arcanist, Nardole, Pillar of the Paruns, and
// Interplanar Beacon targets.
func instantSpellDef() *game.CardDef {
	return oneManaSpellDef("Shock", []types.Card{types.Instant}, color.Red)
}

func sorcerySpellDef() *game.CardDef {
	return oneManaSpellDef("Dig", []types.Card{types.Sorcery}, color.Blue)
}

func planeswalkerSpellDef() *game.CardDef {
	def := oneManaSpellDef("Walker", []types.Card{types.Planeswalker}, color.White)
	def.Loyalty = opt.Val(3)
	return def
}

func multicoloredSpellDef() *game.CardDef {
	return oneManaSpellDef("Hybrid", []types.Card{types.Instant}, color.White, color.Blue)
}

func monocoloredNoncreatureSpellDef() *game.CardDef {
	return oneManaSpellDef("Bolt", []types.Card{types.Instant}, color.Red)
}

// TestInstantOrSorceryManaCastRestriction covers Vodalian Arcanist: the tagged
// mana may pay only to cast an instant or sorcery spell.
func TestInstantOrSorceryManaCastRestriction(t *testing.T) {
	t.Parallel()

	t.Run("accepts instant spell", func(t *testing.T) {
		t.Parallel()
		g, player := gameWithArtifactRestrictedMana(game.ManaSpendCastInstantOrSorcerySpell)
		if !castWithArtifactMana(t, g, instantSpellDef()) {
			t.Fatal("cast instant with instant-or-sorcery mana failed")
		}
		if player.ManaPool.Amount(mana.C) != 0 || len(player.ManaRiders) != 0 {
			t.Fatalf("rider not consumed by instant: pool=%d riders=%d", player.ManaPool.Amount(mana.C), len(player.ManaRiders))
		}
	})

	t.Run("accepts sorcery spell", func(t *testing.T) {
		t.Parallel()
		g, player := gameWithArtifactRestrictedMana(game.ManaSpendCastInstantOrSorcerySpell)
		if !castWithArtifactMana(t, g, sorcerySpellDef()) {
			t.Fatal("cast sorcery with instant-or-sorcery mana failed")
		}
		if player.ManaPool.Amount(mana.C) != 0 || len(player.ManaRiders) != 0 {
			t.Fatalf("rider not consumed by sorcery: pool=%d riders=%d", player.ManaPool.Amount(mana.C), len(player.ManaRiders))
		}
	})

	t.Run("rejects creature spell", func(t *testing.T) {
		t.Parallel()
		g, player := gameWithArtifactRestrictedMana(game.ManaSpendCastInstantOrSorcerySpell)
		if castWithArtifactMana(t, g, nonartifactSpellDef()) {
			t.Fatal("cast creature with instant-or-sorcery mana succeeded")
		}
		if player.ManaPool.Amount(mana.C) != 1 || len(player.ManaRiders) != 1 {
			t.Fatalf("restricted mana changed after rejected cast: pool=%d riders=%d", player.ManaPool.Amount(mana.C), len(player.ManaRiders))
		}
	})
}

// TestNoncreatureManaCastRestriction covers Nardole: the tagged mana may pay
// only to cast a spell that is not a creature.
func TestNoncreatureManaCastRestriction(t *testing.T) {
	t.Parallel()

	t.Run("accepts noncreature spell", func(t *testing.T) {
		t.Parallel()
		g, player := gameWithArtifactRestrictedMana(game.ManaSpendCastNoncreatureSpell)
		if !castWithArtifactMana(t, g, instantSpellDef()) {
			t.Fatal("cast instant with noncreature mana failed")
		}
		if player.ManaPool.Amount(mana.C) != 0 || len(player.ManaRiders) != 0 {
			t.Fatalf("rider not consumed by noncreature spell: pool=%d riders=%d", player.ManaPool.Amount(mana.C), len(player.ManaRiders))
		}
	})

	t.Run("rejects creature spell", func(t *testing.T) {
		t.Parallel()
		g, player := gameWithArtifactRestrictedMana(game.ManaSpendCastNoncreatureSpell)
		if castWithArtifactMana(t, g, nonartifactSpellDef()) {
			t.Fatal("cast creature with noncreature mana succeeded")
		}
		if player.ManaPool.Amount(mana.C) != 1 || len(player.ManaRiders) != 1 {
			t.Fatalf("restricted mana changed after rejected cast: pool=%d riders=%d", player.ManaPool.Amount(mana.C), len(player.ManaRiders))
		}
	})
}

// TestMulticoloredManaCastRestriction covers Pillar of the Paruns: the tagged
// mana may pay only to cast a spell with two or more colors.
func TestMulticoloredManaCastRestriction(t *testing.T) {
	t.Parallel()

	t.Run("accepts multicolored spell", func(t *testing.T) {
		t.Parallel()
		g, player := gameWithArtifactRestrictedMana(game.ManaSpendCastMulticoloredSpell)
		if !castWithArtifactMana(t, g, multicoloredSpellDef()) {
			t.Fatal("cast multicolored spell with multicolored mana failed")
		}
		if player.ManaPool.Amount(mana.C) != 0 || len(player.ManaRiders) != 0 {
			t.Fatalf("rider not consumed by multicolored spell: pool=%d riders=%d", player.ManaPool.Amount(mana.C), len(player.ManaRiders))
		}
	})

	t.Run("rejects monocolored spell", func(t *testing.T) {
		t.Parallel()
		g, player := gameWithArtifactRestrictedMana(game.ManaSpendCastMulticoloredSpell)
		if castWithArtifactMana(t, g, monocoloredNoncreatureSpellDef()) {
			t.Fatal("cast monocolored spell with multicolored mana succeeded")
		}
		if player.ManaPool.Amount(mana.C) != 1 || len(player.ManaRiders) != 1 {
			t.Fatalf("restricted mana changed after rejected cast: pool=%d riders=%d", player.ManaPool.Amount(mana.C), len(player.ManaRiders))
		}
	})
}

// TestPlaneswalkerManaCastRestriction covers the Wizard token / Interplanar
// Beacon: the tagged mana may pay only to cast a planeswalker spell.
func TestPlaneswalkerManaCastRestriction(t *testing.T) {
	t.Parallel()

	t.Run("accepts planeswalker spell", func(t *testing.T) {
		t.Parallel()
		g, player := gameWithArtifactRestrictedMana(game.ManaSpendCastPlaneswalkerSpell)
		if !castWithArtifactMana(t, g, planeswalkerSpellDef()) {
			t.Fatal("cast planeswalker with planeswalker mana failed")
		}
		if player.ManaPool.Amount(mana.C) != 0 || len(player.ManaRiders) != 0 {
			t.Fatalf("rider not consumed by planeswalker: pool=%d riders=%d", player.ManaPool.Amount(mana.C), len(player.ManaRiders))
		}
	})

	t.Run("rejects instant spell", func(t *testing.T) {
		t.Parallel()
		g, player := gameWithArtifactRestrictedMana(game.ManaSpendCastPlaneswalkerSpell)
		if castWithArtifactMana(t, g, instantSpellDef()) {
			t.Fatal("cast instant with planeswalker mana succeeded")
		}
		if player.ManaPool.Amount(mana.C) != 1 || len(player.ManaRiders) != 1 {
			t.Fatalf("restricted mana changed after rejected cast: pool=%d riders=%d", player.ManaPool.Amount(mana.C), len(player.ManaRiders))
		}
	})
}
