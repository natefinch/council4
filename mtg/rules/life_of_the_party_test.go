package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestLifeOfThePartyCreatesGoadedCopiesForEachOpponent verifies the runtime
// mechanics backing Life of the Party's enter trigger: each opponent creates a
// token that's a copy of the source, every created token is published under a
// link key, and a following rest-of-game goad binds exactly those linked tokens.
// It checks that every opponent (not the controller) receives a copy, that the
// copies are goaded by the controller, and that the goad carries the
// rest-of-game flag so it survives the controller's untap.
func TestLifeOfThePartyCreatesGoadedCopiesForEachOpponent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Life of the Party",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})

	obj := &game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackTriggeredAbility,
		Controller:   game.Player1,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
	}
	r := &effectResolver{engine: NewEngine(nil), game: g, obj: obj, log: &TurnLog{}}

	const key game.LinkedKey = "goad-created-tokens"
	created := handleCreateToken(r, game.CreateToken{
		Amount: game.Fixed(1),
		Source: game.TokenCopyOf(game.TokenCopySpec{
			Source: game.TokenCopySourceObject,
			Object: game.SourcePermanentReference(),
		}),
		RecipientGroup: game.OpponentsReference(),
		PublishLinked:  key,
	})
	if !created.succeeded {
		t.Fatal("handleCreateToken did not succeed")
	}

	goaded := handleGoad(r, game.Goad{
		Group:      game.LinkedObjectsGroup(key),
		RestOfGame: true,
	})
	if !goaded.succeeded {
		t.Fatal("handleGoad did not succeed")
	}

	copies := 0
	for _, permanent := range g.Battlefield {
		if permanent == nil || !permanent.Token || permanent.TokenDef == nil ||
			permanent.TokenDef.Name != "Life of the Party" {
			continue
		}
		copies++
		if permanent.Controller == game.Player1 {
			t.Fatal("token copy controlled by the ability controller Player1; want an opponent")
		}
		status, ok := permanent.Goaded[game.Player1]
		if !ok {
			t.Fatalf("token copy for %v is not goaded by the controller", permanent.Controller)
		}
		if !status.RestOfGame {
			t.Fatalf("token copy for %v goad is not rest-of-game", permanent.Controller)
		}
	}
	if copies != game.NumPlayers-1 {
		t.Fatalf("goaded token copies = %d, want %d (one per opponent)", copies, game.NumPlayers-1)
	}
}

// TestRestOfGameGoadSurvivesControllerUntap verifies that a rest-of-game goad is
// not cleared by the goading player's untap-step expiry, unlike the ordinary
// turn-limited goad keyword action.
func TestRestOfGameGoadSurvivesControllerUntap(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	restOfGame := addCombatCreaturePermanent(g, game.Player2)
	turnLimited := addCombatCreaturePermanent(g, game.Player3)
	goadPermanent(g, restOfGame, game.Player1, true)
	goadPermanent(g, turnLimited, game.Player1, false)

	// Advance to a later turn where Player1 (the goading player) is active, so the
	// turn-limited goad reaches its untap-step expiry.
	g.Turn.TurnNumber++
	g.Turn.ActivePlayer = game.Player1
	expireGoadForActivePlayer(g)

	if _, ok := restOfGame.Goaded[game.Player1]; !ok {
		t.Fatal("rest-of-game goad was cleared at the goading player's untap")
	}
	if _, ok := turnLimited.Goaded[game.Player1]; ok {
		t.Fatal("turn-limited goad survived the goading player's untap")
	}
}
