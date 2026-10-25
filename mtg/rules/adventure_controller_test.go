package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
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

// Controller-scoped Adventure exile permission (#3353, bounded Engine slice of
// #170).
//
// Source review:
//   - Comprehensive Rules effective 2026-09-25,
//     https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt
//     (sha256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca):
//     400.7 (a card that changes zones is a new object; none of the 400.7a-m
//     exceptions carries an Adventure permission across a later move), 715.3c
//     (copies of Adventure spells are Adventures), 715.3d ("Instead of putting a
//     spell that was cast as an Adventure into its owner's graveyard as it
//     resolves, its controller exiles it. For as long as that card remains
//     exiled, that player may play it. It can't be cast as an Adventure this
//     way ...") and 715.4.
//   - Scryfall rulings for revision 2a0d430f-da84-4752-940c-8457c525aac9
//     (2019-10-04) and 384ec8b7-472e-4026-9632-349ab9df2b27 (2022-06-10): the
//     spell's controller exiles it and "that player" may later cast it as a
//     permanent spell; countered or all-target-illegal Adventures are not
//     exiled; other exile gives no permission; timing restrictions still apply;
//     copies cease to exist and can't be cast.
//
// The fixtures below are unregistered exact-card proving fixtures from the
// pinned Oracle corpus. They do not establish production admission.

// rosethornAcolyteFixture is Rosethorn Acolyte // Seasonal Ritual, oracle_id
// 639a8804-8513-4ba4-8683-ea1e2e5a9530, revision
// 2a0d430f-da84-4752-940c-8457c525aac9.
func rosethornAcolyteFixture() *game.CardDef {
	const ritualText = "Add one mana of any color. (Then exile this card. You may cast the creature later from exile.)"
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:          "Rosethorn Acolyte",
			ManaCost:      opt.Val(cost.Mana{cost.O(2), cost.G}),
			Colors:        []color.Color{color.Green},
			Types:         []types.Card{types.Creature},
			Subtypes:      []types.Sub{types.Elf, types.Druid},
			Power:         opt.Val(game.PT{Value: 2}),
			Toughness:     opt.Val(game.PT{Value: 3}),
			OracleText:    "{T}: Add one mana of any color.",
			ManaAbilities: []game.ManaAbility{game.TapAnyColorManaAbility()},
		},
		Layout: game.LayoutAdventure,
		Alternate: opt.Val(game.CardFace{
			Name:       "Seasonal Ritual",
			ManaCost:   opt.Val(cost.Mana{cost.G}),
			Colors:     []color.Color{color.Green},
			Types:      []types.Card{types.Sorcery},
			Subtypes:   []types.Sub{types.Adventure},
			OracleText: ritualText,
			SpellAbility: opt.Val(game.Mode{
				Text: ritualText,
				Sequence: []game.Instruction{
					{Primitive: game.Choose{
						Choice: game.ResolutionChoice{
							Kind:   game.ResolutionChoiceMana,
							Prompt: "Choose a color",
							Colors: []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G},
						},
						PublishChoice: game.ChoiceKey("seasonal-ritual-color"),
					}},
					{Primitive: game.AddMana{Amount: game.Fixed(1), ChoiceFrom: game.ChoiceKey("seasonal-ritual-color")}},
				},
			}.Ability()),
		}),
	}
}

// crystalDragonFixture is Crystal Dragon // Rob the Hoard, oracle_id
// 56b0cb64-b886-413e-bb20-392bf12d3926, revision
// 384ec8b7-472e-4026-9632-349ab9df2b27.
func crystalDragonFixture() *game.CardDef {
	const robText = "Return target artifact, enchantment, or legendary card from your graveyard to your hand. (Then exile this card. You may cast the creature later from exile.)"
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:            "Crystal Dragon",
			ManaCost:        opt.Val(cost.Mana{cost.O(4), cost.W, cost.W}),
			Colors:          []color.Color{color.White},
			Types:           []types.Card{types.Creature},
			Subtypes:        []types.Sub{types.Dragon},
			Power:           opt.Val(game.PT{Value: 4}),
			Toughness:       opt.Val(game.PT{Value: 4}),
			OracleText:      "Flying, vigilance",
			StaticAbilities: []game.StaticAbility{game.FlyingStaticBody, game.VigilanceStaticBody},
		},
		Layout: game.LayoutAdventure,
		Alternate: opt.Val(game.CardFace{
			Name:       "Rob the Hoard",
			ManaCost:   opt.Val(cost.Mana{cost.O(1), cost.W}),
			Colors:     []color.Color{color.White},
			Types:      []types.Card{types.Sorcery},
			Subtypes:   []types.Sub{types.Adventure},
			OracleText: robText,
			SpellAbility: opt.Val(game.Mode{
				Text: robText,
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "target artifact, enchantment, or legendary card from your graveyard",
					Allow:      game.TargetAllowCard,
					TargetZone: zone.Graveyard,
					Selection: opt.Val(game.Selection{
						Controller: game.ControllerYou,
						AnyOf: []game.Selection{
							{RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment}},
							{Supertypes: []types.Super{types.Legendary}},
						},
					}),
				}},
				Sequence: []game.Instruction{{Primitive: game.MoveCard{
					Card:        game.CardReference{Kind: game.CardReferenceTarget},
					FromZone:    zone.Graveyard,
					Destination: zone.Hand,
				}}},
			}.Ability()),
		}),
	}
}

// stealStackSpell resolves an effect controlled by newController that gains
// control of the spell (CR 110.2a-style stack control change through the
// shared ChangeStackObjectController operation).
func stealStackSpell(t *testing.T, g *game.Game, engine *Engine, spellID id.ID, newController game.PlayerID) {
	t.Helper()
	addEffectSpellToStack(g, newController, game.ChangeStackObjectController{
		Object:     game.TargetStackObjectReference(0),
		Controller: game.ControllerReference(),
	}, []game.Target{game.StackObjectTarget(spellID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	obj, ok := stackObjectByID(g, spellID)
	if !ok || obj.Controller != newController {
		t.Fatalf("stack spell %v controller = %+v, want %v", spellID, obj, newController)
	}
}

// castSeasonalRitual has Player1 cast Seasonal Ritual from hand during
// Player1's main phase and returns the card and stack object.
func castSeasonalRitual(t *testing.T, g *game.Game, engine *Engine) (id.ID, *game.StackObject) {
	t.Helper()
	cardID := addCardToHand(g, game.Player1, rosethornAcolyteFixture())
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	if !engine.applyAction(g, game.Player1, action.CastSpellFace(cardID, game.FaceAlternate, nil, 0, nil)) {
		t.Fatal("Player1 could not cast Seasonal Ritual")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != cardID || obj.Face != game.FaceAlternate || obj.Controller != game.Player1 {
		t.Fatalf("stack object = %+v, want Player1's Seasonal Ritual", obj)
	}
	return cardID, obj
}

// fundAndGivePriority gives player three untapped Forests and sorcery-speed
// priority on their own main phase with an empty stack.
func fundAndGivePriority(g *game.Game, player game.PlayerID) {
	for range 3 {
		addBasicLandPermanent(g, player, types.Forest)
	}
	giveMainPhasePriority(g, player)
}

func giveMainPhasePriority(g *game.Game, player game.PlayerID) {
	g.Turn.ActivePlayer = player
	g.Turn.PriorityPlayer = player
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
}

func adventureFrontFromExile(cardID id.ID) action.Action {
	return action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil, 0, nil)
}

// assertAdventureCastable checks both legal-action enumeration and direct
// action validation for player casting cardID's creature face from exile.
func assertAdventureCastable(t *testing.T, g *game.Game, engine *Engine, player game.PlayerID, cardID id.ID, want bool) {
	t.Helper()
	giveMainPhasePriority(g, player)
	act := adventureFrontFromExile(cardID)
	if got := actionsContain(engine.legalActions(g, player), act); got != want {
		t.Fatalf("legal actions for %v contain exile creature cast = %v, want %v", player, got, want)
	}
	if got := engine.canCastSpellFaceFromZoneWithKicker(g, player, cardID, zone.Exile, game.FaceFront, nil, 0, nil, false); got != want {
		t.Fatalf("direct validation for %v exile creature cast = %v, want %v", player, got, want)
	}
}

func TestStolenAdventureGrantsPermissionToResolvingController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID, obj := castSeasonalRitual(t, g, engine)
	stealStackSpell(t, g, engine, obj.ID, game.Player2)

	engine.resolveTopOfStack(g, &TurnLog{})

	if !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("resolved Adventure is not in its owner's exile")
	}
	if got := g.Players[game.Player2].ManaPool.Total(); got != 1 {
		t.Fatalf("Player2 mana = %d, want Seasonal Ritual's one mana", got)
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("Player1 mana = %d, want none from the stolen Ritual", got)
	}
	g.Players[game.Player2].ManaPool.Empty()
	fundAndGivePriority(g, game.Player1)
	fundAndGivePriority(g, game.Player2)
	assertAdventureCastable(t, g, engine, game.Player1, cardID, false)
	assertAdventureCastable(t, g, engine, game.Player2, cardID, true)
	giveMainPhasePriority(g, game.Player1)
	if engine.applyAction(g, game.Player1, adventureFrontFromExile(cardID)) {
		t.Fatal("owner cast the creature through another player's Adventure permission")
	}
	giveMainPhasePriority(g, game.Player2)
	if !engine.applyAction(g, game.Player2, adventureFrontFromExile(cardID)) {
		t.Fatal("Player2 could not cast the creature from Player1's exile")
	}
	cast, ok := g.Stack.Peek()
	if !ok || cast.SourceID != cardID || cast.Controller != game.Player2 || cast.SourceZone != zone.Exile || cast.Face != game.FaceFront {
		t.Fatalf("stack object = %+v, want Player2's front-face cast from exile", cast)
	}
	if g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("cast creature remained in its owner's exile")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	permanent := permanentForCard(g, cardID)
	if permanent == nil || effectiveController(g, permanent) != game.Player2 || permanent.Owner != game.Player1 {
		t.Fatalf("permanent = %+v, want Player1-owned Acolyte controlled by Player2", permanent)
	}
}

// stolenResolvedRitual has Player1 cast Seasonal Ritual, Player2 gain control
// of it, and the stolen spell resolve into Player1's exile.
func stolenResolvedRitual(t *testing.T) (*game.Game, *Engine, id.ID) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID, obj := castSeasonalRitual(t, g, engine)
	stealStackSpell(t, g, engine, obj.ID, game.Player2)
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player1].Exile.Contains(cardID) || g.AdventureCards[cardID].Player != game.Player2 {
		t.Fatalf("stolen Ritual exile = %v permission = %+v, want Player2's permission in Player1's exile",
			g.Players[game.Player1].Exile.Contains(cardID), g.AdventureCards[cardID])
	}
	g.Players[game.Player2].ManaPool.Empty()
	return g, engine, cardID
}

func lastSpellCastEvent(t *testing.T, g *game.Game) game.Event {
	t.Helper()
	for i := len(g.Events) - 1; i >= 0; i-- {
		if g.Events[i].Kind == game.EventSpellCast {
			return g.Events[i]
		}
	}
	t.Fatal("no spell cast event")
	return game.Event{}
}

func untappedLandCount(g *game.Game, player game.PlayerID) int {
	count := 0
	for _, permanent := range g.Battlefield {
		if permanent.Controller == player && !permanent.Tapped {
			count++
		}
	}
	return count
}

func TestStolenAdventureTwiceGrantsOnlyFinalController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID, obj := castSeasonalRitual(t, g, engine)
	stealStackSpell(t, g, engine, obj.ID, game.Player2)
	stealStackSpell(t, g, engine, obj.ID, game.Player3)
	engine.resolveTopOfStack(g, &TurnLog{})

	for player, want := range map[game.PlayerID]int{game.Player1: 0, game.Player2: 0, game.Player3: 1} {
		if got := g.Players[player].ManaPool.Total(); got != want {
			t.Fatalf("%v mana = %d, want %d", player, got, want)
		}
	}
	g.Players[game.Player3].ManaPool.Empty()
	for _, player := range []game.PlayerID{game.Player1, game.Player2, game.Player3} {
		fundAndGivePriority(g, player)
	}
	assertAdventureCastable(t, g, engine, game.Player1, cardID, false)
	assertAdventureCastable(t, g, engine, game.Player2, cardID, false)
	assertAdventureCastable(t, g, engine, game.Player4, cardID, false)
	assertAdventureCastable(t, g, engine, game.Player3, cardID, true)
}

func TestStolenAdventurePermissionKeepsFaceCostAndTimingRules(t *testing.T) {
	g, engine, cardID := stolenResolvedRitual(t)

	// Payment failure: two Forests cannot pay {2}{G}; nothing moves or taps.
	addBasicLandPermanent(g, game.Player2, types.Forest)
	addBasicLandPermanent(g, game.Player2, types.Forest)
	assertAdventureCastable(t, g, engine, game.Player2, cardID, false)
	if engine.applyAction(g, game.Player2, adventureFrontFromExile(cardID)) {
		t.Fatal("Player2 cast the creature without paying its mana cost")
	}
	if !g.Players[game.Player1].Exile.Contains(cardID) || g.Stack.Size() != 0 ||
		untappedLandCount(g, game.Player2) != 2 || g.AdventureCards[cardID].Player != game.Player2 {
		t.Fatal("failed payment moved the card, tapped lands or consumed the permission")
	}
	addBasicLandPermanent(g, game.Player2, types.Forest)

	// The permission is for the creature only (CR 715.3d).
	giveMainPhasePriority(g, game.Player2)
	adventureAct := action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceAlternate, nil, 0, nil)
	if actionsContain(engine.legalActions(g, game.Player2), adventureAct) ||
		engine.canCastSpellFaceFromZoneWithKicker(g, game.Player2, cardID, zone.Exile, game.FaceAlternate, nil, 0, nil, false) ||
		engine.applyAction(g, game.Player2, adventureAct) {
		t.Fatal("Player2 recast the Adventure face from exile")
	}

	// Sorcery timing: not Player2's turn.
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player2
	if actionsContain(engine.legalActions(g, game.Player2), adventureFrontFromExile(cardID)) ||
		engine.applyAction(g, game.Player2, adventureFrontFromExile(cardID)) {
		t.Fatal("Player2 cast the creature outside their own main phase")
	}
	// Sorcery timing: non-empty stack.
	giveMainPhasePriority(g, game.Player2)
	pending := addEffectSpellToStack(g, game.Player1, game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}, nil)
	if actionsContain(engine.legalActions(g, game.Player2), adventureFrontFromExile(cardID)) ||
		engine.applyAction(g, game.Player2, adventureFrontFromExile(cardID)) {
		t.Fatal("Player2 cast the creature while the stack was not empty")
	}
	if pendingObj, ok := g.Stack.Peek(); !ok || pendingObj.SourceID != pending {
		t.Fatal("unexpected stack state before clearing the pending spell")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	casts := spellCastEventCount(g)
	giveMainPhasePriority(g, game.Player2)
	if !engine.applyAction(g, game.Player2, adventureFrontFromExile(cardID)) {
		t.Fatal("Player2 could not cast the creature with legal timing and payment")
	}
	if untappedLandCount(g, game.Player2) != 0 {
		t.Fatal("front-face cast did not pay its full mana cost")
	}
	cast := lastSpellCastEvent(t, g)
	if spellCastEventCount(g) != casts+1 || cast.Controller != game.Player2 || cast.CardID != cardID ||
		cast.FromZone != zone.Exile || cast.Face != game.FaceFront {
		t.Fatalf("spell cast event = %+v, want one Player2 front-face cast from exile", cast)
	}
	if _, ok := g.AdventureCards[cardID]; ok {
		t.Fatal("Adventure permission survived the card leaving exile")
	}
}

func TestStolenAdventureIndependentPermissionsRemainUsable(t *testing.T) {
	g, engine, cardID := stolenResolvedRitual(t)
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		ID:             g.IDGen.Next(),
		Kind:           game.RuleEffectCastFromZone,
		Controller:     game.Player1,
		AffectedPlayer: game.PlayerYou,
		CastFromZone:   zone.Exile,
		AffectedCardID: cardID,
	})
	fundAndGivePriority(g, game.Player1)
	fundAndGivePriority(g, game.Player2)
	assertAdventureCastable(t, g, engine, game.Player1, cardID, true)
	assertAdventureCastable(t, g, engine, game.Player2, cardID, true)

	giveMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, adventureFrontFromExile(cardID)) {
		t.Fatal("Player1 could not use the independent cast-from-exile permission")
	}
	if obj, ok := g.Stack.Peek(); !ok || obj.Controller != game.Player1 || obj.SourceID != cardID {
		t.Fatalf("stack object = %+v, want Player1's independent cast", obj)
	}
	if _, ok := g.AdventureCards[cardID]; ok {
		t.Fatal("Player2's Adventure permission survived the card leaving exile")
	}
}

// castSpellsFromExileEffect is a continuous group permission controlled by
// player: "You may cast spells from exile" (RuleEffectCastSpellsFromZone),
// optionally limited to spells of the listed card types.
func castSpellsFromExileEffect(g *game.Game, player game.PlayerID, spellTypes ...types.Card) game.RuleEffect {
	return game.RuleEffect{
		ID:             g.IDGen.Next(),
		Kind:           game.RuleEffectCastSpellsFromZone,
		Controller:     player,
		AffectedPlayer: game.PlayerYou,
		CastFromZone:   zone.Exile,
		SpellTypes:     spellTypes,
	}
}

func TestStolenAdventureGroupPermissionRemainsUsable(t *testing.T) {
	t.Run("matching group permission", func(t *testing.T) {
		g, engine, cardID := stolenResolvedRitual(t)
		g.RuleEffects = append(g.RuleEffects, castSpellsFromExileEffect(g, game.Player1))
		fundAndGivePriority(g, game.Player1)
		fundAndGivePriority(g, game.Player2)
		assertAdventureCastable(t, g, engine, game.Player1, cardID, true)
		assertAdventureCastable(t, g, engine, game.Player2, cardID, true)

		giveMainPhasePriority(g, game.Player1)
		if !engine.applyAction(g, game.Player1, adventureFrontFromExile(cardID)) {
			t.Fatal("Player1 could not use the group cast-spells-from-exile permission")
		}
		if obj, ok := g.Stack.Peek(); !ok || obj.Controller != game.Player1 || obj.SourceID != cardID || obj.Face != game.FaceFront {
			t.Fatalf("stack object = %+v, want Player1's group-permission front-face cast", obj)
		}
		if _, ok := g.AdventureCards[cardID]; ok {
			t.Fatal("Player2's Adventure permission survived the card leaving exile")
		}
	})
	t.Run("filtered group permission that does not cover the creature", func(t *testing.T) {
		g, engine, cardID := stolenResolvedRitual(t)
		g.RuleEffects = append(g.RuleEffects, castSpellsFromExileEffect(g, game.Player1, types.Instant))
		fundAndGivePriority(g, game.Player1)
		fundAndGivePriority(g, game.Player2)
		assertAdventureCastable(t, g, engine, game.Player1, cardID, false)
		assertAdventureCastable(t, g, engine, game.Player2, cardID, true)
	})
	t.Run("another player's group permission", func(t *testing.T) {
		g, engine, cardID := stolenResolvedRitual(t)
		g.RuleEffects = append(g.RuleEffects, castSpellsFromExileEffect(g, game.Player3))
		fundAndGivePriority(g, game.Player1)
		fundAndGivePriority(g, game.Player2)
		assertAdventureCastable(t, g, engine, game.Player1, cardID, false)
		assertAdventureCastable(t, g, engine, game.Player2, cardID, true)
	})
}

// mutatingAdventureFixture is a test-only Adventure card whose creature face
// has Mutate {G}: "Mutating Adventurer" {5} Creature // "Small Errand" {G}
// Sorcery — Adventure, "You gain 1 life."
func mutatingAdventureFixture() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:            "Mutating Adventurer",
			ManaCost:        opt.Val(cost.Mana{cost.O(5)}),
			Colors:          []color.Color{color.Green},
			Types:           []types.Card{types.Creature},
			Power:           opt.Val(game.PT{Value: 3}),
			Toughness:       opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{game.MutateStaticAbility(cost.Mana{cost.G})},
		},
		Layout: game.LayoutAdventure,
		Alternate: opt.Val(game.CardFace{
			Name:     "Small Errand",
			ManaCost: opt.Val(cost.Mana{cost.G}),
			Colors:   []color.Color{color.Green},
			Types:    []types.Card{types.Sorcery},
			Subtypes: []types.Sub{types.Adventure},
			SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{{
				Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()},
			}}}.Ability()),
		}),
	}
}

// TestAdventureMutateFromExileFollowsPermissionHolder covers the Mutate cast
// route from exile: only a player holding the CR 715.3d permission, or an
// independent per-card or group permission, may mutate-cast the exiled
// creature face.
func TestAdventureMutateFromExileFollowsPermissionHolder(t *testing.T) {
	setup := func(t *testing.T, steal bool) (*game.Game, *Engine, id.ID, *game.Permanent) {
		t.Helper()
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		target := addCombatPermanent(g, game.Player1, mutateTargetCard())
		cardID := addCardToHand(g, game.Player1, mutatingAdventureFixture())
		addBasicLandPermanent(g, game.Player1, types.Forest)
		giveMainPhasePriority(g, game.Player1)
		if !engine.applyAction(g, game.Player1, action.CastSpellFace(cardID, game.FaceAlternate, nil, 0, nil)) {
			t.Fatal("Small Errand cast failed")
		}
		if steal {
			obj, _ := g.Stack.Peek()
			stealStackSpell(t, g, engine, obj.ID, game.Player2)
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		if !g.Players[game.Player1].Exile.Contains(cardID) {
			t.Fatal("Small Errand did not resolve into its owner's exile")
		}
		addBasicLandPermanent(g, game.Player1, types.Forest)
		giveMainPhasePriority(g, game.Player1)
		return g, engine, cardID, target
	}
	mutateAct := func(cardID id.ID, target *game.Permanent) action.Action {
		return action.CastMutateSpellFromZone(cardID, zone.Exile, target.ObjectID)
	}
	assertMutate := func(t *testing.T, g *game.Game, engine *Engine, cardID id.ID, target *game.Permanent, want bool) {
		t.Helper()
		act := mutateAct(cardID, target)
		if got := actionsContain(engine.legalActions(g, game.Player1), act); got != want {
			t.Fatalf("legal actions contain owner's exile Mutate cast = %v, want %v", got, want)
		}
		if got := canCastMutateSpell(g, game.Player1, cardID, zone.Exile, target.ObjectID); got != want {
			t.Fatalf("owner's exile Mutate validation = %v, want %v", got, want)
		}
	}

	t.Run("owner-controlled Adventure", func(t *testing.T) {
		g, engine, cardID, target := setup(t, false)
		assertMutate(t, g, engine, cardID, target, true)
		if !engine.applyAction(g, game.Player1, mutateAct(cardID, target)) {
			t.Fatal("owner could not mutate-cast after their own Adventure")
		}
		if obj, ok := g.Stack.Peek(); !ok || !obj.Mutate || obj.SourceZone != zone.Exile || obj.Controller != game.Player1 {
			t.Fatalf("stack object = %+v, want Player1's Mutate cast from exile", obj)
		}
	})
	t.Run("stolen Adventure", func(t *testing.T) {
		g, engine, cardID, target := setup(t, true)
		assertMutate(t, g, engine, cardID, target, false)
		if engine.applyAction(g, game.Player1, mutateAct(cardID, target)) {
			t.Fatal("owner mutate-cast through another player's Adventure permission")
		}
	})
	t.Run("stolen Adventure with owner's group permission", func(t *testing.T) {
		g, engine, cardID, target := setup(t, true)
		g.RuleEffects = append(g.RuleEffects, castSpellsFromExileEffect(g, game.Player1))
		assertMutate(t, g, engine, cardID, target, true)
		if !engine.applyAction(g, game.Player1, mutateAct(cardID, target)) {
			t.Fatal("owner could not mutate-cast through the group permission")
		}
	})
	// CR 702.140a: Mutate is an alternative cost usable whenever an
	// independent effect lets the owner cast the creature from exile.
	cardPermission := func(g *game.Game, controller game.PlayerID, cardID id.ID) game.RuleEffect {
		return game.RuleEffect{
			ID:             g.IDGen.Next(),
			Kind:           game.RuleEffectCastFromZone,
			Controller:     controller,
			AffectedPlayer: game.PlayerYou,
			CastFromZone:   zone.Exile,
			AffectedCardID: cardID,
		}
	}
	t.Run("stolen Adventure with owner's per-card permission", func(t *testing.T) {
		g, engine, cardID, target := setup(t, true)
		g.RuleEffects = append(g.RuleEffects, cardPermission(g, game.Player1, cardID))
		assertMutate(t, g, engine, cardID, target, true)
		if !engine.applyAction(g, game.Player1, mutateAct(cardID, target)) {
			t.Fatal("owner could not mutate-cast through the per-card permission")
		}
		if obj, ok := g.Stack.Peek(); !ok || !obj.Mutate || obj.SourceZone != zone.Exile || obj.Controller != game.Player1 {
			t.Fatalf("stack object = %+v, want Player1's per-card-permission Mutate cast from exile", obj)
		}
	})
	t.Run("stolen Adventure with owner's play permission", func(t *testing.T) {
		g, engine, cardID, target := setup(t, true)
		play := cardPermission(g, game.Player1, cardID)
		play.Kind = game.RuleEffectPlayFromZone
		g.RuleEffects = append(g.RuleEffects, play)
		assertMutate(t, g, engine, cardID, target, true)
		if !engine.applyAction(g, game.Player1, mutateAct(cardID, target)) {
			t.Fatal("owner could not mutate-cast through the per-card play permission")
		}
	})
	t.Run("stolen Adventure with a per-card permission for another card", func(t *testing.T) {
		g, engine, cardID, target := setup(t, true)
		other := addCardToExile(g, game.Player1, mutatingAdventureFixture())
		g.RuleEffects = append(g.RuleEffects, cardPermission(g, game.Player1, other))
		assertMutate(t, g, engine, cardID, target, false)
		if engine.applyAction(g, game.Player1, mutateAct(cardID, target)) {
			t.Fatal("owner mutate-cast through another card's permission")
		}
	})
	t.Run("stolen Adventure with another player's per-card permission", func(t *testing.T) {
		g, engine, cardID, target := setup(t, true)
		g.RuleEffects = append(g.RuleEffects, cardPermission(g, game.Player3, cardID))
		assertMutate(t, g, engine, cardID, target, false)
		if engine.applyAction(g, game.Player1, mutateAct(cardID, target)) {
			t.Fatal("owner mutate-cast through Player3's per-card permission")
		}
	})
}

// TestRegisteredStolenAdventureGroupPermission uses registered Faerie
// Guidemother // Gift of the Fae stolen by registered Aethersnatch: the owner
// may still cast the exiled creature through an independent group permission,
// and enumeration offers it.
func TestRegisteredStolenAdventureGroupPermission(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	guidemother := registry.Lookup("Faerie Guidemother")
	aethersnatch := registry.Lookup("Aethersnatch")
	if guidemother == nil || aethersnatch == nil {
		t.Fatal("missing default registry cards")
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngineWithPrograms(nil, registry.Programs())
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Gift Target", Types: []types.Card{types.Creature},
	}})
	cardID := addCardToHand(g, game.Player1, guidemother)
	for range 2 {
		addBasicLandPermanent(g, game.Player1, types.Plains)
	}
	giveMainPhasePriority(g, game.Player1)
	gift := action.CastSpellFace(cardID, game.FaceAlternate, []game.Target{game.PermanentTarget(creature.ObjectID)}, 0, nil)
	if !engine.applyAction(g, game.Player1, gift) {
		t.Fatal("Gift of the Fae cast failed")
	}
	giftObj, _ := g.Stack.Peek()
	g.Turn.PriorityPlayer = game.Player2
	snatchID := addCardToHand(g, game.Player2, aethersnatch)
	for range 6 {
		addBasicLandPermanent(g, game.Player2, types.Island)
	}
	if !engine.applyAction(g, game.Player2, action.CastSpell(snatchID, []game.Target{game.StackObjectTarget(giftObj.ID)}, 0, nil)) {
		t.Fatal("Aethersnatch cast failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player1].Exile.Contains(cardID) || g.AdventureCards[cardID].Player != game.Player2 {
		t.Fatalf("stolen Gift permission = %+v, want Player2's permission in Player1's exile", g.AdventureCards[cardID])
	}

	addBasicLandPermanent(g, game.Player1, types.Plains)
	assertAdventureCastable(t, g, engine, game.Player1, cardID, false)
	g.RuleEffects = append(g.RuleEffects, castSpellsFromExileEffect(g, game.Player1))
	assertAdventureCastable(t, g, engine, game.Player1, cardID, true)
	if !engine.applyAction(g, game.Player1, adventureFrontFromExile(cardID)) {
		t.Fatal("owner could not cast Faerie Guidemother through the group permission")
	}
	if obj, ok := g.Stack.Peek(); !ok || obj.Controller != game.Player1 || obj.SourceID != cardID || obj.Face != game.FaceFront {
		t.Fatalf("stack object = %+v, want Player1's Faerie Guidemother cast", obj)
	}
}

func TestOwnerControlledAdventuresStillGrantPermission(t *testing.T) {
	t.Run("Seasonal Ritual", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		cardID, _ := castSeasonalRitual(t, g, engine)
		engine.resolveTopOfStack(g, &TurnLog{})
		if got := g.Players[game.Player1].ManaPool.Total(); got != 1 {
			t.Fatalf("Player1 mana = %d, want 1", got)
		}
		g.Players[game.Player1].ManaPool.Empty()
		fundAndGivePriority(g, game.Player1)
		fundAndGivePriority(g, game.Player2)
		assertAdventureCastable(t, g, engine, game.Player2, cardID, false)
		assertAdventureCastable(t, g, engine, game.Player1, cardID, true)
		if !engine.applyAction(g, game.Player1, adventureFrontFromExile(cardID)) {
			t.Fatal("owner could not cast the creature after their own Adventure")
		}
	})
	t.Run("Rob the Hoard", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		cardID := addCardToHand(g, game.Player1, crystalDragonFixture())
		relic := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name: "Graveyard Relic", Types: []types.Card{types.Artifact},
		}})
		for range 2 {
			addBasicLandPermanent(g, game.Player1, types.Plains)
		}
		giveMainPhasePriority(g, game.Player1)
		rob := action.CastSpellFace(cardID, game.FaceAlternate, []game.Target{game.CardTarget(relic)}, 0, nil)
		if !engine.applyAction(g, game.Player1, rob) {
			t.Fatal("Player1 could not cast Rob the Hoard")
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		if !g.Players[game.Player1].Hand.Contains(relic) {
			t.Fatal("Rob the Hoard did not return its target")
		}
		for range 6 {
			addBasicLandPermanent(g, game.Player1, types.Plains)
			addBasicLandPermanent(g, game.Player2, types.Plains)
		}
		assertAdventureCastable(t, g, engine, game.Player2, cardID, false)
		assertAdventureCastable(t, g, engine, game.Player1, cardID, true)
	})
}

func TestAdventureWithoutResolutionIntoExileGrantsNoPermission(t *testing.T) {
	assertNoPermission := func(t *testing.T, g *game.Game, engine *Engine, cardID id.ID) {
		t.Helper()
		if _, ok := g.AdventureCards[cardID]; ok {
			t.Fatalf("Adventure permission = %+v, want none", g.AdventureCards[cardID])
		}
		for _, player := range []game.PlayerID{game.Player1, game.Player2} {
			fundAndGivePriority(g, player)
			assertAdventureCastable(t, g, engine, player, cardID, false)
		}
	}
	t.Run("stolen and countered", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		cardID, obj := castSeasonalRitual(t, g, engine)
		stealStackSpell(t, g, engine, obj.ID, game.Player2)
		if !counterStackObject(g, obj.ID) || !g.Players[game.Player1].Graveyard.Contains(cardID) {
			t.Fatal("countered stolen Adventure did not go to its owner's graveyard")
		}
		assertNoPermission(t, g, engine, cardID)
	})
	t.Run("all targets illegal", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		cardID := addCardToHand(g, game.Player1, crystalDragonFixture())
		relic := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name: "Graveyard Relic", Types: []types.Card{types.Artifact},
		}})
		addBasicLandPermanent(g, game.Player1, types.Plains)
		addBasicLandPermanent(g, game.Player1, types.Plains)
		giveMainPhasePriority(g, game.Player1)
		if !engine.applyAction(g, game.Player1, action.CastSpellFace(cardID, game.FaceAlternate, []game.Target{game.CardTarget(relic)}, 0, nil)) {
			t.Fatal("Player1 could not cast Rob the Hoard")
		}
		obj, _ := g.Stack.Peek()
		stealStackSpell(t, g, engine, obj.ID, game.Player2)
		moveCardBetweenZones(g, game.Player1, relic, zone.Graveyard, zone.Exile)
		engine.resolveTopOfStack(g, &TurnLog{})
		if !g.Players[game.Player1].Graveyard.Contains(cardID) || g.Players[game.Player1].Hand.Contains(relic) {
			t.Fatal("all-target-illegal Adventure resolved instead of going to its owner's graveyard")
		}
		assertNoPermission(t, g, engine, cardID)
	})
	t.Run("ordinary exile", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		cardID := addCardToExile(g, game.Player1, rosethornAcolyteFixture())
		assertNoPermission(t, g, engine, cardID)
	})
	t.Run("destination replaced to graveyard", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		cardID, obj := castSeasonalRitual(t, g, engine)
		stealStackSpell(t, g, engine, obj.ID, game.Player2)
		g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
			MatchEvent: game.EventZoneChanged, MatchFromZone: true, FromZone: zone.Stack,
			MatchToZone: true, ToZone: zone.Exile, ReplaceToZone: zone.Graveyard,
		})
		engine.resolveTopOfStack(g, &TurnLog{})
		if !g.Players[game.Player1].Graveyard.Contains(cardID) || g.Players[game.Player1].Exile.Contains(cardID) {
			t.Fatal("replacement did not redirect the resolving Adventure to the graveyard")
		}
		assertNoPermission(t, g, engine, cardID)
	})
}

func TestAdventurePermissionExpiresWhenCardLeavesExile(t *testing.T) {
	for _, route := range []string{"exile to graveyard and back", "exile to exile"} {
		t.Run(route, func(t *testing.T) {
			g, engine, cardID := stolenResolvedRitual(t)
			stale := g.AdventureCards[cardID]
			if route == "exile to exile" {
				if !moveCardBetweenZones(g, game.Player1, cardID, zone.Exile, zone.Exile) {
					t.Fatal("exile-to-exile move failed")
				}
			} else {
				if !moveCardBetweenZones(g, game.Player1, cardID, zone.Exile, zone.Graveyard) ||
					!moveCardBetweenZones(g, game.Player1, cardID, zone.Graveyard, zone.Exile) {
					t.Fatal("exile round trip failed")
				}
			}
			if !g.Players[game.Player1].Exile.Contains(cardID) {
				t.Fatal("same physical card is not back in exile")
			}
			if _, ok := g.AdventureCards[cardID]; ok {
				t.Fatal("Adventure permission survived the card leaving exile")
			}
			fundAndGivePriority(g, game.Player1)
			fundAndGivePriority(g, game.Player2)
			assertAdventureCastable(t, g, engine, game.Player2, cardID, false)
			assertAdventureCastable(t, g, engine, game.Player1, cardID, false)

			// A stale record cannot revive the permission for the new object.
			g.AdventureCards[cardID] = stale
			assertAdventureCastable(t, g, engine, game.Player2, cardID, false)
		})
	}
}

func TestAdventurePermissionEndsWhenBeneficiaryLeavesGame(t *testing.T) {
	g, _, cardID := stolenResolvedRitual(t)
	g.Players[game.Player2].Eliminated = true
	leaveDepartedOwnedObjects(g)
	if _, ok := g.AdventureCards[cardID]; ok {
		t.Fatal("departed beneficiary's Adventure permission remained")
	}
	if !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("surviving owner's exiled card was removed")
	}
}

func TestAdventureCopiesGrantNoPermission(t *testing.T) {
	t.Run("stack copy", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		cardID, obj := castSeasonalRitual(t, g, engine)
		stealStackSpell(t, g, engine, obj.ID, game.Player2)
		addEffectSpellToStack(g, game.Player3, game.CopyStackObject{Object: game.TargetStackObjectReference(0)},
			[]game.Target{game.StackObjectTarget(obj.ID)})
		engine.resolveTopOfStack(g, &TurnLog{})
		copyObj, ok := g.Stack.Peek()
		if !ok || !copyObj.Copy || copyObj.Controller != game.Player3 {
			t.Fatalf("top of stack = %+v, want Player3's copy", copyObj)
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		if got := g.Players[game.Player3].ManaPool.Total(); got != 1 {
			t.Fatalf("Player3 copy mana = %d, want 1", got)
		}
		if _, ok := g.AdventureCards[cardID]; ok {
			t.Fatal("resolved Adventure copy created an exile permission")
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		if got := g.AdventureCards[cardID].Player; got != game.Player2 {
			t.Fatalf("permission player = %v, want original's final controller Player2", got)
		}
		for _, player := range []game.PlayerID{game.Player1, game.Player2, game.Player3} {
			fundAndGivePriority(g, player)
		}
		assertAdventureCastable(t, g, engine, game.Player3, cardID, false)
		assertAdventureCastable(t, g, engine, game.Player2, cardID, true)
	})
	t.Run("storm copy", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		def := rosethornAcolyteFixture()
		alternate := def.Alternate.Val
		alternate.StaticAbilities = []game.StaticAbility{game.StormStaticBody}
		def.Alternate = opt.Val(alternate)
		cardID := addCardToHand(g, game.Player1, def)
		firstID := addCardToHand(g, game.Player1, simpleGainLifeInstant("First Spell"))
		addBasicLandPermanent(g, game.Player1, types.Forest)
		giveMainPhasePriority(g, game.Player1)
		if !engine.applyAction(g, game.Player1, action.CastSpell(firstID, nil, 0, nil)) {
			t.Fatal("first spell cast failed")
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		casts := spellCastEventCount(g)
		if !engine.applyAction(g, game.Player1, action.CastSpellFace(cardID, game.FaceAlternate, nil, 0, nil)) {
			t.Fatal("storm Seasonal Ritual cast failed")
		}
		if spellCastEventCount(g) != casts+1 || g.Stack.Size() != 2 {
			t.Fatalf("casts = %d stack = %d, want one cast plus one storm copy", spellCastEventCount(g)-casts, g.Stack.Size())
		}
		var original *game.StackObject
		for _, obj := range g.Stack.Objects() {
			if !obj.Copy {
				original = obj
			}
		}
		stealStackSpell(t, g, engine, original.ID, game.Player2)
		engine.resolveTopOfStack(g, &TurnLog{})
		if _, ok := g.AdventureCards[cardID]; ok {
			t.Fatal("resolved storm copy created an exile permission")
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		if g.Players[game.Player1].ManaPool.Total() != 1 || g.Players[game.Player2].ManaPool.Total() != 1 {
			t.Fatal("storm copy and stolen original did not each add mana for their controllers")
		}
		if got := g.AdventureCards[cardID].Player; got != game.Player2 {
			t.Fatalf("permission player = %v, want stolen original's controller Player2", got)
		}
	})
	t.Run("casting a card copy", func(t *testing.T) {
		g, engine, cardID := stolenResolvedRitual(t)
		before := g.AdventureCards[cardID]
		casts := spellCastEventCount(g)
		if !engine.castFreeCopyOfCard(g, game.Player3, cardID, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
			t.Fatal("casting a copy of the exiled card failed")
		}
		copyObj, ok := g.Stack.Peek()
		if !ok || copyObj.Face != game.FaceFront || copyObj.Controller != game.Player3 || spellCastEventCount(g) != casts+1 {
			t.Fatalf("stack object = %+v, want one Player3 front-face card-copy cast", copyObj)
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		if !g.Players[game.Player1].Exile.Contains(cardID) || g.AdventureCards[cardID] != before {
			t.Fatal("casting a card copy disturbed the original's Adventure permission")
		}
		fundAndGivePriority(g, game.Player2)
		assertAdventureCastable(t, g, engine, game.Player2, cardID, true)
	})
}

func TestAdventurePermissionCloneAndReplay(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID, obj := castSeasonalRitual(t, g, engine)
	stealStackSpell(t, g, engine, obj.ID, game.Player2)
	replay := g.Clone()

	engine.resolveTopOfStack(g, &TurnLog{})
	engine.resolveTopOfStack(replay, &TurnLog{})
	if g.AdventureCards[cardID] != replay.AdventureCards[cardID] || g.AdventureCards[cardID].Player != game.Player2 ||
		g.AdventureCards[cardID].ZoneVersion != g.CardInstances[cardID].ZoneVersion {
		t.Fatalf("replayed permission = %+v, original = %+v", replay.AdventureCards[cardID], g.AdventureCards[cardID])
	}
	if len(g.Events) != len(replay.Events) {
		t.Fatalf("replay events = %d, want %d", len(replay.Events), len(g.Events))
	}

	branch := g.Clone()
	if branch.AdventureCards[cardID] != g.AdventureCards[cardID] {
		t.Fatal("clone did not preserve the Adventure beneficiary and incarnation")
	}
	fundAndGivePriority(branch, game.Player2)
	if !engine.applyAction(branch, game.Player2, adventureFrontFromExile(cardID)) {
		t.Fatal("cloned beneficiary could not cast the creature")
	}
	if _, ok := branch.AdventureCards[cardID]; ok {
		t.Fatal("branch permission survived the cast")
	}
	if g.AdventureCards[cardID].Player != game.Player2 || !g.Players[game.Player1].Exile.Contains(cardID) || g.Stack.Size() != 0 {
		t.Fatal("casting in a clone mutated the original game's permission or zones")
	}
	fundAndGivePriority(g, game.Player2)
	assertAdventureCastable(t, g, engine, game.Player2, cardID, true)
}

// castWatcherFixture is a test-only permanent: "Whenever a player casts a
// spell, you gain 1 life."
func castWatcherFixture() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Cast Watcher",
		Types: []types.Card{types.Enchantment},
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{
				Type:    game.TriggerWhenever,
				Pattern: game.TriggerPattern{Event: game.EventSpellCast, Controller: game.TriggerControllerAny},
			},
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()},
			}}}.Ability(),
		}},
	}}
}

// watchCasts puts a Cast Watcher under Player4's control and skips trigger
// detection for earlier events.
func watchCasts(g *game.Game) id.ID {
	watcher := addCombatPermanent(g, game.Player4, castWatcherFixture())
	g.TriggerEventCursor = len(g.Events)
	return watcher.CardInstanceID
}

// resolveCastWatcherTriggers puts pending triggers on the stack, resolves
// every Cast Watcher trigger and returns how many there were.
func resolveCastWatcherTriggers(t *testing.T, g *game.Game, engine *Engine, watcherCardID id.ID) int {
	t.Helper()
	engine.putTriggeredAbilitiesOnStack(g)
	life := g.Players[game.Player4].Life
	count := 0
	for {
		obj, ok := g.Stack.Peek()
		if !ok || obj.Kind != game.StackTriggeredAbility || obj.SourceCardID != watcherCardID {
			break
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		count++
	}
	if got := g.Players[game.Player4].Life - life; got != count {
		t.Fatalf("Cast Watcher gained %d life from %d triggers", got, count)
	}
	return count
}

// TestAdventureCastTriggersFollowGenuineCasts proves cast-keyed triggers see a
// genuine cast of a copy of the exiled Adventure card (CR 707.12) and the
// original storm spell, but not storm copies or other copies put directly onto
// the stack (CR 707.10).
func TestAdventureCastTriggersFollowGenuineCasts(t *testing.T) {
	t.Run("casting a card copy", func(t *testing.T) {
		g, engine, cardID := stolenResolvedRitual(t)
		watcher := watchCasts(g)
		if !engine.castFreeCopyOfCard(g, game.Player3, cardID, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
			t.Fatal("casting a copy of the exiled card failed")
		}
		if got := resolveCastWatcherTriggers(t, g, engine, watcher); got != 1 {
			t.Fatalf("Cast Watcher triggers = %d, want 1 for the card-copy cast", got)
		}
		if obj, ok := g.Stack.Peek(); !ok || obj.Controller != game.Player3 || obj.Face != game.FaceFront {
			t.Fatalf("stack object = %+v, want Player3's cast card copy", obj)
		}
	})
	t.Run("stack copy", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		_, obj := castSeasonalRitual(t, g, engine)
		watcher := watchCasts(g)
		addEffectSpellToStack(g, game.Player3, game.CopyStackObject{Object: game.TargetStackObjectReference(0)},
			[]game.Target{game.StackObjectTarget(obj.ID)})
		engine.resolveTopOfStack(g, &TurnLog{})
		if copyObj, ok := g.Stack.Peek(); !ok || !copyObj.Copy {
			t.Fatalf("top of stack = %+v, want the copy", copyObj)
		}
		if got := resolveCastWatcherTriggers(t, g, engine, watcher); got != 0 {
			t.Fatalf("Cast Watcher triggers = %d, want none for a copied spell", got)
		}
	})
	t.Run("storm", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		def := rosethornAcolyteFixture()
		alternate := def.Alternate.Val
		alternate.StaticAbilities = []game.StaticAbility{game.StormStaticBody}
		def.Alternate = opt.Val(alternate)
		cardID := addCardToHand(g, game.Player1, def)
		firstID := addCardToHand(g, game.Player1, simpleGainLifeInstant("First Spell"))
		addBasicLandPermanent(g, game.Player1, types.Forest)
		giveMainPhasePriority(g, game.Player1)
		if !engine.applyAction(g, game.Player1, action.CastSpell(firstID, nil, 0, nil)) {
			t.Fatal("first spell cast failed")
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		watcher := watchCasts(g)
		if !engine.applyAction(g, game.Player1, action.CastSpellFace(cardID, game.FaceAlternate, nil, 0, nil)) {
			t.Fatal("storm Seasonal Ritual cast failed")
		}
		copies := 0
		for _, obj := range g.Stack.Objects() {
			if obj.Copy && obj.SourceID == cardID {
				copies++
			}
		}
		if copies != 1 {
			t.Fatalf("storm copies = %d, want 1", copies)
		}
		if got := resolveCastWatcherTriggers(t, g, engine, watcher); got != 1 {
			t.Fatalf("Cast Watcher triggers = %d, want 1 for the original storm cast only", got)
		}
	})
}

const adventureReplacementLabel = "Adventure (CR 715.3d)"

// replacementPickingAgent answers each CR 616.1 replacement choice with the
// first option whose label satisfies pick, and takes the first option of every
