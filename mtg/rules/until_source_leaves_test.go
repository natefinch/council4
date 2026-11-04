package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: Banisher Priest M14 printing
// 0e518f9e-d74d-4a6f-8fde-5e3ec5834e63 (oracle
// 9f560b83-32d4-4bb4-a956-8f5db18599db); its Scryfall rulings returned four
// WotC rulings dated 2013-07-01. Banishing Light printing
// d53c7e49-745d-48ba-bc50-240472a4f039 (oracle
// f28b21a6-f7ce-437a-8c5b-0423cb55cefb) returned the same four rulings dated
// 2025-07-25: a source that leaves before its trigger resolves exiles nothing,
// an exiled token never returns, the returned card is a new object, and a
// returning Aura's owner chooses what it enchants or it stays exiled. Static
// Snare printing 1ce50932-03a6-48bc-8aee-bc8defd896cf (oracle
// b5b742d2-f19b-47c9-84bd-e85d975900e6) returned no rulings. Pacifism oracle
// 5f5e0b10-c8cf-450c-bfd3-bcb0528ec330 is the returning Aura. CR 2026-09-25
// from the verified official copy: 610.3 (the return is a second one-shot
// effect created immediately after the source leaves, so it never uses the
// stack), 610.3a-b (the object doesn't move if the source already left),
// 610.3c (it returns under its owner's control), 610.3d (returns caused by
// simultaneous departures are simultaneous), and 400.7 (the returned card and
// a source that left and returned are new objects). Battle at the Helvault
// printing 0de0eaeb-9b62-427d-9873-d84de1740ddc (oracle
// 076584c8-ec58-4711-8f59-4839a4fa88ad) returned three rulings dated
// 2023-07-28 and Vault 13: Dweller's Journey printing
// a22ad436-215a-4da1-b12e-1faac474f948 (oracle
// 586fb987-cd77-420f-978c-2ce140bcdcfb) two dated 2024-03-08: a Saga that
// leaves before its exiling chapter resolves exiles nothing, tokens never
// return, and chapter III's library-bottom cards are no longer in exile.

func untilLeavesRegistryDef(t *testing.T, name string) *game.CardDef {
	t.Helper()
	def := cards.NewDefaultRegistry().Lookup(name)
	if def == nil {
		t.Fatalf("%s is not in the default registry", name)
	}
	return def
}

// castUntilLeavesPermanent casts the named card for its controller with
// enough Plains, resolves it, and resolves its enters trigger.
func castUntilLeavesPermanent(t *testing.T, engine *Engine, g *game.Game, controller game.PlayerID, name string, plains int) *game.Permanent {
	t.Helper()
	cardID := addCardToHand(g, controller, untilLeavesRegistryDef(t, name))
	for range plains {
		addBasicLandPermanent(g, controller, types.Plains)
	}
	setMainPhasePriority(g, controller)
	if !engine.applyAction(g, controller, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatalf("casting %s failed", name)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	source := permanentByCardID(g, cardID)
	if source == nil {
		t.Fatalf("%s did not enter the battlefield", name)
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatalf("%s's enters trigger did not trigger", name)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d after %s's trigger resolved, want 0", g.Stack.Size(), name)
	}
	return source
}

func resolveOpponentDestroy(t *testing.T, engine *Engine, g *game.Game, controller game.PlayerID, victim *game.Permanent) {
	t.Helper()
	addEffectSpellToStack(g, controller, game.Destroy{Object: game.TargetPermanentReference(0)},
		[]game.Target{game.PermanentTarget(victim.ObjectID)})
	engine.resolveTopOfStack(g, &TurnLog{})
}

func requireExiled(t *testing.T, g *game.Game, cardID id.ID) {
	t.Helper()
	if permanentByCardID(g, cardID) != nil {
		t.Fatal("card is on the battlefield, want exiled")
	}
	if zoneOf, ok := cardZone(g, cardID); !ok || zoneOf != zone.Exile {
		t.Fatalf("card zone = %v, want exile", zoneOf)
	}
}

func requireReturned(t *testing.T, g *game.Game, cardID id.ID, previous id.ID, owner game.PlayerID) *game.Permanent {
	t.Helper()
	returned := permanentByCardID(g, cardID)
	if returned == nil {
		t.Fatal("exiled card did not return to the battlefield")
	}
	if returned.ObjectID == previous {
		t.Fatal("returned card kept its old object identity, want a new object (CR 400.7)")
	}
	if returned.Controller != owner {
		t.Fatalf("returned card controller = %v, want owner %v (CR 610.3c)", returned.Controller, owner)
	}
	return returned
}

// TestStaticSnareReturnsExiledCreatureImmediatelyWithoutStack proves the
// CR 610.3 return is immediate: as Static Snare is destroyed, the exiled
// creature is back before anyone gets priority and nothing uses the stack.
func TestStaticSnareReturnsExiledCreatureImmediatelyWithoutStack(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	victim := addCombatCreaturePermanent(g, game.Player2)
	victimCard, victimObject := victim.CardInstanceID, victim.ObjectID

	snare := castUntilLeavesPermanent(t, engine, g, game.Player1, "Static Snare", 5)
	requireExiled(t, g, victimCard)
	if len(g.UntilSourceLeavesExiles) != 1 || g.UntilSourceLeavesExiles[0].SourceObjectID != snare.ObjectID {
		t.Fatalf("until-source-leaves records = %+v, want one for Static Snare", g.UntilSourceLeavesExiles)
	}

	resolveOpponentDestroy(t, engine, g, game.Player2, snare)

	if permanentByCardID(g, snare.CardInstanceID) != nil {
		t.Fatal("Static Snare survived the destroy spell")
	}
	requireReturned(t, g, victimCard, victimObject, game.Player2)
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want the return to use no stack object", g.Stack.Size())
	}
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("Static Snare leaving put a triggered ability on the stack")
	}
	if len(g.UntilSourceLeavesExiles) != 0 {
		t.Fatalf("until-source-leaves records = %+v after the return, want none", g.UntilSourceLeavesExiles)
	}
}

// TestBanisherPriestLeavingBeforeTriggerResolvesExilesNothing proves
// CR 610.3a-b and the Banisher Priest ruling: removing the Priest in response
// to its trigger means the target is never exiled.
func TestBanisherPriestLeavingBeforeTriggerResolvesExilesNothing(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	victim := addCombatCreaturePermanent(g, game.Player2)
	cardID := addCardToHand(g, game.Player1, untilLeavesRegistryDef(t, "Banisher Priest"))
	for range 3 {
		addBasicLandPermanent(g, game.Player1, types.Plains)
	}
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatal("casting Banisher Priest failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	priest := permanentByCardID(g, cardID)
	if priest == nil || !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("Banisher Priest's enters trigger did not trigger")
	}

	resolveOpponentDestroy(t, engine, g, game.Player2, priest)
	if permanentByCardID(g, cardID) != nil {
		t.Fatal("Banisher Priest survived the destroy spell")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if permanentByCardID(g, victim.CardInstanceID) == nil {
		t.Fatal("target was exiled although Banisher Priest left before its trigger resolved")
	}
	if len(g.UntilSourceLeavesExiles) != 0 {
		t.Fatalf("until-source-leaves records = %+v, want none", g.UntilSourceLeavesExiles)
	}
}

// TestBanisherPriestsLeavingTogetherReturnCardsSimultaneously proves
// CR 610.3d: one destroy-all event removing two Banisher Priests returns both
// exiled cards as one simultaneous event, under their owners' control.
func TestBanisherPriestsLeavingTogetherReturnCardsSimultaneously(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	first := addCombatCreaturePermanent(g, game.Player2)
	second := addCombatCreaturePermanent(g, game.Player3)
	castUntilLeavesPermanent(t, engine, g, game.Player1, "Banisher Priest", 3)
	castUntilLeavesPermanent(t, engine, g, game.Player1, "Banisher Priest", 3)
	requireExiled(t, g, first.CardInstanceID)
	requireExiled(t, g, second.CardInstanceID)
	if len(g.UntilSourceLeavesExiles) != 2 {
		t.Fatalf("until-source-leaves records = %+v, want one per Priest", g.UntilSourceLeavesExiles)
	}

	before := len(g.Events)
	addEffectSpellToStack(g, game.Player4, game.Destroy{
		Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
	}, nil)
	engine.resolveTopOfStack(g, &TurnLog{})

	requireReturned(t, g, first.CardInstanceID, first.ObjectID, game.Player2)
	requireReturned(t, g, second.CardInstanceID, second.ObjectID, game.Player3)
	var simultaneous []id.ID
	for _, event := range g.Events[before:] {
		if event.Kind == game.EventZoneChanged && event.FromZone == zone.Exile && event.ToZone == zone.Battlefield {
			simultaneous = append(simultaneous, event.SimultaneousID)
		}
	}
	if len(simultaneous) != 2 || simultaneous[0] == 0 || simultaneous[0] != simultaneous[1] {
		t.Fatalf("return SimultaneousIDs = %v, want two returns sharing one simultaneous event", simultaneous)
	}
}

// TestUntilSourceLeavesReturnsDuringStateBasedActions proves the return
// follows a state-based death in the same check, before any trigger is put on
// the stack, and that the returned card's own enters trigger triggers.
func TestUntilSourceLeavesReturnsDuringStateBasedActions(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	victim := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:      "Returning Watcher",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{Type: game.TriggerWhen, Pattern: game.TriggerPattern{
				Event:  game.EventPermanentEnteredBattlefield,
				Source: game.TriggerSourceSelf,
			}},
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()},
			}}}.Ability(),
		}},
	}})
	priest := castUntilLeavesPermanent(t, engine, g, game.Player1, "Banisher Priest", 3)
	requireExiled(t, g, victim.CardInstanceID)

	priest.MarkedDamage = 2
	engine.applyStateBasedActions(g)

	if permanentByCardID(g, priest.CardInstanceID) != nil {
		t.Fatal("Banisher Priest with lethal damage survived state-based actions")
	}
	requireReturned(t, g, victim.CardInstanceID, victim.ObjectID, game.Player2)
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want no return stack object", g.Stack.Size())
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("returned creature's enters trigger did not trigger")
	}
	top, ok := g.Stack.Peek()
	if !ok || top.SourceCardID != victim.CardInstanceID || g.Stack.Size() != 1 {
		t.Fatalf("stack top = %+v (size %d), want only the returned creature's enters trigger", top, g.Stack.Size())
	}
}

// TestUntilSourceLeavesReturnsStolenPermanentToOwner proves CR 610.3c for a
// permanent controlled by a player other than its owner.
func TestUntilSourceLeavesReturnsStolenPermanentToOwner(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	victim := addCombatCreaturePermanent(g, game.Player2)
	victim.Controller = game.Player3
	priest := castUntilLeavesPermanent(t, engine, g, game.Player1, "Banisher Priest", 3)
	requireExiled(t, g, victim.CardInstanceID)

	resolveOpponentDestroy(t, engine, g, game.Player3, priest)

	requireReturned(t, g, victim.CardInstanceID, victim.ObjectID, game.Player2)
}

// TestUntilSourceLeavesExiledTokenDoesNotReturn proves the token ruling: the
// exiled token ceases to exist and nothing returns.
func TestUntilSourceLeavesExiledTokenDoesNotReturn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	token := &game.Permanent{
		ObjectID:   g.IDGen.Next(),
		Owner:      game.Player2,
		Controller: game.Player2,
		Token:      true,
		TokenDef:   &game.CardDef{CardFace: game.CardFace{Name: "Soldier", Types: []types.Card{types.Creature}}},
	}
	g.Battlefield = append(g.Battlefield, token)
	priest := castUntilLeavesPermanent(t, engine, g, game.Player1, "Banisher Priest", 3)
	if _, ok := permanentByObjectID(g, token.ObjectID); ok {
		t.Fatal("token was not exiled")
	}
	if len(g.UntilSourceLeavesExiles) != 0 {
		t.Fatalf("until-source-leaves records = %+v, want none for a token", g.UntilSourceLeavesExiles)
	}
	before := len(g.Battlefield)

	resolveOpponentDestroy(t, engine, g, game.Player2, priest)

	if len(g.Battlefield) != before-1 {
		t.Fatal("an exiled token returned to the battlefield")
	}
}

// TestUntilSourceLeavesCardThatLeftExileDoesNotReturn proves a card that left
// exile is a new object (CR 400.7), so the source leaving doesn't move it.
func TestUntilSourceLeavesCardThatLeftExileDoesNotReturn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	victim := addCombatCreaturePermanent(g, game.Player2)
	priest := castUntilLeavesPermanent(t, engine, g, game.Player1, "Banisher Priest", 3)
	if !moveCardBetweenZones(g, game.Player2, victim.CardInstanceID, zone.Exile, zone.Graveyard) ||
		!moveCardBetweenZones(g, game.Player2, victim.CardInstanceID, zone.Graveyard, zone.Exile) {
		t.Fatal("could not move the exiled card out of exile and back")
	}

	resolveOpponentDestroy(t, engine, g, game.Player2, priest)

	requireExiled(t, g, victim.CardInstanceID)
}

// TestUntilSourceLeavesPhasedOutSourceKeepsCardExiled proves a phased-out
// source has not left the battlefield (CR 702.26d), so nothing returns.
func TestUntilSourceLeavesPhasedOutSourceKeepsCardExiled(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	victim := addCombatCreaturePermanent(g, game.Player2)
	priest := castUntilLeavesPermanent(t, engine, g, game.Player1, "Banisher Priest", 3)
	priest.PhasedOut = true

	engine.applyStateBasedActions(g)

	requireExiled(t, g, victim.CardInstanceID)
	if len(g.UntilSourceLeavesExiles) != 1 {
		t.Fatalf("until-source-leaves records = %+v, want the record retained", g.UntilSourceLeavesExiles)
	}
}

// TestUntilSourceLeavesCloneIsIndependent proves the duration is game-owned
// state: a return in a clone leaves the original exiled with its record.
func TestUntilSourceLeavesCloneIsIndependent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	victim := addCombatCreaturePermanent(g, game.Player2)
	priest := castUntilLeavesPermanent(t, engine, g, game.Player1, "Banisher Priest", 3)

	clone := g.Clone()
	resolveOpponentDestroy(t, engine, clone, game.Player2, priest)

	requireReturned(t, clone, victim.CardInstanceID, victim.ObjectID, game.Player2)
	requireExiled(t, g, victim.CardInstanceID)
	if len(g.UntilSourceLeavesExiles) != 1 {
		t.Fatalf("original until-source-leaves records = %+v, want one", g.UntilSourceLeavesExiles)
	}
}

// TestUntilSourceLeavesGroupExileReturnsTogether covers the group route a
// later card ("exile ... each ... until this leaves the battlefield") uses:
// every exiled card returns in the same simultaneous event.
func TestUntilSourceLeavesGroupExileReturnsTogether(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Group Warden", Types: []types.Card{types.Enchantment},
	}})
	first := addCombatCreaturePermanent(g, game.Player2)
	second := addCombatCreaturePermanent(g, game.Player3)
	obj := linkedSourceObject(source)
	obj.Kind = game.StackTriggeredAbility
	resolveInstruction(engine, g, obj, game.MovePermanent{
		Group:             game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
		Destination:       zone.Exile,
		UntilSourceLeaves: true,
	}, nil)
	requireExiled(t, g, first.CardInstanceID)
	requireExiled(t, g, second.CardInstanceID)

	before := len(g.Events)
	resolveOpponentDestroy(t, engine, g, game.Player2, source)

	requireReturned(t, g, first.CardInstanceID, first.ObjectID, game.Player2)
	requireReturned(t, g, second.CardInstanceID, second.ObjectID, game.Player3)
	var simultaneous []id.ID
	for _, event := range g.Events[before:] {
		if event.Kind == game.EventZoneChanged && event.FromZone == zone.Exile && event.ToZone == zone.Battlefield {
			simultaneous = append(simultaneous, event.SimultaneousID)
		}
	}
	if len(simultaneous) != 2 || simultaneous[0] == 0 || simultaneous[0] != simultaneous[1] {
		t.Fatalf("return SimultaneousIDs = %v, want one simultaneous return", simultaneous)
	}
}

func addUntilLeavesPacifism(t *testing.T, g *game.Game, owner game.PlayerID, host *game.Permanent) *game.Permanent {
	t.Helper()
	aura := addCombatPermanent(g, owner, untilLeavesRegistryDef(t, "Pacifism"))
	aura.AttachedTo = opt.Val(host.ObjectID)
	return aura
}

// TestBanishingLightReturnsAuraAttachedByItsOwner proves the Aura ruling: the
// returning Aura is attached to a legal object rather than targeting.
func TestBanishingLightReturnsAuraAttachedByItsOwner(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	host := addCombatCreaturePermanent(g, game.Player1)
	aura := addUntilLeavesPacifism(t, g, game.Player2, host)
	light := castUntilLeavesPermanent(t, engine, g, game.Player1, "Banishing Light", 3)
	requireExiled(t, g, aura.CardInstanceID)

	resolveOpponentDestroy(t, engine, g, game.Player2, light)

	returned := requireReturned(t, g, aura.CardInstanceID, aura.ObjectID, game.Player2)
	if !returned.AttachedTo.Exists || returned.AttachedTo.Val != host.ObjectID {
		t.Fatalf("returned Pacifism attached to %v, want the only creature %v", returned.AttachedTo, host.ObjectID)
	}
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want the Aura to return without targeting or the stack", g.Stack.Size())
	}
}

// TestBanishingLightAuraWithoutLegalObjectStaysExiled proves the Aura ruling:
// with nothing it can enchant, the returning Aura remains in exile.
func TestBanishingLightAuraWithoutLegalObjectStaysExiled(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	host := addCombatCreaturePermanent(g, game.Player1)
	aura := addUntilLeavesPacifism(t, g, game.Player2, host)
	light := castUntilLeavesPermanent(t, engine, g, game.Player1, "Banishing Light", 3)
	requireExiled(t, g, aura.CardInstanceID)
	resolveOpponentDestroy(t, engine, g, game.Player2, host)

	resolveOpponentDestroy(t, engine, g, game.Player2, light)

	requireExiled(t, g, aura.CardInstanceID)
	if len(g.UntilSourceLeavesExiles) != 0 {
		t.Fatalf("until-source-leaves records = %+v, want the duration consumed", g.UntilSourceLeavesExiles)
	}
}

// TestUntilSourceLeavesReturnsBeforeTriggersAfterSacrifice covers an
// activated ability's exile and a departure outside resolution (a sacrifice
// such as a cost payment): the return happens at the next Engine checkpoint,
// before any triggered ability is put on the stack or a player gets priority.
func TestUntilSourceLeavesReturnsBeforeTriggersAfterSacrifice(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Sacrificial Warden", Types: []types.Card{types.Artifact},
	}})
	victim := addCombatCreaturePermanent(g, game.Player2)
	obj := linkedSourceObject(source)
	obj.Targets = []game.Target{game.PermanentTarget(victim.ObjectID)}
	resolveInstruction(engine, g, obj, game.MovePermanent{
		Object:            game.TargetPermanentReference(0),
		Destination:       zone.Exile,
		UntilSourceLeaves: true,
	}, nil)
	requireExiled(t, g, victim.CardInstanceID)

	if !sacrificePermanent(g, source) {
		t.Fatal("could not sacrifice the source")
	}
	requireExiled(t, g, victim.CardInstanceID)
	engine.putTriggeredAbilitiesOnStack(g)

	requireReturned(t, g, victim.CardInstanceID, victim.ObjectID, game.Player2)
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want no return stack object", g.Stack.Size())
	}
}

// untilLeavesChapterExile returns the production ForEachPlayer declaration of
// a Saga's first chapter ability.
func untilLeavesChapterExile(t *testing.T, def *game.CardDef) game.ForEachPlayer {
	t.Helper()
	exile, ok := def.ChapterAbilities[0].Content.Modes[0].Sequence[0].Primitive.(game.ForEachPlayer)
	if !ok || !exile.UntilSourceLeaves {
		t.Fatalf("%s chapter I = %T, want an until-source-leaves ForEachPlayer", def.Name, def.ChapterAbilities[0].Content.Modes[0].Sequence[0].Primitive)
	}
	return exile
}

func untilLeavesChapterObject(source *game.Permanent) *game.StackObject {
	obj := linkedSourceObject(source)
	obj.Kind = game.StackTriggeredAbility
	return obj
}

// TestBattleAtTheHelvaultCaptivesReturnTogetherWhenSagaLeaves proves the
// distributive exile route: each player's captive returns immediately and
// simultaneously, under its owner's control, when the Saga leaves.
func TestBattleAtTheHelvaultCaptivesReturnTogetherWhenSagaLeaves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	def := untilLeavesRegistryDef(t, "Battle at the Helvault")
	saga := addCombatPermanent(g, game.Player1, def)
	first := addCombatCreaturePermanent(g, game.Player2)
	second := addCombatCreaturePermanent(g, game.Player3)
	resolveInstruction(engine, g, untilLeavesChapterObject(saga), untilLeavesChapterExile(t, def), nil)
	requireExiled(t, g, first.CardInstanceID)
	requireExiled(t, g, second.CardInstanceID)

	before := len(g.Events)
	resolveOpponentDestroy(t, engine, g, game.Player2, saga)

	requireReturned(t, g, first.CardInstanceID, first.ObjectID, game.Player2)
	requireReturned(t, g, second.CardInstanceID, second.ObjectID, game.Player3)
	var simultaneous []id.ID
	for _, event := range g.Events[before:] {
		if event.Kind == game.EventZoneChanged && event.FromZone == zone.Exile && event.ToZone == zone.Battlefield {
			simultaneous = append(simultaneous, event.SimultaneousID)
		}
	}
	if len(simultaneous) != 2 || simultaneous[0] == 0 || simultaneous[0] != simultaneous[1] {
		t.Fatalf("return SimultaneousIDs = %v, want one simultaneous return", simultaneous)
	}
	if g.Stack.Size() != 0 || engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("the Saga leaving used the stack, want an immediate return")
	}
}

// TestVault13LeavingBeforeChapterOneExilesNothing proves the Vault 13 ruling
// on the distributive route (CR 610.3a-b).
func TestVault13LeavingBeforeChapterOneExilesNothing(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	def := untilLeavesRegistryDef(t, "Vault 13: Dweller's Journey")
	saga := addCombatPermanent(g, game.Player1, def)
	victim := addCombatCreaturePermanent(g, game.Player2)
	obj := untilLeavesChapterObject(saga)
	resolveOpponentDestroy(t, engine, g, game.Player2, saga)

	resolveInstruction(engine, g, obj, untilLeavesChapterExile(t, def), nil)

	if permanentByCardID(g, victim.CardInstanceID) == nil {
		t.Fatal("Vault 13 exiled a permanent after it left the battlefield")
	}
	if len(g.UntilSourceLeavesExiles) != 0 {
		t.Fatalf("until-source-leaves records = %+v, want none", g.UntilSourceLeavesExiles)
	}
}

// TestVault13ChapterThreeLibraryCardsDoNotReturn proves cards chapter III put
// on the bottom of a library are new objects that the Saga's later departure
// doesn't return, while chapter III's chosen returns stay put.
func TestVault13ChapterThreeLibraryCardsDoNotReturn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	def := untilLeavesRegistryDef(t, "Vault 13: Dweller's Journey")
	saga := addCombatPermanent(g, game.Player1, def)
	captives := []*game.Permanent{
		addCombatCreaturePermanent(g, game.Player2),
		addCombatCreaturePermanent(g, game.Player3),
		addCombatCreaturePermanent(g, game.Player4),
	}
	obj := untilLeavesChapterObject(saga)
	resolveInstruction(engine, g, obj, untilLeavesChapterExile(t, def), nil)
	for _, captive := range captives {
		requireExiled(t, g, captive.CardInstanceID)
	}
	chapterThree := def.ChapterAbilities[2].Content.Modes[0].Sequence[0].Primitive
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{0, 1}}}}
	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: chapterThree}, agents, &TurnLog{})
	var bottomed *game.Permanent
	onBattlefield := 0
	for _, captive := range captives {
		if permanentByCardID(g, captive.CardInstanceID) != nil {
			onBattlefield++
			continue
		}
		bottomed = captive
	}
	if onBattlefield != 2 || bottomed == nil || !g.Players[bottomed.Owner].Library.Contains(bottomed.CardInstanceID) {
		t.Fatalf("chapter III returned %d captives, want two returned and one on the library bottom", onBattlefield)
	}

	resolveOpponentDestroy(t, engine, g, game.Player2, saga)

	if !g.Players[bottomed.Owner].Library.Contains(bottomed.CardInstanceID) {
		t.Fatal("the library-bottom card left its library when the Saga left")
	}
	if len(g.UntilSourceLeavesExiles) != 0 {
		t.Fatalf("until-source-leaves records = %+v, want the duration consumed", g.UntilSourceLeavesExiles)
	}
}

// exileUntilLeavesWithTrigger resolves the named until-leaves card's enters
// trigger from an existing permanent against an explicit target.
func exileUntilLeavesWithTrigger(t *testing.T, engine *Engine, g *game.Game, source *game.Permanent, def *game.CardDef, victim *game.Permanent) {
	t.Helper()
	obj := linkedSourceObject(source)
	obj.Kind = game.StackTriggeredAbility
	obj.Targets = []game.Target{game.PermanentTarget(victim.ObjectID)}
	engine.resolveInstructionSequence(g, obj, def.TriggeredAbilities[0].Content.Modes[0].Sequence, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	requireExiled(t, g, victim.CardInstanceID)
	g.TriggerEventCursor = len(g.Events)
}

// requireReturnBeforeTokens reports the exiled card's return entry index and
// requires it before every named token's entry in the given events.
func requireReturnBeforeTokens(t *testing.T, events []game.Event, cardID id.ID, token string, tokens int) {
	t.Helper()
	returned, entered := -1, 0
	for index, event := range events {
		if event.Kind != game.EventPermanentEnteredBattlefield {
			continue
		}
		if event.CardID == cardID {
			returned = index
		}
		if event.TokenName == token {
			if returned < 0 {
				t.Fatalf("%s entered at event %d before the exiled card returned", token, index)
			}
			entered++
		}
	}
	if returned < 0 || entered != tokens {
		t.Fatalf("return entry index = %d, %s entries = %d, want a return before %d tokens", returned, token, entered, tokens)
	}
}

// requireSoulWardenTriggers puts pending triggers on the stack and requires
// exactly want Soul Warden triggers controlled by the returned Warden.
func requireSoulWardenTriggers(t *testing.T, engine *Engine, g *game.Game, warden id.ID, want int) {
	t.Helper()
	engine.putTriggeredAbilitiesOnStack(g)
	got := 0
	for _, obj := range g.Stack.Objects() {
		if obj.SourceCardID == warden {
			got++
		}
	}
	if got != want || g.Stack.Size() != want {
		t.Fatalf("Soul Warden triggers = %d (stack size %d), want %d: the returned Warden must see every later token", got, g.Stack.Size(), want)
	}
}

// TestGenerousGiftReturnsBanishedCardBeforeCreatingElephant proves the
// authored-program destruction route: registered Generous Gift destroys
// Banisher Priest, and the exiled Soul Warden returns immediately after that
// destruction (CR 610.3), before Gift creates the Elephant, so Warden sees it.
func TestGenerousGiftReturnsBanishedCardBeforeCreatingElephant(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngineWithPrograms(nil, registry.Programs())
	priestDef := untilLeavesRegistryDef(t, "Banisher Priest")
	warden := addCombatPermanent(g, game.Player2, untilLeavesRegistryDef(t, "Soul Warden"))
	priest := addCombatPermanent(g, game.Player1, priestDef)
	exileUntilLeavesWithTrigger(t, engine, g, priest, priestDef, warden)

	spell := addCardToHand(g, game.Player2, untilLeavesRegistryDef(t, "Generous Gift"))
	g.Players[game.Player2].ManaPool.Add(mana.W, 1)
	g.Players[game.Player2].ManaPool.Add(mana.C, 2)
	setMainPhasePriority(g, game.Player2)
	if !engine.applyAction(g, game.Player2, action.CastSpell(spell, []game.Target{game.PermanentTarget(priest.ObjectID)}, 0, nil)) {
		t.Fatal("casting Generous Gift failed")
	}
	before := len(g.Events)
	engine.resolveTopOfStack(g, &TurnLog{})

	if !g.Players[game.Player1].Graveyard.Contains(priest.CardInstanceID) {
		t.Fatal("Generous Gift did not destroy Banisher Priest")
	}
	returned := requireReturned(t, g, warden.CardInstanceID, warden.ObjectID, game.Player2)
	requireReturnBeforeTokens(t, g.Events[before:], warden.CardInstanceID, "Elephant Token", 1)
	if namedTokenCount(g, game.Player1, "Elephant Token") != 1 {
		t.Fatal("Generous Gift did not give the Priest's controller an Elephant")
	}
	requireSoulWardenTriggers(t, engine, g, returned.CardInstanceID, 1)
}

// TestTerastodonReturnsBanishedCardBeforeCreatingElephants proves the same
// rule for an authored batch destruction: Terastodon destroys Banishing Light
// together with another permanent, and the exiled Soul Warden returns before
// either Elephant is created.
func TestTerastodonReturnsBanishedCardBeforeCreatingElephants(t *testing.T) {
	g, engine, targets := terastodonStack(t, 1)
	lightDef := untilLeavesRegistryDef(t, "Banishing Light")
	light := addCombatPermanent(g, game.Player2, lightDef)
	warden := addCombatPermanent(g, game.Player1, untilLeavesRegistryDef(t, "Soul Warden"))
	exileUntilLeavesWithTrigger(t, engine, g, light, lightDef, warden)
	trigger, _ := g.Stack.Peek()
	trigger.Targets = append(trigger.Targets, game.PermanentTarget(light.ObjectID))
	trigger.TargetCounts = []int{2}

	before := len(g.Events)
	engine.resolveTopOfStack(g, &TurnLog{})

	if !g.Players[game.Player2].Graveyard.Contains(light.CardInstanceID) ||
		!g.Players[game.Player2].Graveyard.Contains(targets[0].CardInstanceID) {
		t.Fatal("Terastodon did not destroy both chosen permanents")
	}
	returned := requireReturned(t, g, warden.CardInstanceID, warden.ObjectID, game.Player1)
	requireReturnBeforeTokens(t, g.Events[before:], warden.CardInstanceID, "Elephant", 2)
	requireSoulWardenTriggers(t, engine, g, returned.CardInstanceID, 2)
}
