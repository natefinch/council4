package rules

import (
	"fmt"
	"slices"
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

// Source review: Engine gap #3645. Default-registry cards checked against the
// pinned corpus 6da9ced0…a1b4 (oracle_id): Rest in Peace
// 087f9ad7-e74f-40e2-8102-1ed2925d0418 "If a card or token would be put into a
// graveyard from anywhere, exile it instead."; Progenitus
// 2d7e00b6-12f0-4b03-82a6-50e3d1b5395e "If Progenitus would be put into a
// graveyard from anywhere, reveal Progenitus and shuffle it into its owner's
// library instead."; Cathartic Reunion 0f3c3e5f-6af3-4af2-8703-4ccc8ed8f675
// {1}{R} "As an additional cost to cast this spell, discard two cards. Draw
// three cards."; Zombie Infestation bdce1af0-3643-4e77-88f9-320206a191d4
// "Discard two cards: Create a 2/2 black Zombie creature token."; Stitchwing
// Skaab 0b40e537-d385-450d-93a9-b60a25bfe508 "{1}{U}, Discard two cards:
// Return this card from your graveyard to the battlefield tapped."
// Comprehensive Rules effective 2026-09-25 (official copy retained under
// .cardwork/cr-20260925.txt): 101.4 choices are made first, "Then the actions
// happen simultaneously."; 601.2h/602.2b the player pays the total cost, the
// nonrandom costs first; 614.1 replacement effects apply as the event happens;
// 616.1 the affected object's owner chooses which replacement applies; 701.9a
// discard moves a card from its owner's hand to that player's graveyard.

// discardCostRoute names one way a "discard two cards" cost is paid. setup
// prepares the payment's source and mana and returns the action that pays it.
type discardCostRoute struct {
	name  string
	setup func(*testing.T, *game.Game) action.Action
	// keyed routes publish both discarded Progenitus as paid-cost subjects.
	keyed bool
	check func(t *testing.T, g *game.Game, life int)
}

// keyedDiscardCostSpell is a sorcery whose additional costs combine paying 2
// life with discarding two cards bound as paid-cost subjects.
func keyedDiscardCostSpell() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Keyed Discard Spell",
		ManaCost: opt.Val(cost.Mana{cost.R}),
		Types:    []types.Card{types.Sorcery},
		AdditionalCosts: []cost.Additional{
			{Kind: cost.AdditionalPayLife, Text: "pay 2 life", Amount: 2},
			{Kind: cost.AdditionalDiscard, Text: "discard two cards", Amount: 2, Source: zone.Hand, SubjectKey: "discarded"},
		},
		SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{{
			Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()},
		}}}.Ability()),
	}}
}

// keyedDiscardCostArtifact has one ability whose cost combines {1}, paying 2
// life, sacrificing a creature and discarding two cards bound as paid-cost
// subjects.
func keyedDiscardCostArtifact() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Keyed Discard Engine",
		Types: []types.Card{types.Artifact},
		ActivatedAbilities: []game.ActivatedAbility{{
			Text:     "{1}, Pay 2 life, Sacrifice a creature, Discard two cards: You gain 1 life.",
			ManaCost: opt.Val(cost.Mana{cost.O(1)}),
			AdditionalCosts: []cost.Additional{
				{Kind: cost.AdditionalPayLife, Text: "Pay 2 life", Amount: 2},
				{Kind: cost.AdditionalSacrifice, Text: "Sacrifice a creature", Amount: 1, MatchPermanentType: true, PermanentType: types.Creature},
				{Kind: cost.AdditionalDiscard, Text: "Discard two cards", Amount: 2, Source: zone.Hand, SubjectKey: "discarded"},
			},
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()},
			}}}.Ability(),
		}},
	}}
}

func discardCostRoutes() []discardCostRoute {
	stackTopNamed := func(t *testing.T, g *game.Game, kind game.StackObjectKind, name string) *game.StackObject {
		t.Helper()
		obj, ok := g.Stack.Peek()
		if !ok || obj.Kind != kind || obj.Controller != game.Player1 {
			t.Fatalf("stack top = %+v, want Player1's paid %s", obj, name)
		}
		cardID := obj.SourceCardID
		if kind == game.StackSpell {
			cardID = obj.SourceID
		}
		if card, ok := g.GetCardInstance(cardID); !ok || card.Def == nil || card.Def.Name != name {
			t.Fatalf("stack top source card %v is not %s", cardID, name)
		}
		return obj
	}
	var sacrificed *game.Permanent
	return []discardCostRoute{
		{
			name: "Cathartic Reunion cast from hand",
			setup: func(t *testing.T, g *game.Game) action.Action {
				spellID := addCardToHand(g, game.Player1, registryCardDef(t, "Cathartic Reunion"))
				g.Players[game.Player1].ManaPool.Add(mana.R, 2)
				return action.CastSpell(spellID, nil, 0, nil)
			},
			check: func(t *testing.T, g *game.Game, _ int) {
				stackTopNamed(t, g, game.StackSpell, "Cathartic Reunion")
			},
		},
		{
			name: "Cathartic Reunion cast from exile",
			setup: func(t *testing.T, g *game.Game) action.Action {
				spellID := addCardToHand(g, game.Player1, registryCardDef(t, "Cathartic Reunion"))
				if !moveCardBetweenZones(g, game.Player1, spellID, zone.Hand, zone.Exile) {
					t.Fatal("could not exile Cathartic Reunion")
				}
				grantCastFromExile(g, game.Player1, spellID)
				g.Players[game.Player1].ManaPool.Add(mana.R, 2)
				return action.CastSpellFromZone(spellID, zone.Exile, nil, 0, nil)
			},
			check: func(t *testing.T, g *game.Game, _ int) {
				if obj := stackTopNamed(t, g, game.StackSpell, "Cathartic Reunion"); obj.SourceZone != zone.Exile {
					t.Fatalf("Cathartic Reunion cast from %v, want exile", obj.SourceZone)
				}
			},
		},
		{
			name: "Zombie Infestation activated on the battlefield",
			setup: func(t *testing.T, g *game.Game) action.Action {
				source := addCombatPermanent(g, game.Player1, registryCardDef(t, "Zombie Infestation"))
				return action.ActivateAbility(source.ObjectID, 0, nil, 0)
			},
			check: func(t *testing.T, g *game.Game, _ int) {
				stackTopNamed(t, g, game.StackActivatedAbility, "Zombie Infestation")
			},
		},
		{
			name: "Stitchwing Skaab activated from the graveyard",
			setup: func(t *testing.T, g *game.Game) action.Action {
				skaabID := addCardToGraveyard(g, game.Player1, registryCardDef(t, "Stitchwing Skaab"))
				g.Players[game.Player1].ManaPool.Add(mana.U, 2)
				return action.ActivateAbility(skaabID, 0, nil, 0)
			},
			check: func(t *testing.T, g *game.Game, _ int) {
				stackTopNamed(t, g, game.StackActivatedAbility, "Stitchwing Skaab")
			},
		},
		{
			name:  "keyed discard combined with life on a spell",
			keyed: true,
			setup: func(_ *testing.T, g *game.Game) action.Action {
				spellID := addCardToHand(g, game.Player1, keyedDiscardCostSpell())
				g.Players[game.Player1].ManaPool.Add(mana.R, 1)
				return action.CastSpell(spellID, nil, 0, nil)
			},
			check: func(t *testing.T, g *game.Game, life int) {
				stackTopNamed(t, g, game.StackSpell, "Keyed Discard Spell")
				if got := g.Players[game.Player1].Life; got != life-2 {
					t.Fatalf("life = %d, want %d after paying 2 life", got, life-2)
				}
			},
		},
		{
			name:  "keyed discard combined with mana, life and sacrifice on an ability",
			keyed: true,
			setup: func(_ *testing.T, g *game.Game) action.Action {
				source := addCombatPermanent(g, game.Player1, keyedDiscardCostArtifact())
				sacrificed = addCombatPermanent(g, game.Player1, vanillaCreature("Sacrificed creature", 1, 1))
				g.Players[game.Player1].ManaPool.Add(mana.C, 1)
				return action.ActivateAbility(source.ObjectID, 0, nil, 0)
			},
			check: func(t *testing.T, g *game.Game, life int) {
				stackTopNamed(t, g, game.StackActivatedAbility, "Keyed Discard Engine")
				if got := g.Players[game.Player1].Life; got != life-2 {
					t.Fatalf("life = %d, want %d after paying 2 life", got, life-2)
				}
				if _, ok := permanentByObjectID(g, sacrificed.ObjectID); ok ||
					!g.Players[game.Player1].Exile.Contains(sacrificed.CardInstanceID) {
					t.Fatal("sacrificed creature was not sacrificed into exile under Rest in Peace")
				}
			},
		},
	}
}

// TestDiscardCostDecidesReplacementsBeforeMovingAnyCard proves a "discard two
// cards" cost is one simultaneous discard on every payment route: Player 1's
// replacement choice for each discarded Progenitus (its own replacement or
// Player 2's Rest in Peace) is made while both Progenitus are still in hand and
// unseen elsewhere, then both move in one cohort to the chosen zone. Keyed
// routes capture both Progenitus as paid-cost subjects from their hand
// incarnation, alongside the other costs paid with them.
func TestDiscardCostDecidesReplacementsBeforeMovingAnyCard(t *testing.T) {
	for _, rt := range discardCostRoutes() {
		for _, pick := range []struct {
			label string
			want  zone.Type
		}{
			{label: "Progenitus", want: zone.Library},
			{label: "exile it instead", want: zone.Exile},
		} {
			t.Run(rt.name+"/"+pick.label, func(t *testing.T) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				g.Turn.ActivePlayer = game.Player1
				engine := NewEngine(nil)
				set := hiddenReplacementRegistry(t)
				resolveHiddenReplacementPermanent(t, g, engine, game.Player2, set.restInPeace)
				resolveWholeStack(t, engine, g)
				setSorcerySpeedTurn(g, game.Player1)
				act := rt.setup(t, g)
				progenitus := []id.ID{addCardToHand(g, game.Player1, set.progenitus), addCardToHand(g, game.Player1, set.progenitus)}
				versions := make([]uint64, len(progenitus))
				for i, cardID := range progenitus {
					card, _ := g.GetCardInstance(cardID)
					card.ZoneVersion = uint64(7 + i)
					versions[i] = card.ZoneVersion
				}
				kept := addCardToHand(g, game.Player1, discardableCard("Kept card"))
				pending := map[id.ID]game.PlayerID{progenitus[0]: game.Player1, progenitus[1]: game.Player1}
				var asked []game.PlayerID
				recorder := replacementChoiceRecorder{t: t, g: g, pending: pending, pick: pick.label, asked: &asked}
				life := g.Players[game.Player1].Life
				start := len(g.Events)

				turnLog := &TurnLog{}
				if !engine.applyActionWithChoices(g, game.Player1, act, progenitusDiscardAgents(recorder), turnLog) {
					t.Fatalf("paying the discard-two cost failed: %+v", turnLog.Entries)
				}

				if want := []game.PlayerID{game.Player1, game.Player1}; !slices.Equal(asked, want) {
					t.Fatalf("replacement choosers = %v, want %v (the discarding owner, once per Progenitus)", asked, want)
				}
				assertDiscardCostCohort(t, g, start, progenitus)
				zones := map[zone.Type]*zone.Zone{zone.Library: &g.Players[game.Player1].Library, zone.Exile: &g.Players[game.Player1].Exile}
				for _, cardID := range progenitus {
					var got []zone.Type
					for _, event := range g.Events[start:] {
						if event.Kind == game.EventZoneChanged && event.FromZone == zone.Hand && event.CardID == cardID {
							got = append(got, event.ToZone)
						}
					}
					if !slices.Equal(got, []zone.Type{pick.want}) || !zones[pick.want].Contains(cardID) {
						t.Fatalf("Progenitus %d moved from hand to %v, want only to %v", cardID, got, pick.want)
					}
				}
				if !g.Players[game.Player1].Hand.Contains(kept) {
					t.Fatal("unchosen card left the hand")
				}
				obj, _ := g.Stack.Peek()
				subjects := obj.PaidCostSubjects
				if !rt.keyed {
					if len(subjects) != 0 {
						t.Fatalf("unkeyed cost published paid subjects %+v", subjects)
					}
				} else {
					if len(subjects) != len(progenitus) {
						t.Fatalf("paid subjects = %+v, want one per discarded Progenitus", subjects)
					}
					for _, subject := range subjects {
						i := slices.Index(progenitus, subject.Snapshot.CardID)
						if i < 0 || subject.Key != "discarded" || subject.Kind != game.PaidCostDiscard ||
							subject.Snapshot.Name != "Progenitus" || subject.Snapshot.Owner != game.Player1 ||
							subject.Snapshot.FromZone != zone.Hand || !subject.CharacteristicsKnown ||
							!subject.CardZoneVersion.Exists || subject.CardZoneVersion.Val != versions[i] {
							t.Fatalf("paid subject %+v, want a discarded Progenitus captured in hand at its pre-discard zone version", subject)
						}
						progenitus[i] = 0
					}
				}
				if rt.check != nil {
					rt.check(t, g, life)
				}
			})
		}
	}
}

// TestUnpayableDiscardCostMovesNoCard proves an activation whose "discard two
// cards" cost can't be paid fully (CR 118.3) discards nothing: with one card in
// hand, Zombie Infestation's ability isn't activated, no replacement choice is
// asked and the card stays in hand.
func TestUnpayableDiscardCostMovesNoCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	engine := NewEngine(nil)
	set := hiddenReplacementRegistry(t)
	resolveHiddenReplacementPermanent(t, g, engine, game.Player2, set.restInPeace)
	resolveWholeStack(t, engine, g)
	setSorcerySpeedTurn(g, game.Player1)
	source := addCombatPermanent(g, game.Player1, registryCardDef(t, "Zombie Infestation"))
	progenitus := addCardToHand(g, game.Player1, set.progenitus)
	var asked []game.PlayerID
	recorder := replacementChoiceRecorder{t: t, g: g, pending: map[id.ID]game.PlayerID{progenitus: game.Player1}, pick: "Progenitus", asked: &asked}
	start := len(g.Events)

	if engine.applyActionWithChoices(g, game.Player1, action.ActivateAbility(source.ObjectID, 0, nil, 0), progenitusDiscardAgents(recorder), &TurnLog{}) {
		t.Fatal("activated Zombie Infestation with one card in hand")
	}
	if len(asked) != 0 || len(cohortDiscardEvents(g, start)) != 0 || !g.Stack.IsEmpty() ||
		!g.Players[game.Player1].Hand.Contains(progenitus) {
		t.Fatalf("unpayable cost asked %v, discarded %d cards or left a stack object", asked, len(cohortDiscardEvents(g, start)))
	}
}

// assertDiscardCostCohort requires the cost's discards to be exactly the given
// cards, each player's first discard this turn, sharing one nonzero cohort with
// their zone changes out of the hand. A spell's own move from hand to the stack
// (CR 601.2a) is not part of the discard.
func assertDiscardCostCohort(t *testing.T, g *game.Game, start int, cards []id.ID) {
	t.Helper()
	discards := cohortDiscardEvents(g, start)
	if len(discards) != len(cards) {
		t.Fatalf("discard events = %d, want %d", len(discards), len(cards))
	}
	cohort := discards[0].SimultaneousID
	if cohort == 0 {
		t.Fatal("discard cost has no simultaneous cohort")
	}
	for _, event := range discards {
		if event.Player != game.Player1 || !slices.Contains(cards, event.CardID) ||
			event.SimultaneousID != cohort || event.PlayerEventOrdinalThisTurn != 1 {
			t.Fatalf("discard %+v, want Player1's first discard of a Progenitus in cohort %d", event, cohort)
		}
	}
	for _, event := range g.Events[start:] {
		if event.Kind == game.EventZoneChanged && event.FromZone == zone.Hand &&
			slices.Contains(cards, event.CardID) && event.SimultaneousID != cohort {
			t.Fatalf("zone change of card %d in batch %d, want shared cohort %d", event.CardID, event.SimultaneousID, cohort)
		}
	}
}

// Source review (separate discard costs): pinned corpus 6da9ced0…a1b4 Six
// dbcbdf37-c40f-4068-b4a7-a849cab1056c "During your turn, nonland permanent
// cards in your graveyard have retrace."; Lesser Masticore
// a6a257bd-ee19-4246-b997-5288e6fa41e1 {2} "As an additional cost to cast this
// spell, discard a card."; Rielle, the Everwise
// 6f981e70-5828-49ac-8a72-cf1cc09be991 "Whenever you discard one or more cards
// for the first time each turn, draw that many cards." CR 702.81a retrace's
// land discard is an additional cost paid under 601.2f–h, separate from the
// spell's own discard cost; 601.2h pays each cost in turn, so only the cards of
// one cost component are discarded simultaneously.

// twoDiscardCosts is "discard two cards" followed by a separate "discard a card".
func twoDiscardCosts() []cost.Additional {
	return []cost.Additional{
		{Kind: cost.AdditionalDiscard, Text: "discard two cards", Amount: 2, Source: zone.Hand},
		{Kind: cost.AdditionalDiscard, Text: "discard a card", Amount: 1, Source: zone.Hand},
	}
}

// TestSeparateDiscardCostsAreSeparateDiscards proves each discard cost
// component is its own simultaneous discard on every payment route: the cards
// of one component share a cohort, a later component is a later discard
// (ordinal 2), and Rielle, the Everwise draws only for the first component.
func TestSeparateDiscardCostsAreSeparateDiscards(t *testing.T) {
	for _, tc := range []struct {
		name      string
		setup     func(*testing.T, *game.Game) action.Action
		wantSizes []int
	}{
		{
			name: "Lesser Masticore cast with retrace granted by Six",
			setup: func(t *testing.T, g *game.Game) action.Action {
				addPermanentForSBA(g, game.Player1, registryCardDef(t, "Six"))
				card := addCardToGraveyard(g, game.Player1, registryCardDef(t, "Lesser Masticore"))
				addCardToHand(g, game.Player1, sixGraveLand())
				addCardToHand(g, game.Player1, discardableCard("Nonland fuel"))
				g.Players[game.Player1].ManaPool.Add(mana.C, 2)
				return action.CastSpellFromZone(card, zone.Graveyard, nil, 0, nil)
			},
			wantSizes: []int{1, 1},
		},
		{
			name: "spell with two discard costs cast from hand",
			setup: func(_ *testing.T, g *game.Game) action.Action {
				spell := &game.CardDef{CardFace: game.CardFace{
					Name:            "Two Discard Costs Spell",
					ManaCost:        opt.Val(cost.Mana{cost.R}),
					Types:           []types.Card{types.Sorcery},
					AdditionalCosts: twoDiscardCosts(),
					SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{{
						Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()},
					}}}.Ability()),
				}}
				spellID := addCardToHand(g, game.Player1, spell)
				for i := range 3 {
					addCardToHand(g, game.Player1, discardableCard(fmt.Sprintf("Fuel %d", i)))
				}
				g.Players[game.Player1].ManaPool.Add(mana.R, 1)
				return action.CastSpell(spellID, nil, 0, nil)
			},
			wantSizes: []int{2, 1},
		},
		{
			name: "ability with two discard costs activated on the battlefield",
			setup: func(_ *testing.T, g *game.Game) action.Action {
				source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
					Name:  "Two Discard Costs Engine",
					Types: []types.Card{types.Artifact},
					ActivatedAbilities: []game.ActivatedAbility{{
						Text:            "Discard two cards, Discard a card: You gain 1 life.",
						AdditionalCosts: twoDiscardCosts(),
						Content: game.Mode{Sequence: []game.Instruction{{
							Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()},
						}}}.Ability(),
					}},
				}})
				for i := range 3 {
					addCardToHand(g, game.Player1, discardableCard(fmt.Sprintf("Fuel %d", i)))
				}
				return action.ActivateAbility(source.ObjectID, 0, nil, 0)
			},
			wantSizes: []int{2, 1},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			addCombatPermanent(g, game.Player1, registryCardDef(t, "Rielle, the Everwise"))
			for range 4 {
				addCardToLibrary(g, game.Player1, discardableCard("Draw fuel"))
			}
			act := tc.setup(t, g)
			setMainPhasePriority(g, game.Player1)
			agents := [game.NumPlayers]PlayerAgent{}
			agents[game.Player1] = authoredChoicePolicy{choose: func(_ PlayerObservation, req game.ChoiceRequest) []int {
				return req.DefaultSelection
			}}
			start := len(g.Events)
			turnLog := &TurnLog{}
			if !engine.applyActionWithChoices(g, game.Player1, act, agents, turnLog) {
				t.Fatalf("paying both discard costs failed: %+v", turnLog.Entries)
			}

			var cohorts []id.ID
			sizes := map[id.ID]int{}
			for _, event := range cohortDiscardEvents(g, start) {
				if event.SimultaneousID == 0 || event.Player != game.Player1 {
					t.Fatalf("discard %+v has no cohort or wrong player", event)
				}
				if !slices.Contains(cohorts, event.SimultaneousID) {
					cohorts = append(cohorts, event.SimultaneousID)
				}
				if want := slices.Index(cohorts, event.SimultaneousID) + 1; event.PlayerEventOrdinalThisTurn != want {
					t.Fatalf("discard %+v ordinal = %d, want %d", event, event.PlayerEventOrdinalThisTurn, want)
				}
				sizes[event.SimultaneousID]++
			}
			var gotSizes []int
			for _, cohort := range cohorts {
				gotSizes = append(gotSizes, sizes[cohort])
			}
			if !slices.Equal(gotSizes, tc.wantSizes) {
				t.Fatalf("discard cohort sizes = %v, want %v (one discard per cost component)", gotSizes, tc.wantSizes)
			}

			hand := g.Players[game.Player1].Hand.Size()
			if !engine.putTriggeredAbilitiesOnStack(g) {
				t.Fatal("Rielle did not trigger")
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			if engine.putTriggeredAbilitiesOnStack(g) {
				t.Fatal("Rielle triggered again for the second discard cost")
			}
			if got := g.Players[game.Player1].Hand.Size() - hand; got != tc.wantSizes[0] {
				t.Fatalf("Rielle drew %d cards, want %d for the first discard cost only", got, tc.wantSizes[0])
			}
		})
	}
}
