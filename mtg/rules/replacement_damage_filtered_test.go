package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
)

func filteredDamageReplacementCardDef(spec *game.DamageReplacementSpec) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Mayhem Dominus",
		Types: []types.Card{types.Enchantment},
		ReplacementAbilities: []game.ReplacementAbility{
			game.DamageReplacementFiltered("filtered damage replacement", spec),
		},
	}}
}

func TestDamageReplacementOpponentRecipientOnly(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, filteredDamageReplacementCardDef(&game.DamageReplacementSpec{
		Multiplier:        2,
		RecipientOpponent: true,
		Controller:        game.TriggerControllerYou,
	}))
	sourceID := addColoredSourceCard(g, game.Player1, color.Red)

	if dealt := dealPlayerDamage(g, sourceID, 0, game.Player1, game.Player2, 3, false); dealt != 6 {
		t.Fatalf("damage to opponent = %d, want 6", dealt)
	}
	if dealt := dealPlayerDamage(g, sourceID, 0, game.Player1, game.Player1, 3, false); dealt != 3 {
		t.Fatalf("damage to controller = %d, want 3", dealt)
	}
}

func TestDamageReplacementNoncombatOnly(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, filteredDamageReplacementCardDef(&game.DamageReplacementSpec{
		Multiplier:    2,
		NoncombatOnly: true,
		Controller:    game.TriggerControllerYou,
	}))
	sourceID := addColoredSourceCard(g, game.Player1, color.Red)

	if dealt := dealPlayerDamage(g, sourceID, 0, game.Player1, game.Player2, 3, false); dealt != 6 {
		t.Fatalf("noncombat damage = %d, want 6", dealt)
	}
	if dealt := dealPlayerDamage(g, sourceID, 0, game.Player1, game.Player2, 3, true); dealt != 3 {
		t.Fatalf("combat damage = %d, want 3", dealt)
	}
}

func TestDamageReplacementCreatureSourceOnly(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, filteredDamageReplacementCardDef(&game.DamageReplacementSpec{
		Multiplier:  2,
		SourceTypes: []types.Card{types.Creature},
		Controller:  game.TriggerControllerYou,
	}))
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	noncreatureID := addColoredSourceCard(g, game.Player1, color.Red)

	if dealt := dealPlayerDamage(g, creature.CardInstanceID, creature.ObjectID, game.Player1, game.Player2, 3, false); dealt != 6 {
		t.Fatalf("creature-source damage = %d, want 6", dealt)
	}
	if dealt := dealPlayerDamage(g, noncreatureID, 0, game.Player1, game.Player2, 3, false); dealt != 3 {
		t.Fatalf("noncreature-source damage = %d, want 3", dealt)
	}
}

func TestDamageReplacementAnyControllerSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, filteredDamageReplacementCardDef(&game.DamageReplacementSpec{
		Multiplier: 2,
		Controller: game.TriggerControllerAny,
	}))
	ownSource := addColoredSourceCard(g, game.Player1, color.Red)
	opponentSource := addColoredSourceCard(g, game.Player2, color.Red)

	if dealt := dealPlayerDamage(g, ownSource, 0, game.Player1, game.Player2, 3, false); dealt != 6 {
		t.Fatalf("own-source damage = %d, want 6", dealt)
	}
	if dealt := dealPlayerDamage(g, opponentSource, 0, game.Player2, game.Player1, 3, false); dealt != 6 {
		t.Fatalf("opponent-source damage = %d, want 6", dealt)
	}
}

func TestDamageReplacementAddendForOpponentRecipient(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, filteredDamageReplacementCardDef(&game.DamageReplacementSpec{
		Addend:            2,
		SourceColors:      []color.Color{color.Red},
		RecipientOpponent: true,
		Controller:        game.TriggerControllerYou,
	}))
	redSource := addColoredSourceCard(g, game.Player1, color.Red)
	blueSource := addColoredSourceCard(g, game.Player1, color.Blue)

	if dealt := dealPlayerDamage(g, redSource, 0, game.Player1, game.Player2, 3, false); dealt != 5 {
		t.Fatalf("red-source damage to opponent = %d, want 5", dealt)
	}
	if dealt := dealPlayerDamage(g, blueSource, 0, game.Player1, game.Player2, 3, false); dealt != 3 {
		t.Fatalf("blue-source damage = %d, want 3", dealt)
	}
}
