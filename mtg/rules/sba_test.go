package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
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

	changed, deaths := engine.checkPermanentStateBasedActions(g, newPassBatchID(g))

	if !changed {
		t.Fatal("checkPermanentStateBasedActions() = false, want true")
	}
	if _, ok := permanentByObjectID(g, creature.ObjectID); ok {
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

	changed, deaths := engine.checkPermanentStateBasedActions(g, newPassBatchID(g))

	if changed {
		t.Fatalf("checkPermanentStateBasedActions() = true, deaths %+v; want pumped 3 toughness creature to survive 2 damage", deaths)
	}
	pumped.MarkedDamage = 3

	changed, deaths = engine.checkPermanentStateBasedActions(g, newPassBatchID(g))

	if !changed {
		t.Fatal("checkPermanentStateBasedActions() = false, want true after lethal counter-adjusted damage")
	}
	if _, ok := permanentByObjectID(g, pumped.ObjectID); ok {
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

	changed, deaths := engine.checkPermanentStateBasedActions(g, newPassBatchID(g))

	if !changed {
		t.Fatal("checkPermanentStateBasedActions() = false, want true")
	}
	if _, ok := permanentByObjectID(g, creature.ObjectID); ok {
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

	changed, deaths := engine.checkPermanentStateBasedActions(g, newPassBatchID(g))

	if changed {
		t.Fatalf("checkPermanentStateBasedActions() = true, deaths %+v; want indestructible creature to survive", deaths)
	}
	if _, ok := permanentByObjectID(g, creature.ObjectID); !ok {
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

	changed, deaths := engine.checkPermanentStateBasedActions(g, newPassBatchID(g))

	if changed {
		t.Fatalf("checkPermanentStateBasedActions() = true, deaths %+v; want indestructible creature to survive deathtouch damage", deaths)
	}
	if _, ok := permanentByObjectID(g, creature.ObjectID); !ok {
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
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Indestructible Zero Toughness",
		Types:           []types.Card{types.Creature},
		Power:           opt.Val(zero),
		Toughness:       opt.Val(zero),
		StaticAbilities: []game.StaticAbility{game.IndestructibleStaticBody}},
	})

	changed, deaths := engine.checkPermanentStateBasedActions(g, newPassBatchID(g))

	if !changed {
		t.Fatal("checkPermanentStateBasedActions() = false, want true")
	}
	if _, ok := permanentByObjectID(g, creature.ObjectID); ok {
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
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Zero Toughness",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(zero),
		Toughness: opt.Val(zero)},
	})

	changed, deaths := engine.checkPermanentStateBasedActions(g, newPassBatchID(g))

	if !changed {
		t.Fatal("checkPermanentStateBasedActions() = false, want true")
	}
	if _, ok := permanentByObjectID(g, creature.ObjectID); ok {
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

	changed, deaths := engine.checkPermanentStateBasedActions(g, newPassBatchID(g))

	if !changed {
		t.Fatal("checkPermanentStateBasedActions() = false, want true")
	}
	if _, ok := permanentByObjectID(g, creature.ObjectID); ok {
		t.Fatal("minus-countered zero-toughness creature remained on battlefield")
	}
	if len(deaths) != 1 || deaths[0].Reason != PermanentDeathReasonZeroToughness {
		t.Fatalf("death logs = %+v, want zero-toughness death", deaths)
	}
}

func TestCounterStateBasedActionsCancelPlusOneAndMinusOneCounters(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	creature.Counters.Add(counter.PlusOnePlusOne, 2)
	creature.Counters.Add(counter.MinusOneMinusOne, 1)

	if !checkCounterStateBasedActions(g) {
		t.Fatal("checkCounterStateBasedActions() = false, want true")
	}
	if got := creature.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("+1/+1 counters = %d, want 1", got)
	}
	if got := creature.Counters.Get(counter.MinusOneMinusOne); got != 0 {
		t.Fatalf("-1/-1 counters = %d, want 0", got)
	}
	if checkCounterStateBasedActions(g) {
		t.Fatal("checkCounterStateBasedActions() = true after counters already canceled, want false")
	}
}

func TestStateBasedActionsMoveThenRemoveLethalToken(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	pt := game.PT{Value: 1}
	token := &game.Permanent{
		ObjectID:     g.IDGen.Next(),
		Owner:        game.Player1,
		Controller:   game.Player1,
		MarkedDamage: 1,
		Token:        true,
		TokenDef: &game.CardDef{CardFace: game.CardFace{Name: "Token",
			Types:     []types.Card{types.Creature},
			Power:     opt.Val(pt),
			Toughness: opt.Val(pt)},
		},
	}
	g.Battlefield = append(g.Battlefield, token)

	_, deaths := engine.applyStateBasedActionsWithDeaths(g)

	if _, ok := permanentByObjectID(g, token.ObjectID); ok {
		t.Fatal("lethally damaged token remained on battlefield")
	}
	if g.Players[game.Player1].Graveyard.Size() != 0 {
		t.Fatalf("graveyard size = %d, want 0 for token death", g.Players[game.Player1].Graveyard.Size())
	}
	if len(deaths) != 1 || deaths[0].Permanent != token.ObjectID {
		t.Fatalf("death logs = %+v, want token death", deaths)
	}
	if deaths[0].TokenName != "Token" {
		t.Fatalf("death token name = %q, want Token", deaths[0].TokenName)
	}
}

func TestLegendaryRuleKeepsOldestPermanentPerControllerAndName(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	first := addLegendaryPermanent(g, game.Player1, "Godo")
	second := addLegendaryPermanent(g, game.Player1, "Godo")
	otherController := addLegendaryPermanent(g, game.Player2, "Godo")
	differentName := addLegendaryPermanent(g, game.Player1, "Sisay")

	changed, deaths := checkLegendaryRuleStateBasedActions(g, newPassBatchID(g))

	if g.InStaticSourceFrame() {
		t.Fatal("legendary-rule scan leaked a static-source frame")
	}
	if !changed {
		t.Fatal("checkLegendaryRuleStateBasedActions() = false, want true")
	}
	if _, ok := permanentByObjectID(g, first.ObjectID); !ok {
		t.Fatal("oldest legendary permanent should remain on battlefield")
	}
	if _, ok := permanentByObjectID(g, second.ObjectID); ok {
		t.Fatal("newer duplicate legendary permanent remained on battlefield")
	}
	if _, ok := permanentByObjectID(g, otherController.ObjectID); !ok {
		t.Fatal("same-name legendary permanent controlled by another player should remain")
	}
	if _, ok := permanentByObjectID(g, differentName.ObjectID); !ok {
		t.Fatal("different-name legendary permanent should remain")
	}
	if !g.Players[game.Player1].Graveyard.Contains(second.CardInstanceID) {
		t.Fatal("legendary rule permanent did not move to owner's graveyard")
	}
	if len(deaths) != 1 || deaths[0].Permanent != second.ObjectID || deaths[0].Reason != PermanentDeathReasonLegendaryRule {
		t.Fatalf("death logs = %+v, want newer duplicate legendary rule death", deaths)
	}
}

func TestStateBasedActionsConvergeAfterLegendaryRuleDetachesAura(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	keep := addLegendaryPermanent(g, game.Player1, "Godo")
	duplicate := addLegendaryPermanent(g, game.Player1, "Godo")
	aura := addPermanentForSBA(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Test Aura",
		Types:    []types.Card{types.Enchantment},
		Subtypes: []types.Sub{types.Aura}},
	})
	attachPermanent(g, aura, duplicate)

	_, deaths := engine.applyStateBasedActionsWithDeaths(g)

	if _, ok := permanentByObjectID(g, keep.ObjectID); !ok {
		t.Fatal("oldest legendary permanent should remain on battlefield")
	}
	if _, ok := permanentByObjectID(g, duplicate.ObjectID); ok {
		t.Fatal("duplicate legendary permanent remained on battlefield")
	}
	if _, ok := permanentByObjectID(g, aura.ObjectID); ok {
		t.Fatal("aura detached by legendary rule should be put into graveyard by next SBA pass")
	}
	if !g.Players[game.Player1].Graveyard.Contains(duplicate.CardInstanceID) {
		t.Fatal("duplicate legendary permanent did not move to graveyard")
	}
	if !g.Players[game.Player1].Graveyard.Contains(aura.CardInstanceID) {
		t.Fatal("illegal detached aura did not move to graveyard")
	}
	if !deathReasonFound(deaths, duplicate.ObjectID, PermanentDeathReasonLegendaryRule) {
		t.Fatalf("death logs = %+v, want duplicate legendary rule death", deaths)
	}
	if !deathReasonFound(deaths, aura.ObjectID, PermanentDeathReasonIllegalAura) {
		t.Fatalf("death logs = %+v, want detached aura illegal aura death", deaths)
	}
	if g.InStaticSourceFrame() {
		t.Fatal("state-based actions leaked a static-source frame")
	}
}

func addLegendaryPermanent(g *game.Game, controller game.PlayerID, name string) *game.Permanent {
	return addPermanentForSBA(g, controller, &game.CardDef{CardFace: game.CardFace{Name: name,
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		Power:      opt.Val(game.PT{Value: 2}),
		Toughness:  opt.Val(game.PT{Value: 2})},
	})
}

func addPermanentForSBA(g *game.Game, controller game.PlayerID, def *game.CardDef) *game.Permanent {
	cardID := g.IDGen.Next()
	card := &game.CardInstance{
		ID:    cardID,
		Def:   def,
		Owner: controller,
	}
	g.CardInstances[cardID] = card
	permanent, ok := createCardPermanent(g, card, controller, zone.Stack)
	if !ok {
		panic("permanent for SBA was not created")
	}
	return permanent
}

func deathReasonFound(deaths []PermanentDeathLog, objectID id.ID, reason PermanentDeathReason) bool {
	for _, death := range deaths {
		if death.Permanent == objectID && death.Reason == reason {
			return true
		}
	}
	return false
}
