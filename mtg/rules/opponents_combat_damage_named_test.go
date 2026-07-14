package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// gollumNamedGroupGame builds a fresh game with a Gollum permanent (the named
// source) and a differently named permanent, both controlled by Player1, plus
// the combat-damage event history the group recipient reads. It returns the game
// and the Gollum permanent.
func gollumNamedGroupGame(t *testing.T) (*game.Game, *game.Permanent) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	gollum := addPermanentForSBA(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Gollum, Obsessed Stalker",
	}})
	other := addPermanentForSBA(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Bilbo, Retired Burglar",
	}})
	victimPermanent := addPermanentForSBA(g, game.Player3, &game.CardDef{CardFace: game.CardFace{
		Name: "Grond",
	}})

	g.Events = append(g.Events,
		// Player2 dealt combat damage by Gollum this game: qualifies.
		game.Event{
			Kind:            game.EventDamageDealt,
			SourceID:        gollum.CardInstanceID,
			SourceObjectID:  gollum.ObjectID,
			Controller:      game.Player1,
			Player:          game.Player2,
			Amount:          2,
			DamageRecipient: game.DamageRecipientPlayer,
			CombatDamage:    true,
		},
		// Player3 dealt combat damage by a differently named creature: does not qualify.
		game.Event{
			Kind:            game.EventDamageDealt,
			SourceID:        other.CardInstanceID,
			SourceObjectID:  other.ObjectID,
			Controller:      game.Player1,
			Player:          game.Player3,
			Amount:          2,
			DamageRecipient: game.DamageRecipientPlayer,
			CombatDamage:    true,
		},
		// Combat damage dealt to a permanent by Gollum: does not qualify (its
		// controller must not be pulled in).
		game.Event{
			Kind:            game.EventDamageDealt,
			SourceID:        gollum.CardInstanceID,
			SourceObjectID:  gollum.ObjectID,
			Controller:      game.Player1,
			Player:          game.Player3,
			PermanentID:     victimPermanent.ObjectID,
			Amount:          1,
			DamageRecipient: game.DamageRecipientPermanent,
			CombatDamage:    true,
		},
		// Non-combat damage to Player4 by Gollum: does not qualify.
		game.Event{
			Kind:            game.EventDamageDealt,
			SourceID:        gollum.CardInstanceID,
			SourceObjectID:  gollum.ObjectID,
			Controller:      game.Player1,
			Player:          game.Player4,
			Amount:          3,
			DamageRecipient: game.DamageRecipientPlayer,
			CombatDamage:    false,
		},
	)
	return g, gollum
}

// TestOpponentsDealtCombatDamageThisGameByNamedResolution proves the group
// resolver returns exactly the opponents dealt combat damage this game by a
// creature with the given name: an opponent damaged by a differently named
// creature, combat damage dealt to a permanent, and non-combat damage must not
// qualify.
func TestOpponentsDealtCombatDamageThisGameByNamedResolution(t *testing.T) {
	g, gollum := gollumNamedGroupGame(t)

	resolver := newReferenceResolverWithControllerAndSource(g, game.Player1, gollum)
	members := resolver.playerGroup(
		game.OpponentsDealtCombatDamageThisGameByNamedReference("Gollum, Obsessed Stalker"),
	)

	if len(members) != 1 || members[0] != game.Player2 {
		t.Fatalf("members = %v, want [Player2] only", members)
	}
}

func TestOpponentsDealtCombatDamageThisGameUsesEventTimeName(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	gollum := addPermanentForSBA(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Gollum, Obsessed Stalker",
	}})
	emitEvent(g, game.Event{
		Kind:            game.EventDamageDealt,
		SourceID:        gollum.CardInstanceID,
		SourceObjectID:  gollum.ObjectID,
		Controller:      game.Player1,
		Player:          game.Player2,
		Amount:          2,
		DamageRecipient: game.DamageRecipientPlayer,
		CombatDamage:    true,
	})
	g.CardInstances[gollum.CardInstanceID].Def.Name = "Smeagol"

	resolver := newReferenceResolverWithControllerAndSource(g, game.Player1, gollum)
	members := resolver.playerGroup(
		game.OpponentsDealtCombatDamageThisGameByNamedReference("Gollum, Obsessed Stalker"),
	)
	if len(members) != 1 || members[0] != game.Player2 {
		t.Fatalf("members = %v, want [Player2] from the event-time source name", members)
	}
}

// TestOpponentsDealtCombatDamageThisGameByNamedLifeLoss drives the runtime
// LoseLife handler: only the qualifying opponent loses life equal to the
// controller's life gained this turn.
func TestOpponentsDealtCombatDamageThisGameByNamedLifeLoss(t *testing.T) {
	g, gollum := gollumNamedGroupGame(t)

	// The controller gained 5 life this turn.
	g.Events = append(g.Events, game.Event{
		Kind:   game.EventLifeGained,
		Player: game.Player1,
		Amount: 5,
	})

	startingLife := make([]int, game.NumPlayers)
	for i := range g.Players {
		startingLife[i] = g.Players[i].Life
	}

	obj := &game.StackObject{Controller: game.Player1, SourceID: gollum.CardInstanceID}
	r := &effectResolver{engine: NewEngine(nil), game: g, obj: obj, log: &TurnLog{}}

	res := handleLoseLife(r, game.LoseLife{
		Amount:      game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountLifeGainedThisTurn}),
		PlayerGroup: game.OpponentsDealtCombatDamageThisGameByNamedReference("Gollum, Obsessed Stalker"),
	})
	if !res.accepted {
		t.Fatal("lose-life not accepted")
	}

	if got := g.Players[game.Player2].Life; got != startingLife[game.Player2]-5 {
		t.Fatalf("Player2 life = %d, want %d (lost 5)", got, startingLife[game.Player2]-5)
	}
	for _, pid := range []game.PlayerID{game.Player1, game.Player3, game.Player4} {
		if got := g.Players[pid].Life; got != startingLife[pid] {
			t.Fatalf("Player%d life = %d, want unchanged %d", pid+1, got, startingLife[pid])
		}
	}
}
