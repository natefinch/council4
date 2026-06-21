package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestCreateMultipleCopyTokensOfTarget verifies that a create-copy-token effect
// with a fixed count greater than one ("Create two tokens that are copies of
// target creature.", Saw in Half's token clause) creates that many token copies
// of the targeted creature under the effect's controller.
func TestCreateMultipleCopyTokensOfTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Grizzly Bears",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})

	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		Controller: game.Player1,
		Targets:    []game.Target{{Kind: game.TargetPermanent, PermanentID: target.ObjectID}},
	}
	r := &effectResolver{engine: NewEngine(nil), game: g, obj: obj, log: &TurnLog{}}

	resolved := handleCreateToken(r, game.CreateToken{
		Amount: game.Fixed(2),
		Source: game.TokenCopyOf(game.TokenCopySpec{
			Source: game.TokenCopySourceObject,
			Object: game.TargetPermanentReference(0),
		}),
	})
	if !resolved.succeeded {
		t.Fatal("handleCreateToken did not succeed")
	}

	copies := 0
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanent.Controller == game.Player1 && permanent.TokenDef != nil &&
			permanent.TokenDef.Name == "Grizzly Bears" {
			copies++
		}
	}
	if copies != 2 {
		t.Fatalf("created copy tokens = %d, want 2", copies)
	}
}

// TestCreateTappedCopyTokens verifies that a create-copy-token effect with the
// "tapped" entry modifier ("Create two tapped tokens that are copies of ...",
// Skyclave Relic) creates its copies tapped.
func TestCreateTappedCopyTokens(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Grizzly Bears",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})

	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		Controller: game.Player1,
		Targets:    []game.Target{{Kind: game.TargetPermanent, PermanentID: target.ObjectID}},
	}
	r := &effectResolver{engine: NewEngine(nil), game: g, obj: obj, log: &TurnLog{}}

	resolved := handleCreateToken(r, game.CreateToken{
		Amount:      game.Fixed(2),
		EntryTapped: true,
		Source: game.TokenCopyOf(game.TokenCopySpec{
			Source: game.TokenCopySourceObject,
			Object: game.TargetPermanentReference(0),
		}),
	})
	if !resolved.succeeded {
		t.Fatal("handleCreateToken did not succeed")
	}

	tapped := 0
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanent.TokenDef != nil && permanent.TokenDef.Name == "Grizzly Bears" {
			if !permanent.Tapped {
				t.Fatal("copy token created untapped, want tapped")
			}
			tapped++
		}
	}
	if tapped != 2 {
		t.Fatalf("tapped copy tokens = %d, want 2", tapped)
	}
}
