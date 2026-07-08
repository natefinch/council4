package rules

import (
	"testing"

	cardc "github.com/natefinch/council4/mtg/cards/c"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

// vantressChoiceAgent drives Court of Vantress's upkeep trigger through the real
// choice pipeline. It records every target enumeration the engine offers (so a
// test can assert the up-to-one "other" artifact-or-enchantment candidates the
// driver would surface), selects the target permanent named by targetPick (or
// the "choose none" option when targetPick is zero), and answers the "you may"
// prompt with mayAccept.
type vantressChoiceAgent struct {
	targetPick    id.ID
	mayAccept     bool
	targetOptions [][]game.Target
	sawTarget     bool
	sawMay        bool
}

func (*vantressChoiceAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *vantressChoiceAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceTarget {
		a.sawTarget = true
		for _, option := range request.Options {
			a.targetOptions = append(a.targetOptions, option.Targets)
		}
		for i, option := range request.Options {
			if a.targetPick == 0 && len(option.Targets) == 0 {
				return []int{i}
			}
			if a.targetPick != 0 && len(option.Targets) == 1 && option.Targets[0].PermanentID == a.targetPick {
				return []int{i}
			}
		}
		return []int{0}
	}
	if request.Kind == game.ChoiceMay {
		a.sawMay = true
		if a.mayAccept {
			return []int{1}
		}
		return []int{0}
	}
	return request.DefaultSelection
}

// offersTarget reports whether the recorded target enumeration includes a single
// choice for the given permanent.
func (a *vantressChoiceAgent) offersTarget(permanentID id.ID) bool {
	for _, targets := range a.targetOptions {
		if len(targets) == 1 && targets[0].PermanentID == permanentID {
			return true
		}
	}
	return false
}

// offersNoTarget reports whether the recorded enumeration includes the empty
// "choose zero" choice permitted by "up to one".
func (a *vantressChoiceAgent) offersNoTarget() bool {
	for _, targets := range a.targetOptions {
		if len(targets) == 0 {
			return true
		}
	}
	return false
}

func vantressArtifactTarget() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Artifact Target",
		Types: []types.Card{types.Artifact},
	}}
}

func vantressEnchantmentTarget() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Enchantment Target",
		Types: []types.Card{types.Enchantment},
	}}
}

// vantressUpkeepSetup holds a game staged with Court of Vantress and the two
// candidate permanents for its upkeep target.
type vantressUpkeepSetup struct {
	game        *game.Game
	engine      *Engine
	vantress    *game.Permanent
	artifact    *game.Permanent
	enchantment *game.Permanent
}

// setupVantressUpkeep stages Court of Vantress (from the registered card
// definition) on Player1's battlefield alongside one other artifact and one
// other enchantment, and sets the monarchy, so a test can then fire the real
// beginning-of-your-upkeep trigger.
func setupVantressUpkeep(monarch bool) vantressUpkeepSetup {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	setup := vantressUpkeepSetup{
		game:        g,
		engine:      engine,
		vantress:    addCombatPermanent(g, game.Player1, cardc.CourtOfVantress()),
		artifact:    addCombatPermanent(g, game.Player1, vantressArtifactTarget()),
		enchantment: addCombatPermanent(g, game.Player1, vantressEnchantmentTarget()),
	}
	g.Players[game.Player1].IsMonarch = monarch
	g.Turn.ActivePlayer = game.Player1
	return setup
}

// fireVantressUpkeep emits the beginning-of-your-upkeep turn-based event, puts
// the resulting Court of Vantress trigger on the stack (choosing targets through
// the real enumeration path via agent), and resolves it (answering the "you may"
// prompt via agent).
func fireVantressUpkeep(t *testing.T, setup vantressUpkeepSetup, agent *vantressChoiceAgent) {
	t.Helper()
	emitBeginningOfStepEvent(setup.game, game.StepUpkeep)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}
	log := TurnLog{}
	if !setup.engine.putTriggeredAbilitiesOnStackWithChoices(setup.game, agents, &log) {
		t.Fatal("Court of Vantress upkeep trigger was not put on the stack")
	}
	setup.engine.resolveTopOfStackWithChoices(setup.game, agents, &log)
}

func vantressTokenCopyName(g *game.Game) (string, bool) {
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanent.TokenDef != nil {
			return permanent.TokenDef.Name, true
		}
	}
	return "", false
}

func permanentHasUpkeepTrigger(g *game.Game, permanent *game.Permanent) bool {
	for _, ability := range permanentEffectiveAbilities(g, permanent) {
		triggered, ok := ability.(*game.TriggeredAbility)
		if ok && triggered.Trigger.Pattern.Step == game.StepUpkeep {
			return true
		}
	}
	return false
}

// TestCourtOfVantressUpkeepEnumeratesUpToOneOtherTarget proves the driver-facing
// target enumeration for the upkeep trigger: it offers each other artifact or
// enchantment and the "choose none" option permitted by "up to one", and never
// offers Court of Vantress itself ("other").
func TestCourtOfVantressUpkeepEnumeratesUpToOneOtherTarget(t *testing.T) {
	setup := setupVantressUpkeep(true)
	agent := &vantressChoiceAgent{targetPick: 0, mayAccept: false}
	emitBeginningOfStepEvent(setup.game, game.StepUpkeep)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}
	log := TurnLog{}
	if !setup.engine.putTriggeredAbilitiesOnStackWithChoices(setup.game, agents, &log) {
		t.Fatal("Court of Vantress upkeep trigger was not put on the stack")
	}

	if !agent.sawTarget {
		t.Fatal("the upkeep trigger did not enumerate any targets")
	}
	if !agent.offersNoTarget() {
		t.Error(`"up to one" must offer the choose-none option`)
	}
	if !agent.offersTarget(setup.artifact.ObjectID) {
		t.Error("target enumeration must offer the other artifact")
	}
	if !agent.offersTarget(setup.enchantment.ObjectID) {
		t.Error("target enumeration must offer the other enchantment")
	}
	if agent.offersTarget(setup.vantress.ObjectID) {
		t.Error(`"other" must exclude Court of Vantress itself`)
	}
}

// TestCourtOfVantressUpkeepMonarchCreatesTokenCopy proves the monarch branch:
// while you're the monarch, resolving the upkeep trigger lets you create a token
// that's a copy of the chosen artifact or enchantment. The not-monarch
// become-a-copy branch is gated off, so Court of Vantress does not transform.
func TestCourtOfVantressUpkeepMonarchCreatesTokenCopy(t *testing.T) {
	setup := setupVantressUpkeep(true)
	agent := &vantressChoiceAgent{targetPick: setup.artifact.ObjectID, mayAccept: true}
	fireVantressUpkeep(t, setup, agent)

	name, ok := vantressTokenCopyName(setup.game)
	if !ok {
		t.Fatal("monarch upkeep did not create a token copy of the chosen artifact")
	}
	if name != "Test Artifact Target" {
		t.Fatalf("token copy name = %q, want the chosen artifact", name)
	}
	if got := permanentEffectiveName(setup.game, setup.vantress); got != "Court of Vantress" {
		t.Fatalf("Court of Vantress transformed to %q; the become-a-copy branch must be gated off for the monarch", got)
	}
}

// TestCourtOfVantressUpkeepMonarchMayDecline proves the token-copy branch is a
// "you may": a monarch who declines the optional effect creates no token.
func TestCourtOfVantressUpkeepMonarchMayDecline(t *testing.T) {
	setup := setupVantressUpkeep(true)
	agent := &vantressChoiceAgent{targetPick: setup.artifact.ObjectID, mayAccept: false}
	fireVantressUpkeep(t, setup, agent)

	if !agent.sawMay {
		t.Fatal("the monarch was never offered the optional token copy")
	}
	if _, ok := vantressTokenCopyName(setup.game); ok {
		t.Fatal("declining the optional effect must not create a token copy")
	}
}

// TestCourtOfVantressUpkeepMonarchNoTargetNoOps proves the "up to one" target is
// optional: a monarch who chooses no target resolves the trigger harmlessly with
// no token created.
func TestCourtOfVantressUpkeepMonarchNoTargetNoOps(t *testing.T) {
	setup := setupVantressUpkeep(true)
	agent := &vantressChoiceAgent{targetPick: 0, mayAccept: true}
	fireVantressUpkeep(t, setup, agent)

	if _, ok := vantressTokenCopyName(setup.game); ok {
		t.Fatal("choosing no target must not create a token copy")
	}
	if got := permanentEffectiveName(setup.game, setup.vantress); got != "Court of Vantress" {
		t.Fatalf("Court of Vantress unexpectedly transformed to %q with no target", got)
	}
}

// TestCourtOfVantressUpkeepNotMonarchBecomesCopyRetainingAbility proves the
// not-monarch branch: while you're not the monarch, resolving the upkeep trigger
// lets Court of Vantress become a copy of the chosen permanent, except it keeps
// its own upkeep ability. The monarch token-copy branch is gated off, so no
// token is created.
func TestCourtOfVantressUpkeepNotMonarchBecomesCopyRetainingAbility(t *testing.T) {
	setup := setupVantressUpkeep(false)
	agent := &vantressChoiceAgent{targetPick: setup.enchantment.ObjectID, mayAccept: true}
	fireVantressUpkeep(t, setup, agent)

	if _, ok := vantressTokenCopyName(setup.game); ok {
		t.Fatal("the monarch token-copy branch must be gated off when not the monarch")
	}
	if got := permanentEffectiveName(setup.game, setup.vantress); got != "Test Enchantment Target" {
		t.Fatalf("Court of Vantress effective name = %q, want the copied enchantment", got)
	}
	if !permanentHasUpkeepTrigger(setup.game, setup.vantress) {
		t.Fatal(`the copy must retain Court of Vantress's own upkeep ability ("except it has this ability")`)
	}
}

// TestCourtOfVantressUpkeepNotMonarchMayDecline proves the become-a-copy branch
// is a "you may": a non-monarch who declines leaves Court of Vantress unchanged.
func TestCourtOfVantressUpkeepNotMonarchMayDecline(t *testing.T) {
	setup := setupVantressUpkeep(false)
	agent := &vantressChoiceAgent{targetPick: setup.enchantment.ObjectID, mayAccept: false}
	fireVantressUpkeep(t, setup, agent)

	if !agent.sawMay {
		t.Fatal("the non-monarch was never offered the optional become-a-copy")
	}
	if got := permanentEffectiveName(setup.game, setup.vantress); got != "Court of Vantress" {
		t.Fatalf("declining the optional effect must leave Court of Vantress unchanged, got %q", got)
	}
}
