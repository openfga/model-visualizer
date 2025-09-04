# OpenFGA Model Visualizer

*In-browser visualizer for OpenFGA authorization models as a **weighted graph** which offers insights into its performance characteristics*


[](./screenshot.png)

## Getting Started

1. Run: `PORT=8080 go run ./cmd/main.go`
2. Visit: `http://localhost:8080` 


## Weighted Graph

A weighted graph assigns point values (from 1 to ∞) to nodes and edges. These weights represent the relative complexity of resolving that section of the model:

 - **Lower weights:** relatively faster, cheaper resolution
 - **Higher weights:** relatively slower, more resource-intensive resolution
 - **∞ (infinity):** recursive or cyclical behavior, resolution costs cannot be determined from the model alone

The weighted graph is primarily an internal construct, designed to help the system coordinate resolution optimizations and improve query planning. It isn't required knowledge for an OpenFGA operator but can be a useful diagnostic tool for identifying potential performance bottlenecks in an authorization model.