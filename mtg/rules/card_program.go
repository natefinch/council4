package rules

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"slices"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// NewEngineWithPrograms installs an immutable catalog before any game is loaded.
// Search Engines share that same catalog, not mutable per-game execution contexts.
func NewEngineWithPrograms(rng *rand.Rand, programs cardprogram.Catalog) *Engine {
	engine := NewEngine(rng)
	engine.programs = programs
	return engine
}

// ProgramBindings exposes immutable source/version evidence for replay/reporting.
func (e *Engine) ProgramBindings(references []game.ProgramReference) ([]cardprogram.Binding, error) {
	return e.programs.Bindings(references)
}

// ValidatePrograms rejects missing/version-mismatched code before game setup.
// Immutable CardDefs may be cached; copies/modifications must use fresh definitions.
func (e *Engine) ValidatePrograms(configs [game.NumPlayers]game.PlayerConfig) error {
	for _, config := range configs {
		defs := append(slices.Clone(config.CommanderDefs()), config.Deck...)
		for _, def := range defs {
			if def == nil {
				return errors.New("deck contains a nil Card Definition")
			}
			if e.checkedProgramDefs[def] {
				continue
			}
			if err := e.programs.ValidateCard(def); err != nil {
				return err
			}
			for _, face := range []*game.CardFace{&def.CardFace, programFace(def.Back.Exists, &def.Back.Val), programFace(def.Alternate.Exists, &def.Alternate.Val)} {
				if face != nil && face.ImplementationID != "" {
					if _, installed := e.cardImplementations[face.ImplementationID]; !installed {
						return fmt.Errorf("%s: legacy implementation %s is not installed", def.Name, face.ImplementationID)
					}
				}
			}
			if e.checkedProgramDefs == nil {
				e.checkedProgramDefs = make(map[*game.CardDef]bool)
			}

			e.checkedProgramDefs[def] = true
		}
	}
	return nil
}

func programFace(exists bool, face *game.CardFace) *game.CardFace {
	if exists {
		return face
	}
	return nil
}

func (e *Engine) validateLoadedPrograms(g *game.Game) {
	if g == nil {
		panic(UnsupportedError{Reason: "cannot load a nil game"})
	}
	var configs [game.NumPlayers]game.PlayerConfig
	for _, instance := range g.CardInstances {
		configs[0].Deck = append(configs[0].Deck, instance.Def)
	}
	if err := e.ValidatePrograms(configs); err != nil {
		panic(UnsupportedError{Reason: err.Error()})
	}
	e.validatePendingPrograms(g)
}

func (e *Engine) validateProgramReferences(references []game.ProgramReference) error {
	for _, reference := range references {
		if _, err := e.programs.Lookup(reference); err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) validatePendingPrograms(g *game.Game) {
	var references []game.ProgramReference
	for i := range g.ReplacementEffects {
		if err := g.ReplacementEffects[i].ValidateProgram(); err != nil {
			panic(UnsupportedError{Reason: "pending replacement: " + err.Error()})
		}
		references = append(references, g.ReplacementEffects[i].ProgramReferences()...)
	}
	for _, pending := range g.DelayedTriggers {
		references = append(references, pending.Ability.ProgramReferences()...)
	}
	for _, pending := range g.PendingReflexiveTriggers {
		references = append(references, pending.Ability.ProgramReferences()...)
	}
	for _, obj := range g.Stack.Objects() {
		if obj.InlineTrigger != nil {
			references = append(references, obj.InlineTrigger.ProgramReferences()...)
		}
		if obj.InlineActivated != nil {
			references = append(references, obj.InlineActivated.ProgramReferences()...)
		}
		if obj.InlineLoyalty != nil {
			references = append(references, obj.InlineLoyalty.ProgramReferences()...)
		}
		for _, content := range obj.SplicedContent {
			references = append(references, content.ProgramReferences()...)
		}
		if obj.SourceTokenDef != nil {
			references = append(references, obj.SourceTokenDef.ProgramReferences()...)
		}
	}
	if err := e.validateProgramReferences(references); err != nil {
		panic(UnsupportedError{Reason: "pending ability: " + err.Error()})
	}
}

func (e *Engine) resolveProgram(g *game.Game, obj *game.StackObject, mode game.Mode, shared []game.TargetSpec, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	if g == nil || obj == nil {
		panic(UnsupportedError{Reason: "authored program requires game and stack object"})
	}
	if log == nil {
		log = &TurnLog{}
	}
	cardID, objectID := damageSourceIDs(g, obj)
	ctx := &programContext{
		engine: e, game: g, obj: obj, agents: agents, log: log, active: true,
		targets: slices.Clone(obj.Targets),
		specs:   append(slices.Clone(shared), mode.Targets...),
		source: cardprogram.Source{
			CardID: cardID, ObjectID: objectID, StackID: obj.ID, Controller: obj.Controller, X: obj.XValue,
		},
	}
	if obj.HasTriggerEvent {
		if player, known := triggeringEventPlayer(obj.TriggerEvent); known {
			ctx.source.EventPlayer = opt.Val(player)
		}
	}
	defer func() { ctx.active = false }()
	previous, _ := g.ChoiceContext()
	e.setReplacementChoiceContext(g, agents, log)
	defer g.SetChoiceContext(previous)
	if err := e.programs.Run(mode.Program, ctx); err != nil {
		panic(UnsupportedError{Reason: err.Error()})
	}
}

type programContext struct {
	engine       *Engine
	game         *game.Game
	obj          *game.StackObject
	agents       [game.NumPlayers]PlayerAgent
	log          *TurnLog
	active       bool
	source       cardprogram.Source
	targets      []game.Target
	specs        []game.TargetSpec
	found        map[game.ObjectID]game.ZoneCardSnapshot
	libraryEpoch map[game.ObjectID]uint64
}

func (c *programContext) View() cardprogram.View { return c }

func (c *programContext) Source() cardprogram.Source { return c.source }

func (c *programContext) Targets() []game.Target { return slices.Clone(c.targets) }

func (c *programContext) Execute(primitive game.Primitive) (cardprogram.Outcome, error) {
	if err := c.valid(); err != nil {
		return cardprogram.Outcome{}, err
	}
	if err := game.ValidatePrimitive(primitive, c.specs); err != nil {
		return cardprogram.Outcome{}, fmt.Errorf("invalid authored operation: %w", err)
	}
	if err := game.ValidateLinkedProducerOperation(primitive, c.obj.ResolvingLinkedProducers); err != nil {
		return cardprogram.Outcome{}, err
	}
	if provider, ok := primitive.(interface {
		ProgramReferences() []game.ProgramReference
	}); ok {
		if err := c.engine.validateProgramReferences(provider.ProgramReferences()); err != nil {
			return cardprogram.Outcome{}, err
		}
	}
	// This key is implementation-local, never a variable name card authors need
	// to manage. Each call restores the enclosing declarative receipt/products.
	const resultKey = game.ResultKey("@authored-operation")
	restore := isolateResultProducts(c.obj, string(resultKey))
	defer restore()
	resolver := newEffectResolver(c.engine, c.game, c.obj, c.agents, c.log)
	resolver.resolveInstruction(&game.Instruction{Primitive: primitive, PublishResult: resultKey})
	receipt, exists := c.obj.ResolutionResults[string(resultKey)]
	outcome := cardprogram.Outcome{Available: exists, Succeeded: receipt.Succeeded}
	for _, snapshot := range c.obj.ResolutionResultObjects[string(resultKey)] {
		outcome.Objects = append(outcome.Objects, resultCharacteristicSnapshot(snapshot))
	}
	return outcome, nil
}

func (c *programContext) Ask(request game.ChoiceRequest) ([]int, error) {
	if err := c.valid(); err != nil {
		return nil, err
	}
	if request.DecisionKind != game.DecisionRules {
		return nil, errors.New("authored choices must be rules decisions, not personal decisions")
	}
	if !isPlayerAlive(c.game, request.Player) || request.MinChoices < 0 ||
		request.MaxChoices < request.MinChoices {
		return nil, errors.New("authored choice has an invalid player or cardinality")
	}
	repeats := request.Kind == game.ChoiceDamageAllocation || request.Kind == game.ChoiceCounterAllocation ||
		request.Kind == game.ChoiceManaCombination
	if (!repeats && request.MaxChoices > len(request.Options)) ||
		(repeats && request.MinChoices > 0 && len(request.Options) == 0) {
		return nil, errors.New("authored choice exceeds available distinct options")
	}
	seen := make(map[int]bool, len(request.Options))
	for _, option := range request.Options {
		if option.Index < 0 || seen[option.Index] {
			return nil, errors.New("authored choice has ambiguous option indices")
		}
		seen[option.Index] = true
	}
	return c.engine.chooseChoice(c.game, c.agents, request, c.log), nil
}

func (c *programContext) valid() error {
	if c == nil || !c.active || c.engine == nil || c.game == nil || c.obj == nil || c.log == nil {
		return errors.New("authored execution context is missing or no longer active")
	}
	return nil
}
