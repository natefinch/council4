package agent

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/rules"
)

func creatureWithKeywords(name string, power, toughness int, keywords ...game.Keyword) *game.CardDef {
	def := creatureCardDef(name, power, toughness)
	def.StaticAbilities = []game.StaticAbility{{KeywordAbilities: game.SimpleKeywords(keywords...)}}
	return def
}

func TestPermanentThreatRanksDangerousCreaturesHigher(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	vanilla := addObservedPermanent(g, game.Player2, creatureCardDef("Vanilla", 3, 3))
	flyer := addObservedPermanent(g, game.Player2, creatureWithKeywords("Flyer", 3, 3, game.Flying))
	bigger := addObservedPermanent(g, game.Player2, creatureCardDef("Bigger", 6, 6))
	rock := addObservedPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Mana Rock",
		Types: []types.Card{types.Artifact},
	}})

	obs := rules.NewObservation(g, game.Player1)
	views := permanentViewMap(obs)

	if permanentThreat(views[flyer.ObjectID]) <= permanentThreat(views[vanilla.ObjectID]) {
		t.Error("a flyer should be more threatening than a vanilla creature of equal stats")
	}
	if permanentThreat(views[bigger.ObjectID]) <= permanentThreat(views[vanilla.ObjectID]) {
		t.Error("a bigger creature should be more threatening")
	}
	if permanentThreat(views[rock.ObjectID]) >= permanentThreat(views[vanilla.ObjectID]) {
		t.Error("a noncreature should be less threatening than a creature")
	}
}

func TestPermanentThreatTappedIsLessImmediate(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	untapped := addObservedPermanent(g, game.Player2, creatureCardDef("Untapped", 4, 4))
	tapped := addObservedPermanent(g, game.Player2, creatureCardDef("Tapped", 4, 4))
	tapped.Tapped = true

	obs := rules.NewObservation(g, game.Player1)
	views := permanentViewMap(obs)

	if permanentThreat(views[tapped.ObjectID]) >= permanentThreat(views[untapped.ObjectID]) {
		t.Error("a tapped creature should be a less immediate threat than the same untapped creature")
	}
}

func TestThreatModelRanksOpponentsByBoard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addObservedPermanent(g, game.Player2, creatureCardDef("Small", 1, 1))
	addObservedPermanent(g, game.Player3, creatureCardDef("Huge", 8, 8))

	model := NewThreatModel(rules.NewObservation(g, game.Player1))

	if model.PlayerThreat(game.Player3) <= model.PlayerThreat(game.Player2) {
		t.Error("the opponent with the bigger board should be the bigger threat")
	}
	biggest, _, ok := model.HighestThreatOpponent()
	if !ok || biggest != game.Player3 {
		t.Errorf("HighestThreatOpponent = (%v, ok=%v), want Player3", biggest, ok)
	}
}

func TestThreatModelIgnoresPhasedOutBoard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	// Player2 phased their whole board out (e.g. Teferi's Protection); Player3
	// has a real board. Player2 must not be scored as the bigger threat.
	phased := addObservedPermanent(g, game.Player2, creatureCardDef("Phased", 9, 9))
	phased.PhasedOut = true
	addObservedPermanent(g, game.Player3, creatureCardDef("Real", 3, 3))

	model := NewThreatModel(rules.NewObservation(g, game.Player1))

	if model.PlayerThreat(game.Player2) >= model.PlayerThreat(game.Player3) {
		t.Errorf("a phased-out board should not outscore a real board: P2 %v vs P3 %v",
			model.PlayerThreat(game.Player2), model.PlayerThreat(game.Player3))
	}
	biggest, _, ok := model.HighestThreatOpponent()
	if !ok || biggest != game.Player3 {
		t.Errorf("HighestThreatOpponent = (%v, ok=%v), want Player3 with the real board", biggest, ok)
	}
}

func TestThreatModelExcludesObserverAndEliminated(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addObservedPermanent(g, game.Player1, creatureCardDef("Mine", 9, 9))
	addObservedPermanent(g, game.Player4, creatureCardDef("Theirs", 2, 2))
	g.Players[game.Player2].Eliminated = true
	g.Players[game.Player3].Eliminated = true

	model := NewThreatModel(rules.NewObservation(g, game.Player1))

	biggest, _, ok := model.HighestThreatOpponent()
	if !ok || biggest != game.Player4 {
		t.Errorf("HighestThreatOpponent = (%v, ok=%v), want the only living opponent Player4", biggest, ok)
	}
}

func TestGenericStrategyRemovalPrefersEvasiveThreat(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	removalID := addObservedHandCard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Removal",
		Types: []types.Card{types.Instant},
	}})
	ground := addObservedPermanent(g, game.Player2, creatureCardDef("Ground", 3, 3))
	flyer := addObservedPermanent(g, game.Player2, creatureWithKeywords("Flyer", 3, 3, game.Flying))

	obs := rules.NewObservation(g, game.Player1)
	strategy := GenericStrategy{}

	scoreGround := strategy.ScoreAction(obs, action.CastSpell(removalID, []game.Target{game.PermanentTarget(ground.ObjectID)}, 0, nil))
	scoreFlyer := strategy.ScoreAction(obs, action.CastSpell(removalID, []game.Target{game.PermanentTarget(flyer.ObjectID)}, 0, nil))

	if scoreFlyer <= scoreGround {
		t.Errorf("removal should prefer the evasive threat: flyer %v vs ground %v", scoreFlyer, scoreGround)
	}
}

func TestGenericStrategyBurnAvoidsKingmaking(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	burnID := addObservedHandCard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Burn",
		Types: []types.Card{types.Sorcery},
	}})
	// Player2 is a real threat; Player3 has an empty board (near-irrelevant).
	addObservedPermanent(g, game.Player2, creatureCardDef("Threat", 7, 7))

	obs := rules.NewObservation(g, game.Player1)
	strategy := GenericStrategy{}

	scoreThreat := strategy.ScoreAction(obs, action.CastSpell(burnID, []game.Target{game.PlayerTarget(game.Player2)}, 0, nil))
	scoreWeak := strategy.ScoreAction(obs, action.CastSpell(burnID, []game.Target{game.PlayerTarget(game.Player3)}, 0, nil))

	if scoreThreat <= scoreWeak {
		t.Errorf("burn should target the bigger threat (Player2 %v) over the weak board (Player3 %v)", scoreThreat, scoreWeak)
	}
}

func permanentViewMap(obs rules.PlayerObservation) map[id.ID]rules.PermanentView {
	views := map[id.ID]rules.PermanentView{}
	for _, view := range obs.Battlefield() {
		views[view.ObjectID] = view
	}
	return views
}
