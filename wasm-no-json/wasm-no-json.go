//go:build js && wasm

package main

import (
	"syscall/js"

	"github.com/openfga/language/pkg/go/graph"
	language "github.com/openfga/language/pkg/go/transformer"
)

var (
	directEdgeStr   = "Direct Edge"
	rewriteEdgeStr  = "Rewrite Edge"
	ttuEdgeStr      = "TTU Edge"
	computedEdgeStr = "Computed Edge"

	specificTypeStr            = "SpecificType"
	specificTypeAndRelationStr = "SpecificTypeAndRelation"
	operatorNodeStr            = "OperatorNodeType"
	specificTypeWildcardStr    = "SpecificTypeWildcard"
)

func main() {
	c := make(chan struct{}, 0)

	// Register the transformation function
	js.Global().Set("transformModelDSL", js.FuncOf(transformModelDSLWrapper))

	// Keep the program running
	<-c
}

func transformModelDSLWrapper(this js.Value, inputs []js.Value) interface{} {
	// Get the DSL string from JavaScript
	dslString := inputs[0].String()

	// Transform DSL to proto
	authorizationModel, err := language.TransformDSLToProto(dslString)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Build weighted graph
	wgb := graph.NewWeightedAuthorizationModelGraphBuilder()
	weightedGraph, err := wgb.Build(authorizationModel)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Translate and return as JS object directly
	translated := translate(weightedGraph)

	return js.ValueOf(map[string]interface{}{
		"weightedGraph": translated.ToJSValue(),
	})
}

func translateNode(node *graph.WeightedAuthorizationModelNode) *WeightedAuthorizationModelNode {
	var nodeType string
	switch node.GetNodeType() {
	case graph.SpecificType:
		nodeType = specificTypeStr
	case graph.SpecificTypeAndRelation:
		nodeType = specificTypeAndRelationStr
	case graph.OperatorNode:
		nodeType = operatorNodeStr
	case graph.SpecificTypeWildcard:
		nodeType = specificTypeWildcardStr
	}
	return &WeightedAuthorizationModelNode{
		Weights:     node.GetWeights(),
		NodeType:    nodeType,
		Label:       node.GetLabel(),
		UniqueLabel: node.GetUniqueLabel(),
		Wildcards:   node.GetWildcards(),
	}
}

func translateEdge(e *graph.WeightedAuthorizationModelEdge) *WeightedAuthorizationModelEdge {
	var edgeType string
	switch e.GetEdgeType() {
	case graph.DirectEdge:
		edgeType = directEdgeStr
	case graph.RewriteEdge:
		edgeType = rewriteEdgeStr
	case graph.TTUEdge:
		edgeType = ttuEdgeStr
	case graph.ComputedEdge:
		edgeType = computedEdgeStr
	}

	return &WeightedAuthorizationModelEdge{
		Weights:          e.GetWeights(),
		EdgeType:         edgeType,
		TuplesetRelation: e.GetTuplesetRelation(),
		From:             translateNode(e.GetFrom()),
		To:               translateNode(e.GetTo()),
		Wildcards:        e.GetWildcards(),
		Conditions:       e.GetConditions(),
	}
}

func translate(weighteGraph *graph.WeightedAuthorizationModelGraph) WeightedAuthorizationModelGraph {
	graphNodes := weighteGraph.GetNodes()
	graphEdges := weighteGraph.GetEdges()

	nodes := make(map[string]*WeightedAuthorizationModelNode, len(graphNodes))
	for key, node := range graphNodes {
		nodes[key] = translateNode(node)
	}

	edges := make(map[string][]*WeightedAuthorizationModelEdge, len(graphEdges))
	for key, edgeSlice := range graphEdges {
		transformedEdges := make([]*WeightedAuthorizationModelEdge, 0, len(edgeSlice))
		for _, e := range edgeSlice {
			transformedEdges = append(transformedEdges, translateEdge(e))
		}
		edges[key] = transformedEdges
	}

	return WeightedAuthorizationModelGraph{
		Nodes: nodes,
		Edges: edges,
	}
}

// ----------------------------- Types ----------------------------------

type WeightedAuthorizationModelGraph struct {
	Edges map[string][]*WeightedAuthorizationModelEdge
	Nodes map[string]*WeightedAuthorizationModelNode
}

func (g WeightedAuthorizationModelGraph) ToJSValue() js.Value {
	nodes := make(map[string]interface{})
	for key, node := range g.Nodes {
		nodes[key] = node.ToJSValue()
	}

	edges := make(map[string]interface{})
	for key, edgeSlice := range g.Edges {
		jsEdges := make([]interface{}, len(edgeSlice))
		for i, edge := range edgeSlice {
			jsEdges[i] = edge.ToJSValue()
		}
		edges[key] = jsEdges
	}

	return js.ValueOf(map[string]interface{}{
		"nodes": nodes,
		"edges": edges,
	})
}

type WeightedAuthorizationModelEdge struct {
	Weights          map[string]int
	EdgeType         string
	TuplesetRelation string
	From             *WeightedAuthorizationModelNode
	To               *WeightedAuthorizationModelNode
	Wildcards        []string
	Conditions       []string
}

func (e *WeightedAuthorizationModelEdge) ToJSValue() js.Value {
	return js.ValueOf(map[string]interface{}{
		"weights":          e.Weights,
		"edgeType":         e.EdgeType,
		"tuplesetRelation": e.TuplesetRelation,
		"from":             e.From.ToJSValue(),
		"to":               e.To.ToJSValue(),
		"wildcards":        e.Wildcards,
		"conditions":       e.Conditions,
	})
}

type WeightedAuthorizationModelNode struct {
	Weights     map[string]int
	NodeType    string
	Label       string
	UniqueLabel string
	Wildcards   []string
}

func (n *WeightedAuthorizationModelNode) ToJSValue() js.Value {
	return js.ValueOf(map[string]interface{}{
		"weights":     n.Weights,
		"nodeType":    n.NodeType,
		"label":       n.Label,
		"uniqueLabel": n.UniqueLabel,
		"wildcards":   n.Wildcards,
	})
}
