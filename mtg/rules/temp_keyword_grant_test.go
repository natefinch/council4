package rules

import (
	"testing"

	cardf "github.com/natefinch/council4/mtg/cards/f"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestFeignDeathGrantsDeathTriggerUntilEndOfTurn proves the temporary
// granted-quoted-ability generalization end to end with the real Feign Death
// card ("Until end of turn, target creature gains 'When this creature dies,
// return it to the battlefield ...'"): resolving the spell adds the quoted death
// trigger to the chosen creature only, and the grant expires at the end-of-turn
// cleanup.
func TestFeignDeathGrantsDeathTriggerUntilEndOfTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	bystander := addCombatCreaturePermanentWithPower(g, game.Player1, 2)

	if got := countGrantedTriggeredAbilities(g, target); got != 0 {
		t.Fatalf("granted triggered abilities before resolution = %d, want 0", got)
	}

	addImplementationSpellToStack(g, game.Player1, cardf.FeignDeath(),
		[]game.Target{game.PermanentTarget(target.ObjectID)})
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := countGrantedTriggeredAbilities(g, target); got != 1 {
		t.Fatalf("granted triggered abilities on target after resolution = %d, want 1", got)
	}
	if got := countGrantedTriggeredAbilities(g, bystander); got != 0 {
		t.Fatalf("granted triggered abilities on non-target = %d, want 0 (grant must not spread)", got)
	}

	// The until-end-of-turn grant expires at the cleanup step.
	expireCleanupDurations(g)
	if got := countGrantedTriggeredAbilities(g, target); got != 0 {
		t.Fatalf("granted triggered abilities after cleanup = %d, want 0 (grant must expire)", got)
	}
}

// TestFeignDeathGrantedTriggerReturnsTappedWithCounter drives the real Feign
// Death card end to end through the dies→trigger→stack→resolve path and proves
// its granted quoted ability honors the return riders ("return it to the
// battlefield tapped ... with a +1/+1 counter on it"): after the creature dies
// and the granted trigger resolves, the returned permanent is tapped and carries
// one +1/+1 counter. This fails on the earlier rider-dropping lowering (which
// emitted a bare PutOnBattlefield) and passes once the riders are forwarded.
func TestFeignDeathGrantedTriggerReturnsTappedWithCounter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creatureCardID := addCardInstance(g, game.Player1, vanillaCreatureDef())
	creature := &game.Permanent{
		ObjectID:       g.IDGen.Next(),
		CardInstanceID: creatureCardID,
		Owner:          game.Player1,
		Controller:     game.Player1,
		Face:           game.FaceFront,
	}
	g.Battlefield = append(g.Battlefield, creature)

	addImplementationSpellToStack(g, game.Player1, cardf.FeignDeath(),
		[]game.Target{game.PermanentTarget(creature.ObjectID)})
	engine.resolveTopOfStack(g, &TurnLog{})

	if _, ok := destroyPermanent(g, creature.ObjectID); !ok {
		t.Fatal("destroyPermanent() = false, want the granted creature to die")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("putTriggeredAbilitiesOnStack() = false, want the granted dies trigger")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	returned := permanentForCard(g, creatureCardID)
	if returned == nil {
		t.Fatal("the granted dies trigger did not return the creature to the battlefield")
	}
	if !returned.Tapped {
		t.Fatal("returned creature is untapped, want tapped (the \"tapped\" rider was dropped)")
	}
	if got := returned.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("returned creature +1/+1 counters = %d, want 1 (the counter rider was dropped)", got)
	}
}

func vanillaCreatureDef() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Test Bear",
			Types:     []types.Card{types.Creature},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}

// until-your-next-turn duration (Elspeth, Storm Slayer's "Those creatures gain
// flying until your next turn."): the keyword is granted to every creature the
// resolving player controls at resolution — and to no one else — survives the
// end-of-turn cleanup, and only expires at the start of that player's next turn.
func TestGroupKeywordGrantUntilYourNextTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	mine1 := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	mine2 := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	theirs := addCombatCreaturePermanentWithPower(g, game.Player2, 2)

	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		{Primitive: game.ApplyContinuous{
			ContinuousEffects: []game.ContinuousEffect{{
				Layer:       game.LayerAbility,
				Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
				AddKeywords: []game.Keyword{game.Flying},
			}},
			Duration: game.DurationUntilYourNextTurn,
		}},
	}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	if !hasKeyword(g, mine1, game.Flying) || !hasKeyword(g, mine2, game.Flying) {
		t.Fatal("creatures the resolving player controls did not gain flying")
	}
	if hasKeyword(g, theirs, game.Flying) {
		t.Fatal("an opponent's creature gained flying (grant must be limited to the controller's creatures)")
	}

	// End-of-turn cleanup must not touch an until-your-next-turn grant.
	expireCleanupDurations(g)
	if !hasKeyword(g, mine1, game.Flying) || !hasKeyword(g, mine2, game.Flying) {
		t.Fatal("until-your-next-turn grant expired at cleanup; it must last until the controller's next turn")
	}

	// At the start of the controller's next turn the grant expires.
	g.Turn.TurnNumber = 2
	g.Turn.ActivePlayer = game.Player1
	expireTurnStartDurations(g)
	if hasKeyword(g, mine1, game.Flying) || hasKeyword(g, mine2, game.Flying) {
		t.Fatal("until-your-next-turn grant did not expire at the start of the controller's next turn")
	}
}
