package rules

// Source review: Engine gap #3440 (sub-issue of #3333) computes a spell's
// effective cascade as it becomes cast. Pinned corpus
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4 and Scryfall
// rulings retrieved for this unit. The test definitions below mirror these
// declarations; the cards themselves are admitted by a later card unit.
//   - Flamekin Herald 1b3d7411-7a25-4353-83c0-d09525a4fa08 {2}{R} Creature —
//     Elemental Bard 2/2: "Commander spells you cast have cascade." Rulings
//     2020-11-10: if you lose control of Flamekin Herald while casting, the
//     commander spell won't have cascade; once the spell has been cast, causing
//     Flamekin Herald to leave the battlefield won't stop its cascade ability.
//   - The First Sliver 55e4d4b2-1bd8-4c29-a21a-c014a974b715 {W}{U}{B}{R}{G}
//     Legendary Creature — Sliver 7/7: "Cascade" / "Sliver spells you cast have
//     cascade." Rulings 2019-06-14: casting a second The First Sliver gives it
//     two cascade instances, each triggering and resolving separately; the last
//     ability only applies while The First Sliver is on the battlefield, so a
//     Sliver cast by its own cascade won't have cascade.
//   - Quandrix, the Proof 2cbccc46-bdef-4dfb-90a4-0278c5c8488a {4}{G}{U}
//     Legendary Creature — Elder Dragon 6/6: "Flying, trample" / "Cascade" /
//     "Instant and sorcery spells you cast from your hand have cascade."
//   - Bloodbraid Marauder 25ca4b6c-2c4a-40f1-aee8-3bd373572eb7 {1}{R} Creature —
//     Human Berserker 3/1: "This creature can't block." / "Delirium — This spell
//     has cascade as long as there are four or more card types among cards in
//     your graveyard." Ruling 2021-06-18: delirium is checked at the moment you
//     finish casting it.
//   - The First Doctor 5611b70d-f8d1-4a8f-b9a2-a356264a7c46 {1}{W}{U} Legendary
//     Creature — Time Lord Doctor 2/4: "When The First Doctor enters, search your
//     library for an artifact card, reveal it, put it into your hand, then
//     shuffle." / "Whenever you cast a spell with cascade, put a +1/+1 counter on
//     target artifact or creature." Ruling 2023-10-13: its target is chosen
//     before cascade reveals anything. The test stand-in draws a card instead of
//     the targeted counter; only the trigger condition is under test here.
//   - Muscle Sliver {1}{G} Creature — Sliver: a Sliver spell with no cascade.
//   - Woodland Changeling {1}{G} Creature — Shapeshifter 2/2: "Changeling (This
//     card is every creature type.)" (default registry), and a Kindred
//     Changeling instant stand-in: Changeling spells are Sliver spells.
//   - Mistform Ultimus 7bf0e628-06b5-4e12-a526-a2bce08844f2 {3}{U} Legendary
//     Creature — Illusion 3/3: "Mistform Ultimus is every creature type (even if
//     this card isn't on the battlefield)." (default registry). Rulings
//     2006-09-25: it has every creature type, and keeps them even if an effect
//     removes its ability. A Mistform Ultimus spell is a Sliver spell.
//   - Adaptive Automaton 53c730c6-2f8c-4af8-b400-b9d573a71e60 {3} Artifact
//     Creature — Construct: "As this creature enters, choose a creature type." /
//     "This creature is the chosen type in addition to its other types."
//     (default registry). Ruling 2024-11-08: the type is chosen as it enters, so
//     as a spell it is only a Construct.
//   - Transguild Courier 86f94f8e-f45e-42fc-8ad6-e9bdfaaa997d {4} Artifact
//     Creature — Golem: "Transguild Courier is all colors." The color stand-ins
//     below print no colors (or a color plus Devoid) so only the CDA decides.
// Comprehensive Rules (verified copy effective 2026-09-25): 601.2i (a spell
// becomes cast once costs are paid; "whenever you cast" abilities trigger then),
// 604.3/604.3a (characteristic-defining abilities define colors, subtypes,
// power or toughness, are unconditional, and function in all zones), 702.73a
// (Changeling) and 702.114a (Devoid are such abilities),
// 603.2 (triggers check the event as it happens), 702.85a (cascade is a
// triggered ability that functions only while the spell is on the stack),
// 702.85c (each instance triggers separately), 707.10 (non-cast copies are not
// cast), 707.12 (a cast copy of a card is cast) and 903.3 (being a commander is a
// property of the card itself, which a copy does not have).

import (
	"testing"

	carda "github.com/natefinch/council4/mtg/cards/a"
	cardm "github.com/natefinch/council4/mtg/cards/m"
	cardw "github.com/natefinch/council4/mtg/cards/w"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func cascadeGrantStatic(text string, selection game.Selection, fromZone zone.Type) game.StaticAbility {
	return game.StaticAbility{Text: text, RuleEffects: []game.RuleEffect{{
		Kind:               game.RuleEffectGrantSpellKeyword,
		AffectedController: game.ControllerYou,
		CardSelection:      selection,
		GrantedKeyword:     game.Cascade,
		CastFromZone:       fromZone,
	}}}
}

func flamekinHeraldTestDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Flamekin Herald",
		ManaCost:  opt.Val(cost.Mana{cost.O(2), cost.R}),
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Elemental, types.Bard},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{
			cascadeGrantStatic("Commander spells you cast have cascade.", game.Selection{MatchCommander: true}, zone.None),
		},
	}}
}

func theFirstSliverTestDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:       "The First Sliver",
		ManaCost:   opt.Val(cost.Mana{cost.W, cost.U, cost.B, cost.R, cost.G}),
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		Subtypes:   []types.Sub{types.Sliver},
		Power:      opt.Val(game.PT{Value: 7}),
		Toughness:  opt.Val(game.PT{Value: 7}),
		StaticAbilities: []game.StaticAbility{
			game.CascadeStaticBody,
			cascadeGrantStatic("Sliver spells you cast have cascade.", game.Selection{SubtypesAny: []types.Sub{types.Sliver}}, zone.None),
		},
	}}
}

func quandrixTheProofTestDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:       "Quandrix, the Proof",
		ManaCost:   opt.Val(cost.Mana{cost.O(4), cost.G, cost.U}),
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		Subtypes:   []types.Sub{types.Elder, types.Dragon},
		Power:      opt.Val(game.PT{Value: 6}),
		Toughness:  opt.Val(game.PT{Value: 6}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
			game.TrampleStaticBody,
			game.CascadeStaticBody,
			cascadeGrantStatic("Instant and sorcery spells you cast from your hand have cascade.",
				game.Selection{RequiredTypesAny: []types.Card{types.Instant, types.Sorcery}}, zone.Hand),
		},
	}}
}

func bloodbraidMarauderTestDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Bloodbraid Marauder",
		ManaCost:  opt.Val(cost.Mana{cost.O(1), cost.R}),
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Human, types.Berserker},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{{
			Text:             "Delirium — This spell has cascade as long as there are four or more card types among cards in your graveyard.",
			ZoneOfFunction:   zone.Stack,
			KeywordAbilities: []game.KeywordAbility{game.SimpleKeyword{Kind: game.Cascade}},
			Condition: opt.Val(game.Condition{Aggregates: []game.AggregateComparison{{
				Aggregate: game.AggregateControllerGraveyardCardTypeCount,
				Op:        compare.GreaterOrEqual,
				Value:     4,
			}}}),
		}},
	}}
}

func theFirstDoctorTestDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:       "The First Doctor",
		ManaCost:   opt.Val(cost.Mana{cost.O(1), cost.W, cost.U}),
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		Power:      opt.Val(game.PT{Value: 2}),
		Toughness:  opt.Val(game.PT{Value: 4}),
		TriggeredAbilities: []game.TriggeredAbility{{
			Text: "Whenever you cast a spell with cascade, draw a card.",
			Trigger: game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event:           game.EventSpellCast,
					Controller:      game.TriggerControllerYou,
					CardSelection:   game.Selection{Keyword: game.Cascade},
					StackObjectKind: game.StackSpell,
				},
			},
			Content: game.AbilityContent{Modes: []game.Mode{{Sequence: []game.Instruction{{
				Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
			}}}}, MinModes: 1, MaxModes: 1},
		}},
	}}
}

func muscleSliverTestDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Muscle Sliver",
		ManaCost:  opt.Val(cost.Mana{cost.O(1), cost.G}),
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Sliver},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}}
}

func plainInstantTestDef(name string, manaValue int) *game.CardDef {
	return simpleGainLifeInstantWithManaValue(name, manaValue)
}

func newEffectiveCascadeGame() (*Engine, *game.Game) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMainPhasePriority(g, game.Player1)
	return NewEngine(nil), g
}

// castCountingCascade applies a cast action for playerID, puts the waiting
// triggers on the stack and returns the spell cast and how many cascade
// triggers it produced.
func castCountingCascade(t *testing.T, engine *Engine, g *game.Game, playerID game.PlayerID, act action.Action) (*game.StackObject, int) {
	t.Helper()
	before := len(cascadeTriggersOnStack(g))
	if !engine.applyAction(g, playerID, act) {
		t.Fatalf("cast %+v failed", act)
	}
	spell := lastSpellCastStackObject(t, g)
	engine.putTriggeredAbilitiesOnStack(g)
	return spell, len(cascadeTriggersOnStack(g)) - before
}

func lastSpellCastStackObject(t *testing.T, g *game.Game) *game.StackObject {
	t.Helper()
	for i := len(g.Events) - 1; i >= 0; i-- {
		if g.Events[i].Kind == game.EventSpellCast {
			obj, ok := stackObjectByID(g, g.Events[i].StackObjectID)
			if !ok {
				t.Fatal("last cast spell is not on the stack")
			}
			return obj
		}
	}
	t.Fatal("no spell was cast")
	return nil
}

// drainStack resolves every object on the stack so a later creature spell can
// be cast at sorcery speed.
func drainStack(engine *Engine, g *game.Game) {
	for !g.Stack.IsEmpty() {
		engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
		engine.putTriggeredAbilitiesOnStack(g)
	}
}

func addCommanderToCommandZone(g *game.Game, playerID game.PlayerID, def *game.CardDef) id.ID {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{ID: cardID, Def: def, Owner: playerID}
	g.CommanderIDs[cardID] = true
	g.Players[playerID].CommanderInstanceID = cardID
	g.Players[playerID].CommandZone.Add(cardID)
	return cardID
}

func cascadeTriggersFromSpell(g *game.Game, spell *game.StackObject) int {
	count := 0
	for _, obj := range cascadeTriggersOnStack(g) {
		if obj.TriggerEvent.StackObjectID == spell.ID {
			count++
		}
	}
	return count
}

func triggersFromSource(g *game.Game, sourceID id.ID) int {
	count := 0
	for _, obj := range g.Stack.Objects() {
		if obj.Kind == game.StackTriggeredAbility && obj.SourceID == sourceID {
			count++
		}
	}
	return count
}

func TestEffectiveCascadeTestDefinitionsValidate(t *testing.T) {
	for _, def := range []*game.CardDef{
		flamekinHeraldTestDef(), theFirstSliverTestDef(), quandrixTheProofTestDef(),
		bloodbraidMarauderTestDef(), theFirstDoctorTestDef(),
	} {
		if issues := game.ValidateCardDef(def); len(issues) != 0 {
			t.Fatalf("%s: %v", def.Name, issues)
		}
	}
}

// TestFlamekinHeraldGivesOwnCommanderSpellCascade proves "Commander spells you
// cast have cascade" applies to the controller's commander cast from the
// command zone, not to a noncommander spell or an opponent's commander, and that
// the Herald leaving after the cast does not remove the trigger.
func TestFlamekinHeraldGivesOwnCommanderSpellCascade(t *testing.T) {
	engine, g := newEffectiveCascadeGame()
	herald := addCombatPermanent(g, game.Player1, flamekinHeraldTestDef())
	addCardToLibrary(g, game.Player1, simpleGainLifeInstantWithManaValue("Hit", 1))
	commanderID := addCommanderToCommandZone(g, game.Player1, muscleSliverTestDef())
	plainID := addCardToHand(g, game.Player1, plainInstantTestDef("Plain Instant", 1))
	opponentCommanderID := addCommanderToCommandZone(g, game.Player2, muscleSliverTestDef())

	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	if _, got := castCountingCascade(t, engine, g, game.Player1, action.CastSpell(plainID, nil, 0, nil)); got != 0 {
		t.Fatalf("noncommander spell cascade triggers = %d, want 0", got)
	}
	drainStack(engine, g)

	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	if !engine.applyAction(g, game.Player1, action.CastSpellFromZone(commanderID, zone.Command, nil, 0, nil)) {
		t.Fatal("commander cast from the command zone failed")
	}
	if event := lastSpellCastEvent(t, g); event.CascadeInstances != opt.Val(1) {
		t.Fatalf("commander cast event cascade = %+v, want 1", event.CascadeInstances)
	}
	// The trigger was captured as the spell became cast; Flamekin Herald
	// leaving before it is put on the stack does not stop it.
	if _, ok := destroyPermanent(g, herald.ObjectID); !ok {
		t.Fatal("destroying Flamekin Herald failed")
	}
	engine.putTriggeredAbilitiesOnStack(g)
	commanderSpell := g.Stack.Objects()
	if got := len(cascadeTriggersOnStack(g)); got != 1 || commanderSpell[len(commanderSpell)-1].Kind != game.StackTriggeredAbility {
		t.Fatalf("commander cascade triggers = %d, want 1 after Herald left", got)
	}

	// An opponent's commander cast while Player1 controls a Herald gets nothing.
	engine2, g2 := newEffectiveCascadeGame()
	addCombatPermanent(g2, game.Player1, flamekinHeraldTestDef())
	opponentCommanderID = addCommanderToCommandZone(g2, game.Player2, muscleSliverTestDef())
	setMainPhasePriority(g2, game.Player2)
	g2.Players[game.Player2].ManaPool.Add(mana.G, 1)
	g2.Players[game.Player2].ManaPool.Add(mana.C, 1)
	if _, got := castCountingCascade(t, engine2, g2, game.Player2, action.CastSpellFromZone(opponentCommanderID, zone.Command, nil, 0, nil)); got != 0 {
		t.Fatalf("opponent commander cascade triggers = %d, want 0", got)
	}
}

// TestFlamekinHeraldNeedsCommanderCardNotCopy proves a cast copy of a commander
// (CR 707.12) is not itself a commander (CR 903.3): casting a copy of a prepared
// commander's spell face gains no cascade from Flamekin Herald.
func TestFlamekinHeraldNeedsCommanderCardNotCopy(t *testing.T) {
	engine, g := newEffectiveCascadeGame()
	addCombatPermanent(g, game.Player1, flamekinHeraldTestDef())
	def := vanillaCreature("Prepared Commander", 2, 2)
	def.Layout, def.Alternate = game.LayoutPrepare, opt.Val(plainInstantTestDef("Prepared Spell", 2).CardFace)
	commander := addCombatPermanent(g, game.Player1, def)
	commander.Prepared = true
	g.CommanderIDs[commander.CardInstanceID] = true
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	cast := action.CastSpellFaceFromZone(commander.CardInstanceID, zone.Battlefield, game.FaceAlternate, nil, 0, nil)
	spell, got := castCountingCascade(t, engine, g, game.Player1, cast)
	if !spell.Copy || got != 0 {
		t.Fatalf("cast copy of a commander: copy=%v cascade triggers=%d, want copy with 0", spell.Copy, got)
	}
}

// TestTheFirstSliverGrantsCascadeToSliverSpells covers The First Sliver's own
// printed cascade, the grant to another Sliver spell only while it is on the
// battlefield, no grant to a non-Sliver, two instances for a second First
// Sliver, and the Sliver it cascades into from the stack getting no cascade.
func TestTheFirstSliverGrantsCascadeToSliverSpells(t *testing.T) {
	engine, g := newEffectiveCascadeGame()
	muscleHit := addCardToLibrary(g, game.Player1, muscleSliverTestDef())
	firstID := addCardToHand(g, game.Player1, theFirstSliverTestDef())
	for _, c := range []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G} {
		g.Players[game.Player1].ManaPool.Add(c, 1)
	}
	if _, got := castCountingCascade(t, engine, g, game.Player1, action.CastSpell(firstID, nil, 0, nil)); got != 1 {
		t.Fatalf("The First Sliver cascade triggers = %d, want its printed cascade only", got)
	}
	// Its own cascade casts Muscle Sliver while The First Sliver is still on
	// the stack, so that Sliver spell has no cascade.
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	top, _ := g.Stack.Peek()
	if top == nil || top.SourceID != muscleHit {
		t.Fatalf("stack top = %+v, want cascaded Muscle Sliver", top)
	}
	if event := lastSpellCastEvent(t, g); event.CascadeInstances != opt.Val(0) {
		t.Fatalf("cascaded Muscle Sliver cascade = %+v, want 0", event.CascadeInstances)
	}
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("Muscle Sliver cast by The First Sliver's own cascade triggered cascade")
	}

	engine, g = newEffectiveCascadeGame()
	addCombatPermanent(g, game.Player1, theFirstSliverTestDef())
	sliverID := addCardToHand(g, game.Player1, muscleSliverTestDef())
	instantID := addCardToHand(g, game.Player1, plainInstantTestDef("Plain Instant", 2))
	secondFirstID := addCardToHand(g, game.Player1, theFirstSliverTestDef())
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	if _, got := castCountingCascade(t, engine, g, game.Player1, action.CastSpell(sliverID, nil, 0, nil)); got != 1 {
		t.Fatalf("Muscle Sliver with The First Sliver out: cascade triggers = %d, want 1", got)
	}
	drainStack(engine, g)
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	if _, got := castCountingCascade(t, engine, g, game.Player1, action.CastSpell(instantID, nil, 0, nil)); got != 0 {
		t.Fatalf("non-Sliver spell cascade triggers = %d, want 0", got)
	}
	drainStack(engine, g)
	for _, c := range []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G} {
		g.Players[game.Player1].ManaPool.Add(c, 1)
	}
	if spell, _ := castCountingCascade(t, engine, g, game.Player1, action.CastSpell(secondFirstID, nil, 0, nil)); cascadeTriggersFromSpell(g, spell) != 2 {
		t.Fatalf("second The First Sliver cascade triggers = %d, want 2", cascadeTriggersFromSpell(g, spell))
	}
}

// TestTheFirstSliverGrantChainsThroughFreeCast proves the grant reaches a Sliver
// cast without paying its mana cost by another cascade while The First Sliver is
// on the battlefield: the cascaded Sliver spell cascades in turn.
func TestTheFirstSliverGrantChainsThroughFreeCast(t *testing.T) {
	engine, g := newEffectiveCascadeGame()
	addCombatPermanent(g, game.Player1, theFirstSliverTestDef())
	secondHit := addCardToLibrary(g, game.Player1, simpleGainLifeInstantWithManaValue("Second Hit", 1))
	sliverHit := addCardToLibrary(g, game.Player1, muscleSliverTestDef())
	bigID := addCardToHand(g, game.Player1, cascadeSpell(4))
	g.Players[game.Player1].ManaPool.Add(mana.C, 4)
	if _, got := castCountingCascade(t, engine, g, game.Player1, action.CastSpell(bigID, nil, 0, nil)); got != 1 {
		t.Fatalf("printed cascade triggers = %d, want 1", got)
	}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if top, _ := g.Stack.Peek(); top == nil || top.SourceID != sliverHit {
		t.Fatalf("stack top = %+v, want cascaded Muscle Sliver", top)
	}
	resolveCascadeTrigger(t, engine, g, nil)
	if top, _ := g.Stack.Peek(); top == nil || top.SourceID != secondHit {
		t.Fatalf("stack top = %+v, want the granted cascade's hit", top)
	}
}

// TestTheFirstSliverGrantReachesKindredCastCopy proves a genuinely cast copy of
// a card (CR 707.12) is matched by its characteristics: a prepared Kindred
// Sliver spell face cast as a copy gains cascade from The First Sliver.
func TestTheFirstSliverGrantReachesKindredCastCopy(t *testing.T) {
	engine, g := newEffectiveCascadeGame()
	addCombatPermanent(g, game.Player1, theFirstSliverTestDef())
	face := plainInstantTestDef("Sliver Spell Face", 2).CardFace
	face.Types = []types.Card{types.Kindred, types.Instant}
	face.Subtypes = []types.Sub{types.Sliver}
	def := vanillaCreature("Prepared Sliver Source", 2, 2)
	def.Layout, def.Alternate = game.LayoutPrepare, opt.Val(face)
	permanent := addCombatPermanent(g, game.Player1, def)
	permanent.Prepared = true
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	cast := action.CastSpellFaceFromZone(permanent.CardInstanceID, zone.Battlefield, game.FaceAlternate, nil, 0, nil)
	if spell, got := castCountingCascade(t, engine, g, game.Player1, cast); !spell.Copy || got != 1 {
		t.Fatalf("cast Kindred Sliver copy: copy=%v cascade triggers=%d, want copy with 1", spell.Copy, got)
	}
}

// TestGrantedCascadeDoesNotReachNonCastStackCopy proves a non-cast copy of a
// spell that had granted cascade does not cascade (CR 707.10) and does not
// trigger "whenever you cast a spell with cascade".
func TestGrantedCascadeDoesNotReachNonCastStackCopy(t *testing.T) {
	engine, g := newEffectiveCascadeGame()
	addCombatPermanent(g, game.Player1, theFirstSliverTestDef())
	doctor := addCombatPermanent(g, game.Player1, theFirstDoctorTestDef())
	sliverID := addCardToHand(g, game.Player1, muscleSliverTestDef())
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	spell, got := castCountingCascade(t, engine, g, game.Player1, action.CastSpell(sliverID, nil, 0, nil))
	if got != 1 || triggersFromSource(g, doctor.ObjectID) != 1 {
		t.Fatalf("granted Sliver cast: cascade=%d doctor=%d, want 1 and 1", got, triggersFromSource(g, doctor.ObjectID))
	}
	beforeCascade, beforeDoctor := len(cascadeTriggersOnStack(g)), triggersFromSource(g, doctor.ObjectID)
	beforeCasts := spellCastEventCount(g)
	addEffectSpellToStack(g, game.Player1, game.CopyStackObject{
		Object: game.TargetStackObjectReference(0),
	}, []game.Target{game.StackObjectTarget(spell.ID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	if copied, _ := g.Stack.Peek(); copied == nil || !copied.Copy {
		t.Fatalf("stack top = %+v, want the non-cast copy", copied)
	}
	engine.putTriggeredAbilitiesOnStack(g)
	if spellCastEventCount(g) != beforeCasts || len(cascadeTriggersOnStack(g)) != beforeCascade ||
		triggersFromSource(g, doctor.ObjectID) != beforeDoctor {
		t.Fatal("a non-cast copy of a granted-cascade spell cascaded or triggered The First Doctor")
	}
}

// TestQuandrixGrantsCascadeOnlyToInstantsAndSorceriesFromHand proves the
// "from your hand" restriction and the instant-or-sorcery filter: a hand
// instant cascades, the same instant cast from the graveyard does not, a
// creature from hand does not, and an instant Quandrix's own cascade casts
// from exile does not.
func TestQuandrixGrantsCascadeOnlyToInstantsAndSorceriesFromHand(t *testing.T) {
	engine, g := newEffectiveCascadeGame()
	addCombatPermanent(g, game.Player1, quandrixTheProofTestDef())
	handInstant := addCardToHand(g, game.Player1, plainInstantTestDef("Hand Instant", 1))
	handCreature := addCardToHand(g, game.Player1, muscleSliverTestDef())
	graveyardDef := plainInstantTestDef("Graveyard Instant", 1)
	graveyardDef.StaticAbilities = append(graveyardDef.StaticAbilities, game.StaticAbility{
		Text:           "You may cast this card from your graveyard.",
		ZoneOfFunction: zone.Graveyard,
		RuleEffects: []game.RuleEffect{{
			Kind:           game.RuleEffectCastFromZone,
			AffectedPlayer: game.PlayerYou,
			CastFromZone:   zone.Graveyard,
			AffectedSource: true,
		}},
	})
	graveyardInstant := addCardToGraveyard(g, game.Player1, graveyardDef)

	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	if _, got := castCountingCascade(t, engine, g, game.Player1, action.CastSpell(handInstant, nil, 0, nil)); got != 1 {
		t.Fatalf("instant from hand cascade triggers = %d, want 1", got)
	}
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	if _, got := castCountingCascade(t, engine, g, game.Player1, action.CastSpellFromZone(graveyardInstant, zone.Graveyard, nil, 0, nil)); got != 0 {
		t.Fatalf("instant from graveyard cascade triggers = %d, want 0", got)
	}
	drainStack(engine, g)
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	if _, got := castCountingCascade(t, engine, g, game.Player1, action.CastSpell(handCreature, nil, 0, nil)); got != 0 {
		t.Fatalf("creature from hand cascade triggers = %d, want 0", got)
	}

	engine, g = newEffectiveCascadeGame()
	exileHit := addCardToLibrary(g, game.Player1, plainInstantTestDef("Exile Hit", 1))
	quandrixID := addCardToHand(g, game.Player1, quandrixTheProofTestDef())
	addCombatPermanent(g, game.Player1, quandrixTheProofTestDef())
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 4)
	if _, got := castCountingCascade(t, engine, g, game.Player1, action.CastSpell(quandrixID, nil, 0, nil)); got != 1 {
		t.Fatalf("Quandrix creature spell cascade triggers = %d, want its printed cascade only", got)
	}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if top, _ := g.Stack.Peek(); top == nil || top.SourceID != exileHit {
		t.Fatalf("stack top = %+v, want the cascaded instant", top)
	}
	if event := lastSpellCastEvent(t, g); event.FromZone != zone.Exile || event.CascadeInstances != opt.Val(0) {
		t.Fatalf("instant cast from exile by cascade: from=%v cascade=%+v, want exile and 0", event.FromZone, event.CascadeInstances)
	}
}

// TestBloodbraidMarauderDeliriumCascade proves the conditional stack static is
// read as the spell becomes cast: with four card types in the graveyard it
// cascades, with three it does not, and the result does not change afterwards.
func TestBloodbraidMarauderDeliriumCascade(t *testing.T) {
	setup := func(cardTypes []types.Card) (*Engine, *game.Game, id.ID) {
		engine, g := newEffectiveCascadeGame()
		for _, cardType := range cardTypes {
			addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: string(cardType) + " Card", Types: []types.Card{cardType}}})
		}
		marauderID := addCardToHand(g, game.Player1, bloodbraidMarauderTestDef())
		g.Players[game.Player1].ManaPool.Add(mana.R, 1)
		g.Players[game.Player1].ManaPool.Add(mana.C, 1)
		return engine, g, marauderID
	}

	engine, g, marauderID := setup([]types.Card{types.Artifact, types.Creature, types.Land})
	if _, got := castCountingCascade(t, engine, g, game.Player1, action.CastSpell(marauderID, nil, 0, nil)); got != 0 {
		t.Fatalf("Marauder without delirium cascade triggers = %d, want 0", got)
	}

	engine, g, marauderID = setup([]types.Card{types.Artifact, types.Creature, types.Land, types.Instant})
	doctor := addCombatPermanent(g, game.Player1, theFirstDoctorTestDef())
	if !engine.applyAction(g, game.Player1, action.CastSpell(marauderID, nil, 0, nil)) {
		t.Fatal("Marauder cast failed")
	}
	// Emptying the graveyard after the cast cannot take back cascade.
	for _, cardID := range g.Players[game.Player1].Graveyard.All() {
		g.Players[game.Player1].Graveyard.Remove(cardID)
	}
	engine.putTriggeredAbilitiesOnStack(g)
	if got := len(cascadeTriggersOnStack(g)); got != 1 || triggersFromSource(g, doctor.ObjectID) != 1 {
		t.Fatalf("Marauder with delirium: cascade=%d doctor=%d, want 1 and 1", got, triggersFromSource(g, doctor.ObjectID))
	}
}

// TestTheFirstDoctorTriggersOnSpellsWithCascade proves "Whenever you cast a
// spell with cascade" reads printed, granted and conditional cascade from the
// cast event, ignores spells without cascade and opponents' spells, and its
// trigger is put on the stack alongside the cascade trigger.
func TestTheFirstDoctorTriggersOnSpellsWithCascade(t *testing.T) {
	engine, g := newEffectiveCascadeGame()
	doctor := addCombatPermanent(g, game.Player1, theFirstDoctorTestDef())
	printedID := addCardToHand(g, game.Player1, cascadeSpell(2))
	plainID := addCardToHand(g, game.Player1, plainInstantTestDef("Plain Instant", 1))
	sliverID := addCardToHand(g, game.Player1, muscleSliverTestDef())

	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	if _, got := castCountingCascade(t, engine, g, game.Player1, action.CastSpell(printedID, nil, 0, nil)); got != 1 ||
		triggersFromSource(g, doctor.ObjectID) != 1 {
		t.Fatalf("printed cascade: cascade=%d doctor=%d, want 1 and 1", got, triggersFromSource(g, doctor.ObjectID))
	}
	drainStack(engine, g)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	castCountingCascade(t, engine, g, game.Player1, action.CastSpell(plainID, nil, 0, nil))
	if triggersFromSource(g, doctor.ObjectID) != 0 {
		t.Fatal("The First Doctor triggered on a spell without cascade")
	}
	drainStack(engine, g)
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	castCountingCascade(t, engine, g, game.Player1, action.CastSpell(sliverID, nil, 0, nil))
	if triggersFromSource(g, doctor.ObjectID) != 0 {
		t.Fatal("The First Doctor triggered on a Sliver without The First Sliver")
	}

	engine, g = newEffectiveCascadeGame()
	doctor = addCombatPermanent(g, game.Player1, theFirstDoctorTestDef())
	addCombatPermanent(g, game.Player1, theFirstSliverTestDef())
	sliverID = addCardToHand(g, game.Player1, muscleSliverTestDef())
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	castCountingCascade(t, engine, g, game.Player1, action.CastSpell(sliverID, nil, 0, nil))
	if triggersFromSource(g, doctor.ObjectID) != 1 {
		t.Fatal("The First Doctor did not trigger on a Sliver granted cascade")
	}
	clone := g.Clone()
	if event := lastSpellCastEvent(t, clone); event.CascadeInstances != opt.Val(1) {
		t.Fatalf("cloned cast event cascade = %+v, want 1", event.CascadeInstances)
	}

	engine, g = newEffectiveCascadeGame()
	addCombatPermanent(g, game.Player1, theFirstDoctorTestDef())
	opponentID := addCardToHand(g, game.Player2, cascadeSpell(2))
	setMainPhasePriority(g, game.Player2)
	g.Players[game.Player2].ManaPool.Add(mana.C, 2)
	castCountingCascade(t, engine, g, game.Player2, action.CastSpell(opponentID, nil, 0, nil))
	for _, obj := range g.Stack.Objects() {
		if obj.Kind == game.StackTriggeredAbility && obj.Controller == game.Player1 {
			t.Fatal("The First Doctor triggered on an opponent's cascade spell")
		}
	}
}

// TestFlamekinHeraldAndFirstSliverStackOnCommanderSliver proves independent
// grants each add an instance: a Sliver commander cast with both Flamekin
// Herald and The First Sliver out has two cascade triggers.
func TestFlamekinHeraldAndFirstSliverStackOnCommanderSliver(t *testing.T) {
	engine, g := newEffectiveCascadeGame()
	addCombatPermanent(g, game.Player1, flamekinHeraldTestDef())
	addCombatPermanent(g, game.Player1, theFirstSliverTestDef())
	commanderID := addCommanderToCommandZone(g, game.Player1, muscleSliverTestDef())
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	if _, got := castCountingCascade(t, engine, g, game.Player1, action.CastSpellFromZone(commanderID, zone.Command, nil, 0, nil)); got != 2 {
		t.Fatalf("Sliver commander cascade triggers = %d, want 2", got)
	}
}

// TestTheFirstSliverGrantsCascadeToChangelingSpells proves "Sliver spells you
// cast have cascade" reads a spell's effective subtypes: Changeling makes a
// spell every creature type (CR 702.73a), so the default-registry Woodland
// Changeling creature spell and a Kindred Changeling instant are Sliver spells
// that cascade and count as spells with cascade for The First Doctor, while a
// Shapeshifter without Changeling is not a Sliver spell.
func TestTheFirstSliverGrantsCascadeToChangelingSpells(t *testing.T) {
	kindredChangeling := plainInstantTestDef("Kindred Changeling Instant", 2)
	kindredChangeling.Types = []types.Card{types.Kindred, types.Instant}
	kindredChangeling.Subtypes = []types.Sub{types.Shapeshifter}
	kindredChangeling.StaticAbilities = append(kindredChangeling.StaticAbilities, game.ChangelingStaticBody)
	plainShapeshifter := vanillaCreature("Plain Shapeshifter", 2, 2)
	plainShapeshifter.Subtypes = []types.Sub{types.Shapeshifter}
	for _, tc := range []struct {
		def  *game.CardDef
		want int
	}{
		{cardw.WoodlandChangeling(), 1},
		{kindredChangeling, 1},
		{plainShapeshifter, 0},
	} {
		t.Run(tc.def.Name, func(t *testing.T) {
			engine, g := newEffectiveCascadeGame()
			addCombatPermanent(g, game.Player1, theFirstSliverTestDef())
			doctor := addCombatPermanent(g, game.Player1, theFirstDoctorTestDef())
			addCardToLibrary(g, game.Player1, simpleGainLifeInstantWithManaValue("Hit", 1))
			cardID := addCardToHand(g, game.Player1, tc.def)
			g.Players[game.Player1].ManaPool.Add(mana.G, 1)
			g.Players[game.Player1].ManaPool.Add(mana.C, 1)
			if _, got := castCountingCascade(t, engine, g, game.Player1, action.CastSpell(cardID, nil, 0, nil)); got != tc.want {
				t.Fatalf("%s cascade triggers = %d, want %d", tc.def.Name, got, tc.want)
			}
			if got := triggersFromSource(g, doctor.ObjectID); got != tc.want {
				t.Fatalf("%s The First Doctor triggers = %d, want %d", tc.def.Name, got, tc.want)
			}
		})
	}
}

// TestSpellKeywordGrantSubtypeSelectionSeesChangeling proves the payment
// planner and one-shot consumption routes read Changeling the same way: a
// Sliver-filtered improvise grant applies to and is consumed by a Changeling
// spell, and the shared CardDef is not mutated.
func TestSpellKeywordGrantSubtypeSelectionSeesChangeling(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.RuleEffects = append(g.RuleEffects, nextSpellImproviseGrant(g, game.Selection{SubtypesAny: []types.Sub{types.Sliver}}))
	changeling := cardw.WoodlandChangeling()
	if !spellGrantedKeyword(g, game.Player1, changeling, 0, zone.Hand, game.Improvise) {
		t.Fatal("Sliver-filtered improvise grant does not apply to a Changeling spell")
	}
	if len(changeling.Subtypes) != 1 || changeling.Subtypes[0] != types.Shapeshifter {
		t.Fatalf("Changeling card subtypes mutated: %v", changeling.Subtypes)
	}
	if spellGrantedKeyword(g, game.Player1, elfCreatureDef(), 0, zone.Hand, game.Improvise) {
		t.Fatal("Sliver-filtered improvise grant applies to a non-Sliver spell")
	}
	spell := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell, Controller: game.Player1, SourceTokenDef: changeling}
	consumeNextSpellKeywordGrantEffects(g, spell, zone.Hand)
	if len(g.RuleEffects) != 0 {
		t.Fatalf("rule effects = %#v, want the Sliver grant consumed by the Changeling spell", g.RuleEffects)
	}
}

// TestTheFirstSliverGrantsCascadeToCharacteristicDefiningSubtypes proves
// "Sliver spells you cast have cascade" reads subtypes from any
// characteristic-defining ability, not only Changeling (CR 604.3). The
// default-registry Mistform Ultimus is every creature type in every zone, so
// its spell cascades and is a spell with cascade for The First Doctor. A
// self-type effect keyed to an entry choice (the default-registry Adaptive
// Automaton) or gated by a condition (CR 604.3a(5)) is not a CDA, so neither
// spell is a Sliver spell.
func TestTheFirstSliverGrantsCascadeToCharacteristicDefiningSubtypes(t *testing.T) {
	for _, tc := range []struct {
		def  *game.CardDef
		pay  func(*mana.Pool)
		want int
	}{
		{cardm.MistformUltimus(), func(pool *mana.Pool) { pool.Add(mana.U, 1); pool.Add(mana.C, 3) }, 1},
		{carda.AdaptiveAutomaton(), func(pool *mana.Pool) { pool.Add(mana.C, 3) }, 0},
		{conditionalSliverTestDef(), func(pool *mana.Pool) { pool.Add(mana.C, 2) }, 0},
	} {
		t.Run(tc.def.Name, func(t *testing.T) {
			engine, g := newEffectiveCascadeGame()
			addCombatPermanent(g, game.Player1, theFirstSliverTestDef())
			doctor := addCombatPermanent(g, game.Player1, theFirstDoctorTestDef())
			addCardToLibrary(g, game.Player1, simpleGainLifeInstantWithManaValue("Hit", 1))
			cardID := addCardToHand(g, game.Player1, tc.def)
			tc.pay(&g.Players[game.Player1].ManaPool)
			if _, got := castCountingCascade(t, engine, g, game.Player1, action.CastSpell(cardID, nil, 0, nil)); got != tc.want {
				t.Fatalf("%s cascade triggers = %d, want %d", tc.def.Name, got, tc.want)
			}
			if got := triggersFromSource(g, doctor.ObjectID); got != tc.want {
				t.Fatalf("%s The First Doctor triggers = %d, want %d", tc.def.Name, got, tc.want)
			}
		})
	}
}

// TestColorFilteredCascadeGrantReadsColorCDAs proves a color-filtered spell
// grant reads colors as the spell's characteristic-defining abilities define
// them: a card that prints no colors but "is all colors" is a red spell, and a
// red card with Devoid is colorless (CR 604.3, 702.114a).
func TestColorFilteredCascadeGrantReadsColorCDAs(t *testing.T) {
	allColors := vanillaCreature("All-Colors Golem", 2, 2)
	allColors.ManaCost = opt.Val(cost.Mana{cost.O(2)})
	allColors.StaticAbilities = []game.StaticAbility{{
		Text: "This creature is all colors.",
		ContinuousEffects: []game.ContinuousEffect{{
			AffectedSource: true,
			Layer:          game.LayerColor,
			SetColors:      []color.Color{color.White, color.Blue, color.Black, color.Red, color.Green},
		}},
	}}
	devoid := vanillaCreature("Devoid Red Drone", 2, 2, game.Devoid)
	devoid.ManaCost = opt.Val(cost.Mana{cost.O(2)})
	devoid.Colors = []color.Color{color.Red}
	printedRed := vanillaCreature("Printed Red Bear", 2, 2)
	printedRed.ManaCost = opt.Val(cost.Mana{cost.O(2)})
	printedRed.Colors = []color.Color{color.Red}
	redGrant := vanillaCreature("Red Cascade Grant", 1, 1)
	redGrant.StaticAbilities = []game.StaticAbility{
		cascadeGrantStatic("Red spells you cast have cascade.", game.Selection{ColorsAny: []color.Color{color.Red}}, zone.None),
	}
	for _, tc := range []struct {
		def  *game.CardDef
		want int
	}{
		{allColors, 1},
		{devoid, 0},
		{printedRed, 1},
	} {
		t.Run(tc.def.Name, func(t *testing.T) {
			engine, g := newEffectiveCascadeGame()
			addCombatPermanent(g, game.Player1, redGrant)
			addCardToLibrary(g, game.Player1, simpleGainLifeInstantWithManaValue("Hit", 1))
			cardID := addCardToHand(g, game.Player1, tc.def)
			g.Players[game.Player1].ManaPool.Add(mana.C, 2)
			if _, got := castCountingCascade(t, engine, g, game.Player1, action.CastSpell(cardID, nil, 0, nil)); got != tc.want {
				t.Fatalf("%s cascade triggers = %d, want %d", tc.def.Name, got, tc.want)
			}
		})
	}
}

// TestSpellKeywordGrantSubtypeSelectionSeesMistformUltimus proves the payment
// planner and one-shot consumption routes read a non-Changeling subtype CDA:
// a Sliver-filtered improvise grant applies to and is consumed by a Mistform
// Ultimus spell, and the shared CardDef is not mutated.
func TestSpellKeywordGrantSubtypeSelectionSeesMistformUltimus(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.RuleEffects = append(g.RuleEffects, nextSpellImproviseGrant(g, game.Selection{SubtypesAny: []types.Sub{types.Sliver}}))
	ultimus := cardm.MistformUltimus()
	if !spellGrantedKeyword(g, game.Player1, ultimus, 0, zone.Hand, game.Improvise) {
		t.Fatal("Sliver-filtered improvise grant does not apply to a Mistform Ultimus spell")
	}
	if len(ultimus.Subtypes) != 1 || ultimus.Subtypes[0] != types.Illusion {
		t.Fatalf("Mistform Ultimus subtypes mutated: %v", ultimus.Subtypes)
	}
	if spellGrantedKeyword(g, game.Player1, conditionalSliverTestDef(), 0, zone.Hand, game.Improvise) {
		t.Fatal("Sliver-filtered improvise grant applies to a conditional, non-CDA Sliver type")
	}
	spell := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell, Controller: game.Player1, SourceTokenDef: ultimus}
	consumeNextSpellKeywordGrantEffects(g, spell, zone.Hand)
	if len(g.RuleEffects) != 0 {
		t.Fatalf("rule effects = %#v, want the Sliver grant consumed by the Mistform Ultimus spell", g.RuleEffects)
	}
}

// conditionalSliverTestDef is a stand-in whose self type effect is gated by a
// condition ("As long as this creature has three or more +1/+1 counters on it,
// it's a Sliver."), which CR 604.3a(5) excludes from characteristic-defining
// abilities.
func conditionalSliverTestDef() *game.CardDef {
	def := vanillaCreature("Conditional Sliver", 2, 2)
	def.ManaCost = opt.Val(cost.Mana{cost.O(2)})
	def.StaticAbilities = []game.StaticAbility{{
		Text: "As long as this creature has three or more +1/+1 counters on it, it's a Sliver.",
		Condition: opt.Val(game.Condition{
			Object:        opt.Val(game.SourcePermanentReference()),
			ObjectMatches: opt.Val(game.Selection{RequiredCounterCount: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 3}), RequiredCounter: counter.PlusOnePlusOne}),
		}),
		ContinuousEffects: []game.ContinuousEffect{{
			AffectedSource: true,
			Layer:          game.LayerType,
			AddSubtypes:    []types.Sub{types.Sliver},
		}},
	}}
	return def
}

func TestCascadeSpellGrantValidation(t *testing.T) {
	valid := flamekinHeraldTestDef()
	if issues := game.ValidateCardDef(valid); len(issues) != 0 {
		t.Fatalf("valid cascade grant: %v", issues)
	}
	nextOnly := flamekinHeraldTestDef()
	nextOnly.StaticAbilities[0].RuleEffects[0].AppliesToNextSpellOnly = true
	if issues := game.ValidateCardDef(nextOnly); len(issues) == 0 {
		t.Fatal("next-spell-only cascade grant validated")
	}
	badZone := quandrixTheProofTestDef()
	badZone.StaticAbilities[3].RuleEffects[0].CastFromZone = zone.Battlefield
	if issues := game.ValidateCardDef(badZone); len(issues) == 0 {
		t.Fatal("battlefield cast-from zone validated")
	}
	copyTrigger := theFirstDoctorTestDef()
	copyTrigger.TriggeredAbilities[0].Trigger.Pattern.MatchSpellCopy = true
	if issues := game.ValidateCardDef(copyTrigger); len(issues) == 0 {
		t.Fatal("cascade keyword filter on a spell-copy trigger validated")
	}
}
