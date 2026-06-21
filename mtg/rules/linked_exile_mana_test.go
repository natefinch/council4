package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

const testImprintLink = "imprint"

// chromeMoxDef builds a reusable imprint artifact: an ETB that may exile a
// nonartifact, nonland card from hand (publishing the imprint link by object
// identity) and a mana ability that adds one mana of the imprinted card's
// colors. No card-name-specific behavior backs it.
func chromeMoxDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Chrome Mox",
		Types: []types.Card{types.Artifact},
		ManaAbilities: []game.ManaAbility{
			game.TapLinkedExileColorManaAbility(testImprintLink),
		},
	}}
}

// addExiledColorCard creates a card instance owned by player with the given
// colors and places it in exile, returning its instance ID.
func addExiledColorCard(g *game.Game, owner game.PlayerID, name string, colors ...color.Color) game.ObjectID {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID:    cardID,
		Owner: owner,
		Def: &game.CardDef{CardFace: game.CardFace{
			Name:   name,
			Types:  []types.Card{types.Creature},
			Colors: colors,
		}},
	}
	g.Players[owner].Exile.Add(cardID)
	return cardID
}

// triggeredObjFor builds a triggered-ability stack object whose source is the
// given permanent, matching the object the ETB ability resolves under.
func triggeredObjFor(permanent *game.Permanent) *game.StackObject {
	return &game.StackObject{
		Kind:         game.StackTriggeredAbility,
		SourceID:     permanent.ObjectID,
		SourceCardID: permanent.CardInstanceID,
		Controller:   permanent.Controller,
	}
}

func imprintManaChoice() *game.ResolutionChoice {
	return &game.ResolutionChoice{
		Kind:        game.ResolutionChoiceMana,
		ColorSource: game.ResolutionChoiceColorSourceLinkedExileColors,
		LinkID:      testImprintLink,
	}
}

// TestLinkedExileColorsManaReadsImprintedColors verifies the mana ability offers
// exactly the imprinted card's colors in WUBRG order, including every color of a
// multicolored imprint.
func TestLinkedExileColorsManaReadsImprintedColors(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	mox := addCombatPermanent(g, game.Player1, chromeMoxDef())
	obj := triggeredObjFor(mox)

	cardID := addExiledColorCard(g, game.Player1, "Multi", color.Blue, color.White)
	rememberLinkedObject(g, linkedObjectByObjectKey(g, obj, testImprintLink), game.LinkedObjectRef{CardID: cardID})

	got := linkedExileColorsMana(g, obj, imprintManaChoice())
	if want := []mana.Color{mana.W, mana.U}; !slices.Equal(got, want) {
		t.Fatalf("imprinted colors = %v, want %v (WUBRG order)", got, want)
	}
}

// TestLinkedExileColorsManaColorlessImprintEmpty verifies a colorless imprint
// yields no colors, leaving the mana ability unusable.
func TestLinkedExileColorsManaColorlessImprintEmpty(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	mox := addCombatPermanent(g, game.Player1, chromeMoxDef())
	obj := triggeredObjFor(mox)

	cardID := addExiledColorCard(g, game.Player1, "Colorless")
	rememberLinkedObject(g, linkedObjectByObjectKey(g, obj, testImprintLink), game.LinkedObjectRef{CardID: cardID})

	if got := linkedExileColorsMana(g, obj, imprintManaChoice()); len(got) != 0 {
		t.Fatalf("colorless imprint colors = %v, want empty", got)
	}
}

// TestLinkedExileColorsManaNoImprintEmpty verifies that with no imprint recorded
// (declined or never exiled) the mana ability offers no colors.
func TestLinkedExileColorsManaNoImprintEmpty(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	mox := addCombatPermanent(g, game.Player1, chromeMoxDef())
	obj := triggeredObjFor(mox)

	if got := linkedExileColorsMana(g, obj, imprintManaChoice()); len(got) != 0 {
		t.Fatalf("no-imprint colors = %v, want empty", got)
	}
}

// TestLinkedExileColorsManaReentryDropsImprint verifies that the link follows the
// permanent's object identity: a fresh entry (new object ID) sees no prior
// imprint even though the same card instance returns to the battlefield.
func TestLinkedExileColorsManaReentryDropsImprint(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	mox := addCombatPermanent(g, game.Player1, chromeMoxDef())
	obj := triggeredObjFor(mox)

	cardID := addExiledColorCard(g, game.Player1, "Multi", color.Red, color.Green)
	rememberLinkedObject(g, linkedObjectByObjectKey(g, obj, testImprintLink), game.LinkedObjectRef{CardID: cardID})

	// Same card instance re-enters as a new object: a fresh object ID, no link.
	reentered := &game.Permanent{
		ObjectID:       g.IDGen.Next(),
		CardInstanceID: mox.CardInstanceID,
		Owner:          game.Player1,
		Controller:     game.Player1,
	}
	if got := linkedExileColorsMana(g, triggeredObjFor(reentered), imprintManaChoice()); len(got) != 0 {
		t.Fatalf("re-entered object colors = %v, want empty (object-scoped link)", got)
	}
}

// TestLinkedExileColorManaAbilityActivationGating verifies activation legality
// tracks the imprint: unusable with no imprint or a colorless imprint, usable
// once a colored card is imprinted.
func TestLinkedExileColorManaAbilityActivationGating(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setSorcerySpeedTurn(g, game.Player1)
	mox := addCombatPermanent(g, game.Player1, chromeMoxDef())
	card, ok := permanentCardDef(g, mox)
	if !ok {
		t.Fatal("permanent card definition not found")
	}
	obj := triggeredObjFor(mox)
	key := linkedObjectByObjectKey(g, obj, testImprintLink)

	if canActivateManaAbility(g, game.Player1, mox, &card.ManaAbilities[0], 0) {
		t.Fatal("canActivateManaAbility() = true, want false with no imprint")
	}

	colorless := addExiledColorCard(g, game.Player1, "Colorless")
	rememberLinkedObject(g, key, game.LinkedObjectRef{CardID: colorless})
	if canActivateManaAbility(g, game.Player1, mox, &card.ManaAbilities[0], 0) {
		t.Fatal("canActivateManaAbility() = true, want false with colorless imprint")
	}

	clearLinkedObjects(g, key)
	colored := addExiledColorCard(g, game.Player1, "Multi", color.Black, color.Green)
	rememberLinkedObject(g, key, game.LinkedObjectRef{CardID: colored})
	if !canActivateManaAbility(g, game.Player1, mox, &card.ManaAbilities[0], 0) {
		t.Fatal("canActivateManaAbility() = false, want true with a colored imprint")
	}
}

// imprintExileAgent answers the optional "may" prompt per accept, then selects
// the hand card whose label matches wanted for the exile choice.
type imprintExileAgent struct {
	accept bool
	wanted string
}

func (imprintExileAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a imprintExileAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceMay {
		if a.accept {
			return []int{1}
		}
		return []int{0}
	}
	for _, option := range request.Options {
		if option.Label == a.wanted {
			return []int{option.Index}
		}
	}
	return []int{}
}

func exileFromHandInstruction() *game.Instruction {
	return &game.Instruction{
		Optional: true,
		Primitive: game.ExileFromHand{
			Player:        game.ControllerReference(),
			Selection:     game.Selection{ExcludedTypes: []types.Card{types.Artifact, types.Land}},
			Amount:        game.Fixed(1),
			PublishLinked: testImprintLink,
		},
	}
}

// TestExileFromHandImprintsChosenColoredCard verifies the ETB optional exile
// moves the chosen matching card to exile, links it to the source object, and
// makes exactly its colors available to the imprint mana ability.
func TestExileFromHandImprintsChosenColoredCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	mox := addCombatPermanent(g, game.Player1, chromeMoxDef())
	obj := triggeredObjFor(mox)

	multi := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:   "Multi",
		Types:  []types.Card{types.Creature},
		Colors: []color.Color{color.Blue, color.Red},
	}})

	agents := [game.NumPlayers]PlayerAgent{game.Player1: imprintExileAgent{accept: true, wanted: "Multi"}}
	engine.resolveInstructionWithChoices(g, obj, exileFromHandInstruction(), agents, &TurnLog{})

	if g.Players[game.Player1].Hand.Contains(multi) {
		t.Fatal("imprinted card still in hand")
	}
	if !g.Players[game.Player1].Exile.Contains(multi) {
		t.Fatal("imprinted card not moved to exile")
	}
	if got := linkedExileColorsMana(g, obj, imprintManaChoice()); !slices.Equal(got, []mana.Color{mana.U, mana.R}) {
		t.Fatalf("imprinted colors = %v, want [U R]", got)
	}
}

// TestExileFromHandDeclineLeavesNoImprint verifies declining the optional exile
// keeps the hand intact and records no imprint, leaving the ability unusable.
func TestExileFromHandDeclineLeavesNoImprint(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	mox := addCombatPermanent(g, game.Player1, chromeMoxDef())
	obj := triggeredObjFor(mox)

	multi := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:   "Multi",
		Types:  []types.Card{types.Creature},
		Colors: []color.Color{color.Blue, color.Red},
	}})

	agents := [game.NumPlayers]PlayerAgent{game.Player1: imprintExileAgent{accept: false}}
	engine.resolveInstructionWithChoices(g, obj, exileFromHandInstruction(), agents, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(multi) {
		t.Fatal("declined exile removed card from hand")
	}
	if got := linkedExileColorsMana(g, obj, imprintManaChoice()); len(got) != 0 {
		t.Fatalf("declined exile colors = %v, want empty", got)
	}
}

// TestLinkedExileColorManaAbilityActivationAddsImprintedColor exercises the full
// activation path: with a multicolored card imprinted on the source object, the
// mana ability offers exactly the imprint's colors and produces the chosen one.
func TestLinkedExileColorManaAbilityActivationAddsImprintedColor(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	setSorcerySpeedTurn(g, game.Player1)
	mox := addCombatPermanent(g, game.Player1, chromeMoxDef())
	obj := triggeredObjFor(mox)

	cardID := addExiledColorCard(g, game.Player1, "Multi", color.White, color.Black)
	rememberLinkedObject(g, linkedObjectByObjectKey(g, obj, testImprintLink), game.LinkedObjectRef{CardID: cardID})

	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}},
	}
	log := TurnLog{}
	if !engine.applyActionWithChoices(g, game.Player1, action.ActivateAbility(mox.ObjectID, 0, nil, 0), agents, &log) {
		t.Fatal("applyActionWithChoices(Chrome Mox mana) = false, want true")
	}
	if !mox.Tapped {
		t.Fatal("Chrome Mox was not tapped to produce mana")
	}
	if len(log.Choices) != 1 {
		t.Fatalf("choices = %+v, want one mana choice", log.Choices)
	}
	options := log.Choices[0].Request.Options
	if len(options) != 2 || options[0].Label != "W" || options[1].Label != "B" {
		t.Fatalf("choice options = %+v, want [W B] (imprint colors, WUBRG order)", options)
	}
	// The agent selected option index 1 ("B"), so one black mana is added.
	if got := g.Players[game.Player1].ManaPool.Amount(mana.B); got != 1 {
		t.Fatalf("black mana = %d, want 1", got)
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.W); got != 0 {
		t.Fatalf("white mana = %d, want 0 (chose black)", got)
	}
}

// TestExileFromHandSkipsExcludedTypes verifies the selection filter: with only
// artifact and land cards in hand, accepting the optional exile imprints nothing.
func TestExileFromHandSkipsExcludedTypes(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	mox := addCombatPermanent(g, game.Player1, chromeMoxDef())
	obj := triggeredObjFor(mox)

	artifact := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:   "Relic",
		Types:  []types.Card{types.Artifact},
		Colors: []color.Color{color.Blue},
	}})
	land := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Island",
		Types: []types.Card{types.Land},
	}})

	agents := [game.NumPlayers]PlayerAgent{game.Player1: imprintExileAgent{accept: true, wanted: "Relic"}}
	engine.resolveInstructionWithChoices(g, obj, exileFromHandInstruction(), agents, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(artifact) || !g.Players[game.Player1].Hand.Contains(land) {
		t.Fatal("excluded-type card was exiled")
	}
	if got := linkedExileColorsMana(g, obj, imprintManaChoice()); len(got) != 0 {
		t.Fatalf("no eligible card colors = %v, want empty", got)
	}
}
