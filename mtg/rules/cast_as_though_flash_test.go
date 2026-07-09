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

// TestApplyRuleGrantsSorceryFlashTimingForTurn proves that resolving the
// activated ability's lowered ApplyRule (Alchemist's Refuge:
// "{G}{U}, {T}: You may cast spells this turn as though they had flash.") lets the
// controller cast a sorcery-speed spell at instant speed for the rest of the turn,
// while an unfiltered creature filter would not have matched. The permission is
// created by resolving the primitive rather than a static ability, exercising the
// handleApplyRule → createRuleEffectTemplates path.
func TestApplyRuleGrantsSorceryFlashTimingForTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	// Make it not the controller's main phase with an empty stack, so sorcery
	// speed is otherwise unavailable to them.
	g.Turn.ActivePlayer = game.Player2
	sorcery := &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Sorcery",
		Types: []types.Card{types.Sorcery},
	}}
	if canCastAtCurrentTiming(g, game.Player1, sorcery) {
		t.Fatal("sorcery castable at instant speed before activating the ability")
	}

	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Alchemist's Refuge",
		Types: []types.Card{types.Land},
	}})
	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackActivatedAbility,
		Controller: game.Player1,
		SourceID:   source.ObjectID,
	}
	r := &effectResolver{engine: NewEngine(nil), game: g, obj: obj, log: &TurnLog{}}

	resolved := handleApplyRule(r, game.ApplyRule{
		RuleEffects: []game.RuleEffect{{
			Kind:           game.RuleEffectCastSpellsAsThoughFlash,
			AffectedPlayer: game.PlayerYou,
		}},
		Duration: game.DurationThisTurn,
	})
	if !resolved.succeeded {
		t.Fatal("handleApplyRule did not create the flash-timing rule effect")
	}

	if !canCastAtCurrentTiming(g, game.Player1, sorcery) {
		t.Fatal("activating the ability did not allow casting a sorcery at instant speed")
	}
	if canCastAtCurrentTiming(g, game.Player2, sorcery) {
		t.Fatal("the controller's activated permission leaked to an opponent")
	}
}
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
