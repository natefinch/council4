package rules

import (
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Source review (#3447, with the land-play counterpart #3406; Engine gap, no
// card admitted; blocks Maralen, Fae Ascendant 114e1d2c-cefc-4e2a-922e-3b2ceed1871d):
//   - Evelyn, the Covetous (oracle 4ab6f615-1f04-4ed2-88b2-49faa8356dd0):
//     "Once each turn, you may play a card from exile with a collection counter
//     on it if it was exiled by an ability you controlled, and you may spend
//     mana as though it were mana of any color to cast it." No Scryfall rulings.
//   - Sakashima of a Thousand Faces (oracle 8ecdaf4b-4442-42da-9714-4257a83faf50):
//     "The \"legend rule\" doesn't apply to permanents you control." Ruling
//     2020-11-10: you can control any number of legendary permanents with the
//     same name and none of them will be put into the graveyard.
//   - Vandalblast (oracle 3567c3c8-b3c7-45b7-935b-b1fdbc973720): "Destroy target
//     artifact you don't control. Overload {4}{R}".
//   - Maralen, Fae Ascendant: "Once each turn, you may cast a spell ... from
//     among cards exiled with Maralen this turn without paying its mana cost."
//     Two Maralens likewise offer same-named once-each-turn permissions.
//   - CR (2026-09-25) 601.2b: the player announces alternative costs and how the
//     spell is cast; previously made choices restrict later ones. 601.2f: total
//     cost from the chosen mana or alternative cost. 704.5j: the legend rule.
//     116.2a/305.1: a land play is a special action. Each once-each-turn
//     permission is its own source's use, so the choice must name which
//     same-named permanent's use is spent.

// sourcePermanentAgent answers the choice with prompt by selecting the option
// whose SourcePermanent is source, recording every option it was offered.
type sourcePermanentAgent struct {
	prompt  string
	source  id.ID
	offered *[]game.ChoiceOption
}

func (sourcePermanentAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a sourcePermanentAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Prompt != a.prompt {
		return request.DefaultSelection
	}
	*a.offered = append(*a.offered, request.Options...)
	for _, option := range request.Options {
		if option.SourcePermanent == a.source {
			return []int{option.Index}
		}
	}
	return request.DefaultSelection
}

// twoRegisteredEvelyns puts registered Sakashima of a Thousand Faces and two
// registered Evelyn, the Covetous under Player1's control and checks that
// state-based actions keep both Evelyns, since the legend rule doesn't apply.
func twoRegisteredEvelyns(t *testing.T, g *game.Game, engine *Engine) (first, second *game.Permanent) {
	t.Helper()
	addCombatPermanent(g, game.Player1, registryCardDef(t, "Sakashima of a Thousand Faces"))
	first = addCombatPermanent(g, game.Player1, registryCardDef(t, "Evelyn, the Covetous"))
	second = addCombatPermanent(g, game.Player1, registryCardDef(t, "Evelyn, the Covetous"))
	engine.applyStateBasedActions(g)
	for _, evelyn := range []*game.Permanent{first, second} {
		if _, ok := permanentByObjectID(g, evelyn.ObjectID); !ok {
			t.Fatal("the legend rule removed an Evelyn while Sakashima is on the battlefield")
		}
	}
	return first, second
}

// assertSourceOptions checks that the offered options name each Evelyn exactly
// once, by SourcePermanent and by a distinct label carrying its object ID.
func assertSourceOptions(t *testing.T, offered []game.ChoiceOption, evelyns ...*game.Permanent) {
	t.Helper()
	if len(offered) != len(evelyns) {
		t.Fatalf("offered %+v, want one option per Evelyn", offered)
	}
	labels := map[string]bool{}
	for _, evelyn := range evelyns {
		index := slices.IndexFunc(offered, func(option game.ChoiceOption) bool { return option.SourcePermanent == evelyn.ObjectID })
		if index < 0 {
			t.Fatalf("offered %+v, want an option sourced from Evelyn %d", offered, evelyn.ObjectID)
		}
		label := offered[index].Label
		if !strings.Contains(label, "Evelyn, the Covetous ["+strconv.FormatUint(uint64(evelyn.ObjectID), 10)+"]") {
			t.Fatalf("option label %q does not identify Evelyn %d", label, evelyn.ObjectID)
		}
		labels[label] = true
	}
	if len(labels) != len(evelyns) {
		t.Fatalf("offered %+v, want distinct labels", offered)
	}
}

// TestExileCastPermissionChoiceIdentifiesSameNamedSources casts registered
// Vandalblast from exile with one blue mana, which only an Evelyn's any-color
// rider can pay. Two registered Evelyns (legal under Sakashima) both authorize
// it, so the cost choice offers one "Normal cost" option per Evelyn; each names
// its own Evelyn by object ID and carries it as SourcePermanent. Choosing
// either spends only that Evelyn's once-each-turn use, in a clone independent
// of the original, and the other Evelyn still authorizes a later cast.
func TestExileCastPermissionChoiceIdentifiesSameNamedSources(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	first, second := twoRegisteredEvelyns(t, g, engine)
	artifact := addOverloadArtifact(g, game.Player2, false)
	spellID := collectionExiledCard(g, registryCardDef(t, "Vandalblast"))
	laterID := collectionExiledCard(g, collectedSorcery())
	giveMainPhasePriority(g, game.Player1)
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	cast := action.CastSpellFaceFromZone(spellID, zone.Exile, game.FaceFront, []game.Target{game.PermanentTarget(artifact.ObjectID)}, 0, nil)
	assertExileCastListed(t, g, engine, cast, true)

	for _, chosen := range []*game.Permanent{first, second} {
		other := first
		if chosen == first {
			other = second
		}
		clone := g.Clone()
		var offered []game.ChoiceOption
		log := &TurnLog{}
		agents := [game.NumPlayers]PlayerAgent{game.Player1: sourcePermanentAgent{prompt: "Choose spell cost", source: chosen.ObjectID, offered: &offered}}
		if !engine.applyActionWithChoices(clone, game.Player1, cast, agents, log) {
			t.Fatalf("casting Vandalblast through Evelyn %d was rejected", chosen.ObjectID)
		}
		assertSourceOptions(t, offered, first, second)
		for _, option := range offered {
			if !strings.HasPrefix(option.Label, "Normal cost (via Evelyn, the Covetous [") {
				t.Fatalf("option label %q, want the normal cost through an Evelyn", option.Label)
			}
		}
		if index := slices.IndexFunc(log.Choices, func(decision game.ChoiceDecision) bool {
			return decision.Request.Prompt == "Choose spell cost"
		}); index < 0 || len(log.Choices[index].Request.Options) != 2 ||
			log.Choices[index].Request.Options[log.Choices[index].Selected[0]].SourcePermanent != chosen.ObjectID {
			t.Fatalf("logged choices %+v, want the recorded selection of Evelyn %d", log.Choices, chosen.ObjectID)
		}
		if obj, ok := clone.Stack.Peek(); !ok || obj.SourceID != spellID || obj.ManaSpentToCast != 1 {
			t.Fatalf("stack object = %+v, want Vandalblast paid with one blue mana", obj)
		}
		if !clone.ExilePlayPermissionUsedThisTurn[chosen.ObjectID] || clone.ExilePlayPermissionUsedThisTurn[other.ObjectID] {
			t.Fatalf("used = %v, want only the chosen Evelyn %d spent", clone.ExilePlayPermissionUsedThisTurn, chosen.ObjectID)
		}

		engine.resolveTopOfStack(clone, &TurnLog{})
		giveMainPhasePriority(clone, game.Player1)
		clone.Players[game.Player1].ManaPool.Add(mana.U, 1)
		if !castCollectedSorcery(clone, engine, laterID) {
			t.Fatalf("the unchosen Evelyn %d could not authorize a later cast", other.ObjectID)
		}
		if !clone.ExilePlayPermissionUsedThisTurn[other.ObjectID] {
			t.Fatal("the later cast did not spend the remaining Evelyn's use")
		}
	}
	if len(g.ExilePlayPermissionUsedThisTurn) != 0 || !g.Players[game.Player2].Exile.Contains(spellID) {
		t.Fatal("casting in a clone changed the original game")
	}
}

// TestMutateCastPermissionChoiceIdentifiesSameNamedSources casts an
// Adventure-exiled Mutate creature with one blue mana through either of two
// same-named Evelyn permissions; the Mutate route offers the same identified
// options and spends only the chosen source's use.
func TestMutateCastPermissionChoiceIdentifiesSameNamedSources(t *testing.T) {
	t.Parallel()
	for _, chooseSecond := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		first := evelynCovetousPermanent(g, game.Player1)
		second := evelynCovetousPermanent(g, game.Player1)
		chosen, other := first, second
		if chooseSecond {
			chosen, other = second, first
		}
		target := addCombatPermanent(g, game.Player1, mutateTargetCard())
		cardID := addCardToExile(g, game.Player1, mutatingAdventureFixture())
		g.AdventureCards[cardID] = game.AdventureExile{Player: game.Player2, ZoneVersion: g.CardInstances[cardID].ZoneVersion}
		g.AddExileCounterFromController(cardID, counter.Collection, 1, game.Player1)
		g.Players[game.Player1].ManaPool.Add(mana.U, 1)
		giveMainPhasePriority(g, game.Player1)

		var offered []game.ChoiceOption
		agents := [game.NumPlayers]PlayerAgent{game.Player1: sourcePermanentAgent{prompt: "Choose spell cost", source: chosen.ObjectID, offered: &offered}}
		if !engine.applyActionWithChoices(g, game.Player1, action.CastMutateSpellFromZone(cardID, zone.Exile, target.ObjectID), agents, &TurnLog{}) {
			t.Fatalf("chooseSecond=%v: Mutate cast rejected", chooseSecond)
		}
		if len(offered) != 2 || offered[0].SourcePermanent == offered[1].SourcePermanent || offered[0].Label == offered[1].Label ||
			!slices.ContainsFunc(offered, func(option game.ChoiceOption) bool { return option.SourcePermanent == chosen.ObjectID }) {
			t.Fatalf("chooseSecond=%v: offered %+v, want one identified option per Evelyn", chooseSecond, offered)
		}
		if obj, ok := g.Stack.Peek(); !ok || obj.SourceID != cardID || !obj.Mutate || obj.ManaSpentToCast != 1 {
			t.Fatalf("chooseSecond=%v: stack object = %+v", chooseSecond, obj)
		}
		if !g.ExilePlayPermissionUsedThisTurn[chosen.ObjectID] || g.ExilePlayPermissionUsedThisTurn[other.ObjectID] {
			t.Fatalf("chooseSecond=%v: used = %v, want only the chosen Evelyn spent", chooseSecond, g.ExilePlayPermissionUsedThisTurn)
		}
	}
}

// TestLandPlayPermissionChoiceIdentifiesSameNamedSources plays a
// collection-counter land that two registered Evelyns (legal under Sakashima)
// both cover. Each land-play option names its own Evelyn by object ID and
// carries it as SourcePermanent; choosing either spends only that Evelyn's use.
func TestLandPlayPermissionChoiceIdentifiesSameNamedSources(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	first, second := twoRegisteredEvelyns(t, g, engine)
	landID := exiledLandOwnedBy(g, game.Player1, "Island")
	g.AddExileCounterFromController(landID, counter.Collection, 1, game.Player1)
	giveMainPhasePriority(g, game.Player1)

	for _, chosen := range []*game.Permanent{first, second} {
		other := first
		if chosen == first {
			other = second
		}
		clone := g.Clone()
		var offered []game.ChoiceOption
		agents := [game.NumPlayers]PlayerAgent{game.Player1: sourcePermanentAgent{prompt: "Choose land play permission", source: chosen.ObjectID, offered: &offered}}
		assertLandPlayedFromExile(t, clone, engine, landID, agents)
		assertSourceOptions(t, offered, first, second)
		if !clone.ExilePlayPermissionUsedThisTurn[chosen.ObjectID] || clone.ExilePlayPermissionUsedThisTurn[other.ObjectID] {
			t.Fatalf("used = %v, want only the chosen Evelyn %d spent", clone.ExilePlayPermissionUsedThisTurn, chosen.ObjectID)
		}
	}
	if len(g.ExilePlayPermissionUsedThisTurn) != 0 || !g.Players[game.Player1].Exile.Contains(landID) {
		t.Fatal("playing the land in a clone changed the original game")
	}
}

// TestPermissionSourcesNumberSameNamedNonPermanentSources names permission
// sources that are no longer on the battlefield: distinct names stay plain,
// while same-named cards with no battlefield object are numbered, and a
// sourceless permission is numbered.
func TestPermissionSourcesNumberSameNamedNonPermanentSources(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	first := addCardToExile(g, game.Player1, collectedSorcery())
	second := addCardToExile(g, game.Player1, collectedSorcery())
	land := exiledLandOwnedBy(g, game.Player1, "Island")
	sources := permissionSources(g, []game.RuleEffect{
		{SourceCardID: first},
		{SourceCardID: second},
		{SourceCardID: land},
		{},
	})
	want := []string{
		"via Collected Sorcery (permission 1)",
		"via Collected Sorcery (permission 2)",
		"via Island",
		"permission 4",
	}
	for i, source := range sources {
		if source.label != want[i] || source.permanent != 0 {
			t.Fatalf("sources = %+v, want labels %q with no permanent", sources, want)
		}
	}
}
