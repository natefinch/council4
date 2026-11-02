package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: Oracle text from the pinned corpus for Molten Duplication
// (fdbe1ac1-461f-4746-a8d8-6c8dea2c97c6): "Create a token that's a copy of
// target artifact or creature you control, except it's an artifact in addition
// to its other types. It gains haste until end of turn. Sacrifice it at the
// beginning of the next end step."; Flameshadow Conjuring
// (ae113258-288f-47d5-8f79-e2e51b6326de): "Whenever a nontoken creature you
// control enters, you may pay {R}. If you do, create a token that's a copy of
// that creature. That token gains haste. Exile it at the beginning of the next
// end step."; and Ojer Taq, Deepest Foundation (1ca79dd4-67fc-496c-96fc-
// 489b039c4932): "If one or more creature tokens would be created under your
// control, three times that many of those tokens are created instead."
// https://api.scryfall.com/cards/<id>/rulings returned five 2024-04-12 copy
// rulings for Molten Duplication and eight 2015-06-22 rulings for Flameshadow
// Conjuring, including "If the token isn't exiled at the beginning of the next
// end step (perhaps because the delayed triggered ability is countered), it
// remains on the battlefield indefinitely. It continues to have haste." and
// "If another creature becomes or enters the battlefield as a copy of the
// token, that creature won't have haste and it won't be exiled." CR 2026-09-25
// from the official TXT
// (https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt):
// 603.7c (a delayed trigger affects the particular objects it refers to), 608.2c
// ("it" is the created token, and every token a replaced creation made), 611.2c
// (a resolving effect's characteristic change locks in its affected objects),
// 616.1 (replacement effects modify the creation event) and 707.9 (only an
// "except" clause of a copy effect becomes a copiable value; "gains haste" does
// not). So the haste and the delayed disposal act on every created token, never
// the copied original, and each resolution keeps its own tokens.

func moltenDuplicationGame(t *testing.T, ojerTaq bool) (*game.Game, *Engine) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	if ojerTaq {
		addReplacementPermanent(t, g, game.Player1, registryCard(t, "Ojer Taq, Deepest Foundation"))
	}
	return g, NewEngineWithPrograms(nil, cards.NewDefaultRegistry().Programs())
}

func copyTestBear(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}}
}

// castMoltenDuplication casts a real Molten Duplication targeting original and
// resolves it, returning the copy tokens that resolution created.
func castMoltenDuplication(t *testing.T, g *game.Game, engine *Engine, original *game.Permanent) []id.ID {
	t.Helper()
	name := permanentName(g, original)
	before := playerTokensNamed(g, game.Player1, name)
	spell := addCardToHand(g, game.Player1, registryCard(t, "Molten Duplication"))
	g.Players[game.Player1].ManaPool.Add(mana.R, 2)
	setMainPhasePriority(g, game.Player1)
	targets := []game.Target{game.PermanentTarget(original.ObjectID)}
	if !engine.applyAction(g, game.Player1, action.CastSpell(spell, targets, 0, nil)) {
		t.Fatal("could not cast Molten Duplication")
	}
	resolveStackWithTriggers(engine, g, [game.NumPlayers]PlayerAgent{})
	return newTokens(before, playerTokensNamed(g, game.Player1, name))
}

// Molten Duplication's "It gains haste ... Sacrifice it" is the copy token, not
// the targeted original: the token alone gains haste until end of turn, the
// delayed sacrifice is bound to it, and the original survives the end step.
func TestMoltenDuplicationHastesAndSacrificesTheTokenNotTheOriginal(t *testing.T) {
	g, engine := moltenDuplicationGame(t, false)
	original := addCombatPermanent(g, game.Player1, copyTestBear("Molten Bear"))
	tokens := castMoltenDuplication(t, g, engine, original)
	if len(tokens) != 1 {
		t.Fatalf("copy tokens = %d, want 1", len(tokens))
	}
	token, _ := g.PermanentByID(tokens[0])
	if !hasKeyword(g, token, game.Haste) || !permanentHasType(g, token, types.Artifact) {
		t.Fatal("the copy token is not a hasty artifact")
	}
	if hasKeyword(g, original, game.Haste) || permanentHasType(g, original, types.Artifact) {
		t.Fatal("the targeted original gained haste or the artifact type")
	}
	views := observe(g, game.Player2).DelayedTriggers()
	if len(views) != 1 || !sameIDs(views[0].BoundObjectIDs, tokens) {
		t.Fatalf("pending delayed triggers = %+v, want one sacrifice bound to token %v", views, tokens)
	}

	cleanup := g.Clone()
	expireCleanupDurations(cleanup)
	if cleanupToken, _ := cleanup.PermanentByID(tokens[0]); hasKeyword(cleanup, cleanupToken, game.Haste) {
		t.Fatal("the copy token kept haste after the turn's cleanup")
	}

	fireEndStep(t, g, engine)
	engine.resolveTopOfStack(g, &TurnLog{})
	if left := onBattlefield(g, tokens); len(left) != 0 {
		t.Fatalf("copy token %v remained after the next end step's sacrifice", left)
	}
	if _, ok := g.PermanentByID(original.ObjectID); !ok {
		t.Fatal("the targeted original was sacrificed instead of its copy")
	}
}

// Under Ojer Taq, Molten Duplication creates three copies; each gains haste and
// the one delayed sacrifice is bound to, and sacrifices, all three.
func TestMoltenDuplicationSacrificesEveryTokenOjerTaqCreates(t *testing.T) {
	g, engine := moltenDuplicationGame(t, true)
	original := addCombatPermanent(g, game.Player1, copyTestBear("Molten Bear"))
	tokens := castMoltenDuplication(t, g, engine, original)
	if len(tokens) != 3 {
		t.Fatalf("copy tokens under Ojer Taq = %d, want 3", len(tokens))
	}
	for _, objectID := range tokens {
		token, _ := g.PermanentByID(objectID)
		if !hasKeyword(g, token, game.Haste) {
			t.Fatalf("copy token %d lacks haste", objectID)
		}
	}
	if hasKeyword(g, original, game.Haste) {
		t.Fatal("the targeted original gained haste")
	}
	views := observe(g, game.Player2).DelayedTriggers()
	if len(views) != 1 || !sameIDs(views[0].BoundObjectIDs, tokens) {
		t.Fatalf("pending delayed triggers = %+v, want one sacrifice bound to %v", views, tokens)
	}
	fireEndStep(t, g, engine)
	engine.resolveTopOfStack(g, &TurnLog{})
	if left := onBattlefield(g, tokens); len(left) != 0 {
		t.Fatalf("copy tokens %v remained after the next end step's sacrifice", left)
	}
	if _, ok := g.PermanentByID(original.ObjectID); !ok {
		t.Fatal("the targeted original was sacrificed")
	}
}

// Two Molten Duplication resolutions under Ojer Taq keep independent bindings:
// each delayed sacrifice acts on exactly its own resolution's three copies.
func TestMoltenDuplicationSeparateResolutionsKeepTheirOwnTokens(t *testing.T) {
	g, engine := moltenDuplicationGame(t, true)
	original := addCombatPermanent(g, game.Player1, copyTestBear("Molten Bear"))
	first := castMoltenDuplication(t, g, engine, original)
	second := castMoltenDuplication(t, g, engine, original)
	if len(first) != 3 || len(second) != 3 {
		t.Fatalf("copy tokens per resolution = %d and %d, want 3 each", len(first), len(second))
	}
	views := observe(g, game.Player1).DelayedTriggers()
	if len(views) != 2 || !sameIDs(views[0].BoundObjectIDs, first) || !sameIDs(views[1].BoundObjectIDs, second) {
		t.Fatalf("pending delayed triggers = %+v, want sacrifices bound to %v and %v", views, first, second)
	}
	fireEndStep(t, g, engine)
	top, _ := g.Stack.Peek()
	topGroup, otherGroup := first, second
	if !sameIDs(top.CapturedObjectIDs, first) {
		topGroup, otherGroup = second, first
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if left := onBattlefield(g, topGroup); len(left) != 0 {
		t.Fatalf("copy tokens %v remained after their own sacrifice resolved", left)
	}
	if left := onBattlefield(g, otherGroup); len(left) != 3 {
		t.Fatalf("the other resolution's tokens on the battlefield = %v, want all of %v", left, otherGroup)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if left := onBattlefield(g, append(slices.Clone(first), second...)); len(left) != 0 {
		t.Fatalf("copy tokens %v remained after both sacrifices resolved", left)
	}
	if _, ok := g.PermanentByID(original.ObjectID); !ok {
		t.Fatal("the targeted original was sacrificed")
	}
}

func flameshadowGame(t *testing.T, ojerTaq bool) (*game.Game, *Engine) {
	t.Helper()
	g, engine := moltenDuplicationGame(t, ojerTaq)
	addCombatPermanent(g, game.Player1, registryCard(t, "Flameshadow Conjuring"))
	return g, engine
}

// castCreatureUnderFlameshadow casts a nontoken creature, pays {R} for
// Flameshadow Conjuring's trigger, and returns the new creature and the copy
// tokens that trigger's resolution created.
func castCreatureUnderFlameshadow(t *testing.T, g *game.Game, engine *Engine, name string) (*game.Permanent, []id.ID) {
	t.Helper()
	def := copyTestBear(name)
	def.ManaCost = opt.Val(cost.Mana{cost.O(1)})
	spell := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.R, 2)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(spell, nil, 0, nil)) {
		t.Fatalf("could not cast %s", name)
	}
	resolveStackWithTriggers(engine, g, [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: true}})
	var creature *game.Permanent
	for _, permanent := range g.Battlefield {
		if !permanent.Token && permanent.CardInstanceID == spell {
			creature = permanent
		}
	}
	if creature == nil {
		t.Fatalf("%s did not enter the battlefield", name)
	}
	return creature, playerTokensNamed(g, game.Player1, name)
}

// Under Ojer Taq, Flameshadow Conjuring's paid trigger creates three copies;
// each gains haste and the one delayed exile is bound to, and exiles, all three
// while the nontoken original stays.
func TestFlameshadowConjuringExilesEveryCopyOjerTaqCreates(t *testing.T) {
	g, engine := flameshadowGame(t, true)
	original, tokens := castCreatureUnderFlameshadow(t, g, engine, "Flameshadow Bear")
	if len(tokens) != 3 {
		t.Fatalf("copy tokens under Ojer Taq = %d, want 3", len(tokens))
	}
	for _, objectID := range tokens {
		token, _ := g.PermanentByID(objectID)
		if !hasKeyword(g, token, game.Haste) {
			t.Fatalf("copy token %d lacks haste", objectID)
		}
	}
	views := observe(g, game.Player2).DelayedTriggers()
	if len(views) != 1 || !sameIDs(views[0].BoundObjectIDs, tokens) {
		t.Fatalf("pending delayed triggers = %+v, want one exile bound to %v", views, tokens)
	}
	fireEndStep(t, g, engine)
	engine.resolveTopOfStack(g, &TurnLog{})
	if left := onBattlefield(g, tokens); len(left) != 0 {
		t.Fatalf("copy tokens %v remained after the next end step's exile", left)
	}
	if _, ok := g.PermanentByID(original.ObjectID); !ok {
		t.Fatal("the nontoken original left the battlefield")
	}
}

// Two Flameshadow Conjuring resolutions under Ojer Taq keep independent
// bindings: each delayed exile acts on exactly its own creature's three copies.
func TestFlameshadowConjuringSeparateResolutionsKeepTheirOwnCopies(t *testing.T) {
	g, engine := flameshadowGame(t, true)
	_, first := castCreatureUnderFlameshadow(t, g, engine, "First Bear")
	_, second := castCreatureUnderFlameshadow(t, g, engine, "Second Bear")
	if len(first) != 3 || len(second) != 3 {
		t.Fatalf("copy tokens per resolution = %d and %d, want 3 each", len(first), len(second))
	}
	views := observe(g, game.Player1).DelayedTriggers()
	if len(views) != 2 || !sameIDs(views[0].BoundObjectIDs, first) || !sameIDs(views[1].BoundObjectIDs, second) {
		t.Fatalf("pending delayed triggers = %+v, want exiles bound to %v and %v", views, first, second)
	}
	fireEndStep(t, g, engine)
	top, _ := g.Stack.Peek()
	topGroup, otherGroup := first, second
	if !sameIDs(top.CapturedObjectIDs, first) {
		topGroup, otherGroup = second, first
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if left := onBattlefield(g, topGroup); len(left) != 0 {
		t.Fatalf("copy tokens %v remained after their own exile resolved", left)
	}
	if left := onBattlefield(g, otherGroup); len(left) != 3 {
		t.Fatalf("the other resolution's copies on the battlefield = %v, want all of %v", left, otherGroup)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if left := onBattlefield(g, append(slices.Clone(first), second...)); len(left) != 0 {
		t.Fatalf("copy tokens %v remained after both exiles resolved", left)
	}
}

// Flameshadow Conjuring's ruling: a creature that enters as a copy of its
// token (Cackling Counterpart) doesn't have haste and isn't exiled, because
// "That token gains haste" is not part of the copy effect (CR 707.9).
func TestFlameshadowConjuringCopyOfTheTokenLacksHasteAndStays(t *testing.T) {
	g, engine := flameshadowGame(t, false)
	_, tokens := castCreatureUnderFlameshadow(t, g, engine, "Flameshadow Bear")
	if len(tokens) != 1 {
		t.Fatalf("copy tokens = %d, want 1", len(tokens))
	}
	flameshadowToken, _ := g.PermanentByID(tokens[0])
	if !hasKeyword(g, flameshadowToken, game.Haste) {
		t.Fatal("Flameshadow Conjuring's token lacks haste")
	}

	counterpart := addCardToHand(g, game.Player1, registryCard(t, "Cackling Counterpart"))
	g.Players[game.Player1].ManaPool.Add(mana.U, 3)
	setMainPhasePriority(g, game.Player1)
	targets := []game.Target{game.PermanentTarget(flameshadowToken.ObjectID)}
	if !engine.applyAction(g, game.Player1, action.CastSpell(counterpart, targets, 0, nil)) {
		t.Fatal("could not cast Cackling Counterpart")
	}
	resolveStackWithTriggers(engine, g, [game.NumPlayers]PlayerAgent{})
	copies := newTokens(tokens, playerTokensNamed(g, game.Player1, "Flameshadow Bear"))
	if len(copies) != 1 {
		t.Fatalf("Cackling Counterpart copies = %d, want 1", len(copies))
	}
	copyOfToken, _ := g.PermanentByID(copies[0])
	if hasKeyword(g, copyOfToken, game.Haste) {
		t.Fatal("the copy of Flameshadow Conjuring's token copied its haste")
	}

	fireEndStep(t, g, engine)
	engine.resolveTopOfStack(g, &TurnLog{})
	if _, ok := g.PermanentByID(flameshadowToken.ObjectID); ok {
		t.Fatal("Flameshadow Conjuring's token was not exiled")
	}
	if _, ok := g.PermanentByID(copyOfToken.ObjectID); !ok {
		t.Fatal("the copy of Flameshadow Conjuring's token was exiled")
	}
}
