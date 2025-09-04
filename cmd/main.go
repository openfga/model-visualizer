package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/openfga/language/pkg/go/graph"
	language "github.com/openfga/language/pkg/go/transformer"
)

const (
	defaultPort = 8080
)

func main() {
	port := defaultPort
	if os.Getenv("PORT") != "" {
		p, err := strconv.Atoi(os.Getenv("PORT"))
		if err != nil {
			log.Fatal("Error: cannot parse PORT to int")
		}
		port = p
	}

	http.Handle("/", http.FileServer(http.Dir("./ui")))
	http.HandleFunc("/transform", transformModelDSLToWeightedGraph)
	fmt.Printf("Model analyzer UI: http://localhost:%d\n", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

type Request struct {
	AuthorizationModelDSL string `json:"authorizationModelDSL"`
}

type Response struct {
	WeightedGraph interface{} `json:"weightedGraph"`
}

func transformModelDSLToWeightedGraph(w http.ResponseWriter, r *http.Request) {
	var req Request
	json.NewDecoder(r.Body).Decode(&req)

	authorizationModel, err := language.TransformDSLToProto(req.AuthorizationModelDSL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid authorization model DSL: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	wgb := graph.NewWeightedAuthorizationModelGraphBuilder()
	weightedGraph, err := wgb.Build(authorizationModel)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error building weighted graph: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	translated := translate(weightedGraph)
	json.NewEncoder(w).Encode(Response{
		WeightedGraph: translated,
	})
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
