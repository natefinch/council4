package rules

import (
	"testing"

	cardsb "github.com/natefinch/council4/mtg/cards/b"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// braidsLinkKey and braidsResultKey mirror the cardgen lowerer's linked-object
// and result keys so these runtime tests exercise the exact keys the generated
// Braids, Arisen Nightmare card publishes under.
const braidsLinkKey = game.LinkedKey("braids-sacrificed-permanent")
const braidsResultKey = game.ResultKey("braids-sacrificed")

// braidsStackObject builds a triggered-ability stack object controlled by
// Player1 with a distinct source id, so the controller's optional sacrifice and
// the following punisher share one resolution the way the real trigger does.
func braidsStackObject(g *game.Game) *game.StackObject {
	return &game.StackObject{
		Kind:         game.StackTriggeredAbility,
		Controller:   game.Player1,
		SourceID:     g.IDGen.Next(),
		SourceCardID: g.IDGen.Next(),
	}
}

// braidsSacrificeInstruction is the controller's optional sacrifice of one
// artifact/creature/enchantment/land/planeswalker, publishing the sacrificed
// permanent as a linked object and its success as a result, matching the
// lowerer's first instruction.
func braidsSacrificeInstruction() *game.Instruction {
	return &game.Instruction{
		Primitive: game.SacrificePermanents{
			Player: game.ControllerReference(),
			Amount: game.Fixed(1),
			Selection: game.Selection{RequiredTypesAny: []types.Card{
				types.Artifact,
				types.Creature,
				types.Enchantment,
				types.Land,
				types.Planeswalker,
			}},
			PublishLinked: braidsLinkKey,
		},
		Optional:      true,
		PublishResult: braidsResultKey,
	}
}

// braidsPunisherInstruction is the each-opponent shared-card-type punisher gated
// on the controller's sacrifice succeeding, matching the lowerer's second
// instruction: each opponent may sacrifice a permanent sharing a card type with
// the sacrificed permanent; each who doesn't loses 2 life and the controller
// draws a card.
func braidsPunisherInstruction() *game.Instruction {
	return &game.Instruction{
		Primitive: game.PunisherEachLoseLife{
			PlayerGroup:        game.OpponentsReference(),
			Amount:             game.Fixed(2),
			AllowSacrifice:     true,
			SacrificeSelection: game.Selection{SharesCardTypeFromLinked: braidsLinkKey},
			ControllerDrawEach: true,
		},
		ResultGate: opt.Val(game.InstructionResultGate{
			Key:       braidsResultKey,
			Succeeded: game.TriTrue,
		}),
	}
}

func stockLibraryForBraids(g *game.Game, player game.PlayerID, count int) {
	for range count {
		addLibraryCard(g, player, &game.CardDef{CardFace: game.CardFace{
			Name:  "Library Card",
			Types: []types.Card{types.Land},
		}})
	}
}

// TestBraidsPunisherMixedOutcomesAcrossFourPlayers proves the whole Braids
// sequence with a multi-type sacrifice: the controller sacrifices an
// artifact-creature, so an opponent who shares only Creature and one who shares
// only Artifact are both offered the alternative sacrifice (union match through
// last-known information after the permanent has left the battlefield). The
// opponent who sacrifices avoids the loss; the opponent who declines and the
// opponent who controls only a non-sharing land each lose 2 life and the
// controller draws one card per punished opponent.
func TestBraidsPunisherMixedOutcomesAcrossFourPlayers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	engine := NewEngine(nil)
	obj := braidsStackObject(g)

	sacrificed := addTypedPermanent(g, game.Player1, []types.Card{types.Artifact, types.Creature}, nil)
	p2creature := addTypedPermanent(g, game.Player2, []types.Card{types.Creature}, nil)
	addTypedPermanent(g, game.Player3, []types.Card{types.Land}, nil)
	p4artifact := addTypedPermanent(g, game.Player4, []types.Card{types.Artifact}, nil)
	stockLibraryForBraids(g, game.Player1, 4)

	startLife := [game.NumPlayers]int{}
	for p := range startLife {
		startLife[p] = g.Players[p].Life
	}
	startHand := g.Players[game.Player1].Hand.Size()

	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}}, // accept the sacrifice
		game.Player2: &choiceOnlyAgent{choices: [][]int{{1}}}, // sacrifice to avoid
		game.Player4: &choiceOnlyAgent{choices: [][]int{{0}}}, // decline, take the loss
	}
	log := &TurnLog{}
	engine.resolveInstructionWithChoices(g, obj, braidsSacrificeInstruction(), agents, log)
	engine.resolveInstructionWithChoices(g, obj, braidsPunisherInstruction(), agents, log)

	if _, ok := permanentByObjectID(g, sacrificed.ObjectID); ok {
		t.Fatal("controller's artifact-creature was not sacrificed")
	}
	if _, ok := permanentByObjectID(g, p2creature.ObjectID); ok {
		t.Fatal("Player2 shared Creature and sacrificed, but its creature remained")
	}
	if _, ok := permanentByObjectID(g, p4artifact.ObjectID); !ok {
		t.Fatal("Player4 declined, but its artifact was sacrificed anyway")
	}
	if got := g.Players[game.Player2].Life; got != startLife[game.Player2] {
		t.Fatalf("Player2 life = %d, want %d (sacrificed instead of losing life)", got, startLife[game.Player2])
	}
	if got := g.Players[game.Player3].Life; got != startLife[game.Player3]-2 {
		t.Fatalf("Player3 life = %d, want %d (no sharing permanent, lost 2)", got, startLife[game.Player3]-2)
	}
	if got := g.Players[game.Player4].Life; got != startLife[game.Player4]-2 {
		t.Fatalf("Player4 life = %d, want %d (declined, lost 2)", got, startLife[game.Player4]-2)
	}
	if got := g.Players[game.Player1].Hand.Size() - startHand; got != 2 {
		t.Fatalf("controller drew %d cards, want 2 (one per punished opponent)", got)
	}
}

// TestBraidsPunisherControllerDeclineSkipsOpponents proves the result gate: when
// the controller declines the optional sacrifice, nothing is published as
// succeeded, so the punisher is skipped entirely — no opponent loses life and
// the controller draws nothing.
func TestBraidsPunisherControllerDeclineSkipsOpponents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	engine := NewEngine(nil)
	obj := braidsStackObject(g)

	addTypedPermanent(g, game.Player1, []types.Card{types.Creature}, nil)
	addTypedPermanent(g, game.Player3, []types.Card{types.Land}, nil)
	stockLibraryForBraids(g, game.Player1, 2)

	startLife := g.Players[game.Player3].Life
	startHand := g.Players[game.Player1].Hand.Size()

	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}}, // decline the sacrifice
	}
	log := &TurnLog{}
	engine.resolveInstructionWithChoices(g, obj, braidsSacrificeInstruction(), agents, log)
	engine.resolveInstructionWithChoices(g, obj, braidsPunisherInstruction(), agents, log)

	if got := g.Players[game.Player3].Life; got != startLife {
		t.Fatalf("Player3 life = %d, want %d (controller declined, punisher skipped)", got, startLife)
	}
	if got := g.Players[game.Player1].Hand.Size(); got != startHand {
		t.Fatalf("controller hand = %d, want %d (no draws when sacrifice declined)", got, startHand)
	}
}

// TestBraidsPunisherNoSharingOpponentsAllLoseLife proves the shares-a-card-type
// gate rejects permanents with no shared type: when the controller sacrifices a
// creature and no opponent controls a creature, none can pay, so every opponent
// loses 2 life and the controller draws one card each.
func TestBraidsPunisherNoSharingOpponentsAllLoseLife(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	engine := NewEngine(nil)
	obj := braidsStackObject(g)

	addTypedPermanent(g, game.Player1, []types.Card{types.Creature}, nil)
	p2artifact := addTypedPermanent(g, game.Player2, []types.Card{types.Artifact}, nil)
	p3land := addTypedPermanent(g, game.Player3, []types.Card{types.Land}, nil)
	p4enchantment := addTypedPermanent(g, game.Player4, []types.Card{types.Enchantment}, nil)
	stockLibraryForBraids(g, game.Player1, 4)

	startLife := [game.NumPlayers]int{}
	for p := range startLife {
		startLife[p] = g.Players[p].Life
	}
	startHand := g.Players[game.Player1].Hand.Size()

	// No opponent shares Creature, so none is offered the sacrifice and all
	// silently take the loss without any choice prompt.
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}},
	}
	log := &TurnLog{}
	engine.resolveInstructionWithChoices(g, obj, braidsSacrificeInstruction(), agents, log)
	engine.resolveInstructionWithChoices(g, obj, braidsPunisherInstruction(), agents, log)

	for _, p := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		if got := g.Players[p].Life; got != startLife[p]-2 {
			t.Fatalf("Player%d life = %d, want %d (no sharing permanent, lost 2)", p+1, got, startLife[p]-2)
		}
	}
	for _, perm := range []*game.Permanent{p2artifact, p3land, p4enchantment} {
		if _, ok := permanentByObjectID(g, perm.ObjectID); !ok {
			t.Fatal("a non-sharing opponent's permanent was sacrificed, but none should share a card type")
		}
	}
	if got := g.Players[game.Player1].Hand.Size() - startHand; got != 3 {
		t.Fatalf("controller drew %d cards, want 3 (one per punished opponent)", got)
	}
}

// braidsCardDef builds Braids, Arisen Nightmare with the exact trigger and
// two-instruction sequence the cardgen lowerer emits, so the harness test drives
// the real end-step trigger detection and resolution path.
func braidsCardDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:       "Braids, Arisen Nightmare",
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		Subtypes:   []types.Sub{types.Nightmare},
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{
				Type: game.TriggerAt,
				Pattern: game.TriggerPattern{
					Event:      game.EventBeginningOfStep,
					Controller: game.TriggerControllerYou,
					Step:       game.StepEnd,
				},
			},
			Content: game.Mode{Sequence: []game.Instruction{
				*braidsSacrificeInstruction(),
				*braidsPunisherInstruction(),
			}}.Ability(),
		}},
	}}
}

// TestBraidsTriggerSacrificesSourceAndPunishes proves the whole card end to end
// through the real trigger: the ability fires only on the controller's end step,
// the controller sacrifices Braids itself (the source leaving the battlefield),
// and the punisher still reads Braids' Creature type through last-known
// information — an opponent who declines and one who controls only a land each
// lose 2 life while the controller draws a card each, and an opponent who
// sacrifices a creature is spared.
func TestBraidsTriggerSacrificesSourceAndPunishes(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, braidsCardDef())
	addTypedPermanent(g, game.Player2, []types.Card{types.Creature}, nil)
	addTypedPermanent(g, game.Player3, []types.Card{types.Land}, nil)
	p4creature := addTypedPermanent(g, game.Player4, []types.Card{types.Creature}, nil)
	stockLibraryForBraids(g, game.Player1, 4)

	startLife := [game.NumPlayers]int{}
	for p := range startLife {
		startLife[p] = g.Players[p].Life
	}
	startHand := g.Players[game.Player1].Hand.Size()

	g.Turn.ActivePlayer = game.Player1
	emitBeginningOfStepEvent(g, game.StepEnd)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("Braids end-step trigger was not put on the stack")
	}

	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}}, // sacrifice Braids
		game.Player2: &choiceOnlyAgent{choices: [][]int{{0}}}, // decline, take the loss
		game.Player4: &choiceOnlyAgent{choices: [][]int{{1}}}, // sacrifice to avoid
	}
	log := &TurnLog{}
	engine.resolveTopOfStackWithChoices(g, agents, log)

	if _, ok := permanentByObjectID(g, source.ObjectID); ok {
		t.Fatal("Braids sacrificed itself but remained on the battlefield")
	}
	if _, ok := permanentByObjectID(g, p4creature.ObjectID); ok {
		t.Fatal("Player4 shared Creature and sacrificed, but its creature remained")
	}
	if got := g.Players[game.Player2].Life; got != startLife[game.Player2]-2 {
		t.Fatalf("Player2 life = %d, want %d (declined, lost 2)", got, startLife[game.Player2]-2)
	}
	if got := g.Players[game.Player3].Life; got != startLife[game.Player3]-2 {
		t.Fatalf("Player3 life = %d, want %d (land-only, lost 2)", got, startLife[game.Player3]-2)
	}
	if got := g.Players[game.Player4].Life; got != startLife[game.Player4] {
		t.Fatalf("Player4 life = %d, want %d (sacrificed instead)", got, startLife[game.Player4])
	}
	if got := g.Players[game.Player1].Hand.Size() - startHand; got != 2 {
		t.Fatalf("controller drew %d cards, want 2 (one per punished opponent)", got)
	}
}

// TestBraidsTriggerSkipsOpponentEndStep proves the trigger's controller scope:
// on an opponent's end step the ability does not fire, so no one sacrifices or
// loses life.
func TestBraidsTriggerSkipsOpponentEndStep(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, braidsCardDef())

	g.Turn.ActivePlayer = game.Player2
	emitBeginningOfStepEvent(g, game.StepEnd)
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("Braids trigger fired on an opponent's end step, want controller-only")
	}
}

// TestBraidsRealCardSacrificesTokenOffersSharedType is the token regression: it
// drives the real generated Braids, Arisen Nightmare card and has the controller
// sacrifice a Treasure token (CardInstanceID == 0). Because the generated card
// binds the sacrificed permanent by ObjectID (PublishObjectBinding), the token's
// Artifact type survives through last-known information, so an opponent who
// controls an artifact is offered the shared-card-type sacrifice and takes it to
// avoid the loss. An opponent with only a land and an opponent with no permanent
// share nothing, so each loses 2 life and the controller draws one card apiece.
// Before the ObjectID binding, the token published an empty ref, no opponent
// would have shared a type, and the artifact opponent would have wrongly lost
// life instead of being offered the sacrifice.
func TestBraidsRealCardSacrificesTokenOffersSharedType(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	braids := addCombatPermanent(g, game.Player1, cardsb.BraidsArisenNightmare())
	treasure, ok := createTokenPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Treasure",
		Types: []types.Card{types.Artifact},
	}})
	if !ok {
		t.Fatal("Treasure token was not created")
	}
	if treasure.CardInstanceID != 0 {
		t.Fatalf("Treasure token CardInstanceID = %d, want 0 (real token)", treasure.CardInstanceID)
	}
	p2artifact := addTypedPermanent(g, game.Player2, []types.Card{types.Artifact}, nil)
	addTypedPermanent(g, game.Player3, []types.Card{types.Land}, nil)
	stockLibraryForBraids(g, game.Player1, 4)

	startLife := [game.NumPlayers]int{}
	for p := range startLife {
		startLife[p] = g.Players[p].Life
	}
	startHand := g.Players[game.Player1].Hand.Size()

	g.Turn.ActivePlayer = game.Player1
	emitBeginningOfStepEvent(g, game.StepEnd)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("Braids end-step trigger was not put on the stack")
	}

	agents := [game.NumPlayers]PlayerAgent{
		// Accept the optional sacrifice, then choose the Treasure token (candidate
		// index 1, after Braids itself) rather than Braids.
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}, {1}}},
		game.Player2: &choiceOnlyAgent{choices: [][]int{{1}}}, // sacrifice the artifact to avoid the loss
	}
	log := &TurnLog{}
	engine.resolveTopOfStackWithChoices(g, agents, log)

	if _, ok := permanentByObjectID(g, treasure.ObjectID); ok {
		t.Fatal("Treasure token was not sacrificed")
	}
	if _, ok := permanentByObjectID(g, braids.ObjectID); !ok {
		t.Fatal("Braids was sacrificed, but the controller chose the Treasure token")
	}
	if _, ok := permanentByObjectID(g, p2artifact.ObjectID); ok {
		t.Fatal("Player2 shared Artifact with the token and sacrificed, but its artifact remained")
	}
	if got := g.Players[game.Player2].Life; got != startLife[game.Player2] {
		t.Fatalf("Player2 life = %d, want %d (offered and sacrificed the shared artifact)", got, startLife[game.Player2])
	}
	if got := g.Players[game.Player3].Life; got != startLife[game.Player3]-2 {
		t.Fatalf("Player3 life = %d, want %d (land only, lost 2)", got, startLife[game.Player3]-2)
	}
	if got := g.Players[game.Player4].Life; got != startLife[game.Player4]-2 {
		t.Fatalf("Player4 life = %d, want %d (no permanent, lost 2)", got, startLife[game.Player4]-2)
	}
	if got := g.Players[game.Player1].Hand.Size() - startHand; got != 2 {
		t.Fatalf("controller drew %d cards, want 2 (Player3 and Player4 punished)", got)
	}
}
