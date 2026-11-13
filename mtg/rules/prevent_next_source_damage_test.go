package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: one-shot "a source of your choice" prevention shields.
// Comprehensive Rules effective 2026-11-06 (local verified copy
// .cardwork/MagicCompRules-20261106.txt):
//   - 609.7a: the source is chosen when the prevention effect is created;
//     permanents, spells on the stack (including permanent spells, whose
//     prevention also covers the resulting permanent) and objects referred to
//     by objects on the stack may be chosen.
//   - 609.7b: the chosen source's characteristics required by the effect are
//     checked again when the damage would be dealt; if they no longer match,
//     the damage is not prevented (and the shield remains).
//   - 400.7c: an effect that prevents damage from a permanent spell continues
//     to apply to the permanent the spell becomes.
//   - 615.1, 615.3, 616.1: prevention shields are prevention effects applied
//     to the next damage event they match.
// Representative cards: Circle of Protection: Red/Artifacts, Rune of
// Protection: Lands, Story Circle ("of the chosen color"), Deflecting Palm.

// protectionShieldSource adds a Circle of Protection-style enchantment for
// controller and returns the activated-ability stack object whose resolution
// creates the one-shot shield.
func protectionShieldSource(g *game.Game, controller game.PlayerID) (*game.Permanent, *game.StackObject) {
	circle := addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:   "Circle of Protection",
		Types:  []types.Card{types.Enchantment},
		Colors: []color.Color{color.White},
	}})
	return circle, &game.StackObject{
		ID: g.IDGen.Next(), Kind: game.StackActivatedAbility, SourceID: circle.ObjectID, Controller: controller,
	}
}

// oneShotShield is the "The next time a source of your choice would deal
// damage to you this turn, prevent that damage." primitive.
func oneShotShield(colors []color.Color, cardTypes []types.Card) game.PreventDamage {
	return game.PreventDamage{
		Player: game.ControllerReference(), All: true, OneShot: true,
		SourceColors: colors, SourceTypes: cardTypes,
	}
}

// sourceChoiceAgent picks the offered source whose card is cardID and records
// the cards it was offered.
func sourceChoiceAgent(t *testing.T, cardID id.ID, offered *[]id.ID) *controlScript {
	t.Helper()
	return &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
		if request.Kind != game.ChoicePreventionAllocation {
			return request.DefaultSelection
		}
		selected := -1
		for _, option := range request.Options {
			if !option.Card.Exists {
				continue
			}
			if offered != nil {
				*offered = append(*offered, option.Card.Val.CardID)
			}
			if option.Card.Val.CardID == cardID {
				selected = option.Index
			}
		}
		if selected < 0 {
			t.Fatalf("source card %d was not offered in %+v", cardID, request.Options)
		}
		return []int{selected}
	}}
}

func createOneShotShield(t *testing.T, engine *Engine, g *game.Game, obj *game.StackObject, prim game.PreventDamage, agent PlayerAgent) {
	t.Helper()
	var agents [game.NumPlayers]PlayerAgent
	agents[obj.Controller] = agent
	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: prim}, agents, &TurnLog{})
}

func dealDamageFrom(g *game.Game, source *game.Permanent, player game.PlayerID, amount int) int {
	return dealPlayerDamage(g, source.CardInstanceID, source.ObjectID, source.Controller, player, amount, false)
}

// TestOneShotShieldChoosesRedSpellOnStack proves the Circle of Protection: Red
// shape: with a red Lightning Bolt-style spell on the stack aimed at its
// controller, the shield chooses that spell as it is created (CR 609.7a) and
// prevents the spell's damage when it resolves, while a red creature that was
// not chosen still deals damage.
func TestOneShotShieldChoosesRedSpellOnStack(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addColoredCreaturePermanent(g, game.Player1, 2, color.Red)
	boltCard := addEffectSpellToStack(g, game.Player1, game.Damage{Amount: game.Fixed(3), Recipient: game.AnyTargetDamageRecipient(0)},
		[]game.Target{{Kind: game.TargetPlayer, PlayerID: game.Player2}})
	boltDef := g.CardInstances[boltCard].Def
	boltDef.Name, boltDef.Types, boltDef.Colors = "Lightning Bolt", []types.Card{types.Instant}, []color.Color{color.Red}
	_, ability := protectionShieldSource(g, game.Player2)

	var offered []id.ID
	createOneShotShield(t, engine, g, ability, oneShotShield([]color.Color{color.Red}, nil), sourceChoiceAgent(t, boltCard, &offered))
	if !slices.Contains(offered, boltCard) || !slices.Contains(offered, creature.CardInstanceID) || len(offered) != 2 {
		t.Fatalf("offered sources = %v, want the red spell and the red creature", offered)
	}
	if len(g.PreventionShields) != 1 || g.PreventionShields[0].ChosenSourceObjectID != g.Stack.Objects()[0].ID {
		t.Fatalf("shields = %+v, want one bound to the Bolt on the stack", g.PreventionShields)
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player2].Life != 40 || len(g.PreventionShields) != 0 {
		t.Fatalf("after Bolt: life %d, shields %+v; want 40 and the shield used", g.Players[game.Player2].Life, g.PreventionShields)
	}
	assertEvent(t, g.Events, game.EventDamagePrevented, func(event game.Event) bool {
		return event.Player == game.Player2 && event.Amount == 3 && event.SourceID == boltCard
	})
	if dealt := dealDamageFrom(g, creature, game.Player2, 2); dealt != 2 {
		t.Fatalf("unchosen creature dealt %d, want 2", dealt)
	}
}

// TestOneShotShieldBindsOnlyChosenSource proves the binding is literal: damage
// from another source with the required color passes and leaves the shield,
// and the chosen source's next damage is then prevented.
func TestOneShotShieldBindsOnlyChosenSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	chosen := addColoredCreaturePermanent(g, game.Player1, 2, color.Red)
	other := addColoredCreaturePermanent(g, game.Player1, 3, color.Red)
	_, ability := protectionShieldSource(g, game.Player2)
	createOneShotShield(t, engine, g, ability, oneShotShield([]color.Color{color.Red}, nil), sourceChoiceAgent(t, chosen.CardInstanceID, nil))

	if dealt := dealDamageFrom(g, other, game.Player2, 3); dealt != 3 {
		t.Fatalf("unchosen red source dealt %d, want 3", dealt)
	}
	if len(g.PreventionShields) != 1 {
		t.Fatalf("shields = %+v, want the shield still waiting for its chosen source", g.PreventionShields)
	}
	if dealt := dealDamageFrom(g, chosen, game.Player2, 2); dealt != 0 {
		t.Fatalf("chosen source dealt %d, want 0", dealt)
	}
	if len(g.PreventionShields) != 0 || g.Players[game.Player2].Life != 37 {
		t.Fatalf("after chosen damage: life %d, shields %+v; want 37 and no shield", g.Players[game.Player2].Life, g.PreventionShields)
	}
}

// TestOneShotShieldArtifactFilterOffersOnlyArtifacts proves the Circle of
// Protection: Artifacts filter: only artifact sources, including a creature
// that an effect made an artifact, are offered, and with no artifact source no
// shield is created.
func TestOneShotShieldArtifactFilterOffersOnlyArtifacts(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addColoredCreaturePermanent(g, game.Player1, 2, color.Red)
	_, ability := protectionShieldSource(g, game.Player2)
	createOneShotShield(t, engine, g, ability, oneShotShield(nil, []types.Card{types.Artifact}), nil)
	if len(g.PreventionShields) != 0 {
		t.Fatalf("shields = %+v, want none without an artifact source", g.PreventionShields)
	}

	golem := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Golem", Types: []types.Card{types.Artifact, types.Creature},
		Power: opt.Val(game.PT{Value: 3}), Toughness: opt.Val(game.PT{Value: 3}),
	}})
	animated := addColoredCreaturePermanent(g, game.Player1, 1, color.Blue)
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID: g.IDGen.Next(), Layer: game.LayerType, AffectedObjectID: animated.ObjectID,
		AddTypes: []types.Card{types.Artifact}, Duration: game.DurationPermanent,
	})
	var offered []id.ID
	createOneShotShield(t, engine, g, ability, oneShotShield(nil, []types.Card{types.Artifact}), sourceChoiceAgent(t, golem.CardInstanceID, &offered))
	if len(offered) != 2 || !slices.Contains(offered, golem.CardInstanceID) || !slices.Contains(offered, animated.CardInstanceID) {
		t.Fatalf("offered sources = %v, want only the two artifacts", offered)
	}
	if dealt := dealDamageFrom(g, creature, game.Player2, 2); dealt != 2 {
		t.Fatalf("non-artifact dealt %d, want 2", dealt)
	}
	if dealt := dealDamageFrom(g, golem, game.Player2, 3); dealt != 0 {
		t.Fatalf("chosen artifact dealt %d, want 0", dealt)
	}
}

// TestOneShotShieldLandFilterPreventsAnimatedManland proves the Rune of
// Protection: Lands filter chooses a land that has become a creature and
// prevents its damage.
func TestOneShotShieldLandFilterPreventsAnimatedManland(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addColoredCreaturePermanent(g, game.Player1, 2, color.Red)
	manland := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Mutavault", Types: []types.Card{types.Land},
	}})
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID: g.IDGen.Next(), Layer: game.LayerType, AffectedObjectID: manland.ObjectID,
		AddTypes: []types.Card{types.Creature}, Duration: game.DurationUntilEndOfTurn,
	})
	_, ability := protectionShieldSource(g, game.Player2)
	createOneShotShield(t, engine, g, ability, oneShotShield(nil, []types.Card{types.Land}), nil)
	if len(g.PreventionShields) != 1 || g.PreventionShields[0].ChosenSourceObjectID != manland.ObjectID {
		t.Fatalf("shields = %+v, want one bound to the manland", g.PreventionShields)
	}
	if dealt := dealDamageFrom(g, manland, game.Player2, 2); dealt != 0 {
		t.Fatalf("manland dealt %d, want 0", dealt)
	}
}

// TestOneShotShieldChosenColorReadsEntryChoice proves the Story Circle shape
// ("a source of your choice of the chosen color"): the color the shield's
// source chose as it entered fixes the filter, only sources of that color are
// eligible, and without a color choice no shield is created.
func TestOneShotShieldChosenColorReadsEntryChoice(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	red := addColoredCreaturePermanent(g, game.Player1, 3, color.Red)
	green := addColoredCreaturePermanent(g, game.Player1, 2, color.Green)
	circle, ability := protectionShieldSource(g, game.Player2)
	prim := oneShotShield(nil, nil)
	prim.SourceColorFromEntryChoice = true

	createOneShotShield(t, engine, g, ability, prim, nil)
	if len(g.PreventionShields) != 0 {
		t.Fatalf("shields = %+v, want none without an entry color choice", g.PreventionShields)
	}

	circle.EntryChoices = map[game.ChoiceKey]game.ResolutionChoiceResult{
		game.EntryColorChoiceKey: {Kind: game.ResolutionChoiceMana, Color: mana.G},
	}
	createOneShotShield(t, engine, g, ability, prim, nil)
	if len(g.PreventionShields) != 1 || g.PreventionShields[0].ChosenSourceObjectID != green.ObjectID ||
		!slices.Equal(g.PreventionShields[0].SourceColors, []color.Color{color.Green}) {
		t.Fatalf("shields %+v; want one bound to the only green source with a green filter", g.PreventionShields)
	}
	if dealt := dealDamageFrom(g, red, game.Player2, 3); dealt != 3 {
		t.Fatalf("red source dealt %d, want 3", dealt)
	}
	if dealt := dealDamageFrom(g, green, game.Player2, 2); dealt != 0 {
		t.Fatalf("chosen green source dealt %d, want 0", dealt)
	}
}

// TestOneShotShieldRechecksChosenSourceColor proves CR 609.7b: a red source
// chosen for a red-source shield that is no longer red when it deals damage is
// not prevented, and the shield remains.
func TestOneShotShieldRechecksChosenSourceColor(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addColoredCreaturePermanent(g, game.Player1, 2, color.Red)
	_, ability := protectionShieldSource(g, game.Player2)
	createOneShotShield(t, engine, g, ability, oneShotShield([]color.Color{color.Red}, nil), nil)
	if len(g.PreventionShields) != 1 || g.PreventionShields[0].ChosenSourceObjectID != source.ObjectID {
		t.Fatalf("shields = %+v, want one bound to the red source", g.PreventionShields)
	}
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID: g.IDGen.Next(), Layer: game.LayerColor, AffectedObjectID: source.ObjectID,
		SetColors: []color.Color{color.White}, Duration: game.DurationUntilEndOfTurn,
	})
	if dealt := dealDamageFrom(g, source, game.Player2, 2); dealt != 2 {
		t.Fatalf("no-longer-red source dealt %d, want 2", dealt)
	}
	if len(g.PreventionShields) != 1 {
		t.Fatalf("shields = %+v, want the shield to remain", g.PreventionShields)
	}
}

// TestOneShotShieldFollowsPermanentSpellOntoBattlefield proves a creature spell
// chosen on the stack stays chosen as the creature it becomes (CR 400.7c), so
// that creature's later damage is prevented.
func TestOneShotShieldFollowsPermanentSpellOntoBattlefield(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spell := pushFaceUpCardSpell(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Goblin Piker", Types: []types.Card{types.Creature}, Colors: []color.Color{color.Red},
		Power: opt.Val(game.PT{Value: 2}), Toughness: opt.Val(game.PT{Value: 1}),
	}})
	_, ability := protectionShieldSource(g, game.Player2)
	createOneShotShield(t, engine, g, ability, oneShotShield([]color.Color{color.Red}, nil), nil)
	if len(g.PreventionShields) != 1 || g.PreventionShields[0].ChosenSourceObjectID != spell.ID {
		t.Fatalf("shields = %+v, want one bound to the creature spell", g.PreventionShields)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	creature := permanentForCard(g, spell.SourceID)
	if creature == nil {
		t.Fatal("creature spell did not enter the battlefield")
	}
	if dealt := dealDamageFrom(g, creature, game.Player2, 2); dealt != 0 {
		t.Fatalf("creature from the chosen spell dealt %d, want 0", dealt)
	}
	if len(g.PreventionShields) != 0 {
		t.Fatalf("shields = %+v, want the shield used", g.PreventionShields)
	}
}

// addPreventionSourceSpell adds a resolving Deflecting Palm-style instant
// stack object for owner.
func addPreventionSourceSpell(g *game.Game, owner game.PlayerID, name string) *game.StackObject {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID: cardID,
		Def: &game.CardDef{CardFace: game.CardFace{
			Name:   name,
			Types:  []types.Card{types.Instant},
			Colors: []color.Color{color.Red, color.White},
		}},
		Owner: owner,
	}
	return &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: cardID, Controller: owner}
}

func deflectingPalmShield() game.PreventDamage {
	prim := oneShotShield(nil, nil)
	prim.PreventedRiders = []game.PreventedDamageRider{game.PreventedDamageRiderDamageSourceController}
	return prim
}

// TestPreventNextSourceRedirectDealsPreventedAmountToSourceController proves the
// Deflecting Palm shield ("The next time a source of your choice would deal
// damage to you this turn, prevent that damage. If damage is prevented this way,
// Deflecting Palm deals that much damage to that source's controller.") prevents
// the chosen source's next damage to its controller and deals that prevented
// amount back to the source's controller, as damage from Deflecting Palm.
func TestPreventNextSourceRedirectDealsPreventedAmountToSourceController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addColoredCreaturePermanent(g, game.Player1, 4, color.Red)
	palm := addPreventionSourceSpell(g, game.Player2, "Deflecting Palm")

	resolveInstruction(engine, g, palm, deflectingPalmShield(), nil)

	if dealt := dealDamageFrom(g, source, game.Player2, 4); dealt != 0 {
		t.Fatalf("dealt damage to shielded player = %d, want 0 after prevention", dealt)
	}
	if g.Players[game.Player2].Life != 40 || g.Players[game.Player1].Life != 36 {
		t.Fatalf("life = %d/%d, want 36/40 after redirect", g.Players[game.Player1].Life, g.Players[game.Player2].Life)
	}
	if len(g.PreventionShields) != 0 {
		t.Fatalf("prevention shields = %+v, want expired after one event", g.PreventionShields)
	}
	assertEvent(t, g.Events, game.EventDamagePrevented, func(event game.Event) bool {
		return event.Player == game.Player2 && event.Amount == 4
	})
	assertEvent(t, g.Events, game.EventDamageDealt, func(event game.Event) bool {
		return event.Player == game.Player1 && event.SourceID == palm.SourceID && event.Amount == 4
	})
}

// TestPreventNextSourceRedirectIsItselfPreventable proves the redirected damage
// is a new damage event from Deflecting Palm: the source's controller, having
// chosen the Palm spell on the stack for their own one-shot shield, prevents it.
func TestPreventNextSourceRedirectIsItselfPreventable(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addColoredCreaturePermanent(g, game.Player1, 5, color.Red)
	palm := addPreventionSourceSpell(g, game.Player2, "Deflecting Palm")
	g.Stack.Push(palm)

	// Player1 responds to Deflecting Palm by choosing it as their shield's source.
	_, ability := protectionShieldSource(g, game.Player1)
	createOneShotShield(t, engine, g, ability, oneShotShield(nil, nil), sourceChoiceAgent(t, palm.SourceID, nil))
	g.Stack.Pop()
	resolveInstruction(engine, g, palm, deflectingPalmShield(), nil)

	if dealt := dealDamageFrom(g, source, game.Player2, 5); dealt != 0 {
		t.Fatalf("dealt damage to shielded player = %d, want 0", dealt)
	}
	if g.Players[game.Player2].Life != 40 || g.Players[game.Player1].Life != 40 {
		t.Fatalf("life = %d/%d, want 40/40 after the redirect is prevented", g.Players[game.Player1].Life, g.Players[game.Player2].Life)
	}
	if len(g.PreventionShields) != 0 {
		t.Fatalf("prevention shields = %+v, want both expired", g.PreventionShields)
	}
}

// TestOneShotShieldResolvingSpellMayChooseItself proves the resolving spell is
// still on the stack while it resolves (CR 608.2n), so a real Deflecting Palm
// cast with no other source in the game may choose itself (CR 609.7a) and its
// shield is created.
func TestOneShotShieldResolvingSpellMayChooseItself(t *testing.T) {
	g, engine := spellTypeScenario(t)
	card := addCardToHand(g, game.Player1, productionChoiceCard(t, "Deflecting Palm"))
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	g.Players[game.Player1].ManaPool.Add(mana.W, 1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(card, nil, 0, nil)) {
		t.Fatal("casting Deflecting Palm failed")
	}
	palm, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("Deflecting Palm is not on the stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if len(g.PreventionShields) != 1 || g.PreventionShields[0].ChosenSourceObjectID != palm.ID {
		t.Fatalf("shields = %+v, want one bound to the resolving Palm spell %d", g.PreventionShields, palm.ID)
	}
}

// TestOneShotShieldResolvingAbilityMayChooseItsDepartedSource proves a
// resolving ability's source is a legal choice even after it left the
// battlefield (sacrificed as a cost): the ability is still on the stack and
// refers to it (CR 608.2n, 609.7a).
func TestOneShotShieldResolvingAbilityMayChooseItsDepartedSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	circle, ability := protectionShieldSource(g, game.Player2)
	if !sacrificePermanent(g, circle) {
		t.Fatal("sacrificing the shield source failed")
	}
	createOneShotShield(t, engine, g, ability, oneShotShield([]color.Color{color.White}, nil), nil)
	if len(g.PreventionShields) != 1 || g.PreventionShields[0].ChosenSourceObjectID != circle.ObjectID {
		t.Fatalf("shields = %+v, want one bound to the departed ability source %d", g.PreventionShields, circle.ObjectID)
	}
}
