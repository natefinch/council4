package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func spiritTokenDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Spirit",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Sub("Spirit")},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}}
}

// A CreateToken whose Recipient references a targeted player creates the token
// under that player's control and ownership, not the resolving controller's
// ("Target opponent creates a 1/1 colorless Spirit creature token.", Forbidden
// Orchard).
func TestCreateTokenTargetedPlayerRecipientControlsToken(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	obj := &game.StackObject{
		Controller: game.Player1,
		Targets:    []game.Target{{Kind: game.TargetPlayer, PlayerID: game.Player2}},
	}
	resolveInstruction(engine, g, obj, game.CreateToken{
		Amount:    game.Fixed(1),
		Source:    game.TokenDef(spiritTokenDef()),
		Recipient: opt.Val(game.TargetPlayerReference(0)),
	}, &TurnLog{})

	token := newlyCreatedToken(g)
	if token == nil {
		t.Fatal("targeted-recipient token did not enter the battlefield")
	}
	if token.Controller != game.Player2 {
		t.Fatalf("token controller = %v, want the targeted opponent Player2", token.Controller)
	}
	if token.Owner != game.Player2 {
		t.Fatalf("token owner = %v, want the targeted opponent Player2", token.Owner)
	}
}

// tokensByController counts the tokens each player controls on the battlefield.
func tokensByController(g *game.Game) map[game.PlayerID]int {
	counts := make(map[game.PlayerID]int)
	for _, permanent := range g.Battlefield {
		if permanent.Token {
			counts[permanent.Controller]++
		}
	}
	return counts
}

// A CreateToken whose RecipientGroup is all players creates one token under each
// player's control ("Each player creates a 1/1 ... creature token.", Grismold,
// the Dreadsower).
func TestCreateTokenEachPlayerRecipientGroup(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	obj := &game.StackObject{Controller: game.Player1}
	resolveInstruction(engine, g, obj, game.CreateToken{
		Amount:         game.Fixed(1),
		Source:         game.TokenDef(spiritTokenDef()),
		RecipientGroup: game.AllPlayersReference(),
	}, &TurnLog{})

	counts := tokensByController(g)
	if len(counts) != game.NumPlayers {
		t.Fatalf("tokens created for %d players, want all %d", len(counts), game.NumPlayers)
	}
	for player, n := range counts {
		if n != 1 {
			t.Fatalf("player %v controls %d tokens, want 1", player, n)
		}
	}
}

// addControlledPermanent puts a battlefield permanent under playerID's control
// whose card types are cardTypes, backing the "who controls <selection>" group
// recipient tests. It mirrors the minimal card-instance-and-permanent wiring the
// other rules tests use.
func addControlledPermanent(g *game.Game, playerID game.PlayerID, cardTypes ...types.Card) *game.Permanent {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID:    cardID,
		Def:   &game.CardDef{CardFace: game.CardFace{Name: "Test Permanent", Types: cardTypes}},
		Owner: playerID,
	}
	permanent := &game.Permanent{
		ObjectID:       g.IDGen.Next(),
		CardInstanceID: cardID,
		Owner:          playerID,
		Controller:     playerID,
	}
	g.Battlefield = append(g.Battlefield, permanent)
	return permanent
}

// artifactOrEnchantmentSelection is the "artifact or enchantment" union the
// Fade from History group recipient filters on.
func artifactOrEnchantmentSelection() game.Selection {
	return game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment}}
}

// A CreateToken whose RecipientGroup carries a ControllingMatching qualifier
// creates a token only for group members controlling a matching permanent ("Each
// player who controls an artifact or enchantment creates a 2/2 green Bear
// creature token.", Fade from History). A member controlling neither is skipped;
// a member controlling both an artifact and an enchantment still receives a
// single token.
func TestCreateTokenGroupControlsMatchingFiltersMembers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	addControlledPermanent(g, game.Player1, types.Artifact)
	addControlledPermanent(g, game.Player2, types.Enchantment)
	addControlledPermanent(g, game.Player3, types.Creature)
	addControlledPermanent(g, game.Player4, types.Artifact)
	addControlledPermanent(g, game.Player4, types.Enchantment)

	obj := &game.StackObject{Controller: game.Player1}
	resolveInstruction(engine, g, obj, game.CreateToken{
		Amount:         game.Fixed(1),
		Source:         game.TokenDef(spiritTokenDef()),
		RecipientGroup: game.AllPlayersReference().ControllingMatching(artifactOrEnchantmentSelection()),
	}, &TurnLog{})

	counts := tokensByController(g)
	if counts[game.Player3] != 0 {
		t.Fatalf("Player3 controls neither type but got %d tokens, want 0", counts[game.Player3])
	}
	for _, player := range []game.PlayerID{game.Player1, game.Player2, game.Player4} {
		if counts[player] != 1 {
			t.Fatalf("player %v controls %d tokens, want 1", player, counts[player])
		}
	}
}

// A ControllingMatching group recipient skips eliminated players even when they
// control a matching permanent, because the base group already excludes them.
func TestCreateTokenGroupControlsMatchingSkipsEliminated(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	addControlledPermanent(g, game.Player1, types.Artifact)
	addControlledPermanent(g, game.Player2, types.Enchantment)
	g.Players[game.Player2].Eliminated = true

	obj := &game.StackObject{Controller: game.Player1}
	resolveInstruction(engine, g, obj, game.CreateToken{
		Amount:         game.Fixed(1),
		Source:         game.TokenDef(spiritTokenDef()),
		RecipientGroup: game.AllPlayersReference().ControllingMatching(artifactOrEnchantmentSelection()),
	}, &TurnLog{})

	counts := tokensByController(g)
	if counts[game.Player2] != 0 {
		t.Fatalf("eliminated Player2 got %d tokens, want 0", counts[game.Player2])
	}
	if counts[game.Player1] != 1 {
		t.Fatalf("Player1 controls %d tokens, want 1", counts[game.Player1])
	}
}

// A ControllingMatching group recipient composes with the opponents base group:
// only opponents controlling a matching permanent receive a token, and the
// resolving controller never does even while controlling one.
func TestCreateTokenGroupControlsMatchingOpponentsBase(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	addControlledPermanent(g, game.Player1, types.Artifact)
	addControlledPermanent(g, game.Player2, types.Enchantment)
	addControlledPermanent(g, game.Player3, types.Land)

	obj := &game.StackObject{Controller: game.Player1}
	resolveInstruction(engine, g, obj, game.CreateToken{
		Amount:         game.Fixed(1),
		Source:         game.TokenDef(spiritTokenDef()),
		RecipientGroup: game.OpponentsReference().ControllingMatching(artifactOrEnchantmentSelection()),
	}, &TurnLog{})

	counts := tokensByController(g)
	if counts[game.Player1] != 0 {
		t.Fatalf("controller Player1 got %d tokens, want 0", counts[game.Player1])
	}
	if counts[game.Player2] != 1 {
		t.Fatalf("opponent Player2 controls an enchantment but got %d tokens, want 1", counts[game.Player2])
	}
	if counts[game.Player3] != 0 {
		t.Fatalf("opponent Player3 controls neither type but got %d tokens, want 0", counts[game.Player3])
	}
}

// Resolving the full Fade from History sequence — create a token for each player
// controlling an artifact or enchantment, then destroy all artifacts and
// enchantments — creates the tokens before the mass destroy runs, so the
// creation observes the qualifying permanents and the created creature tokens
// survive the destroy that removes the artifacts and enchantments.
func TestCreateTokenGroupControlsMatchingThenDestroyOrdering(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	artifact := addControlledPermanent(g, game.Player1, types.Artifact)
	enchantment := addControlledPermanent(g, game.Player2, types.Enchantment)
	addControlledPermanent(g, game.Player3, types.Creature)

	content := game.Mode{Sequence: []game.Instruction{
		{Primitive: game.CreateToken{
			Amount:         game.Fixed(1),
			Source:         game.TokenDef(spiritTokenDef()),
			RecipientGroup: game.AllPlayersReference().ControllingMatching(artifactOrEnchantmentSelection()),
		}},
		{Primitive: game.Destroy{Group: game.BattlefieldGroup(artifactOrEnchantmentSelection())}},
	}}.Ability()

	obj := &game.StackObject{Controller: game.Player1}
	engine.resolveAbilityContentWithChoices(g, obj, content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if _, ok := permanentByObjectID(g, artifact.ObjectID); ok {
		t.Fatal("artifact survived the mass destroy")
	}
	if _, ok := permanentByObjectID(g, enchantment.ObjectID); ok {
		t.Fatal("enchantment survived the mass destroy")
	}
	counts := tokensByController(g)
	if counts[game.Player1] != 1 {
		t.Fatalf("Player1 controls %d Bear tokens, want 1 surviving the destroy", counts[game.Player1])
	}
	if counts[game.Player2] != 1 {
		t.Fatalf("Player2 controls %d Bear tokens, want 1 surviving the destroy", counts[game.Player2])
	}
	if counts[game.Player3] != 0 {
		t.Fatalf("Player3 controls neither type but got %d tokens, want 0", counts[game.Player3])
	}
}

// A ControllingMatching group recipient applies token-creation replacements per
// recipient: with an any-controller token doubler (Primal Vigor) in play, each
// qualifying member's single Bear is doubled to two, while a member controlling
// neither type still receives none. This confirms the group path routes each
// recipient's creation through the replacement system independently.
func TestCreateTokenGroupControlsMatchingReplacementPerRecipient(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	// Player1's doubler is itself an enchantment, so Player1 qualifies too.
	addReplacementPermanent(t, g, game.Player1, anyControllerTokenDoublingCardDef())
	addControlledPermanent(g, game.Player2, types.Artifact)
	addControlledPermanent(g, game.Player3, types.Creature)

	obj := &game.StackObject{Controller: game.Player1}
	resolveInstruction(engine, g, obj, game.CreateToken{
		Amount:         game.Fixed(1),
		Source:         game.TokenDef(spiritTokenDef()),
		RecipientGroup: game.AllPlayersReference().ControllingMatching(artifactOrEnchantmentSelection()),
	}, &TurnLog{})

	counts := tokensByController(g)
	if counts[game.Player1] != 2 {
		t.Fatalf("Player1 controls %d tokens, want 2 (doubled)", counts[game.Player1])
	}
	if counts[game.Player2] != 2 {
		t.Fatalf("Player2 controls %d tokens, want 2 (doubled by any-controller replacement)", counts[game.Player2])
	}
	if counts[game.Player3] != 0 {
		t.Fatalf("Player3 controls neither type but got %d tokens, want 0", counts[game.Player3])
	}
}

// A CreateToken whose RecipientGroup is the controller's opponents creates one
// token under each opponent's control and none for the controller ("Each
// opponent creates a 1/1 white Human creature token.", Slaughter Specialist).
func TestCreateTokenEachOpponentRecipientGroup(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	obj := &game.StackObject{Controller: game.Player1}
	resolveInstruction(engine, g, obj, game.CreateToken{
		Amount:         game.Fixed(1),
		Source:         game.TokenDef(spiritTokenDef()),
		RecipientGroup: game.OpponentsReference(),
	}, &TurnLog{})

	counts := tokensByController(g)
	if counts[game.Player1] != 0 {
		t.Fatalf("controller Player1 controls %d tokens, want 0", counts[game.Player1])
	}
	if len(counts) != game.NumPlayers-1 {
		t.Fatalf("tokens created for %d players, want %d opponents", len(counts), game.NumPlayers-1)
	}
	for player, n := range counts {
		if n != 1 {
			t.Fatalf("opponent %v controls %d tokens, want 1", player, n)
		}
	}
}
