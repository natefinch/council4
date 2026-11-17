package rules

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: pinned corpus 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4.
// Tidehollow Sculler (51b55057-8b14-4277-894b-4813b4f7eb39) and Mesmeric Fiend
// (fb46b300-8a39-40d2-8bc7-31b90c822e22): "When this creature enters, target
// opponent reveals their hand and you choose a nonland card from it. Exile that
// card. When this creature leaves the battlefield, return the exiled card to its
// owner's hand." Sculler ruling 2021-03-19: if it leaves before its enters
// ability resolves, the leaves ability resolves first and does nothing, then the
// enters ability exiles the card indefinitely. Severance Priest
// (3a38bf69-511f-4851-870f-8a65641d832e): "When this creature enters, target
// opponent reveals their hand. You may choose a nonland card from it. If you do,
// exile that card. When this creature leaves the battlefield, the exiled card's
// owner creates an X/X white Spirit creature token, where X is the mana value of
// the exiled card." Ruling 2025-04-04: if no card was exiled, no token. Brain
// Maggot (419b5b4f-cac7-44bf-b1e2-655b5a5d37b1): "When this creature enters,
// target opponent reveals their hand and you choose a nonland card from it.
// Exile that card until this creature leaves the battlefield." Rulings
// 2014-04-26: the exile and the return are two one-shot effects of one ability;
// if it left before the ability resolves the hand is still revealed but nothing
// is exiled; the card returns immediately, with no state-based actions between.
// Kitesail Freebooter (8aa1bbc9-b047-4d68-93f4-728e40dafc5e): "target opponent
// reveals their hand. You choose a noncreature, nonland card from it. Exile that
// card until this creature leaves the battlefield." Rulings 2020-06-23 match
// Brain Maggot's. CR 2026-11-06 verified copy: 400.7 (a card that changes zones
// is a new object), 603.10a (leaves-the-battlefield abilities look back), 607.1
// and 607.2a (a linked "the exiled card" ability refers only to cards the first
// ability put in exile that are still there), 607.3 (all such cards, summing a
// variable), 610.3 (the second one-shot effect returns the object to its
// previous zone immediately after the source leaves), 610.3a-b (nothing moves
// if the source already left), 610.3d (simultaneous returns). These test-local
// programs prove the shared operations; the cards stay unregistered.
const linkedHandExileProducer game.LinkedProducerKey = "test/linked-hand-exile/enter/exiles"

var (
	scullerEnterTestReference  = game.ProgramReference{ID: "test/tidehollow-sculler/enter", Version: "1"}
	scullerLeaveTestReference  = game.ProgramReference{ID: "test/tidehollow-sculler/leave", Version: "1"}
	priestEnterTestReference   = game.ProgramReference{ID: "test/severance-priest/enter", Version: "1"}
	priestLeaveTestReference   = game.ProgramReference{ID: "test/severance-priest/leave", Version: "1"}
	maggotEnterTestReference   = game.ProgramReference{ID: "test/brain-maggot/enter", Version: "1"}
	kitesailEnterTestReference = game.ProgramReference{ID: "test/kitesail-freebooter/enter", Version: "1"}
)

func targetOpponentSpec() game.TargetSpec {
	return game.TargetSpec{
		MinTargets: 1, MaxTargets: 1, Constraint: "target opponent", Allow: game.TargetAllowPlayer,
		Selection: opt.Val(game.Selection{Player: game.PlayerOpponent}),
	}
}

// revealHandExileEnterProgram has the target opponent reveal their hand, then
// exiles the chosen matching card through placement's duration fields.
func revealHandExileEnterProgram(reference game.ProgramReference, filter game.Selection, optional bool, placement cardprogram.CardPlacement) cardprogram.Program {
	return cardprogram.Program{Reference: reference, Resolve: func(ctx cardprogram.Context) error {
		revealed, err := ctx.RevealHand(cardprogram.HandReveal{
			Player: game.TargetPlayerReference(0), Filter: filter, Max: 1, Optional: optional,
		})
		if err != nil || len(revealed.Chosen) == 0 {
			return err
		}
		placement.Cards, placement.Destination = revealed.Chosen, zone.Exile
		_, err = ctx.MoveCards(placement)
		return err
	}}
}

// scullerLeaveProgram returns the linked exiled cards to their owners' hands,
// recording the handles it read.
func scullerLeaveProgram(read *[]cardprogram.FoundCard) cardprogram.Program {
	return cardprogram.Program{Reference: scullerLeaveTestReference, Resolve: func(ctx cardprogram.Context) error {
		exiled, err := ctx.LinkedExiledCards(linkedHandExileProducer)
		if err != nil {
			return err
		}
		*read = append(*read, exiled...)
		if len(exiled) == 0 {
			return nil
		}
		_, err = ctx.MoveCards(cardprogram.CardPlacement{Cards: exiled, Destination: zone.Hand})
		return err
	}}
}

// priestLeaveProgram has the exiled card's owner create an X/X white Spirit,
// X the exiled cards' total mana value (CR 607.3).
func priestLeaveProgram() cardprogram.Program {
	return cardprogram.Program{Reference: priestLeaveTestReference, Resolve: func(ctx cardprogram.Context) error {
		exiled, err := ctx.LinkedExiledCards(linkedHandExileProducer)
		if err != nil || len(exiled) == 0 {
			return err
		}
		x := 0
		for _, card := range exiled {
			x += card.Info.ManaValue
		}
		spirit := &game.CardDef{CardFace: game.CardFace{
			Name: "Spirit", Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Spirit},
			Colors: []color.Color{color.White},
			Power:  opt.Val(game.PT{Value: x}), Toughness: opt.Val(game.PT{Value: x}),
		}}
		_, err = ctx.CreateTokens(spirit, cardprogram.TokenRecipient{Player: exiled[0].Owner, Count: 1})
		return err
	}}
}

// linkedHandExileCreature is a {B} creature whose enters trigger targets an
// opponent and runs enter, declaring the linked producer, with an optional
// leaves-the-battlefield trigger running leave.
func linkedHandExileCreature(name string, enter, leave game.ProgramReference) *game.CardDef {
	enterContent := game.Mode{Targets: []game.TargetSpec{targetOpponentSpec()}, Program: enter}.Ability()
	enterContent.LinkedProducers = []game.LinkedProducerKey{linkedHandExileProducer}
	triggers := []game.TriggeredAbility{{Trigger: game.TriggerCondition{Pattern: game.TriggerPattern{
		Event: game.EventPermanentEnteredBattlefield, Source: game.TriggerSourceSelf,
	}}, Content: enterContent}}
	if leave.ID != "" {
		triggers = append(triggers, game.TriggeredAbility{Trigger: game.TriggerCondition{Pattern: game.TriggerPattern{
			Event: game.EventZoneChanged, Source: game.TriggerSourceSelf, MatchFromZone: true, FromZone: zone.Battlefield,
		}}, Content: game.Mode{Program: leave}.Ability()})
	}
	return &game.CardDef{CardFace: game.CardFace{
		Name: name, ManaCost: opt.Val(cost.Mana{cost.B}),
		Types: []types.Card{types.Creature}, Power: opt.Val(game.PT{Value: 2}), Toughness: opt.Val(game.PT{Value: 2}),
		TriggeredAbilities: triggers,
	}}
}

// castLinkedHandExileCreature casts def for Player1 and leaves its enters
// trigger, targeting Player2, on the stack.
func castLinkedHandExileCreature(t *testing.T, e *Engine, g *game.Game, def *game.CardDef) *game.Permanent {
	t.Helper()
	if issues := game.ValidateCardDef(def); len(issues) != 0 {
		t.Fatalf("test declaration: %+v", issues)
	}
	cardID := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.B, 1)
	if !e.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatalf("casting %s failed", def.Name)
	}
	e.resolveTopOfStack(g, &TurnLog{})
	source := permanentByCardID(g, cardID)
	if source == nil || !e.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 1 {
		t.Fatalf("%s did not enter with exactly one enters trigger", def.Name)
	}
	trigger, _ := g.Stack.Peek()
	if !slices.Equal(trigger.Targets, []game.Target{game.PlayerTarget(game.Player2)}) {
		t.Fatalf("enters trigger targets %+v, want Player2", trigger.Targets)
	}
	return source
}

// chooseRevealedCard answers the revealed-hand choice with the named card, or
// with nothing when name is empty.
func chooseRevealedCard(t *testing.T, name string) [game.NumPlayers]PlayerAgent {
	return [game.NumPlayers]PlayerAgent{&controlScript{choice: func(_ PlayerObservation, req game.ChoiceRequest) []int {
		if name == "" {
			return []int{}
		}
		index := slices.Index(optionCardNames(t, req.Options), name)
		if index < 0 {
			t.Fatalf("%s was not a revealed-hand option", name)
		}
		return []int{index}
	}}}
}

func linkedHandExileEngine(t *testing.T, programs ...cardprogram.Program) (*game.Game, *Engine) {
	t.Helper()
	g, _ := setupBestowMain(t)
	return g, revealHandEngine(t, programs...)
}

func requireCardInZone(t *testing.T, g *game.Game, cardID id.ID, want zone.Type) {
	t.Helper()
	if got, ok := cardZone(g, cardID); !ok || got != want {
		t.Fatalf("card zone = %v, want %v", got, want)
	}
}

// TestLinkedHandExileTidehollowSculler plays Tidehollow Sculler's text: the
// enters ability exiles the chosen nonland card into the source's linked set,
// and the leaves ability reads that card and returns it to its owner's hand as
// a new object. A clone branch whose exiled card left exile and came back has
// nothing for the leaves ability to refer to (CR 400.7, 607.2a).
func TestLinkedHandExileTidehollowSculler(t *testing.T) {
	t.Parallel()
	var read []cardprogram.FoundCard
	g, e := linkedHandExileEngine(t,
		revealHandExileEnterProgram(scullerEnterTestReference, nonlandSelection(), false,
			cardprogram.CardPlacement{PublishLinkedProducer: linkedHandExileProducer}),
		scullerLeaveProgram(&read))
	addCardToHand(g, game.Player2, revealHandLand())
	addCardToHand(g, game.Player2, evidenceCard("Lightning Bolt", 1))
	titan := addCardToHand(g, game.Player2, evidenceCard("Grave Titan", 6))
	sculler := castLinkedHandExileCreature(t, e, g,
		linkedHandExileCreature("Tidehollow Sculler", scullerEnterTestReference, scullerLeaveTestReference))

	e.resolveTopOfStackWithChoices(g, chooseRevealedCard(t, "Grave Titan"), &TurnLog{})
	requireCardInZone(t, g, titan, zone.Exile)
	exiledVersion := g.CardInstances[titan].ZoneVersion
	if len(g.UntilSourceLeavesExiles) != 0 {
		t.Fatal("a CR 607 linked exile created a CR 610.3 duration")
	}
	branch := g.Clone()

	movePermanentToZone(g, sculler, zone.Graveyard)
	if !g.Players[game.Player2].Exile.Contains(titan) || !e.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("the linked card returned without the leaves ability")
	}
	e.resolveTopOfStack(g, &TurnLog{})
	requireCardInZone(t, g, titan, zone.Hand)
	if g.CardInstances[titan].ZoneVersion == exiledVersion || !g.Stack.IsEmpty() {
		t.Fatal("the returned card is not a new object in hand")
	}
	if len(read) != 1 || read[0].Identity.CardID != titan || read[0].Owner != game.Player2 ||
		read[0].Info.Name != "Grave Titan" || read[0].Info.ManaValue != 6 || read[0].Zone != zone.Exile {
		t.Fatalf("the leaves ability read %+v, want Grave Titan owned by Player2 in exile", read)
	}

	// The clone kept the exile and its link independently.
	requireCardInZone(t, branch, titan, zone.Exile)
	branchSculler := permanentByCardID(branch, sculler.CardInstanceID)
	if !moveCardBetweenZonesInBatch(branch, game.Player2, titan, zone.Exile, zone.Graveyard, false, 0) ||
		!moveCardBetweenZonesInBatch(branch, game.Player2, titan, zone.Graveyard, zone.Exile, false, 0) {
		t.Fatal("the branch could not move the exiled card out and back")
	}
	read = nil
	movePermanentToZone(branch, branchSculler, zone.Graveyard)
	e.putTriggeredAbilitiesOnStack(branch)
	e.resolveTopOfStack(branch, &TurnLog{})
	requireCardInZone(t, branch, titan, zone.Exile)
	if len(read) != 0 {
		t.Fatalf("the leaves ability referred to a new object %+v", read)
	}
}

// TestLinkedHandExileTidehollowLeavesBeforeEnterExilesIndefinitely proves the
// Sculler ruling: the leaves ability resolves first and does nothing, then the
// enters ability exiles the card with no CR 610.3 return.
func TestLinkedHandExileTidehollowLeavesBeforeEnterExilesIndefinitely(t *testing.T) {
	t.Parallel()
	var read []cardprogram.FoundCard
	g, e := linkedHandExileEngine(t,
		revealHandExileEnterProgram(scullerEnterTestReference, nonlandSelection(), false,
			cardprogram.CardPlacement{PublishLinkedProducer: linkedHandExileProducer}),
		scullerLeaveProgram(&read))
	bolt := addCardToHand(g, game.Player2, evidenceCard("Lightning Bolt", 1))
	sculler := castLinkedHandExileCreature(t, e, g,
		linkedHandExileCreature("Tidehollow Sculler", scullerEnterTestReference, scullerLeaveTestReference))

	movePermanentToZone(g, sculler, zone.Graveyard)
	if !e.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 2 {
		t.Fatal("the leaves ability did not go on the stack above the enters ability")
	}
	e.resolveTopOfStack(g, &TurnLog{})
	if len(read) != 0 {
		t.Fatalf("the early leaves ability read %+v", read)
	}
	requireCardInZone(t, g, bolt, zone.Hand)
	e.resolveTopOfStack(g, &TurnLog{})
	requireCardInZone(t, g, bolt, zone.Exile)
	e.applyStateBasedActions(g)
	if e.putTriggeredAbilitiesOnStack(g) || !g.Stack.IsEmpty() || len(g.UntilSourceLeavesExiles) != 0 {
		t.Fatal("the indefinite exile created a later return")
	}
	requireCardInZone(t, g, bolt, zone.Exile)
}

// TestLinkedHandExileSeverancePriest plays Severance Priest's text: the exiled
// card's owner creates an X/X white Spirit with X its mana value, and the card
// stays exiled. Declining the optional choice exiles nothing, so no token.
func TestLinkedHandExileSeverancePriest(t *testing.T) {
	t.Parallel()
	g, e := linkedHandExileEngine(t,
		revealHandExileEnterProgram(priestEnterTestReference, nonlandSelection(), true,
			cardprogram.CardPlacement{PublishLinkedProducer: linkedHandExileProducer}),
		priestLeaveProgram())
	addCardToHand(g, game.Player2, evidenceCard("Lightning Bolt", 1))
	titan := addCardToHand(g, game.Player2, evidenceCard("Grave Titan", 6))
	def := linkedHandExileCreature("Severance Priest", priestEnterTestReference, priestLeaveTestReference)
	priest := castLinkedHandExileCreature(t, e, g, def)
	declined := g.Clone()

	e.resolveTopOfStackWithChoices(g, chooseRevealedCard(t, "Grave Titan"), &TurnLog{})
	requireCardInZone(t, g, titan, zone.Exile)
	movePermanentToZone(g, priest, zone.Graveyard)
	e.putTriggeredAbilitiesOnStack(g)
	e.resolveTopOfStack(g, &TurnLog{})
	if namedTokenCount(g, game.Player2, "Spirit") != 1 || namedTokenCount(g, game.Player1, "Spirit") != 0 {
		t.Fatal("the exiled card's owner did not create the Spirit")
	}
	for _, permanent := range g.Battlefield {
		if permanent != nil && permanent.Token && effectivePower(g, permanent) != 6 {
			t.Fatalf("Spirit power = %d, want the exiled card's mana value 6", effectivePower(g, permanent))
		}
	}
	requireCardInZone(t, g, titan, zone.Exile)

	e.resolveTopOfStackWithChoices(declined, chooseRevealedCard(t, ""), &TurnLog{})
	requireCardInZone(t, declined, titan, zone.Hand)
	movePermanentToZone(declined, permanentByCardID(declined, priest.CardInstanceID), zone.Graveyard)
	e.putTriggeredAbilitiesOnStack(declined)
	e.resolveTopOfStack(declined, &TurnLog{})
	for player := range game.NumPlayers {
		if namedTokenCount(declined, game.PlayerID(player), "Spirit") != 0 {
			t.Fatal("a Spirit was created although no card was exiled")
		}
	}
}

// TestUntilSourceLeavesHandExileBrainMaggot plays Brain Maggot's text: the
// chosen card returns to its owner's hand as a new object immediately when
// Maggot leaves, using no stack object and before anyone gets priority. A
// clone keeps the exile and its duration independently.
func TestUntilSourceLeavesHandExileBrainMaggot(t *testing.T) {
	t.Parallel()
	g, e := linkedHandExileEngine(t,
		revealHandExileEnterProgram(maggotEnterTestReference, nonlandSelection(), false,
			cardprogram.CardPlacement{UntilSourceLeaves: true}))
	addCardToHand(g, game.Player2, revealHandLand())
	titan := addCardToHand(g, game.Player2, evidenceCard("Grave Titan", 6))
	maggot := castLinkedHandExileCreature(t, e, g,
		linkedHandExileCreature("Brain Maggot", maggotEnterTestReference, game.ProgramReference{}))

	e.resolveTopOfStack(g, &TurnLog{})
	requireCardInZone(t, g, titan, zone.Exile)
	exiledVersion := g.CardInstances[titan].ZoneVersion
	if len(g.UntilSourceLeavesExiles) != 1 || g.UntilSourceLeavesExiles[0].ReturnZone != zone.Hand ||
		g.UntilSourceLeavesExiles[0].SourceObjectID != maggot.ObjectID {
		t.Fatalf("until-source-leaves records = %+v, want one hand return for Brain Maggot", g.UntilSourceLeavesExiles)
	}
	branch := g.Clone()

	before := len(g.Events)
	resolveOpponentDestroy(t, e, g, game.Player2, maggot)
	if permanentByCardID(g, maggot.CardInstanceID) != nil {
		t.Fatal("Brain Maggot survived the destroy spell")
	}
	requireCardInZone(t, g, titan, zone.Hand)
	if g.CardInstances[titan].ZoneVersion == exiledVersion || !g.Stack.IsEmpty() ||
		e.putTriggeredAbilitiesOnStack(g) || len(g.UntilSourceLeavesExiles) != 0 {
		t.Fatal("the return was not an immediate one-shot effect producing a new object")
	}
	returned := false
	for _, event := range g.Events[before:] {
		if event.Kind == game.EventZoneChanged && event.CardID == titan &&
			event.FromZone == zone.Exile && event.ToZone == zone.Hand && event.Player == game.Player2 {
			returned = true
		}
	}
	if !returned {
		t.Fatal("no exile-to-hand zone change was recorded for the owner")
	}

	requireCardInZone(t, branch, titan, zone.Exile)
	if len(branch.UntilSourceLeavesExiles) != 1 {
		t.Fatal("the clone lost its until-source-leaves duration")
	}
}

// TestUntilSourceLeavesHandExileBrainMaggotsReturnSimultaneously proves CR
// 610.3d for hand returns: one destroy-all removing two Maggots returns both
// cards as one simultaneous event.
func TestUntilSourceLeavesHandExileBrainMaggotsReturnSimultaneously(t *testing.T) {
	t.Parallel()
	g, e := linkedHandExileEngine(t,
		revealHandExileEnterProgram(maggotEnterTestReference, nonlandSelection(), false,
			cardprogram.CardPlacement{UntilSourceLeaves: true}))
	bolt := addCardToHand(g, game.Player2, evidenceCard("Lightning Bolt", 1))
	titan := addCardToHand(g, game.Player2, evidenceCard("Grave Titan", 6))
	def := linkedHandExileCreature("Brain Maggot", maggotEnterTestReference, game.ProgramReference{})
	castLinkedHandExileCreature(t, e, g, def)
	e.resolveTopOfStack(g, &TurnLog{})
	castLinkedHandExileCreature(t, e, g, def)
	e.resolveTopOfStack(g, &TurnLog{})
	requireCardInZone(t, g, bolt, zone.Exile)
	requireCardInZone(t, g, titan, zone.Exile)

	before := len(g.Events)
	addEffectSpellToStack(g, game.Player4, game.Destroy{
		Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
	}, nil)
	e.resolveTopOfStack(g, &TurnLog{})
	requireCardInZone(t, g, bolt, zone.Hand)
	requireCardInZone(t, g, titan, zone.Hand)
	var simultaneous []id.ID
	for _, event := range g.Events[before:] {
		if event.Kind == game.EventZoneChanged && event.FromZone == zone.Exile && event.ToZone == zone.Hand {
			simultaneous = append(simultaneous, event.SimultaneousID)
		}
	}
	if len(simultaneous) != 2 || simultaneous[0] == 0 || simultaneous[0] != simultaneous[1] {
		t.Fatalf("return SimultaneousIDs = %v, want two returns sharing one simultaneous event", simultaneous)
	}
}

// TestUntilSourceLeavesHandExileKitesailFreebooterLeftBeforeResolution proves
// the Freebooter ruling: the hand is still revealed, but nothing is exiled.
func TestUntilSourceLeavesHandExileKitesailFreebooterLeftBeforeResolution(t *testing.T) {
	t.Parallel()
	g, e := linkedHandExileEngine(t,
		revealHandExileEnterProgram(kitesailEnterTestReference,
			game.Selection{ExcludedTypes: []types.Card{types.Land, types.Creature}}, false,
			cardprogram.CardPlacement{UntilSourceLeaves: true}))
	bolt := addCardToHand(g, game.Player2, evidenceCard("Lightning Bolt", 1))
	freebooter := castLinkedHandExileCreature(t, e, g,
		linkedHandExileCreature("Kitesail Freebooter", kitesailEnterTestReference, game.ProgramReference{}))

	movePermanentToZone(g, freebooter, zone.Graveyard)
	start := len(g.Events)
	e.resolveTopOfStack(g, &TurnLog{})
	if !slices.Equal(handRevealEvents(g, start), []id.ID{bolt}) {
		t.Fatal("the hand was not revealed")
	}
	requireCardInZone(t, g, bolt, zone.Hand)
	if len(g.UntilSourceLeavesExiles) != 0 {
		t.Fatal("a departed source recorded an until-source-leaves exile")
	}
}

// TestLinkedHandExilePlacementRejectsInvalidDurations proves the placement
// fails explicitly instead of approximating a duration it cannot honor.
func TestLinkedHandExilePlacementRejectsInvalidDurations(t *testing.T) {
	t.Parallel()
	spellUntil := game.ProgramReference{ID: "test/linked-hand-exile/spell-until", Version: "1"}
	spellLinked := game.ProgramReference{ID: "test/linked-hand-exile/spell-linked", Version: "1"}
	undeclared := game.ProgramReference{ID: "test/linked-hand-exile/undeclared", Version: "1"}
	toHand := game.ProgramReference{ID: "test/linked-hand-exile/until-to-hand", Version: "1"}
	tests := []struct {
		name    string
		program cardprogram.Program
		setup   func(*testing.T, *Engine, *game.Game)
		want    string
	}{
		{
			name: "until from a spell",
			program: revealHandExileEnterProgram(spellUntil, nonlandSelection(), false,
				cardprogram.CardPlacement{UntilSourceLeaves: true}),
			setup: func(_ *testing.T, _ *Engine, g *game.Game) {
				addImplementationSpellToStack(g, game.Player1, revealHandSpellDef("Spell", spellUntil),
					[]game.Target{game.PlayerTarget(game.Player2)})
			},
			want: "activated or triggered ability",
		},
		{
			name: "linked from a spell",
			program: revealHandExileEnterProgram(spellLinked, nonlandSelection(), false,
				cardprogram.CardPlacement{PublishLinkedProducer: linkedHandExileProducer}),
			setup: func(_ *testing.T, _ *Engine, g *game.Game) {
				def := revealHandSpellDef("Spell", spellLinked)
				def.SpellAbility.Val.LinkedProducers = []game.LinkedProducerKey{linkedHandExileProducer}
				addImplementationSpellToStack(g, game.Player1, def, []game.Target{game.PlayerTarget(game.Player2)})
			},
			want: "original battlefield source",
		},
		{
			name: "undeclared producer",
			program: revealHandExileEnterProgram(undeclared, nonlandSelection(), false,
				cardprogram.CardPlacement{PublishLinkedProducer: "test/linked-hand-exile/other"}),
			setup: func(t *testing.T, e *Engine, g *game.Game) {
				castLinkedHandExileCreature(t, e, g, linkedHandExileCreature("Creature", undeclared, game.ProgramReference{}))
			},
			want: "not declared",
		},
		{
			name: "until outside exile",
			program: cardprogram.Program{Reference: toHand, Resolve: func(ctx cardprogram.Context) error {
				revealed, err := ctx.RevealHand(cardprogram.HandReveal{Player: game.TargetPlayerReference(0), Max: 1})
				if err != nil {
					return err
				}
				_, err = ctx.MoveCards(cardprogram.CardPlacement{Cards: revealed.Chosen, Destination: zone.Graveyard, UntilSourceLeaves: true})
				return err
			}},
			setup: func(t *testing.T, e *Engine, g *game.Game) {
				castLinkedHandExileCreature(t, e, g, linkedHandExileCreature("Creature", toHand, game.ProgramReference{}))
			},
			want: "requires exile",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			g, e := linkedHandExileEngine(t, test.program)
			bolt := addCardToHand(g, game.Player2, evidenceCard("Lightning Bolt", 1))
			test.setup(t, e, g)
			func() {
				defer func() {
					failure, ok := recover().(UnsupportedError)
					if !ok || !strings.Contains(failure.Reason, test.want) {
						t.Fatalf("failure = %+v, want an unsupported error containing %q", failure, test.want)
					}
				}()
				e.resolveTopOfStack(g, &TurnLog{})
			}()
			requireCardInZone(t, g, bolt, zone.Hand)
		})
	}
}
