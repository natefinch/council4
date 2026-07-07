package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// driveLocthwainUpkeepWithCard resolves the real Court of Locthwain upkeep ability
// against g, exiling the top card of Player2's library — supplied by def — into
// Player2's exile bucket and installing clause 2's any-mana play permission (and,
// when Player1 is already the monarch, clause 3's until-end-of-turn free cast)
// exactly as a live upkeep trigger would. The exiled card lands in the OPPONENT's
// exile bucket, never the controller's, so callers exercise the cross-player
// action-enumeration path a game driver hits rather than the caster's own bucket.
func driveLocthwainUpkeepWithCard(t *testing.T, engine *Engine, g *game.Game, def *game.CardDef) id.ID {
	t.Helper()
	seq := locthwainUpkeepSequence(t)
	topID := addCardToLibrary(g, game.Player2, def)
	obj := &game.StackObject{
		Kind:         game.StackTriggeredAbility,
		Controller:   game.Player1,
		SourceCardID: g.IDGen.Next(),
		SourceID:     g.IDGen.Next(),
		Targets:      []game.Target{game.PlayerTarget(game.Player2)},
	}
	engine.resolveInstructionWithChoices(g, obj, &seq[0], [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	engine.resolveInstructionWithChoices(g, obj, &seq[1], [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if !g.Players[game.Player2].Exile.Contains(topID) {
		t.Fatal("upkeep did not exile the top card into the opponent's exile bucket")
	}
	return topID
}

// opponentSpellDef is a plain {2}{B}{B} creature standing in for a card Court of
// Locthwain exiles from an opponent's library.
func opponentSpellDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Opponent Spell",
		Types:    []types.Card{types.Creature},
		ManaCost: opt.Val(cost.Mana{cost.O(2), cost.B, cost.B}),
	}}
}

// containsCastFromExile reports whether actions offers a cast of cardID from
// exile, regardless of the alternative cost the payment step later selects.
func containsCastFromExile(actions []action.Action, cardID id.ID) bool {
	for _, act := range actions {
		if payload, ok := act.CastSpellPayload(); ok &&
			payload.CardID == cardID && payload.SourceZone == zone.Exile {
			return true
		}
	}
	return false
}

// containsPlayLandFromExile reports whether actions offers a play of the land
// cardID from exile.
func containsPlayLandFromExile(actions []action.Action, cardID id.ID) bool {
	for _, act := range actions {
		if payload, ok := act.PlayLandPayload(); ok &&
			payload.CardID == cardID && payload.SourceZone == zone.Exile {
			return true
		}
	}
	return false
}

// TestCourtOfLocthwainLegalActionsSurfacesAnyManaPlay proves the "You may play
// that card ... and mana of any type can be spent to cast it" clause is reachable
// from a real game driver: the cast of the opponent-owned exiled card appears in
// the controller's legal actions (payable with off-color mana), even though the
// card physically rests in the opponent's exile bucket. Before the enumeration
// fix this action was never produced because the enumerators scanned only the
// acting player's own exile bucket.
func TestCourtOfLocthwainLegalActionsSurfacesAnyManaPlay(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	topID := driveLocthwainUpkeepWithCard(t, engine, g, opponentSpellDef())

	setSorcerySpeedTurn(g, game.Player1)
	g.Players[game.Player1].ManaPool.Add(mana.G, 4)

	if !containsCastFromExile(engine.legalActions(g, game.Player1), topID) {
		t.Fatal("controller's legal actions must offer casting the opponent-exiled card with any-color mana")
	}
}

// TestCourtOfLocthwainLegalActionsMonarchFreeCast proves the monarch free cast is
// reachable from a game driver and gated on the monarchy. With no mana available
// the only way to cast the pooled opponent-owned card is the free cast, so the
// cast action appears in the controller's legal actions exactly while they are
// the monarch and disappears when they are not.
func TestCourtOfLocthwainLegalActionsMonarchFreeCast(t *testing.T) {
	monarchGame := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	monarchEngine := NewEngine(nil)
	monarchGame.Players[game.Player1].IsMonarch = true
	monarchTopID := driveLocthwainUpkeepWithCard(t, monarchEngine, monarchGame, opponentSpellDef())
	setSorcerySpeedTurn(monarchGame, game.Player1)

	if !containsCastFromExile(monarchEngine.legalActions(monarchGame, game.Player1), monarchTopID) {
		t.Fatal("monarch's legal actions must offer the free cast of the pooled opponent-exiled card with no mana available")
	}

	commonerGame := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	commonerEngine := NewEngine(nil)
	commonerTopID := driveLocthwainUpkeepWithCard(t, commonerEngine, commonerGame, opponentSpellDef())
	setSorcerySpeedTurn(commonerGame, game.Player1)

	if containsCastFromExile(commonerEngine.legalActions(commonerGame, game.Player1), commonerTopID) {
		t.Fatal("a non-monarch with no mana must not be offered the free cast of the pooled card")
	}
}

// TestCourtOfLocthwainLegalActionsGatedToController proves the cross-player
// enumeration is gated to the player who holds the permission: the opponent who
// owns the exiled card — in whose exile bucket it rests — is never offered to
// cast it, even with priority and mana in hand.
func TestCourtOfLocthwainLegalActionsGatedToController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Players[game.Player1].IsMonarch = true
	topID := driveLocthwainUpkeepWithCard(t, engine, g, opponentSpellDef())

	setSorcerySpeedTurn(g, game.Player2)
	g.Players[game.Player2].ManaPool.Add(mana.G, 4)

	if containsCastFromExile(engine.legalActions(g, game.Player2), topID) {
		t.Fatal("the opponent who owns the exiled card must not be offered to cast it")
	}

	setSorcerySpeedTurn(g, game.Player3)
	if containsCastFromExile(engine.legalActions(g, game.Player3), topID) {
		t.Fatal("an unrelated player must not be offered to cast the exiled card")
	}
}

// TestCourtOfLocthwainLegalActionsSurfacesLandPlay proves the "You may play that
// card" clause reaches lands too: when the exiled card is a land, the controller's
// legal actions offer playing it from the opponent's exile bucket.
func TestCourtOfLocthwainLegalActionsSurfacesLandPlay(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	topID := driveLocthwainUpkeepWithCard(t, engine, g, &game.CardDef{CardFace: game.CardFace{
		Name:  "Opponent Land",
		Types: []types.Card{types.Land},
	}})

	setSorcerySpeedTurn(g, game.Player1)

	if !containsPlayLandFromExile(engine.legalActions(g, game.Player1), topID) {
		t.Fatal("controller's legal actions must offer playing the opponent-exiled land from exile")
	}
}
