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

// countGraveyardCopyTokens returns how many Feldon-style copy tokens (copies of
// the "Shivan Dragon" graveyard card) are on the battlefield.
func countGraveyardCopyTokens(g *game.Game) int {
	count := 0
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanent.TokenDef != nil && permanent.TokenDef.Name == "Shivan Dragon" {
			count++
		}
	}
	return count
}

// TestCreateCopyTokenOfGraveyardCardTwiceSacrificesBoth exercises the repeatable
// Feldon of the Third Path activation. Each activation creates a copy token and
// binds a delayed "Sacrifice it at the beginning of the next end step" to it via
// a source-and-link-scoped linked reference that is constant across activations.
// After the first end-step sacrifice the dead token lingers in last-known
// information, so the second activation must rebind the link to the new token;
// otherwise the delayed sacrifice resolves the dead first token and leaks the
// second copy. This is the regression the clear-before-publish fix prevents.
func TestCreateCopyTokenOfGraveyardCardTwiceSacrificesBoth(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	feldon := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Feldon of the Third Path",
		Types: []types.Card{types.Creature},
	}})
	cardID := addCardToGraveyard(g, game.Player1, feldonGraveyardCreature())

	// activate builds the resolver for one activation of Feldon's ability. Every
	// activation shares Feldon's source identity, so the published link key is
	// the same across activations — exactly the condition that made the leak
	// possible.
	activate := func() *effectResolver {
		obj := &game.StackObject{
			ID:           g.IDGen.Next(),
			Kind:         game.StackActivatedAbility,
			Controller:   game.Player1,
			SourceID:     feldon.ObjectID,
			SourceCardID: feldon.CardInstanceID,
			Targets:      []game.Target{{Kind: game.TargetCard, CardID: cardID}},
		}
		return &effectResolver{engine: NewEngine(nil), game: g, obj: obj, log: &TurnLog{}}
	}

	createToken := game.CreateToken{
		Amount: game.Fixed(1),
		Source: game.TokenCopyOf(game.TokenCopySpec{
			Source:      game.TokenCopySourceObject,
			Object:      game.TargetCardReference(0),
			AddTypes:    []types.Card{types.Artifact},
			AddKeywords: []game.Keyword{game.Haste},
		}),
		PublishLinked: game.LinkedKey("delayed-sacrifice-1"),
	}
	sacrifice := game.Sacrifice{Object: game.LinkedObjectReference("delayed-sacrifice-1")}

	// First activation, then its end-step delayed sacrifice.
	r1 := activate()
	if !handleCreateToken(r1, createToken).succeeded {
		t.Fatal("first activation did not create a copy token")
	}
	if got := countGraveyardCopyTokens(g); got != 1 {
		t.Fatalf("after first activation, copy tokens = %d, want 1", got)
	}
	handleSacrifice(r1, sacrifice)
	if got := countGraveyardCopyTokens(g); got != 0 {
		t.Fatalf("first end-step sacrifice left %d copy token(s), want 0", got)
	}

	// Second activation, then its end-step delayed sacrifice. The delayed
	// sacrifice must bind to the freshly-created token rather than no-op on the
	// dead first token's last-known information.
	r2 := activate()
	if !handleCreateToken(r2, createToken).succeeded {
		t.Fatal("second activation did not create a copy token")
	}
	if got := countGraveyardCopyTokens(g); got != 1 {
		t.Fatalf("after second activation, copy tokens = %d, want 1", got)
	}
	handleSacrifice(r2, sacrifice)
	if got := countGraveyardCopyTokens(g); got != 0 {
		t.Fatalf("second end-step sacrifice leaked %d copy token(s): the delayed sacrifice must bind to the new token, not the dead first token", got)
	}
}
