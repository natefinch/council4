package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

func TestDrawEffectDrawsRequestedCards(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	sourceID := addEffectSpellToStack(g, game.Player1, game.Effect{
		Type:        game.EffectDraw,
		Amount:      2,
		TargetIndex: -1,
	}, nil)
	firstDraw := addCardToLibrary(g, game.Player1, &game.CardDef{Name: "First"})
	secondDraw := addCardToLibrary(g, game.Player1, &game.CardDef{Name: "Second"})
	log := TurnLog{}

	engine.resolveTopOfStack(g, &log)

	if !g.Players[game.Player1].Hand.Contains(firstDraw) {
		t.Fatal("first card was not drawn")
	}
	if !g.Players[game.Player1].Hand.Contains(secondDraw) {
		t.Fatal("second card was not drawn")
	}
	if len(log.Draws) != 2 {
		t.Fatalf("draw logs = %d, want 2", len(log.Draws))
	}
	if log.Resolves[0].SourceID != sourceID {
		t.Fatalf("resolve source = %v, want %v", log.Resolves[0].SourceID, sourceID)
	}
}

func TestGainLifeEffectIncreasesTargetLife(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addEffectSpellToStack(g, game.Player1, game.Effect{
		Type:        game.EffectGainLife,
		Amount:      3,
		TargetIndex: 0,
	}, []game.Target{game.PlayerTarget(game.Player2)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Players[game.Player2].Life != 43 {
		t.Fatalf("player 2 life = %d, want 43", g.Players[game.Player2].Life)
	}
}

func TestDamageAndLoseLifeEffectsCanEliminatePlayers(t *testing.T) {
	tests := []struct {
		name       string
		effectType game.EffectType
	}{
		{name: "damage", effectType: game.EffectDamage},
		{name: "lose life", effectType: game.EffectLoseLife},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			g.Players[game.Player2].Life = 3
			addEffectSpellToStack(g, game.Player1, game.Effect{
				Type:        tt.effectType,
				Amount:      3,
				TargetIndex: 0,
			}, []game.Target{game.PlayerTarget(game.Player2)})

			engine.resolveTopOfStack(g, &TurnLog{})
			losses := engine.applyStateBasedActions(g)

			if len(losses) != 1 {
				t.Fatalf("losses = %d, want 1", len(losses))
			}
			if losses[0].Player != game.Player2 {
				t.Fatalf("loss player = %v, want %v", losses[0].Player, game.Player2)
			}
			if !g.Players[game.Player2].Eliminated {
				t.Fatal("player 2 was not eliminated")
			}
		})
	}
}

func TestFailedDrawEffectLogsAndEliminatesPlayer(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addEffectSpellToStack(g, game.Player1, game.Effect{
		Type:        game.EffectDraw,
		Amount:      1,
		TargetIndex: -1,
	}, nil)
	log := TurnLog{}

	engine.resolveTopOfStack(g, &log)
	losses := engine.applyStateBasedActions(g)
	log.Losses = append(log.Losses, losses...)

	if len(log.Draws) != 1 {
		t.Fatalf("draw logs = %d, want 1", len(log.Draws))
	}
	if !log.Draws[0].Failed {
		t.Fatal("draw log did not record failed draw")
	}
	if len(log.Losses) != 1 {
		t.Fatalf("loss logs = %d, want 1", len(log.Losses))
	}
	if log.Losses[0].Player != game.Player1 || log.Losses[0].Reason != LossReasonEmptyLibraryDraw {
		t.Fatalf("loss log = %+v, want player %v reason %q", log.Losses[0], game.Player1, LossReasonEmptyLibraryDraw)
	}
	if !g.Players[game.Player1].Eliminated {
		t.Fatal("player 1 was not eliminated")
	}
}

func addEffectSpellToStack(g *game.Game, controller game.PlayerID, effect game.Effect, targets []game.Target) id.ID {
	sourceID := g.IDGen.Next()
	g.CardInstances[sourceID] = &game.CardInstance{
		ID: sourceID,
		Def: &game.CardDef{
			Name:  "Effect Spell",
			Types: []game.CardType{game.TypeSorcery},
			Abilities: []game.AbilityDef{
				{
					Kind:    game.SpellAbility,
					Effects: []game.Effect{effect},
				},
			},
		},
		Owner: controller,
	}
	g.Stack.Push(&game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   sourceID,
		Controller: controller,
		Targets:    targets,
	})
	return sourceID
}
