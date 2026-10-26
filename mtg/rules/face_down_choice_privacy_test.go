package rules

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: CR 708.2/708.2a (a face-down permanent has only the
// characteristics its face-down rules list; morph gives a nameless, colorless
// 2/2 creature with no subtypes and no mana cost) and CR 708.5 (only the
// controller may look at it). Registered Exalted Angel (morph), Arc Lightning
// ("3 damage divided as you choose among one, two, or three targets") and
// Diabolic Edict ("Target player sacrifices a creature of their choice") and
// Birthday Escape ("Draw a card. The Ring tempts you.") drive the choices.
// Engine gap #3419.

type faceDownChoiceRecorder struct {
	simPassPolicy

	selections [][]int
	requests   []game.ChoiceRequest
}

func (r *faceDownChoiceRecorder) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	r.requests = append(r.requests, request)
	if len(r.selections) > 0 {
		selected := r.selections[0]
		r.selections = r.selections[1:]
		return selected
	}
	return request.DefaultSelection
}

// assertFaceDownChoicePublic checks that request describes each face-down
// permanent only by its public characteristics: a label naming it as a
// face-down permanent by object ID and a nameless, colorless creature with mana
// value 0, while each face-up permanent keeps its printed name.
func assertFaceDownChoicePublic(t *testing.T, request game.ChoiceRequest, hiddenName string, faceDown []*game.Permanent, faceUp map[*game.Permanent]string) {
	t.Helper()
	if strings.Contains(request.Prompt, hiddenName) {
		t.Fatalf("prompt %q reveals a face-down card", request.Prompt)
	}
	byCard := map[string]game.ChoiceOption{}
	for _, option := range request.Options {
		if strings.Contains(option.Label, hiddenName) {
			t.Fatalf("option label %q reveals a face-down card", option.Label)
		}
		if !option.Card.Exists {
			t.Fatalf("option %q has no card info", option.Label)
		}
		if option.Card.Val.Name == hiddenName {
			t.Fatalf("option card info %+v reveals a face-down card", option.Card.Val)
		}
		byCard[fmt.Sprint(option.Card.Val.CardID)] = option
	}
	for _, permanent := range faceDown {
		option, ok := byCard[fmt.Sprint(permanent.CardInstanceID)]
		if !ok {
			t.Fatalf("face-down permanent %d not offered in %+v", permanent.ObjectID, request.Options)
		}
		if want := fmt.Sprintf("Face-down permanent %d", permanent.ObjectID); option.Label != want {
			t.Fatalf("face-down label = %q, want %q", option.Label, want)
		}
		info := option.Card.Val
		if info.Name != "" || len(info.Colors) != 0 || info.ManaValue != 0 || !slices.Equal(info.Types, []types.Card{types.Creature}) {
			t.Fatalf("face-down info = %+v, want a nameless colorless creature with mana value 0", info)
		}
	}
	for permanent, name := range faceUp {
		option, ok := byCard[fmt.Sprint(permanent.CardInstanceID)]
		if !ok {
			t.Fatalf("face-up %s not offered in %+v", name, request.Options)
		}
		if option.Label != name || option.Card.Val.Name != name {
			t.Fatalf("face-up option = %q/%+v, want %s", option.Label, option.Card.Val, name)
		}
	}
}

// TestDividedTargetDamageChoiceHidesOpponentFaceDownPermanent casts registered
// Arc Lightning at an opponent's face-down Exalted Angel and Grizzly Bears. The
// resolution-time division over those targets, and its logged copy, describe
// the morph only as a face-down 2/2, and the chosen division is still honored.
func TestDividedTargetDamageChoiceHidesOpponentFaceDownPermanent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	angel := addFaceDownPermanent(g, game.Player2, registryCardDef(t, "Exalted Angel"), game.FaceDownMorph)
	bears := addCombatPermanent(g, game.Player2, registryCardDef(t, "Grizzly Bears"))
	arcID := addCardToHand(g, game.Player1, registryCardDef(t, "Arc Lightning"))
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(arcID, []game.Target{
		game.PermanentTarget(angel.ObjectID),
		game.PermanentTarget(bears.ObjectID),
	}, 0, nil)) {
		t.Fatal("Arc Lightning cast failed")
	}
	caster := &faceDownChoiceRecorder{selections: [][]int{{0, 0, 1}}}
	agents := simPassPolicies()
	agents[game.Player1] = caster
	log := &TurnLog{}

	engine.resolveTopOfStackWithChoices(g, agents, log)

	if len(caster.requests) != 1 || len(log.Choices) != 1 {
		t.Fatalf("requests/logged = %d/%d, want one damage division", len(caster.requests), len(log.Choices))
	}
	for _, request := range []game.ChoiceRequest{caster.requests[0], log.Choices[0].Request} {
		if request.Kind != game.ChoiceDamageAllocation {
			t.Fatalf("request kind = %v, want ChoiceDamageAllocation", request.Kind)
		}
		for _, option := range request.Options {
			if strings.Contains(option.Label, "Exalted Angel") {
				t.Fatalf("division label %q reveals the face-down card", option.Label)
			}
		}
		labels := []string{request.Options[0].Label, request.Options[1].Label}
		if want := []string{fmt.Sprintf("Face-down permanent %d", angel.ObjectID), "Grizzly Bears"}; !slices.Equal(labels, want) {
			t.Fatalf("division labels = %q, want %q", labels, want)
		}
	}
	if angel.MarkedDamage != 2 || bears.MarkedDamage != 1 {
		t.Fatalf("marked damage angel/bears = %d/%d, want the chosen 2/1", angel.MarkedDamage, bears.MarkedDamage)
	}
}

// TestSacrificeChoiceHidesFaceDownPermanents resolves registered Diabolic Edict
// against an opponent who controls two face-down Exalted Angels and Grizzly
// Bears. The sacrifice choice, and its logged copy, tell the face-down
// permanents apart by object ID without revealing the card underneath; the
// card ID lets their controller match the private face-down view it may look
// at (CR 708.5). The chosen face-down permanent is the one sacrificed.
func TestSacrificeChoiceHidesFaceDownPermanents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	first := addFaceDownPermanent(g, game.Player2, registryCardDef(t, "Exalted Angel"), game.FaceDownMorph)
	second := addFaceDownPermanent(g, game.Player2, registryCardDef(t, "Exalted Angel"), game.FaceDownMorph)
	bears := addCombatPermanent(g, game.Player2, registryCardDef(t, "Grizzly Bears"))
	edictID := addCardToHand(g, game.Player1, registryCardDef(t, "Diabolic Edict"))
	g.Players[game.Player1].ManaPool.Add(mana.B, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(edictID, []game.Target{game.PlayerTarget(game.Player2)}, 0, nil)) {
		t.Fatal("Diabolic Edict cast failed")
	}
	opponent := &faceDownChoiceRecorder{}
	agents := simPassPolicies()
	agents[game.Player2] = opponent
	log := &TurnLog{}

	engine.resolveTopOfStackWithChoices(g, agents, log)

	if len(opponent.requests) != 1 || len(log.Choices) != 1 {
		t.Fatalf("requests/logged = %d/%d, want one sacrifice choice", len(opponent.requests), len(log.Choices))
	}
	var secondIndex int
	for _, request := range []game.ChoiceRequest{opponent.requests[0], log.Choices[0].Request} {
		if request.Kind != game.ChoicePayment || request.Player != game.Player2 || len(request.Options) != 3 {
			t.Fatalf("request = %+v, want the opponent's three-way sacrifice choice", request)
		}
		assertFaceDownChoicePublic(t, request, "Exalted Angel", []*game.Permanent{first, second}, map[*game.Permanent]string{bears: "Grizzly Bears"})
		secondIndex = slices.IndexFunc(request.Options, func(option game.ChoiceOption) bool {
			return option.Card.Val.CardID == second.CardInstanceID
		})
	}
	obs := observe(g, game.Player2)
	if view := findPermanentView(t, obs, second.ObjectID); len(view.FaceDownCards) != 1 || view.FaceDownCards[0].Name != "Exalted Angel" {
		t.Fatalf("controller's private face-down view = %+v, want Exalted Angel", view.FaceDownCards)
	}
	if view := findPermanentView(t, observe(g, game.Player1), second.ObjectID); len(view.FaceDownCards) != 0 {
		t.Fatalf("opponent sees face-down cards %+v", view.FaceDownCards)
	}

	// Replay with the controller choosing the second face-down permanent.
	g2 := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	first = addFaceDownPermanent(g2, game.Player2, registryCardDef(t, "Exalted Angel"), game.FaceDownMorph)
	second = addFaceDownPermanent(g2, game.Player2, registryCardDef(t, "Exalted Angel"), game.FaceDownMorph)
	bears = addCombatPermanent(g2, game.Player2, registryCardDef(t, "Grizzly Bears"))
	edictID = addCardToHand(g2, game.Player1, registryCardDef(t, "Diabolic Edict"))
	g2.Players[game.Player1].ManaPool.Add(mana.B, 1)
	g2.Players[game.Player1].ManaPool.Add(mana.C, 1)
	setMainPhasePriority(g2, game.Player1)
	if !engine.applyAction(g2, game.Player1, action.CastSpell(edictID, []game.Target{game.PlayerTarget(game.Player2)}, 0, nil)) {
		t.Fatal("Diabolic Edict replay cast failed")
	}
	agents[game.Player2] = &faceDownChoiceRecorder{selections: [][]int{{secondIndex}}}
	engine.resolveTopOfStackWithChoices(g2, agents, &TurnLog{})
	if _, ok := permanentByObjectID(g2, second.ObjectID); ok {
		t.Fatal("chosen face-down permanent was not sacrificed")
	}
	if _, ok := permanentByObjectID(g2, first.ObjectID); !ok {
		t.Fatal("unchosen face-down permanent was sacrificed")
	}
	if _, ok := permanentByObjectID(g2, bears.ObjectID); !ok {
		t.Fatal("Grizzly Bears was sacrificed")
	}
}

// TestPermanentChoiceHelpersUseEffectiveCharacteristics pins the shared
// helpers directly: a face-down permanent is a nameless colorless creature with
// mana value 0 labeled by object ID, and the same card face up keeps its
// printed name, colors, types and mana value.
func TestPermanentChoiceHelpersUseEffectiveCharacteristics(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	hidden := addFaceDownPermanent(g, game.Player2, registryCardDef(t, "Exalted Angel"), game.FaceDownMorph)
	shown := addCombatPermanent(g, game.Player2, registryCardDef(t, "Exalted Angel"))

	if got, want := permanentChoiceLabel(g, hidden), fmt.Sprintf("Face-down permanent %d", hidden.ObjectID); got != want {
		t.Fatalf("face-down label = %q, want %q", got, want)
	}
	info := permanentChoiceInfo(g, hidden)
	if !info.Exists || info.Val.Name != "" || len(info.Val.Colors) != 0 || info.Val.ManaValue != 0 || !slices.Equal(info.Val.Types, []types.Card{types.Creature}) || info.Val.CardID != hidden.CardInstanceID {
		t.Fatalf("face-down info = %+v", info)
	}

	if got := permanentChoiceLabel(g, shown); got != "Exalted Angel" {
		t.Fatalf("face-up label = %q, want Exalted Angel", got)
	}
	def := registryCardDef(t, "Exalted Angel")
	want := opt.Val(game.ChoiceCardInfo{
		CardID: shown.CardInstanceID, Name: "Exalted Angel", Types: def.Types,
		ManaValue: def.ManaValue(), Colors: def.Colors,
	})
	if got := permanentChoiceInfo(g, shown); !got.Exists || got.Val.Name != want.Val.Name || got.Val.ManaValue != want.Val.ManaValue ||
		!slices.Equal(got.Val.Types, want.Val.Types) || !slices.Equal(got.Val.Colors, want.Val.Colors) || got.Val.CardID != want.Val.CardID {
		t.Fatalf("face-up info = %+v, want %+v", got, want)
	}
}

// TestRingBearerChoiceLabelsFaceDownPermanentsDistinctly resolves registered
// Birthday Escape ("Draw a card. The Ring tempts you.") for a player who
// controls two face-down Exalted Angels and Grizzly Bears. The Ring-bearer
// choice, and its logged copy, label each face-down creature by
// object ID instead of an empty or printed name, and the chosen face-down
// creature becomes the Ring-bearer.
func TestRingBearerChoiceLabelsFaceDownPermanentsDistinctly(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	first := addFaceDownPermanent(g, game.Player1, registryCardDef(t, "Exalted Angel"), game.FaceDownMorph)
	second := addFaceDownPermanent(g, game.Player1, registryCardDef(t, "Exalted Angel"), game.FaceDownMorph)
	bears := addCombatPermanent(g, game.Player1, registryCardDef(t, "Grizzly Bears"))
	addCardToLibrary(g, game.Player1, registryCardDef(t, "Grizzly Bears"))
	escapeID := addCardToHand(g, game.Player1, registryCardDef(t, "Birthday Escape"))
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(escapeID, nil, 0, nil)) {
		t.Fatal("Birthday Escape cast failed")
	}
	controller := &faceDownChoiceRecorder{selections: [][]int{{1}}}
	agents := simPassPolicies()
	agents[game.Player1] = controller
	log := &TurnLog{}

	engine.resolveTopOfStackWithChoices(g, agents, log)

	if len(controller.requests) != 1 || len(log.Choices) != 1 {
		t.Fatalf("requests/logged = %d/%d, want one Ring-bearer choice", len(controller.requests), len(log.Choices))
	}
	for _, request := range []game.ChoiceRequest{controller.requests[0], log.Choices[0].Request} {
		if request.Kind != game.ChoiceResolution || request.Player != game.Player1 || len(request.Options) != 3 {
			t.Fatalf("request = %+v, want the controller's three-way Ring-bearer choice", request)
		}
		assertFaceDownChoicePublic(t, request, "Exalted Angel", []*game.Permanent{first, second}, map[*game.Permanent]string{bears: "Grizzly Bears"})
		if request.Options[1].Card.Val.CardID != second.CardInstanceID {
			t.Fatalf("option 1 = %+v, want the second face-down creature", request.Options[1])
		}
	}
	if got := g.Players[game.Player1].RingBearerID; got != second.ObjectID {
		t.Fatalf("RingBearerID = %v, want chosen face-down creature %v", got, second.ObjectID)
	}
}

// TestPermanentChoiceRoutesLabelFaceDownPermanentsDistinctly covers the other
// permanent choices that label options by name: the ninjutsu attacker to
// return, the triggering-batch creature to copy, and a resolution-time
// permanent choice. Each labels two face-down creatures by object ID (CR
// 708.2a), never with an empty or printed name, and honors the choice.
func TestPermanentChoiceRoutesLabelFaceDownPermanentsDistinctly(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	first := addFaceDownPermanent(g, game.Player1, registryCardDef(t, "Exalted Angel"), game.FaceDownMorph)
	second := addFaceDownPermanent(g, game.Player1, registryCardDef(t, "Exalted Angel"), game.FaceDownMorph)
	candidates := []*game.Permanent{first, second}
	wantLabels := []string{
		fmt.Sprintf("Face-down permanent %d", first.ObjectID),
		fmt.Sprintf("Face-down permanent %d", second.ObjectID),
	}
	labelsOf := func(options []game.ChoiceOption) []string {
		labels := make([]string, len(options))
		for i, option := range options {
			labels[i] = option.Label
		}
		return labels
	}

	t.Run("ninjutsu attacker", func(t *testing.T) {
		recorder := &faceDownChoiceRecorder{selections: [][]int{{1}}}
		agents := simPassPolicies()
		agents[game.Player1] = recorder
		log := &TurnLog{}
		if got := chooseNinjutsuAttacker(engine, g, game.Player1, candidates, agents, log); got != second {
			t.Fatalf("returned attacker = %v, want the chosen second face-down creature", got)
		}
		if len(recorder.requests) != 1 || len(log.Choices) != 1 {
			t.Fatalf("requests/logged = %d/%d, want one choice", len(recorder.requests), len(log.Choices))
		}
		for _, request := range []game.ChoiceRequest{recorder.requests[0], log.Choices[0].Request} {
			if got := labelsOf(request.Options); !slices.Equal(got, wantLabels) {
				t.Fatalf("labels = %q, want %q", got, wantLabels)
			}
		}
	})

	t.Run("triggering batch copy", func(t *testing.T) {
		recorder := &faceDownChoiceRecorder{selections: [][]int{{1}}}
		agents := simPassPolicies()
		agents[game.Player1] = recorder
		log := &TurnLog{}
		r := &effectResolver{engine: engine, game: g, obj: &game.StackObject{Controller: game.Player1}, agents: agents, log: log}
		if got := r.chooseTriggeringBatchMember(game.Player1, candidates); got != second {
			t.Fatalf("chosen member = %v, want the chosen second face-down creature", got)
		}
		if len(recorder.requests) != 1 || len(log.Choices) != 1 {
			t.Fatalf("requests/logged = %d/%d, want one choice", len(recorder.requests), len(log.Choices))
		}
		for _, request := range []game.ChoiceRequest{recorder.requests[0], log.Choices[0].Request} {
			if got := labelsOf(request.Options); !slices.Equal(got, wantLabels) {
				t.Fatalf("labels = %q, want %q", got, wantLabels)
			}
		}
	})

	t.Run("resolution permanent choice", func(t *testing.T) {
		player := game.ControllerReference()
		options, values := resolutionChoiceOptions(g, &game.StackObject{Controller: game.Player1}, game.Player1, &game.ResolutionChoice{
			Kind:            game.ResolutionChoicePermanent,
			PlayerReference: &player,
			Selection:       &game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou},
		})
		if got := labelsOf(options); !slices.Equal(got, wantLabels) {
			t.Fatalf("labels = %q, want %q", got, wantLabels)
		}
		if result := values[1]; result.PermanentID != second.ObjectID || result.CardName != "" {
			t.Fatalf("result = %#v, want the second face-down creature with no name", result)
		}
	})
}
