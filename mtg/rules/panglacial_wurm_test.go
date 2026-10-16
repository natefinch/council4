package rules

import (
	"reflect"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestPanglacialWurmCastsBeforeFinding(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	wurm := addCardToLibrary(g, game.Player1, authored.PanglacialWurm.New())
	other := addCardToLibrary(g, game.Player1, vanillaCreature("Find this card", 2, 2))
	g.Players[game.Player1].ManaPool.Add(mana.G, 7)
	var offered bool
	policies := simPassPolicies()
	policies[game.Player1] = &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
		if request.Prompt == "Cast a card before finding cards?" {
			offered = true
			for _, option := range request.Options {
				if option.Card.Exists && option.Card.Val.CardID == wurm {
					return []int{option.Index}
				}
			}
		}
		if request.Kind == game.ChoiceSearch {
			for _, option := range request.Options {
				if option.Card.Exists && option.Card.Val.CardID == wurm {
					t.Fatal("cast Wurm remained a search candidate")
				}
			}
		}
		return request.DefaultSelection
	}}
	found := engine.searchLibraryRevealOnlyCards(g, &game.StackObject{Controller: game.Player1}, policies, &TurnLog{},
		game.Player1, game.SearchSpec{RevealOnly: true}, 1)
	if !offered || !slices.Equal(found, []id.ID{other}) || g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatalf("search did not cast before finding: offered %v, found %v", offered, found)
	}
	top, present := g.Stack.Peek()
	if !present || top.SourceID != wurm || top.Controller != game.Player1 || top.SourceZone != zone.Library {
		t.Fatal("paid Wurm is not on the stack under its searching actor's control")
	}
}

func panglacialCastChoices(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Prompt == "Cast a card before finding cards?" {
		return []int{1}
	}
	return request.DefaultSelection
}

func TestPanglacialWurmMultipleCastsKeepOrderAndDeferTriggers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.PriorityPlayer = game.Player3
	wurms := []id.ID{
		addCardToLibrary(g, game.Player1, authored.PanglacialWurm.New()),
		addCardToLibrary(g, game.Player1, authored.PanglacialWurm.New()),
	}
	other := addCardToLibrary(g, game.Player1, vanillaCreature("Unrelated top card", 2, 2))
	witness := vanillaCreature("Cast witness", 1, 3)
	witness.TriggeredAbilities = []game.TriggeredAbility{{
		Trigger: game.TriggerCondition{Type: game.TriggerWhenever, Pattern: game.TriggerPattern{
			Event: game.EventSpellCast, Controller: game.TriggerControllerYou,
		}},
		Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.GainLife{
			Player: game.ControllerReference(), Amount: game.Fixed(1),
		}}}}.Ability(),
	}}
	addCombatPermanent(g, game.Player1, witness)
	g.Players[game.Player1].ManaPool.Add(mana.G, 14)
	policies := simPassPolicies()
	policies[game.Player1] = &controlScript{choice: func(obs PlayerObservation, request game.ChoiceRequest) []int {
		if request.Kind == game.ChoiceSearch && (g.Stack.Size() != 2 || g.Players[game.Player1].Life != 40) {
			t.Fatal("cast triggers or Wurms resolved before finding")
		}
		return panglacialCastChoices(obs, request)
	}}
	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{{Primitive: game.Search{
		Player: game.ControllerReference(), Amount: game.Fixed(1),
		Spec: game.SearchSpec{SourceZone: zone.Library, RevealOnly: true, Reveal: true},
	}}}, nil)
	log := &TurnLog{}
	engine.resolveTopOfStackWithChoices(g, policies, log)
	if !slices.Equal(g.Players[game.Player1].Library.All(), []id.ID{other}) || g.Stack.Size() != 2 ||
		g.Players[game.Player1].ManaPool.Total() != 0 || g.Turn.PriorityPlayer != game.Player3 {
		t.Fatal("multiple paid casts changed unrelated order, priority, or resources")
	}
	for _, object := range g.Stack.Objects() {
		if !slices.Contains(wurms, object.SourceID) {
			t.Fatal("search left a spurious stack object")
		}
	}
	if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 4 {
		t.Fatal("cast triggers did not wait for the next priority boundary")
	}
	for _, decision := range log.Choices {
		for _, option := range decision.Request.Options {
			if option.Card.Exists {
				t.Fatal("private search-cast identities escaped into replay choice logs")
			}
		}
	}
}

func TestPanglacialWurmDeclineAndFailedPaymentAreAtomic(t *testing.T) {
	for _, scenario := range []string{"decline", "insufficient mana", "declined mana abilities", "unpayable additional cost"} {
		t.Run(scenario, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			def := authored.PanglacialWurm.New()
			if scenario == "declined mana abilities" {
				addComplexManaAbilityPermanent(g, game.Player1,
					&game.CardDef{CardFace: game.CardFace{Name: "Unused pain source", Types: []types.Card{types.Artifact}}},
					&game.ManaAbility{
						AdditionalCosts: []cost.Additional{cost.T, {Kind: cost.AdditionalPayLife, Amount: 1}},
						Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.AddMana{
							Player: opt.Val(game.ControllerReference()), ManaColor: mana.G, Amount: game.Fixed(1),
						}}}}.Ability(),
					})
			}
			if scenario == "unpayable additional cost" {
				def.AdditionalCosts = []cost.Additional{{Kind: cost.AdditionalDiscard, Amount: 1}}
			}
			addCardToLibrary(g, game.Player1, vanillaCreature("Bottom card", 2, 2))
			wurm := addCardToLibrary(g, game.Player1, def)
			addCardToLibrary(g, game.Player1, vanillaCreature("Top card", 2, 2))
			card, _ := g.GetCardInstance(wurm)
			version := card.ZoneVersion
			before := g.Players[game.Player1].Library.All()
			policies := simPassPolicies()
			policies[game.Player1] = &controlScript{choice: func(obs PlayerObservation, request game.ChoiceRequest) []int {
				if request.Kind == game.ChoiceSearch {
					return nil
				}
				if scenario == "decline" {
					return request.DefaultSelection
				}
				return panglacialCastChoices(obs, request)
			}}
			engine.searchLibraryRevealOnlyCards(g, &game.StackObject{Controller: game.Player1}, policies, &TurnLog{},
				game.Player1, game.SearchSpec{RevealOnly: true, FailToFindPolicy: game.SearchMayFailToFind}, 1)
			if !slices.Equal(before, g.Players[game.Player1].Library.All()) || card.ZoneVersion != version ||
				g.Players[game.Player1].ManaPool.Total() != 0 || g.Stack.Size() != 0 || len(g.SearchControlScopes) != 0 ||
				len(g.ManaPaymentScopes) != 0 || g.Execution.InvalidChoices != 0 {
				t.Fatal("declined or failed cast spent resources, changed order/incarnation, or retained permission")
			}
			for _, event := range g.Events {
				if event.Kind == game.EventSpellCast || event.Kind == game.EventZoneChanged {
					t.Fatal("failed cast published a cast or zone transition")
				}
			}
		})
	}
}

func TestPanglacialWurmManaAbilityChangesLiveSearchCandidatesAndLocksCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.PriorityPlayer = game.Player3
	engine := NewEngine(nil)
	wurm := addCardToLibrary(g, game.Player1, authored.PanglacialWurm.New())
	remaining := addCardToLibrary(g, game.Player1, vanillaCreature("Still searchable", 2, 2))
	drawn := addCardToLibrary(g, game.Player1, vanillaCreature("Drawn during payment", 2, 2))
	source := addComplexManaAbilityPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Cost-reducing mana source", Types: []types.Card{types.Artifact},
		StaticAbilities: []game.StaticAbility{{RuleEffects: []game.RuleEffect{{
			Kind: game.RuleEffectCostModifier, AffectedPlayer: game.PlayerYou,
			CostModifier: game.CostModifier{GenericReduction: 5},
		}}}},
	}}, &game.ManaAbility{
		Text:            "Sacrifice this artifact: Add G and draw a card.",
		AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalSacrificeSource, Amount: 1}},
		Content: game.Mode{Sequence: []game.Instruction{
			{Primitive: game.AddMana{Player: opt.Val(game.ControllerReference()), ManaColor: mana.G, Amount: game.Fixed(1)}},
			{Primitive: game.Draw{Player: game.ControllerReference(), Amount: game.Fixed(1)}},
		}}.Ability(),
	})
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	var activated bool
	policies := simPassPolicies()
	policies[game.Player1] = &controlScript{choice: func(obs PlayerObservation, request game.ChoiceRequest) []int {
		if request.Prompt == "Activate a mana ability while paying?" {
			if g.Players[game.Player1].Library.Contains(wurm) || g.Stack.Size() != 1 ||
				g.Turn.PriorityPlayer != game.Player3 || !g.ManaPaymentAllowed(game.Player1) {
				t.Fatal("payment did not move the proposal first or granted ordinary priority")
			}
			clone := g.Clone()
			clone.ManaPaymentScopes[0] = game.Player2
			clone.SearchControlScopes[0].BeforeFind = false
			if !g.SearchCastingAllowed(game.Player1) || !g.ManaPaymentAllowed(game.Player1) {
				t.Fatal("payment/search metadata was shared with a clone")
			}
			activated = true
			return []int{1}
		}
		if request.Kind == game.ChoiceSearch {
			for _, option := range request.Options {
				if option.Card.Exists && option.Card.Val.CardID != remaining {
					t.Fatal("search candidates were cached before payment changed the library")
				}
			}
		}
		return panglacialCastChoices(obs, request)
	}}
	found := engine.searchLibraryRevealOnlyCards(g, &game.StackObject{Controller: game.Player1}, policies, &TurnLog{},
		game.Player1, game.SearchSpec{RevealOnly: true}, 1)
	if !activated || !slices.Equal(found, []id.ID{remaining}) || !g.Players[game.Player1].Hand.Contains(drawn) ||
		!g.Players[game.Player1].Graveyard.Contains(source.CardInstanceID) || g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("manual payment did not use its locked two-mana cost and resume the actual library")
	}
	if top, ok := g.Stack.Peek(); !ok || top.SourceID != wurm || top.ManaSpentToCast != 2 {
		t.Fatal("cost modifier was reevaluated after its mana-source sacrifice")
	}
}

func TestPanglacialWurmControlKeepsActorPayerAndPrivateViewer(t *testing.T) {
	for _, opposition := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		g.Turn.TurnNumber = 7
		if opposition {
			addCombatPermanent(g, game.Player2, oppositionAgentDef())
		} else {
			g.PlayerControlEffects = []game.PlayerControlEffect{{
				ID: g.IDGen.Next(), Player: game.Player1, Controller: game.Player2, Turn: 7,
			}}
		}
		wurm := addCardToLibrary(g, game.Player1, authored.PanglacialWurm.New())
		other := addCardToLibrary(g, game.Player1, vanillaCreature("Controlled find", 2, 2))
		g.Players[game.Player1].ManaPool.Add(mana.G, 7)
		g.Players[game.Player2].ManaPool.Add(mana.G, 9)
		policies := simPassPolicies()
		decisions := 0
		policies[game.Player2] = &controlScript{choice: func(obs PlayerObservation, request game.ChoiceRequest) []int {
			decisions++
			if obs.Player != game.Player1 || obs.DecisionMaker != game.Player2 || request.Player != game.Player1 ||
				len(observe(g, game.Player3).LibraryLookable(game.Player1)) != 0 {
				t.Fatal("search control changed the actor or disclosed its library to an uninvolved player")
			}
			return panglacialCastChoices(obs, request)
		}}
		NewEngine(nil).searchLibrary(g, &game.StackObject{Controller: game.Player1}, policies, &TurnLog{},
			game.Player1, game.Player1, game.SearchSpec{SourceZone: zone.Library, Destination: zone.Hand}, 1)
		top, ok := g.Stack.Peek()
		if decisions == 0 || !ok || top.SourceID != wurm || top.Controller != game.Player1 ||
			g.Players[game.Player1].ManaPool.Total() != 0 || g.Players[game.Player2].ManaPool.Total() != 9 {
			t.Fatal("search controller became the caster/payer")
		}
		if opposition {
			if !g.Players[game.Player1].Exile.Contains(other) {
				t.Fatal("Opposition Agent did not still redirect the actual find")
			}
			if _, permission := playFromExilePermissionFor(g, wurm); permission {
				t.Fatal("cast Wurm was mistaken for a found card")
			}
		} else if !g.Players[game.Player1].Hand.Contains(other) {
			t.Fatal("ordinary turn control lost the actual find")
		}
		if g.LookingAtLibrary(game.Player2, game.Player1) || g.SearchCastingAllowed(game.Player1) {
			t.Fatal("controlled search permission escaped its scope")
		}
	}
}

func TestPanglacialWurmPermissionIsNotTopLookForeignOrPriority(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	wurm := addCardToLibrary(g, game.Player1, authored.PanglacialWurm.New())
	card, _ := g.GetCardInstance(wurm)
	g.Players[game.Player1].ManaPool.Add(mana.G, 7)
	if containsAction(engine.legalActions(g, game.Player1), action.CastSpellFromZone(wurm, zone.Library, nil, 0, nil)) {
		t.Fatal("search permission became an ordinary library cast permission")
	}
	before := g.Players[game.Player1].Library.All()
	ctx := &programContext{active: true, engine: engine, game: g, obj: &game.StackObject{Controller: game.Player1}, log: &TurnLog{}}
	if _, err := ctx.Execute(game.LookAtLibraryTop{Player: game.ControllerReference(), PublishLinked: "look"}); err != nil {
		t.Fatal(err)
	}
	if searchCastCardAllowed(g, game.Player1, card) || !reflect.DeepEqual(before, g.Players[game.Player1].Library.All()) {
		t.Fatal("looking at the top opened a search cast window")
	}
	end := beginSearchControl(g, searchControl{searcher: game.Player2, decisionMaker: game.Player2})
	g.SearchControlScopes[0].BeforeFind = true
	if searchCastCardAllowed(g, game.Player2, card) || searchCastCardAllowed(g, game.Player1, card) {
		t.Fatal("another library's search granted this card's permission")
	}
	end()
	if searchCastCardAllowed(g, game.Player1, card) {
		t.Fatal("closed search retained permission")
	}
}

func TestPanglacialWurmAuthoredPrivateTutorAndSimulator(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	wurm := addCardToLibrary(g, game.Player1, authored.PanglacialWurm.New())
	find := addCardToLibrary(g, game.Player1, vanillaCreature("Private find", 2, 2))
	g.Players[game.Player1].ManaPool.Add(mana.G, 7)
	addImplementationSpellToStack(g, game.Player1, authored.DemonicTutor.New(), nil)
	engine := removalTokenEngine(t, authored.DemonicTutor)
	policies := simPassPolicies()
	policies[game.Player1] = &controlScript{choice: panglacialCastChoices}
	before := g.Players[game.Player1].Library.All()
	branch := engine.Simulator().ResolvePriority(g, policies)
	if !slices.Equal(g.Players[game.Player1].Library.All(), before) || g.Stack.Size() != 1 ||
		g.Players[game.Player1].ManaPool.Total() != 7 {
		t.Fatal("simulating a search cast mutated its source")
	}
	engine.resolveTopOfStackWithChoices(g, policies, &TurnLog{})
	if !g.Players[game.Player1].Hand.Contains(find) || g.Players[game.Player1].Library.Contains(wurm) {
		t.Fatal("actual authored private tutor did not resume after the cast")
	}
	// ResolvePriority also resolves the new Wurm after the search spell finishes.
	if permanentForCard(branch, wurm) == nil || !branch.Players[game.Player1].Hand.Contains(find) ||
		len(branch.SearchControlScopes) != 0 || len(branch.ManaPaymentScopes) != 0 {
		t.Fatal("Simulator did not finish the complete paid search and later Wurm resolution")
	}
}

func TestSearchCastingProcessesTargetsAndAdditionalCosts(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := authored.PanglacialWurm.New()
	def.SpellAbility = opt.Val(game.Mode{
		Targets:  []game.TargetSpec{{Allow: game.TargetAllowPlayer, MinTargets: 1, MaxTargets: 1}},
		Sequence: []game.Instruction{{Primitive: game.GainLife{Player: game.TargetPlayerReference(0), Amount: game.Fixed(1)}}},
	}.Ability())
	def.AdditionalCosts = []cost.Additional{{Kind: cost.AdditionalDiscard, Amount: 1, Text: "Discard a card"}}
	wurm := addCardToLibrary(g, game.Player1, def)
	discard := addCardToHand(g, game.Player1, vanillaCreature("Additional payment", 1, 1))
	g.Players[game.Player1].ManaPool.Add(mana.G, 7)
	policies := simPassPolicies()
	policies[game.Player1] = &controlScript{choice: panglacialCastChoices}
	NewEngine(nil).searchLibraryRevealOnlyCards(g, &game.StackObject{Controller: game.Player1}, policies, &TurnLog{},
		game.Player1, game.SearchSpec{RevealOnly: true}, 1)
	top, ok := g.Stack.Peek()
	if !ok || top.SourceID != wurm || len(top.Targets) != 1 || top.Targets[0].Kind != game.TargetPlayer ||
		!g.Players[game.Player1].Graveyard.Contains(discard) || !slices.Contains(top.AdditionalCostsPaid, "Discard a card") ||
		g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("search permission bypassed ordinary target announcement or additional payment")
	}
}

func TestPanglacialWurmOptionalSearchDeclineAndProhibitions(t *testing.T) {
	for _, scenario := range []string{"no search", "cast prohibition", "source-zone prohibition"} {
		t.Run(scenario, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			wurm := addCardToLibrary(g, game.Player1, authored.PanglacialWurm.New())
			g.Players[game.Player1].ManaPool.Add(mana.G, 7)
			switch scenario {
			case "cast prohibition":
				g.RuleEffects = []game.RuleEffect{{Kind: game.RuleEffectCantCastSpells, Controller: game.Player1, AffectedPlayer: game.PlayerYou}}
			case "source-zone prohibition":
				g.RuleEffects = []game.RuleEffect{{
					Kind: game.RuleEffectCantCastFromZones, Controller: game.Player1, AffectedPlayer: game.PlayerYou,
					CantCastFromZones: []zone.Type{zone.Library},
				}}
			case "no search":
			default:
				t.Fatal("unknown search prohibition scenario")
			}

			policies := simPassPolicies()
			policies[game.Player1] = &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
				if request.Prompt == "Cast a card before finding cards?" {
					t.Fatal("permission bypassed a prohibition or an unchosen search")
				}
				if scenario == "no search" && request.Kind == game.ChoiceMay {
					return []int{0}
				}
				return request.DefaultSelection
			}}
			addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{{
				Optional: scenario == "no search",
				Primitive: game.Search{Player: game.ControllerReference(), Amount: game.Fixed(1),
					Spec: game.SearchSpec{SourceZone: zone.Library, RevealOnly: true, Reveal: true, FailToFindPolicy: game.SearchMayFailToFind}},
			}}, nil)
			NewEngine(nil).resolveTopOfStackWithChoices(g, policies, &TurnLog{})
			if !g.Players[game.Player1].Library.Contains(wurm) || g.Players[game.Player1].ManaPool.Total() != 7 || g.Stack.Size() != 0 {
				t.Fatal("declined or prohibited search casting changed a card or payment")
			}
		})
	}
}

func TestPanglacialWurmUnpaidProposalReversesReversibleManaAbilities(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	wurm := addCardToLibrary(g, game.Player1, authored.PanglacialWurm.New())
	source := addComplexManaAbilityPermanent(g, game.Player1,
		&game.CardDef{CardFace: game.CardFace{Name: "Insufficient pain source", Types: []types.Card{types.Artifact}}},
		&game.ManaAbility{
			Text:            "Tap, pay 1 life: Add G.",
			AdditionalCosts: []cost.Additional{cost.T, {Kind: cost.AdditionalPayLife, Amount: 1}},
			Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.AddMana{
				Player: opt.Val(game.ControllerReference()), ManaColor: mana.G, Amount: game.Fixed(1),
			}}}}.Ability(),
		})
	policies := simPassPolicies()
	policies[game.Player1] = &controlScript{choice: func(obs PlayerObservation, request game.ChoiceRequest) []int {
		if request.Prompt == "Activate a mana ability while paying?" {
			return []int{1}
		}
		return panglacialCastChoices(obs, request)
	}}
	NewEngine(nil).searchLibraryRevealOnlyCards(g, &game.StackObject{Controller: game.Player1}, policies, &TurnLog{},
		game.Player1, game.SearchSpec{RevealOnly: true, FailToFindPolicy: game.SearchMayFailToFind}, 1)
	if !g.Players[game.Player1].Library.Contains(wurm) || g.Stack.Size() != 0 || g.Players[game.Player1].Life != 40 ||
		g.Players[game.Player1].ManaPool.Total() != 0 || source.Tapped || len(g.ManaPaymentScopes) != 0 {
		t.Fatal("an unpaid proposal retained reversible mana activation costs or output")
	}
}

func TestPanglacialWurmUnpaidProposalKeepsIrreversibleManaEffects(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	remaining := addCardToLibrary(g, game.Player1, vanillaCreature("Remaining bottom", 1, 1))
	wurm := addCardToLibrary(g, game.Player1, authored.PanglacialWurm.New())
	drawn := addCardToLibrary(g, game.Player1, vanillaCreature("Drawn for mana", 1, 1))
	drawSource := addComplexManaAbilityPermanent(g, game.Player1,
		&game.CardDef{CardFace: game.CardFace{Name: "Irreversible source", Types: []types.Card{types.Artifact}}},
		&game.ManaAbility{
			Text: "Tap: Add G and draw a card.", AdditionalCosts: []cost.Additional{cost.T},
			Content: game.Mode{Sequence: []game.Instruction{
				{Primitive: game.AddMana{Player: opt.Val(game.ControllerReference()), ManaColor: mana.G, Amount: game.Fixed(1)}},
				{Primitive: game.Draw{Player: game.ControllerReference(), Amount: game.Fixed(1)}},
			}}.Ability(),
		})
	painSource := addComplexManaAbilityPermanent(g, game.Player1,
		&game.CardDef{CardFace: game.CardFace{Name: "Later reversible source", Types: []types.Card{types.Artifact}}},
		&game.ManaAbility{
			Text:            "Tap, pay 1 life: Add G.",
			AdditionalCosts: []cost.Additional{cost.T, {Kind: cost.AdditionalPayLife, Amount: 1}},
			Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.AddMana{
				Player: opt.Val(game.ControllerReference()), ManaColor: mana.G, Amount: game.Fixed(1),
			}}}}.Ability(),
		})
	policies := simPassPolicies()
	policies[game.Player1] = &controlScript{choice: func(obs PlayerObservation, request game.ChoiceRequest) []int {
		if request.Prompt == "Activate a mana ability while paying?" {
			return []int{1}
		}
		if request.Kind == game.ChoiceSearch {
			return nil
		}
		return panglacialCastChoices(obs, request)
	}}
	NewEngine(nil).searchLibraryRevealOnlyCards(g, &game.StackObject{Controller: game.Player1}, policies, &TurnLog{},
		game.Player1, game.SearchSpec{RevealOnly: true, FailToFindPolicy: game.SearchMayFailToFind}, 1)
	if !slices.Equal(g.Players[game.Player1].Library.All(), []id.ID{wurm, remaining}) ||
		!g.Players[game.Player1].Hand.Contains(drawn) || !drawSource.Tapped || painSource.Tapped ||
		g.Players[game.Player1].Life != 40 || g.Players[game.Player1].ManaPool.Total() != 1 || g.Stack.Size() != 0 ||
		len(g.ManaPaymentScopes) != 0 {
		t.Fatal("illegal-cast reversal either undid a forbidden draw or retained later reversible costs")
	}
}

func TestPanglacialWurmEveryLibrarySearchEntryPath(t *testing.T) {
	for _, path := range []string{"private", "revealed", "slots", "named multizone", "chosen multizone", "face-down exile"} {
		t.Run(path, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			wurm := addCardToLibrary(g, game.Player1, authored.PanglacialWurm.New())
			addCardToLibrary(g, game.Player1, vanillaCreature("Findable creature", 2, 2))
			g.Players[game.Player1].ManaPool.Add(mana.G, 7)
			policies := simPassPolicies()
			policies[game.Player1] = &controlScript{choice: panglacialCastChoices}
			engine, object, log := NewEngine(nil), &game.StackObject{Controller: game.Player1}, &TurnLog{}
			spec := game.SearchSpec{SourceZone: zone.Library, Destination: zone.Hand}
			switch path {
			case "private":
				engine.searchLibrary(g, object, policies, log, game.Player1, game.Player1, spec, 1)
			case "revealed":
				spec.Destination, spec.RevealOnly, spec.Reveal = zone.None, true, true
				engine.searchLibraryRevealOnlyCards(g, object, policies, log, game.Player1, spec, 1)
			case "slots":
				spec.SlotFilters = []game.Selection{{RequiredTypes: []types.Card{types.Creature}}}
				engine.searchLibrary(g, object, policies, log, game.Player1, game.Player1, spec, 1)
			case "named multizone":
				spec.AlsoGraveyard, spec.Name = true, "Findable creature"
				engine.searchLibraryAndGraveyard(g, object, policies, log, game.Player1, spec)
			case "chosen multizone":
				spec.AlsoGraveyard, spec.ConditionalShuffle = true, true
				engine.searchLibraryAndGraveyardChoice(g, object, policies, log, game.Player1, game.Player1, spec)
			case "face-down exile":
				spec.Destination, spec.ExileFaceDown = zone.Exile, true
				engine.searchLibraryExileFaceDown(g, object, policies, log, game.Player1, spec, 1)
			default:
				t.Fatal("unknown library search entry path")
			}
			top, ok := g.Stack.Peek()
			if !ok || top.SourceID != wurm || g.Players[game.Player1].ManaPool.Total() != 0 ||
				len(g.SearchControlScopes) != 0 || len(g.ManaPaymentScopes) != 0 {
				t.Fatal("search path did not expose and close the shared paid before-find opportunity")
			}
		})
	}
}

func TestSearchCastingRejectsMidResolutionSimulatorResume(t *testing.T) {
	for _, payment := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		if payment {
			g.ManaPaymentScopes = []game.PlayerID{game.Player2}
		} else {
			g.SearchControlScopes = []game.SearchControlScope{{Player: game.Player2, BeforeFind: true}}
		}
		clone := g.Clone()
		if len(engine.Simulator().LegalActions(clone, game.Player1)) != 0 ||
			engine.applyAction(clone, game.Player1, action.Pass()) {
			t.Fatal("synchronous metadata granted ordinary priority")
		}
		if branch, accepted := engine.Simulator().Apply(clone, game.Player1, action.Pass(), simPassPolicies()); accepted || branch != nil {
			t.Fatal("Simulator admitted an ordinary action while an unclonable synchronous procedure was active")
		}
		func() {
			defer func() {
				if _, explicit := recover().(UnsupportedError); !explicit {
					t.Fatal("unsupported mid-resolution continuation was not reported explicitly")
				}
			}()
			engine.Simulator().ResolvePriority(clone, simPassPolicies())
		}()
	}
}
