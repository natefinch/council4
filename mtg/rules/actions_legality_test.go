package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestLegalActionsIncludesPlayableLandBeforePass(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	landID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Forest",
		Types: []types.Card{types.Land}},
	})
	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Unsupported Card"}})
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
	landID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Forest",
		Types: []types.Card{types.Land}},
	})
	spellID := addCardToHand(g, game.Player1, greenCreature())
	addBasicLandPermanent(g, game.Player1, types.Forest)
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
			addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Forest",
				Types: []types.Card{types.Land}},
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
			addBasicLandPermanent(g, game.Player1, types.Forest)
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
	addBasicLandPermanent(g, game.Player2, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player2
	g.Stack.Push(&game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell})

	if !containsAction(engine.legalActions(g, game.Player2), action.CastSpell(instantID, nil, 0, nil)) {
		t.Fatal("instant was not legal while another object was on the stack")
	}
}

func TestFlashCreatureLegalAtInstantSpeed(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	flashID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Ambush Viper",
		ManaCost:        greenCost(),
		Types:           []types.Card{types.Creature},
		StaticAbilities: []game.StaticAbility{game.FlashStaticBody}},
	})
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhaseBeginning
	g.Turn.Step = game.StepUpkeep
	g.Stack.Push(&game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell})

	if !containsAction(engine.legalActions(g, game.Player1), action.CastSpell(flashID, nil, 0, nil)) {
		t.Fatal("flash creature was not legal at instant speed")
	}
}

func TestUnpayableSpellIsNotLegal(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, greenCreature())
	addBasicLandPermanent(g, game.Player1, types.Mountain)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if containsAction(engine.legalActions(g, game.Player1), action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("unpayable spell was legal")
	}
}

func TestLegalActionsIncludesPayableXValues(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	manaCost := cost.Mana{cost.X, cost.G}
	spellID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Gelatinous Genesis",
		ManaCost: opt.Val(manaCost),
		Types:    []types.Card{types.Sorcery}},
	})
	addBasicLandPermanent(g, game.Player1, types.Forest)
	addBasicLandPermanent(g, game.Player1, types.Mountain)
	addBasicLandPermanent(g, game.Player1, types.Island)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)
	for _, xValue := range []int{0, 1, 2} {
		if !containsAction(legal, action.CastSpell(spellID, nil, xValue, nil)) {
			t.Fatalf("legal actions do not include X=%d cast: %+v", xValue, legal)
		}
	}
	if containsAction(legal, action.CastSpell(spellID, nil, 3, nil)) {
		t.Fatalf("legal actions include unpayable X=3 cast: %+v", legal)
	}
}

func TestLegalActionsIncludesModalSpellModeChoices(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, modalCharm())
	target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Silvercoat Lion",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2})},
	})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)
	if !containsAction(legal, action.CastSpell(spellID, nil, 0, []int{0})) {
		t.Fatalf("legal actions do not include untargeted modal choice: %+v", legal)
	}
	if !containsAction(legal, action.CastSpell(spellID, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, []int{1})) {
		t.Fatalf("legal actions do not include targeted modal choice: %+v", legal)
	}
	if containsAction(legal, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatalf("legal actions include modal spell without chosen mode: %+v", legal)
	}
}

func TestModalSpellSupportsChooseTwo(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, modalSpellWithModeRange(2, 2))
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	if !containsAction(legal, action.CastSpell(spellID, nil, 0, []int{0, 1})) {
		t.Fatal("legal actions did not include first choose-two combination")
	}
	if !containsAction(legal, action.CastSpell(spellID, nil, 0, []int{1, 2})) {
		t.Fatal("legal actions did not include second choose-two combination")
	}
	if containsAction(legal, action.CastSpell(spellID, nil, 0, []int{0})) {
		t.Fatal("legal actions included too few chosen modes")
	}
}

func TestModalSpellSupportsOneOrBothAndUpToOne(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	oneOrBothID := addCardToHand(g, game.Player1, modalSpellWithModeRange(1, 2))
	upToOneID := addCardToHand(g, game.Player1, modalSpellWithModeRange(0, 1))
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	if !containsAction(legal, action.CastSpell(oneOrBothID, nil, 0, []int{0, 1})) {
		t.Fatal("one-or-both modal spell did not include both modes")
	}
	if containsAction(legal, action.CastSpell(oneOrBothID, nil, 0, nil)) {
		t.Fatal("one-or-both modal spell included no modes")
	}
	if !containsAction(legal, action.CastSpell(upToOneID, nil, 0, nil)) {
		t.Fatal("choose-up-to-one modal spell did not include no modes")
	}
	if !containsAction(legal, action.CastSpell(upToOneID, nil, 0, []int{1})) {
		t.Fatal("choose-up-to-one modal spell did not include one mode")
	}
}

func TestModalSpellSupportsChooseThreeAndAllModes(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	chooseThreeID := addCardToHand(g, game.Player1, modalSpellWithModeRange(3, 3))
	allModesID := addCardToHand(g, game.Player1, modalSpellWithModeRange(3, 3))
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	if !containsAction(legal, action.CastSpell(chooseThreeID, nil, 0, []int{0, 1, 2})) {
		t.Fatal("choose-three modal spell did not include all three modes")
	}
	if containsAction(legal, action.CastSpell(chooseThreeID, nil, 0, []int{0, 1})) {
		t.Fatal("choose-three modal spell included too few modes")
	}
	if !containsAction(legal, action.CastSpell(allModesID, nil, 0, []int{0, 1, 2})) {
		t.Fatal("all-modes modal spell did not include all modes")
	}
}

func TestModalSpellDuplicateModesAreCanonicalized(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, modalSpellWithDuplicateModes(2, 2))
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	for _, modes := range [][]int{{0, 0}, {0, 1}, {1, 1}} {
		if !containsAction(legal, action.CastSpell(spellID, nil, 0, modes)) {
			t.Fatalf("legal actions did not include duplicate-mode choice %+v", modes)
		}
	}
	if containsAction(legal, action.CastSpell(spellID, nil, 0, []int{1, 0})) {
		t.Fatal("legal actions included non-canonical duplicate-mode permutation")
	}
}

func TestModalSpellResolvesChosenModeOnly(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, modalCharm())
	target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Silvercoat Lion",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2})},
	})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	act := action.CastSpell(spellID, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, []int{1})
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(modal spell) = false, want true")
	}
	engine.resolveTopOfStack(g, nil)
	if target.MarkedDamage != 2 {
		t.Fatalf("target damage = %d, want 2", target.MarkedDamage)
	}
	if got := g.Players[game.Player1].Life; got != 40 {
		t.Fatalf("controller life = %d, want 40 because unchosen mode did not resolve", got)
	}
}
