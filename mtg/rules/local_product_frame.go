package rules

import "github.com/natefinch/council4/mtg/game"

// Each declared cell is fresh while this frame resolves. Restoring the outer
// cells also makes recursive/repeated invocations independent without rewriting
// references, changing persistent links, or confusing an unavailable result with
// the previous invocation's success.
func enterLocalProductFrame(g *game.Game, obj *game.StackObject, sequence []game.Instruction) func() {
	var restore []func()
	keys := make(map[string]bool)
	links := make(map[game.LinkedKey]bool)
	enterResult := func(key string) {
		if key == "" || keys[key] || obj == nil {
			return
		}
		keys[key] = true
		restore = append(restore,
			isolateProductCell(&obj.ResolutionResults, key),
			isolateProductCell(&obj.ResolutionResultObjects, key),
			isolateProductCell(&obj.ResolvedAmounts, key),
			isolateProductCell(&obj.ResolvedExcessDamage, key),
		)
	}
	for _, instruction := range sequence {
		for _, key := range instruction.LocalProducts.Results {
			enterResult(string(key))
		}
		for _, key := range instruction.LocalProducts.Links {
			if key == "" || obj == nil {
				continue
			}
			enterResult(string(key))
			if !links[key] {
				links[key] = true
				restore = append(restore, isolateProductCell(&obj.LocalLinkedProducts, key))
				if obj.LocalLinkedProducts == nil {
					obj.LocalLinkedProducts = make(map[game.LinkedKey]game.LinkedObjectKey)
				}
				scope := g.IDGen.Next()
				address := game.LinkedObjectKey{SourceID: scope, LinkID: string(key), ResolutionScope: scope}
				obj.LocalLinkedProducts[key] = address
				restore = append(restore, isolateProductCell(&g.LinkedObjects, address))
			}
		}
	}
	return func() {
		for i := len(restore) - 1; i >= 0; i-- {
			restore[i]()
		}
	}
}

func isolateProductCell[K comparable, V any](values *map[K]V, key K) func() {
	previous, exists := (*values)[key]
	delete(*values, key)
	return func() {
		if exists {
			if *values == nil {
				*values = make(map[K]V)
			}
			(*values)[key] = previous
		} else {
			delete(*values, key)
		}
	}
}
