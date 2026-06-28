package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// grandAbolisherPermanent gives controller a battlefield permanent whose static
// ability stops opponents from casting spells and activating abilities of
// artifacts, creatures, and enchantments during the controller's turn.
func grandAbolisherPermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:  "Grand Abolisher",
		Types: []types.Card{types.Creature},
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{
				{
					Kind:                           game.RuleEffectCantCastSpells,
					AffectedPlayer:                 game.PlayerOpponent,
					RestrictedDuringControllerTurn: true,
				},
				{
					Kind:                           game.RuleEffectCantActivateAbilities,
					AffectedPlayer:                 game.PlayerOpponent,
					PermanentTypes:                 []types.Card{types.Artifact, types.Creature, types.Enchantment},
					RestrictedDuringControllerTurn: true,
				},
			},
		}},
	}})
}

func TestSpellCastProhibitedByGrandAbolisher(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	grandAbolisherPermanent(g, game.Player1)
	spell := &game.CardDef{CardFace: game.CardFace{Name: "Test Bolt", Types: []types.Card{types.Instant}}}

	g.Turn.ActivePlayer = game.Player1
	if !spellCastProhibited(g, game.Player2, spell) {
		t.Fatal("opponent should not be able to cast during the controller's turn")
	}
	if spellCastProhibited(g, game.Player1, spell) {
		t.Fatal("the controller is never restricted by their own Grand Abolisher")
	}

	g.Turn.ActivePlayer = game.Player2
	if spellCastProhibited(g, game.Player2, spell) {
		t.Fatal("the turn-scoped restriction must lift on the opponent's turn")
	}
}

func TestAbilityActivationProhibitedByGrandAbolisher(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	grandAbolisherPermanent(g, game.Player1)
	g.Turn.ActivePlayer = game.Player1

	opponentCreature := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Opponent Creature", Types: []types.Card{types.Creature}}})
	if !abilityActivationProhibited(g, game.Player2, opponentCreature, false) {
		t.Fatal("opponent creature abilities should be restricted during the controller's turn")
	}

	opponentLand := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Opponent Land", Types: []types.Card{types.Land}}})
	if abilityActivationProhibited(g, game.Player2, opponentLand, false) {
		t.Fatal("land abilities are not in the restricted permanent-type set")
	}

	g.Turn.ActivePlayer = game.Player2
	if abilityActivationProhibited(g, game.Player2, opponentCreature, false) {
		t.Fatal("the turn-scoped restriction must lift on the opponent's turn")
	}
}

func TestActionRestrictionWithoutTurnScopeAlwaysApplies(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Always Abolisher",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:           game.RuleEffectCantCastSpells,
				AffectedPlayer: game.PlayerOpponent,
			}},
		}},
	}})
	spell := &game.CardDef{CardFace: game.CardFace{Name: "Test Bolt", Types: []types.Card{types.Instant}}}

	g.Turn.ActivePlayer = game.Player2
	if !spellCastProhibited(g, game.Player2, spell) {
		t.Fatal("an unscoped restriction must apply on every turn")
	}
}

// drannithPermanent gives controller a battlefield permanent whose static
// ability stops opponents from casting spells out of any non-hand zone, mirroring
// Drannith Magistrate.
func drannithPermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:  "Drannith Magistrate",
		Types: []types.Card{types.Creature},
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:              game.RuleEffectCantCastFromZones,
				AffectedPlayer:    game.PlayerOpponent,
				CantCastFromZones: []zone.Type{zone.Graveyard, zone.Exile, zone.Library, zone.Command},
			}},
		}},
	}})
}

func TestCastFromZoneProhibitedByDrannith(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	drannithPermanent(g, game.Player1)

	if !castFromZoneProhibited(g, game.Player2, zone.Graveyard) {
		t.Fatal("opponent should not be able to cast from their graveyard")
	}
	if !castFromZoneProhibited(g, game.Player2, zone.Command) {
		t.Fatal("opponent should not be able to cast from the command zone")
	}
	if castFromZoneProhibited(g, game.Player2, zone.Hand) {
		t.Fatal("opponent must still be able to cast from their hand")
	}
	if castFromZoneProhibited(g, game.Player1, zone.Graveyard) {
		t.Fatal("the controller is never restricted by their own Drannith Magistrate")
	}
}
