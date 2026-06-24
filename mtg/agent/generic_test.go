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

func TestGenericStrategyAvoidsResourceSpendingActivations(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	sacCosts := append(append([]cost.Additional(nil), cost.Tap...),
		cost.Additional{Kind: cost.AdditionalSacrifice, Amount: 1})
	sac := addObservedPermanent(g, game.Player1, activatedArtifact("Altar", sacCosts))
	tap := addObservedPermanent(g, game.Player1, activatedArtifact("Engine", cost.Tap))
	obs := rules.NewObservation(g, game.Player1)
	strategy := GenericStrategy{}

	pass := strategy.ScoreAction(obs, action.Pass())
	sacScore := strategy.ScoreAction(obs, action.ActivateAbility(sac.ObjectID, 0, nil, 0))
	tapScore := strategy.ScoreAction(obs, action.ActivateAbility(tap.ObjectID, 0, nil, 0))

	if sacScore >= pass {
		t.Fatalf("resource-spending activation scored %v, want below pass %v", sacScore, pass)
	}
	if tapScore <= pass {
		t.Fatalf("plain activation scored %v, want above pass %v", tapScore, pass)
	}
}
