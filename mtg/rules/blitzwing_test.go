package rules

import (
	"math/rand/v2"
	"testing"

	cards "github.com/natefinch/council4/mtg/cards/b"
	"github.com/natefinch/council4/mtg/game"
)

// newBlitzwingFront puts the real Blitzwing, Cruel Tormentor card onto the
// controller's battlefield as its front face so its "At the beginning of your
// end step, target opponent loses life equal to the life that player lost this
// turn. If no life is lost this way, convert Blitzwing." trigger runs through
// the real resolution path.
func newBlitzwingFront(g *game.Game, controller game.PlayerID) *game.Permanent {
	permanent := addCombatPermanent(g, controller, cards.BlitzwingCruelTormentor())
	permanent.Face = game.FaceFront
	return permanent
}

// resolveBlitzwingEndStep resolves the real front-face end-step ability against
// the given target opponent, driving the actual instruction sequence (the
// target-scoped life loss followed by the loss-gated convert).
func resolveBlitzwingEndStep(g *game.Game, engine *Engine, permanent *game.Permanent, controller, target game.PlayerID) {
	content := cards.BlitzwingCruelTormentor().TriggeredAbilities[0].Content
	obj := &game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackTriggeredAbility,
		SourceID:     permanent.ObjectID,
		SourceCardID: permanent.CardInstanceID,
		Face:         game.FaceFront,
		Controller:   controller,
		Targets:      []game.Target{game.PlayerTarget(target)},
	}
	engine.resolveAbilityContentWithChoices(g, obj, content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
}

// TestBlitzwingEndStepDrainScalesWithTargetLifeLost proves gap #1 on the real
// card: the front-face end-step trigger drains the TARGET opponent for exactly
// the life that same opponent lost this turn (the dynamic amount reads the
// target player, not Blitzwing's controller). Because life is lost this way,
// gap #2's loss gate is satisfied and Blitzwing does not convert.
func TestBlitzwingEndStepDrainScalesWithTargetLifeLost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	blitzwing := newBlitzwingFront(g, game.Player1)

	// The controller loses life this turn to prove the amount is NOT scaled by
	// the controller: only the target opponent's own loss counts.
	loseLife(g, game.Player1, 9)
	loseLife(g, game.Player2, 4)
	if g.Players[game.Player2].Life != 36 {
		t.Fatalf("setup: target life = %d, want 36 after losing 4", g.Players[game.Player2].Life)
	}

	resolveBlitzwingEndStep(g, engine, blitzwing, game.Player1, game.Player2)

	if g.Players[game.Player2].Life != 32 {
		t.Fatalf("target life = %d, want 32 (lost another 4 = its own life lost this turn)", g.Players[game.Player2].Life)
	}
	if blitzwing.Face != game.FaceFront || blitzwing.Transformed {
		t.Fatalf("Blitzwing face/transformed = %v/%v, want front/false (life was lost, no convert)", blitzwing.Face, blitzwing.Transformed)
	}
}

// TestBlitzwingEndStepConvertsWhenNoLifeLost proves gap #2 on the real card:
// when the target opponent lost no life this turn, the drain removes zero life
// and the "If no life is lost this way, convert Blitzwing." gate fires, flipping
// the card to its back face (Blitzwing, Adaptive Assailant).
func TestBlitzwingEndStepConvertsWhenNoLifeLost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	blitzwing := newBlitzwingFront(g, game.Player1)

	resolveBlitzwingEndStep(g, engine, blitzwing, game.Player1, game.Player2)

	if g.Players[game.Player2].Life != 40 {
		t.Fatalf("target life = %d, want 40 (lost no life this turn, drains zero)", g.Players[game.Player2].Life)
	}
	if blitzwing.Face != game.FaceBack || !blitzwing.Transformed {
		t.Fatalf("Blitzwing face/transformed = %v/%v, want back/true (no life lost, converts)", blitzwing.Face, blitzwing.Transformed)
	}
}

// grantedBlitzwingCombatKeyword drives the real back-face beginning-of-combat
// trigger once with the supplied random source and returns the keyword granted
// to Blitzwing, asserting exactly one of {flying, indestructible} was granted
// until end of turn.
func grantedBlitzwingCombatKeyword(t *testing.T, rng *rand.Rand) game.Keyword {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(rng)
	g.Turn.ActivePlayer = game.Player1
	blitzwing := addCombatPermanent(g, game.Player1, cards.BlitzwingCruelTormentor())
	blitzwing.Face = game.FaceBack
	blitzwing.Transformed = true

	emitEvent(g, beginningStepEvent(game.Player1, game.StepBeginningOfCombat))
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("Blitzwing beginning-of-combat random-keyword trigger was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	flying := hasKeyword(g, blitzwing, game.Flying)
	indestructible := hasKeyword(g, blitzwing, game.Indestructible)
	if flying == indestructible {
		t.Fatalf("granted keywords flying=%v indestructible=%v, want exactly one", flying, indestructible)
	}
	if flying {
		return game.Flying
	}
	return game.Indestructible
}

// TestBlitzwingCombatGrantsOneRandomKeyword proves gap #3 on the real card: the
// back-face "At the beginning of combat on your turn, choose flying or
// indestructible at random. Blitzwing gains that ability until end of turn."
// trigger selects a single keyword through the engine's random source and grants
// it to Blitzwing.
func TestBlitzwingCombatGrantsOneRandomKeyword(t *testing.T) {
	_ = grantedBlitzwingCombatKeyword(t, rand.New(rand.NewPCG(1, 2)))
}

// TestBlitzwingCombatReachesBothKeywords proves the random selection is not
// pinned to one keyword: across several seeds both flying and indestructible are
// chosen at least once, so the at-random grant draws from both keyword modes.
func TestBlitzwingCombatReachesBothKeywords(t *testing.T) {
	seen := map[game.Keyword]bool{}
	for seed := range uint64(40) {
		seen[grantedBlitzwingCombatKeyword(t, rand.New(rand.NewPCG(seed, seed+1)))] = true
	}
	for _, keyword := range []game.Keyword{game.Flying, game.Indestructible} {
		if !seen[keyword] {
			t.Fatalf("random keyword selection never chose %v across seeds", keyword)
		}
	}
}
