## package graphs - Go lang library that provides mathematical graph-theory algorithms
###import "github.com/alonsovidales/go_graph"

[![GoDoc](https://godoc.org/github.com/alonsovidales/go_graph?status.png)](https://godoc.org/github.com/alonsovidales/go_graph)


This library implements the next graph algorithms:

* **BFS:** Breadth-first search. Can be used to find the shortest path between a vertex and the others: <http://en.wikipedia.org/wiki/Breadth-first_search>
* **DFS:** Depth-first search. As Breadth-first search, but using recursion, a common case of use could be search all the connected vertices to a given one< <http://en.wikipedia.org/wiki/Depth-first_search>
* **Connected component:** Using the BFS implementation searchs for all the sets of connected vertices: <http://en.wikipedia.org/wiki/Connected_component_%28graph_theory%29>
* **Eulerian Cycle:** Calculates a cycle that starting and ending on a vertex walks through all the edges on the graph: <http://en.wikipedia.org/wiki/Eulerian_path>
* **Eulerian Path:** As the Eulerian cycle, calculates a path that starting on a vertex and ending on another one walks through all the edges on the graph: <http://en.wikipedia.org/wiki/Eulerian_path>
* **Hamilton Path / Tour:** Calculates a path that visits each vertex exactly once. A same origin and destination can be specified in order to calculate a Hamilton tour: <http://en.wikipedia.org/wiki/Hamiltonian_path>
* **MST:** Minimum Spanning Tree. Calculates the tree of edges who connects all the vertices with a minimun cost using the Kruskal's algorithm: <http://en.wikipedia.org/wiki/Kruskal%27s_algorithm>
* **Min Cut Max Flow:** Ford-Fulkerson algorithm. Min Cut Max Flow. Used to calculate the min number of edges to be removed in order to get two disjoint sets of vertices, and the max capacity between two vertices of a graph: <http://en.wikipedia.org/wiki/Ford%E2%80%93Fulkerson_algorithm>
* **shortest path:** Bellman-Ford. Algorithm used to find the shortest path between two vertices: <http://en.wikipedia.org/wiki/Bellman%E2%80%93Ford_algorithm>
* **Kosaraju-Sharir's algorithm:** Strongly Connected Components. Search for all the strongly connected components in a graph: <http://en.wikipedia.org/wiki/Kosaraju%27s_algorithm>
* **Topological Order:** Used to calculate the topological order of a graph: <http://en.wikipedia.org/wiki/Topological_order>

Author: Alonso Vidales <alonso.vidales@tras2.es>

Use of this source code is governed by a BSD-style. These programs and documents are distributed without any warranty, express or implied. All use of these programs is entirely at the user's own risk.

For further information about this libraries, please visit the online documentation on: <http://godoc.org/github.com/alonsovidales/go_graph>
