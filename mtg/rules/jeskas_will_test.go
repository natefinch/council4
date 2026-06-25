package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestCommanderModalBonusIsChosenAtCastTime(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	targetPlayer := game.TargetPlayerReference(0)
	spell := &game.CardDef{CardFace: game.CardFace{
		Name:  "Commander Modal Spell",
		Types: []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.AbilityContent{
			Modes: []game.Mode{
				{
					Targets: []game.TargetSpec{{
						MinTargets: 1,
						MaxTargets: 1,
						Allow:      game.TargetAllowPlayer,
						Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),
					}},
					Sequence: []game.Instruction{{Primitive: game.AddMana{
						Amount: game.Dynamic(game.DynamicAmount{
							Kind:       game.DynamicAmountCountCardsInZone,
							Multiplier: 1,
							Player:     &targetPlayer,
							CardZone:   zone.Hand,
							Selection:  &game.Selection{},
						}),
						ManaColor: mana.R,
					}}},
				},
				{Sequence: []game.Instruction{{Primitive: game.ImpulseExile{
					Player:   game.ControllerReference(),
					Amount:   game.Fixed(3),
					Duration: game.DurationThisTurn,
				}}}},
			},
			MinModes: 1,
			MaxModes: 1,
			ModeChoiceBonus: game.ModeChoiceBonus{
				Condition:          game.ModeChoiceConditionControlsCommander,
				AdditionalMaxModes: 1,
			},
		}),
	}}
	spellID := addCardToHand(g, game.Player1, spell)
	for range 2 {
		addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Opponent Hand Card"}})
	}
	exiledIDs := []id.ID{
		addCardToLibrary(g, game.Player1, zeroCostImpulseSpell("First Exiled Spell")),
		addCardToLibrary(g, game.Player1, zeroCostImpulseSpell("Second Exiled Spell")),
		addCardToLibrary(g, game.Player1, zeroCostImpulseSpell("Third Exiled Spell")),
	}
	setMainPhasePriority(g, game.Player1)
	targets := []game.Target{game.PlayerTarget(game.Player2)}
	castBoth := action.CastSpell(spellID, targets, 0, []int{0, 1})

	legal := engine.legalActions(g, game.Player1)
	if actionsContain(legal, castBoth) {
		t.Fatal("both modes were legal without controlling a commander")
	}

	commander := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Commander",
		Types: []types.Card{types.Creature},
	}})
	g.Players[game.Player1].CommanderInstanceID = commander.CardInstanceID
	commander.PhasedOut = true
	if actionsContain(engine.legalActions(g, game.Player1), castBoth) {
		t.Fatal("phased-out commander enabled both modes")
	}
	commander.PhasedOut = false
	legal = engine.legalActions(g, game.Player1)
	if !actionsContain(legal, castBoth) {
		t.Fatalf("legal actions = %+v, want both-mode cast", legal)
	}
	if !engine.applyAction(g, game.Player1, castBoth) {
		t.Fatal("both-mode cast was rejected while controlling a commander")
	}
	movePermanentToZone(g, commander, zone.Graveyard)
	obj, ok := g.Stack.Peek()
	if !ok || !slices.Equal(obj.ChosenModes, []int{0, 1}) {
		t.Fatalf("stack chosen modes = %+v, want [0 1] locked at cast time", obj)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].ManaPool.Amount(mana.R); got != 2 {
		t.Fatalf("red mana after resolution = %d, want 2", got)
	}
	for _, cardID := range exiledIDs {
		if !g.Players[game.Player1].Exile.Contains(cardID) {
			t.Fatalf("card %d was not impulse-exiled", cardID)
		}
	}
}

func TestCommanderModalBonusRecognizesMergedCommander(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	commanderID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Merged Commander",
		Types: []types.Card{types.Creature},
	}})
	g.Players[game.Player1].Hand.Remove(commanderID)
	g.Players[game.Player1].CommanderInstanceID = commanderID
	top := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Mutate Top",
		Types: []types.Card{types.Creature},
	}})
	top.MergedCards = []game.MergedCard{{CardInstanceID: commanderID, Face: game.FaceFront}}

	spell := &game.CardDef{CardFace: game.CardFace{
		Name:  "Commander Modal Spell",
		Types: []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.AbilityContent{
			Modes:    []game.Mode{{}, {}},
			MinModes: 1,
			MaxModes: 1,
			ModeChoiceBonus: game.ModeChoiceBonus{
				Condition:          game.ModeChoiceConditionControlsCommander,
				AdditionalMaxModes: 1,
			},
		}),
	}}
	if choices := modeChoicesForSpellAt(g, game.Player1, spell); !slices.ContainsFunc(choices, func(choice []int) bool {
		return slices.Equal(choice, []int{0, 1})
	}) {
		t.Fatalf("mode choices = %v, want [0 1] while commander is merged beneath controlled permanent", choices)
	}
}

func TestTargetOpponentHandCountAddsRedMana(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	for range 3 {
		addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Hand Card"}})
	}
	target := game.TargetPlayerReference(0)
	obj := &game.StackObject{
		Controller: game.Player1,
		Targets:    []game.Target{game.PlayerTarget(game.Player2)},
	}
	resolveInstruction(engine, g, obj, game.AddMana{
		Amount: game.Dynamic(game.DynamicAmount{
			Kind:       game.DynamicAmountCountCardsInZone,
			Multiplier: 1,
			Player:     &target,
			CardZone:   zone.Hand,
			Selection:  &game.Selection{},
		}),
		ManaColor: mana.R,
	}, &TurnLog{})
	if got := g.Players[game.Player1].ManaPool.Amount(mana.R); got != 3 {
		t.Fatalf("red mana = %d, want 3", got)
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 3 {
		t.Fatalf("total mana units = %d, want 3", got)
	}
}

func TestImpulseExilePermitsSpellAndLandOnlyThisTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToLibrary(g, game.Player1, zeroCostImpulseSpell("Impulse Spell"))
	landID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Impulse Land",
		Types: []types.Card{types.Land},
	}})
	thirdID := addCardToLibrary(g, game.Player1, kickerSpell())
	unrelatedID := addCardToHand(g, game.Player1, zeroCostImpulseSpell("Unrelated Exile"))
	g.Players[game.Player1].Hand.Remove(unrelatedID)
	g.Players[game.Player1].Exile.Add(unrelatedID)

	resolveInstruction(engine, g, &game.StackObject{
		Controller:   game.Player1,
		SourceCardID: g.IDGen.Next(),
		SourceID:     g.IDGen.Next(),
	}, game.ImpulseExile{
		Player:   game.ControllerReference(),
		Amount:   game.Fixed(3),
		Duration: game.DurationThisTurn,
	}, &TurnLog{})

	for _, cardID := range []id.ID{spellID, landID, thirdID} {
		if !g.Players[game.Player1].Exile.Contains(cardID) {
			t.Fatalf("card %d was not exiled", cardID)
		}
	}
	setMainPhasePriority(g, game.Player1)
	g.Players[game.Player1].ManaPool.Add(mana.G, 2)
	legal := engine.legalActions(g, game.Player1)
	if !actionsContain(legal, action.CastSpellFromZone(spellID, zone.Exile, nil, 0, nil)) {
		t.Fatalf("legal actions = %+v, want impulse spell cast", legal)
	}
	if !actionsContain(legal, action.PlayLandFaceFromZone(landID, zone.Exile, game.FaceFront)) {
		t.Fatalf("legal actions = %+v, want impulse land play", legal)
	}
	if !actionsContain(legal, action.CastKickedSpellFromZone(thirdID, zone.Exile, nil, 0, nil)) {
		t.Fatalf("legal actions = %+v, want kicked impulse spell cast", legal)
	}
	if actionsContain(legal, action.CastSpellFromZone(unrelatedID, zone.Exile, nil, 0, nil)) {
		t.Fatal("unrelated exiled card received play permission")
	}
	if !engine.applyAction(g, game.Player1, action.PlayLandFaceFromZone(landID, zone.Exile, game.FaceFront)) {
		t.Fatal("playing impulse land from exile failed")
	}
	if !engine.applyAction(g, game.Player1, action.CastSpellFromZone(spellID, zone.Exile, nil, 0, nil)) {
		t.Fatal("casting impulse spell from exile failed")
	}

	expireRuleEffects(g)
	if actionsContain(engine.legalActions(g, game.Player1), action.CastSpellFromZone(thirdID, zone.Exile, nil, 0, nil)) {
		t.Fatal("impulse permission survived end-of-turn expiry")
	}
}

func TestImpulsePermissionClearsWhenCardChangesZones(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToLibrary(g, game.Player1, zeroCostImpulseSpell("Moving Spell"))
	resolveInstruction(engine, g, &game.StackObject{Controller: game.Player1}, game.ImpulseExile{
		Player:   game.ControllerReference(),
		Amount:   game.Fixed(1),
		Duration: game.DurationThisTurn,
	}, &TurnLog{})
	if !moveCardBetweenZones(g, game.Player1, cardID, zone.Exile, zone.Graveyard) ||
		!moveCardBetweenZones(g, game.Player1, cardID, zone.Graveyard, zone.Exile) {
		t.Fatal("failed to move impulse card through zones")
	}
	setMainPhasePriority(g, game.Player1)
	if actionsContain(engine.legalActions(g, game.Player1), action.CastSpellFromZone(cardID, zone.Exile, nil, 0, nil)) {
		t.Fatal("play permission followed the card through a zone change")
	}
}

func zeroCostImpulseSpell(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  name,
		Types: []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{{
			Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()},
		}}}.Ability()),
	}}
}

func TestImpulseExileUntilEndOfNextTurnSurvivesCurrentTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.TurnNumber = 5
	g.Turn.ActivePlayer = game.Player1
	spellID := addCardToLibrary(g, game.Player1, zeroCostImpulseSpell("Delayed Impulse"))

	resolveInstruction(engine, g, &game.StackObject{
		Controller:   game.Player1,
		SourceCardID: g.IDGen.Next(),
		SourceID:     g.IDGen.Next(),
	}, game.ImpulseExile{
		Player:   game.ControllerReference(),
		Amount:   game.Fixed(1),
		Duration: game.DurationUntilEndOfYourNextTurn,
	}, &TurnLog{})

	expireRuleEffects(g)
	setMainPhasePriority(g, game.Player1)
	if !actionsContain(engine.legalActions(g, game.Player1), action.CastSpellFromZone(spellID, zone.Exile, nil, 0, nil)) {
		t.Fatal("impulse permission expired before the controller's next turn")
	}

	g.Turn.TurnNumber = 7
	expireRuleEffects(g)
	if actionsContain(engine.legalActions(g, game.Player1), action.CastSpellFromZone(spellID, zone.Exile, nil, 0, nil)) {
		t.Fatal("impulse permission survived past the controller's next turn")
	}
}
