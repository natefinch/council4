package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// flashTimingStaticPermanent gives playerID a battlefield permanent whose static
// ability lets them cast spells as though they had flash. The optional
// spellTypes/spellSubtypes filters narrow the grant the same way the lowered
// static does.
func flashTimingStaticPermanent(g *game.Game, playerID game.PlayerID, spellTypes []types.Card, spellSubtypes []types.Sub) *game.Permanent {
	return addCombatPermanent(g, playerID, &game.CardDef{CardFace: game.CardFace{
		Name: "Test Flash Permission",
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:           game.RuleEffectCastSpellsAsThoughFlash,
				AffectedPlayer: game.PlayerYou,
				SpellTypes:     spellTypes,
				SpellSubtypes:  spellSubtypes,
			}},
		}},
	}})
}

func TestPlayerCanCastAsThoughFlashHonorsRuleEffect(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	sorcery := &game.CardDef{CardFace: game.CardFace{Name: "Test Sorcery", Types: []types.Card{types.Sorcery}}}
	if playerCanCastAsThoughFlash(g, game.Player1, sorcery) {
		t.Fatal("player has flash timing without any rule effect")
	}
	flashTimingStaticPermanent(g, game.Player1, nil, nil)
	if !playerCanCastAsThoughFlash(g, game.Player1, sorcery) {
		t.Fatal("player lacks flash timing despite an active rule effect")
	}
	if playerCanCastAsThoughFlash(g, game.Player2, sorcery) {
		t.Fatal("opponent gained flash timing from the controller's rule effect")
	}
}

// TestPlayerCanCastAsThoughFlashHonorsFilters proves the optional card-type and
// subtype filters narrow the grant to matching spells only.
func TestPlayerCanCastAsThoughFlashHonorsFilters(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	sorcery := &game.CardDef{CardFace: game.CardFace{Name: "Test Sorcery", Types: []types.Card{types.Sorcery}}}
	creature := &game.CardDef{CardFace: game.CardFace{Name: "Test Creature", Types: []types.Card{types.Creature}}}
	aura := &game.CardDef{CardFace: game.CardFace{Name: "Test Aura", Types: []types.Card{types.Enchantment}, Subtypes: []types.Sub{types.Sub("Aura")}}}

	flashTimingStaticPermanent(g, game.Player1, []types.Card{types.Sorcery}, nil)
	if !playerCanCastAsThoughFlash(g, game.Player1, sorcery) {
		t.Fatal("sorcery-filtered permission did not match a sorcery spell")
	}
	if playerCanCastAsThoughFlash(g, game.Player1, creature) {
		t.Fatal("sorcery-filtered permission matched a creature spell")
	}

	g2 := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	flashTimingStaticPermanent(g2, game.Player1, nil, []types.Sub{types.Sub("Aura"), types.Sub("Equipment")})
	if !playerCanCastAsThoughFlash(g2, game.Player1, aura) {
		t.Fatal("subtype-filtered permission did not match an Aura spell")
	}
	if playerCanCastAsThoughFlash(g2, game.Player1, creature) {
		t.Fatal("subtype-filtered permission matched a non-Aura/Equipment spell")
	}
}

// TestCanCastAtCurrentTimingAllowsSorceryWithFlashPermission proves the timing
// permission lets a sorcery-speed card be cast at instant speed, while leaving
// it sorcery-speed for a player without the permission.
func TestCanCastAtCurrentTimingAllowsSorceryWithFlashPermission(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	// Make it not the active player's main phase with an empty stack, so sorcery
	// speed is unavailable to anyone by default.
	g.Turn.ActivePlayer = game.Player2
	sorcery := &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Sorcery",
		Types: []types.Card{types.Sorcery},
	}}
	if canCastAtCurrentTiming(g, game.Player1, sorcery) {
		t.Fatal("non-active player may cast a sorcery without flash timing")
	}
	flashTimingStaticPermanent(g, game.Player1, nil, nil)
	if !canCastAtCurrentTiming(g, game.Player1, sorcery) {
		t.Fatal("flash timing permission did not allow casting a sorcery at instant speed")
	}
}
