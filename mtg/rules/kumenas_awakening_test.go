package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// kumenasAwakeningUpkeepSequence mirrors the two-instruction upkeep body cardgen
// lowers Kumena's Awakening into: without the city's blessing every player draws
// one card, and with the city's blessing only the controller draws one card. The
// two draws are complementary gates on the controller's live HasCityBlessing
// flag, so exactly one resolves. City's blessing can never be lost (CR 702.131),
// so the branches are mutually exclusive at resolution.
func kumenasAwakeningUpkeepSequence() []game.Instruction {
	return []game.Instruction{
		{
			Primitive: game.Draw{
				Amount:      game.Fixed(1),
				PlayerGroup: game.AllPlayersReference(),
			},
			Condition: opt.Val(game.EffectCondition{
				Condition: opt.Val(game.Condition{
					Negate:                    true,
					ControllerHasCityBlessing: true,
				}),
			}),
		},
		{
			Primitive: game.Draw{
				Amount: game.Fixed(1),
				Player: game.ControllerReference(),
			},
			Condition: opt.Val(game.EffectCondition{
				Condition: opt.Val(game.Condition{
					ControllerHasCityBlessing: true,
				}),
			}),
		},
	}
}

func kumenasAwakeningSourceDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:            "Kumena's Awakening",
		Types:           []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{game.AscendStaticBody},
	}}
}

// resolveKumenaUpkeep resolves the trigger body for a stack object controlled by
// controller and returns each player's hand-size delta.
func resolveKumenaUpkeep(t *testing.T, g *game.Game, obj *game.StackObject) [game.NumPlayers]int {
	t.Helper()
	engine := NewEngine(nil)
	var before [game.NumPlayers]int
	for i := range game.NumPlayers {
		before[i] = g.Players[i].Hand.Size()
	}
	resolver := newEffectResolver(engine, g, obj, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	for _, instr := range kumenasAwakeningUpkeepSequence() {
		resolver.resolveInstruction(&instr)
	}
	var delta [game.NumPlayers]int
	for i := range game.NumPlayers {
		delta[i] = g.Players[i].Hand.Size() - before[i]
	}
	return delta
}

func stockAllLibraries(g *game.Game, count int) {
	for i := range game.NumPlayers {
		stockLibrary(g, game.PlayerID(i), count)
	}
}

// TestKumenasAwakeningWithoutBlessingEachPlayerDraws covers the base branch: with
// no city's blessing, every player draws exactly one card in APNAP order and the
// controller does not draw a second card.
func TestKumenasAwakeningWithoutBlessing(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	stockAllLibraries(g, 3)
	source := addCombatPermanent(g, game.Player1, kumenasAwakeningSourceDef())
	obj := triggeredObjFor(source)

	delta := resolveKumenaUpkeep(t, g, obj)
	for i := range game.NumPlayers {
		if delta[i] != 1 {
			t.Fatalf("player %d drew %d cards, want 1 (each player draws without blessing)", i, delta[i])
		}
	}
}

// TestKumenasAwakeningWithBlessingOnlyControllerDraws covers the replacement
// branch: with the city's blessing, only the controller draws a card and the
// opponents draw nothing.
func TestKumenasAwakeningWithBlessing(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	stockAllLibraries(g, 3)
	source := addCombatPermanent(g, game.Player1, kumenasAwakeningSourceDef())
	obj := triggeredObjFor(source)
	g.Players[game.Player1].HasCityBlessing = true

	delta := resolveKumenaUpkeep(t, g, obj)
	if delta[game.Player1] != 1 {
		t.Fatalf("controller drew %d cards, want 1 (only you draw with blessing)", delta[game.Player1])
	}
	for i := 1; i < game.NumPlayers; i++ {
		if delta[i] != 0 {
			t.Fatalf("opponent %d drew %d cards, want 0 (only controller draws with blessing)", i, delta[i])
		}
	}
}

// TestKumenasAwakeningControllerChangeChecksNewController proves the blessing
// gate is evaluated against the ability's current controller at resolution, not
// the printed owner: when a player without the blessing controls the trigger,
// every player draws even though a different player has the blessing.
func TestKumenasAwakeningControllerChangeChecksNewController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	stockAllLibraries(g, 3)
	source := addCombatPermanent(g, game.Player1, kumenasAwakeningSourceDef())
	// A different player has the blessing, but the ability is controlled by a
	// player who does not, so the base each-player branch resolves.
	g.Players[game.Player2].HasCityBlessing = true
	obj := triggeredObjFor(source)

	delta := resolveKumenaUpkeep(t, g, obj)
	for i := range game.NumPlayers {
		if delta[i] != 1 {
			t.Fatalf("player %d drew %d cards, want 1 (controller lacks blessing)", i, delta[i])
		}
	}

	// Now the controller itself has the blessing: only it draws.
	g.Players[game.Player1].HasCityBlessing = true
	delta = resolveKumenaUpkeep(t, g, obj)
	if delta[game.Player1] != 1 {
		t.Fatalf("controller drew %d cards, want 1 (blessed controller draws)", delta[game.Player1])
	}
	for i := 1; i < game.NumPlayers; i++ {
		if delta[i] != 0 {
			t.Fatalf("opponent %d drew %d cards, want 0 (blessed controller draws alone)", i, delta[i])
		}
	}
}

// TestKumenasAwakeningDeckoutDrawsWhatItCan proves an empty-library player still
// participates in the each-player draw: the draw is attempted (marking the player
// for the empty-library loss state-based action) while the other players draw
// normally in APNAP order.
func TestKumenasAwakeningDeckoutDrawsWhatItCan(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	// Player3 has no library; every other player can draw.
	stockLibrary(g, game.Player1, 3)
	stockLibrary(g, game.Player2, 3)
	stockLibrary(g, game.Player4, 3)
	source := addCombatPermanent(g, game.Player1, kumenasAwakeningSourceDef())
	obj := triggeredObjFor(source)

	delta := resolveKumenaUpkeep(t, g, obj)
	for _, id := range []game.PlayerID{game.Player1, game.Player2, game.Player4} {
		if delta[id] != 1 {
			t.Fatalf("player %d drew %d cards, want 1", id, delta[id])
		}
	}
	if delta[game.Player3] != 0 {
		t.Fatalf("empty-library player drew %d cards, want 0", delta[game.Player3])
	}
	if !g.FailedDraws[game.Player3] {
		t.Fatal("empty-library player was not marked for the draw-from-empty loss")
	}
}
