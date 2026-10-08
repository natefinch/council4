package rules

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func compiledReferenceRoleOriginal(t *testing.T, name string) *game.CardDef {
	t.Helper()
	data, err := os.ReadFile("../../cardgen/testdata/reference-role-originals.json")
	if err != nil {
		t.Fatal(err)
	}
	var cards []cardgen.ScryfallCard
	if err := json.Unmarshal(data, &cards); err != nil {
		t.Fatal(err)
	}
	for _, card := range cards {
		if card.Name == name {
			return compileUnlessCard(t, card)
		}
	}
	t.Fatalf("missing full original %q", name)
	return nil
}

func TestReferenceRoleOriginalSpellDamageAttribution(t *testing.T) {
	def := compiledReferenceRoleOriginal(t, "Beacon of Destruction")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	card := addCardToHand(g, game.Player1, def)
	obj := &game.StackObject{
		ID: g.IDGen.Next(), SourceID: card, SourceCardID: card,
		Kind: game.StackSpell, Controller: game.Player2,
		Targets: []game.Target{{Kind: game.TargetPlayer, PlayerID: game.Player3}},
	}
	before := g.Players[game.Player3].Life
	engine := NewEngine(nil)
	engine.resolveAbilityContentWithChoices(g, obj, def.SpellAbility.Val, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if got := g.Players[game.Player3].Life; got != before-5 {
		t.Fatalf("resolving spell damage = %d, want %d", before-got, 5)
	}
}

func TestReferenceRoleOriginalMirkwoodDeparturePower(t *testing.T) {
	def := compiledReferenceRoleOriginal(t, "Mirkwood Elk")
	for _, power := range []int{0, 4} {
		t.Run(string(rune('0'+power)), func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			source := addCombatPermanent(g, game.Player2, def)
			elf := vanillaCreature("Departure Elf", power, 4)
			elf.Subtypes = []types.Sub{types.Elf}
			card := addCardToGraveyard(g, game.Player2, elf)
			decoy := addCardToGraveyard(g, game.Player1, elf)
			obj := &game.StackObject{
				ID: g.IDGen.Next(), Kind: game.StackTriggeredAbility, Controller: game.Player2,
				SourceID: source.ObjectID, SourceCardID: source.CardInstanceID,
				Targets: []game.Target{{Kind: game.TargetCard, CardID: card,
					CardZoneVersion: g.CardInstances[card].ZoneVersion, CardZoneVersionSet: true}},
			}
			before := g.Players[game.Player2].Life
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.TriggeredAbilities[0].Content,
				[game.NumPlayers]PlayerAgent{}, &TurnLog{})
			if !g.Players[game.Player2].Hand.Contains(card) || !g.Players[game.Player1].Graveyard.Contains(decoy) {
				t.Fatal("return moved wrong owner/card")
			}
			if got := g.Players[game.Player2].Life - before; got != power {
				t.Fatalf("departure-card power life = %d, want %d", got, power)
			}
		})
	}
}

func TestReferenceRoleOriginalKardurEnteredKeyword(t *testing.T) {
	def := compiledReferenceRoleOriginal(t, "Kardur's Vicious Return")
	if len(def.ChapterAbilities) != 3 {
		t.Fatalf("full Saga chapter count = %d", len(def.ChapterAbilities))
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player2, def)
	card := addCardToGraveyard(g, game.Player2, referenceSubjectCreature(t))
	decoy := addCombatPermanent(g, game.Player3, referenceSubjectCreature(t))
	obj := &game.StackObject{
		ID: g.IDGen.Next(), Kind: game.StackTriggeredAbility, Controller: game.Player2,
		SourceID: source.ObjectID, SourceCardID: source.CardInstanceID,
		Targets: []game.Target{{Kind: game.TargetCard, CardID: card,
			CardZoneVersion: g.CardInstances[card].ZoneVersion, CardZoneVersionSet: true}},
	}
	NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.ChapterAbilities[2].Content,
		[game.NumPlayers]PlayerAgent{}, &TurnLog{})
	entered := permanentForCard(g, card)
	if entered == nil || !hasKeyword(g, entered, game.Haste) || hasKeyword(g, decoy, game.Haste) {
		t.Fatal("haste must follow the actual entered occurrence across its counter clause")
	}
	if entered.Counters.Get(counter.PlusOnePlusOne) != 1 {
		t.Fatal("same entered creature lost its counter")
	}
}

func TestReferenceRoleOriginalResolvingSpellReturnHand(t *testing.T) {
	def := compiledReferenceRoleOriginal(t, "View from Above")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addColoredCombatCreature(g, game.Player2, "White condition permanent", color.White)
	target := addCombatPermanent(g, game.Player2, referenceSubjectCreature(t))
	card := addCardToHand(g, game.Player1, def)
	obj := &game.StackObject{
		ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: card, SourceCardID: card,
		Controller: game.Player2, Targets: []game.Target{game.PermanentTarget(target.ObjectID)},
	}

	g.Players[game.Player1].Hand.Remove(card)
	g.Stack.Push(obj)
	NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player1].Hand.Contains(card) || permanentForCard(g, target.CardInstanceID) == nil {
		t.Fatalf("only resolving spell should return to owner hand; target stays: hand=%v graveyard=%v flags=%+v content=%+v",
			g.Players[game.Player1].Hand.All(), g.Players[game.Player1].Graveyard.All(), obj, def.SpellAbility.Val)
	}
	if !hasKeyword(g, target, game.Flying) {
		t.Fatal("target lost complete spell's flying effect")
	}
}

func TestReferenceRoleOriginalGraveyardFunctionSource(t *testing.T) {
	def := compiledReferenceRoleOriginal(t, "Squee, Goblin Nabob")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	card := addCardToGraveyard(g, game.Player2, def)
	version := g.CardInstances[card].ZoneVersion
	g.AppendEvent(game.Event{Kind: game.EventBeginningOfStep, Step: game.StepUpkeep, Player: game.Player2})
	engine := NewEngine(nil)
	agents := [game.NumPlayers]PlayerAgent{}
	agents[game.Player2] = &repeatLandChoiceAgent{mayAnswers: []bool{true}}
	log := &TurnLog{}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, log) {
		t.Fatal("complete original's graveyard-function trigger was not detected")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceCardID != card || obj.SourceZone != zone.Graveyard ||
		obj.SourceZoneVersion != version {
		t.Fatal("graveyard-function trigger did not retain its exact source incarnation")
	}
	engine.resolveTopOfStackWithChoices(g, agents, log)
	if !g.Players[game.Player2].Hand.Contains(card) {
		t.Fatal("complete original did not return its exact graveyard source")
	}
}

func TestReferenceRoleOriginalTopExiledCardProduct(t *testing.T) {
	def := compiledReferenceRoleOriginal(t, "Court of Locthwain")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, def)
	first := addCardToLibrary(g, game.Player3, vanillaCreature("Opponent Top", 2, 2))
	second := addCardToLibrary(g, game.Player3, vanillaCreature("Opponent Other", 2, 2))
	own := addCardToLibrary(g, game.Player1, vanillaCreature("Controller Top", 2, 2))
	obj := &game.StackObject{
		ID: g.IDGen.Next(), Kind: game.StackTriggeredAbility, Controller: game.Player1,
		SourceID: source.ObjectID, SourceCardID: source.CardInstanceID,
		Targets: []game.Target{{Kind: game.TargetPlayer, PlayerID: game.Player3}},
	}
	NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.TriggeredAbilities[1].Content,
		[game.NumPlayers]PlayerAgent{}, &TurnLog{})
	exiled := 0
	var card id.ID
	for _, candidate := range []id.ID{first, second} {
		if g.Players[game.Player3].Exile.Contains(candidate) {
			exiled++
			card = candidate
		}
	}
	if exiled != 1 || !g.Players[game.Player1].Library.Contains(own) {
		t.Fatalf("exiled %d target-opponent cards; controller library intact=%v", exiled,
			g.Players[game.Player1].Library.Contains(own))
	}
	linked := linkedObjects(g, linkedObjectSourceKey(g, obj, "court-of-locthwain-exile"))
	if len(linked) != 1 || linked[0].CardID != card ||
		linked[0].CardZoneVersion != g.CardInstances[card].ZoneVersion {
		t.Fatalf("permission does not name the actual exiled card version: %+v", linked)
	}
	if !moveCardBetweenZonesWithPlacement(g, game.Player3, card, zone.Exile, zone.Graveyard, false) ||
		!moveCardBetweenZonesWithPlacement(g, game.Player3, card, zone.Graveyard, zone.Exile, false) {
		t.Fatal("could not create a later exile incarnation")
	}
	if _, available := linkedCardInstance(g, linked[0]); available {
		t.Fatal("a later exile incarnation satisfied the original exile product")
	}
}

func TestReferenceRoleOriginalEnclosingSpellTarget(t *testing.T) {
	def := compiledReferenceRoleOriginal(t, "Code of Constraint")
	for _, main := range []bool{true, false} {
		t.Run(map[bool]string{true: "main phase", false: "not main phase"}[main], func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			addCardToLibrary(g, game.Player1, vanillaCreature("Drawn", 1, 1))
			target := addCombatPermanent(g, game.Player2, vanillaCreature("Target", 5, 5))
			decoy := addCombatPermanent(g, game.Player2, vanillaCreature("Decoy", 5, 5))
			card := addCardToHand(g, game.Player1, def)
			obj := &game.StackObject{
				ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: card, SourceCardID: card,
				Controller: game.Player1, CastDuringControllerMainPhase: main,
				Targets: []game.Target{game.PermanentTarget(target.ObjectID)},
			}
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.SpellAbility.Val,
				[game.NumPlayers]PlayerAgent{}, &TurnLog{})
			if effectivePower(g, target) != 1 || effectivePower(g, decoy) != 5 {
				t.Fatal("power modification did not follow the exact target")
			}
			if target.Tapped != main || decoy.Tapped {
				t.Fatalf("addendum tapped target=%v decoy=%v, want target=%v", target.Tapped, decoy.Tapped, main)
			}
		})
	}
}

func TestReferenceRoleOriginalResolvingSpellReturnHandControls(t *testing.T) {
	def := compiledReferenceRoleOriginal(t, "View from Above")
	for _, tc := range []struct {
		name               string
		white, flashback   bool
		copy               bool
		hand, grave, exile bool
	}{
		{name: "condition false", grave: true},
		{name: "flashback replacement", white: true, flashback: true, exile: true},
		{name: "copy has no card", white: true, copy: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			if tc.white {
				addColoredCombatCreature(g, game.Player2, "White condition permanent", color.White)
			}
			target := addCombatPermanent(g, game.Player2, referenceSubjectCreature(t))
			card := addCardToHand(g, game.Player1, def)
			g.Players[game.Player1].Hand.Remove(card)
			spell := func(copy bool) *game.StackObject {
				return &game.StackObject{
					ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: card, SourceCardID: card,
					Controller: game.Player2, Flashback: tc.flashback, Copy: copy,
					Targets: []game.Target{game.PermanentTarget(target.ObjectID)},
				}
			}
			if tc.copy {
				g.Stack.Push(spell(false))
			}
			g.Stack.Push(spell(tc.copy))
			NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
			owner := g.Players[game.Player1]
			if owner.Hand.Contains(card) != tc.hand || owner.Graveyard.Contains(card) != tc.grave ||
				owner.Exile.Contains(card) != tc.exile {
				t.Fatalf("card disposition hand=%v graveyard=%v exile=%v", owner.Hand.Contains(card),
					owner.Graveyard.Contains(card), owner.Exile.Contains(card))
			}
			if permanentForCard(g, target.CardInstanceID) == nil || !hasKeyword(g, target, game.Flying) {
				t.Fatal("resolving-spell disposition disturbed its target")
			}
		})
	}
}

func TestReferenceRoleOriginalMirkwoodStaleTarget(t *testing.T) {
	def := compiledReferenceRoleOriginal(t, "Mirkwood Elk")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player2, def)
	elf := vanillaCreature("Departure Elf", 4, 4)
	elf.Subtypes = []types.Sub{types.Elf}
	card := addCardToGraveyard(g, game.Player2, elf)
	obj := &game.StackObject{
		ID: g.IDGen.Next(), Kind: game.StackTriggeredAbility, Controller: game.Player2,
		SourceID: source.ObjectID, SourceCardID: source.CardInstanceID,
		Targets: []game.Target{{Kind: game.TargetCard, CardID: card,
			CardZoneVersion: g.CardInstances[card].ZoneVersion, CardZoneVersionSet: true}},
	}
	if !moveCardBetweenZonesWithPlacement(g, game.Player2, card, zone.Graveyard, zone.Exile, false) ||
		!moveCardBetweenZonesWithPlacement(g, game.Player2, card, zone.Exile, zone.Graveyard, false) {
		t.Fatal("could not create a later graveyard incarnation")
	}
	before := g.Players[game.Player2].Life
	NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.TriggeredAbilities[0].Content,
		[game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if g.Players[game.Player2].Hand.Contains(card) || g.Players[game.Player2].Life != before {
		t.Fatal("a later incarnation was returned or supplied departure power")
	}
}

func TestReferenceRoleOriginalConditionedSpellDamageSource(t *testing.T) {
	def := compiledReferenceRoleOriginal(t, "Jilt")
	for _, kicked := range []bool{true, false} {
		t.Run(map[bool]string{true: "kicked", false: "not kicked"}[kicked], func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			returned := addCombatPermanent(g, game.Player2, vanillaCreature("Returned", 2, 3))
			damaged := addCombatPermanent(g, game.Player2, vanillaCreature("Damaged", 2, 3))
			card := addCardToHand(g, game.Player1, def)
			obj := &game.StackObject{
				ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: card, SourceCardID: card,
				Controller: game.Player1, KickerPaid: kicked,
				Targets: []game.Target{game.PermanentTarget(returned.ObjectID), game.PermanentTarget(damaged.ObjectID)},
			}
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.SpellAbility.Val,
				[game.NumPlayers]PlayerAgent{}, &TurnLog{})
			if !g.Players[game.Player2].Hand.Contains(returned.CardInstanceID) {
				t.Fatal("first target was not returned")
			}
			want := 0
			if kicked {
				want = 2
			}
			if damaged.MarkedDamage != want {
				t.Fatalf("second target damage = %d, want %d", damaged.MarkedDamage, want)
			}
			for _, event := range g.Events {
				if event.Kind == game.EventDamageDealt && event.SourceID == returned.ObjectID {
					t.Fatal("returned creature was used as the spell's damage source")
				}
			}
		})
	}
}

func TestReferenceRoleOriginalKardurStaleTarget(t *testing.T) {
	def := compiledReferenceRoleOriginal(t, "Kardur's Vicious Return")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player2, def)
	card := addCardToGraveyard(g, game.Player2, referenceSubjectCreature(t))
	decoy := addCombatPermanent(g, game.Player2, referenceSubjectCreature(t))
	obj := &game.StackObject{
		ID: g.IDGen.Next(), Kind: game.StackTriggeredAbility, Controller: game.Player2,
		SourceID: source.ObjectID, SourceCardID: source.CardInstanceID,
		Targets: []game.Target{{Kind: game.TargetCard, CardID: card,
			CardZoneVersion: g.CardInstances[card].ZoneVersion, CardZoneVersionSet: true}},
	}
	if !moveCardBetweenZonesWithPlacement(g, game.Player2, card, zone.Graveyard, zone.Exile, false) ||
		!moveCardBetweenZonesWithPlacement(g, game.Player2, card, zone.Exile, zone.Graveyard, false) {
		t.Fatal("could not create a later graveyard incarnation")
	}
	NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.ChapterAbilities[2].Content,
		[game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if permanentForCard(g, card) != nil || hasKeyword(g, decoy, game.Haste) ||
		decoy.Counters.Get(counter.PlusOnePlusOne) != 0 {
		t.Fatal("stale return produced an entry or its riders reached another creature")
	}
}
