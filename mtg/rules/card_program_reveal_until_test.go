package rules

import (
	"errors"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: Amplifire (oracle c6861c50-d336-4352-a76d-ec22a86ec09e,
// printing 7a58c77f-7c15-44f4-8011-6046482a2d08) from the pinned corpus
// 6da9ced0...: {2}{R}{R} Creature — Elemental 1/1, "At the beginning of your
// upkeep, reveal cards from the top of your library until you reveal a creature
// card. Until your next turn, this creature's base power becomes twice that
// card's power and its base toughness becomes twice that card's toughness. Put
// the revealed cards on the bottom of your library in a random order."
// https://api.scryfall.com/cards/7a58c77f-7c15-44f4-8011-6046482a2d08/rulings
// returned three WotC rulings (2019-01-25): it is 1/1 from your untap step until
// the trigger resolves; effects that modify its power or toughness (and
// counters) apply after the base set; if you reveal your entire library without
// a creature card, its base power and toughness become 0/0 and the library is
// put back in a random order. Issue #3894's "no P/T effect" wording for that
// case is superseded by the ruling. CR 2026-11-06 verified copy: 701.20a-b
// (reveal shows a card to all players without moving it), 701.20d (reordered
// revealed library cards become new objects), 611.2c (affected set fixed when
// the effect begins), 613.4b (layer 7b base set). The card stays unregistered;
// these test-local programs prove the shared reveal-until operation.
var (
	amplifireTestReference         = game.ProgramReference{ID: "test/amplifire", Version: "1"}
	revealUntilRevokeTestReference = game.ProgramReference{ID: "test/reveal-until-revoke", Version: "1"}
)

func amplifireTestProgram() cardprogram.Program {
	return cardprogram.Program{Reference: amplifireTestReference, Resolve: func(ctx cardprogram.Context) error {
		revealed, err := ctx.RevealUntil(game.RevealUntil{
			Player:      game.ControllerReference(),
			Until:       game.Selection{RequiredTypes: []types.Card{types.Creature}},
			Destination: zone.Library,
		})
		if err != nil || !revealed.Available {
			return err
		}
		// Ruling 2019-01-25: revealing the whole library without a creature
		// card makes the base power and toughness 0/0.
		power, toughness := 0, 0
		if revealed.Hit.Exists {
			hit := revealed.Hit.Val
			if !hit.Power.Exists || !hit.Toughness.Exists {
				return errors.New("amplifire: revealed creature's power or toughness is not evaluable outside the battlefield (#3836)")
			}
			power, toughness = 2*hit.Power.Val, 2*hit.Toughness.Val
		}
		_, err = ctx.Execute(game.ApplyContinuous{
			Object:   opt.Val(game.SourcePermanentReference()),
			Duration: game.DurationUntilYourNextTurn,
			ContinuousEffects: []game.ContinuousEffect{{
				Layer:        game.LayerPowerToughnessSet,
				SetPower:     opt.Val(game.PT{Value: power}),
				SetToughness: opt.Val(game.PT{Value: toughness}),
			}},
		})
		return err
	}}
}

func upkeepProgramTestCard(name string, reference game.ProgramReference) *game.CardDef {
	one := game.PT{Value: 1}
	return &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		ManaCost:  opt.Val(cost.Mana{cost.O(2), cost.R, cost.R}),
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Elemental},
		Power:     opt.Val(one),
		Toughness: opt.Val(one),
		TriggeredAbilities: []game.TriggeredAbility{{
			Text: "At the beginning of your upkeep, reveal cards from the top of your library until you reveal a creature card. Until your next turn, this creature's base power becomes twice that card's power and its base toughness becomes twice that card's toughness. Put the revealed cards on the bottom of your library in a random order.",
			Trigger: game.TriggerCondition{
				Type: game.TriggerAt,
				Pattern: game.TriggerPattern{
					Event:      game.EventBeginningOfStep,
					Controller: game.TriggerControllerYou,
					Step:       game.StepUpkeep,
				},
			},
			Content: game.Mode{Program: reference}.Ability(),
		}},
	}}
}

func revealUntilTestInstant(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		ManaCost: opt.Val(cost.Mana{cost.O(1)}),
		Types:    []types.Card{types.Instant},
	}}
}

type amplifireObservation struct {
	turn             int
	step             game.Step
	triggerOnStack   bool
	power, toughness int
	library          []id.ID
	reveals          []game.Event
}

// amplifireScenario is shared by every seat's agent; agents only pass and
// record each priority and choice request they receive.
type amplifireScenario struct {
	g            *game.Game
	amplifire    id.ID
	observations []amplifireObservation
	choices      []game.ChoiceRequest
}

func (s *amplifireScenario) agents() [game.NumPlayers]PlayerAgent {
	var agents [game.NumPlayers]PlayerAgent
	for seat := range agents {
		agents[seat] = amplifireAgent{scenario: s}
	}
	return agents
}

// upkeep returns Player1's upkeep observations on turn: the first is taken with
// the trigger on the stack, the last after it resolved.
func (s *amplifireScenario) upkeep(t *testing.T, turn int) (first, last amplifireObservation) {
	t.Helper()
	var matching []amplifireObservation
	for _, observation := range s.observations {
		if observation.turn == turn && observation.step == game.StepUpkeep {
			matching = append(matching, observation)
		}
	}
	if len(matching) < 2 || !matching[0].triggerOnStack || matching[len(matching)-1].triggerOnStack {
		t.Fatalf("turn %d upkeep observations = %+v, want the trigger on the stack and then resolved", turn, matching)
	}
	return matching[0], matching[len(matching)-1]
}

type amplifireAgent struct{ scenario *amplifireScenario }

func (a amplifireAgent) ChooseAction(_ PlayerObservation, _ []action.Action) action.Action {
	s := a.scenario
	permanent, ok := permanentByObjectID(s.g, s.amplifire)
	if !ok {
		return action.Pass()
	}
	top, onStack := s.g.Stack.Peek()
	toughness, _ := effectiveToughness(s.g, permanent)
	var reveals []game.Event
	for _, event := range s.g.Events {
		if event.Kind == game.EventCardRevealed {
			reveals = append(reveals, event)
		}
	}
	s.observations = append(s.observations, amplifireObservation{
		turn:           s.g.Turn.TurnNumber,
		step:           s.g.Turn.Step,
		triggerOnStack: onStack && top.Kind == game.StackTriggeredAbility && top.SourceID == s.amplifire,
		power:          effectivePower(s.g, permanent),
		toughness:      toughness,
		library:        slices.Clone(s.g.Players[game.Player1].Library.All()),
		reveals:        reveals,
	})
	return action.Pass()
}

func (a amplifireAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	a.scenario.choices = append(a.scenario.choices, request)
	return request.DefaultSelection
}

func newAmplifireEngine(t *testing.T, programs ...cardprogram.Program) *Engine {
	t.Helper()
	catalog, err := cardprogram.NewCatalog(programs)
	if err != nil {
		t.Fatal(err)
	}
	return NewEngineWithPrograms(nil, catalog)
}

func wantAmplifirePT(t *testing.T, label string, observation amplifireObservation, power, toughness int) {
	t.Helper()
	if observation.power != power || observation.toughness != toughness {
		t.Fatalf("%s: Amplifire is %d/%d, want %d/%d", label, observation.power, observation.toughness, power, toughness)
	}
}

func revealedCardIDs(events []game.Event) []id.ID {
	cards := make([]id.ID, 0, len(events))
	for _, event := range events {
		cards = append(cards, event.CardID)
	}
	return cards
}

// TestAmplifireRevealsUntilCreatureAndDoublesItsBase plays Amplifire's upkeep
// trigger over a land, an instant and a 3/4 creature card. Exactly those three
// cards are publicly revealed in order with no player choice, they become the
// bottom three cards of an unchanged-size library in a seed-determined order,
// and Amplifire is 6/8 through every opponent's turn. At Player1's next turn it
// is 1/1 again before the new trigger resolves, which then reveals a 1/2.
func TestAmplifireRevealsUntilCreatureAndDoublesItsBase(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := newAmplifireEngine(t, amplifireTestProgram())
	amplifire := addCombatPermanent(g, game.Player1, upkeepProgramTestCard("Amplifire", amplifireTestReference))
	for _, player := range []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4} {
		stockLibrary(g, player, 12)
	}
	creature := addCardToLibrary(g, game.Player1, vanillaCreature("Revealed Brute", 3, 4))
	instant := addCardToLibrary(g, game.Player1, revealUntilTestInstant("Revealed Instant"))
	land := addCardToLibrary(g, game.Player1, basicLand())
	revealed := []id.ID{land, instant, creature}
	untouched := slices.Clone(g.Players[game.Player1].Library.All()[3:])
	librarySize := g.Players[game.Player1].Library.Size()
	clones := [2]*game.Game{g.Clone(), g.Clone()}

	s := &amplifireScenario{g: g, amplifire: amplifire.ObjectID}
	agents := s.agents()
	engine.runTurn(g, agents)
	first, last := s.upkeep(t, 1)
	wantAmplifirePT(t, "before the trigger resolves", first, 1, 1)
	wantAmplifirePT(t, "after the trigger resolves", last, 6, 8)
	if got := revealedCardIDs(last.reveals); !slices.Equal(got, revealed) {
		t.Fatalf("revealed cards = %v, want exactly %v in reveal order", got, revealed)
	}
	for _, event := range last.reveals {
		if event.Player != game.Player1 || event.FromZone != zone.Library || event.SourceID != amplifire.CardInstanceID {
			t.Fatalf("reveal event = %+v, want Player1's library card revealed by Amplifire", event)
		}
	}
	if len(s.choices) != 0 {
		t.Fatalf("players received %d choice requests, want none: the bottom order is random", len(s.choices))
	}
	if len(last.library) != librarySize {
		t.Fatalf("library size after the trigger = %d, want unchanged %d", len(last.library), librarySize)
	}
	bottom := last.library[len(last.library)-3:]
	if !slices.Equal(last.library[:len(last.library)-3], untouched) {
		t.Fatal("the unrevealed cards did not keep their order above the revealed cards")
	}
	sortedBottom, sortedRevealed := sortedIDs(bottom), sortedIDs(revealed)
	if !slices.Equal(sortedBottom, sortedRevealed) {
		t.Fatalf("library bottom = %v, want the revealed cards %v", bottom, revealed)
	}

	// The random bottom order comes from the game's seeded RNG: two clones of
	// the same state (each seeded identically, see Game.Clone) reproduce the
	// same reveals, bottom order and power/toughness without any decision.
	var cloneResults [2]amplifireObservation
	for i, clone := range clones {
		cloned := &amplifireScenario{g: clone, amplifire: amplifire.ObjectID}
		engine.runTurn(clone, cloned.agents())
		_, cloneResults[i] = cloned.upkeep(t, 1)
		wantAmplifirePT(t, "clone after the trigger resolves", cloneResults[i], 6, 8)
		if !slices.Equal(revealedCardIDs(cloneResults[i].reveals), revealed) {
			t.Fatalf("clone reveals = %v, want %v", revealedCardIDs(cloneResults[i].reveals), revealed)
		}
	}
	if !slices.Equal(cloneResults[0].library, cloneResults[1].library) {
		t.Fatalf("clone libraries differ: %v and %v", cloneResults[0].library, cloneResults[1].library)
	}

	for turn := 2; turn <= 4; turn++ {
		engine.runTurn(g, agents)
		permanent, _ := permanentByObjectID(g, amplifire.ObjectID)
		toughness, _ := effectiveToughness(g, permanent)
		if power := effectivePower(g, permanent); power != 6 || toughness != 8 {
			t.Fatalf("after turn %d Amplifire is %d/%d, want 6/8 until Player1's next turn", turn, power, toughness)
		}
	}
	if g.Turn.TurnNumber != 5 || g.Turn.ActivePlayer != game.Player1 {
		t.Fatalf("turn = %d for %v, want Player1's turn 5 next", g.Turn.TurnNumber, g.Turn.ActivePlayer)
	}

	addCardToLibrary(g, game.Player1, vanillaCreature("Small Sprite", 1, 2))
	engine.runTurn(g, agents)
	first, last = s.upkeep(t, 5)
	wantAmplifirePT(t, "Player1's next upkeep before the new trigger", first, 1, 1)
	wantAmplifirePT(t, "Player1's next upkeep after revealing a 1/2", last, 2, 4)
}

// TestAmplifireWithoutCreatureRevealsWholeLibrary follows the 2019-01-25
// ruling: with no creature card every library card is revealed and put back in
// a random order, and Amplifire's base becomes 0/0, leaving its two +1/+1
// counters (layer 7c) to keep it alive as a 2/2.
func TestAmplifireWithoutCreatureRevealsWholeLibrary(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := newAmplifireEngine(t, amplifireTestProgram())
	amplifire := addCombatPermanent(g, game.Player1, upkeepProgramTestCard("Amplifire", amplifireTestReference))
	addCountersToPermanent(g, amplifire, counter.PlusOnePlusOne, 2)
	for _, player := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		stockLibrary(g, player, 4)
	}
	stockLibrary(g, game.Player1, 4)
	addCardToLibrary(g, game.Player1, revealUntilTestInstant("Revealed Instant"))
	before := slices.Clone(g.Players[game.Player1].Library.All())

	s := &amplifireScenario{g: g, amplifire: amplifire.ObjectID}
	engine.runTurn(g, s.agents())
	first, last := s.upkeep(t, 1)
	wantAmplifirePT(t, "before the trigger resolves", first, 3, 3)
	wantAmplifirePT(t, "after revealing no creature", last, 2, 2)
	if got := revealedCardIDs(last.reveals); !slices.Equal(got, before) {
		t.Fatalf("revealed cards = %v, want the whole library %v in order", got, before)
	}
	if !slices.Equal(sortedIDs(last.library), sortedIDs(before)) {
		t.Fatalf("library after the trigger = %v, want the same cards as %v", last.library, before)
	}
	if len(s.choices) != 0 {
		t.Fatalf("players received %d choice requests, want none", len(s.choices))
	}
}

// TestRevealUntilRevokesReorderedCardHandles proves CR 701.20d for the shared
// operation: a card revealed earlier in the invocation and then reordered to
// the library bottom is a new object, so its old handle cannot move it.
func TestRevealUntilRevokesReorderedCardHandles(t *testing.T) {
	var moveErr error
	var result cardprogram.UntilResult
	program := cardprogram.Program{Reference: revealUntilRevokeTestReference, Resolve: func(ctx cardprogram.Context) error {
		peeked, err := ctx.Execute(game.Reveal{Player: game.ControllerReference(), Amount: game.Fixed(1)})
		if err != nil || len(peeked.Cards) != 1 {
			return errors.New("expected one revealed top-card handle")
		}
		result, err = ctx.RevealUntil(game.RevealUntil{
			Player:      game.ControllerReference(),
			Until:       game.Selection{RequiredTypes: []types.Card{types.Creature}},
			Destination: zone.Library,
		})
		if err != nil {
			return err
		}
		_, moveErr = ctx.MoveCards(cardprogram.CardPlacement{Cards: peeked.Cards, Destination: zone.Hand})
		return nil
	}}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := newAmplifireEngine(t, program)
	source := addCombatPermanent(g, game.Player1, upkeepProgramTestCard("Reveal Probe", revealUntilRevokeTestReference))
	for _, player := range []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4} {
		stockLibrary(g, player, 6)
	}
	creature := addCardToLibrary(g, game.Player1, vanillaCreature("Revealed Brute", 3, 4))
	top := addCardToLibrary(g, game.Player1, basicLand())
	handSize := g.Players[game.Player1].Hand.Size()

	s := &amplifireScenario{g: g, amplifire: source.ObjectID}
	engine.runTurn(g, s.agents())
	s.upkeep(t, 1)
	if !result.Available || !result.Hit.Exists || result.Hit.Val.Identity.CardID != creature ||
		len(result.Others) != 1 || result.Others[0].Identity.CardID != top {
		t.Fatalf("reveal-until result = %+v, want hit %v after other %v", result, creature, top)
	}
	if hit := result.Hit.Val; hit.Power != opt.Val(3) || hit.Toughness != opt.Val(4) || hit.Info.Name != "Revealed Brute" {
		t.Fatalf("hit handle = %+v, want Revealed Brute's 3/4 card characteristics", hit)
	}
	if moveErr == nil {
		t.Fatal("moving a reordered card through its earlier handle succeeded, want CR 701.20d new object")
	}
	if g.Players[game.Player1].Hand.Contains(top) || g.Players[game.Player1].Hand.Size() != handSize+1 {
		t.Fatal("the reordered top card moved to hand (only the draw-step card should arrive)")
	}
}

// TestDeclarativeRevealUntilLibraryDestinationBottomsRandomly proves the
// declarative RevealUntil route shares the same public reveals and random
// bottom placement.
func TestDeclarativeRevealUntilLibraryDestinationBottomsRandomly(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	stockLibrary(g, game.Player1, 5)
	creature := addCardToLibrary(g, game.Player1, vanillaCreature("Revealed Brute", 3, 4))
	land := addCardToLibrary(g, game.Player1, basicLand())
	obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1}
	resolver := newEffectResolver(NewEngine(nil), g, obj, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	resolver.resolveInstruction(&game.Instruction{Primitive: game.RevealUntil{
		Player:      game.ControllerReference(),
		Until:       game.Selection{RequiredTypes: []types.Card{types.Creature}},
		Destination: zone.Library,
	}})
	var reveals []id.ID
	for _, event := range g.Events {
		if event.Kind == game.EventCardRevealed {
			reveals = append(reveals, event.CardID)
		}
	}
	if !slices.Equal(reveals, []id.ID{land, creature}) {
		t.Fatalf("revealed = %v, want land then creature", reveals)
	}
	library := g.Players[game.Player1].Library.All()
	if len(library) != 7 || !slices.Equal(sortedIDs(library[5:]), sortedIDs([]id.ID{land, creature})) {
		t.Fatalf("library = %v, want the two revealed cards on the bottom", library)
	}
}
