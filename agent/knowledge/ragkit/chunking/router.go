package chunking

import (
	"context"

	"github.com/cloudwego/eino/schema"
)

// Matcher 决定一个策略是否处理该文档。
type Matcher func(req Request) bool

// RoutedStrategy 是一条路由规则。
type RoutedStrategy struct {
	Name     string
	Match    Matcher
	Strategy Strategy
}

// StrategyRouter 按顺序匹配，首个命中者切；未命中走 default。
type StrategyRouter struct {
	defaultStrategy Strategy
	routes          []RoutedStrategy
}

func NewStrategyRouter(defaultStrategy Strategy, routes ...RoutedStrategy) *StrategyRouter {
	return &StrategyRouter{defaultStrategy: defaultStrategy, routes: routes}
}

// Split 路由切块，chunk metadata 记录 chunking_route。
func (r *StrategyRouter) Split(ctx context.Context, req Request) ([]*schema.Document, error) {
	chosen := r.defaultStrategy
	routeName := chosen.Name()
	for _, rt := range r.routes {
		if rt.Match != nil && rt.Match(req) {
			chosen = rt.Strategy
			routeName = rt.Name
			break
		}
	}
	chunks, err := chosen.Split(ctx, req)
	if err != nil {
		return nil, err
	}
	for _, c := range chunks {
		if c.MetaData == nil {
			c.MetaData = map[string]any{}
		}
		c.MetaData["chunking_route"] = routeName
	}
	return chunks, nil
}

func (r *StrategyRouter) Name() string { return "router" }
