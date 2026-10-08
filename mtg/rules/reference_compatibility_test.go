package rules

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestCompiledOptionalLifePaymentDrawCompatibility(t *testing.T) {
	def := compileUnlessCard(t, cardgen.ScryfallCard{Name: "Life Payment", Layout: "normal",
		TypeLine: "Sorcery", OracleText: "You may pay 2 life. If you do, draw a card."})
	for _, tc := range []struct {
		name                  string
		life                  int
		accept, forbidden     bool
		wantLife, draws, asks int
	}{
		{"paid", 20, true, false, 18, 1, 1},
		{"exact", 2, true, false, 0, 1, 1},
		{"declined", 20, false, false, 20, 0, 1},
		{"insufficient", 1, true, false, 1, 0, 0},
		{"forbidden", 20, true, true, 20, 0, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			g.Players[game.Player1].Life = tc.life
			addCardToLibrary(g, game.Player1, vanillaCreature("Drawn", 1, 1))
			if tc.forbidden {
				g.RuleEffects = append(g.RuleEffects, game.RuleEffect{ID: g.IDGen.Next(),
					Kind: game.RuleEffectLifeTotalCantChange, Controller: game.Player1, AffectedPlayer: game.PlayerYou})
			}
			agent := &libraryPaymentAgent{accept: tc.accept}
			obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1}
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.SpellAbility.Val,
				[game.NumPlayers]PlayerAgent{game.Player1: agent}, &TurnLog{})
			if g.Players[game.Player1].Life != tc.wantLife || g.Players[game.Player1].Hand.Size() != tc.draws ||
				len(agent.prompts) != tc.asks || g.Players[game.Player2].Life != 40 {
				t.Fatalf("life/draws/prompts = %d/%d/%d, want %d/%d/%d", g.Players[game.Player1].Life,
					g.Players[game.Player1].Hand.Size(), len(agent.prompts), tc.wantLife, tc.draws, tc.asks)
			}
			assertResultKeysCleaned(t, obj, "if-you-do")
		})
	}
}

func TestCompiledAnotherUnionRecipientExcludesDealer(t *testing.T) {
	def := compileUnlessCard(t, cardgen.ScryfallCard{Name: "Another Recipient", Layout: "normal",
		TypeLine: "Instant", OracleText: "Target creature you control deals damage equal to its power to another target creature, planeswalker, or battle."})
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	card := addCardToHand(g, game.Player1, def)
	dealer := addCombatPermanent(g, game.Player1, vanillaCreature("Dealer", 3, 5))
	recipient := addCombatPermanent(g, game.Player2, vanillaCreature("Recipient", 1, 5))
	g.Turn.Phase, g.Turn.Step = game.PhasePrecombatMain, game.StepNone
	found := false
	for _, act := range NewEngine(nil).legalActions(g, game.Player1) {
		cast, ok := act.CastSpellPayload()
		if !ok || cast.CardID != card {
			continue
		}
		found = true
		if len(cast.Targets) != 2 || cast.Targets[0].PermanentID != dealer.ObjectID ||
			cast.Targets[1].PermanentID != recipient.ObjectID {
			t.Fatalf("another-recipient targets = %+v, want distinct dealing/recipient objects", cast.Targets)
		}
	}
	if !found {
		t.Fatal("valid distinct target pair was not offered")
	}
	obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1, Kind: game.StackSpell,
		SourceID: card, SourceCardID: card,
		Targets: []game.Target{game.PermanentTarget(dealer.ObjectID), game.PermanentTarget(recipient.ObjectID)}}
	NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.SpellAbility.Val, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if recipient.MarkedDamage != 3 || dealer.MarkedDamage != 0 {
		t.Fatalf("dealer/recipient damage = %d/%d, want 0/3", dealer.MarkedDamage, recipient.MarkedDamage)
	}
}

func TestCompiledSplitAddendumExactTargetCompatibility(t *testing.T) {
	def := compileUnlessCard(t, cardgen.ScryfallCard{Name: "Split Addendum", Layout: "normal",
		TypeLine: "Instant", OracleText: "Target creature gets +2/+2 until end of turn.\nAddendum — If you cast this spell during your main phase, that creature gains flying until end of turn."})
	for _, main := range []bool{false, true} {
		t.Run(fmt.Sprint(main), func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			target := addCombatPermanent(g, game.Player2, vanillaCreature("Target", 2, 3))
			decoy := addCombatPermanent(g, game.Player2, vanillaCreature("Decoy", 2, 3))
			obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1, Kind: game.StackSpell,
				SourceCardID: addCardInstance(g, game.Player1, def), CastDuringControllerMainPhase: main,
				Targets: []game.Target{game.PermanentTarget(target.ObjectID)}}
			// Resolution phase deliberately disagrees with the captured cast phase.
			g.Turn.Phase = game.PhasePrecombatMain
			if main {
				g.Turn.Phase = game.PhaseCombat
			}
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.SpellAbility.Val, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			if effectivePower(g, target) != 4 || effectivePower(g, decoy) != 2 ||
				hasKeyword(g, target, game.Flying) != main || hasKeyword(g, decoy, game.Flying) {
				t.Fatal("split Addendum lost exact target, unconditional pump, or captured main-phase timing")
			}
		})
	}
}

func TestCompiledObservedCardPaymentCompatibility(t *testing.T) {
	def := compileUnlessCard(t, cardgen.ScryfallCard{Name: "Observed Payment", Layout: "normal", TypeLine: "Artifact",
		OracleText: "{T}: Look at the top card of target player's library. If it's a nonland card, you may pay 2 life. If you do, put it into that player's graveyard."})
	for _, tc := range []struct {
		name                    string
		land, empty, accept     bool
		forbidden, changed      bool
		life, wantLife, prompts int
		moved                   bool
	}{
		{name: "paid", accept: true, life: 20, wantLife: 18, prompts: 1, moved: true},
		{name: "declined", life: 20, wantLife: 20, prompts: 1},
		{name: "land", land: true, accept: true, life: 20, wantLife: 20},
		{name: "empty", empty: true, accept: true, life: 20, wantLife: 20},
		{name: "insufficient", accept: true, life: 1, wantLife: 1},
		{name: "forbidden", accept: true, forbidden: true, life: 20, wantLife: 20},
		{name: "incarnation changed after observation", accept: true, changed: true, life: 20, wantLife: 18, prompts: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			g.Players[game.Player1].Life = tc.life
			decoy := addCardToLibrary(g, game.Player1, vanillaCreature("Decoy", 2, 3))
			observed := addCardToLibrary(g, game.Player3, vanillaCreature("Observed", 2, 3))
			if tc.empty {
				g.Players[game.Player3].Library.Remove(observed)
			}
			if tc.land {
				g.CardInstances[observed].Def.Types = []types.Card{types.Land}
			}
			if tc.forbidden {
				g.RuleEffects = append(g.RuleEffects, game.RuleEffect{ID: g.IDGen.Next(),
					Kind: game.RuleEffectLifeTotalCantChange, Controller: game.Player1, AffectedPlayer: game.PlayerYou})
			}
			controller := &libraryPaymentAgent{accept: tc.accept}
			if tc.changed {
				controller.afterMay = func() { g.CardInstances[observed].ZoneVersion++ }
			}
			owner := &libraryPaymentAgent{accept: true}
			source := addCombatPermanent(g, game.Player1, def)
			obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1,
				SourceID: source.ObjectID, SourceCardID: source.CardInstanceID,
				Targets: []game.Target{game.PlayerTarget(game.Player3)}}
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.ActivatedAbilities[0].Content,
				[game.NumPlayers]PlayerAgent{game.Player1: controller, game.Player3: owner}, &TurnLog{})
			if g.Players[game.Player1].Life != tc.wantLife || g.Players[game.Player3].Life != 40 ||
				len(controller.prompts) != tc.prompts || len(owner.prompts) != 0 ||
				g.Players[game.Player3].Graveyard.Contains(observed) != tc.moved ||
				!g.Players[game.Player1].Library.Contains(decoy) {
				t.Fatalf("observed payment lost payer, card incarnation, or owner: life=%d prompts=%d moved=%v",
					g.Players[game.Player1].Life, len(controller.prompts), g.Players[game.Player3].Graveyard.Contains(observed))
			}
		})
	}
}

func TestCompiledOptionalReturnConditionCompatibility(t *testing.T) {
	def := compileUnlessCard(t, cardgen.ScryfallCard{Name: "Optional Returned Elf", Layout: "normal", TypeLine: "Instant",
		OracleText: "You may return target creature card from your graveyard to the battlefield. If it's an Elf, draw a card."})
	for _, tc := range []struct {
		name                string
		accept, elf, stale  bool
		wantReturned, draws bool
	}{
		{"returned Elf", true, true, false, true, true},
		{"returned non-Elf", true, false, false, true, false},
		{"declined Elf", false, true, false, false, false},
		{"stale Elf", true, true, true, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			subject := vanillaCreature("Subject", 2, 3)
			if tc.elf {
				subject.Subtypes = []types.Sub{types.Elf}
			}
			card := addCardToGraveyard(g, game.Player1, subject)
			decoyDef := vanillaCreature("Subject", 2, 3)
			decoyDef.Subtypes = []types.Sub{types.Elf}
			decoy := addCombatPermanent(g, game.Player2, decoyDef)
			drawn := addCardToLibrary(g, game.Player1, vanillaCreature("Drawn", 1, 1))
			obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1, Kind: game.StackSpell,
				SourceCardID: addCardInstance(g, game.Player1, def), Targets: []game.Target{currentCardTarget(t, g, card)}}
			if tc.stale {
				moveCardBetweenZones(g, game.Player1, card, zone.Graveyard, zone.Hand)
				moveCardBetweenZones(g, game.Player1, card, zone.Hand, zone.Graveyard)
			}
			agent := &libraryPaymentAgent{accept: tc.accept}
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.SpellAbility.Val,
				[game.NumPlayers]PlayerAgent{game.Player1: agent}, &TurnLog{})
			if (permanentForCard(g, card) != nil) != tc.wantReturned ||
				g.Players[game.Player1].Hand.Contains(drawn) != tc.draws ||
				g.Players[game.Player2].Hand.Size() != 0 || decoy.Controller != game.Player2 {
				t.Fatal("optional condition acquired a skipped/stale/decoy subject or wrong player")
			}
			key := game.PublishedLinkedKey(def.SpellAbility.Val.Modes[0].Sequence[0].Primitive)
			if _, ok := resolveObjectReference(g, obj, game.LinkedObjectReference(string(key))); ok {
				t.Fatal("optional entered-object publication leaked")
			}
		})
	}
}

func TestCompiledLatestObservationAfterReturnCompatibility(t *testing.T) {
	def := compileUnlessCard(t, cardgen.ScryfallCard{Name: "Latest Observed Elf", Layout: "normal", TypeLine: "Instant",
		OracleText: "Return target creature card from your graveyard to the battlefield. Reveal the top card of your library. If it's an Elf, draw a card."})
	for _, returnedElf := range []bool{false, true} {
		for _, observedElf := range []bool{false, true} {
			t.Run(fmt.Sprintf("returned=%t/observed=%t", returnedElf, observedElf), func(t *testing.T) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				returnedDef, observedDef := vanillaCreature("Returned", 1, 1), vanillaCreature("Observed", 1, 1)
				if returnedElf {
					returnedDef.Subtypes = []types.Sub{types.Elf}
				}
				if observedElf {
					observedDef.Subtypes = []types.Sub{types.Elf}
				}
				card := addCardToGraveyard(g, game.Player1, returnedDef)
				observed := addCardToLibrary(g, game.Player1, observedDef)
				obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1,
					Targets: []game.Target{currentCardTarget(t, g, card)}}
				NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.SpellAbility.Val, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
				if permanentForCard(g, card) == nil || g.Players[game.Player1].Hand.Contains(observed) != observedElf {
					t.Fatal("latest observation condition selected the returned permanent instead of the revealed card")
				}
			})
		}
	}
}

func TestCompiledObservedManaValueThresholdCompatibility(t *testing.T) {
	def := compileUnlessCard(t, cardgen.ScryfallCard{Name: "Observed Mana Value", Layout: "normal", TypeLine: "Creature",
		OracleText: "When this creature enters, reveal the top card of your library. If it had mana value 3 or less, draw a card."})
	for _, tc := range []struct {
		name         string
		value        int
		empty, stale bool
		draw         bool
	}{
		{"zero", 0, false, false, true},
		{"boundary", 3, false, false, true},
		{"above", 4, false, false, false},
		{"empty", 0, true, false, false},
		{"stale", 3, false, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			subject := evidenceCard("Observed", tc.value)
			card := addCardToLibrary(g, game.Player1, subject)
			if tc.empty {
				g.Players[game.Player1].Library.Remove(card)
			}
			source := addCombatPermanent(g, game.Player1, def)
			obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1, SourceID: source.ObjectID, SourceCardID: source.CardInstanceID}
			sequence := def.TriggeredAbilities[0].Content.Modes[0].Sequence
			if tc.stale {
				restore := enterLocalProductFrame(g, obj, sequence)
				defer restore()
				resolver := newEffectResolver(NewEngine(nil), g, obj, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
				resolver.resolveInstruction(&sequence[0])
				g.CardInstances[card].ZoneVersion++
				resolver.resolveInstruction(&sequence[1])
			} else {
				NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.TriggeredAbilities[0].Content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			}
			if g.Players[game.Player1].Hand.Contains(card) != tc.draw || g.Players[game.Player2].Hand.Size() != 0 {
				t.Fatal("observed mana-value condition lost threshold, availability, incarnation, or player isolation")
			}
		})
	}
}

func TestCompiledReturnedManaValueLifeRiderCompatibility(t *testing.T) {
	def := compileUnlessCard(t, cardgen.ScryfallCard{Name: "Returned Mana Value", Layout: "normal", TypeLine: "Instant",
		OracleText: "Put target creature card from a graveyard onto the battlefield under your control. You lose life equal to its mana value."})
	for _, stale := range []bool{false, true} {
		t.Run(fmt.Sprint(stale), func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			targetDef := evidenceCard("Returned", 4)
			targetDef.Types = []types.Card{types.Creature}
			card := addCardToGraveyard(g, game.Player3, targetDef)
			addCombatPermanent(g, game.Player1, evidenceCard("Decoy", 9))
			key := string(def.SpellAbility.Val.Modes[0].Sequence[0].PublishResult)
			obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1,
				Targets:           []game.Target{currentCardTarget(t, g, card)},
				ResolutionResults: map[string]game.InstructionResolutionResult{key: {Succeeded: true}}}
			if stale {
				moveCardBetweenZones(g, game.Player3, card, zone.Graveyard, zone.Hand)
				moveCardBetweenZones(g, game.Player3, card, zone.Hand, zone.Graveyard)
			}
			NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.SpellAbility.Val, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			wantLife := 36
			if stale {
				wantLife = 40
			}
			returned := permanentForCard(g, card)
			if g.Players[game.Player1].Life != wantLife || g.Players[game.Player3].Life != 40 ||
				(returned != nil) == stale || returned != nil && returned.Controller != game.Player1 {
				t.Fatal("life rider read the wrong subject/player or ignored failed battlefield entry")
			}
			if result, ok := obj.ResolutionResults[key]; !ok || result.Succeeded == stale {
				t.Fatalf("entry receipt = %+v, want fresh success=%t", result, !stale)
			}
		})
	}
}
