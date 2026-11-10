package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Source review: pinned Party Dude Oracle ID 8df7a58c-053f-4ead-a778-2747718e5f10,
// printing d27b6f2a-84df-4097-a66a-8e463db47f58; all 11 Scryfall rulings retrieved
// 2026-11-10 (including that X is determined on resolution and Class abilities
// remain when leveling). Official WotC Comprehensive Rules TXT, effective
// 2026-11-06, https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt:
// 601.2, 602.2, 107.3, 111.1-2/111.10b, 115.1d, 508.3b, 603.2c/603.4,
// 611.2a, 717.2a/717.3; these govern cost payment, Food ownership, target
// choice, attack coalescing, resolution-time X, temporary P/T and level gates.
func TestPartyDudeOwnText(t *testing.T) {
	g, engine, class := castClassForTest(t, "8df7a58c-053f-4ead-a778-2747718e5f10", mana.G, 0)
	if class.ClassLevel != 1 {
		t.Fatalf("Party Dude entered at level %d, want 1", class.ClassLevel)
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("Party Dude did not trigger when it entered")
	}
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})
	for _, player := range []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4} {
		if got := countTokenNamed(g, "Food", player); got != 1 {
			t.Fatalf("player %v has %d Food tokens after Party Dude entered, want 1", player, got)
		}
	}
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanentEffectiveName(g, permanent) == "Food" &&
			(!permanentHasType(g, permanent, types.Artifact) || !permanentHasSubtype(g, permanent, types.Food) ||
				len(permanentEffectiveColors(g, permanent)) != 0) {
			t.Fatal("Party Dude's Food token is not a colorless Food artifact")
		}
	}

	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Party Dude draw"}})
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	activateClassLevelForTest(t, g, engine, class, 0)
	if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("level 2 activation left %d mana, want exact {1}{G} payment", got)
	}

	opponentArtifact := enterClassTestPermanent(t, g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Opponent Artifact", Types: []types.Card{types.Artifact},
	}})
	if !movePermanentsToZoneSimultaneously(g, []*game.Permanent{opponentArtifact}, zone.Graveyard) {
		t.Fatal("moving the opponent's artifact to its graveyard failed")
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("Party Dude did not trigger for an opponent artifact put into a graveyard")
	}
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size after Party Dude's level-2 trigger = %d, want 1", got)
	}

	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 4)
	activateClassLevelForTest(t, g, engine, class, 1)
	if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("level 3 activation left %d mana, want exact {4}{G} payment", got)
	}

	attackers := []*game.Permanent{
		addCombatCreaturePermanentWithPower(g, game.Player1, 2),
		addCombatCreaturePermanentWithPower(g, game.Player1, 2),
	}
	declareClassAttackers(t, g, engine, attackers...)
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("Party Dude did not trigger when its opponent was attacked")
	}
	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("Party Dude attack triggers = %d, want one for two attackers against one opponent", got)
	}

	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Party Dude resolution card"}})
	agents := simPassPolicies()
	agents[game.Player1] = firstTargetChoiceAgent{}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	changed := 0
	for _, attacker := range attackers {
		if power := effectivePower(g, attacker); power == 4 && effectiveToughnessMust(t, g, attacker) == 4 {
			changed++
		} else if power != 2 || effectiveToughnessMust(t, g, attacker) != 2 {
			t.Fatalf("attacker stats = %d/%d, want either 4/4 or 2/2", power, effectiveToughnessMust(t, g, attacker))
		}
	}
	if changed != 1 {
		t.Fatalf("attackers receiving Party Dude's +X/+X = %d, want exactly one with X=2", changed)
	}
	engine.runCleanupStep(g, simPassPolicies(), &TurnLog{})
	for _, attacker := range attackers {
		if power := effectivePower(g, attacker); power != 2 || effectiveToughnessMust(t, g, attacker) != 2 {
			t.Fatalf("attacker stats after cleanup = %d/%d, want 2/2", power, effectiveToughnessMust(t, g, attacker))
		}
	}
}

// Source review: pinned Sengir Autocrat Oracle ID 0561043a-3896-4bb1-8ff5-ee674ddfea2a,
// printing 83fa8a17-ff86-465e-9092-0968081f4e28; both current Scryfall rulings
// retrieved 2026-11-10 were reviewed (all Serf tokens are exiled; leaving with
// its entry trigger pending resolves the leave trigger first, then creates the
// three Serfs). Official WotC Comprehensive Rules TXT, effective 2026-11-06,
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt:
// 111.7-8, 400.7, 603.3, and 603.6c govern token cessation, trigger ordering,
// and leaves-the-battlefield abilities.
func TestSengirAutocratOwnText(t *testing.T) {
	g, engine, autocrat := castClassForTest(t, "0561043a-3896-4bb1-8ff5-ee674ddfea2a", mana.B, 3)
	if permanentEffectiveName(g, autocrat) != "Sengir Autocrat" {
		t.Fatalf("cast permanent = %q, want Sengir Autocrat", permanentEffectiveName(g, autocrat))
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("Sengir Autocrat did not trigger when it entered")
	}
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})
	if got := countTokenNamed(g, "Serf", game.Player1); got != 3 {
		t.Fatalf("Sengir Autocrat created %d Serf tokens, want 3", got)
	}
	assertSerfTokens(t, g, game.Player1, 3)

	def, err := cards.NewDefaultRegistry().LookupOracleID("0561043a-3896-4bb1-8ff5-ee674ddfea2a")
	if err != nil {
		t.Fatalf("Sengir Autocrat registry lookup: %v", err)
	}
	enterClassTestPermanent(t, g, game.Player2, def)
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("the second Sengir Autocrat did not trigger when it entered")
	}
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})
	if got := countTokenNamed(g, "Serf", game.Player2); got != 3 {
		t.Fatalf("second Sengir Autocrat created %d Serf tokens, want 3", got)
	}
	assertSerfTokens(t, g, game.Player2, 3)

	if !movePermanentsToZoneSimultaneously(g, []*game.Permanent{autocrat}, zone.Graveyard) {
		t.Fatal("moving Sengir Autocrat off the battlefield failed")
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("Sengir Autocrat did not trigger when it left the battlefield")
	}
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})
	if got := countSerfTokens(g); got != 0 {
		t.Fatalf("Serf tokens remaining after Sengir Autocrat left = %d, want all exiled", got)
	}

	pendingAutocrat := enterClassTestPermanent(t, g, game.Player3, def)
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("the third Sengir Autocrat's entry trigger was not put on the stack")
	}
	if !movePermanentsToZoneSimultaneously(g, []*game.Permanent{pendingAutocrat}, zone.Graveyard) {
		t.Fatal("moving the third Sengir Autocrat off the battlefield failed")
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("the third Sengir Autocrat's leaves trigger was not put on the stack")
	}
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})
	if got := countTokenNamed(g, "Serf", game.Player3); got != 3 {
		t.Fatalf("Serf tokens created after the source left with its entry trigger pending = %d, want 3", got)
	}
	assertSerfTokens(t, g, game.Player3, 3)
}

// Source review: pinned White Glove Gourmand Oracle ID 3771c0d7-cea3-4c03-933d-f236f32ef2db,
// printing c5bf1f6c-4653-4734-9bd7-af689a1b679d; all five current Scryfall
// rulings retrieved 2026-11-10 were reviewed, including that an earlier Human
// death counts and the condition is checked as the end step begins. Official
// WotC Comprehensive Rules TXT, effective 2026-11-06,
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt:
// 601.2, 111.1-2/111.10b, 603.2b, 603.4, and 700.4 require the cast cost,
// token ownership and standard Food, the step-start/intervening-if checks, and
// a Human put into a graveyard from the battlefield.
func TestWhiteGloveGourmandOwnText(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.ActivePlayer = game.Player1
	earlierHuman := enterClassTestPermanent(t, g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Earlier Human", Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Human},
	}})
	if !movePermanentsToZoneSimultaneously(g, []*game.Permanent{earlierHuman}, zone.Graveyard) {
		t.Fatal("moving the earlier Human to its owner's graveyard failed")
	}

	def, err := cards.NewDefaultRegistry().LookupOracleID("3771c0d7-cea3-4c03-933d-f236f32ef2db")
	if err != nil {
		t.Fatalf("White Glove Gourmand registry lookup: %v", err)
	}
	cardID := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	g.Players[game.Player1].ManaPool.Add(mana.W, 1)
	g.Players[game.Player1].ManaPool.Add(mana.B, 1)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatal("casting White Glove Gourmand for {2}{W}{B} failed")
	}
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})
	gourmand := classTestPermanentForCard(g, g.CardInstances[cardID])
	if gourmand == nil {
		t.Fatal("White Glove Gourmand did not resolve onto the battlefield")
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("casting White Glove Gourmand left %d mana, want exact {2}{W}{B} payment", got)
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("White Glove Gourmand did not trigger when it entered")
	}
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})
	humans := make([]*game.Permanent, 0, 2)
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanentEffectiveName(g, permanent) == "Human Soldier" {
			humans = append(humans, permanent)
		}
	}
	if len(humans) != 2 {
		t.Fatalf("White Glove Gourmand created %d Human Soldier tokens, want 2", len(humans))
	}
	for _, human := range humans {
		if effectivePower(g, human) != 1 || effectiveToughnessMust(t, g, human) != 1 ||
			!permanentHasType(g, human, types.Creature) ||
			!permanentHasSubtype(g, human, types.Human) ||
			!permanentHasSubtype(g, human, types.Soldier) ||
			!hasEffectiveColor(g, human, color.White) {
			t.Fatal("White Glove Gourmand's token is not a 1/1 white Human Soldier creature")
		}
	}
	if !movePermanentsToZoneSimultaneously(g, []*game.Permanent{humans[0]}, zone.Graveyard) {
		t.Fatal("moving a Human Soldier token to its owner's graveyard failed")
	}

	g.Turn.Step = game.StepEnd
	emitBeginningOfStepEvent(g, game.StepEnd)
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("White Glove Gourmand did not trigger at the beginning of the end step")
	}
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})
	if got := countTokenNamed(g, "Food", game.Player1); got != 1 {
		t.Fatalf("White Glove Gourmand created %d Food tokens, want 1", got)
	}
}

func countTokenNamed(g *game.Game, name string, controller game.PlayerID) int {
	count := 0
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanent.Controller == controller && permanentEffectiveName(g, permanent) == name {
			count++
		}
	}
	return count
}

func countSerfTokens(g *game.Game) int {
	count := 0
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanentEffectiveName(g, permanent) == "Serf" {
			count++
		}
	}
	return count
}

func assertSerfTokens(t *testing.T, g *game.Game, controller game.PlayerID, want int) {
	t.Helper()
	checked := 0
	for _, permanent := range g.Battlefield {
		if !permanent.Token || permanent.Controller != controller || permanentEffectiveName(g, permanent) != "Serf" {
			continue
		}
		checked++
		if effectivePower(g, permanent) != 0 || effectiveToughnessMust(t, g, permanent) != 1 ||
			!permanentHasType(g, permanent, types.Creature) ||
			!permanentHasSubtype(g, permanent, types.Serf) ||
			!hasEffectiveColor(g, permanent, color.Black) {
			t.Fatal("Sengir Autocrat's token is not a 0/1 black Serf creature")
		}
	}
	if checked != want {
		t.Fatalf("checked %d Serf tokens for player %v, want %d", checked, controller, want)
	}
}

func hasEffectiveColor(g *game.Game, permanent *game.Permanent, want color.Color) bool {
	for _, got := range permanentEffectiveColors(g, permanent) {
		if got == want {
			return true
		}
	}
	return false
}

type firstTargetChoiceAgent struct {
	simPassPolicy
}

func (firstTargetChoiceAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceTarget && len(request.Options) > 0 {
		return []int{0}
	}
	return request.DefaultSelection
}

var _ PlayerAgent = firstTargetChoiceAgent{}
