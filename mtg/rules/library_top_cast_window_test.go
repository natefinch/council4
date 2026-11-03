package rules

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: Engine gap #3624 (sub-issue of #3562). Default-registry cards
// checked against the pinned corpus 6da9ced0…a1b4 (oracle_id): Future Sight
// 93a0e938-b916-47ec-9803-369e885e253c {2}{U}{U}{U} "Play with the top card of
// your library revealed. You may play lands and cast spells from the top of
// your library."; Bolas's Citadel 2bd111bb-ce02-414c-b5b7-e0e037d8d96b "You may
// look at the top card of your library any time. You may play lands and cast
// spells from the top of your library. If you cast a spell this way, pay life
// equal to its mana value rather than pay its mana cost. …"; Village Rites
// 365548fb-5acc-4a8a-b20b-26d28b7d029f {B} "As an additional cost to cast this
// spell, sacrifice a creature. Draw two cards."; Parcelbeast
// de83b309-ac31-4a57-95e6-ccd3659e6c1f "Mutate {G}{U} …"; Millikin
// fa4dffda-6f04-4d0b-829d-28a1a5794dee "{T}, Mill a card: Add {C}. …"; Silent
// Clearing fd45063f-c83c-431c-9104-f139c497ec0d "{T}, Pay 1 life: Add {W} or
// {B}. …"; Waterlogged Grove 70fa2eba-565e-4fed-adc9-7f5d9fcbf1fa; Thalia,
// Guardian of Thraben 9b7f1d05-707c-4ed3-9f0e-8ced1232c2ee "Noncreature spells
// cost {1} more to cast."; Muldrotha, the Gravetide
// e4625704-1d52-44e4-804f-2f45644d76ac; War Behemoth
// 4792a05e-b557-4132-8303-235feefe0d99 (morph); Conduit of Worlds
// ed14be15-8f8d-4fe3-a147-f5da8ed873bf; Sol Ring
// 6ad8011d-3471-4369-9d68-b264cc027487. Rulings: Future Sight 2019-06-14 "If
// the top card of your library changes while you're casting a spell, playing a
// land, or activating an ability, the new top card won't be revealed until you
// finish doing so."; Bolas's Citadel 2019-05-03, the same for looking at the
// new top card. Comprehensive Rules effective 2026-09-25: 401.5 "If an effect
// causes a player to play with the top card of their library revealed, or
// allows a player to look at the top card of their library, and the top card
// of that library changes during the process of casting a spell or activating
// an ability, the new top card won't be revealed, and can't be looked at, until
// the spell becomes cast or the ability becomes activated (see rule 601.2i)";
// 601.2a moves the card to the stack first; 601.2g/h pay costs, mana abilities
// included; 601.2i the spell becomes cast; 733.1 a reversed cast doesn't undo
// a mill from the library. Activations, land plays and special actions are
// #3625.

// libraryTopWindowAgent answers a cast's choices by default, optionally
// activating only Millikin's mana ability while paying, and checks at each
// choice what every observer is shown of owner's library top: an unchanged
// top (start) may stay visible, but a changed one must be hidden from all.
type libraryTopWindowAgent struct {
	owner game.PlayerID
	start id.ID
	mill  bool
	// onChanged runs once, at the first choice made after the top changed;
	// onUnchanged runs at each choice made before.
	onChanged, onUnchanged func(g *game.Game)

	milled, unchanged, unchangedVisible, changed int
	leaks                                        []string
}

func (*libraryTopWindowAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *libraryTopWindowAgent) ChooseChoice(obs PlayerObservation, request game.ChoiceRequest) []int {
	a.check(obs, request.Prompt)
	switch {
	case request.Kind == game.ChoiceMay:
		return []int{1}
	case a.mill && a.milled == 0 && request.Prompt == "Activate a mana ability while paying?":
		for _, option := range request.Options {
			if strings.Contains(option.Label, "Mill a card") {
				a.milled++
				return []int{option.Index}
			}
		}
	}
	return request.DefaultSelection
}

func (a *libraryTopWindowAgent) check(obs PlayerObservation, prompt string) {
	g := obs.g
	top, ok := g.Players[a.owner].Library.Top()
	observations := []PlayerObservation{obs}
	for viewer := range game.NumPlayers {
		if g.Players[viewer] != nil {
			observations = append(observations, NewObservation(g, game.PlayerID(viewer)))
		}
	}
	if ok && top == a.start {
		a.unchanged++
		if topCardShown(NewObservation(g, a.owner), a.owner) == top {
			a.unchangedVisible++
		}
		if a.onUnchanged != nil {
			a.onUnchanged(g)
		}
		return
	}
	a.changed++
	for _, o := range observations {
		if shown := topCardShown(o, a.owner); shown != 0 {
			a.leaks = append(a.leaks, fmt.Sprintf("viewer %d saw %v at %q", o.viewer, shown, prompt))
		}
	}
	if a.onChanged != nil {
		a.onChanged(g)
		a.onChanged = nil
	}
}

// topCardShown is the card o may see on top of player's library, revealed or
// privately looked at, or zero.
func topCardShown(o PlayerObservation, player game.PlayerID) id.ID {
	if card, ok := o.LibraryTopRevealed(player); ok {
		return card.CardInstanceID
	}
	if card, ok := o.LibraryTopLookable(player); ok {
		return card.CardInstanceID
	}
	return 0
}

func (a *libraryTopWindowAgent) agents() [game.NumPlayers]PlayerAgent {
	var agents [game.NumPlayers]PlayerAgent
	agents[a.owner] = a
	return agents
}

// assertHiddenWhileChanged requires at least one choice after the top changed
// and that no observer was shown the changed top during any of them.
func (a *libraryTopWindowAgent) assertHiddenWhileChanged(t *testing.T) {
	t.Helper()
	if a.changed == 0 {
		t.Fatal("no choice was made after the library top changed")
	}
	if len(a.leaks) != 0 {
		t.Fatalf("the changed library top was shown before the spell became cast (CR 401.5): %v", a.leaks)
	}
}

func assertNoLibraryTopWindow(t *testing.T, g *game.Game) {
	t.Helper()
	if len(g.LibraryTopWindows) != 0 || g.LibraryTopHidden(game.Player1) {
		t.Fatalf("library top windows = %v, want none open after the action", g.LibraryTopWindows)
	}
}

// villageRitesLibraryGame puts default-registry Village Rites on top of
// Player1's library over a card it returns, behind fillers, with permission
// granting Player1 two creatures to choose a sacrifice from.
func villageRitesLibraryGame(t *testing.T, permission string, fillers int) (*game.Game, *Engine, id.ID, id.ID) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, registryCardDef(t, permission))
	addCombatPermanent(g, game.Player1, vanillaCreature("First Bear", 2, 2))
	addCombatPermanent(g, game.Player1, vanillaCreature("Second Bear", 2, 2))
	for i := range fillers {
		addCardToLibraryNamed(g, game.Player1, fmt.Sprintf("Filler %d", i))
	}
	next := addCardToLibraryNamed(g, game.Player1, "Next Card")
	rites := addCardToLibrary(g, game.Player1, registryCardDef(t, "Village Rites"))
	g.Players[game.Player1].Life = 20
	giveMainPhasePriority(g, game.Player1)
	return g, NewEngine(nil), rites, next
}

func assertVillageRitesCast(t *testing.T, g *game.Game, rites id.ID) {
	t.Helper()
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != rites || len(obj.SacrificedAsCostIDs) != 1 {
		t.Fatalf("stack top = %+v, want Village Rites cast with a creature sacrificed", obj)
	}
}

// TestLibraryTopChangedByCastFromLibraryStaysHidden casts Village Rites off
// the top of Player1's library on each library cast route. The card moves to
// the stack before its sacrifice is chosen (CR 601.2a), so the next card is
// the new top; it isn't revealed (Future Sight) or lookable (Bolas's Citadel)
// to anyone while the sacrifice and payment are chosen, and becomes so once
// the spell is cast (CR 401.5, 601.2i).
func TestLibraryTopChangedByCastFromLibraryStaysHidden(t *testing.T) {
	t.Parallel()
	routes := []struct {
		name       string
		permission string
		cast       func(t *testing.T, e *Engine, g *game.Game, rites id.ID, agents [game.NumPlayers]PlayerAgent)
	}{
		{name: "Future Sight cast", permission: "Future Sight", cast: func(t *testing.T, e *Engine, g *game.Game, rites id.ID, agents [game.NumPlayers]PlayerAgent) {
			g.Players[game.Player1].ManaPool.Add(mana.B, 1)
			cast := action.CastSpellFaceFromZone(rites, zone.Library, game.FaceFront, nil, 0, nil)
			if outcome, _ := e.attemptActionWithChoices(g, game.Player1, cast, agents, &TurnLog{}, nil); outcome != actionApplied {
				t.Fatalf("outcome = %v, want Village Rites cast from the library", outcome)
			}
		}},
		{name: "Bolas's Citadel pay-life cast", permission: "Bolas's Citadel", cast: func(t *testing.T, e *Engine, g *game.Game, rites id.ID, agents [game.NumPlayers]PlayerAgent) {
			cast := action.CastSpellFaceFromZone(rites, zone.Library, game.FaceFront, nil, 0, nil)
			if outcome, _ := e.attemptActionWithChoices(g, game.Player1, cast, agents, &TurnLog{}, nil); outcome != actionApplied {
				t.Fatalf("outcome = %v, want Village Rites cast from the library for 1 life", outcome)
			}
			if life := g.Players[game.Player1].Life; life != 19 {
				t.Fatalf("life = %d, want 19 after paying Village Rites' mana value", life)
			}
		}},
		{name: "paid cast during resolution", permission: "Future Sight", cast: func(t *testing.T, e *Engine, g *game.Game, rites id.ID, agents [game.NumPlayers]PlayerAgent) {
			g.Players[game.Player1].ManaPool.Add(mana.B, 1)
			if !e.castPaidSpellFromSource(g, g.Players[game.Player1], game.Player1, rites, zone.Library, agents, &TurnLog{}) {
				t.Fatal("the resolution-time paid cast from the library failed")
			}
		}},
		{name: "free cast during resolution", permission: "Future Sight", cast: func(t *testing.T, e *Engine, g *game.Game, rites id.ID, agents [game.NumPlayers]PlayerAgent) {
			if !e.chooseAndCastFreeSpellFromSource(g, g.Players[game.Player1], game.Player1, rites, zone.Library,
				game.Selection{}, opt.V[int]{}, false, agents, &TurnLog{}) {
				t.Fatal("the resolution-time free cast from the library failed")
			}
		}},
	}
	for _, route := range routes {
		t.Run(route.name, func(t *testing.T) {
			t.Parallel()
			g, e, rites, next := villageRitesLibraryGame(t, route.permission, 0)
			agent := &libraryTopWindowAgent{owner: game.Player1, start: rites}

			route.cast(t, e, g, rites, agent.agents())

			agent.assertHiddenWhileChanged(t)
			assertVillageRitesCast(t, g, rites)
			assertNoLibraryTopWindow(t, g)
			owner, opponent := NewObservation(g, game.Player1), NewObservation(g, game.Player2)
			if route.permission == "Bolas's Citadel" {
				if card, ok := owner.LibraryTopLookable(game.Player1); !ok || card.CardInstanceID != next {
					t.Fatal("Player1 can't look at the new top card after the spell became cast")
				}
				if topCardShown(opponent, game.Player1) != 0 {
					t.Fatal("Citadel's private look reached the opponent")
				}
				return
			}
			for _, o := range []PlayerObservation{owner, opponent} {
				if card, ok := o.LibraryTopRevealed(game.Player1); !ok || card.CardInstanceID != next {
					t.Fatalf("viewer %d isn't shown the new top card after the spell became cast", o.viewer)
				}
			}
		})
	}
}

// TestLibraryTopChangedByMutateCastStaysHidden mutate-casts Parcelbeast off
// the top of Player1's library with two cast-from-top permissions, so the
// permission is chosen after the card has moved to the stack: the card below
// stays hidden through that choice and is revealed once the spell is cast.
func TestLibraryTopChangedByMutateCastStaysHidden(t *testing.T) {
	t.Parallel()
	f := newParcelbeastLibraryFixture(t, "Future Sight")
	addCombatPermanent(f.game, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Once Seer",
		StaticAbilities: []game.StaticAbility{{RuleEffects: []game.RuleEffect{{
			Kind:           game.RuleEffectCastSpellsFromZone,
			AffectedPlayer: game.PlayerYou,
			CastFromZone:   zone.Library,
			TopCardOnly:    true,
			OncePerTurn:    true,
		}}}},
	}})
	agent := &libraryTopWindowAgent{owner: game.Player1, start: f.cardID}

	if !f.engine.applyActionWithChoices(f.game, game.Player1, f.mutate(), agent.agents(), &TurnLog{}) {
		t.Fatal("the library mutate cast failed")
	}

	agent.assertHiddenWhileChanged(t)
	assertNoLibraryTopWindow(t, f.game)
	if obj, ok := f.game.Stack.Peek(); !ok || obj.SourceID != f.cardID || !obj.Mutate {
		t.Fatalf("stack top = %+v, want the mutating Parcelbeast", obj)
	}
	if card, ok := NewObservation(f.game, game.Player2).LibraryTopRevealed(game.Player1); !ok || card.CardInstanceID != f.below {
		t.Fatal("the card below Parcelbeast isn't revealed after the spell became cast")
	}
}

// TestLibraryTopWindowSurvivesCloneAndDeterminize clones the game while the
// sacrifice for a Future Sight cast of Village Rites is chosen. The clone keeps
// the open window, so its observation hides the new top card and a
// determinized search world doesn't pin it on top; changing the clone's window
// leaves the original's alone.
func TestLibraryTopWindowSurvivesCloneAndDeterminize(t *testing.T) {
	t.Parallel()
	g, e, rites, next := villageRitesLibraryGame(t, "Future Sight", 10)
	g.Players[game.Player1].ManaPool.Add(mana.B, 1)
	agent := &libraryTopWindowAgent{owner: game.Player1, start: rites}
	var checked bool
	agent.onChanged = func(live *game.Game) {
		checked = true
		clone := live.Clone()
		if len(clone.LibraryTopWindows) != 1 || !clone.LibraryTopHidden(game.Player1) {
			t.Fatalf("clone windows = %v, want the open cast window kept", clone.LibraryTopWindows)
		}
		if topCardShown(NewObservation(clone, game.Player1), game.Player1) != 0 {
			t.Fatal("the clone's observation shows the changed top card")
		}
		pinned := 0
		for seed := range uint64(20) {
			world := clone.Clone()
			determinize(world, game.Player1, rand.New(rand.NewPCG(seed, seed)))
			if top, _ := world.Players[game.Player1].Library.Top(); top == next {
				pinned++
			}
		}
		if pinned == 20 {
			t.Fatal("determinize pinned the hidden new top card in every search world")
		}
		clone.LibraryTopWindows[0][game.Player1] = next
		if clone.LibraryTopHidden(game.Player1) || !live.LibraryTopHidden(game.Player1) {
			t.Fatal("the clone's window is shared with the live game")
		}
	}
	cast := action.CastSpellFaceFromZone(rites, zone.Library, game.FaceFront, nil, 0, nil)

	if outcome, _ := e.attemptActionWithChoices(g, game.Player1, cast, agent.agents(), &TurnLog{}, nil); outcome != actionApplied {
		t.Fatalf("outcome = %v, want Village Rites cast from the library", outcome)
	}

	if !checked {
		t.Fatal("the clone check never ran")
	}
	agent.assertHiddenWhileChanged(t)
	assertNoLibraryTopWindow(t, g)
	if card, ok := NewObservation(g, game.Player1).LibraryTopRevealed(game.Player1); !ok || card.CardInstanceID != next {
		t.Fatal("the new top card isn't revealed after the spell became cast")
	}
}

// millDuringPaymentRoute builds a reversible cast, from a zone other than the
// library, that opens a payment-time mana ability window in which Player1 can
// pay only with Millikin plus another manual mana source. (A cast from hand
// with priority pays from mana activated beforehand.)
type millDuringPaymentRoute struct {
	name string
	run  func(t *testing.T) (*game.Game, func([game.NumPlayers]PlayerAgent))
}

func millDuringPaymentRoutes() []millDuringPaymentRoute {
	routes := []millDuringPaymentRoute{
		{name: "madness cast", run: func(t *testing.T) (*game.Game, func([game.NumPlayers]PlayerAgent)) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			giveMainPhasePriority(g, game.Player1)
			addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Silent Clearing"))
			card := addCardToExile(g, game.Player1, productionChoiceCard(t, "Sol Ring"))
			e := NewEngine(nil)
			return g, func(agents [game.NumPlayers]PlayerAgent) {
				if e.castMadnessSpellWithChoices(g, game.Player1, g.CardInstances[card], cost.Mana{cost.O(2)}, agents, &TurnLog{}) {
					t.Fatal("the unpaid madness cast succeeded")
				}
				if g.Stack.Size() != 0 || !g.Players[game.Player1].Exile.Contains(card) {
					t.Fatal("the unpaid madness cast didn't leave the card in exile")
				}
			}
		}},
	}
	for _, route := range graveyardRestoreRoutes() {
		routes = append(routes, millDuringPaymentRoute{name: route.name, run: func(t *testing.T) (*game.Game, func([game.NumPlayers]PlayerAgent)) {
			g, _, run := route.build(t)
			addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Waterlogged Grove"))
			return g, run
		}})
	}
	return routes
}

// TestLibraryTopChangedByPaymentStaysHidden casts a spell from outside the
// library under Future Sight on each remaining cast route. The revealed top
// card stays revealed while the cast begins, but Millikin's mill during
// payment (CR 601.2g) changes the top, and the new top stays hidden through
// the rest of the payment. The cost goes unpaid and the cast is reversed (CR
// 733.1), which ends the cast without undoing the mill, so the new top card is
// then revealed.
func TestLibraryTopChangedByPaymentStaysHidden(t *testing.T) {
	t.Parallel()
	for _, route := range millDuringPaymentRoutes() {
		t.Run(route.name, func(t *testing.T) {
			t.Parallel()
			g, run := route.run(t)
			addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Future Sight"))
			addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Millikin"))
			next := addCardToLibraryNamed(g, game.Player1, "Next Card")
			start := addCardToLibraryNamed(g, game.Player1, "Revealed Card")
			agent := &libraryTopWindowAgent{owner: game.Player1, start: start, mill: true}

			run(agent.agents())

			if agent.milled != 1 || !g.Players[game.Player1].Graveyard.Contains(start) {
				t.Fatal("Millikin didn't mill the revealed card while paying")
			}
			if agent.unchanged == 0 || agent.unchangedVisible != agent.unchanged {
				t.Fatalf("the unchanged top was visible at %d of %d choices, want all", agent.unchangedVisible, agent.unchanged)
			}
			agent.assertHiddenWhileChanged(t)
			assertNoLibraryTopWindow(t, g)
			if card, ok := NewObservation(g, game.Player2).LibraryTopRevealed(game.Player1); !ok || card.CardInstanceID != next {
				t.Fatal("the new top card isn't revealed after the cast was reversed")
			}
		})
	}
}

// TestUnchangedLibraryTopStaysRevealedWhileCastingFromHand casts Village Rites
// from hand under Future Sight. The library's top card doesn't change, so it
// stays revealed to every player while the sacrifice is chosen (CR 401.5 only
// withholds a new top card).
func TestUnchangedLibraryTopStaysRevealedWhileCastingFromHand(t *testing.T) {
	t.Parallel()
	g, e, inLibrary, _ := villageRitesLibraryGame(t, "Future Sight", 0)
	rites := addCardToHand(g, game.Player1, registryCardDef(t, "Village Rites"))
	g.Players[game.Player1].ManaPool.Add(mana.B, 1)
	agent := &libraryTopWindowAgent{owner: game.Player1, start: inLibrary}
	opponentSaw := 0
	agent.onUnchanged = func(live *game.Game) {
		if card, ok := NewObservation(live, game.Player2).LibraryTopRevealed(game.Player1); ok && card.CardInstanceID == inLibrary {
			opponentSaw++
		}
	}

	if outcome, _ := e.attemptActionWithChoices(g, game.Player1, action.CastSpell(rites, nil, 0, nil), agent.agents(), &TurnLog{}, nil); outcome != actionApplied {
		t.Fatalf("outcome = %v, want Village Rites cast from hand", outcome)
	}

	assertVillageRitesCast(t, g, rites)
	assertNoLibraryTopWindow(t, g)
	if agent.changed != 0 || agent.unchanged == 0 || agent.unchangedVisible != agent.unchanged || opponentSaw != agent.unchanged {
		t.Fatalf("choices: %d unchanged (%d shown to Player1, %d to Player2), %d changed; want the unchanged top revealed at every choice",
			agent.unchanged, agent.unchangedVisible, opponentSaw, agent.changed)
	}
}
