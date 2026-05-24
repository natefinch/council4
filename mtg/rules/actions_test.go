package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
)

func TestLegalActionsIncludesPlayableLandBeforePass(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	landID := addCardToHand(g, game.Player1, &game.CardDef{
		Name:  "Forest",
		Types: []game.CardType{game.TypeLand},
	})
	addCardToHand(g, game.Player1, &game.CardDef{
		Name:  "Sol Ring",
		Types: []game.CardType{game.TypeArtifact},
	})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	if len(legal) != 2 {
		t.Fatalf("legal actions = %d, want 2", len(legal))
	}
	if !actionsEqual(legal[0], action.PlayLand(landID)) {
		t.Fatalf("first legal action = %+v, want PlayLand(%v)", legal[0], landID)
	}
	if legal[1].Kind != action.ActionPass {
		t.Fatalf("last legal action kind = %v, want %v", legal[1].Kind, action.ActionPass)
	}
}

func TestLegalActionsIncludesCastSpellAfterPlayLandBeforePass(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	landID := addCardToHand(g, game.Player1, &game.CardDef{
		Name:  "Forest",
		Types: []game.CardType{game.TypeLand},
	})
	spellID := addCardToHand(g, game.Player1, greenCreature())
	addBasicLandPermanent(g, game.Player1, "Forest")
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	if len(legal) != 3 {
		t.Fatalf("legal actions = %d, want 3", len(legal))
	}
	if !actionsEqual(legal[0], action.PlayLand(landID)) {
		t.Fatalf("first legal action = %+v, want PlayLand(%v)", legal[0], landID)
	}
	if !actionsEqual(legal[1], action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatalf("second legal action = %+v, want CastSpell(%v)", legal[1], spellID)
	}
	if legal[2].Kind != action.ActionPass {
		t.Fatalf("last legal action kind = %v, want %v", legal[2].Kind, action.ActionPass)
	}
}

func TestLegalActionsDoesNotIncludePlayLandWhenUnavailable(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*game.Game)
	}{
		{
			name: "outside main phase",
			setup: func(g *game.Game) {
				g.Turn.Phase = game.PhaseBeginning
				g.Turn.Step = game.StepDraw
			},
		},
		{
			name: "land already played",
			setup: func(g *game.Game) {
				g.Turn.Phase = game.PhasePrecombatMain
				g.Turn.Step = game.StepNone
				g.Turn.LandsPlayedThisTurn = 1
			},
		},
		{
			name: "non-active player",
			setup: func(g *game.Game) {
				g.Turn.Phase = game.PhasePrecombatMain
				g.Turn.Step = game.StepNone
				g.Turn.PriorityPlayer = game.Player2
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			addCardToHand(g, game.Player1, &game.CardDef{
				Name:  "Forest",
				Types: []game.CardType{game.TypeLand},
			})
			tt.setup(g)

			legal := engine.legalActions(g, game.Player1)

			if len(legal) != 1 {
				t.Fatalf("legal actions = %d, want 1", len(legal))
			}
			if legal[0].Kind != action.ActionPass {
				t.Fatalf("legal action kind = %v, want %v", legal[0].Kind, action.ActionPass)
			}
		})
	}
}

func TestCreatureAndSorceryLegalOnlyAtSorcerySpeed(t *testing.T) {
	tests := []struct {
		name  string
		card  *game.CardDef
		setup func(*game.Game)
	}{
		{
			name: "creature outside main phase",
			card: greenCreature(),
			setup: func(g *game.Game) {
				g.Turn.Phase = game.PhaseBeginning
				g.Turn.Step = game.StepDraw
			},
		},
		{
			name: "sorcery while stack non-empty",
			card: greenSorcery(),
			setup: func(g *game.Game) {
				g.Turn.Phase = game.PhasePrecombatMain
				g.Turn.Step = game.StepNone
				g.Stack.Push(&game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell})
			},
		},
		{
			name: "creature non-active player",
			card: greenCreature(),
			setup: func(g *game.Game) {
				g.Turn.Phase = game.PhasePrecombatMain
				g.Turn.Step = game.StepNone
				g.Turn.PriorityPlayer = game.Player2
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			spellID := addCardToHand(g, game.Player1, tt.card)
			addBasicLandPermanent(g, game.Player1, "Forest")
			tt.setup(g)

			if containsAction(engine.legalActions(g, game.Player1), action.CastSpell(spellID, nil, 0, nil)) {
				t.Fatal("cast spell action was legal outside sorcery timing")
			}
		})
	}
}

func TestInstantLegalWhileStackNonEmpty(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	instantID := addCardToHand(g, game.Player2, greenInstant())
	addBasicLandPermanent(g, game.Player2, "Forest")
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player2
	g.Stack.Push(&game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell})

	if !containsAction(engine.legalActions(g, game.Player2), action.CastSpell(instantID, nil, 0, nil)) {
		t.Fatal("instant was not legal while another object was on the stack")
	}
}

func TestUnpayableSpellIsNotLegal(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, greenCreature())
	addBasicLandPermanent(g, game.Player1, "Mountain")
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if containsAction(engine.legalActions(g, game.Player1), action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("unpayable spell was legal")
	}
}

func TestTargetedSpellIsNotLegalBeforeTargetingSupport(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, &game.CardDef{
		Name:     "Giant Growth",
		ManaCost: greenCost(),
		Types:    []game.CardType{game.TypeInstant},
		Abilities: []game.AbilityDef{
			{
				Kind: game.SpellAbility,
				Targets: []game.TargetSpec{
					{MinTargets: 1, MaxTargets: 1, Constraint: "creature"},
				},
			},
		},
	})
	addBasicLandPermanent(g, game.Player1, "Forest")
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if containsAction(engine.legalActions(g, game.Player1), action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("targeted spell was legal before targeting support")
	}
}

func TestApplyActionPlayLandMovesCardToBattlefield(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	landID := addCardToHand(g, game.Player1, &game.CardDef{
		Name:  "Forest",
		Types: []game.CardType{game.TypeLand},
	})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.PlayLand(landID)) {
		t.Fatal("applyAction() = false, want true")
	}
	if g.Players[game.Player1].Hand.Contains(landID) {
		t.Fatal("land remained in hand")
	}
	if g.Turn.LandsPlayedThisTurn != 1 {
		t.Fatalf("lands played = %d, want 1", g.Turn.LandsPlayedThisTurn)
	}
	if len(g.Battlefield) != 1 {
		t.Fatalf("battlefield permanents = %d, want 1", len(g.Battlefield))
	}
	permanent := g.Battlefield[0]
	if permanent.CardInstanceID != landID {
		t.Fatalf("permanent card ID = %v, want %v", permanent.CardInstanceID, landID)
	}
	if permanent.Controller != game.Player1 {
		t.Fatalf("permanent controller = %v, want %v", permanent.Controller, game.Player1)
	}
	if !permanent.SummoningSick {
		t.Fatal("permanent summoning sick = false, want true")
	}
}

func TestApplyActionCastSpellPaysAndPushesStackObject(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, greenCreature())
	forest := addBasicLandPermanent(g, game.Player1, "Forest")
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("applyAction() = false, want true")
	}
	if g.Players[game.Player1].Hand.Contains(spellID) {
		t.Fatal("spell remained in hand")
	}
	if !forest.Tapped {
		t.Fatal("forest was not tapped to pay cost")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want 1", g.Stack.Size())
	}
	obj := g.Stack.Peek()
	if obj.SourceID != spellID {
		t.Fatalf("stack source ID = %v, want %v", obj.SourceID, spellID)
	}
	if obj.Controller != game.Player1 {
		t.Fatalf("stack controller = %v, want %v", obj.Controller, game.Player1)
	}
}

func TestApplyActionInvalidPlayLandDoesNotMutate(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	landID := addCardToHand(g, game.Player1, &game.CardDef{
		Name:  "Forest",
		Types: []game.CardType{game.TypeLand},
	})
	g.Turn.Phase = game.PhaseBeginning
	g.Turn.Step = game.StepDraw

	if engine.applyAction(g, game.Player1, action.PlayLand(landID)) {
		t.Fatal("applyAction() = true, want false")
	}
	if !g.Players[game.Player1].Hand.Contains(landID) {
		t.Fatal("land was removed from hand")
	}
	if len(g.Battlefield) != 0 {
		t.Fatalf("battlefield permanents = %d, want 0", len(g.Battlefield))
	}
	if g.Turn.LandsPlayedThisTurn != 0 {
		t.Fatalf("lands played = %d, want 0", g.Turn.LandsPlayedThisTurn)
	}
}

func TestApplyActionInvalidCastDoesNotMutate(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, greenCreature())
	forest := addBasicLandPermanent(g, game.Player1, "Forest")
	g.Turn.Phase = game.PhaseBeginning
	g.Turn.Step = game.StepDraw

	if engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("applyAction() = true, want false")
	}
	if !g.Players[game.Player1].Hand.Contains(spellID) {
		t.Fatal("spell was removed from hand")
	}
	if forest.Tapped {
		t.Fatal("forest was tapped by invalid cast")
	}
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want 0", g.Stack.Size())
	}
}

func addCardToHand(g *game.Game, playerID game.PlayerID, def *game.CardDef) id.ID {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID:    cardID,
		Def:   def,
		Owner: playerID,
	}
	g.Players[playerID].Hand.Add(cardID)
	return cardID
}

func greenCreature() *game.CardDef {
	return &game.CardDef{
		Name:      "Runeclaw Bear",
		ManaCost:  greenCost(),
		ManaValue: 1,
		Types:     []game.CardType{game.TypeCreature},
	}
}

func greenInstant() *game.CardDef {
	return &game.CardDef{
		Name:     "Giant Growth",
		ManaCost: greenCost(),
		Types:    []game.CardType{game.TypeInstant},
	}
}

func greenSorcery() *game.CardDef {
	return &game.CardDef{
		Name:     "Explore",
		ManaCost: greenCost(),
		Types:    []game.CardType{game.TypeSorcery},
	}
}

func greenCost() *mana.Cost {
	cost := mana.Cost{mana.ColoredMana(mana.Green)}
	return &cost
}
