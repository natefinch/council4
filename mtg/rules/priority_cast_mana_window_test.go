package rules

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: Engine gap #3627 (remainder of #3625). Default-registry cards
// checked against the pinned corpus 6da9ced0…a1b4 (oracle_id): Sol Ring
// 6ad8011d-3471-4369-9d68-b264cc027487 {1} "{T}: Add {C}{C}."; Thalia,
// Guardian of Thraben 9b7f1d05-707c-4ed3-9f0e-8ced1232c2ee "First strike /
// Noncreature spells cost {1} more to cast."; Waterlogged Grove
// 70fa2eba-565e-4fed-adc9-7f5d9fcbf1fa "{T}, Pay 1 life: Add {G} or {U}. …";
// Nurturing Peatland 8ed932ff-986c-4592-ad70-53b3fac80d69 "{T}, Pay 1 life: Add
// {B} or {G}. …"; Future Sight 93a0e938-b916-47ec-9803-369e885e253c "Play with
// the top card of your library revealed. / You may play lands and cast spells
// from the top of your library."; Muldrotha, the Gravetide
// e4625704-1d52-44e4-804f-2f45644d76ac "During each of your turns, you may play
// a land and cast a permanent spell of each permanent type from your
// graveyard. …"; Mutagenic Growth c17b9e5b-58b6-4ccf-8283-945ba77429e7 {G/P}
// "Target creature gets +2/+2 until end of turn."; Parcelbeast
// de83b309-ac31-4a57-95e6-ccd3659e6c1f "Mutate {G}{U} …"; Carnivorous
// Cultivator // Enroot db655c5a-06f6-42bc-8180-d9ca1c1d939a "This creature
// enters prepared." // {G} sorcery "Search your library for a land card, put it
// into your graveyard, then shuffle."; Village Rites
// 365548fb-5acc-4a8a-b20b-26d28b7d029f {B} "As an additional cost to cast this
// spell, sacrifice a creature. / Draw two cards."; Millikin
// fa4dffda-6f04-4d0b-829d-28a1a5794dee "{T}, Mill a card: Add {C}. (Activate
// only as an instant. …)", not a mana ability (605.1a; ruling 2026-08-05, see
// library_moving_mana_ability_test.go); Phyrexian Altar
// 8d02b297-97c4-4379-9862-0a462400f66f "Sacrifice a creature: Add one mana of
// any color."; Blightsteel Colossus e80772e2-8623-4094-81a2-70828b2b151c "… If
// Blightsteel Colossus would be put into a graveyard from anywhere, reveal
// Blightsteel Colossus and shuffle it into its owner's library instead.", a
// replacement effect not the Altar's own, so the Altar's ability is still a
// mana ability (605.1a) whose sacrifice of the Colossus moves a card into a
// library; Silent Clearing
// fd45063f-c83c-431c-9104-f139c497ec0d "{T}, Pay 1 life: Add {W} or {B}. …".
// Rulings: Future Sight 2019-06-14 "If the top card of
// your library changes while you're casting a spell, …, the new top card won't
// be revealed until you finish doing so." Comprehensive Rules effective
// 2026-09-25: 601.2a moves the card to the stack first; 601.2f locks the total
// cost; 601.2g "If the total cost includes a mana payment, the player then has
// a chance to activate mana abilities. Mana abilities must be activated before
// costs are paid."; 601.2h pays it; 605.3a a player may activate a mana
// ability "whenever they have priority, or when they are asked to pay a cost";
// 401.5 keeps a changed library top hidden until the spell becomes cast
// (601.2i); 733.1 reverses an illegal cast, the player reversing legal mana
// abilities activated while making it unless mana from them was spent on
// another mana ability that wasn't reversed, but not actions that moved cards
// from a library to any zone other than the stack; 733.2 the player keeps priority.
// Neither land's mana ability is activated automatically, so each cast below is
// offered only when the caster's payment-time mana abilities could pay its
// locked total: one land can't make two mana, both can.

// priorityCastWindowAgent answers a priority cast's choices by default,
// activating each payment-time mana ability whose label contains "Pay 1 life"
// up to limit times (no limit when negative), and choosing the first offered of
// colors when a mana ability adds one of several.
type priorityCastWindowAgent struct {
	limit     int
	colors    []string
	windows   int
	activated int
}

func (*priorityCastWindowAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *priorityCastWindowAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Prompt == "Activate a mana ability while paying?" {
		a.windows++
		if a.limit >= 0 && a.activated >= a.limit {
			return request.DefaultSelection
		}
		for _, option := range request.Options {
			if strings.Contains(option.Label, "Pay 1 life") {
				a.activated++
				return []int{option.Index}
			}
		}
		return request.DefaultSelection
	}
	for _, want := range a.colors {
		for _, option := range request.Options {
			if option.Label == want {
				return []int{option.Index}
			}
		}
	}
	return request.DefaultSelection
}

// priorityCastRoute builds a game in which casting a card with priority,
// through one route, needs a locked total of two mana, none of it in the pool.
type priorityCastRoute struct {
	name   string
	colors []string
	build  func(t *testing.T) (*game.Game, *Engine, id.ID)
	// cast reports whether act is the route's cast of cardID.
	cast func(act action.Action, cardID id.ID) bool
	// done checks the spell cast with spent mana; undone checks the game after
	// a reversed or refused cast.
	done   func(t *testing.T, g *game.Game, cardID id.ID)
	undone func(t *testing.T, g *game.Game, cardID id.ID)
}

func castOfCardFrom(sourceZone zone.Type) func(action.Action, id.ID) bool {
	return func(act action.Action, cardID id.ID) bool {
		cast, ok := act.CastSpellPayload()
		return ok && cast.CardID == cardID && normalizedCastSourceZone(cast) == sourceZone && !cast.Mutate
	}
}

func spellOnStackFrom(sourceZone zone.Type, isCopy bool) func(*testing.T, *game.Game, id.ID) {
	return func(t *testing.T, g *game.Game, cardID id.ID) {
		t.Helper()
		obj, ok := g.Stack.Peek()
		if !ok || obj.Kind != game.StackSpell || obj.SourceID != cardID || obj.SourceZone != sourceZone || obj.Copy != isCopy || obj.ManaSpentToCast != 2 {
			t.Fatalf("stack top = %+v, want the spell cast from %v with two mana spent", obj, sourceZone)
		}
	}
}

func cardStillIn(sourceZone zone.Type) func(*testing.T, *game.Game, id.ID) {
	return func(t *testing.T, g *game.Game, cardID id.ID) {
		t.Helper()
		p := g.Players[game.Player1]
		var in bool
		switch sourceZone {
		case zone.Hand:
			in = p.Hand.Contains(cardID)
		case zone.Library:
			top, ok := p.Library.Top()
			in = ok && top == cardID
		case zone.Graveyard:
			in = p.Graveyard.Contains(cardID)
		}
		if !in || g.Stack.Size() != 0 {
			t.Fatalf("card in %v = %t, stack size = %d; want the card back where it was and nothing cast", sourceZone, in, g.Stack.Size())
		}
		for _, event := range g.Events {
			if event.Kind == game.EventSpellCast && event.CardID == cardID {
				t.Fatal("an uncompleted cast produced a spell-cast event")
			}
		}
	}
}

func priorityCastRoutes() []priorityCastRoute {
	solRingGame := func(permission string, put func(*testing.T, *game.Game) id.ID) func(*testing.T) (*game.Game, *Engine, id.ID) {
		return func(t *testing.T) (*game.Game, *Engine, id.ID) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			if permission != "" {
				addCombatPermanent(g, game.Player1, registryCardDef(t, permission))
			}
			addCombatPermanent(g, game.Player2, registryCardDef(t, "Thalia, Guardian of Thraben"))
			cardID := put(t, g)
			g.Players[game.Player1].Life = 20
			giveMainPhasePriority(g, game.Player1)
			return g, NewEngine(nil), cardID
		}
	}
	return []priorityCastRoute{
		{
			name: "hand",
			build: solRingGame("", func(t *testing.T, g *game.Game) id.ID {
				return addCardToHand(g, game.Player1, registryCardDef(t, "Sol Ring"))
			}),
			cast: castOfCardFrom(zone.Hand), done: spellOnStackFrom(zone.Hand, false), undone: cardStillIn(zone.Hand),
		},
		{
			name: "library top with Future Sight",
			build: solRingGame("Future Sight", func(t *testing.T, g *game.Game) id.ID {
				addCardToLibraryNamed(g, game.Player1, "Below")
				return addCardToLibrary(g, game.Player1, registryCardDef(t, "Sol Ring"))
			}),
			cast: castOfCardFrom(zone.Library), done: spellOnStackFrom(zone.Library, false), undone: cardStillIn(zone.Library),
		},
		{
			name: "graveyard with Muldrotha",
			build: solRingGame("Muldrotha, the Gravetide", func(t *testing.T, g *game.Game) id.ID {
				return addCardToGraveyard(g, game.Player1, registryCardDef(t, "Sol Ring"))
			}),
			cast: castOfCardFrom(zone.Graveyard), done: spellOnStackFrom(zone.Graveyard, false), undone: cardStillIn(zone.Graveyard),
		},
		{
			// Two Thalias make Mutagenic Growth's total {2}{G/P}; its {G/P} is
			// announced as 2 life before targets (CR 601.2b).
			name: "Phyrexian announcement",
			build: func(t *testing.T) (*game.Game, *Engine, id.ID) {
				g, e, cardID := solRingGame("", func(t *testing.T, g *game.Game) id.ID {
					return addCardToHand(g, game.Player1, registryCardDef(t, "Mutagenic Growth"))
				})(t)
				addCombatPermanent(g, game.Player2, registryCardDef(t, "Thalia, Guardian of Thraben"))
				return g, e, cardID
			},
			cast: castOfCardFrom(zone.Hand),
			done: func(t *testing.T, g *game.Game, cardID id.ID) {
				t.Helper()
				spellOnStackFrom(zone.Hand, false)(t, g, cardID)
				if life := g.Players[game.Player1].Life; life != 16 {
					t.Fatalf("life = %d, want 20 less 1 per land and 2 for {G/P}", life)
				}
			},
			undone: cardStillIn(zone.Hand),
		},
		{
			name:   "mutate",
			colors: []string{"U", "G"},
			build: func(t *testing.T) (*game.Game, *Engine, id.ID) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				addCombatPermanent(g, game.Player1, mutateTargetCard())
				cardID := addCardToHand(g, game.Player1, registryCardDef(t, "Parcelbeast"))
				g.Players[game.Player1].Life = 20
				giveMainPhasePriority(g, game.Player1)
				return g, NewEngine(nil), cardID
			},
			cast: func(act action.Action, cardID id.ID) bool {
				cast, ok := act.CastSpellPayload()
				return ok && cast.CardID == cardID && cast.Mutate
			},
			done: func(t *testing.T, g *game.Game, cardID id.ID) {
				t.Helper()
				spellOnStackFrom(zone.Hand, false)(t, g, cardID)
				if obj, _ := g.Stack.Peek(); !obj.Mutate {
					t.Fatalf("stack top = %+v, want the mutating spell", obj)
				}
			},
			undone: cardStillIn(zone.Hand),
		},
		{
			name: "prepared copy",
			build: func(t *testing.T) (*game.Game, *Engine, id.ID) {
				g, e, cardID, _ := preparedExileCastGame(t, "Carnivorous Cultivator", mana.G, 2)
				addCombatPermanent(g, game.Player2, registryCardDef(t, "Thalia, Guardian of Thraben"))
				g.Players[game.Player1].Life = 20
				return g, e, cardID
			},
			cast: castOfCardFrom(zone.Battlefield),
			done: spellOnStackFrom(zone.Exile, true),
			undone: func(t *testing.T, g *game.Game, cardID id.ID) {
				t.Helper()
				if permanent := permanentForCard(g, cardID); permanent == nil || !permanent.Prepared || g.Stack.Size() != 0 {
					t.Fatalf("permanent = %+v, stack size = %d; want it still prepared and nothing cast", permanent, g.Stack.Size())
				}
			},
		},
	}
}

func findPriorityCast(e *Engine, g *game.Game, route priorityCastRoute, cardID id.ID) (action.Action, bool) {
	for _, act := range e.legalActions(g, game.Player1) {
		if route.cast(act, cardID) {
			return act, true
		}
	}
	return action.Action{}, false
}

func requireLandsUnused(t *testing.T, g *game.Game, life int, names ...string) {
	t.Helper()
	for _, name := range names {
		if reboundPermanentNamed(t, g, name).Tapped {
			t.Fatalf("%s is tapped, want the payment undone", name)
		}
	}
	if got := g.Players[game.Player1].Life; got != life {
		t.Fatalf("life = %d, want %d with the payment undone", got, life)
	}
}

// TestPriorityCastPaysWithPaymentTimeManaAbilities casts with priority on each
// route — from hand, from the top of the library with Future Sight, from the
// graveyard with Muldrotha, a cost with a Phyrexian symbol announced before
// targets, a mutate cast and a prepared copy — with a locked two-mana total
// and no mana in the pool. With Waterlogged Grove alone the cast is not
// offered. With Nurturing Peatland too it is offered, both lands are activated
// while paying (CR 601.2g) and pay the total (CR 601.2h).
func TestPriorityCastPaysWithPaymentTimeManaAbilities(t *testing.T) {
	t.Parallel()
	for _, route := range priorityCastRoutes() {
		for _, peatland := range []bool{false, true} {
			name := route.name + "/Grove alone"
			if peatland {
				name = route.name + "/Grove and Peatland"
			}
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				g, e, cardID := route.build(t)
				addCombatPermanent(g, game.Player1, registryCardDef(t, "Waterlogged Grove"))
				if peatland {
					addCombatPermanent(g, game.Player1, registryCardDef(t, "Nurturing Peatland"))
				}
				act, offered := findPriorityCast(e, g, route, cardID)
				if offered != peatland {
					t.Fatalf("cast offered = %t, want %t", offered, peatland)
				}
				if !peatland {
					return
				}
				agent := &priorityCastWindowAgent{limit: -1, colors: route.colors}
				outcome, _ := e.attemptActionWithChoices(g, game.Player1, act, reboundKickerAgents(agent), &TurnLog{}, nil)
				if outcome != actionApplied {
					t.Fatalf("outcome = %v, want the cast applied", outcome)
				}
				route.done(t, g, cardID)
				if agent.activated != 2 || !reboundPermanentNamed(t, g, "Waterlogged Grove").Tapped ||
					!reboundPermanentNamed(t, g, "Nurturing Peatland").Tapped || g.Players[game.Player1].ManaPool.Total() != 0 {
					t.Fatalf("activated = %d, pool = %d; want both lands to pay the total", agent.activated, g.Players[game.Player1].ManaPool.Total())
				}
			})
		}
	}
}

// TestUnpaidPriorityCastIsReversed casts on each route offered because both
// lands could pay its two-mana total, but activates only Waterlogged Grove
// while paying. The total can't then be paid, so the cast is reversed (CR
// 733.1): the card returns where it was, and the Grove's activation is undone
// with it, untapping the Grove and restoring the life paid. Declining the
// window is reversed the same way, and a proposal reversed twice is no longer
// accepted.
func TestUnpaidPriorityCastIsReversed(t *testing.T) {
	t.Parallel()
	for _, route := range priorityCastRoutes() {
		t.Run(route.name, func(t *testing.T) {
			t.Parallel()
			g, e, cardID := route.build(t)
			addCombatPermanent(g, game.Player1, registryCardDef(t, "Waterlogged Grove"))
			addCombatPermanent(g, game.Player1, registryCardDef(t, "Nurturing Peatland"))
			life := g.Players[game.Player1].Life
			act, ok := findPriorityCast(e, g, route, cardID)
			if !ok {
				t.Fatal("cast payable with both lands was not offered")
			}
			attempts := &castAttempts{}
			for i, limit := range []int{1, 0} {
				agent := &priorityCastWindowAgent{limit: limit, colors: route.colors}
				outcome, _ := e.attemptActionWithChoices(g, game.Player1, act, reboundKickerAgents(agent), &TurnLog{}, attempts)
				if outcome != actionReversed || agent.windows == 0 || agent.activated != limit {
					t.Fatalf("attempt %d: outcome = %v, windows = %d, activated = %d; want a reversed cast", i, outcome, agent.windows, agent.activated)
				}
				route.undone(t, g, cardID)
				requireLandsUnused(t, g, life, "Waterlogged Grove", "Nurturing Peatland")
			}
			agent := &priorityCastWindowAgent{limit: 0, colors: route.colors}
			if outcome, _ := e.attemptActionWithChoices(g, game.Player1, act, reboundKickerAgents(agent), &TurnLog{}, attempts); outcome != actionReversed {
				t.Fatalf("third attempt outcome = %v, want reversed", outcome)
			}
			if outcome, _ := e.attemptActionWithChoices(g, game.Player1, act, reboundKickerAgents(agent), &TurnLog{}, attempts); outcome == actionApplied {
				t.Fatal("a repeatedly reversed proposal was cast")
			}
			route.undone(t, g, cardID)
			requireLandsUnused(t, g, life, "Waterlogged Grove", "Nurturing Peatland")
		})
	}
}

// priorityLoopCaster casts cardID whenever offered and never activates a mana
// ability while paying.
type priorityLoopCaster struct {
	cardID id.ID
	casts  int
}

func (a *priorityLoopCaster) ChooseAction(_ PlayerObservation, legal []action.Action) action.Action {
	for _, act := range legal {
		if cast, ok := act.CastSpellPayload(); ok && cast.CardID == a.cardID {
			a.casts++
			return act
		}
	}
	return action.Pass()
}

func (*priorityLoopCaster) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	return request.DefaultSelection
}

// TestPriorityCastReversalsInPriorityLoop runs a caster who casts Sol Ring
// whenever it is offered but never activates the lands while paying. Each
// unpaid cast is reversed and the caster keeps priority (CR 733.2); once the
// repeated proposal is withheld the cast is no longer accepted, so the caster
// passes and the step ends with Sol Ring still in hand and nothing paid.
func TestPriorityCastReversalsInPriorityLoop(t *testing.T) {
	t.Parallel()
	route := priorityCastRoutes()[0]
	g, e, cardID := route.build(t)
	addCombatPermanent(g, game.Player1, registryCardDef(t, "Waterlogged Grove"))
	addCombatPermanent(g, game.Player1, registryCardDef(t, "Nurturing Peatland"))
	life := g.Players[game.Player1].Life
	agent := &priorityLoopCaster{cardID: cardID}

	e.runPriorityLoop(g, reboundKickerAgents(agent), &TurnLog{})

	if agent.casts < 2 {
		t.Fatalf("casts = %d, want repeated reversed attempts", agent.casts)
	}
	route.undone(t, g, cardID)
	requireLandsUnused(t, g, life, "Waterlogged Grove", "Nurturing Peatland")
}

// TestFutureSightVillageRitesNotPaidWithMillikin casts Village Rites off the
// top of the library with Future Sight under Thalia, with {B} in the pool, so
// its locked total is {1}{B} (CR 601.2f). Millikin's ability mills a card, so
// it isn't a mana ability (CR 605.1a) and can't be activated while paying (CR
// 601.2g): without other mana the cast isn't offered even with Millikin
// untapped. With {C} in the pool too, the cast is paid from the pool, Millikin
// is never offered while paying and the card under Village Rites stays in the
// library, the new top revealed once the spell is cast (CR 401.5).
func TestFutureSightVillageRitesNotPaidWithMillikin(t *testing.T) {
	t.Parallel()
	for _, spare := range []bool{false, true} {
		name := "no other mana"
		if spare {
			name = "spare colorless mana"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g, e, rites, next := villageRitesLibraryGame(t, "Future Sight", 0)
			addCombatPermanent(g, game.Player2, registryCardDef(t, "Thalia, Guardian of Thraben"))
			millikin := addCombatPermanent(g, game.Player1, registryCardDef(t, "Millikin"))
			p := g.Players[game.Player1]
			p.ManaPool.Add(mana.B, 1)
			if spare {
				p.ManaPool.Add(mana.C, 1)
			}
			act := action.CastSpellFaceFromZone(rites, zone.Library, game.FaceFront, nil, 0, nil)
			if offered := actionsContain(e.legalActions(g, game.Player1), act); offered != spare {
				t.Fatalf("Village Rites offered = %t, want %t: Millikin can't pay while casting", offered, spare)
			}
			if !spare {
				return
			}
			agent := &windowOptionsAgent{priorityCastWindowAgent: priorityCastWindowAgent{limit: -1}}

			outcome, _ := e.attemptActionWithChoices(g, game.Player1, act, reboundKickerAgents(agent), &TurnLog{}, nil)

			if outcome != actionApplied {
				t.Fatalf("outcome = %v, want Village Rites cast", outcome)
			}
			assertVillageRitesCast(t, g, rites)
			requireNotOffered(t, agent.offered, millProducerMana)
			if millikin.Tapped || !p.Library.Contains(next) || p.Graveyard.Contains(next) {
				t.Fatalf("Millikin tapped = %t, next card milled = %t; want Millikin unused", millikin.Tapped, p.Graveyard.Contains(next))
			}
			if obj, _ := g.Stack.Peek(); obj.ManaSpentToCast != 2 || p.ManaPool.Total() != 0 {
				t.Fatalf("mana spent = %d, pool = %d; want Thalia's {1} and {B} paid from the pool", obj.ManaSpentToCast, p.ManaPool.Total())
			}
			assertTopShown(t, g, "Future Sight", next)
			assertNoLibraryTopWindow(t, g)
		})
	}
}

// manaOrderAgent activates, while paying, the mana abilities whose labels
// contain each of order in turn, then finishes. It chooses the first offered
// of colors when a mana ability adds one of several.
type manaOrderAgent struct {
	order     []string
	colors    []string
	activated int
}

func (*manaOrderAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *manaOrderAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Prompt == "Activate a mana ability while paying?" {
		if a.activated < len(a.order) {
			for _, option := range request.Options {
				if strings.Contains(option.Label, a.order[a.activated]) {
					a.activated++
					return []int{option.Index}
				}
			}
		}
		return request.DefaultSelection
	}
	for _, want := range a.colors {
		for _, option := range request.Options {
			if option.Label == want {
				return []int{option.Index}
			}
		}
	}
	return request.DefaultSelection
}

// Labels of the mana abilities of Silent Clearing and Waterlogged Grove.
const (
	clearingMana = "Pay 1 life: Add {W} or {B}"
	groveMana    = "Pay 1 life: Add {G} or {U}"
)

// TestUnpaidPriorityCastReversesManaAroundShuffle casts on each route offered
// because Waterlogged Grove, Nurturing Peatland and Phyrexian Altar together
// could pay its total, raised to three mana by another Thalia for a
// noncreature spell, but activates only the Grove and then the Altar,
// sacrificing Blightsteel Colossus, while paying. The Altar's ability is a
// mana ability: the Colossus's replacement effect that shuffles it into its
// owner's library isn't the Altar's own (CR 605.1a). The total can't be paid,
// so the cast is reversed (CR 733.1). The shuffle can't be reversed, but the
// Grove's activation came before it and none of its mana was spent on the
// Altar, so it is reversed anyway: the Grove untaps and the life is restored,
// while the Colossus stays in the library and the Altar's mana in the pool.
func TestUnpaidPriorityCastReversesManaAroundShuffle(t *testing.T) {
	t.Parallel()
	for _, route := range priorityCastRoutes() {
		t.Run(route.name, func(t *testing.T) {
			t.Parallel()
			g, e, cardID := route.build(t)
			if route.name != "mutate" {
				addCombatPermanent(g, game.Player2, registryCardDef(t, "Thalia, Guardian of Thraben"))
			}
			addCombatPermanent(g, game.Player1, registryCardDef(t, "Waterlogged Grove"))
			addCombatPermanent(g, game.Player1, registryCardDef(t, "Nurturing Peatland"))
			colossus := addAltarAndColossus(t, g, game.Player1)
			p := g.Players[game.Player1]
			life := p.Life
			act, ok := findPriorityCast(e, g, route, cardID)
			if !ok {
				t.Fatal("cast payable with both lands and Phyrexian Altar was not offered")
			}
			order := &manaOrderAgent{order: []string{groveMana, altarMana}, colors: route.colors}
			agent := &sacrificingAgent{next: order, victim: "Blightsteel Colossus"}

			outcome, _ := e.attemptActionWithChoices(g, game.Player1, act, reboundKickerAgents(agent), &TurnLog{}, nil)

			if outcome != actionReversed || order.activated != 2 || agent.sacrifices != 1 {
				t.Fatalf("outcome = %v, activated = %d, Colossus sacrificed %d times; want a reversed cast after two activations", outcome, order.activated, agent.sacrifices)
			}
			route.undone(t, g, cardID)
			requireLandsUnused(t, g, life, "Waterlogged Grove", "Nurturing Peatland")
			requireColossusShuffled(t, g, colossus)
			if got := p.ManaPool.Total(); got != 1 {
				t.Fatalf("pool = %d, want the Altar's mana kept", got)
			}
		})
	}
}

// colossusFilterDef is a test artifact whose mana ability spends mana and
// sacrifices a creature: "{1}, Sacrifice a creature: Add {C}{C}."
func colossusFilterDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Sacrifice Filter",
		Types: []types.Card{types.Artifact},
		ManaAbilities: []game.ManaAbility{{
			Text:     "{1}, Sacrifice a creature: Add {C}{C}.",
			ManaCost: opt.Val(cost.Mana{cost.O(1)}),
			AdditionalCosts: []cost.Additional{{
				Kind: cost.AdditionalSacrifice, Text: "Sacrifice a creature", Amount: 1,
				MatchPermanentType: true, PermanentType: types.Creature,
			}},
			Content: game.AbilityContent{
				Modes:    []game.Mode{{Sequence: []game.Instruction{{Primitive: game.AddMana{Amount: game.Fixed(2), ManaColor: mana.C}}}}},
				MinModes: 1, MaxModes: 1,
			},
		}},
	}}
}

// TestUnpaidVillageRitesReversesIndependentMana casts Village Rites off the
// top of the library with Future Sight under Thalia, a locked {1}{B} total
// (CR 601.2f), and leaves it unpaid after activating mana abilities while
// paying (CR 601.2g), so the cast is reversed (CR 733.1). A player may reverse
// legal mana abilities activated while casting it, unless mana from them was
// spent on another mana ability that isn't reversed, but not an action that
// moved a card to a library (CR 733.1), here Blightsteel Colossus shuffled
// into the library when sacrificed by a mana ability (CR 605.1a):
//   - Silent Clearing for {W}, Phyrexian Altar sacrificing the Colossus, then
//     Waterlogged Grove for {G}: both lands are reversed, before and after the
//     shuffle, and the Altar's activation is kept.
//   - Silent Clearing for {W}, then Sacrifice Filter spending that {W} and
//     sacrificing the Colossus: the Filter's activation is kept, so the
//     Clearing, whose mana it spent, is kept too.
func TestUnpaidVillageRitesReversesIndependentMana(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name     string
		order    []string
		reversed []string
		kept     string
		life     int
		pool     int
	}{
		{name: "lands around Phyrexian Altar", order: []string{clearingMana, altarMana, groveMana}, reversed: []string{"Silent Clearing", "Waterlogged Grove"}, life: 20, pool: 1},
		{name: "land spent on Sacrifice Filter", order: []string{clearingMana, "Add {C}{C}"}, reversed: []string{"Waterlogged Grove"}, kept: "Silent Clearing", life: 19, pool: 2},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			g, e, rites, _ := villageRitesLibraryGame(t, "Future Sight", 0)
			addCombatPermanent(g, game.Player2, registryCardDef(t, "Thalia, Guardian of Thraben"))
			for _, name := range []string{"Silent Clearing", "Waterlogged Grove"} {
				addCombatPermanent(g, game.Player1, registryCardDef(t, name))
			}
			addCombatPermanent(g, game.Player1, colossusFilterDef())
			colossus := addAltarAndColossus(t, g, game.Player1)
			act := action.CastSpellFaceFromZone(rites, zone.Library, game.FaceFront, nil, 0, nil)
			if !actionsContain(e.legalActions(g, game.Player1), act) {
				t.Fatal("Village Rites payable while paying was not offered")
			}
			order := &manaOrderAgent{order: test.order, colors: []string{"W", "G"}}
			agent := &sacrificingAgent{next: order, victim: "Blightsteel Colossus"}

			outcome, _ := e.attemptActionWithChoices(g, game.Player1, act, reboundKickerAgents(agent), &TurnLog{}, nil)

			if outcome != actionReversed || order.activated != len(test.order) || agent.sacrifices != 1 {
				t.Fatalf("outcome = %v, activated = %d, Colossus sacrificed %d times; want a reversed cast after %d activations", outcome, order.activated, agent.sacrifices, len(test.order))
			}
			p := g.Players[game.Player1]
			if top, _ := p.Library.Top(); top != rites {
				t.Fatalf("library top = %v, want Village Rites back on top", top)
			}
			requireColossusShuffled(t, g, colossus)
			requireLandsUnused(t, g, test.life, test.reversed...)
			if test.kept != "" && !reboundPermanentNamed(t, g, test.kept).Tapped {
				t.Fatalf("%s is untapped, want its activation kept", test.kept)
			}
			if got := p.ManaPool.Total(); got != test.pool {
				t.Fatalf("pool = %d, want %d", got, test.pool)
			}
		})
	}
}
