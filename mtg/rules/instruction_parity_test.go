package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"

	cardc "github.com/natefinch/council4/mtg/cards/c"
	carde "github.com/natefinch/council4/mtg/cards/e"
	cardl "github.com/natefinch/council4/mtg/cards/l"
	cardn "github.com/natefinch/council4/mtg/cards/n"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLightningBoltInstructionSequenceDealsDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	sourceID := addImplementationSpellToStack(g, game.Player1, cardl.LightningBolt, []game.Target{game.PlayerTarget(game.Player2)})
	log := TurnLog{}

	engine.resolveTopOfStack(g, &log)

	if got := g.Players[game.Player2].Life; got != 37 {
		t.Fatalf("player 2 life = %d, want 37", got)
	}
	if !g.Players[game.Player1].Graveyard.Contains(sourceID) {
		t.Fatal("lightning bolt did not move to graveyard")
	}
}

func TestChaosWarpInstructionSequenceUsesOwnerRevealAndPutOnBattlefield(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Warped Creature", Types: []types.Card{types.Creature}}})
	target.Controller = game.Player3
	sourceID := addImplementationSpellToStack(g, game.Player1, cardc.ChaosWarp, []game.Target{game.PermanentTarget(target.ObjectID)})
	obj, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("chaos warp was not put on the stack")
	}
	log := TurnLog{}

	engine.resolveTopOfStack(g, &log)

	if _, ok := permanentByObjectID(g, target.ObjectID); ok {
		t.Fatal("original permanent object remained on battlefield")
	}
	if got := countCardPermanentsControlledBy(g, game.Player2, target.CardInstanceID); got != 1 {
		t.Fatalf("owner battlefield copies = %d, want 1", got)
	}
	if got := countCardPermanentsControlledBy(g, game.Player1, target.CardInstanceID); got != 0 {
		t.Fatalf("spell controller battlefield copies = %d, want 0", got)
	}
	if got := countCardPermanentsControlledBy(g, game.Player3, target.CardInstanceID); got != 0 {
		t.Fatalf("old controller battlefield copies = %d, want 0", got)
	}
	if !eventRevealedCard(g, target.CardInstanceID, obj.ID) {
		t.Fatal("reveal event for chaos warp target card was not emitted")
	}
	if !g.Players[game.Player1].Graveyard.Contains(sourceID) {
		t.Fatal("chaos warp did not move to graveyard")
	}
}

func TestNeyithFightTriggerInstructionSequenceDrawsCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	neyith := addCombatPermanent(g, game.Player1, cardn.NeyithOfTheDireHunt)
	drawn := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn Card"}})
	g.Stack.Push(&game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackTriggeredAbility,
		SourceID:     neyith.ObjectID,
		SourceCardID: neyith.CardInstanceID,
		Controller:   game.Player1,
		AbilityIndex: 0,
	})
	log := TurnLog{}

	engine.resolveTopOfStack(g, &log)

	if !g.Players[game.Player1].Hand.Contains(drawn) {
		t.Fatal("neyith fight trigger did not draw a card")
	}
}

func TestNeyithCombatTriggerInstructionSequencePaysAndAppliesEffects(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	neyith := addCombatPermanent(g, game.Player1, cardn.NeyithOfTheDireHunt)
	target := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	g.Players[game.Player1].ManaPool.Add(mana.R, 3)
	g.Stack.Push(&game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackTriggeredAbility,
		SourceID:     neyith.ObjectID,
		SourceCardID: neyith.CardInstanceID,
		Controller:   game.Player1,
		AbilityIndex: 2,
		Targets:      []game.Target{game.PermanentTarget(target.ObjectID)},
	})
	log := TurnLog{}

	engine.resolveTopOfStack(g, &log)

	if got := effectivePower(g, target); got != 6 {
		t.Fatalf("target power = %d, want 6", got)
	}
	if g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatalf("remaining mana = %d, want 0 after paying {2}{R/G}", g.Players[game.Player1].ManaPool.Total())
	}
	if len(log.Choices) != 2 {
		t.Fatalf("choice count = %d, want 2 optional/payment prompts", len(log.Choices))
	}
	if !slices.ContainsFunc(activeRuleEffects(g), func(effect game.RuleEffect) bool {
		return effect.Kind == game.RuleEffectMustBeBlocked && effect.AffectedObjectID == target.ObjectID
	}) {
		t.Fatal("neyith combat trigger did not create must-be-blocked rule effect")
	}
}

func TestActivatedAbilityInstructionSequenceResolves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Typed Activated Ability",
			Types: []types.Card{types.Creature},
			ActivatedAbilities: []game.ActivatedAbilityBody{{
				Content: game.Mode{
					Sequence: []game.Instruction{{
						Primitive: game.Draw{
							Amount:      game.Fixed(1),
							TargetIndex: game.TargetIndexController,
						},
					}},
				}.Ability(),
			}},
		},
	})
	drawn := addCardToLibrary(g, game.Player1, &game.CardDef{
		CardFace: game.CardFace{Name: "Drawn Card"},
	})
	g.Stack.Push(&game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackActivatedAbility,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
		Controller:   game.Player1,
		AbilityIndex: 0,
	})

	engine.resolveTopOfStack(g, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(drawn) {
		t.Fatal("typed activated ability did not draw a card")
	}
}

func TestEnduringCourageInstructionReturnsAsEnchantment(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	courage := addCombatPermanent(g, game.Player1, carde.EnduringCourage)
	if !movePermanentToZone(g, courage, zone.Graveyard) {
		t.Fatal("could not move Enduring Courage to graveyard")
	}
	g.Stack.Push(&game.StackObject{
		ID:              g.IDGen.Next(),
		Kind:            game.StackTriggeredAbility,
		SourceID:        courage.ObjectID,
		SourceCardID:    courage.CardInstanceID,
		Controller:      game.Player1,
		AbilityIndex:    1,
		HasTriggerEvent: true,
		TriggerEvent: game.GameEvent{
			Kind:        game.EventPermanentDied,
			PermanentID: courage.ObjectID,
		},
	})

	engine.resolveTopOfStack(g, &TurnLog{})

	returned := permanentForCard(g, courage.CardInstanceID)
	if returned == nil {
		t.Fatal("Enduring Courage did not return to the battlefield")
	}
	if permanentHasType(g, returned, types.Creature) {
		t.Fatal("returned Enduring Courage is still a creature")
	}
	if !permanentHasType(g, returned, types.Enchantment) {
		t.Fatal("returned Enduring Courage is not an enchantment")
	}
	content := carde.EnduringCourage.TriggeredAbilities[1].Content
	if content.IsModal() || len(content.Modes) != 1 {
		t.Fatal("Enduring Courage return ability does not use one non-modal mode")
	}
	primitive, ok := content.Modes[0].Sequence[0].Primitive.(game.PutOnBattlefield)
	if !ok {
		t.Fatal("Enduring Courage return ability does not use PutOnBattlefield")
	}
	if primitive.ContinuousEffects[0].ID != 0 {
		t.Fatal("resolution mutated Enduring Courage's shared continuous-effect template")
	}
}
