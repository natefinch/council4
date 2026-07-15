package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// estridGrantedUpkeepBlink builds the printed self-blink upkeep ability Estrid's
// Invocation's "except it has \"At the beginning of your upkeep, you may exile
// this enchantment. If you do, return it to the battlefield under its owner's
// control.\"" rider grants to the copy: the exile is Optional and publishes its
// result under "if-you-do", and the gated return brings the same card back under
// its owner's control.
func estridGrantedUpkeepBlink() *game.TriggeredAbility {
	const link = game.LinkedKey("estrid-blink")
	return &game.TriggeredAbility{
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
					Primitive:     game.Exile{Object: game.SourceCardPermanentReference(), ExileLinkedKey: link},
					Optional:      true,
					PublishResult: game.ResultKey("if-you-do"),
				},
				{
					Primitive: game.PutOnBattlefield{Source: game.LinkedBattlefieldSource(link)},
					ResultGate: opt.Val(game.InstructionResultGate{
						Key:       game.ResultKey("if-you-do"),
						Succeeded: game.TriTrue,
					}),
				},
			},
		}.Ability(),
	}
}

// estridReplacement builds the Estrid's Invocation enters-as-copy replacement:
// the optional "enter as a copy of an enchantment you control" copy carrying the
// granted upkeep self-blink ability as a copiable rider.
func estridReplacement() game.ReplacementAbility {
	return game.EntersAsCopyWithAddedAbilities(
		game.EntersAsCopyReplacement(
			"You may have this enchantment enter as a copy of an enchantment you control, except it has \"At the beginning of your upkeep, you may exile this enchantment. If you do, return it to the battlefield under its owner's control.\"",
			&game.Selection{RequiredTypes: []types.Card{types.Enchantment}, Controller: game.ControllerYou},
			true, false, nil, false, nil, nil,
		),
		estridGrantedUpkeepBlink(),
	)
}

// addEstridPermanent puts an Estrid's Invocation permanent on the battlefield
// owned by owner but controlled by controller, so the granted ability's "return
// under its owner's control" can be distinguished from the controller.
func addEstridPermanent(g *game.Game, owner, controller game.PlayerID) *game.Permanent {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID: cardID,
		Def: &game.CardDef{CardFace: game.CardFace{
			Name:  "Estrid's Invocation",
			Types: []types.Card{types.Enchantment},
		}},
		Owner: owner,
	}
	permanent := &game.Permanent{
		ObjectID:       g.IDGen.Next(),
		CardInstanceID: cardID,
		Owner:          owner,
		Controller:     controller,
	}
	g.Battlefield = append(g.Battlefield, permanent)
	return permanent
}

func addVanillaEnchantment(g *game.Game, controller game.PlayerID, name string) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:  name,
		Types: []types.Card{types.Enchantment},
	}})
}

// TestEstridEntersAsCopyGrantsUpkeepBlinkAbility proves the copy takes the chosen
// controlled enchantment's characteristics and additionally carries the granted
// upkeep self-blink ability from the "except it has \"...\"" rider (CR 706.2: the
// granted ability is a copiable characteristic of the copy).
func TestEstridEntersAsCopyGrantsUpkeepBlinkAbility(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addVanillaEnchantment(g, game.Player1, "Oblivion Ring")
	estrid := addEstridPermanent(g, game.Player1, game.Player1)

	accept := &sequencedChoiceAgent{choices: [][]int{{1}}}
	ctx := enterBattlefieldContext{
		engine: NewEngine(nil),
		agents: [game.NumPlayers]PlayerAgent{accept, accept},
		log:    &TurnLog{},
	}
	replacement := estridReplacement()
	applyEntersAsCopy(ctx, g, estrid, &replacement.Replacement)

	if got := permanentEffectiveName(g, estrid); got != "Oblivion Ring" {
		t.Fatalf("effective name = %q, want copied Oblivion Ring", got)
	}
	if !hasGrantedUpkeepBlink(g, estrid) {
		t.Fatal("copy did not carry the granted upkeep self-blink ability")
	}
}

// TestEstridEntersAsCopyDeclineOmitsGrantedAbility proves declining the optional
// copy leaves the permanent as itself and grants no ability: the granted rider
// only rides on the chosen copy.
func TestEstridEntersAsCopyDeclineOmitsGrantedAbility(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addVanillaEnchantment(g, game.Player1, "Oblivion Ring")
	estrid := addEstridPermanent(g, game.Player1, game.Player1)

	decline := &sequencedChoiceAgent{choices: [][]int{{0}}}
	ctx := enterBattlefieldContext{
		engine: NewEngine(nil),
		agents: [game.NumPlayers]PlayerAgent{decline, decline},
		log:    &TurnLog{},
	}
	replacement := estridReplacement()
	applyEntersAsCopy(ctx, g, estrid, &replacement.Replacement)

	if got := permanentEffectiveName(g, estrid); got != "Estrid's Invocation" {
		t.Fatalf("effective name = %q, want Estrid's Invocation after declining", got)
	}
	if hasGrantedUpkeepBlink(g, estrid) {
		t.Fatal("declined copy must not carry the granted upkeep ability")
	}
}

// TestEstridGrantedBlinkReturnsUnderOwnerControl drives the granted upkeep
// ability on an Estrid permanent owned by Player2 but controlled by Player1:
// accepting the optional exile blinks the permanent and the gated return brings
// the same card back as a new object under its owner's (Player2's) control.
func TestEstridGrantedBlinkReturnsUnderOwnerControl(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	estrid := addEstridPermanent(g, game.Player2, game.Player1)
	cardID := estrid.CardInstanceID

	obj := &game.StackObject{
		Kind:          game.StackTriggeredAbility,
		SourceID:      estrid.ObjectID,
		SourceCardID:  estrid.CardInstanceID,
		Controller:    game.Player1,
		InlineTrigger: estridGrantedUpkeepBlink(),
	}
	sourceDef, _ := stackObjectSourceDef(g, obj)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: true}}
	engine.resolveTriggeredAbilityBodyWithChoices(g, obj, sourceDef, obj.InlineTrigger, agents, &TurnLog{})

	if _, ok := permanentByObjectID(g, estrid.ObjectID); ok {
		t.Fatal("accepting the optional exile must remove the original permanent object")
	}
	var returned *game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == cardID {
			returned = permanent
		}
	}
	if returned == nil {
		t.Fatal("blinked enchantment card did not return to the battlefield")
	}
	if returned.ObjectID == estrid.ObjectID {
		t.Fatal("returned permanent reused the original object identity; re-entry must be a fresh object")
	}
	if returned.Owner != game.Player2 || returned.Controller != game.Player2 {
		t.Fatalf("returned owner/controller = %v/%v, want Player2 under its owner's control", returned.Owner, returned.Controller)
	}
}

// TestEstridGrantedBlinkDeclineKeepsPermanent proves declining the optional exile
// leaves the permanent in place with its original object identity: the gated
// return never runs.
func TestEstridGrantedBlinkDeclineKeepsPermanent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	estrid := addEstridPermanent(g, game.Player1, game.Player1)

	obj := &game.StackObject{
		Kind:          game.StackTriggeredAbility,
		SourceID:      estrid.ObjectID,
		SourceCardID:  estrid.CardInstanceID,
		Controller:    game.Player1,
		InlineTrigger: estridGrantedUpkeepBlink(),
	}
	sourceDef, _ := stackObjectSourceDef(g, obj)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: false}}
	engine.resolveTriggeredAbilityBodyWithChoices(g, obj, sourceDef, obj.InlineTrigger, agents, &TurnLog{})

	remaining, ok := permanentByObjectID(g, estrid.ObjectID)
	if !ok {
		t.Fatal("declining the optional exile must leave the original permanent on the battlefield")
	}
	if remaining.ObjectID != estrid.ObjectID {
		t.Fatalf("permanent object identity changed to %v, want unchanged %v", remaining.ObjectID, estrid.ObjectID)
	}
}

func hasGrantedUpkeepBlink(g *game.Game, permanent *game.Permanent) bool {
	for _, ability := range effectivePermanentValues(g, permanent).abilities {
		trigger, ok := ability.(*game.TriggeredAbility)
		if !ok {
			continue
		}
		if trigger.Trigger.Type == game.TriggerAt &&
			trigger.Trigger.Pattern.Event == game.EventBeginningOfStep &&
			trigger.Trigger.Pattern.Step == game.StepUpkeep {
			return true
		}
	}
	return false
}
