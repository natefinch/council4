package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// feldonGraveyardCreature is a plain creature card used as the graveyard-card
// copy blueprint for the Feldon-style copy-token tests.
func feldonGraveyardCreature() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Shivan Dragon",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Dragon},
		Power:     opt.Val(game.PT{Value: 5}),
		Toughness: opt.Val(game.PT{Value: 5}),
	}}
}

// TestResolveTargetCardReference verifies that ObjectReferenceTargetCard resolves
// a card chosen for a card target slot (a creature card in a graveyard) to that
// card's snapshot, the blueprint of a Feldon-style copy token.
func TestResolveTargetCardReference(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	cardID := addCardToGraveyard(g, game.Player1, feldonGraveyardCreature())

	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackActivatedAbility,
		Controller: game.Player1,
		Targets:    []game.Target{{Kind: game.TargetCard, CardID: cardID}},
	}

	resolved, ok := resolveObjectReference(g, obj, game.TargetCardReference(0))
	if !ok {
		t.Fatal("resolveObjectReference(TargetCardReference) did not resolve")
	}
	if resolved.snapshot.CardID != cardID {
		t.Fatalf("resolved card = %v, want %v", resolved.snapshot.CardID, cardID)
	}
	if resolved.snapshot.Face != game.FaceFront {
		t.Fatalf("resolved face = %v, want FaceFront", resolved.snapshot.Face)
	}
}

// TestResolveTargetCardReferenceWrongKindFailsClosed verifies that a
// TargetCardReference declines when the target slot is not a card target (e.g. a
// permanent target), so the copy fails closed rather than resolving the wrong
// object.
func TestResolveTargetCardReferenceWrongKindFailsClosed(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := addCombatPermanent(g, game.Player1, feldonGraveyardCreature())

	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackActivatedAbility,
		Controller: game.Player1,
		Targets:    []game.Target{{Kind: game.TargetPermanent, PermanentID: target.ObjectID}},
	}

	if _, ok := resolveObjectReference(g, obj, game.TargetCardReference(0)); ok {
		t.Fatal("resolveObjectReference(TargetCardReference) on a permanent target must fail closed")
	}
}

// TestCreateCopyTokenOfGraveyardCard covers the Feldon of the Third Path copy:
// creating a token that copies a creature card in the controller's graveyard,
// adding the artifact type and haste, without disturbing the copied card in the
// graveyard.
func TestCreateCopyTokenOfGraveyardCard(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	cardID := addCardToGraveyard(g, game.Player1, feldonGraveyardCreature())

	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackActivatedAbility,
		Controller: game.Player1,
		Targets:    []game.Target{{Kind: game.TargetCard, CardID: cardID}},
	}
	r := &effectResolver{engine: NewEngine(nil), game: g, obj: obj, log: &TurnLog{}}

	resolved := handleCreateToken(r, game.CreateToken{
		Amount: game.Fixed(1),
		Source: game.TokenCopyOf(game.TokenCopySpec{
			Source:      game.TokenCopySourceObject,
			Object:      game.TargetCardReference(0),
			AddTypes:    []types.Card{types.Artifact},
			AddKeywords: []game.Keyword{game.Haste},
		}),
	})
	if !resolved.succeeded {
		t.Fatal("handleCreateToken did not succeed")
	}

	var token *game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanent.TokenDef != nil && permanent.TokenDef.Name == "Shivan Dragon" {
			token = permanent
			break
		}
	}
	if token == nil {
		t.Fatal("copy token of the graveyard card was not created")
	}
	if !hasType(token.TokenDef, types.Creature) {
		t.Error("copy token lost the copied creature type")
	}
	if !hasType(token.TokenDef, types.Artifact) {
		t.Error("copy token is not an artifact in addition to its other types")
	}
	if !hasKeyword(g, token, game.Haste) {
		t.Error("copy token did not gain haste")
	}

	// The copied card remains untouched in the graveyard.
	if !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("copied card left the graveyard")
	}
}

func hasType(def *game.CardDef, cardType types.Card) bool {
	return slices.Contains(def.Types, cardType)
}
