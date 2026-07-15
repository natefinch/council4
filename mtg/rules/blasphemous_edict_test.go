package rules

import (
	"reflect"
	"slices"
	"testing"

	cardb "github.com/natefinch/council4/mtg/cards/b"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// blasphemousEdictEdict extracts the sacrifice edict primitive from the real
// generated Blasphemous Edict definition so the runtime tests drive the curated
// card, not a hand-written stand-in.
func blasphemousEdictEdict(t *testing.T) game.SacrificePermanents {
	t.Helper()
	def := cardb.BlasphemousEdict()
	if !def.SpellAbility.Exists {
		t.Fatal("Blasphemous Edict has no spell ability")
	}
	seq := def.SpellAbility.Val.Modes[0].Sequence
	if len(seq) != 1 {
		t.Fatalf("spell ability sequence length = %d, want 1", len(seq))
	}
	prim, ok := seq[0].Primitive.(game.SacrificePermanents)
	if !ok {
		t.Fatalf("spell ability primitive is %T, want game.SacrificePermanents", seq[0].Primitive)
	}
	return prim
}

// edictSacrificeIndices returns the choice indices [0, n) a player uses to
// sacrifice the first n of their eligible creatures.
func edictSacrificeIndices(n int) []int {
	indices := make([]int, n)
	for i := range indices {
		indices[i] = i
	}
	return indices
}

// creaturesControlled counts the creatures playerID still controls on the
// battlefield.
func creaturesControlled(g *game.Game, playerID game.PlayerID) int {
	count := 0
	for _, permanent := range g.Battlefield {
		if permanent.Controller == playerID && permanentHasType(g, permanent, types.Creature) {
			count++
		}
	}
	return count
}

// TestBlasphemousEdictRealCardDefShape locks the curated definition to both
// features issue #1865 added: the board-state-conditional {B} alternative cost
// and the each-player-sacrifices-thirteen edict. Sourcing the assertions from
// the registered definition proves the generated card — not a stand-in — carries
// the typed shapes the parser, compiler, lowering, and render produced.
func TestBlasphemousEdictRealCardDefShape(t *testing.T) {
	def := cardb.BlasphemousEdict()

	if got, want := def.ManaCost.Val, (cost.Mana{cost.O(3), cost.B, cost.B}); !def.ManaCost.Exists || !reflect.DeepEqual(got, want) {
		t.Fatalf("mana cost = %v (exists %t), want %v", got, def.ManaCost.Exists, want)
	}
	if !slices.Equal(def.Types, []types.Card{types.Sorcery}) {
		t.Fatalf("types = %v, want [Sorcery]", def.Types)
	}

	if got := len(def.AlternativeCosts); got != 1 {
		t.Fatalf("alternative costs length = %d, want 1", got)
	}
	wantAlt := cost.Alternative{
		Label:                  "Pay {B}",
		ManaCost:               opt.Val(cost.Mana{cost.B}),
		Condition:              cost.AlternativeConditionPermanentsOnBattlefield,
		ConditionCount:         13,
		ConditionPermanentType: types.Creature,
	}
	if got := def.AlternativeCosts[0]; !reflect.DeepEqual(got, wantAlt) {
		t.Fatalf("alternative cost = %+v, want %+v", got, wantAlt)
	}

	wantEdict := game.SacrificePermanents{
		Amount:      game.Fixed(13),
		PlayerGroup: game.AllPlayersReference(),
		Selection:   game.Selection{RequiredTypes: []types.Card{types.Creature}},
	}
	if got := blasphemousEdictEdict(t); !reflect.DeepEqual(got, wantEdict) {
		t.Fatalf("edict primitive = %+v, want %+v", got, wantEdict)
	}
}

// TestBlasphemousEdictEachPlayerSacrificesUpToThirteen resolves the real edict
// with every board-count band — zero, fewer than thirteen, exactly thirteen, and
// more than thirteen — proving each player (the controller included) sacrifices
// min(13, controlled) of their own creatures, keeps exactly the creatures they
// chose to spare, and never reaches across to another player's board. It also
// covers a token among the victims to prove tokens are ordinary edict fodder.
func TestBlasphemousEdictEachPlayerSacrificesUpToThirteen(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	engine := NewEngine(nil)

	// Player1 (the controller) controls fourteen creatures: one more than the
	// edict count, so they must choose which thirteen die and keep one.
	p1 := make([]*game.Permanent, 14)
	for i := range p1 {
		p1[i] = addCreaturePermanent(g, game.Player1)
	}
	// Player2 controls exactly thirteen: all die with no choice offered.
	p2 := make([]*game.Permanent, 13)
	for i := range p2 {
		p2[i] = addCreaturePermanent(g, game.Player2)
	}
	// Player3 controls five (one a token): fewer than thirteen, so all die.
	p3 := make([]*game.Permanent, 5)
	for i := range p3 {
		if i == 2 {
			p3[i] = addTokenCreaturePermanent(g, game.Player3, "Zombie")
			continue
		}
		p3[i] = addCreaturePermanent(g, game.Player3)
	}
	// Player4 controls none: unaffected.

	addEffectSpellToStack(g, game.Player1, blasphemousEdictEdict(t), nil)
	agents := [game.NumPlayers]PlayerAgent{
		// Sacrifice the first thirteen, sparing the last creature (p1[13]).
		game.Player1: &sacrificeChoiceAgent{t: t, g: g, choice: edictSacrificeIndices(13)},
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if _, ok := permanentByObjectID(g, p1[13].ObjectID); !ok {
		t.Error("controller's spared creature left the battlefield")
	}
	for i, permanent := range p1[:13] {
		if _, ok := permanentByObjectID(g, permanent.ObjectID); ok {
			t.Errorf("controller's chosen victim p1[%d] survived", i)
		}
	}
	if got := creaturesControlled(g, game.Player1); got != 1 {
		t.Errorf("controller kept %d creatures, want 1", got)
	}
	if got := creaturesControlled(g, game.Player2); got != 0 {
		t.Errorf("player with exactly thirteen kept %d creatures, want 0", got)
	}
	if got := creaturesControlled(g, game.Player3); got != 0 {
		t.Errorf("player with fewer than thirteen kept %d creatures, want 0", got)
	}
	if _, ok := permanentByObjectID(g, p3[2].ObjectID); ok {
		t.Error("sacrificed token creature remained on the battlefield")
	}
}

// TestBlasphemousEdictGathersChoicesInAPNAPOrderBeforeSimultaneousSacrifice
// proves that when several players must each choose thirteen victims, choices are
// gathered in APNAP order and every chosen creature stays on the battlefield
// until all players have chosen, so the sacrifices happen simultaneously (deaths,
// triggers, and last-known information see one shared board). Each player only
// reaches their own creatures, so identical choice indices target different
// boards with no cross-player leakage.
func TestBlasphemousEdictGathersChoicesInAPNAPOrderBeforeSimultaneousSacrifice(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player2
	engine := NewEngine(nil)

	// Player1 controls no creatures; Players 2, 3, and 4 each control fourteen so
	// every one of them must make a choice, letting us observe the order.
	boards := map[game.PlayerID][]*game.Permanent{}
	for _, playerID := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		board := make([]*game.Permanent, 14)
		for i := range board {
			board[i] = addCreaturePermanent(g, playerID)
		}
		boards[playerID] = board
	}

	addEffectSpellToStack(g, game.Player1, blasphemousEdictEdict(t), nil)
	var choiceOrder []game.PlayerID
	// Player2 is active and chooses first; its first victim must still be present
	// when the later players choose, proving choices precede any sacrifice.
	firstVictim := boards[game.Player2][0].ObjectID
	agents := [game.NumPlayers]PlayerAgent{
		game.Player2: &sacrificeChoiceAgent{t: t, g: g, choice: edictSacrificeIndices(13), order: &choiceOrder},
		game.Player3: &sacrificeChoiceAgent{t: t, g: g, choice: edictSacrificeIndices(13), order: &choiceOrder, mustRemainForChoice: firstVictim},
		game.Player4: &sacrificeChoiceAgent{t: t, g: g, choice: edictSacrificeIndices(13), order: &choiceOrder, mustRemainForChoice: firstVictim},
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if want := []game.PlayerID{game.Player2, game.Player3, game.Player4}; !slices.Equal(choiceOrder, want) {
		t.Fatalf("choice order = %v, want APNAP order %v", choiceOrder, want)
	}
	for _, playerID := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		board := boards[playerID]
		if _, ok := permanentByObjectID(g, board[13].ObjectID); !ok {
			t.Errorf("player %v's spared creature left the battlefield", playerID)
		}
		for i, permanent := range board[:13] {
			if _, ok := permanentByObjectID(g, permanent.ObjectID); ok {
				t.Errorf("player %v's chosen victim [%d] survived", playerID, i)
			}
		}
		if got := creaturesControlled(g, playerID); got != 1 {
			t.Errorf("player %v kept %d creatures, want 1", playerID, got)
		}
	}
}

// TestBlasphemousEdictExcludesEliminatedPlayers proves the all-players edict skips
// players who have left the game: an eliminated player is never asked to choose
// and keeps every creature, while the players still in the game sacrifice theirs.
func TestBlasphemousEdictExcludesEliminatedPlayers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	engine := NewEngine(nil)

	addCreaturePermanent(g, game.Player1)
	addCreaturePermanent(g, game.Player2)
	// Player4 has left the game but still has creatures on the battlefield.
	g.Players[game.Player4].Eliminated = true
	addCreaturePermanent(g, game.Player4)
	addCreaturePermanent(g, game.Player4)

	addEffectSpellToStack(g, game.Player1, blasphemousEdictEdict(t), nil)
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if got := creaturesControlled(g, game.Player1); got != 0 {
		t.Errorf("active player kept %d creatures, want 0", got)
	}
	if got := creaturesControlled(g, game.Player2); got != 0 {
		t.Errorf("opponent kept %d creatures, want 0", got)
	}
	if got := creaturesControlled(g, game.Player4); got != 2 {
		t.Errorf("eliminated player kept %d creatures, want 2 (untouched)", got)
	}
}

// TestBlasphemousEdictRespectsCantBeSacrificed proves the edict honors a
// "can't be sacrificed" restriction: a protected creature is excluded from the
// victim pool entirely, so its controller sacrifices only their remaining
// creatures and the protected one survives.
func TestBlasphemousEdictRespectsCantBeSacrificed(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	engine := NewEngine(nil)

	// Player2 controls a shield that protects the creatures they control but do
	// not own, then a borrowed creature it protects plus three of their own.
	cantSacrificeControlNotOwnEnchantment(g, game.Player2)
	protected := controlNotOwnCreature(g, game.Player1, game.Player2)
	own := make([]*game.Permanent, 3)
	for i := range own {
		own[i] = addCreaturePermanent(g, game.Player2)
	}

	addEffectSpellToStack(g, game.Player1, blasphemousEdictEdict(t), nil)
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if _, ok := permanentByObjectID(g, protected.ObjectID); !ok {
		t.Error("creature that can't be sacrificed was sacrificed")
	}
	for i, permanent := range own {
		if _, ok := permanentByObjectID(g, permanent.ObjectID); ok {
			t.Errorf("player's own creature own[%d] survived the edict", i)
		}
	}
}
