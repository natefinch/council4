package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func commandeerAbility() game.AbilityContent {
	return game.Mode{
		Targets: []game.TargetSpec{{
			MinTargets: 1,
			MaxTargets: 1,
			Allow:      game.TargetAllowStackObject,
			Predicate: game.TargetPredicate{
				StackObjectKinds:       []game.StackObjectKind{game.StackSpell},
				ExcludedSpellCardTypes: []types.Card{types.Creature},
			},
		}},
		Sequence: []game.Instruction{
			{Primitive: game.ChangeStackObjectController{
				Object:     game.TargetStackObjectReference(0),
				Controller: game.ControllerReference(),
			}},
			{
				Primitive: game.ChooseNewTargets{Object: game.TargetStackObjectReference(0)},
				Optional:  true,
			},
		},
	}.Ability()
}

func commandeerDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Commandeer",
		ManaCost: opt.Val(cost.Mana{cost.O(5), cost.U, cost.U}),
		Colors:   []color.Color{color.Blue},
		Types:    []types.Card{types.Instant},
		AlternativeCosts: []cost.Alternative{{
			Label: "Exile 2 blue cards",
			AdditionalCosts: []cost.Additional{{
				Kind:           cost.AdditionalExile,
				Amount:         2,
				Source:         zone.Hand,
				MatchCardColor: true,
				CardColor:      color.Blue,
			}},
		}},
		SpellAbility: opt.Val(commandeerAbility()),
	}}
}

func targetedStackSpell(g *game.Game, owner, controller game.PlayerID, cardTypes []types.Card, target game.Target) (*game.StackObject, *game.CardInstance) {
	cardID := g.IDGen.Next()
	card := &game.CardInstance{
		ID: cardID,
		Def: &game.CardDef{CardFace: game.CardFace{
			Name:  "Targeted Spell",
			Types: cardTypes,
			SpellAbility: opt.Val(game.Mode{Targets: []game.TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      game.TargetAllowPermanent,
				Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
			}}}.Ability()),
		}},
		Owner: owner,
	}
	g.CardInstances[cardID] = card
	obj := &game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackSpell,
		SourceID:     cardID,
		Controller:   controller,
		Targets:      []game.Target{target},
		TargetCounts: []int{1},
	}
	g.Stack.Push(obj)
	return obj, card
}

func resolveCommandeerEffect(g *game.Game, victim *game.StackObject, agents [game.NumPlayers]PlayerAgent) {
	addInstructionSpellToStackForController(g, game.Player1, commandeerAbility().Modes[0].Sequence,
		[]game.Target{game.StackObjectTarget(victim.ID)})
	NewEngine(nil).resolveTopOfStackWithChoices(g, agents, &TurnLog{})
}

func TestCommandeerChangesControllerThenMayRetarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	first := addCreaturePermanent(g, game.Player1)
	second := addCreaturePermanent(g, game.Player1)
	victim, card := targetedStackSpell(g, game.Player2, game.Player2, []types.Card{types.Instant}, game.PermanentTarget(first.ObjectID))
	originalID, originalSource := victim.ID, victim.SourceID

	resolveCommandeerEffect(g, victim, [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}, {1}}},
	})

	if victim.Controller != game.Player1 {
		t.Fatalf("controller = %v, want Player1", victim.Controller)
	}
	if victim.ID != originalID || victim.SourceID != originalSource || card.Owner != game.Player2 {
		t.Fatal("controller change altered stack identity, source card, or ownership")
	}
	if got := victim.Targets[0].PermanentID; got != second.ObjectID {
		t.Fatalf("new target = %v, want %v", got, second.ObjectID)
	}
}

func TestCommandeerMayDeclineRetarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	first := addCreaturePermanent(g, game.Player1)
	addCreaturePermanent(g, game.Player1)
	victim, _ := targetedStackSpell(g, game.Player2, game.Player2, []types.Card{types.Instant}, game.PermanentTarget(first.ObjectID))

	resolveCommandeerEffect(g, victim, [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}},
	})

	if victim.Controller != game.Player1 {
		t.Fatalf("controller = %v, want Player1", victim.Controller)
	}
	if got := victim.Targets[0].PermanentID; got != first.ObjectID {
		t.Fatalf("declined retarget changed target to %v", got)
	}
}

func TestCommandeerRejectsCreatureSpellTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := addCreaturePermanent(g, game.Player1)
	victim, _ := targetedStackSpell(g, game.Player2, game.Player2, []types.Card{types.Creature}, game.PermanentTarget(target.ObjectID))
	spellID := addCardToHand(g, game.Player1, commandeerDef())
	first := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Blue One", Colors: []color.Color{color.Blue}}})
	second := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Blue Two", Colors: []color.Color{color.Blue}}})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if NewEngine(nil).applyAction(g, game.Player1,
		action.CastSpell(spellID, []game.Target{game.StackObjectTarget(victim.ID)}, 0, nil)) {
		t.Fatal("Commandeer targeted a creature spell")
	}
	if !g.Players[game.Player1].Hand.Contains(spellID) ||
		!g.Players[game.Player1].Hand.Contains(first) ||
		!g.Players[game.Player1].Hand.Contains(second) {
		t.Fatal("illegal target attempt did not leave the hand intact")
	}
}

func TestCommandeerFizzlesWhenTargetLeavesStack(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := addCreaturePermanent(g, game.Player1)
	victim, _ := targetedStackSpell(g, game.Player2, game.Player2, []types.Card{types.Instant}, game.PermanentTarget(target.ObjectID))
	addInstructionSpellToStackForController(g, game.Player1, commandeerAbility().Modes[0].Sequence,
		[]game.Target{game.StackObjectTarget(victim.ID)})
	if _, ok := g.Stack.RemoveByID(victim.ID); !ok {
		t.Fatal("failed to remove targeted spell")
	}

	NewEngine(nil).resolveTopOfStack(g, &TurnLog{})

	if victim.Controller != game.Player2 {
		t.Fatal("fizzled Commandeer changed the removed spell's controller")
	}
	if g.Stack.Size() != 0 {
		t.Fatal("fizzled Commandeer remained on the stack")
	}
}

func TestCommandeerControlsUncounterableSpellAndUsesOwnerDestination(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := addCreaturePermanent(g, game.Player1)
	victim, card := targetedStackSpell(g, game.Player2, game.Player2, []types.Card{types.Instant}, game.PermanentTarget(target.ObjectID))
	victim.RuleEffects = []game.RuleEffect{{Kind: game.RuleEffectCantBeCountered}}

	resolveCommandeerEffect(g, victim, [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}},
	})
	if victim.Controller != game.Player1 {
		t.Fatal("uncounterable spell did not change controller")
	}
	NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player2].Graveyard.Contains(card.ID) {
		t.Fatal("stolen instant did not go to its owner's graveyard")
	}
	if g.Players[game.Player1].Graveyard.Contains(card.ID) {
		t.Fatal("stolen instant went to its controller's graveyard")
	}
}

func TestCommandeerPermanentSpellEntersUnderNewController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := addCreaturePermanent(g, game.Player1)
	victim, card := targetedStackSpell(g, game.Player2, game.Player2, []types.Card{types.Artifact}, game.PermanentTarget(target.ObjectID))

	resolveCommandeerEffect(g, victim, [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}},
	})
	NewEngine(nil).resolveTopOfStack(g, &TurnLog{})

	permanent := permanentByCardID(g, card.ID)
	if permanent == nil || permanent.Controller != game.Player1 {
		t.Fatalf("resolved permanent = %#v, want controlled by Player1", permanent)
	}
	if card.Owner != game.Player2 {
		t.Fatalf("card owner = %v, want Player2", card.Owner)
	}
}

func TestCommandeerPermanentSpellCopyBecomesTokenForNewController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := &game.CardInstance{
		ID:    g.IDGen.Next(),
		Def:   &game.CardDef{CardFace: game.CardFace{Name: "Copied Artifact", Types: []types.Card{types.Artifact}}},
		Owner: game.Player2,
	}
	g.CardInstances[source.ID] = source
	copyObj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   source.ID,
		Controller: game.Player2,
		Copy:       true,
	}
	g.Stack.Push(copyObj)
	addEffectSpellToStack(g, game.Player1, game.ChangeStackObjectController{
		Object:     game.TargetStackObjectReference(0),
		Controller: game.ControllerReference(),
	}, []game.Target{game.StackObjectTarget(copyObj.ID)})
	engine := NewEngine(nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	engine.resolveTopOfStack(g, &TurnLog{})

	var token *game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanent.Controller == game.Player1 {
			token = permanent
			break
		}
	}
	if token == nil {
		t.Fatal("controlled permanent-spell copy did not become a token")
	}
	if source.Owner != game.Player2 {
		t.Fatal("copy controller change altered source-card ownership")
	}
}

func TestCommandeerAuraSpellCopyEntersAttachedForNewController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := addCreaturePermanent(g, game.Player1)
	auraDef := &game.CardDef{CardFace: game.CardFace{
		Name:     "Copied Aura",
		Types:    []types.Card{types.Enchantment},
		Subtypes: []types.Sub{types.Aura},
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.EnchantKeyword{Target: game.TargetSpec{
				Allow:     game.TargetAllowPermanent,
				Selection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
			}}},
		}},
	}}
	source := &game.CardInstance{ID: g.IDGen.Next(), Def: auraDef, Owner: game.Player2}
	g.CardInstances[source.ID] = source
	copyObj := &game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackSpell,
		SourceID:     source.ID,
		Controller:   game.Player2,
		Copy:         true,
		Targets:      []game.Target{game.PermanentTarget(target.ObjectID)},
		TargetCounts: []int{1},
	}
	g.Stack.Push(copyObj)
	addEffectSpellToStack(g, game.Player1, game.ChangeStackObjectController{
		Object:     game.TargetStackObjectReference(0),
		Controller: game.ControllerReference(),
	}, []game.Target{game.StackObjectTarget(copyObj.ID)})
	engine := NewEngine(nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	engine.resolveTopOfStack(g, &TurnLog{})

	for _, permanent := range g.Battlefield {
		if permanent.Token && permanent.Controller == game.Player1 &&
			permanent.AttachedTo.Exists && permanent.AttachedTo.Val == target.ObjectID {
			return
		}
	}
	t.Fatal("controlled Aura-spell copy did not become a token attached to its target")
}

func TestCommandeerPitchCostExactCardsAndSourceExclusion(t *testing.T) {
	t.Run("succeeds with exactly two other blue cards", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		target := addCreaturePermanent(g, game.Player1)
		victim, _ := targetedStackSpell(g, game.Player2, game.Player2, []types.Card{types.Instant}, game.PermanentTarget(target.ObjectID))
		spellID := addCardToHand(g, game.Player1, commandeerDef())
		first := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Blue One", Colors: []color.Color{color.Blue}}})
		second := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Blue Two", Colors: []color.Color{color.Blue}}})
		lands := make([]*game.Permanent, 7)
		for i := range lands {
			lands[i] = addBasicLandPermanent(g, game.Player1, types.Island)
		}
		g.Turn.Phase = game.PhasePrecombatMain
		g.Turn.Step = game.StepNone

		if !NewEngine(nil).applyActionWithChoices(g, game.Player1,
			action.CastSpell(spellID, []game.Target{game.StackObjectTarget(victim.ID)}, 0, nil),
			[game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{1}, {0, 1}}}},
			&TurnLog{}) {
			t.Fatal("pitch cast failed")
		}
		if !g.Players[game.Player1].Exile.Contains(first) || !g.Players[game.Player1].Exile.Contains(second) {
			t.Fatal("exactly two blue cards were not exiled")
		}
		for _, land := range lands {
			if land.Tapped {
				t.Fatal("chosen pitch alternative also paid the mana cost")
			}
		}
	})

	t.Run("cannot use the spell itself or a nonblue card and rolls back", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		target := addCreaturePermanent(g, game.Player1)
		victim, _ := targetedStackSpell(g, game.Player2, game.Player2, []types.Card{types.Instant}, game.PermanentTarget(target.ObjectID))
		spellID := addCardToHand(g, game.Player1, commandeerDef())
		blue := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Only Blue", Colors: []color.Color{color.Blue}}})
		red := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Red", Colors: []color.Color{color.Red}}})
		g.Turn.Phase = game.PhasePrecombatMain
		g.Turn.Step = game.StepNone

		if NewEngine(nil).applyAction(g, game.Player1,
			action.CastSpell(spellID, []game.Target{game.StackObjectTarget(victim.ID)}, 0, nil)) {
			t.Fatal("pitch cast succeeded without two other blue cards")
		}
		if !g.Players[game.Player1].Hand.Contains(spellID) ||
			!g.Players[game.Player1].Hand.Contains(blue) ||
			!g.Players[game.Player1].Hand.Contains(red) ||
			g.Players[game.Player1].Exile.Size() != 0 {
			t.Fatal("failed pitch cast did not roll back cleanly")
		}
	})
}

func TestFreeCastDoesNotAlsoPayCommandeerPitchCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := addCreaturePermanent(g, game.Player1)
	victim, _ := targetedStackSpell(g, game.Player2, game.Player2, []types.Card{types.Instant}, game.PermanentTarget(target.ObjectID))
	spellID := addCardToExile(g, game.Player1, commandeerDef())
	first := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Blue One", Colors: []color.Color{color.Blue}}})
	second := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Blue Two", Colors: []color.Color{color.Blue}}})
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		Kind:                  game.RuleEffectPlayFromZone,
		Controller:            game.Player1,
		AffectedPlayer:        game.PlayerYou,
		Duration:              game.DurationThisTurn,
		CastFromZone:          zone.Exile,
		AffectedCardID:        spellID,
		WithoutPayingManaCost: true,
		ExpiresFor:            game.Player1,
	})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1

	if !NewEngine(nil).applyAction(g, game.Player1,
		action.CastSpellFaceFromZone(spellID, zone.Exile, game.FaceFront,
			[]game.Target{game.StackObjectTarget(victim.ID)}, 0, nil)) {
		t.Fatal("free cast failed")
	}
	if !g.Players[game.Player1].Hand.Contains(first) || !g.Players[game.Player1].Hand.Contains(second) {
		t.Fatal("free cast incorrectly paid the pitch alternative")
	}
}

func TestCommandeerPitchCostStillPaysCommanderTax(t *testing.T) {
	g := newCommanderCastGame(commandeerDef())
	target := addCreaturePermanent(g, game.Player1)
	victim, _ := targetedStackSpell(g, game.Player2, game.Player2, []types.Card{types.Instant}, game.PermanentTarget(target.ObjectID))
	first := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Blue One", Colors: []color.Color{color.Blue}}})
	second := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Blue Two", Colors: []color.Color{color.Blue}}})
	island := addBasicLandPermanent(g, game.Player1, types.Island)
	forest := addBasicLandPermanent(g, game.Player1, types.Forest)
	player := g.Players[game.Player1]
	player.CommanderCastCount = 1
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1

	if !NewEngine(nil).applyActionWithChoices(g, game.Player1,
		action.CastCommanderSpell(player.CommanderInstanceID, []game.Target{game.StackObjectTarget(victim.ID)}, 0, nil),
		[game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{0, 1}}}},
		&TurnLog{}) {
		t.Fatal("pitch commander cast with tax failed")
	}
	if !g.Players[game.Player1].Exile.Contains(first) || !g.Players[game.Player1].Exile.Contains(second) {
		t.Fatal("pitch commander cast did not exile two blue cards")
	}
	if !island.Tapped || !forest.Tapped {
		t.Fatal("pitch commander cast did not pay the {2} commander tax")
	}
	if player.CommanderCastCount != 2 {
		t.Fatalf("commander cast count = %d, want 2", player.CommanderCastCount)
	}
}
