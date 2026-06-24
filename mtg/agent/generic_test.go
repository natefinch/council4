package agent

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/rules"
	"github.com/natefinch/council4/opt"
)

// Compile-time check that GenericStrategy drives an Agent.
var _ Strategy = GenericStrategy{}

func creatureCardDef(name string, power, toughness int) *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:      name,
			Types:     []types.Card{types.Creature},
			Power:     opt.Val(game.PT{Value: power}),
			Toughness: opt.Val(game.PT{Value: toughness}),
		},
	}
}

func addObservedHandCard(g *game.Game, owner game.PlayerID, def *game.CardDef) id.ID {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{ID: cardID, Def: def, Owner: owner}
	g.Players[owner].Hand.Add(cardID)
	return cardID
}

func addObservedPermanent(g *game.Game, controller game.PlayerID, def *game.CardDef) *game.Permanent {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{ID: cardID, Def: def, Owner: controller}
	permanent := &game.Permanent{
		ObjectID:       g.IDGen.Next(),
		CardInstanceID: cardID,
		Owner:          controller,
		Controller:     controller,
	}
	g.Battlefield = append(g.Battlefield, permanent)
	return permanent
}

func TestGenericStrategyPrefersLandAndCreatureOverPass(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creatureID := addObservedHandCard(g, game.Player1, creatureCardDef("Bear", 2, 2))
	obs := rules.NewObservation(g, game.Player1)
	strategy := GenericStrategy{}

	pass := strategy.ScoreAction(obs, action.Pass())
	land := strategy.ScoreAction(obs, action.PlayLandFace(g.IDGen.Next(), game.FaceFront))
	cast := strategy.ScoreAction(obs, action.CastSpell(creatureID, nil, 0, nil))

	if land <= pass {
		t.Errorf("land score %v should beat pass %v", land, pass)
	}
	if cast <= pass {
		t.Errorf("cast-creature score %v should beat pass %v", cast, pass)
	}
}

func TestGenericStrategyTargetsBiggestThreat(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	removalID := addObservedHandCard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Removal",
		Types: []types.Card{types.Instant},
	}})
	small := addObservedPermanent(g, game.Player2, creatureCardDef("Small", 2, 2))
	big := addObservedPermanent(g, game.Player2, creatureCardDef("Big", 6, 6))

	obs := rules.NewObservation(g, game.Player1)
	strategy := GenericStrategy{}

	scoreSmall := strategy.ScoreAction(obs, action.CastSpell(removalID, []game.Target{game.PermanentTarget(small.ObjectID)}, 0, nil))
	scoreBig := strategy.ScoreAction(obs, action.CastSpell(removalID, []game.Target{game.PermanentTarget(big.ObjectID)}, 0, nil))

	if scoreBig <= scoreSmall {
		t.Errorf("targeting the 6/6 (%v) should outscore targeting the 2/2 (%v)", scoreBig, scoreSmall)
	}
}

func TestGenericStrategyPenalizesSelfTargeting(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	removalID := addObservedHandCard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Removal",
		Types: []types.Card{types.Instant},
	}})
	// A genuine threat: instant removal is deliberately held for low-value
	// targets, so the self-vs-enemy comparison is only meaningful against a
	// target the agent would actually spend removal on.
	own := addObservedPermanent(g, game.Player1, creatureCardDef("Mine", 6, 6))
	enemy := addObservedPermanent(g, game.Player2, creatureCardDef("Theirs", 6, 6))

	obs := rules.NewObservation(g, game.Player1)
	strategy := GenericStrategy{}

	scoreOwn := strategy.ScoreAction(obs, action.CastSpell(removalID, []game.Target{game.PermanentTarget(own.ObjectID)}, 0, nil))
	scoreEnemy := strategy.ScoreAction(obs, action.CastSpell(removalID, []game.Target{game.PermanentTarget(enemy.ObjectID)}, 0, nil))

	if scoreOwn >= scoreEnemy {
		t.Errorf("self-target score %v should be worse than enemy-target score %v", scoreOwn, scoreEnemy)
	}
}

func TestAgentWithGenericStrategyPicksHighestScore(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creatureID := addObservedHandCard(g, game.Player1, creatureCardDef("Bear", 2, 2))
	obs := rules.NewObservation(g, game.Player1)

	agent := Agent{Strategy: GenericStrategy{}}
	legal := []action.Action{
		action.Pass(),
		action.CastSpell(creatureID, nil, 0, nil),
		action.PlayLandFace(g.IDGen.Next(), game.FaceFront),
	}

	got := agent.ChooseAction(obs, legal)
	if got.Kind != action.ActionPlayLand {
		t.Errorf("Agent picked %v, want ActionPlayLand (highest-scoring)", got.Kind)
	}
}

func TestGenericStrategyDeterministic(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creatureID := addObservedHandCard(g, game.Player1, creatureCardDef("Bear", 2, 2))
	obs := rules.NewObservation(g, game.Player1)
	strategy := GenericStrategy{}
	act := action.CastSpell(creatureID, nil, 0, nil)

	first := strategy.ScoreAction(obs, act)
	for range 20 {
		if again := strategy.ScoreAction(obs, act); again != first {
			t.Fatalf("ScoreAction not deterministic: %v vs %v", again, first)
		}
	}
}

func activatedArtifact(name string, costs []cost.Additional) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  name,
		Types: []types.Card{types.Artifact},
		ActivatedAbilities: []game.ActivatedAbility{{
			AdditionalCosts: costs,
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
			}}}.Ability(),
		}},
	}}
}

func sacrificeDrawArtifact(name string, sacCount int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  name,
		Types: []types.Card{types.Artifact},
		ActivatedAbilities: []game.ActivatedAbility{{
			AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalSacrifice, Amount: sacCount}},
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
			}}}.Ability(),
		}},
	}}
}

func sacrificeDestroyArtifact(name string, sacCount int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  name,
		Types: []types.Card{types.Artifact},
		ActivatedAbilities: []game.ActivatedAbility{{
			AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalSacrifice, Amount: sacCount}},
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.Destroy{},
			}}}.Ability(),
		}},
	}}
}

func modalDrawOrLoseLifeArtifact(name string) *game.CardDef {
	you := game.ControllerReference()
	return &game.CardDef{CardFace: game.CardFace{
		Name:  name,
		Types: []types.Card{types.Artifact},
		ActivatedAbilities: []game.ActivatedAbility{{
			Content: game.AbilityContent{
				MinModes: 1,
				MaxModes: 1,
				Modes: []game.Mode{
					{Sequence: []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: you}}}},
					{Sequence: []game.Instruction{{Primitive: game.LoseLife{Amount: game.Fixed(8), Player: you}}}},
				},
			},
		}},
	}}
}

func drawXArtifact(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  name,
		Types: []types.Card{types.Artifact},
		ActivatedAbilities: []game.ActivatedAbility{{
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.Draw{Amount: game.Dynamic(game.DynamicAmount{}), Player: game.ControllerReference()},
			}}}.Ability(),
		}},
	}}
}

func TestGenericStrategyScoresModalAbilityByChosenMode(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	artifact := addObservedPermanent(g, game.Player1, modalDrawOrLoseLifeArtifact("Oracle"))
	obs := rules.NewObservation(g, game.Player1)
	strategy := GenericStrategy{}

	pass := strategy.ScoreAction(obs, action.Pass())
	drawMode := strategy.ScoreAction(obs, action.ActivateAbilityWithModes(artifact.ObjectID, 0, nil, 0, []int{0}))
	lifeMode := strategy.ScoreAction(obs, action.ActivateAbilityWithModes(artifact.ObjectID, 0, nil, 0, []int{1}))

	if drawMode <= pass {
		t.Fatalf("draw-mode activation scored %v, want above pass %v", drawMode, pass)
	}
	if lifeMode >= pass {
		t.Fatalf("lose-life-mode activation scored %v, want below pass %v", lifeMode, pass)
	}
}

func TestGenericStrategyScoresDynamicDrawByAnnouncedX(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	artifact := addObservedPermanent(g, game.Player1, drawXArtifact("Engine"))
	obs := rules.NewObservation(g, game.Player1)
	strategy := GenericStrategy{}

	smallX := strategy.ScoreAction(obs, action.ActivateAbilityWithModes(artifact.ObjectID, 0, nil, 0, nil))
	largeX := strategy.ScoreAction(obs, action.ActivateAbilityWithModes(artifact.ObjectID, 0, nil, 5, nil))

	if largeX <= smallX {
		t.Fatalf("draw-X with X=5 scored %v, want above X=0 score %v", largeX, smallX)
	}
}

func TestGenericStrategyActivatesFetchlandStyleSearch(t *testing.T) {
	// A fetchland pays by sacrificing its own source (not another permanent), so
	// its land-search effect must keep it worth activating.
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	fetch := addObservedPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Fetchland",
		Types: []types.Card{types.Land},
		ActivatedAbilities: []game.ActivatedAbility{{
			AdditionalCosts: []cost.Additional{
				{Kind: cost.AdditionalSacrificeSource},
				{Kind: cost.AdditionalPayLife, Amount: 1},
			},
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.Search{Amount: game.Fixed(1), Player: game.ControllerReference()},
			}}}.Ability(),
		}},
	}})
	obs := rules.NewObservation(g, game.Player1)
	strategy := GenericStrategy{}

	pass := strategy.ScoreAction(obs, action.Pass())
	fetchScore := strategy.ScoreAction(obs, action.ActivateAbility(fetch.ObjectID, 0, nil, 0))
	if fetchScore <= pass {
		t.Fatalf("fetchland search scored %v, want above pass %v", fetchScore, pass)
	}
}

func TestGenericStrategyValuesFreeActivationAbovePass(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	tap := addObservedPermanent(g, game.Player1, activatedArtifact("Engine", cost.Tap))
	obs := rules.NewObservation(g, game.Player1)
	strategy := GenericStrategy{}

	pass := strategy.ScoreAction(obs, action.Pass())
	tapScore := strategy.ScoreAction(obs, action.ActivateAbility(tap.ObjectID, 0, nil, 0))

	if tapScore <= pass {
		t.Fatalf("free draw activation scored %v, want above pass %v", tapScore, pass)
	}
}

func TestGenericStrategySacrificeValuedAgainstEffect(t *testing.T) {
	// "Sacrifice a creature: Draw a card." is worth chumping a useless creature
	// but not feeding a real one.
	weak := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addObservedPermanent(weak, game.Player1, creatureCardDef("Goblin", 0, 1))
	weakAltar := addObservedPermanent(weak, game.Player1, sacrificeDrawArtifact("Altar", 1))
	weakObs := rules.NewObservation(weak, game.Player1)
	strategy := GenericStrategy{}
	weakPass := strategy.ScoreAction(weakObs, action.Pass())
	weakScore := strategy.ScoreAction(weakObs, action.ActivateAbility(weakAltar.ObjectID, 0, nil, 0))
	if weakScore <= weakPass {
		t.Fatalf("sacrificing a 0/1 to draw scored %v, want above pass %v", weakScore, weakPass)
	}

	strong := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addObservedPermanent(strong, game.Player1, creatureCardDef("Wurm", 10, 10))
	strongAltar := addObservedPermanent(strong, game.Player1, sacrificeDrawArtifact("Altar", 1))
	strongObs := rules.NewObservation(strong, game.Player1)
	strongPass := strategy.ScoreAction(strongObs, action.Pass())
	strongScore := strategy.ScoreAction(strongObs, action.ActivateAbility(strongAltar.ObjectID, 0, nil, 0))
	if strongScore >= strongPass {
		t.Fatalf("sacrificing a 10/10 to draw scored %v, want below pass %v", strongScore, strongPass)
	}
}

func TestGenericStrategySacrificesWeakToRemoveThreat(t *testing.T) {
	// "Sacrifice three creatures: Destroy target creature." Worth three useless
	// 1/1s to kill a 10/10, but not three 5/5s to kill a 1/1.
	strategy := GenericStrategy{}

	worth := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	for range 3 {
		addObservedPermanent(worth, game.Player1, creatureCardDef("Soldier", 1, 1))
	}
	bigThreat := addObservedPermanent(worth, game.Player2, creatureCardDef("Wurm", 10, 10))
	worthAltar := addObservedPermanent(worth, game.Player1, sacrificeDestroyArtifact("Altar", 3))
	worthObs := rules.NewObservation(worth, game.Player1)
	worthPass := strategy.ScoreAction(worthObs, action.Pass())
	worthTargets := []game.Target{{Kind: game.TargetPermanent, PermanentID: bigThreat.ObjectID}}
	worthScore := strategy.ScoreAction(worthObs, action.ActivateAbility(worthAltar.ObjectID, 0, worthTargets, 0))
	if worthScore <= worthPass {
		t.Fatalf("sacrificing three 1/1s to kill a 10/10 scored %v, want above pass %v", worthScore, worthPass)
	}

	wasteful := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	for range 3 {
		addObservedPermanent(wasteful, game.Player1, creatureCardDef("Bear", 5, 5))
	}
	smallThreat := addObservedPermanent(wasteful, game.Player2, creatureCardDef("Mouse", 1, 1))
	wastefulAltar := addObservedPermanent(wasteful, game.Player1, sacrificeDestroyArtifact("Altar", 3))
	wastefulObs := rules.NewObservation(wasteful, game.Player1)
	wastefulPass := strategy.ScoreAction(wastefulObs, action.Pass())
	wastefulTargets := []game.Target{{Kind: game.TargetPermanent, PermanentID: smallThreat.ObjectID}}
	wastefulScore := strategy.ScoreAction(wastefulObs, action.ActivateAbility(wastefulAltar.ObjectID, 0, wastefulTargets, 0))
	if wastefulScore >= wastefulPass {
		t.Fatalf("sacrificing three 5/5s to kill a 1/1 scored %v, want below pass %v", wastefulScore, wastefulPass)
	}
}
