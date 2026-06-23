package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ringTemptsObj builds a controller-scoped triggered-ability stack object that
// resolves "The Ring tempts you." for the given player.
func ringTemptsObj(g *game.Game, controller game.PlayerID) *game.StackObject {
	source := addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:  "Tempter",
		Types: []types.Card{types.Enchantment},
	}})
	return &game.StackObject{
		Kind:         game.StackTriggeredAbility,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
		Controller:   controller,
	}
}

// TestRingTemptsGrantsEmblemDesignatesBearerAndAdvancesLevel covers the core of
// CR 701.51: the first tempting gives the controller the Ring emblem at level 1,
// designates the single creature they control as their Ring-bearer, and counts
// the tempting.
func TestRingTemptsGrantsEmblemDesignatesBearerAndAdvancesLevel(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	bearer := addCreaturePermanent(g, game.Player1)
	obj := ringTemptsObj(g, game.Player1)

	resolveInstruction(engine, g, obj, game.RingTempts{Player: game.ControllerReference()}, &TurnLog{})

	player := g.Players[game.Player1]
	if player.RingLevel != 1 {
		t.Fatalf("RingLevel = %d, want 1", player.RingLevel)
	}
	if player.RingTemptedCount != 1 {
		t.Fatalf("RingTemptedCount = %d, want 1", player.RingTemptedCount)
	}
	if player.RingBearerID != bearer.ObjectID {
		t.Fatalf("RingBearerID = %v, want the only controlled creature %v", player.RingBearerID, bearer.ObjectID)
	}
}

// TestRingTemptsAdvancesAndCapsAtLevelFour confirms each tempting advances the
// Ring one level and that level 4 is the ceiling (CR 701.51).
func TestRingTemptsAdvancesAndCapsAtLevelFour(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCreaturePermanent(g, game.Player1)
	obj := ringTemptsObj(g, game.Player1)

	for i := 1; i <= 6; i++ {
		resolveInstruction(engine, g, obj, game.RingTempts{Player: game.ControllerReference()}, &TurnLog{})
	}

	player := g.Players[game.Player1]
	if player.RingLevel != ringMaxLevel {
		t.Fatalf("RingLevel = %d, want %d (capped)", player.RingLevel, ringMaxLevel)
	}
	if player.RingTemptedCount != 6 {
		t.Fatalf("RingTemptedCount = %d, want 6", player.RingTemptedCount)
	}
}

// TestRingTemptsLetsControllerChooseBearer confirms the controller chooses
// which creature they control becomes the Ring-bearer when more than one is
// available.
func TestRingTemptsLetsControllerChooseBearer(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	_ = addCreaturePermanent(g, game.Player1)
	second := addCreaturePermanent(g, game.Player1)
	obj := ringTemptsObj(g, game.Player1)
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}},
	}

	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{
		Primitive: game.RingTempts{Player: game.ControllerReference()},
	}, agents, &TurnLog{})

	if got := g.Players[game.Player1].RingBearerID; got != second.ObjectID {
		t.Fatalf("RingBearerID = %v, want chosen creature %v", got, second.ObjectID)
	}
}

// TestRingTemptsWithNoCreaturesAdvancesWithoutBearer confirms a player who
// controls no creatures still advances the Ring but gets no Ring-bearer
// (CR 701.51c).
func TestRingTemptsWithNoCreaturesAdvancesWithoutBearer(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	obj := ringTemptsObj(g, game.Player1)

	resolveInstruction(engine, g, obj, game.RingTempts{Player: game.ControllerReference()}, &TurnLog{})

	player := g.Players[game.Player1]
	if player.RingLevel != 1 {
		t.Fatalf("RingLevel = %d, want 1", player.RingLevel)
	}
	if player.RingBearerID != 0 {
		t.Fatalf("RingBearerID = %v, want none", player.RingBearerID)
	}
}

// TestRingTemptsDesignatesTokenBearerByObjectID confirms a token creature can be
// the Ring-bearer: tokens have a zero CardInstanceID, so the designation tracks
// the permanent's ObjectID and the level-4 drain still fires for a token bearer.
func TestRingTemptsDesignatesTokenBearerByObjectID(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	token, ok := createTokenPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Soldier",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
	if !ok {
		t.Fatal("token was not created")
	}
	obj := ringTemptsObj(g, game.Player1)

	resolveInstruction(engine, g, obj, game.RingTempts{Player: game.ControllerReference()}, &TurnLog{})

	if got := g.Players[game.Player1].RingBearerID; got != token.ObjectID {
		t.Fatalf("RingBearerID = %v, want token ObjectID %v", got, token.ObjectID)
	}
	if !isRingBearer(g, token) {
		t.Fatal("token is not recognized as the Ring-bearer")
	}

	g.Players[game.Player1].RingLevel = ringMaxLevel
	before := g.Players[game.Player2].Life
	markPlayerCombatDamage(g, token, game.Player3, 2, &TurnLog{})
	if got := before - g.Players[game.Player2].Life; got != ringBearerLoseLifeOnDamage {
		t.Fatalf("opponent lost %d life from token Ring-bearer combat damage, want %d", got, ringBearerLoseLifeOnDamage)
	}
}

// TestRingBearerCombatDamageDrainsOpponentsAtLevelFour covers the Ring's fourth
// ability: when a level-4 Ring-bearer deals combat damage to a player, each
// opponent loses 3 life.
func TestRingBearerCombatDamageDrainsOpponentsAtLevelFour(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	bearer := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Ring-bearer",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
	g.Players[game.Player1].RingLevel = ringMaxLevel
	g.Players[game.Player1].RingBearerID = bearer.ObjectID
	before := [game.NumPlayers]int{}
	for i := range g.Players {
		before[i] = g.Players[i].Life
	}

	markPlayerCombatDamage(g, bearer, game.Player2, 2, &TurnLog{})

	for i := range g.Players {
		want := before[i]
		if game.PlayerID(i) != game.Player1 && !g.Players[i].Eliminated {
			want -= ringBearerLoseLifeOnDamage
		}
		// Player2 also took 2 combat damage on top of the ring drain.
		if game.PlayerID(i) == game.Player2 {
			want -= 2
		}
		if got := g.Players[i].Life; got != want {
			t.Fatalf("player %d life = %d, want %d", i, got, want)
		}
	}
}

// TestRingBearerCombatDamageNoDrainBelowLevelFour confirms opponents lose no
// life from Ring-bearer combat damage before the Ring reaches level 4.
func TestRingBearerCombatDamageNoDrainBelowLevelFour(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	bearer := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Ring-bearer",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
	g.Players[game.Player1].RingLevel = 3
	g.Players[game.Player1].RingBearerID = bearer.ObjectID
	thirdBefore := g.Players[game.Player3].Life

	markPlayerCombatDamage(g, bearer, game.Player2, 2, &TurnLog{})

	if got := g.Players[game.Player3].Life; got != thirdBefore {
		t.Fatalf("non-defending opponent life = %d, want unchanged %d", got, thirdBefore)
	}
}
