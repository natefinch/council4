package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
)

func TestCheckStateBasedActionsEliminatesPlayers(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*game.Player, *game.Game)
		wantReason LossReason
	}{
		{
			name:       "zero life",
			wantReason: LossReasonZeroLife,
			setup: func(p *game.Player, g *game.Game) {
				p.Life = 0
			},
		},
		{
			name:       "lethal poison",
			wantReason: LossReasonPoisonCounters,
			setup: func(p *game.Player, g *game.Game) {
				p.PoisonCounters = 10
			},
		},
		{
			name:       "lethal commander damage",
			wantReason: LossReasonCommanderDamage,
			setup: func(p *game.Player, g *game.Game) {
				p.CommanderDamage[id.ID(99)] = 21
			},
		},
		{
			name:       "failed draw",
			wantReason: LossReasonEmptyLibraryDraw,
			setup: func(p *game.Player, g *game.Game) {
				g.FailedDraws[p.ID] = true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			player := g.Players[game.Player1]
			tt.setup(player, g)

			changed, losses := engine.checkStateBasedActions(g)
			if !changed {
				t.Fatal("checkStateBasedActions() = false, want true")
			}
			if len(losses) != 1 {
				t.Fatalf("losses = %d, want 1", len(losses))
			}
			if losses[0].Player != player.ID {
				t.Fatalf("loss player = %v, want %v", losses[0].Player, player.ID)
			}
			if losses[0].Reason != tt.wantReason {
				t.Fatalf("loss reason = %q, want %q", losses[0].Reason, tt.wantReason)
			}
			if !player.Eliminated {
				t.Fatal("player was not marked eliminated")
			}
			if !g.TurnOrder.IsEliminated(player.ID) {
				t.Fatal("turn order was not marked eliminated")
			}
			if g.FailedDraws[player.ID] {
				t.Fatal("failed draw flag was not cleared")
			}
		})
	}
}

func TestCheckStateBasedActionsAlreadyEliminatedIsStable(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	if !engine.eliminatePlayer(g, game.Player1) {
		t.Fatal("first eliminatePlayer() = false, want true")
	}
	changed, losses := engine.checkStateBasedActions(g)
	if changed {
		t.Fatal("checkStateBasedActions() = true for stable eliminated player, want false")
	}
	if len(losses) != 0 {
		t.Fatalf("losses = %d, want 0", len(losses))
	}
}

func TestCheckStateBasedActionsClearsFailedDrawForAlreadyEliminatedPlayer(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	if !engine.eliminatePlayer(g, game.Player1) {
		t.Fatal("eliminatePlayer() = false, want true")
	}
	g.FailedDraws[game.Player1] = true

	changed, losses := engine.checkStateBasedActions(g)
	if changed {
		t.Fatal("checkStateBasedActions() = true for already eliminated player, want false")
	}
	if len(losses) != 0 {
		t.Fatalf("losses = %d, want 0", len(losses))
	}
	if g.FailedDraws[game.Player1] {
		t.Fatal("failed draw flag was not cleared")
	}
}

func TestApplyStateBasedActionsReturnsLosses(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Players[game.Player1].Life = 0

	losses := engine.applyStateBasedActions(g)

	if len(losses) != 1 {
		t.Fatalf("losses = %d, want 1", len(losses))
	}
	if losses[0].Reason != LossReasonZeroLife {
		t.Fatalf("loss reason = %q, want %q", losses[0].Reason, LossReasonZeroLife)
	}
}

func TestCheckPermanentStateBasedActionsDestroysCreatureWithLethalDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	creature.MarkedDamage = 2

	changed, deaths := engine.checkPermanentStateBasedActions(g)

	if !changed {
		t.Fatal("checkPermanentStateBasedActions() = false, want true")
	}
	if permanentByObjectID(g, creature.ObjectID) != nil {
		t.Fatal("creature with lethal damage remained on battlefield")
	}
	if !g.Players[game.Player1].Graveyard.Contains(creature.CardInstanceID) {
		t.Fatal("destroyed creature did not move to graveyard")
	}
	if len(deaths) != 1 || deaths[0].Permanent != creature.ObjectID || deaths[0].Reason != PermanentDeathReasonLethalDamage {
		t.Fatalf("death logs = %+v, want lethal damage death", deaths)
	}
}

func TestCheckPermanentStateBasedActionsUsesCounterAdjustedToughness(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	pumped := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	pumped.Counters.Add(counter.PlusOnePlusOne, 1)
	pumped.MarkedDamage = 2

	changed, deaths := engine.checkPermanentStateBasedActions(g)

	if changed {
		t.Fatalf("checkPermanentStateBasedActions() = true, deaths %+v; want pumped 3 toughness creature to survive 2 damage", deaths)
	}
	pumped.MarkedDamage = 3

	changed, deaths = engine.checkPermanentStateBasedActions(g)

	if !changed {
		t.Fatal("checkPermanentStateBasedActions() = false, want true after lethal counter-adjusted damage")
	}
	if permanentByObjectID(g, pumped.ObjectID) != nil {
		t.Fatal("creature with lethal counter-adjusted damage remained on battlefield")
	}
	if len(deaths) != 1 || deaths[0].Permanent != pumped.ObjectID || deaths[0].Reason != PermanentDeathReasonLethalDamage {
		t.Fatalf("death logs = %+v, want pumped lethal damage death", deaths)
	}
}

func TestCheckPermanentStateBasedActionsDestroysCreatureWithDeathtouchDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 5)
	creature.MarkedDamage = 1
	creature.MarkedDeathtouchDamage = true

	changed, deaths := engine.checkPermanentStateBasedActions(g)

	if !changed {
		t.Fatal("checkPermanentStateBasedActions() = false, want true")
	}
	if permanentByObjectID(g, creature.ObjectID) != nil {
		t.Fatal("creature with deathtouch damage remained on battlefield")
	}
	if len(deaths) != 1 || deaths[0].Permanent != creature.ObjectID || deaths[0].Reason != PermanentDeathReasonLethalDamage {
		t.Fatalf("death logs = %+v, want deathtouch lethal damage death", deaths)
	}
}

func TestCheckPermanentStateBasedActionsDoesNotDestroyIndestructibleWithLethalDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2, game.Indestructible)
	creature.MarkedDamage = 5

	changed, deaths := engine.checkPermanentStateBasedActions(g)

	if changed {
		t.Fatalf("checkPermanentStateBasedActions() = true, deaths %+v; want indestructible creature to survive", deaths)
	}
	if permanentByObjectID(g, creature.ObjectID) == nil {
		t.Fatal("indestructible creature with lethal damage left the battlefield")
	}
	if creature.MarkedDamage != 5 {
		t.Fatalf("marked damage = %d, want 5 until cleanup", creature.MarkedDamage)
	}
	if g.Players[game.Player1].Graveyard.Contains(creature.CardInstanceID) {
		t.Fatal("indestructible creature moved to graveyard")
	}
}

func TestCheckPermanentStateBasedActionsDoesNotDestroyIndestructibleWithDeathtouchDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 5, game.Indestructible)
	creature.MarkedDamage = 1
	creature.MarkedDeathtouchDamage = true

	changed, deaths := engine.checkPermanentStateBasedActions(g)

	if changed {
		t.Fatalf("checkPermanentStateBasedActions() = true, deaths %+v; want indestructible creature to survive deathtouch damage", deaths)
	}
	if permanentByObjectID(g, creature.ObjectID) == nil {
		t.Fatal("indestructible creature with deathtouch damage left the battlefield")
	}
	if creature.MarkedDamage != 1 || !creature.MarkedDeathtouchDamage {
		t.Fatalf("marked damage = %d deathtouch=%v, want retained marked damage until cleanup", creature.MarkedDamage, creature.MarkedDeathtouchDamage)
	}
}

func TestCheckPermanentStateBasedActionsDestroysIndestructibleZeroToughnessCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	zero := game.PT{Value: 0}
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{
		Name:      "Indestructible Zero Toughness",
		Types:     []game.CardType{game.TypeCreature},
		Power:     &zero,
		Toughness: &zero,
		Abilities: []game.AbilityDef{
			{
				Kind:     game.StaticAbility,
				Keywords: []game.Keyword{game.Indestructible},
			},
		},
	})

	changed, deaths := engine.checkPermanentStateBasedActions(g)

	if !changed {
		t.Fatal("checkPermanentStateBasedActions() = false, want true")
	}
	if permanentByObjectID(g, creature.ObjectID) != nil {
		t.Fatal("indestructible zero-toughness creature remained on battlefield")
	}
	if len(deaths) != 1 || deaths[0].Reason != PermanentDeathReasonZeroToughness {
		t.Fatalf("death logs = %+v, want zero-toughness death", deaths)
	}
}

func TestCheckPermanentStateBasedActionsDestroysZeroToughnessCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	zero := game.PT{Value: 0}
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{
		Name:      "Zero Toughness",
		Types:     []game.CardType{game.TypeCreature},
		Power:     &zero,
		Toughness: &zero,
	})

	changed, deaths := engine.checkPermanentStateBasedActions(g)

	if !changed {
		t.Fatal("checkPermanentStateBasedActions() = false, want true")
	}
	if permanentByObjectID(g, creature.ObjectID) != nil {
		t.Fatal("zero-toughness creature remained on battlefield")
	}
	if len(deaths) != 1 || deaths[0].Reason != PermanentDeathReasonZeroToughness {
		t.Fatalf("death logs = %+v, want zero-toughness death", deaths)
	}
}

func TestCheckPermanentStateBasedActionsUsesMinusCounterForZeroToughness(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 1)
	creature.Counters.Add(counter.MinusOneMinusOne, 1)

	changed, deaths := engine.checkPermanentStateBasedActions(g)

	if !changed {
		t.Fatal("checkPermanentStateBasedActions() = false, want true")
	}
	if permanentByObjectID(g, creature.ObjectID) != nil {
		t.Fatal("minus-countered zero-toughness creature remained on battlefield")
	}
	if len(deaths) != 1 || deaths[0].Reason != PermanentDeathReasonZeroToughness {
		t.Fatalf("death logs = %+v, want zero-toughness death", deaths)
	}
}

func TestCheckPermanentStateBasedActionsRemovesLethalTokenWithoutGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	pt := game.PT{Value: 1}
	token := &game.Permanent{
		ObjectID:     g.IDGen.Next(),
		Owner:        game.Player1,
		Controller:   game.Player1,
		MarkedDamage: 1,
		Token:        true,
		TokenDef: &game.CardDef{
			Name:      "Token",
			Types:     []game.CardType{game.TypeCreature},
			Power:     &pt,
			Toughness: &pt,
		},
	}
	g.Battlefield = append(g.Battlefield, token)

	changed, deaths := engine.checkPermanentStateBasedActions(g)

	if !changed {
		t.Fatal("checkPermanentStateBasedActions() = false, want true")
	}
	if permanentByObjectID(g, token.ObjectID) != nil {
		t.Fatal("lethally damaged token remained on battlefield")
	}
	if g.Players[game.Player1].Graveyard.Size() != 0 {
		t.Fatalf("graveyard size = %d, want 0 for token death", g.Players[game.Player1].Graveyard.Size())
	}
	if len(deaths) != 1 || deaths[0].Permanent != token.ObjectID {
		t.Fatalf("death logs = %+v, want token death", deaths)
	}
}
