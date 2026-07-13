package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// throneManaSource builds a permanent carrying Throne of Eldraine's mana
// ability: "{T}: Add four mana of the chosen color. Spend this mana only to cast
// monocolored spells of that color." Its entry choice is fixed to chosen.
func throneManaSource(g *game.Game, controller game.PlayerID, chosen mana.Color) *game.Permanent {
	source := addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:       "Throne of Eldraine",
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Artifact},
		ManaAbilities: []game.ManaAbility{{
			AdditionalCosts: cost.Tap,
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.AddMana{
					Amount:          game.Fixed(4),
					EntryChoiceFrom: game.EntryColorChoiceKey,
					SpendRider: opt.Val(game.ManaSpendRider{
						Condition:   game.ManaSpendCastMonocoloredSpellOfChosenColor,
						Restriction: game.ManaSpendRestrictedToCondition,
					}),
				},
			}}}.Ability(),
		}},
	}})
	source.EntryChoices = map[game.ChoiceKey]game.ResolutionChoiceResult{
		game.EntryColorChoiceKey: {Kind: game.ResolutionChoiceMana, Color: chosen},
	}
	return source
}

// activateThroneForFourMana activates the given Throne's mana ability, producing
// four chosen-color mana tagged with the monocolored-chosen-color spend rider.
func activateThroneForFourMana(t *testing.T, g *game.Game, throne *game.Permanent) {
	t.Helper()
	engine := NewEngine(nil)
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(throne.ObjectID, 0, nil, 0)) {
		t.Fatal("activating Throne mana ability = false, want true")
	}
}

// spellDefWithCost returns a castable instant of the given colors and mana cost
// symbols, usable as a cast target for the monocolored-chosen-color spend tests.
func spellDefWithCost(name string, manaCost cost.Mana, colors []color.Color) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		Types:    []types.Card{types.Instant},
		Colors:   colors,
		ManaCost: opt.Val(manaCost),
	}}
}

// genericSpellDef returns an instant whose cost is a single generic pip of the
// given value, payable by mana of any color.
func genericSpellDef(name string, value int, colors ...color.Color) *game.CardDef {
	return spellDefWithCost(name, cost.Mana{cost.O(value)}, colors)
}

// gameWithChosenColorRestrictedMana seeds Player1 with count mana of the chosen
// color, each tagged with Throne's monocolored-chosen-color spend rider.
func gameWithChosenColorRestrictedMana(chosen mana.Color, count int) (*game.Game, *game.Player) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	player := g.Players[game.Player1]
	player.ManaPool.Add(chosen, count)
	for range count {
		player.ManaRiders = append(player.ManaRiders, game.ManaRiderInstance{
			Unit:       mana.Unit{Color: chosen},
			Controller: game.Player1,
			Rider: game.ManaSpendRider{
				Condition:   game.ManaSpendCastMonocoloredSpellOfChosenColor,
				Restriction: game.ManaSpendRestrictedToCondition,
			},
		})
	}
	return g, player
}

// TestThroneAddsFourChosenColorManaWithMonocoloredRider proves Throne's mana
// ability: activating it adds four mana of the entry-time chosen color (here
// blue) and tags four spend riders restricting that mana to monocolored spells
// of the chosen color.
func TestThroneAddsFourChosenColorManaWithMonocoloredRider(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	throne := throneManaSource(g, game.Player1, mana.U)
	activateThroneForFourMana(t, g, throne)

	player := g.Players[game.Player1]
	if got := player.ManaPool.Amount(mana.U); got != 4 {
		t.Fatalf("blue mana = %d, want 4 (chosen color)", got)
	}
	if got := len(player.ManaRiders); got != 4 {
		t.Fatalf("tagged spend riders = %d, want 4", got)
	}
	for _, rider := range player.ManaRiders {
		if rider.Unit.Color != mana.U {
			t.Fatalf("rider unit color = %q, want blue (chosen color)", rider.Unit.Color)
		}
		if rider.Rider.Condition != game.ManaSpendCastMonocoloredSpellOfChosenColor {
			t.Fatalf("rider condition = %v, want ManaSpendCastMonocoloredSpellOfChosenColor", rider.Rider.Condition)
		}
		if rider.Rider.Restriction != game.ManaSpendRestrictedToCondition {
			t.Fatalf("rider restriction = %v, want ManaSpendRestrictedToCondition", rider.Rider.Restriction)
		}
	}
}

// TestThroneChosenColorManaCastsMatchingMonocoloredSpell proves the end-to-end
// spend through the real payment planner: Throne produces four white mana, and a
// monocolored white spell can spend all four, consuming their spend riders.
func TestThroneChosenColorManaCastsMatchingMonocoloredSpell(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	throne := throneManaSource(g, game.Player1, mana.W)
	activateThroneForFourMana(t, g, throne)
	player := g.Players[game.Player1]

	if !castWithArtifactMana(t, g, genericSpellDef("White Burst", 4, color.White)) {
		t.Fatal("casting monocolored white spell with chosen-color mana = false, want true")
	}
	if got := player.ManaPool.Amount(mana.W); got != 0 {
		t.Fatalf("white mana = %d, want 0 (all four spent on the matching spell)", got)
	}
	if got := len(player.ManaRiders); got != 0 {
		t.Fatalf("spend riders remaining = %d, want 0 (all consumed)", got)
	}
}

// TestThroneChosenColorManaRejectsIneligibleSpells proves the restriction
// through the real payment planner: Throne's four white mana cannot pay for a
// monocolored spell of a different color, a multicolored spell (even one that
// includes the chosen color), or a colorless spell. The tagged mana and its
// riders remain untouched after each rejected cast.
func TestThroneChosenColorManaRejectsIneligibleSpells(t *testing.T) {
	cases := []struct {
		name string
		def  *game.CardDef
	}{
		{"different-color monocolored", genericSpellDef("Blue Bolt", 4, color.Blue)},
		{"multicolored including chosen", genericSpellDef("Azorius Charm", 4, color.White, color.Blue)},
		{"colorless", genericSpellDef("Colorless Shard", 4)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			throne := throneManaSource(g, game.Player1, mana.W)
			activateThroneForFourMana(t, g, throne)
			player := g.Players[game.Player1]

			if castWithArtifactMana(t, g, tc.def) {
				t.Fatalf("casting %s spell with chosen-color mana = true, want false", tc.name)
			}
			if got := player.ManaPool.Amount(mana.W); got != 4 {
				t.Fatalf("white mana = %d, want 4 (restricted mana unspent after rejected cast)", got)
			}
			if got := len(player.ManaRiders); got != 4 {
				t.Fatalf("spend riders = %d, want 4 (unchanged after rejected cast)", got)
			}
		})
	}
}

// TestThroneChosenColorManaCombinesWithUnrestrictedMana proves the interaction
// with ordinary mana through the real payment planner: unrestricted mana pays for
// an ineligible spell while the restricted white mana stays unspent, and an
// eligible monocolored white spell can combine the restricted white mana with
// unrestricted mana to pay a larger cost.
func TestThroneChosenColorManaCombinesWithUnrestrictedMana(t *testing.T) {
	t.Run("unrestricted pays ineligible spell, restricted stays", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		throne := throneManaSource(g, game.Player1, mana.W)
		activateThroneForFourMana(t, g, throne)
		player := g.Players[game.Player1]
		player.ManaPool.Add(mana.C, 2)

		if !castWithArtifactMana(t, g, genericSpellDef("Colorless Shard", 2)) {
			t.Fatal("casting colorless spell with unrestricted mana = false, want true")
		}
		if got := player.ManaPool.Amount(mana.C); got != 0 {
			t.Fatalf("colorless mana = %d, want 0 (spent on the ineligible spell)", got)
		}
		if got := player.ManaPool.Amount(mana.W); got != 4 {
			t.Fatalf("white mana = %d, want 4 (restricted mana ineligible, left unspent)", got)
		}
		if got := len(player.ManaRiders); got != 4 {
			t.Fatalf("spend riders = %d, want 4 (unchanged)", got)
		}
	})

	t.Run("restricted white combines with unrestricted for a matching spell", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		throne := throneManaSource(g, game.Player1, mana.W)
		activateThroneForFourMana(t, g, throne)
		player := g.Players[game.Player1]
		player.ManaPool.Add(mana.C, 2)

		if !castWithArtifactMana(t, g, genericSpellDef("Grand White Rite", 6, color.White)) {
			t.Fatal("casting monocolored white {6} spell = false, want true")
		}
		if got := player.ManaPool.Amount(mana.W); got != 0 {
			t.Fatalf("white mana = %d, want 0 (four restricted white spent)", got)
		}
		if got := player.ManaPool.Amount(mana.C); got != 0 {
			t.Fatalf("colorless mana = %d, want 0 (two unrestricted combined into the cost)", got)
		}
		if got := len(player.ManaRiders); got != 0 {
			t.Fatalf("spend riders = %d, want 0 (all four consumed)", got)
		}
	})
}

// TestThroneChosenColorManaHybridColorIdentity proves the hybrid color-identity
// rule (CR 202.2f) through the real payment planner. A card's color is fixed by
// the color symbols in its cost regardless of how it is paid, matching Scryfall:
// a two-color hybrid ({G/W}) is multicolored and cannot spend chosen-color mana
// even though white mana can physically pay it, while a monocolored hybrid
// ({2/W}) is a single white color and can.
func TestThroneChosenColorManaHybridColorIdentity(t *testing.T) {
	t.Run("two-color hybrid is multicolored and rejected", func(t *testing.T) {
		g, player := gameWithChosenColorRestrictedMana(mana.W, 1)
		// Kitchen Finks-style {G/W}: white mana can pay the pip, but the spell is
		// green and white, so it is multicolored and not a monocolored white spell.
		twoColorHybrid := spellDefWithCost(
			"Gilded Finks",
			cost.Mana{cost.HybridMana(mana.G, mana.W)},
			[]color.Color{color.Green, color.White},
		)

		if castWithArtifactMana(t, g, twoColorHybrid) {
			t.Fatal("casting two-color hybrid with chosen-color mana = true, want false (multicolored)")
		}
		if got := player.ManaPool.Amount(mana.W); got != 1 {
			t.Fatalf("white mana = %d, want 1 (restricted mana unspent on a multicolored spell)", got)
		}
		if got := len(player.ManaRiders); got != 1 {
			t.Fatalf("spend riders = %d, want 1 (unchanged)", got)
		}
	})

	t.Run("monocolored hybrid of the chosen color is accepted", func(t *testing.T) {
		g, player := gameWithChosenColorRestrictedMana(mana.W, 1)
		// Spectral Procession-style {2/W}: a monocolored white spell, payable by a
		// single white mana via the {W} half.
		monoHybrid := spellDefWithCost(
			"Spectral Rite",
			cost.Mana{cost.Twobrid(mana.W)},
			[]color.Color{color.White},
		)

		if !castWithArtifactMana(t, g, monoHybrid) {
			t.Fatal("casting monocolored white hybrid with chosen-color mana = false, want true")
		}
		if got := player.ManaPool.Amount(mana.W); got != 0 {
			t.Fatalf("white mana = %d, want 0 (spent on the monocolored white spell)", got)
		}
		if got := len(player.ManaRiders); got != 0 {
			t.Fatalf("spend riders = %d, want 0 (consumed)", got)
		}
	})
}
