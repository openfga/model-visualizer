package main

import (
	"encoding/json"
	"fmt"
	"os"

	// Remove: "syscall/js"

	"github.com/openfga/language/pkg/go/graph"
	language "github.com/openfga/language/pkg/go/transformer"
)

// ...existing var declarations...

func main() {
	// Replace WASM-specific code with CLI or server logic
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run wasm.go '<DSL string>'")
		os.Exit(1)
	}

	dslString := os.Args[1]
	result := transformModelDSL(dslString)

	if errorMsg, hasError := result["error"]; hasError {
		fmt.Printf("Error: %s\n", errorMsg)
		os.Exit(1)
	}

	fmt.Println(result["weightedGraph"])
}

func transformModelDSL(dslString string) map[string]interface{} {
	// Transform DSL to proto
	authorizationModel, err := language.TransformDSLToProto(dslString)
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	// Build weighted graph
	wgb := graph.NewWeightedAuthorizationModelGraphBuilder()
	weightedGraph, err := wgb.Build(authorizationModel)
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	// Translate and return
	translated := translate(weightedGraph)
	jsonBytes, _ := json.Marshal(translated)

	return map[string]interface{}{
		"weightedGraph": string(jsonBytes),
	}
}

func translateNode(node *graph.WeightedAuthorizationModelNode) *WeightedAuthorizationModelNode {
	var nodeType string
	switch node.GetNodeType() {
	case graph.SpecificType:
		nodeType = "SpecificType"
	case graph.SpecificTypeAndRelation:
		nodeType = "SpecificTypeAndRelation"
	case graph.OperatorNode:
		nodeType = "OperatorNodeType"
	case graph.SpecificTypeWildcard:
		nodeType = "SpecificTypeWildcard"
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
		edgeType = "Direct Edge"
	case graph.RewriteEdge:
		edgeType = "Rewrite Edge"
	case graph.TTUEdge:
		edgeType = "TTU Edge"
	case graph.ComputedEdge:
		edgeType = "Computed Edge"
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
	nodes := map[string]*WeightedAuthorizationModelNode{}
	for key, node := range weighteGraph.GetNodes() {
		nodes[key] = translateNode(node)
	}

	edges := map[string][]*WeightedAuthorizationModelEdge{}
	for key, edgeSlice := range weighteGraph.GetEdges() {

		transformedEdges := []*WeightedAuthorizationModelEdge{}
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

type NodeType int64

const (
	SpecificType            NodeType = 0
	SpecificTypeAndRelation NodeType = 1
	OperatorNode            NodeType = 2
	SpecificTypeWildcard    NodeType = 3

	UnionOperator        = "union"
	IntersectionOperator = "intersection"
	ExclusionOperator    = "exclusion"
)

type WeightedAuthorizationModelGraph struct {
	Edges map[string][]*WeightedAuthorizationModelEdge
	Nodes map[string]*WeightedAuthorizationModelNode
}

type EdgeType int64

const (
	DirectEdge   EdgeType = 0
	RewriteEdge  EdgeType = 1
	TTUEdge      EdgeType = 2
	ComputedEdge EdgeType = 3
)

type WeightedAuthorizationModelEdge struct {
	Weights          map[string]int
	EdgeType         string
	TuplesetRelation string
	From             *WeightedAuthorizationModelNode
	To               *WeightedAuthorizationModelNode
	Wildcards        []string
	Conditions       []string
}

type WeightedAuthorizationModelNode struct {
	Weights     map[string]int
	NodeType    string
	Label       string
	UniqueLabel string
	Wildcards   []string
}
