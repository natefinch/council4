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

func TestMillScryAndSurveilLibraryEffectsUseDeterministicFallback(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	top := addCardToLibrary(g, game.Player1, &game.CardDef{Name: "Top"})
	second := addCardToLibrary(g, game.Player1, &game.CardDef{Name: "Second"})
	third := addCardToLibrary(g, game.Player1, &game.CardDef{Name: "Third"})
	addEffectSpellToStack(g, game.Player1, game.Effect{Type: game.EffectScry, Amount: 2, TargetIndex: -1}, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Library.All(); len(got) < 3 || got[0] != third || got[1] != second || got[2] != top {
		t.Fatalf("library after scry = %+v, want deterministic keep-top order", got)
	}

	addEffectSpellToStack(g, game.Player1, game.Effect{Type: game.EffectSurveil, Amount: 2, TargetIndex: -1}, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Library.All(); len(got) < 3 || got[0] != third || got[1] != second || got[2] != top {
		t.Fatalf("library after surveil = %+v, want deterministic keep-top order", got)
	}

	addEffectSpellToStack(g, game.Player1, game.Effect{Type: game.EffectMill, Amount: 2, TargetIndex: -1}, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player1].Graveyard.Contains(third) || !g.Players[game.Player1].Graveyard.Contains(second) {
		t.Fatal("mill did not move top two cards to graveyard")
	}
	if got := g.Players[game.Player1].Library.All(); len(got) != 1 || got[0] != top {
		t.Fatalf("library after mill = %+v, want only original bottom card", got)
	}
}

func TestScryAndSurveilUseChoiceAgent(t *testing.T) {
	t.Run("scry bottom", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		bottom := addCardToLibrary(g, game.Player1, &game.CardDef{Name: "Bottom"})
		top := addCardToLibrary(g, game.Player1, &game.CardDef{Name: "Top"})
		addEffectSpellToStack(g, game.Player1, game.Effect{Type: game.EffectScry, Amount: 1, TargetIndex: -1}, nil)
		log := TurnLog{}
		agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}}}

		engine.resolveTopOfStackWithChoices(g, agents, &log)

		if got := g.Players[game.Player1].Library.All(); len(got) != 2 || got[0] != bottom || got[1] != top {
			t.Fatalf("library after scry = %+v, want chosen card on bottom", got)
		}
		if len(log.Choices) != 1 || log.Choices[0].Request.Kind != game.ChoiceScry || log.Choices[0].UsedFallback {
			t.Fatalf("choices = %+v, want non-fallback scry choice", log.Choices)
		}
	})
	t.Run("surveil graveyard", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		top := addCardToLibrary(g, game.Player1, &game.CardDef{Name: "Top"})
		addEffectSpellToStack(g, game.Player1, game.Effect{Type: game.EffectSurveil, Amount: 1, TargetIndex: -1}, nil)
		log := TurnLog{}
		agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}}}

		engine.resolveTopOfStackWithChoices(g, agents, &log)

		if g.Players[game.Player1].Library.Contains(top) || !g.Players[game.Player1].Graveyard.Contains(top) {
			t.Fatal("surveil choice did not move card to graveyard")
		}
		if len(log.Choices) != 1 || log.Choices[0].Request.Kind != game.ChoiceSurveil || log.Choices[0].UsedFallback {
			t.Fatalf("choices = %+v, want non-fallback surveil choice", log.Choices)
		}
	})
}

func TestDestroyEffectMovesPermanentToGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCreaturePermanent(g, game.Player2)
	addEffectSpellToStack(g, game.Player1, game.Effect{Type: game.EffectDestroy, TargetIndex: 0}, []game.Target{game.PermanentTarget(target.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if permanentByObjectID(g, target.ObjectID) != nil {
		t.Fatal("destroyed permanent remained on battlefield")
	}
	if !g.Players[game.Player2].Graveyard.Contains(target.CardInstanceID) {
		t.Fatal("destroyed card was not in owner's graveyard")
	}
}

func TestExileAndBounceEffectsMovePermanentsToOwnerZones(t *testing.T) {
	tests := []struct {
		name        string
		effectType  game.EffectType
		destination *game.Zone
	}{
		{name: "exile", effectType: game.EffectExile, destination: nil},
		{name: "bounce", effectType: game.EffectBounce, destination: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			target := addCreaturePermanent(g, game.Player2)
			addEffectSpellToStack(g, game.Player1, game.Effect{Type: tt.effectType, TargetIndex: 0}, []game.Target{game.PermanentTarget(target.ObjectID)})

			engine.resolveTopOfStack(g, &TurnLog{})

			if permanentByObjectID(g, target.ObjectID) != nil {
				t.Fatal("moved permanent remained on battlefield")
			}
			var zone *game.Zone
			switch tt.effectType {
			case game.EffectExile:
				zone = &g.Players[game.Player2].Exile
			case game.EffectBounce:
				zone = &g.Players[game.Player2].Hand
			}
			if zone == nil || !zone.Contains(target.CardInstanceID) {
				t.Fatalf("card was not moved to expected zone for %s", tt.name)
			}
		})
	}
}

func TestSacrificeEffectMovesControllerPermanentThroughGraveyardIgnoringIndestructible(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatCreaturePermanent(g, game.Player1, game.Indestructible)
	addEffectSpellToStack(g, game.Player1, game.Effect{Type: game.EffectSacrifice, TargetIndex: 0}, []game.Target{game.PermanentTarget(target.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if permanentByObjectID(g, target.ObjectID) != nil {
		t.Fatal("sacrificed permanent remained on battlefield")
	}
	if !g.Players[game.Player1].Graveyard.Contains(target.CardInstanceID) {
		t.Fatal("sacrificed permanent did not move to graveyard")
	}
}

func TestTapAndUntapEffectsChangeTappedState(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCreaturePermanent(g, game.Player2)
	addEffectSpellToStack(g, game.Player1, game.Effect{Type: game.EffectTap, TargetIndex: 0}, []game.Target{game.PermanentTarget(target.ObjectID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	if !target.Tapped {
		t.Fatal("tap effect did not tap permanent")
	}

	addEffectSpellToStack(g, game.Player1, game.Effect{Type: game.EffectUntap, TargetIndex: 0}, []game.Target{game.PermanentTarget(target.ObjectID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	if target.Tapped {
		t.Fatal("untap effect did not untap permanent")
	}
}

func TestDamageToPermanentEffectCanCauseLethalSBA(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	addEffectSpellToStack(g, game.Player1, game.Effect{Type: game.EffectDamage, Amount: 3, TargetIndex: 0}, []game.Target{game.PermanentTarget(target.ObjectID)})
	engine.resolveTopOfStack(g, &TurnLog{})

	_, deaths := engine.applyStateBasedActionsWithDeaths(g)

	if len(deaths) != 1 {
		t.Fatalf("deaths = %d, want 1", len(deaths))
	}
	if permanentByObjectID(g, target.ObjectID) != nil {
		t.Fatal("lethally damaged permanent remained on battlefield")
	}
}

func TestMassDestroyCreaturesUsesSnapshotAndRespectsIndestructible(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature1 := addCreaturePermanent(g, game.Player1)
	creature2 := addCreaturePermanent(g, game.Player2)
	indestructible := addCombatCreaturePermanent(g, game.Player3, game.Indestructible)
	artifact := addCombatPermanent(g, game.Player4, &game.CardDef{
		Name:  "Relic",
		Types: []game.CardType{game.TypeArtifact},
	})
	addEffectSpellToStack(g, game.Player1, game.Effect{
		Type:        game.EffectDestroy,
		TargetIndex: -1,
		Selector:    game.EffectSelectorAllCreatures,
	}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	if permanentByObjectID(g, creature1.ObjectID) != nil {
		t.Fatal("first creature survived mass destroy")
	}
	if permanentByObjectID(g, creature2.ObjectID) != nil {
		t.Fatal("second creature survived mass destroy")
	}
	if permanentByObjectID(g, indestructible.ObjectID) == nil {
		t.Fatal("indestructible creature did not survive mass destroy")
	}
	if permanentByObjectID(g, artifact.ObjectID) == nil {
		t.Fatal("noncreature artifact did not survive mass destroy")
	}
}

func TestMassDestroyNonlandPermanentsLeavesLands(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	land := addCombatPermanent(g, game.Player1, &game.CardDef{
		Name:  "Island",
		Types: []game.CardType{game.TypeLand},
	})
	artifact := addCombatPermanent(g, game.Player1, &game.CardDef{
		Name:  "Relic",
		Types: []game.CardType{game.TypeArtifact},
	})
	enchantment := addCombatPermanent(g, game.Player2, &game.CardDef{
		Name:  "Aura",
		Types: []game.CardType{game.TypeEnchantment},
	})
	addEffectSpellToStack(g, game.Player1, game.Effect{
		Type:        game.EffectDestroy,
		TargetIndex: -1,
		Selector:    game.EffectSelectorAllNonlandPermanents,
	}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	if permanentByObjectID(g, land.ObjectID) == nil {
		t.Fatal("land did not survive nonland permanent wipe")
	}
	if permanentByObjectID(g, artifact.ObjectID) != nil {
		t.Fatal("artifact survived nonland permanent wipe")
	}
	if permanentByObjectID(g, enchantment.ObjectID) != nil {
		t.Fatal("enchantment survived nonland permanent wipe")
	}
}

func TestMassDamageDeathsAreLoggedTogetherBySBA(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature1 := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	creature2 := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	artifact := addCombatPermanent(g, game.Player3, &game.CardDef{
		Name:  "Relic",
		Types: []game.CardType{game.TypeArtifact},
	})
	addEffectSpellToStack(g, game.Player1, game.Effect{
		Type:        game.EffectDamage,
		Amount:      3,
		TargetIndex: -1,
		Selector:    game.EffectSelectorAllCreatures,
	}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})
	_, deaths := engine.applyStateBasedActionsWithDeaths(g)

	if len(deaths) != 2 {
		t.Fatalf("deaths = %d, want 2", len(deaths))
	}
	if permanentByObjectID(g, creature1.ObjectID) != nil {
		t.Fatal("first damaged creature survived SBA")
	}
	if permanentByObjectID(g, creature2.ObjectID) != nil {
		t.Fatal("second damaged creature survived SBA")
	}
	if permanentByObjectID(g, artifact.ObjectID) == nil {
		t.Fatal("noncreature artifact was affected by creature mass damage")
	}
}

func TestTemporaryPTModifierChangesCombatDamageAndLethalThreshold(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 4)
	addEffectSpellToStack(g, game.Player1, game.Effect{
		Type:           game.EffectModifyPT,
		TargetIndex:    0,
		PowerDelta:     3,
		ToughnessDelta: 3,
		UntilEndOfTurn: true,
	}, []game.Target{game.PermanentTarget(creature.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: creature.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
		Blockers: []game.BlockDeclaration{
			{Blocker: blocker.ObjectID, Blocking: creature.ObjectID},
		},
	}
	engine.resolveCombatDamage(g, &TurnLog{})
	_, deaths := engine.applyStateBasedActionsWithDeaths(g)

	if blocker.MarkedDamage != 5 {
		t.Fatalf("blocker marked damage = %d, want 5", blocker.MarkedDamage)
	}
	if permanentByObjectID(g, blocker.ObjectID) != nil {
		t.Fatal("blocker survived pumped combat damage")
	}
	if permanentByObjectID(g, creature.ObjectID) == nil {
		t.Fatal("pumped creature died despite increased toughness")
	}
	if len(deaths) != 1 || deaths[0].Permanent != blocker.ObjectID {
		t.Fatalf("deaths = %+v, want blocker death only", deaths)
	}
}

func TestTemporaryPTModifiersStackDeterministically(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	for _, effect := range []game.Effect{
		{Type: game.EffectModifyPT, TargetIndex: 0, PowerDelta: 1, ToughnessDelta: 2, UntilEndOfTurn: true},
		{Type: game.EffectModifyPT, TargetIndex: 0, PowerDelta: -2, ToughnessDelta: -1, UntilEndOfTurn: true},
	} {
		addEffectSpellToStack(g, game.Player1, effect, []game.Target{game.PermanentTarget(creature.ObjectID)})
		engine.resolveTopOfStack(g, &TurnLog{})
	}

	if got := effectivePower(g, creature); got != 1 {
		t.Fatalf("effective power = %d, want 1", got)
	}
	if got, ok := effectiveToughness(g, creature); !ok || got != 3 {
		t.Fatalf("effective toughness = %d ok=%v, want 3 true", got, ok)
	}
}

func TestTemporaryPTModifierExpiresDuringCleanup(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	creature.TemporaryPowerModifier = 3
	creature.TemporaryToughnessModifier = 3

	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	if creature.TemporaryPowerModifier != 0 || creature.TemporaryToughnessModifier != 0 {
		t.Fatalf("temporary modifiers = +%d/+%d, want 0/0", creature.TemporaryPowerModifier, creature.TemporaryToughnessModifier)
	}
	if got := effectivePower(g, creature); got != 2 {
		t.Fatalf("effective power after cleanup = %d, want 2", got)
	}
}

func TestCreateTokenEffectCreatesTokenPermanent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	token := &game.CardDef{
		Name:      "Soldier Token",
		Types:     []game.CardType{game.TypeCreature},
		Power:     &game.PT{Value: 1},
		Toughness: &game.PT{Value: 1},
	}
	addEffectSpellToStack(g, game.Player1, game.Effect{Type: game.EffectCreateToken, Amount: 2, TargetIndex: -1, Token: token}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	var tokens []*game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.Token {
			tokens = append(tokens, permanent)
		}
	}
	if len(tokens) != 2 {
		t.Fatalf("tokens = %d, want 2", len(tokens))
	}
	for _, permanent := range tokens {
		if permanent.TokenDef != token {
			t.Fatalf("token def = %p, want %p", permanent.TokenDef, token)
		}
		if permanent.Controller != game.Player1 || permanent.Owner != game.Player1 {
			t.Fatalf("token owner/controller = %v/%v, want %v", permanent.Owner, permanent.Controller, game.Player1)
		}
		if !permanent.SummoningSick {
			t.Fatal("token did not enter summoning sick")
		}
	}
}

func TestTokenCanBlockTakeCombatDamageAndDie(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	pt := game.PT{Value: 2}
	token := createTokenPermanent(g, game.Player2, &game.CardDef{
		Name:      "Bear Token",
		Types:     []game.CardType{game.TypeCreature},
		Power:     &pt,
		Toughness: &pt,
	})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
		Blockers: []game.BlockDeclaration{
			{Blocker: token.ObjectID, Blocking: attacker.ObjectID},
		},
	}

	engine.resolveCombatDamage(g, &TurnLog{})
	_, deaths := engine.applyStateBasedActionsWithDeaths(g)

	if permanentByObjectID(g, token.ObjectID) != nil {
		t.Fatal("lethally damaged token remained on battlefield")
	}
	if g.Players[game.Player2].Graveyard.Contains(token.ObjectID) {
		t.Fatal("dead token did not cease to exist from graveyard")
	}
	if len(deaths) != 1 || deaths[0].Permanent != token.ObjectID || deaths[0].TokenName != "Bear Token" {
		t.Fatalf("death logs = %+v, want readable token death", deaths)
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
