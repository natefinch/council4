package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const worldgorgerID = "a628186d-b7d9-40a5-9ae2-fbc9d2a14c7c"
const deathmantleID = "66d41377-626d-4ae6-ba86-17bf0c8b3362"

func lifetimeGame(t *testing.T) (*game.Game, *Engine, *cards.Registry) {
	t.Helper()
	g, _ := setupBestowMain(t)
	registry := cards.NewDefaultRegistry()
	return g, NewEngineWithPrograms(nil, registry.Programs()), registry
}

func lifetimeDefinition(t *testing.T, registry *cards.Registry, oracleID string) *game.CardDef {
	t.Helper()
	def, err := registry.LookupOracleID(oracleID)
	if err != nil {
		t.Fatal(err)
	}
	if issues := game.ValidateCardDef(def); len(issues) != 0 {
		t.Fatalf("production declaration: %+v", issues)
	}
	return def
}

func castWorldgorger(t *testing.T, g *game.Game, e *Engine, registry *cards.Registry) *game.Permanent {
	t.Helper()
	card := addCardToHand(g, game.Player1, lifetimeDefinition(t, registry, worldgorgerID))
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	g.Players[game.Player1].ManaPool.Add(mana.R, 3)
	if !e.applyAction(g, game.Player1, action.CastSpell(card, nil, 0, nil)) {
		t.Fatal("production Worldgorger could not be cast")
	}
	e.resolveTopOfStack(g, &TurnLog{})
	source := permanentByCardID(g, card)
	if source == nil || !hasKeyword(g, source, game.Flying) || !hasKeyword(g, source, game.Trample) {
		t.Fatal("production Dragon did not enter with flying and trample")
	}
	if !e.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 1 {
		t.Fatal("actual entry did not queue exactly one exile trigger")
	}
	return source
}

func TestWorldgorgerProductionSimultaneousExileOwnerReturnAndClone(t *testing.T) {
	g, e, registry := lifetimeGame(t)
	stolen := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	stolen.Controller = game.Player1
	artifact := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Relic", Types: []types.Card{types.Artifact}}})
	opponent := addCombatCreaturePermanentWithPower(g, game.Player3, 2)
	createTokenPermanentsWithChoices(e, g, game.Player1, vanillaCreature("Token", 2, 2), 1, false, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	source := castWorldgorger(t, g, e, registry)
	before := g.Clone()
	e.resolveTopOfStack(g, &TurnLog{})
	if permanentByCardID(before, artifact.CardInstanceID) == nil || len(before.LinkedObjects) != 0 ||
		!g.Players[game.Player2].Exile.Contains(stolen.CardInstanceID) || !g.Players[game.Player1].Exile.Contains(artifact.CardInstanceID) ||
		permanentByCardID(g, opponent.CardInstanceID) == nil || permanentByCardID(g, source.CardInstanceID) == nil {
		t.Fatal("batch selection, ownership or pre-publication clone isolation failed")
	}
	var batch game.ObjectID
	for _, event := range g.Events {
		if event.Kind == game.EventZoneChanged && event.ToZone == zone.Exile {
			if event.SimultaneousID == 0 || batch != 0 && batch != event.SimultaneousID {
				t.Fatal("exiles were not one simultaneous event")
			}
			batch = event.SimultaneousID
		}
	}
	e.applyStateBasedActions(g)
	movePermanentToZone(g, source, zone.Hand)
	if !e.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("actual departure did not queue linked return")
	}
	after := g.Clone()
	branch := e.Simulator().ResolvePriority(after, [game.NumPlayers]PlayerAgent{})
	for _, captive := range []*game.Permanent{stolen, artifact} {
		returned := permanentByCardID(branch, captive.CardInstanceID)
		if returned == nil || returned.ObjectID == captive.ObjectID || returned.Controller != captive.Owner {
			t.Fatal("linked return lost exact card, fresh incarnation or owner control")
		}
		if !g.Players[captive.Owner].Exile.Contains(captive.CardInstanceID) {
			t.Fatal("Simulator aliased published links")
		}
	}
	if namedTokenCount(branch, game.Player1, "Token") != 0 || !branch.Stack.IsEmpty() {
		t.Fatal("ceased token returned or priority did not drain")
	}
	e.runPriorityLoop(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if permanentByCardID(g, stolen.CardInstanceID) == nil {
		t.Fatal("same catalog failed live resolution")
	}
}

func TestWorldgorgerLeavesBeforeEnterIsNotAnUntilEffect(t *testing.T) {
	g, e, registry := lifetimeGame(t)
	captive := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	source := castWorldgorger(t, g, e, registry)
	movePermanentToZone(g, source, zone.Graveyard)
	e.putTriggeredAbilitiesOnStack(g)
	e.resolveTopOfStack(g, &TurnLog{})
	if permanentByCardID(g, captive.CardInstanceID) == nil {
		t.Fatal("early leave trigger did not resolve against an empty link")
	}
	e.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player1].Exile.Contains(captive.CardInstanceID) || len(g.DelayedTriggers) != 0 || !g.Stack.IsEmpty() {
		t.Fatal("CR 607 exile was suppressed or automatically returned as CR 610.3")
	}
}

func TestWorldgorgerCopiesAccumulateAndReentryCannotReclaim(t *testing.T) {
	g, e, registry := lifetimeGame(t)
	first := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	source := castWorldgorger(t, g, e, registry)
	enter, _ := g.Stack.Peek()
	g.Stack.Push(game.NewStackObjectCopy(enter, g.IDGen.Next()))
	e.resolveTopOfStack(g, &TurnLog{})
	second := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	e.resolveTopOfStack(g, &TurnLog{})
	movePermanentToZone(g, source, zone.Hand)
	e.putTriggeredAbilitiesOnStack(g)
	oldReturn, _ := g.Stack.Pop()
	resolveInstruction(e, g, &game.StackObject{SourceCardID: source.CardInstanceID, Controller: game.Player1}, game.PutOnBattlefield{
		Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceSource}),
	}, nil)
	reentered := permanentByCardID(g, source.CardInstanceID)
	if reentered == nil || reentered.ObjectID == source.ObjectID {
		t.Fatal("Dragon reentry did not create fresh identity")
	}
	e.putTriggeredAbilitiesOnStack(g)
	third := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	e.resolveTopOfStack(g, &TurnLog{})
	movePermanentToZone(g, reentered, zone.Graveyard)
	e.putTriggeredAbilitiesOnStack(g)
	e.resolveTopOfStack(g, &TurnLog{})
	if permanentByCardID(g, third.CardInstanceID) == nil ||
		!g.Players[game.Player1].Exile.Contains(first.CardInstanceID) || !g.Players[game.Player1].Exile.Contains(second.CardInstanceID) {
		t.Fatal("new incarnation crossed linked namespaces")
	}
	g.Stack.Push(oldReturn)
	g.Stack.Push(game.NewStackObjectCopy(oldReturn, g.IDGen.Next()))
	e.resolveTopOfStack(g, &TurnLog{})
	e.resolveTopOfStack(g, &TurnLog{})
	for _, captive := range []*game.Permanent{first, second} {
		if countCardPermanentsControlledBy(g, game.Player1, captive.CardInstanceID) != 1 {
			t.Fatal("copied enter publications lost accumulation or copied return duplicated cards")
		}
	}
}

func TestWorldgorgerProductionRedirectedAndMovedCaptives(t *testing.T) {
	for _, scenario := range []string{"exile redirect", "return redirect", "moved exile"} {
		t.Run(scenario, func(t *testing.T) {
			g, e, registry := lifetimeGame(t)
			captive := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
			source := castWorldgorger(t, g, e, registry)
			if scenario == "exile redirect" {
				g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
					MatchEvent: game.EventZoneChanged, MatchFromZone: true, FromZone: zone.Battlefield,
					MatchToZone: true, ToZone: zone.Exile, ReplaceToZone: zone.Graveyard,
				})
			}
			e.resolveTopOfStack(g, &TurnLog{})
			if scenario == "return redirect" {
				g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
					MatchEvent: game.EventZoneChanged, MatchFromZone: true, FromZone: zone.Exile,
					MatchToZone: true, ToZone: zone.Battlefield, ReplaceToZone: zone.Graveyard,
				})
			}
			if scenario == "moved exile" {
				moveCardBetweenZones(g, game.Player1, captive.CardInstanceID, zone.Exile, zone.Hand)
				moveCardBetweenZones(g, game.Player1, captive.CardInstanceID, zone.Hand, zone.Exile)
			}
			movePermanentToZone(g, source, zone.Hand)
			e.runPriorityLoop(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			if permanentByCardID(g, captive.CardInstanceID) != nil {
				t.Fatal("returned replaced or later exile incarnation")
			}
			if scenario == "moved exile" && !g.Players[game.Player1].Exile.Contains(captive.CardInstanceID) ||
				scenario != "moved exile" && !g.Players[game.Player1].Graveyard.Contains(captive.CardInstanceID) {
				t.Fatal("actual destination was not honored")
			}
		})
	}
}

func TestWorldgorgerCopiedPermanentKeepsIndependentProducer(t *testing.T) {
	g, e, registry := lifetimeGame(t)
	first := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	source := castWorldgorger(t, g, e, registry)
	e.resolveTopOfStack(g, &TurnLog{})
	second := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	resolveInstruction(e, g, &game.StackObject{
		Controller: game.Player2, Targets: []game.Target{game.PermanentTarget(source.ObjectID)},
	}, game.CreateToken{
		Source: game.TokenCopyOf(game.TokenCopySpec{Source: game.TokenCopySourceObject, Object: game.TargetPermanentReference(0)}),
	}, nil)
	var copied *game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.Token {
			copied = permanent
		}
	}
	if copied == nil || !e.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("actual copy-token entry did not capture copied Worldgorger abilities")
	}
	e.resolveTopOfStack(g, &TurnLog{})
	movePermanentToZone(g, copied, zone.Graveyard)
	e.runPriorityLoop(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if permanentByCardID(g, second.CardInstanceID) == nil || !g.Players[game.Player1].Exile.Contains(first.CardInstanceID) {
		t.Fatal("copied permanent shared another instance's linked channel")
	}
	movePermanentToZone(g, source, zone.Graveyard)
	e.runPriorityLoop(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if permanentByCardID(g, first.CardInstanceID) == nil {
		t.Fatal("original producer lost its independently accumulated cards")
	}
}

type deathmantlePermanents struct {
	equipment *game.Permanent
	creature  *game.Permanent
}

func deathmantleDeath(t *testing.T) (*game.Game, *Engine, deathmantlePermanents) {
	t.Helper()
	g, e, registry := lifetimeGame(t)
	equipment := addCombatPermanent(g, game.Player1, lifetimeDefinition(t, registry, deathmantleID))
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	creature.Controller = game.Player2
	movePermanentToZone(g, creature, zone.Graveyard)
	if !e.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 1 {
		t.Fatal("owner's graveyard death did not trigger Equipment")
	}
	return g, e, deathmantlePermanents{equipment: equipment, creature: creature}
}

func TestNimDeathmantlePaymentReturnAttachmentAndStatic(t *testing.T) {
	for _, scenario := range []string{"paid", "declined", "unpayable", "redirected", "moved and returned", "equipment blink"} {
		t.Run(scenario, func(t *testing.T) {
			g, e, board := deathmantleDeath(t)
			equipment, creature := board.equipment, board.creature
			if scenario != "unpayable" {
				g.Players[game.Player1].ManaPool.Add(mana.C, 4)
			}
			if scenario == "redirected" {
				g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
					MatchEvent: game.EventZoneChanged, MatchFromZone: true, FromZone: zone.Graveyard,
					MatchToZone: true, ToZone: zone.Battlefield, ReplaceToZone: zone.Exile,
				})
			}
			if scenario == "moved and returned" {
				resolveInstruction(e, g, &game.StackObject{SourceCardID: creature.CardInstanceID, Controller: game.Player1}, game.PutOnBattlefield{
					Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceSource}),
				}, nil)
			}
			if scenario == "equipment blink" {
				movePermanentToZone(g, equipment, zone.Hand)
				resolveInstruction(e, g, &game.StackObject{SourceCardID: equipment.CardInstanceID, Controller: game.Player1}, game.PutOnBattlefield{
					Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceSource}),
				}, nil)
			}
			log := TurnLog{}
			e.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{
				game.Player1: optionalMayAgent{accept: scenario != "declined"},
			}, &log)
			currentEquipment := permanentByCardID(g, equipment.CardInstanceID)
			returned := permanentByCardID(g, creature.CardInstanceID)
			paid := scenario != "declined" && scenario != "unpayable"
			if paid && g.Players[game.Player1].ManaPool.Total() != 0 ||
				scenario == "declined" && g.Players[game.Player1].ManaPool.Total() != 4 {
				t.Fatal("payment was inferred rather than paid/declined")
			}
			if scenario == "paid" || scenario == "equipment blink" {
				if returned == nil || returned.ObjectID == creature.ObjectID || returned.Controller != creature.Owner {
					t.Fatal("return did not use exact death card under owner control")
				}
			}
			if scenario != "paid" {
				if currentEquipment.AttachedTo.Exists {
					t.Fatal("failed return or Equipment blink attached to the wrong occurrence")
				}
				if (scenario == "declined" || scenario == "unpayable") && !g.Players[creature.Owner].Graveyard.Contains(creature.CardInstanceID) ||
					scenario == "redirected" && !g.Players[creature.Owner].Exile.Contains(creature.CardInstanceID) {
					t.Fatal("failed payment or redirected return moved to the wrong actual zone")
				}
				return
			}
			if currentEquipment.AttachedTo != opt.Val(returned.ObjectID) || effectivePower(g, returned) != 4 ||
				!hasKeyword(g, returned, game.Intimidate) || !permanentHasSubtype(g, returned, types.Zombie) {
				t.Fatal("return attachment or printed static abilities failed")
			}
			values := effectivePermanentValues(g, returned)
			if len(values.colors) != 1 || values.colors[0] != color.Black {
				t.Fatal("Equipment did not set black")
			}
			if len(log.Choices) == 0 || log.Choices[0].UsedFallback {
				t.Fatal("payment was not a real agent choice")
			}
		})
	}
}

func TestNimDeathmantlePendingCloneChoicesAndIndependentInstances(t *testing.T) {
	g, e, board := deathmantleDeath(t)
	equipment, creature := board.equipment, board.creature
	registry := cards.NewDefaultRegistry()
	secondEquipment := addCombatPermanent(g, game.Player2, lifetimeDefinition(t, registry, deathmantleID))
	secondCreature := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	secondCreature.Controller = game.Player1
	movePermanentToZone(g, secondCreature, zone.Graveyard)
	e.putTriggeredAbilitiesOnStack(g)
	g.Players[game.Player1].ManaPool.Add(mana.C, 4)
	g.Players[game.Player2].ManaPool.Add(mana.C, 4)
	branch := e.Simulator().ResolvePriority(g, [game.NumPlayers]PlayerAgent{
		game.Player1: optionalMayAgent{accept: true}, game.Player2: optionalMayAgent{accept: true},
	})
	e.runPriorityLoop(g, [game.NumPlayers]PlayerAgent{
		game.Player1: optionalMayAgent{accept: false}, game.Player2: optionalMayAgent{accept: false},
	}, &TurnLog{})
	for _, pair := range [][2]*game.Permanent{{equipment, creature}, {secondEquipment, secondCreature}} {
		returned := permanentByCardID(branch, pair[1].CardInstanceID)
		attached := permanentByCardID(branch, pair[0].CardInstanceID)
		if returned == nil || returned.Controller != pair[1].Owner || attached.AttachedTo != opt.Val(returned.ObjectID) ||
			!g.Players[pair[1].Owner].Graveyard.Contains(pair[1].CardInstanceID) || pair[0].AttachedTo.Exists {
			t.Fatal("cloned pending programs shared payment choices, owners or instance attachments")
		}
	}
}

func TestNimDeathmantleSubtypeFamilyAndRepeatedDeaths(t *testing.T) {
	g, e, board := deathmantleDeath(t)
	equipment, creature := board.equipment, board.creature
	card, _ := g.GetCardInstance(creature.CardInstanceID)
	card.Def = &game.CardDef{CardFace: game.CardFace{
		Name: "Artifact Goblin", Types: []types.Card{types.Artifact, types.Creature}, Subtypes: []types.Sub{types.Equipment, types.Goblin},
		Colors: []color.Color{color.Red}, Power: opt.Val(game.PT{Value: 2}), Toughness: opt.Val(game.PT{Value: 2}),
	}}
	g.Players[game.Player1].ManaPool.Add(mana.C, 8)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: true}}
	e.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	firstReturn := permanentByCardID(g, creature.CardInstanceID)
	if !permanentHasSubtype(g, firstReturn, types.Equipment) || !permanentHasSubtype(g, firstReturn, types.Zombie) ||
		permanentHasSubtype(g, firstReturn, types.Goblin) {
		t.Fatal("Zombie setting removed a noncreature subtype or retained an old creature subtype")
	}
	movePermanentToZone(g, firstReturn, zone.Graveyard)
	e.putTriggeredAbilitiesOnStack(g)
	e.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	secondReturn := permanentByCardID(g, creature.CardInstanceID)
	if secondReturn == nil || secondReturn.ObjectID == firstReturn.ObjectID || equipment.AttachedTo != opt.Val(secondReturn.ObjectID) {
		t.Fatal("later death resolution reused an older arrival")
	}
}

func TestNimDeathmantleIgnoresTokensAndOtherOwners(t *testing.T) {
	g, e, registry := lifetimeGame(t)
	addCombatPermanent(g, game.Player1, lifetimeDefinition(t, registry, deathmantleID))
	other := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	other.Controller = game.Player1
	movePermanentToZone(g, other, zone.Graveyard)
	tokens, _ := createTokenPermanentsCollectingWithChoices(e, g, game.Player1, vanillaCreature("Token", 2, 2), 1, false, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	movePermanentToZone(g, tokens[0], zone.Graveyard)
	if e.putTriggeredAbilitiesOnStack(g) || !g.Stack.IsEmpty() {
		t.Fatal("Nim used departed controller rather than graveyard owner, or triggered on token")
	}
}

func TestAuthoredBattlefieldEntrySnapshotsDescribeActualArrivals(t *testing.T) {
	for _, redirected := range []bool{false, true} {
		g, e, board := deathmantleDeath(t)
		creature := board.creature
		obj, _ := g.Stack.Peek()
		if redirected {
			g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
				MatchEvent: game.EventZoneChanged, MatchFromZone: true, FromZone: zone.Graveyard,
				MatchToZone: true, ToZone: zone.Battlefield, ReplaceToZone: zone.Exile,
			})
		}
		ctx := &programContext{engine: e, game: g, obj: obj, log: &TurnLog{}, active: true}
		outcome, err := ctx.Execute(game.PutOnBattlefield{
			Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceEvent}),
		})
		if err != nil || !outcome.Available || outcome.Succeeded == redirected {
			t.Fatalf("actual entry outcome: %+v, %v", outcome, err)
		}
		if redirected {
			if len(outcome.Objects) != 0 || !g.Players[creature.Owner].Exile.Contains(creature.CardInstanceID) {
				t.Fatal("redirected return published a nonexistent arrival")
			}
			continue
		}
		returned := permanentByCardID(g, creature.CardInstanceID)
		card, _ := g.GetCardInstance(creature.CardInstanceID)
		if len(outcome.Objects) != 1 || outcome.Objects[0].ObjectID != returned.ObjectID ||
			len(outcome.Objects[0].ZoneCards) != 1 || outcome.Objects[0].ZoneCards[0].ZoneVersion != card.ZoneVersion ||
			outcome.Objects[0].CopiableDef != nil || outcome.Objects[0].TokenDef != nil {
			t.Fatal("Execute did not return actual incarnation values without live pointers")
		}
		outcome.Objects[0].Types[0] = types.Instant
		version := card.ZoneVersion
		outcome.Objects[0].ZoneCards[0].ZoneVersion++
		if !permanentHasType(g, returned, types.Creature) || card.ZoneVersion != version {
			t.Fatal("authored result aliased mutable game data")
		}
	}
}

func TestNimDeathmantleReturnSurvivesIllegalProtectedAttachment(t *testing.T) {
	for _, scenario := range []string{"artifacts", "everything", "effective red", "black creature not Equipment"} {
		t.Run(scenario, func(t *testing.T) {
			g, e, board := deathmantleDeath(t)
			equipment, creature := board.equipment, board.creature
			card, _ := g.GetCardInstance(creature.CardInstanceID)
			def := *card.Def
			protection := game.ProtectionKeyword{FromTypes: []types.Card{types.Artifact}}
			if scenario == "everything" {
				protection = game.ProtectionKeyword{Everything: true}
			}
			if scenario == "effective red" {
				protection = game.ProtectionKeyword{FromColors: []color.Color{color.Red}}
				g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
					ID: g.IDGen.Next(), Layer: game.LayerColor, AffectedObjectID: equipment.ObjectID, SetColors: []color.Color{color.Red},
				})
			}
			if scenario == "black creature not Equipment" {
				protection = game.ProtectionKeyword{FromColors: []color.Color{color.Black}}
			}
			def.StaticAbilities = []game.StaticAbility{{KeywordAbilities: []game.KeywordAbility{protection}}}
			card.Def = &def
			g.Players[game.Player1].ManaPool.Add(mana.C, 4)
			e.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: true}}, &TurnLog{})
			returned := permanentByCardID(g, creature.CardInstanceID)
			if returned == nil || returned.ObjectID == creature.ObjectID || g.Players[game.Player1].ManaPool.Total() != 0 {
				t.Fatal("illegal attachment undid a nontargeted successful return or payment")
			}
			if scenario == "black creature not Equipment" {
				if equipment.AttachedTo != opt.Val(returned.ObjectID) {
					t.Fatal("attachment protection used the creature's new black color instead of the colorless Equipment")
				}
				return
			}
			if equipment.AttachedTo.Exists || len(returned.Attachments) != 0 || effectivePower(g, returned) != 2 {
				t.Fatal("Equipment attached or granted bonuses through protection")
			}
		})
	}
}

func TestNimDeathmantleMultipleInstancesCopiedTriggersAndEquip(t *testing.T) {
	g, e, board := deathmantleDeath(t)
	first, creature := board.equipment, board.creature
	registry := cards.NewDefaultRegistry()
	second := addCombatPermanent(g, game.Player1, lifetimeDefinition(t, registry, deathmantleID))
	original, _ := g.Stack.Peek()
	g.Stack.Push(game.NewStackObjectCopy(original, g.IDGen.Next()))
	g.Players[game.Player1].ManaPool.Add(mana.C, 8)
	e.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: true}}, &TurnLog{})
	returned := permanentByCardID(g, creature.CardInstanceID)
	e.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: true}}, &TurnLog{})
	if first.AttachedTo != opt.Val(returned.ObjectID) || second.AttachedTo.Exists || g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("copied death trigger duplicated a card, crossed attachments or skipped payment")
	}
	// Equip is a normal inspectable, sorcery-speed paid activation.
	g.Players[game.Player1].ManaPool.Add(mana.C, 4)
	if !e.applyAction(g, game.Player1, action.ActivateAbility(second.ObjectID, 0, []game.Target{game.PermanentTarget(returned.ObjectID)}, 0)) {
		t.Fatal("printed Equip {4} could not activate")
	}
	e.resolveTopOfStack(g, &TurnLog{})
	if second.AttachedTo != opt.Val(returned.ObjectID) || g.Players[game.Player1].ManaPool.Total() != 0 || effectivePower(g, returned) != 6 {
		t.Fatal("Equip or multiple Equipment statics did not apply")
	}
}

func TestWorldgorgerMixedReanimationLoopRetainsNaturalTriggers(t *testing.T) {
	g, e, registry := lifetimeGame(t)
	dragonCard := addCardToGraveyard(g, game.Player1, lifetimeDefinition(t, registry, worldgorgerID))
	// A declarative reanimation fixture supplies an unconditional feedback
	// path; Worldgorger itself is the complete production registration.
	reanimator := &game.CardDef{CardFace: game.CardFace{
		Name: "Reanimation loop fixture", Types: []types.Card{types.Enchantment},
		TriggeredAbilities: []game.TriggeredAbility{
			{Trigger: game.TriggerCondition{Pattern: game.TriggerPattern{
				Event: game.EventPermanentEnteredBattlefield, Source: game.TriggerSourceSelf,
			}}, Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.MassReturnFromGraveyard{
				Player: game.ControllerReference(), Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
				Destination: zone.Battlefield,
			}}}}.Ability()},
			{Trigger: game.TriggerCondition{Pattern: game.TriggerPattern{
				Event: game.EventZoneChanged, Source: game.TriggerSourceSelf, MatchFromZone: true, FromZone: zone.Battlefield,
			}}, Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.Sacrifice{
				Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
			}}}}.Ability()},
		},
	}}
	fixtureCard := addCardToHand(g, game.Player1, reanimator)
	resolveInstruction(e, g, &game.StackObject{SourceCardID: fixtureCard, Controller: game.Player1}, game.PutOnBattlefield{
		Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceSource}),
	}, nil)
	e.putTriggeredAbilitiesOnStack(g)
	e.resolveTopOfStack(g, &TurnLog{})
	var previous id.ID
	// Stop at stable priority seams, not by modifying live card behavior or
	// claiming a rules draw. Each lap must create another incarnation and ETB.
	for range 3 {
		if !e.putTriggeredAbilitiesOnStack(g) {
			t.Fatal("Worldgorger loop lost its next trigger")
		}
		dragon := permanentByCardID(g, dragonCard)
		if dragon == nil || dragon.ObjectID == previous {
			t.Fatal("reanimation did not produce a new Dragon occurrence")
		}
		previous = dragon.ObjectID
		e.resolveTopOfStack(g, &TurnLog{}) // exile reanimator
		e.putTriggeredAbilitiesOnStack(g)
		e.resolveTopOfStack(g, &TurnLog{}) // sacrifice Dragon
		e.putTriggeredAbilitiesOnStack(g)
		e.resolveTopOfStack(g, &TurnLog{}) // return the reanimator
		e.putTriggeredAbilitiesOnStack(g)
		e.resolveTopOfStack(g, &TurnLog{}) // reanimate Dragon
		if g.IsGameOver() || permanentByCardID(g, fixtureCard) == nil || permanentByCardID(g, dragonCard) == nil {
			t.Fatal("natural loop was suppressed, capped, or declared a draw")
		}
	}
}
