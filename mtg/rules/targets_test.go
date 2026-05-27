package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
)

func TestPlayerTargetedSpellCreatesOneLegalActionPerAlivePlayer(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, playerDamageSpell())
	if !engine.eliminatePlayer(g, game.Player4) {
		t.Fatal("eliminatePlayer() = false, want true")
	}
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	var castTargets []game.Target
	for _, act := range legal {
		if act.Kind == action.ActionCastSpell && act.CastSpell.CardID == spellID {
			if len(act.CastSpell.Targets) != 1 {
				t.Fatalf("cast targets = %d, want 1", len(act.CastSpell.Targets))
			}
			castTargets = append(castTargets, act.CastSpell.Targets[0])
		}
	}
	wantTargets := []game.Target{
		game.PlayerTarget(game.Player1),
		game.PlayerTarget(game.Player2),
		game.PlayerTarget(game.Player3),
	}
	if len(castTargets) != len(wantTargets) {
		t.Fatalf("cast actions = %d, want %d", len(castTargets), len(wantTargets))
	}
	for i, want := range wantTargets {
		if castTargets[i] != want {
			t.Fatalf("target %d = %+v, want %+v", i, castTargets[i], want)
		}
	}
}

func TestTargetedEffectUsesSelectedTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, playerDamageSpell())
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, []game.Target{game.PlayerTarget(game.Player2)}, 0, nil)) {
		t.Fatal("applyAction() = false, want true")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Players[game.Player1].Life != 40 {
		t.Fatalf("caster life = %d, want 40", g.Players[game.Player1].Life)
	}
	if g.Players[game.Player2].Life != 37 {
		t.Fatalf("target life = %d, want 37", g.Players[game.Player2].Life)
	}
}

func TestDeadPlayerTargetDoesNotApplyEffect(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	sourceID := addEffectSpellToStack(g, game.Player1, game.Effect{
		Type:        game.EffectDamage,
		Amount:      3,
		TargetIndex: 0,
	}, []game.Target{game.PlayerTarget(game.Player2)})
	if !engine.eliminatePlayer(g, game.Player2) {
		t.Fatal("eliminatePlayer() = false, want true")
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Players[game.Player2].Life != 40 {
		t.Fatalf("dead target life = %d, want 40", g.Players[game.Player2].Life)
	}
	if !g.Players[game.Player1].Graveyard.Contains(sourceID) {
		t.Fatal("spell did not move to graveyard")
	}
}

func TestTargetThatDiesBeforeResolutionDoesNotApplyEffect(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, playerDamageSpell())
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, []game.Target{game.PlayerTarget(game.Player2)}, 0, nil)) {
		t.Fatal("applyAction() = false, want true")
	}
	if !engine.eliminatePlayer(g, game.Player2) {
		t.Fatal("eliminatePlayer() = false, want true")
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Players[game.Player2].Life != 40 {
		t.Fatalf("dead target life = %d, want 40", g.Players[game.Player2].Life)
	}
	if !g.Players[game.Player1].Graveyard.Contains(spellID) {
		t.Fatal("spell did not move to graveyard")
	}
}

func TestPermanentTargetedSpellCreatesActionsForMatchingPermanents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCreaturePermanent(g, game.Player1)
	addBasicLandPermanent(g, game.Player1, "Forest")
	addBasicLandPermanent(g, game.Player2, "Forest")
	spellID := addCardToHand(g, game.Player1, permanentTargetSpell("creature"))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	want := action.CastSpell(spellID, []game.Target{game.PermanentTarget(creature.ObjectID)}, 0, nil)
	if !containsAction(legal, want) {
		t.Fatalf("legal actions did not include creature target action %+v", want)
	}
	for _, act := range legal {
		if act.Kind != action.ActionCastSpell || act.CastSpell.CardID != spellID {
			continue
		}
		if len(act.CastSpell.Targets) != 1 || act.CastSpell.Targets[0] != game.PermanentTarget(creature.ObjectID) {
			t.Fatalf("unexpected cast target %+v", act.CastSpell.Targets)
		}
	}
}

func TestOptionalPermanentTargetAllowsTargetOrNoTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCreaturePermanent(g, game.Player1)
	addBasicLandPermanent(g, game.Player1, "Forest")
	spellID := addCardToHand(g, game.Player1, optionalPermanentTargetSpell("creature"))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	if !containsAction(legal, action.CastSpell(spellID, []game.Target{game.PermanentTarget(creature.ObjectID)}, 0, nil)) {
		t.Fatal("legal actions did not include optional target spell with a target")
	}
	if !containsAction(legal, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("legal actions did not include optional target spell with no target")
	}
	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, []game.Target{game.PermanentTarget(creature.ObjectID)}, 0, nil)) {
		t.Fatal("applyAction() with optional target = false, want true")
	}
}

func TestOptionalTargetWithNoCandidatesHasSingleNoTargetChoice(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	choices := targetChoicesForSpecs(g, game.Player1, nil, 0, []game.TargetSpec{
		{MinTargets: 0, MaxTargets: 2, Constraint: "creature"},
	})

	if len(choices) != 1 {
		t.Fatalf("choices = %d, want one no-target choice", len(choices))
	}
	if choices[0] != nil {
		t.Fatalf("choice = %+v, want nil no-target choice", choices[0])
	}
}

func TestMultiTargetSpellCreatesCombinations(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	first := addCreaturePermanent(g, game.Player1)
	second := addCreaturePermanent(g, game.Player1)
	third := addCreaturePermanent(g, game.Player2)
	addBasicLandPermanent(g, game.Player1, "Forest")
	spellID := addCardToHand(g, game.Player1, permanentTargetSpellWithRange("creature", 2, 2))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	wantTargets := [][]game.Target{
		{game.PermanentTarget(first.ObjectID), game.PermanentTarget(second.ObjectID)},
		{game.PermanentTarget(first.ObjectID), game.PermanentTarget(third.ObjectID)},
		{game.PermanentTarget(second.ObjectID), game.PermanentTarget(third.ObjectID)},
	}
	for _, targets := range wantTargets {
		if !containsAction(legal, action.CastSpell(spellID, targets, 0, nil)) {
			t.Fatalf("legal actions did not include targets %+v", targets)
		}
	}
	if containsAction(legal, action.CastSpell(spellID, []game.Target{game.PermanentTarget(first.ObjectID)}, 0, nil)) {
		t.Fatal("legal actions included too few targets for two-target spell")
	}
}

func TestUpToTwoTargetsIncludesZeroOneAndTwoTargets(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	first := addCreaturePermanent(g, game.Player1)
	second := addCreaturePermanent(g, game.Player2)
	addBasicLandPermanent(g, game.Player1, "Forest")
	spellID := addCardToHand(g, game.Player1, permanentTargetSpellWithRange("creature", 0, 2))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	for _, targets := range [][]game.Target{
		nil,
		{game.PermanentTarget(first.ObjectID)},
		{game.PermanentTarget(second.ObjectID)},
		{game.PermanentTarget(first.ObjectID), game.PermanentTarget(second.ObjectID)},
	} {
		if !containsAction(legal, action.CastSpell(spellID, targets, 0, nil)) {
			t.Fatalf("legal actions did not include targets %+v", targets)
		}
	}
}

func TestMixedTargetSlotsCreateCartesianProduct(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCreaturePermanent(g, game.Player1)
	addBasicLandPermanent(g, game.Player1, "Forest")
	spellID := addCardToHand(g, game.Player1, permanentTargetSpellWithSpecs([]game.TargetSpec{
		{MinTargets: 1, MaxTargets: 1, Constraint: "creature"},
		{MinTargets: 1, MaxTargets: 1, Constraint: "opponent"},
	}))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	for _, playerID := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		targets := []game.Target{game.PermanentTarget(creature.ObjectID), game.PlayerTarget(playerID)}
		if !containsAction(legal, action.CastSpell(spellID, targets, 0, nil)) {
			t.Fatalf("legal actions did not include mixed targets %+v", targets)
		}
	}
}

func TestStructuredTargetPredicates(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	blackCreature := addCombatPermanent(g, game.Player1, &game.CardDef{
		Name:      "Black Creature",
		ManaValue: 2,
		Colors:    []mana.Color{mana.Black},
		Types:     []game.CardType{game.TypeCreature},
		Power:     optPT(game.PT{Value: 3}),
		Toughness: optPT(game.PT{Value: 3}),
	})
	whiteCreature := addCombatPermanent(g, game.Player2, &game.CardDef{
		Name:      "White Creature",
		ManaValue: 4,
		Colors:    []mana.Color{mana.White},
		Types:     []game.CardType{game.TypeCreature},
		Power:     optPT(game.PT{Value: 2}),
		Toughness: optPT(game.PT{Value: 2}),
		Abilities: []game.AbilityDef{{Kind: game.StaticAbility, Keywords: []game.Keyword{game.Flying}}},
	})
	whiteCreature.Tapped = true
	addBasicLandPermanent(g, game.Player1, "Forest")
	spellID := addCardToHand(g, game.Player1, permanentTargetSpellWithSpecs([]game.TargetSpec{
		{
			MinTargets: 1,
			MaxTargets: 1,
			Constraint: "nonblack tapped creature with flying mana value 4 or less an opponent controls",
			Allow:      game.TargetAllowPermanent,
			Predicate: game.TargetPredicate{
				PermanentTypes: []game.CardType{game.TypeCreature},
				ExcludedColors: []mana.Color{mana.Black},
				Controller:     game.ControllerOpponent,
				Tapped:         game.TriTrue,
				Keyword:        game.Flying,
				ManaValue:      optIntComparison(game.IntComparison{Op: game.CompareLessOrEqual, Value: 4}),
			},
		},
	}))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	if !containsAction(legal, action.CastSpell(spellID, []game.Target{game.PermanentTarget(whiteCreature.ObjectID)}, 0, nil)) {
		t.Fatal("structured predicate did not allow matching creature")
	}
	if containsAction(legal, action.CastSpell(spellID, []game.Target{game.PermanentTarget(blackCreature.ObjectID)}, 0, nil)) {
		t.Fatal("structured predicate allowed excluded black creature")
	}
}

func TestStructuredAllowPermanentWithoutTypePredicateAllowsAnyPermanent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	artifact := addCombatPermanent(g, game.Player1, &game.CardDef{
		Name:  "Relic",
		Types: []game.CardType{game.TypeArtifact},
	})
	addBasicLandPermanent(g, game.Player1, "Forest")
	spellID := addCardToHand(g, game.Player1, permanentTargetSpellWithSpecs([]game.TargetSpec{
		{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPermanent},
	}))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	if !containsAction(legal, action.CastSpell(spellID, []game.Target{game.PermanentTarget(artifact.ObjectID)}, 0, nil)) {
		t.Fatal("structured permanent target without type predicate did not allow artifact")
	}
}

func TestStructuredAnyTargetAllowsOnlyDamageablePermanents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	artifact := addCombatPermanent(g, game.Player1, &game.CardDef{
		Name:  "Relic",
		Types: []game.CardType{game.TypeArtifact},
	})
	creature := addCreaturePermanent(g, game.Player1)
	addBasicLandPermanent(g, game.Player1, "Forest")
	spellID := addCardToHand(g, game.Player1, permanentTargetSpellWithSpecs([]game.TargetSpec{
		{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPermanent | game.TargetAllowPlayer},
	}))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	if containsAction(legal, action.CastSpell(spellID, []game.Target{game.PermanentTarget(artifact.ObjectID)}, 0, nil)) {
		t.Fatal("structured any target allowed noncreature artifact")
	}
	if !containsAction(legal, action.CastSpell(spellID, []game.Target{game.PermanentTarget(creature.ObjectID)}, 0, nil)) {
		t.Fatal("structured any target did not allow creature")
	}
	if !containsAction(legal, action.CastSpell(spellID, []game.Target{game.PlayerTarget(game.Player2)}, 0, nil)) {
		t.Fatal("structured any target did not allow player")
	}
}

func TestAnotherTargetPredicateExcludesSourcePermanent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{
		Name:  "Source Creature",
		Types: []game.CardType{game.TypeCreature},
		Abilities: []game.AbilityDef{
			{
				Kind: game.ActivatedAbility,
				Targets: []game.TargetSpec{
					{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "another target creature",
						Allow:      game.TargetAllowPermanent,
						Predicate: game.TargetPredicate{
							PermanentTypes: []game.CardType{game.TypeCreature},
							Another:        true,
						},
					},
				},
				Effects: []game.Effect{{Type: game.EffectTap, TargetIndex: 0}},
			},
		},
	})
	other := addCreaturePermanent(g, game.Player1)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	if containsAction(legal, action.ActivateAbility(source.ObjectID, 0, []game.Target{game.PermanentTarget(source.ObjectID)}, 0)) {
		t.Fatal("another target predicate allowed source permanent")
	}
	if !containsAction(legal, action.ActivateAbility(source.ObjectID, 0, []game.Target{game.PermanentTarget(other.ObjectID)}, 0)) {
		t.Fatal("another target predicate did not allow other creature")
	}
}

func TestPermanentTargetConstraintsCanRequireOpponentControlledNonland(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	ownCreature := addCreaturePermanent(g, game.Player1)
	opponentCreature := addCreaturePermanent(g, game.Player2)
	opponentLand := addBasicLandPermanent(g, game.Player2, "Forest")
	addBasicLandPermanent(g, game.Player1, "Forest")
	spellID := addCardToHand(g, game.Player1, permanentTargetSpell("nonland permanent an opponent controls"))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	want := action.CastSpell(spellID, []game.Target{game.PermanentTarget(opponentCreature.ObjectID)}, 0, nil)
	if !containsAction(legal, want) {
		t.Fatalf("legal actions did not include opponent nonland permanent target %+v", want)
	}
	for _, forbidden := range []game.Target{game.PermanentTarget(ownCreature.ObjectID), game.PermanentTarget(opponentLand.ObjectID)} {
		if containsAction(legal, action.CastSpell(spellID, []game.Target{forbidden}, 0, nil)) {
			t.Fatalf("legal actions included forbidden target %+v", forbidden)
		}
	}
}

func TestIllegalPermanentTargetIsRejectedDuringCast(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	land := addBasicLandPermanent(g, game.Player2, "Forest")
	addBasicLandPermanent(g, game.Player1, "Forest")
	spellID := addCardToHand(g, game.Player1, permanentTargetSpell("creature"))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if engine.applyAction(g, game.Player1, action.CastSpell(spellID, []game.Target{game.PermanentTarget(land.ObjectID)}, 0, nil)) {
		t.Fatal("cast with illegal permanent target succeeded")
	}
}

func TestPermanentTargetThatLeavesBeforeResolutionCountersSpellByRules(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCreaturePermanent(g, game.Player2)
	sourceID := addEffectSpellToStack(g, game.Player1, game.Effect{
		Type:        game.EffectDamage,
		Amount:      3,
		TargetIndex: 0,
	}, []game.Target{game.PermanentTarget(target.ObjectID)})
	card, ok := g.GetCardInstance(sourceID)
	if !ok {
		t.Fatal("source card instance not found")
	}
	card.Def.Abilities[0].Targets = []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "creature"}}
	if !movePermanentToZone(g, target, game.ZoneGraveyard) {
		t.Fatal("movePermanentToZone() = false, want true")
	}
	log := TurnLog{}

	engine.resolveTopOfStack(g, &log)

	if len(log.Resolves) != 1 || log.Resolves[0].Result != "countered by rules" {
		t.Fatalf("resolve log = %+v, want countered by rules", log.Resolves)
	}
	if !g.Players[game.Player1].Graveyard.Contains(sourceID) {
		t.Fatal("countered spell did not move to graveyard")
	}
}

func TestPermanentTargetedDamageMarksDamageOnResolution(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCreaturePermanent(g, game.Player2)
	sourceID := addEffectSpellToStack(g, game.Player1, game.Effect{
		Type:        game.EffectDamage,
		Amount:      3,
		TargetIndex: 0,
	}, []game.Target{game.PermanentTarget(target.ObjectID)})
	card, ok := g.GetCardInstance(sourceID)
	if !ok {
		t.Fatal("source card instance not found")
	}
	card.Def.Abilities[0].Targets = []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "creature"}}

	engine.resolveTopOfStack(g, &TurnLog{})

	if target.MarkedDamage != 3 {
		t.Fatalf("marked damage = %d, want 3", target.MarkedDamage)
	}
	if !g.Players[game.Player1].Graveyard.Contains(sourceID) {
		t.Fatal("resolved spell did not move to graveyard")
	}
}
func playerDamageSpell() *game.CardDef {
	return &game.CardDef{
		Name:  "Needle Drop",
		Types: []game.CardType{game.TypeSorcery},
		Abilities: []game.AbilityDef{
			{
				Kind: game.SpellAbility,
				Targets: []game.TargetSpec{
					{MinTargets: 1, MaxTargets: 1, Constraint: "player"},
				},
				Effects: []game.Effect{
					{Type: game.EffectDamage, Amount: 3, TargetIndex: 0},
				},
			},
		},
	}
}

func permanentTargetSpell(constraint string) *game.CardDef {
	return permanentTargetSpellWithRange(constraint, 1, 1)
}

func optionalPermanentTargetSpell(constraint string) *game.CardDef {
	return permanentTargetSpellWithRange(constraint, 0, 1)
}

func permanentTargetSpellWithRange(constraint string, minTargets int, maxTargets int) *game.CardDef {
	return permanentTargetSpellWithSpecs([]game.TargetSpec{
		{MinTargets: minTargets, MaxTargets: maxTargets, Constraint: constraint},
	})
}

func permanentTargetSpellWithSpecs(specs []game.TargetSpec) *game.CardDef {
	return &game.CardDef{
		Name:     "Permanent Target Spell",
		ManaCost: greenCost(),
		Types:    []game.CardType{game.TypeSorcery},
		Abilities: []game.AbilityDef{
			{
				Kind:    game.SpellAbility,
				Targets: specs,
			},
		},
	}
}
