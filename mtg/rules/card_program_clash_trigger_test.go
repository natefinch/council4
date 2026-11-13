package rules

import (
	"errors"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: CR effective 2025-11-14 (MagicCompRules 20251114), 701.30a-d:
// "clash with an opponent" chooses an opponent and both players clash
// (701.30b); reveals, APNAP placement decisions and the moves all happen
// within the clash (701.30c); the winner revealed a card with a higher mana
// value than all other revealed cards (701.30d). CR 603.2 and 603.3: an ability
// triggers when its event occurs and is put on the stack the next time a player
// would receive priority, so a trigger on an event during a spell's resolution
// waits until that resolution finishes. Pinned Oracle corpus 6da9ced0...:
// Rebellion of the Flamekin 0e2a283b-c06a-4b54-a627-5b4959318b1a ({3}{R}
// Kindred Enchantment — Elemental; "Whenever you clash, you may pay {1}. If you
// do, create a 3/1 red Elemental Shaman creature token. If you won, that token
// gains haste until end of turn. (This ability triggers after the clash
// ends.)"; Scryfall ruling 2013-09-20: it triggers even when an opponent's spell
// or ability made you clash, and you can win a clash you didn't start) and
// Sylvan Echoes 6d19e2bb-2e5f-4e55-bb86-de01ea35263d ({G} Enchantment;
// "Whenever you clash and win, you may draw a card. (This ability triggers
// after the clash ends.)"; ruling 2013-09-20: it triggers when you win a clash
// started by an opponent's spell or ability). Research the Deep and Oaken
// Brawler (see card_program_clash_test.go) supply the clashes. These test-local
// programs prove the shared clash trigger event; a later card unit admits the
// cards themselves.

var (
	rebellionOfTheFlamekinTestReference = game.ProgramReference{ID: "test/rebellion-of-the-flamekin", Version: "1"}
	sylvanEchoesTestReference           = game.ProgramReference{ID: "test/sylvan-echoes", Version: "1"}
)

const rebellionTokenLink game.LinkedKey = "test/rebellion-of-the-flamekin/token"

type clashTriggerFixture struct {
	*clashFixture
	// facts records the captured clash fact each resolving trigger saw.
	facts *[]cardprogram.ClashEvent
}

func newClashTriggerFixture(t *testing.T) *clashTriggerFixture {
	t.Helper()
	base := newClashFixture(t)
	facts := &[]cardprogram.ClashEvent{}
	capture := func(ctx cardprogram.Context) (cardprogram.ClashEvent, error) {
		fact := ctx.View().Source().EventClash
		if !fact.Exists {
			return cardprogram.ClashEvent{}, errors.New("clash trigger resolved without a captured clash fact")
		}
		*facts = append(*facts, fact.Val)
		return fact.Val, nil
	}
	programs := []cardprogram.Program{
		{
			Reference: rebellionOfTheFlamekinTestReference,
			Resolve: func(ctx cardprogram.Context) error {
				fact, err := capture(ctx)
				if err != nil {
					return err
				}
				paid, err := ctx.Execute(game.Pay{Payment: game.ResolutionPayment{
					Prompt: "Pay {1} to create a 3/1 Elemental Shaman?", ManaCost: opt.Val(cost.Mana{cost.O(1)}),
				}})
				if err != nil || !paid.Succeeded {
					return err
				}
				created, err := ctx.Execute(game.CreateToken{
					Amount: game.Fixed(1), Source: game.TokenDef(elementalShamanTestToken()), PublishLinked: rebellionTokenLink,
				})
				if err != nil || !created.Succeeded || !fact.Won {
					return err
				}
				_, err = ctx.Execute(game.ApplyContinuous{
					Object: opt.Val(game.LinkedObjectReference(string(rebellionTokenLink))),
					ContinuousEffects: []game.ContinuousEffect{{
						Duration: game.DurationPermanent, Layer: game.LayerAbility, AddKeywords: []game.Keyword{game.Haste},
					}},
					Duration: game.DurationUntilEndOfTurn,
				})
				return err
			},
		},
		{
			Reference: sylvanEchoesTestReference,
			Resolve: func(ctx cardprogram.Context) error {
				if _, err := capture(ctx); err != nil {
					return err
				}
				controller := ctx.View().Source().Controller
				selected, err := ctx.Ask(game.ChoiceRequest{
					Kind: game.ChoiceMay, Player: controller, Prompt: "Sylvan Echoes: draw a card?",
					Options:    []game.ChoiceOption{{Index: 0, Label: "No"}, {Index: 1, Label: "Yes"}},
					MinChoices: 1, MaxChoices: 1, DefaultSelection: []int{1},
				})
				if err != nil || !slices.Equal(selected, []int{1}) {
					return err
				}
				_, err = ctx.Execute(game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()})
				return err
			},
		},
	}
	catalog, err := cardprogram.NewCatalog(append(clashTestPrograms(base.results), programs...))
	if err != nil {
		t.Fatal(err)
	}
	base.engine = NewEngineWithPrograms(nil, catalog)
	return &clashTriggerFixture{clashFixture: base, facts: facts}
}

func elementalShamanTestToken() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Elemental Shaman",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Elemental, types.Shaman},
		Colors:    []color.Color{color.Red},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}}
}

func rebellionOfTheFlamekinTestDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Rebellion of the Flamekin",
		ManaCost: opt.Val(cost.Mana{cost.O(3), cost.R}),
		Types:    []types.Card{types.Kindred, types.Enchantment},
		Subtypes: []types.Sub{types.Elemental},
		TriggeredAbilities: []game.TriggeredAbility{{
			Text: "Whenever you clash, you may pay {1}. If you do, create a 3/1 red Elemental Shaman creature token. If you won, that token gains haste until end of turn.",
			Trigger: game.TriggerCondition{Type: game.TriggerWhenever, Pattern: game.TriggerPattern{
				Event: game.EventClashed, Player: game.TriggerPlayerYou,
			}},
			Content: game.Mode{Program: rebellionOfTheFlamekinTestReference}.Ability(),
		}},
	}}
}

func sylvanEchoesTestDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Sylvan Echoes",
		ManaCost: opt.Val(cost.Mana{cost.G}),
		Types:    []types.Card{types.Enchantment},
		TriggeredAbilities: []game.TriggeredAbility{{
			Text: "Whenever you clash and win, you may draw a card.",
			Trigger: game.TriggerCondition{Type: game.TriggerWhenever, Pattern: game.TriggerPattern{
				Event: game.EventClashed, Player: game.TriggerPlayerYou, RequireClashWon: true,
			}},
			Content: game.Mode{Program: sylvanEchoesTestReference}.Ability(),
		}},
	}}
}

// declineMay makes player answer "No" to every may prompt (including optional
// payments) while keeping the clash answers of agents.
func declineMay(agents [game.NumPlayers]PlayerAgent, player game.PlayerID) [game.NumPlayers]PlayerAgent {
	inner := agents[player].(authoredChoicePolicy)
	agents[player] = authoredChoicePolicy{
		simPassPolicy: inner.simPassPolicy,
		choose: func(observation PlayerObservation, request game.ChoiceRequest) []int {
			if request.Kind == game.ChoiceMay {
				return []int{0}
			}
			return inner.choose(observation, request)
		},
	}
	return agents
}

func clashEvents(g *game.Game) []game.Event {
	var events []game.Event
	for _, event := range g.Events {
		if event.Kind == game.EventClashed {
			events = append(events, event)
		}
	}
	return events
}

func tokensNamed(g *game.Game, controller game.PlayerID, name string) []*game.Permanent {
	var tokens []*game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanent.Controller == controller && permanent.TokenDef != nil && permanent.TokenDef.Name == name {
			tokens = append(tokens, permanent)
		}
	}
	return tokens
}

// stackTriggers puts pending triggers on the stack and returns the stack.
func (f *clashTriggerFixture) stackTriggers(agents [game.NumPlayers]PlayerAgent) []*game.StackObject {
	f.t.Helper()
	f.engine.putTriggeredAbilitiesOnStackWithChoices(f.g, agents, &TurnLog{})
	return f.g.Stack.Objects()
}

func TestRebellionOfTheFlamekinTriggersAfterItsControllerClashes(t *testing.T) {
	for _, tc := range []struct {
		name         string
		mine, theirs int
		decline      bool
		wantToken    bool
		wantHaste    bool
		wantWon      bool
	}{
		{name: "win and pay creates a hasty token", mine: 4, theirs: 1, wantToken: true, wantHaste: true, wantWon: true},
		{name: "loss and pay creates a token without haste", mine: 1, theirs: 4, wantToken: true},
		{name: "tie and pay creates a token without haste", mine: 2, theirs: 2, wantToken: true},
		{name: "declining to pay creates nothing", mine: 4, theirs: 1, decline: true, wantWon: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newClashTriggerFixture(t)
			rebellion := addCombatPermanent(f.g, game.Player1, rebellionOfTheFlamekinTestDef())
			addCardToLibrary(f.g, game.Player1, clashLibraryCard("Mine", tc.mine))
			addCardToLibrary(f.g, game.Player1, clashLibraryCard("Drawn", 1))
			addCardToLibrary(f.g, game.Player2, clashLibraryCard("Theirs", tc.theirs))
			spell := f.cast(researchTheDeepTestDef(), mana.U, 1, nil)
			agents := clashAgents(t, game.Player2, nil, nil)
			if tc.decline {
				agents = declineMay(agents, game.Player1)
			}

			f.engine.resolveTopOfStackWithChoices(f.g, agents, &TurnLog{})

			// Research the Deep has fully resolved, and its clash trigger is
			// only now put on the stack (CR 603.3).
			if !f.g.Stack.IsEmpty() {
				t.Fatalf("stack = %d objects, want Research the Deep finished and the trigger not yet stacked", f.g.Stack.Size())
			}
			if inHand := f.g.Players[game.Player1].Hand.Contains(spell); inHand != tc.wantWon {
				t.Fatalf("Research the Deep in hand = %v, want %v", inHand, tc.wantWon)
			}
			events := clashEvents(f.g)
			if len(events) != 2 || events[0].SimultaneousID == 0 || events[0].SimultaneousID != events[1].SimultaneousID {
				t.Fatalf("clash events = %+v, want one per clashing player sharing a SimultaneousID", events)
			}
			for _, event := range events {
				other := game.Player2
				if event.Player == game.Player2 {
					other = game.Player1
				}
				if event.ClashOpponent != other || event.ClashWon != (event.Player == game.Player1 && tc.wantWon || event.Player == game.Player2 && tc.theirs > tc.mine) {
					t.Fatalf("clash event %+v has the wrong opponent or won fact", event)
				}
			}
			stacked := f.stackTriggers(agents)
			if len(stacked) != 1 || stacked[0].SourceID != rebellion.ObjectID || stacked[0].Controller != game.Player1 {
				t.Fatalf("stack = %+v, want exactly Rebellion's trigger", stacked)
			}
			f.g.Players[game.Player1].ManaPool.Add(mana.C, 1)

			f.engine.resolveTopOfStackWithChoices(f.g, agents, &TurnLog{})

			want := cardprogram.ClashEvent{Player: game.Player1, Opponent: game.Player2, Won: tc.wantWon}
			if !slices.Equal(*f.facts, []cardprogram.ClashEvent{want}) {
				t.Fatalf("captured facts = %+v, want %+v", *f.facts, want)
			}
			tokens := tokensNamed(f.g, game.Player1, "Elemental Shaman")
			if (len(tokens) == 1) != tc.wantToken || len(tokens) > 1 {
				t.Fatalf("Elemental Shaman tokens = %d, want token=%v", len(tokens), tc.wantToken)
			}
			if paid := f.g.Players[game.Player1].ManaPool.Total() == 0; paid != tc.wantToken {
				t.Fatalf("{1} paid = %v, want %v", paid, tc.wantToken)
			}
			if !tc.wantToken {
				return
			}
			token := tokens[0]
			if toughness, _ := effectiveToughness(f.g, token); effectivePower(f.g, token) != 3 || toughness != 1 ||
				!slices.Equal(token.TokenDef.Colors, []color.Color{color.Red}) {
				t.Fatal("token is not a 3/1 red Elemental Shaman")
			}
			if got := hasKeyword(f.g, token, game.Haste); got != tc.wantHaste {
				t.Fatalf("token haste = %v, want %v", got, tc.wantHaste)
			}
			if tc.wantHaste {
				expireCleanupDurations(f.g)
				if hasKeyword(f.g, token, game.Haste) {
					t.Fatal("haste lasted past end of turn")
				}
			}
		})
	}
}

func TestRebellionOfTheFlamekinTriggersForTheClashedOpponentOnly(t *testing.T) {
	f := newClashTriggerFixture(t)
	opponents := addCombatPermanent(f.g, game.Player2, rebellionOfTheFlamekinTestDef())
	addCombatPermanent(f.g, game.Player3, rebellionOfTheFlamekinTestDef())
	addCardToLibrary(f.g, game.Player1, clashLibraryCard("Mine", 1))
	addCardToLibrary(f.g, game.Player1, clashLibraryCard("Drawn", 1))
	for _, player := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		addCardToLibrary(f.g, player, clashLibraryCard("Theirs", 5))
	}
	f.cast(researchTheDeepTestDef(), mana.U, 1, nil)
	agents := clashAgents(t, game.Player2, nil, nil)

	f.engine.resolveTopOfStackWithChoices(f.g, agents, &TurnLog{})
	stacked := f.stackTriggers(agents)

	if len(stacked) != 1 || stacked[0].SourceID != opponents.ObjectID || stacked[0].Controller != game.Player2 {
		t.Fatalf("stack = %+v, want only the clashed opponent's Rebellion trigger", stacked)
	}
	f.g.Players[game.Player2].ManaPool.Add(mana.C, 1)
	f.engine.resolveTopOfStackWithChoices(f.g, agents, &TurnLog{})

	want := cardprogram.ClashEvent{Player: game.Player2, Opponent: game.Player1, Won: true}
	if !slices.Equal(*f.facts, []cardprogram.ClashEvent{want}) {
		t.Fatalf("captured facts = %+v, want the opponent's won clash %+v", *f.facts, want)
	}
	tokens := tokensNamed(f.g, game.Player2, "Elemental Shaman")
	if len(tokens) != 1 || !hasKeyword(f.g, tokens[0], game.Haste) {
		t.Fatalf("Player2 tokens = %d, want one hasty Elemental Shaman for winning a clash it didn't start", len(tokens))
	}
	if len(tokensNamed(f.g, game.Player3, "Elemental Shaman")) != 0 {
		t.Fatal("the uninvolved opponent's Rebellion created a token")
	}
}

func TestSylvanEchoesTriggersOnlyWhenItsControllerWinsAClash(t *testing.T) {
	for _, tc := range []struct {
		name         string
		mine, theirs int
		decline      bool
		// wantDrawer is the player whose Sylvan Echoes triggers, if any.
		wantDrawer opt.V[game.PlayerID]
	}{
		{name: "win draws", mine: 5, theirs: 1, wantDrawer: opt.Val(game.Player1)},
		{name: "win may decline the draw", mine: 5, theirs: 1, decline: true, wantDrawer: opt.Val(game.Player1)},
		{name: "loss triggers the winning opponent's echoes", mine: 1, theirs: 5, wantDrawer: opt.Val(game.Player2)},
		{name: "tie triggers nothing", mine: 3, theirs: 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newClashTriggerFixture(t)
			echoes := map[game.PlayerID]*game.Permanent{
				game.Player1: addCombatPermanent(f.g, game.Player1, sylvanEchoesTestDef()),
				game.Player2: addCombatPermanent(f.g, game.Player2, sylvanEchoesTestDef()),
			}
			for _, player := range []game.PlayerID{game.Player1, game.Player2} {
				addCardToLibrary(f.g, player, clashLibraryCard("Spare", 1))
			}
			addCardToLibrary(f.g, game.Player1, clashLibraryCard("Mine", tc.mine))
			addCardToLibrary(f.g, game.Player2, clashLibraryCard("Theirs", tc.theirs))
			f.cast(oakenBrawlerTestDef(), mana.W, 3, nil)
			agents := clashAgents(t, game.Player2, nil, nil)
			if tc.decline {
				agents = declineMay(agents, game.Player1)
			}
			f.engine.resolveTopOfStackWithChoices(f.g, agents, &TurnLog{})
			if stacked := f.stackTriggers(agents); len(stacked) != 1 {
				t.Fatalf("stack = %+v, want only Oaken Brawler's enters trigger", stacked)
			}
			hands := [game.NumPlayers]int{}
			for player := range game.PlayerID(game.NumPlayers) {
				hands[player] = f.g.Players[player].Hand.Size()
			}

			f.engine.resolveTopOfStackWithChoices(f.g, agents, &TurnLog{})
			stacked := f.stackTriggers(agents)

			if !tc.wantDrawer.Exists {
				if len(stacked) != 0 {
					t.Fatalf("stack = %+v, want no Sylvan Echoes trigger after a tie", stacked)
				}
				return
			}
			drawer := tc.wantDrawer.Val
			if len(stacked) != 1 || stacked[0].SourceID != echoes[drawer].ObjectID {
				t.Fatalf("stack = %+v, want only Player%d's Sylvan Echoes trigger", stacked, drawer+1)
			}
			f.engine.resolveTopOfStackWithChoices(f.g, agents, &TurnLog{})
			other := game.Player2
			if drawer == game.Player2 {
				other = game.Player1
			}
			if want := []cardprogram.ClashEvent{{Player: drawer, Opponent: other, Won: true}}; !slices.Equal(*f.facts, want) {
				t.Fatalf("captured facts = %+v, want %+v", *f.facts, want)
			}
			for player := range game.PlayerID(game.NumPlayers) {
				want := hands[player]
				if player == drawer && !tc.decline {
					want++
				}
				if got := f.g.Players[player].Hand.Size(); got != want {
					t.Fatalf("Player%d hand = %d, want %d", player+1, got, want)
				}
			}
		})
	}
}

func TestRebellionOfTheFlamekinTriggerKeepsItsClashFactInAClone(t *testing.T) {
	f := newClashTriggerFixture(t)
	addCombatPermanent(f.g, game.Player1, rebellionOfTheFlamekinTestDef())
	addCardToLibrary(f.g, game.Player1, clashLibraryCard("Mine", 4))
	addCardToLibrary(f.g, game.Player1, clashLibraryCard("Drawn", 1))
	addCardToLibrary(f.g, game.Player2, clashLibraryCard("Theirs", 1))
	f.cast(researchTheDeepTestDef(), mana.U, 1, nil)
	agents := clashAgents(t, game.Player2, nil, nil)
	f.engine.resolveTopOfStackWithChoices(f.g, agents, &TurnLog{})
	if stacked := f.stackTriggers(agents); len(stacked) != 1 {
		t.Fatalf("stack = %+v, want Rebellion's trigger", stacked)
	}
	f.g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	clone := f.g.Clone()
	// Changing the original's recorded event after cloning must not reach the
	// clone's captured trigger event.
	trigger, _ := f.g.Stack.Peek()
	trigger.TriggerEvent.ClashWon = false

	f.engine.resolveTopOfStackWithChoices(clone, agents, &TurnLog{})

	want := cardprogram.ClashEvent{Player: game.Player1, Opponent: game.Player2, Won: true}
	if !slices.Equal(*f.facts, []cardprogram.ClashEvent{want}) {
		t.Fatalf("clone captured facts = %+v, want %+v", *f.facts, want)
	}
	tokens := tokensNamed(clone, game.Player1, "Elemental Shaman")
	if len(tokens) != 1 || !hasKeyword(clone, tokens[0], game.Haste) {
		t.Fatal("the clone's Rebellion trigger did not create a hasty token from its won clash")
	}
	if len(tokensNamed(f.g, game.Player1, "Elemental Shaman")) != 0 {
		t.Fatal("resolving the clone changed the original")
	}
}
