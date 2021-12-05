package flowdebug

import (
	"bytes"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"regexp"
	"strings"

	"github.com/brian14708/go-flow/flowdebug/types"
)

var reNamespace = regexp.MustCompile(`[./]`)

func Graphviz(topology *types.Topology) (string, error) {
	if topology == nil {
		return "", errors.New("invalid topology")
	}

	buf := new(bytes.Buffer)
	fmt.Fprint(buf, "digraph {\n")
	fmt.Fprint(buf, "\trankdir=LR;\n")
	fmt.Fprint(buf, "\tnode [shape=record];\n")

	port2id := make(map[string]string)
	for _, node := range topology.Nodes {
		localport2id := writeGraphvizNode(buf, node)
		for k, v := range localport2id {
			port2id[fmt.Sprintf("%s:%s", node.Name, k)] = v
		}
		fmt.Fprint(buf, "\n")
	}

	for idx, conn := range topology.Connections {
		if len(conn.Source) == 1 && len(conn.Destination) == 1 {
			fmt.Fprintf(buf, "\t%s -> %s;\n", port2id[conn.Source[0]], port2id[conn.Destination[0]])
			continue
		}

		tmpPt := hash(fmt.Sprintf("tmp%d", idx))
		namespace := commonNamespace(append(conn.Source, conn.Destination...))
		fmt.Fprint(buf, "\t")
		for _, ns := range namespace {
			fmt.Fprintf(buf, "subgraph cluster_%s { ", ns)
		}
		fmt.Fprintf(buf, "%s [shape=point];", tmpPt)
		for range namespace {
			fmt.Fprint(buf, " }")
		}
		fmt.Fprint(buf, "\n")
		for _, src := range conn.Source {
			fmt.Fprintf(buf, "\t%s -> %s [arrowhead=none];\n", port2id[src], tmpPt)
		}
		for _, dst := range conn.Destination {
			fmt.Fprintf(buf, "\t%s -> %s;\n", tmpPt, port2id[dst])
		}
	}

	fmt.Fprint(buf, "}")
	return buf.String(), nil
}

func writeGraphvizNode(w io.Writer, node types.Node) (port2id map[string]string) {
	port2id = make(map[string]string)

	name := node.Name
	namespace := reNamespace.Split(name, -1)
	label := namespace[len(namespace)-1]
	namespace = namespace[:len(namespace)-1]

	fmt.Fprint(w, "\t")
	for _, ns := range namespace {
		fmt.Fprintf(w, "subgraph cluster_%s { ", ns)
		fmt.Fprintf(w, "label=\"%s\"; ", ns)
	}
	fmt.Fprintf(w, "%s [label=\"{", hash(name))

	in, out := node.InPorts, node.OutPorts
	displayPort := (len(in) != 1 || len(out) != 1)

	if displayPort {
		fmt.Fprint(w, "{")
		first := true
		for _, p := range in {
			if !first {
				fmt.Fprint(w, "|")
			} else {
				first = false
			}
			fmt.Fprintf(w, "<i%s>%s", hash(p.Name), p.Name)
			port2id[p.Name] = fmt.Sprintf("%s:i%s", hash(name), hash(p.Name))
		}
		fmt.Fprint(w, "}|")
	} else {
		for _, p := range in {
			port2id[p.Name] = hash(name)
		}
	}
	fmt.Fprintf(w, "%s", label)
	if displayPort {
		fmt.Fprint(w, "|{")
		first := true
		for _, p := range out {
			if !first {
				fmt.Fprint(w, "|")
			} else {
				first = false
			}
			fmt.Fprintf(w, "<o%s>%s", hash(p.Name), p.Name)
			port2id[p.Name] = fmt.Sprintf("%s:o%s", hash(name), hash(p.Name))
		}
		fmt.Fprint(w, "}")
	} else {
		for _, p := range out {
			port2id[p.Name] = hash(name)
		}
	}
	fmt.Fprint(w, "}\"]")
	for range namespace {
		fmt.Fprint(w, " }")
	}

	return port2id
}

func hash(s string) string {
	h := fnv.New64()
	_, err := io.WriteString(h, s)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("X%x", h.Sum(nil))
}

func commonNamespace(ports []string) []string {
	var common []string
	for _, p := range ports {
		name := strings.Split(p, ":")[0]
		namespace := reNamespace.Split(name, -1)
		namespace = namespace[:len(namespace)-1]
		if common == nil {
			common = namespace
			continue
		}
		for i, n := range common {
			if i >= len(namespace) {
				common = namespace
				break
			}
			if namespace[i] != n {
				common = common[:i]
				break
			}
		}
	}
	return common
}
