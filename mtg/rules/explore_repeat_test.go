package rules

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/internal/carddata"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func compileExploreRepeatDef(t *testing.T, text, typeLine string) *game.CardDef {
	t.Helper()
	card := &carddata.ScryfallCard{Name: "Test Scout", Layout: "normal",
		TypeLine: typeLine, ManaCost: "{X}{G}", OracleText: text}
	if typeLine == "Creature" {
		card.Power, card.Toughness = new("1"), new("1")
	}
	defs := legacyCardDefs(t, *card)

	return defs[0]
}

func exploreRevealIDs(g *game.Game) []id.ID {
	var ids []id.ID
	for _, event := range g.Events {
		if event.Kind == game.EventCardRevealed {
			ids = append(ids, event.CardID)
		}
	}
	return ids
}

func TestCompiledExploreRepeatMutatesOneSubjectSequentially(t *testing.T) {
	t.Parallel()
	for _, subject := range []string{"this creature", "it", "target creature"} {
		for _, variable := range []bool{false, true} {
			for _, count := range []int{0, 1, 2, 3} {
				t.Run(fmt.Sprintf("%s/x=%v/n=%d", subject, variable, count), func(t *testing.T) {
					t.Parallel()
					countText := fmt.Sprintf("%d", count)
					if variable {
						countText = "X"
					}
					def := compileExploreRepeatDef(t, "When this creature enters, "+subject+" explores "+countText+" times.", "Creature")
					g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
					creature := addCombatPermanent(g, game.Player2, def)
					creature.Owner = game.Player4
					g.CardInstances[creature.CardInstanceID].Owner = game.Player4
					last := addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Last land", Types: []types.Card{types.Land}}})
					nonland := addCardToLibrary(g, game.Player2, vanillaCreature("Nonland", 1, 1))
					first := addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "First land", Types: []types.Card{types.Land}}})
					agent := &choiceOnlyAgent{choices: [][]int{{1}}}
					obj := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackTriggeredAbility,
						Controller: game.Player1, SourceID: creature.ObjectID, SourceCardID: creature.CardInstanceID,
						XValue: count, HasTriggerEvent: true,
						TriggerEvent: game.Event{Kind: game.EventPermanentEnteredBattlefield, PermanentID: creature.ObjectID},
						Targets:      []game.Target{game.PermanentTarget(creature.ObjectID)}}
					NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.TriggeredAbilities[0].Content,
						[game.NumPlayers]PlayerAgent{game.Player2: agent}, &TurnLog{})
					if got := len(exploreRevealIDs(g)); got != count {
						t.Fatalf("reveals = %d, want %d", got, count)
					}
					wantCounters := 0
					if count >= 2 {
						wantCounters = 1
					}
					if got := creature.Counters.Get(counter.PlusOnePlusOne); got != wantCounters {
						t.Fatalf("counters = %d, want %d", got, wantCounters)
					}
					if g.Players[game.Player2].Hand.Contains(first) != (count >= 1) ||
						g.Players[game.Player2].Graveyard.Contains(nonland) != (count >= 2) ||
						g.Players[game.Player2].Hand.Contains(last) != (count >= 3) ||
						agent.next != wantCounters {
						t.Fatal("each iteration did not independently draw a land or counter and choose the nonland")
					}
					if g.Players[game.Player1].Hand.Size() != 0 || g.Players[game.Player4].Hand.Size() != 0 {
						t.Fatal("Explore used the stack controller or subject owner instead of its current controller")
					}
					if len(obj.LocalLinkedProducts) != 0 || len(g.LinkedObjects) != 0 {
						t.Fatal("repeat leaked local products")
					}
				})
			}
		}
	}
}

func TestCompiledExploreRepeatRetainedTopIsObservedAgain(t *testing.T) {
	t.Parallel()
	def := compileExploreRepeatDef(t, "When this creature enters, it explores two times.", "Creature")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatPermanent(g, game.Player1, def)
	addCardToLibrary(g, game.Player1, vanillaCreature("Third", 1, 1))
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Land", Types: []types.Card{types.Land}}})
	top := addCardToLibrary(g, game.Player1, vanillaCreature("Top", 1, 1))
	agent := &choiceOnlyAgent{choices: [][]int{{0}, {1}}}
	obj := &game.StackObject{Controller: game.Player1, SourceID: creature.ObjectID, HasTriggerEvent: true,
		TriggerEvent: game.Event{Kind: game.EventPermanentEnteredBattlefield, PermanentID: creature.ObjectID}}
	NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.TriggeredAbilities[0].Content,
		[game.NumPlayers]PlayerAgent{game.Player1: agent}, &TurnLog{})
	reveals := exploreRevealIDs(g)
	if len(reveals) != 2 || reveals[0] != top || reveals[1] != top ||
		agent.next != 2 || creature.Counters.Get(counter.PlusOnePlusOne) != 2 ||
		!g.Players[game.Player1].Graveyard.Contains(top) || g.Players[game.Player1].Hand.Size() != 0 {
		t.Fatal("repeated Explore did not observe and independently choose the retained top card twice")
	}
}

type exploreDepartAgent struct {
	choiceOnlyAgent

	depart func()
}

func (a *exploreDepartAgent) ChooseChoice(obs PlayerObservation, request game.ChoiceRequest) []int {
	if a.next == 0 {
		a.depart()
	}
	return a.choiceOnlyAgent.ChooseChoice(obs, request)
}

func TestCompiledExploreRepeatDoesNotRebindDepartedIncarnation(t *testing.T) {
	t.Parallel()
	for _, subject := range []string{"this creature", "it", "target creature"} {
		t.Run(subject, func(t *testing.T) {
			t.Parallel()
			def := compileExploreRepeatDef(t, "When this creature enters, "+subject+" explores two times.", "Creature")
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			old := addCombatPermanent(g, game.Player1, def)
			addCardToLibrary(g, game.Player1, vanillaCreature("Third", 1, 1))
			second := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Second", Types: []types.Card{types.Land}}})
			addCardToLibrary(g, game.Player1, vanillaCreature("First", 1, 1))
			var replacement *game.Permanent
			agent := &exploreDepartAgent{choiceOnlyAgent: choiceOnlyAgent{choices: [][]int{{1}}}, depart: func() {
				if _, ok := removePermanentFromBattlefield(g, old.ObjectID); !ok {
					t.Fatal("subject did not depart")
				}
				card, _ := g.GetCardInstance(old.CardInstanceID)
				var ok bool
				replacement, ok = createCardPermanent(g, card, game.Player1, zone.Exile)
				if !ok || replacement.ObjectID == old.ObjectID {
					t.Fatal("failed to create a new incarnation of the same card")
				}
			}}
			obj := &game.StackObject{Controller: game.Player1, SourceID: old.ObjectID, SourceCardID: old.CardInstanceID,
				HasTriggerEvent: true, TriggerEvent: game.Event{Kind: game.EventPermanentEnteredBattlefield, PermanentID: old.ObjectID},
				Targets: []game.Target{game.PermanentTarget(old.ObjectID)}}
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.TriggeredAbilities[0].Content,
				[game.NumPlayers]PlayerAgent{game.Player1: agent}, &TurnLog{})
			if len(exploreRevealIDs(g)) != 1 || agent.next != 1 || replacement == nil ||
				replacement.Counters.Get(counter.PlusOnePlusOne) != 0 ||
				g.Players[game.Player1].Hand.Contains(second) {
				t.Fatal("subsequent Explore rebound the departed subject to a new incarnation")
			}
		})
	}
}

func TestCompiledExploreRepeatEventSubjectSurvivesSourceDeparture(t *testing.T) {
	t.Parallel()
	def := compileExploreRepeatDef(t, "Whenever a creature you control enters, it explores two times.", "Enchantment")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, def)
	subject := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	for range 3 {
		addCardToLibrary(g, game.Player1, vanillaCreature("Explore card", 1, 1))
	}
	agent := &exploreDepartAgent{
		choiceOnlyAgent: choiceOnlyAgent{choices: [][]int{{1}, {1}}},
		depart: func() {
			if _, ok := removePermanentFromBattlefield(g, source.ObjectID); !ok {
				t.Fatal("source did not depart")
			}
		},
	}
	obj := &game.StackObject{Controller: game.Player1, SourceID: source.ObjectID, SourceCardID: source.CardInstanceID,
		HasTriggerEvent: true, TriggerEvent: game.Event{Kind: game.EventPermanentEnteredBattlefield, PermanentID: subject.ObjectID}}
	NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.TriggeredAbilities[0].Content,
		[game.NumPlayers]PlayerAgent{game.Player1: agent}, &TurnLog{})
	if subject.Counters.Get(counter.PlusOnePlusOne) != 2 || source.Counters.Get(counter.PlusOnePlusOne) != 0 ||
		len(exploreRevealIDs(g)) != 2 {
		t.Fatal("repeated event subject borrowed the source identity or stopped when the unrelated source left")
	}
}

func TestCompiledExploreRepeatEntryXAndCopy(t *testing.T) {
	t.Parallel()
	for _, cast := range []bool{false, true} {
		for _, count := range []int{0, 1, 2} {
			t.Run(fmt.Sprintf("cast=%v/x=%d", cast, count), func(t *testing.T) {
				t.Parallel()
				def := compileExploreRepeatDef(t, "When this creature enters, it explores X times.", "Creature")
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				addCardToLibrary(g, game.Player1, vanillaCreature("Third", 1, 1))
				addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Second", Types: []types.Card{types.Land}}})
				addCardToLibrary(g, game.Player1, vanillaCreature("First", 1, 1))
				engine := NewEngine(nil)
				cardID := addCardToHand(g, game.Player1, def)
				card, _ := g.GetCardInstance(cardID)
				g.Players[game.Player1].Hand.Remove(cardID)
				creature, ok := createCardPermanentFaceWithOptions(engine, g, card, game.Player1, zone.Stack,
					game.FaceFront, nil, permanentCreationOptions{WasCast: cast, XValue: count},
					[game.NumPlayers]PlayerAgent{}, nil)
				if !ok || !engine.putTriggeredAbilitiesOnStack(g) {
					t.Fatal("entry trigger was not created")
				}
				original, _ := g.Stack.Peek()
				want := count
				if !cast {
					want = 0
				}
				if original.XValue != want || original.TriggerEvent.PermanentID != creature.ObjectID {
					t.Fatal("entry X or actual event subject was not captured")
				}
				addEffectSpellToStack(g, game.Player1, game.CopyStackObject{Object: game.TargetStackObjectReference(0)},
					[]game.Target{game.StackObjectTarget(original.ID)})
				engine.resolveTopOfStack(g, &TurnLog{})
				copied, _ := g.Stack.Peek()
				if !copied.Copy || copied.XValue != want || copied.SourceID != original.SourceID ||
					copied.TriggerEvent.PermanentID != original.TriggerEvent.PermanentID {
					t.Fatal("repeat copy lost its chosen X, original source, or event identity")
				}
				agent := &choiceOnlyAgent{choices: [][]int{{1}, {1}}}
				engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player1: agent}, &TurnLog{})
				if got := len(exploreRevealIDs(g)); got != want {
					t.Fatalf("copy reveals=%d, want %d", got, want)
				}
				if g.Stack.Size() != 1 {
					t.Fatal("resolving the copy consumed the original trigger")
				}
			})
		}
	}
}

func TestJadelightRepeatNormalCastEntryXAndIncarnationCopies(t *testing.T) {
	t.Parallel()
	defs := legacyCardDefs(t, carddata.ScryfallCard{
		Name: "Jadelight Spelunker", Layout: "normal", ManaCost: "{X}{G}",
		TypeLine: "Creature — Merfolk Scout", Power: new("1"), Toughness: new("1"),
		OracleText: "When Jadelight Spelunker enters, it explores X times.",
	})

	for _, x := range []int{0, 1, 2} {
		for _, leave := range []bool{false, true} {
			t.Run(fmt.Sprintf("x=%d/depart-reenter=%v", x, leave), func(t *testing.T) {
				t.Parallel()
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				engine := NewEngine(nil)
				addCardToLibrary(g, game.Player1, vanillaCreature("Third", 1, 1))
				addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
					Name: "Second land", Types: []types.Card{types.Land},
				}})
				topCard := addCardToLibrary(g, game.Player1, vanillaCreature("Retained top", 1, 1))
				cardID := addCardToHand(g, game.Player1, defs[0])
				for range x + 1 {
					addBasicLandPermanent(g, game.Player1, types.Forest)
				}
				g.Turn.Phase, g.Turn.Step = game.PhasePrecombatMain, game.StepNone
				if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, x, nil)) {
					t.Fatalf("normal cast X=%d failed", x)
				}
				spell, ok := g.Stack.Peek()
				if !ok || spell.Kind != game.StackSpell || spell.XValue != x {
					t.Fatal("announced X was not captured on the paid spell")
				}
				engine.resolveTopOfStack(g, &TurnLog{})
				old := permanentForCard(g, cardID)
				if old == nil || !old.EnteredFromCast || old.CastXValue != x {
					t.Fatal("normal spell resolution lost cast X on the entered permanent")
				}
				if !engine.putTriggeredAbilitiesOnStack(g) {
					t.Fatal("normal entry did not create the Explore trigger")
				}
				original, _ := g.Stack.Peek()
				if original.ID == spell.ID || original.XValue != x ||
					!original.HasTriggerEvent || !original.TriggerEvent.EnterWasCast ||
					original.TriggerEvent.EnterXValue != x || original.TriggerEvent.PermanentID != old.ObjectID {
					t.Fatal("the new entry-trigger stack object did not capture the paid spell's own X and event incarnation")
				}
				addEffectSpellToStack(g, game.Player1, game.CopyStackObject{Object: game.TargetStackObjectReference(0)},
					[]game.Target{game.StackObjectTarget(original.ID)})
				engine.resolveTopOfStack(g, &TurnLog{})
				copied, _ := g.Stack.Peek()
				if !copied.Copy || copied.XValue != x || copied.SourceID != old.ObjectID ||
					copied.TriggerEvent.PermanentID != old.ObjectID || copied.TriggerEvent.EnterXValue != x {
					t.Fatal("the copied entry trigger lost its captured X or old incarnation")
				}
				var returned *game.Permanent
				if leave {
					if _, ok := removePermanentFromBattlefield(g, old.ObjectID); !ok {
						t.Fatal("old source did not depart")
					}
					card, _ := g.GetCardInstance(cardID)
					returned, ok = createCardPermanent(g, card, game.Player1, zone.Exile)
					if !ok || returned.ObjectID == old.ObjectID || !engine.putTriggeredAbilitiesOnStack(g) {
						t.Fatal("same card's noncast new incarnation or its entry trigger was not created")
					}
					fresh, _ := g.Stack.Peek()
					if fresh.XValue != 0 || fresh.TriggerEvent.EnterWasCast ||
						fresh.TriggerEvent.PermanentID != returned.ObjectID {
						t.Fatal("new noncast entry borrowed the earlier announced X")
					}
					engine.resolveTopOfStack(g, &TurnLog{})
				}
				agent := &choiceOnlyAgent{choices: [][]int{{0}, {0}, {0}, {0}}}
				agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}
				engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
				if original.XValue != x || copied.XValue != x || g.Stack.Size() != 1 {
					t.Fatal("resolving the copy changed the original announced X or consumed the original")
				}
				engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
				want := 2 * x
				if leave {
					want = 0
				}
				if len(exploreRevealIDs(g)) != want || agent.next != want ||
					old.Counters.Get(counter.PlusOnePlusOne) != want {
					t.Fatalf("normal-cast original/copy repeated actions=%d choices=%d counters=%d, want %d",
						len(exploreRevealIDs(g)), agent.next, old.Counters.Get(counter.PlusOnePlusOne), want)
				}
				if returned != nil && returned.Counters.Get(counter.PlusOnePlusOne) != 0 {
					t.Fatal("old entry trigger rebound to the same card's new incarnation")
				}
				if top, _ := g.Players[game.Player1].Library.Top(); top != topCard {
					t.Fatal("zero, departed, or retained-top Explore moved the wrong library card")
				}
			})
		}
	}
}

func TestCompiledExploreRepeatTargetFizzleAndMixedSlot(t *testing.T) {
	t.Parallel()
	for _, departed := range []bool{false, true} {
		t.Run(fmt.Sprintf("departed=%v", departed), func(t *testing.T) {
			t.Parallel()
			def := compileExploreRepeatDef(t, "Target creature explores two times. Draw a card.", "Sorcery")
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			target := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
			addCardToLibrary(g, game.Player2, vanillaCreature("Third", 1, 1))
			addCardToLibrary(g, game.Player2, vanillaCreature("Second", 1, 1))
			addCardToLibrary(g, game.Player2, vanillaCreature("First", 1, 1))
			addCardToLibrary(g, game.Player1, vanillaCreature("Controller draw", 1, 1))
			source := addCardToHand(g, game.Player1, def)
			g.Players[game.Player1].Hand.Remove(source)
			obj := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: source, SourceCardID: source,
				Controller: game.Player1, Targets: []game.Target{game.PermanentTarget(target.ObjectID)}}
			g.Stack.Push(obj)
			if departed {
				removePermanentFromBattlefield(g, target.ObjectID)
			}
			agent := &choiceOnlyAgent{choices: [][]int{{1}, {1}}}
			NewEngine(nil).resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player2: agent}, &TurnLog{})
			if departed {
				if len(exploreRevealIDs(g)) != 0 || g.Players[game.Player1].Hand.Size() != 0 {
					t.Fatal("fully illegal targets did not fizzle the entire repeat and rider")
				}
			} else if len(exploreRevealIDs(g)) != 2 || target.Counters.Get(counter.PlusOnePlusOne) != 2 ||
				g.Players[game.Player1].Hand.Size() != 1 {
				t.Fatal("targeted repeat or independent controller rider did not resolve")
			}
		})
	}
	def := compileExploreRepeatDef(t, "Target creature gets +1/+1 until end of turn. Target creature explores two times.", "Sorcery")
	repeat, ok := def.SpellAbility.Val.Modes[0].Sequence[1].Primitive.(game.RepeatProcess)
	if !ok {
		t.Fatal("mixed-slot action is not a RepeatProcess")
	}
	explore, ok := repeat.Body.Modes[0].Sequence[0].Primitive.(game.Explore)
	if !ok || explore.Creature != game.TargetPermanentReference(1) {
		t.Fatal("nested repeat target was not remapped to the second global slot")
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	first := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	second := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	addCardToLibrary(g, game.Player2, vanillaCreature("Third", 1, 1))
	addCardToLibrary(g, game.Player2, vanillaCreature("Second", 1, 1))
	addCardToLibrary(g, game.Player2, vanillaCreature("First", 1, 1))
	obj := &game.StackObject{Controller: game.Player1,
		Targets: []game.Target{game.PermanentTarget(first.ObjectID), game.PermanentTarget(second.ObjectID)}}
	NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.SpellAbility.Val,
		[game.NumPlayers]PlayerAgent{game.Player2: &choiceOnlyAgent{choices: [][]int{{1}, {1}}}}, &TurnLog{})
	if first.Counters.Get(counter.PlusOnePlusOne) != 0 || second.Counters.Get(counter.PlusOnePlusOne) != 2 {
		t.Fatal("repeated Explore used the wrong accumulated target slot")
	}
}

func TestCompiledExploreRepeatPartialTargetLegality(t *testing.T) {
	t.Parallel()
	for _, legal := range []bool{false, true} {
		t.Run(fmt.Sprintf("legal=%v", legal), func(t *testing.T) {
			t.Parallel()
			def := compileExploreRepeatDef(t, "Target player draws a card. Target creature you control explores X times.", "Sorcery")
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
			for range 3 {
				addCardToLibrary(g, game.Player1, vanillaCreature("Explore card", 1, 1))
				addCardToLibrary(g, game.Player2, vanillaCreature("Illegal target controller card", 1, 1))
			}
			addCardToLibrary(g, game.Player3, vanillaCreature("Independent draw", 1, 1))
			source := addCardToHand(g, game.Player1, def)
			g.Players[game.Player1].Hand.Remove(source)
			obj := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: source, SourceCardID: source,
				Controller: game.Player1, XValue: 2,
				Targets: []game.Target{game.PlayerTarget(game.Player3), game.PermanentTarget(creature.ObjectID)}}
			g.Stack.Push(obj)
			if !legal {
				creature.Controller = game.Player2
			}
			agent := &choiceOnlyAgent{choices: [][]int{{1}, {1}}}
			NewEngine(nil).resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player1: agent}, &TurnLog{})
			want := 0
			if legal {
				want = 2
			}
			if len(exploreRevealIDs(g)) != want || creature.Counters.Get(counter.PlusOnePlusOne) != want ||
				g.Players[game.Player3].Hand.Size() != 1 {
				t.Fatal("a partially illegal repeated target rebound or suppressed an independent legal target")
			}
		})
	}
}

func TestExploreRepeatLocalProductFrames(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	addCardToLibrary(g, game.Player1, vanillaCreature("Observed", 1, 1))
	const key game.LinkedKey = "iteration-observation"
	body := game.Mode{Sequence: []game.Instruction{
		{Primitive: game.Explore{Creature: game.SourcePermanentReference()}},
		{Primitive: game.LookAtLibraryTop{Player: game.ControllerReference(), PublishLinked: key},
			LocalProducts: game.LocalProducts{Links: []game.LinkedKey{key}}},
		{Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(1)},
			Condition: opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{
				Object:        opt.Val(game.LinkedObjectReference(string(key))),
				ObjectMatches: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
			})})},
	}}.Ability()
	obj := &game.StackObject{Controller: game.Player1, SourceID: creature.ObjectID}
	resolveInstruction(NewEngine(nil), g, obj, game.RepeatProcess{Times: game.Fixed(2), Body: body}, &TurnLog{})
	if g.Players[game.Player1].Life != 42 || creature.Counters.Get(counter.PlusOnePlusOne) != 2 ||
		len(obj.LocalLinkedProducts) != 0 || len(g.LinkedObjects) != 0 {
		t.Fatal("repeat did not refresh and clean up each iteration's local observation")
	}
}

func TestCompiledGenericRepeatTargetAndCurrentQuantity(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"Repeat the following process two times. Target player draws a card.",
		"Target creature explores X times, where X is the number of artifacts you control.",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			def := compileExploreRepeatDef(t, text, "Sorcery")
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			creature := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
			addCardToLibrary(g, game.Player2, vanillaCreature("Third", 1, 1))
			for range 2 {
				addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
					Name: "Artifact", Types: []types.Card{types.Artifact},
				}})
				addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
					Name: "Land", Types: []types.Card{types.Land},
				}})
			}
			target := game.PermanentTarget(creature.ObjectID)
			if text == "Repeat the following process two times. Target player draws a card." {
				target = game.PlayerTarget(game.Player2)
			}
			obj := &game.StackObject{Controller: game.Player1, Targets: []game.Target{target}}
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.SpellAbility.Val, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			if g.Players[game.Player2].Hand.Size() != 2 || g.Players[game.Player1].Hand.Size() != 0 {
				t.Fatal("shared repeat did not use its actual fixed/current modeled quantity and original target")
			}
		})
	}
}
