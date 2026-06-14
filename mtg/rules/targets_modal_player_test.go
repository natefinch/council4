package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestSingleModeContentDoesNotRequireModeChoice(t *testing.T) {
	content := game.Mode{
		Targets: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "player"}},
	}.Ability()

	choices := modeChoicesForContent(content)
	if len(choices) != 1 || len(choices[0]) != 0 {
		t.Fatalf("mode choices = %+v, want one choice with no selected modes", choices)
	}
	if !modesValidForContent(content, nil) {
		t.Fatal("single-mode content rejected an empty mode choice")
	}
	if modesValidForContent(content, []int{0}) {
		t.Fatal("single-mode content required an explicit mode selection")
	}
}

func TestBodyTargetsRejectMissingModalChoice(t *testing.T) {
	body := &game.ActivatedAbility{
		Content: game.AbilityContent{
			SharedTargets: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "player"}},
			MinModes:      1,
			MaxModes:      1,
			Modes:         []game.Mode{{}, {}},
		},
	}

	result := targetChoicesForBodyFromSourceObjectWithModes(nil, game.Player1, nil, 0, body, nil)
	if result.kind != targetInvalidSpec {
		t.Fatalf("target choice kind = %v, want targetInvalidSpec", result.kind)
	}
	if targetsValidForBodyFromSourceObjectWithModes(nil, game.Player1, nil, 0, body, nil, nil) {
		t.Fatal("targets accepted without a required modal choice")
	}
}

func TestModeChoiceRangeWithDuplicateModes(t *testing.T) {
	content := game.AbilityContent{
		MinModes:            3,
		MaxModes:            3,
		AllowDuplicateModes: true,
		Modes:               []game.Mode{{}, {}},
	}

	choices := modeChoicesForContent(content)
	want := [][]int{{0, 0, 0}, {0, 0, 1}, {0, 1, 1}, {1, 1, 1}}
	if !slices.EqualFunc(choices, want, slices.Equal) {
		t.Fatalf("mode choices = %v, want %v", choices, want)
	}
	if !modesValidForContent(content, []int{0, 1, 1}) {
		t.Fatal("valid repeated mode choice was rejected")
	}

	content.AllowDuplicateModes = false
	if choices := modeChoicesForContent(content); len(choices) != 0 {
		t.Fatalf("choices for impossible nonduplicate range = %v, want none", choices)
	}
	if modesValidForContent(content, []int{0, 1, 1}) {
		t.Fatal("impossible nonduplicate mode choice was accepted")
	}
}

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
		cast, ok := act.CastSpellPayload()
		if ok && cast.CardID == spellID {
			if len(cast.Targets) != 1 {
				t.Fatalf("cast targets = %d, want 1", len(cast.Targets))
			}
			castTargets = append(castTargets, cast.Targets[0])
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
	sourceID := addEffectSpellToStack(g, game.Player1, game.Damage{
		Amount:    game.Fixed(3),
		Recipient: game.AnyTargetDamageRecipient(0),
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
	addBasicLandPermanent(g, game.Player1, types.Forest)
	addBasicLandPermanent(g, game.Player2, types.Forest)
	spellID := addCardToHand(g, game.Player1, permanentTargetSpell("creature"))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	want := action.CastSpell(spellID, []game.Target{game.PermanentTarget(creature.ObjectID)}, 0, nil)
	if !containsAction(legal, want) {
		t.Fatalf("legal actions did not include creature target action %+v", want)
	}
	for _, act := range legal {
		cast, ok := act.CastSpellPayload()
		if !ok || cast.CardID != spellID {
			continue
		}
		if len(cast.Targets) != 1 || cast.Targets[0] != game.PermanentTarget(creature.ObjectID) {
			t.Fatalf("unexpected cast target %+v", cast.Targets)
		}
	}
}

func TestOptionalPermanentTargetAllowsTargetOrNoTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCreaturePermanent(g, game.Player1)
	addBasicLandPermanent(g, game.Player1, types.Forest)
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

	result := targetChoicesForSpecs(g, game.Player1, nil, 0, []game.TargetSpec{
		{MinTargets: 0, MaxTargets: 2, Constraint: "creature"},
	})

	if result.kind != targetLegalChoicesFound {
		t.Fatalf("kind = %v, want targetLegalChoicesFound", result.kind)
	}
	if len(result.choices) != 1 {
		t.Fatalf("choices = %d, want one no-target choice", len(result.choices))
	}
	if result.choices[0] != nil {
		t.Fatalf("choice = %+v, want nil no-target choice", result.choices[0])
	}
}
