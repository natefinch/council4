package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestImpulseExileCastFromEventPlayerLibraryGrantsCastOnlyAnyMana proves that an
// impulse exile scoped to the triggering event player's library with the
// cast-only and any-color riders (Grenzo, Havoc Raiser's "Exile the top card of
// that player's library. Until end of turn, you may cast that card and you may
// spend mana as though it were mana of any color to cast that spell.") exiles the
// top card of the event player's library, then grants the ability controller a
// cast-only permission — never a land-play permission — that may be paid with
// mana of any color.
func TestImpulseExileCastFromEventPlayerLibraryGrantsCastOnlyAnyMana(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	// Player2 is the damaged (event) player whose library is exiled.
	topID := addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Event Player Spell",
		Types: []types.Card{types.Sorcery},
	}})

	obj := &game.StackObject{
		Controller:      game.Player1,
		SourceCardID:    g.IDGen.Next(),
		SourceID:        g.IDGen.Next(),
		HasTriggerEvent: true,
		TriggerEvent: game.Event{
			Kind:            game.EventDamageDealt,
			Player:          game.Player2,
			DamageRecipient: game.DamageRecipientPlayer,
		},
	}
	resolveInstruction(engine, g, obj, game.ImpulseExile{
		Player:       game.EventPlayerReference(),
		Amount:       game.Fixed(1),
		Duration:     game.DurationUntilEndOfTurn,
		Cast:         true,
		SpendAnyMana: true,
	}, &TurnLog{})

	if !g.Players[game.Player2].Exile.Contains(topID) {
		t.Fatal("the event player's top library card was not exiled")
	}
	if g.Players[game.Player2].Library.Contains(topID) {
		t.Fatal("the exiled card is still in the event player's library")
	}
	if !hasCastFromZoneRuleEffect(g, game.Player1, topID, zone.Exile, game.FaceFront) {
		t.Fatal("the controller was not granted permission to cast the exiled card")
	}
	if canPlayLandFromZoneByRuleEffect(g, game.Player1, topID, zone.Exile) {
		t.Fatal("cast-only impulse granted land-play permission, but should grant cast only")
	}
	if !castFromZoneAllowsAnyMana(g, game.Player1, topID, zone.Exile, game.FaceFront) {
		t.Fatal("the cast-only permission did not honor the spend-any-color rider")
	}
}

// TestImpulseExilePlayFromEventPlayerLibraryGrantsLandPlay proves the play
// variant (no Cast rider) still grants an ordinary play-from-zone permission that
// includes playing an exiled land, distinguishing it from the cast-only grant.
func TestImpulseExilePlayFromEventPlayerLibraryGrantsLandPlay(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	topID := addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Event Player Land",
		Types: []types.Card{types.Land},
	}})

	obj := &game.StackObject{
		Controller:      game.Player1,
		SourceCardID:    g.IDGen.Next(),
		SourceID:        g.IDGen.Next(),
		HasTriggerEvent: true,
		TriggerEvent: game.Event{
			Kind:            game.EventDamageDealt,
			Player:          game.Player2,
			DamageRecipient: game.DamageRecipientPlayer,
		},
	}
	resolveInstruction(engine, g, obj, game.ImpulseExile{
		Player:   game.EventPlayerReference(),
		Amount:   game.Fixed(1),
		Duration: game.DurationUntilEndOfTurn,
	}, &TurnLog{})

	if !g.Players[game.Player2].Exile.Contains(topID) {
		t.Fatal("the event player's top library card was not exiled")
	}
	if !canPlayLandFromZoneByRuleEffect(g, game.Player1, topID, zone.Exile) {
		t.Fatal("the play variant did not grant land-play permission for the exiled card")
	}
}
