package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// assembleEntmootTreefolkDef is the token minted by Assemble the Entmoot: a green
// Treefolk whose printed X/X is filled in at resolution from the CreateToken
// power/toughness overrides.
func assembleEntmootTreefolkDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Treefolk",
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Treefolk},
		Colors:   []color.Color{color.Green},
	}}
}

// assembleEntmootSequence models the lowered activated body of Assemble the
// Entmoot: create three tapped X/X green Treefolk tokens (X = life gained this
// turn) published under a link key, then put a reach counter on each of them via
// that link group.
func assembleEntmootSequence() []game.Instruction {
	const linkKey = game.LinkedKey("assemble-entmoot-treefolk")
	size := game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountLifeGainedThisTurn})
	return []game.Instruction{
		{Primitive: game.CreateToken{
			Amount:        game.Fixed(3),
			Source:        game.TokenDef(assembleEntmootTreefolkDef()),
			Power:         opt.Val(size),
			Toughness:     opt.Val(size),
			EntryTapped:   true,
			PublishLinked: linkKey,
		}},
		{Primitive: game.AddCounter{
			Group:       game.LinkedObjectsGroup(linkKey),
			CounterKind: counter.Reach,
			Amount:      game.Fixed(1),
		}},
	}
}

func assembleEntmootTreefolkTokens(g *game.Game) []*game.Permanent {
	var tokens []*game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanentTokenName(permanent) == "Treefolk" {
			tokens = append(tokens, permanent)
		}
	}
	return tokens
}

// TestAssembleEntmootThreeTappedTokensReachCounterOnEach resolves the activated
// body with life gained this turn = 4: three 4/4 tapped Treefolk are created and
// each of them receives a reach counter via the linked-object group.
func TestAssembleEntmootThreeTappedTokensReachCounterOnEach(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	emitEvent(g, game.Event{Kind: game.EventLifeGained, Player: game.Player1, Amount: 4})

	addInstructionSpellToStackForController(g, game.Player1, assembleEntmootSequence(), nil)
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	tokens := assembleEntmootTreefolkTokens(g)
	if len(tokens) != 3 {
		t.Fatalf("created Treefolk tokens = %d, want 3", len(tokens))
	}
	for _, token := range tokens {
		if !token.Tapped {
			t.Fatalf("token %d Tapped = false, want tapped entry", token.ObjectID)
		}
		if got := token.Counters.Get(counter.Reach); got != 1 {
			t.Fatalf("token %d reach counters = %d, want 1 on each of them", token.ObjectID, got)
		}
		if !token.TokenDef.Power.Exists || token.TokenDef.Power.Val.Value != 4 ||
			!token.TokenDef.Toughness.Exists || token.TokenDef.Toughness.Val.Value != 4 {
			t.Fatalf("token P/T = %+v/%+v, want 4/4 from life gained this turn",
				token.TokenDef.Power, token.TokenDef.Toughness)
		}
	}
}

// TestAssembleEntmootTokenDoublerCountersEveryCopy proves the "on each of them"
// linked group includes replacement-created copies: with a token doubler in
// play the three tokens become six, and every one of the six receives a reach
// counter because CreateToken.PublishLinked records all created tokens.
func TestAssembleEntmootTokenDoublerCountersEveryCopy(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addReplacementPermanent(t, g, game.Player1, tokenDoublingReplacementCardDef())
	emitEvent(g, game.Event{Kind: game.EventLifeGained, Player: game.Player1, Amount: 2})

	addInstructionSpellToStackForController(g, game.Player1, assembleEntmootSequence(), nil)
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	tokens := assembleEntmootTreefolkTokens(g)
	if len(tokens) != 6 {
		t.Fatalf("doubled Treefolk tokens = %d, want 6", len(tokens))
	}
	for _, token := range tokens {
		if got := token.Counters.Get(counter.Reach); got != 1 {
			t.Fatalf("doubled token %d reach counters = %d, want 1 on every copy", token.ObjectID, got)
		}
	}
}

// TestAssembleEntmootZeroLifeGainedMakesZeroPowerTokens proves the dynamic size
// collapses to 0/0 when no life was gained this turn: the three tapped tokens
// are still created and still receive a reach counter each (the group placement
// does not depend on the token's size).
func TestAssembleEntmootZeroLifeGainedMakesZeroPowerTokens(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	addInstructionSpellToStackForController(g, game.Player1, assembleEntmootSequence(), nil)
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	tokens := assembleEntmootTreefolkTokens(g)
	if len(tokens) != 3 {
		t.Fatalf("created Treefolk tokens = %d, want 3", len(tokens))
	}
	for _, token := range tokens {
		if !token.TokenDef.Power.Exists || token.TokenDef.Power.Val.Value != 0 ||
			!token.TokenDef.Toughness.Exists || token.TokenDef.Toughness.Val.Value != 0 {
			t.Fatalf("token P/T = %+v/%+v, want 0/0 with no life gained",
				token.TokenDef.Power, token.TokenDef.Toughness)
		}
		if got := token.Counters.Get(counter.Reach); got != 1 {
			t.Fatalf("token %d reach counters = %d, want 1 even at 0/0", token.ObjectID, got)
		}
	}
}
