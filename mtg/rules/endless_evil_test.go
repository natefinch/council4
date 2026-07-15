package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// endlessEvilDef mirrors the compiler-generated CardDef for Endless Evil. It is
// built inline so the runtime behavior of the attached-permanent dies trigger,
// its last-known-information subtype gate, and the copy-token 1/1 upkeep ability
// can be exercised without regenerating the card corpus.
func endlessEvilDef() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Endless Evil",
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature you control",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
				}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepUpkeep,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenCopyOf(game.TokenCopySpec{
										Source:       game.TokenCopySourceObject,
										Object:       game.SourceAttachedPermanentReference(),
										SetPower:     opt.Val(game.PT{Value: 1}),
										SetToughness: opt.Val(game.PT{Value: 1}),
									}),
								},
							},
						},
					}.Ability(),
				},
				{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							Source:           game.TriggerSourceAttachedPermanent,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
						InterveningIf: "if that creature was a Horror",
						InterveningCondition: opt.Val(game.Condition{
							Object:        opt.Val(game.EventPermanentReference()),
							ObjectMatches: opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Horror}}),
						}),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.MoveCard{
									Card:        game.CardReference{Kind: game.CardReferenceSource},
									FromZone:    zone.Graveyard,
									Destination: zone.Hand,
								},
							},
						},
					}.Ability(),
				},
			},
		},
	}
}

// endlessEvilCreatureDef builds a creature the Endless Evil Aura can enchant,
// with the given subtype so a Horror and a non-Horror can both be tested.
func endlessEvilCreatureDef(name string, subtype types.Sub) *game.CardDef {
	pt := game.PT{Value: 2}
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      name,
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{subtype},
			Power:     opt.Val(pt),
			Toughness: opt.Val(pt),
		},
	}
}

// endlessEvilAttached bundles the game, engine, and the two permanents an
// Endless Evil attachment test drives, keeping the setup helper within the
// per-function return-result limit.
type endlessEvilAttached struct {
	game     *game.Game
	engine   *Engine
	aura     *game.Permanent
	creature *game.Permanent
}

// setupEndlessEvilAttached puts a creature of the given subtype onto the
// battlefield under Player1's control with an Endless Evil Aura attached to it.
func setupEndlessEvilAttached(t *testing.T, subtype types.Sub) endlessEvilAttached {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatPermanent(g, game.Player1, endlessEvilCreatureDef("Enchanted Creature", subtype))
	aura := addCombatPermanent(g, game.Player1, endlessEvilDef())
	if !attachPermanent(g, aura, creature) {
		t.Fatal("failed to attach Endless Evil to creature")
	}
	return endlessEvilAttached{game: g, engine: engine, aura: aura, creature: creature}
}

// TestEndlessEvilReturnsToHandWhenHorrorDies proves the attached-permanent dies
// trigger fires and returns the Aura card from the graveyard (where SBA put it
// after its enchanted creature died) to its owner's hand, gated on the dead
// creature having been a Horror in last-known information.
func TestEndlessEvilReturnsToHandWhenHorrorDies(t *testing.T) {
	setup := setupEndlessEvilAttached(t, types.Horror)
	g, engine, aura, creature := setup.game, setup.engine, setup.aura, setup.creature
	auraCardID := aura.CardInstanceID

	// The enchanted creature receives lethal damage and dies via SBA; the Aura
	// is put into the graveyard the same round of SBA checks.
	creature.MarkedDamage = 2
	engine.applyStateBasedActions(g)

	if _, ok := permanentByObjectID(g, creature.ObjectID); ok {
		t.Fatal("enchanted creature still on battlefield after lethal damage")
	}
	if !g.Players[game.Player1].Graveyard.Contains(auraCardID) {
		t.Fatal("Aura should be in the graveyard after its creature died")
	}

	log := TurnLog{}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &log) {
		t.Fatal("Endless Evil dies trigger was not put on the stack")
	}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &log)

	if !g.Players[game.Player1].Hand.Contains(auraCardID) {
		t.Fatal("Aura was not returned to its owner's hand after a Horror died")
	}
	if g.Players[game.Player1].Graveyard.Contains(auraCardID) {
		t.Fatal("Aura should have left the graveyard when returned to hand")
	}
}

// TestEndlessEvilStaysInGraveyardWhenNonHorrorDies proves the intervening
// subtype gate fails closed: when the dead creature was not a Horror, the dies
// trigger does not fire (or does nothing), so the Aura remains in the graveyard.
func TestEndlessEvilStaysInGraveyardWhenNonHorrorDies(t *testing.T) {
	setup := setupEndlessEvilAttached(t, types.Human)
	g, engine, aura, creature := setup.game, setup.engine, setup.aura, setup.creature
	auraCardID := aura.CardInstanceID

	creature.MarkedDamage = 2
	engine.applyStateBasedActions(g)

	if !g.Players[game.Player1].Graveyard.Contains(auraCardID) {
		t.Fatal("Aura should be in the graveyard after its creature died")
	}

	log := TurnLog{}
	// The intervening-if gate is evaluated when the trigger would be put on the
	// stack; a non-Horror creature must not place the return ability on the stack.
	engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &log)
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want 0 (intervening-if gate should suppress the trigger)", g.Stack.Size())
	}
	for g.Stack.Size() > 0 {
		engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &log)
	}

	if g.Players[game.Player1].Hand.Contains(auraCardID) {
		t.Fatal("Aura must not return to hand when a non-Horror died")
	}
	if !g.Players[game.Player1].Graveyard.Contains(auraCardID) {
		t.Fatal("Aura should remain in the graveyard when a non-Horror died")
	}
}

// TestEndlessEvilFailsWhenSourceMovedBeforeResolution proves the return uses the
// Aura source's current identity: if the Aura card leaves the graveyard (e.g. is
// exiled) after the trigger is put on the stack but before it resolves, the
// return finds no matching battlefield/graveyard source and the card is not moved
// to hand.
func TestEndlessEvilFailsWhenSourceMovedBeforeResolution(t *testing.T) {
	setup := setupEndlessEvilAttached(t, types.Horror)
	g, engine, aura, creature := setup.game, setup.engine, setup.aura, setup.creature
	auraCardID := aura.CardInstanceID

	creature.MarkedDamage = 2
	engine.applyStateBasedActions(g)

	log := TurnLog{}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &log) {
		t.Fatal("Endless Evil dies trigger was not put on the stack")
	}

	// The Aura card changes zones (graveyard -> exile) before the trigger
	// resolves, so "return this card" no longer has its source to return.
	if !moveCardBetweenZones(g, game.Player1, auraCardID, zone.Graveyard, zone.Exile) {
		t.Fatal("failed to move Aura card to exile before resolution")
	}

	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &log)

	if g.Players[game.Player1].Hand.Contains(auraCardID) {
		t.Fatal("Aura must not be returned to hand after its source changed zones")
	}
	if !g.Players[game.Player1].Exile.Contains(auraCardID) {
		t.Fatal("Aura should remain in exile where it was moved")
	}
}

// TestEndlessEvilReturnsWhenAuraAndCreatureDieSimultaneously proves the return
// still fires when the enchanted creature and the Aura leave the battlefield in
// the same simultaneous batch (e.g. a board wipe destroying both at once) rather
// than the Aura leaving as a follow-up state-based action. The creature's Horror
// subtype is captured in last-known information at the moment of death, so the
// intervening gate holds and the Aura returns from the graveyard.
func TestEndlessEvilReturnsWhenAuraAndCreatureDieSimultaneously(t *testing.T) {
	setup := setupEndlessEvilAttached(t, types.Horror)
	g, engine, aura, creature := setup.game, setup.engine, setup.aura, setup.creature
	auraCardID := aura.CardInstanceID

	if !movePermanentsToZoneSimultaneously(g, []*game.Permanent{creature, aura}, zone.Graveyard) {
		t.Fatal("failed to move creature and Aura to the graveyard simultaneously")
	}
	engine.applyStateBasedActions(g)

	if !g.Players[game.Player1].Graveyard.Contains(auraCardID) {
		t.Fatal("Aura should be in the graveyard after simultaneous death")
	}

	log := TurnLog{}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &log) {
		t.Fatal("Endless Evil dies trigger was not put on the stack after simultaneous death")
	}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &log)

	if !g.Players[game.Player1].Hand.Contains(auraCardID) {
		t.Fatal("Aura was not returned to hand after simultaneous Horror death")
	}
}

// TestEndlessEvilUpkeepTokenIsOneOne proves the upkeep copy-token override mints
// a token that is a copy of the enchanted creature except that it is a 1/1: the
// enchanted creature is a 4/4, but the created token's printed power and
// toughness are overridden to 1.
func TestEndlessEvilUpkeepTokenIsOneOne(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	bigPT := game.PT{Value: 4}
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Overlord of the Balemurk",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Horror},
		Power:     opt.Val(bigPT),
		Toughness: opt.Val(bigPT),
	}})
	aura := addCombatPermanent(g, game.Player1, endlessEvilDef())
	if !attachPermanent(g, aura, creature) {
		t.Fatal("failed to attach Endless Evil to creature")
	}

	obj := &game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackTriggeredAbility,
		Controller:   game.Player1,
		SourceID:     aura.ObjectID,
		SourceCardID: aura.CardInstanceID,
	}
	r := &effectResolver{engine: engine, game: g, obj: obj, log: &TurnLog{}}
	resolved := handleCreateToken(r, game.CreateToken{
		Amount: game.Fixed(1),
		Source: game.TokenCopyOf(game.TokenCopySpec{
			Source:       game.TokenCopySourceObject,
			Object:       game.SourceAttachedPermanentReference(),
			SetPower:     opt.Val(game.PT{Value: 1}),
			SetToughness: opt.Val(game.PT{Value: 1}),
		}),
	})
	if !resolved.succeeded {
		t.Fatal("upkeep copy-token creation did not succeed")
	}

	var token *game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanent.TokenDef != nil {
			token = permanent
			break
		}
	}
	if token == nil {
		t.Fatal("no copy token created")
	}
	if token.TokenDef.Name != "Overlord of the Balemurk" {
		t.Fatalf("token name = %q, want a copy of the enchanted creature", token.TokenDef.Name)
	}
	if got := effectivePower(g, token); got != 1 {
		t.Fatalf("token power = %d, want 1", got)
	}
	if got, _ := effectiveToughness(g, token); got != 1 {
		t.Fatalf("token toughness = %d, want 1", got)
	}
}
