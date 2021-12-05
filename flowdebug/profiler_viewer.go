package flowdebug

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"text/template"

	"github.com/brian14708/go-flow/flowdebug/types"
)

func (p *Profiler) index(w http.ResponseWriter, r *http.Request) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	type state struct {
		ID    string `json:"id"`
		State string `json:"state"`
	}
	var states []state

	for k, v := range p.graphs {
		s := state{
			ID: k,
		}
		if v.err == nil {
			s.State = "RUNNING"
		} else if *v.err == nil {
			s.State = "STOPPED"
		} else {
			s.State = fmt.Sprintf("ERROR (%s)", (*v.err).Error())
		}
		states = append(states, s)
	}
	sort.Slice(states, func(a, b int) bool {
		return states[a].ID < states[b].ID
	})

	accept := r.Header.Get("Accept")
	if strings.Contains(accept, "application/json") {
		if err := json.NewEncoder(w).Encode(states); err != nil {
			panic(err)
		}
	} else {
		if err := indexTmpl.Execute(w, states); err != nil {
			panic(err)
		}
	}
}

var indexTmpl = template.Must(template.New("index").Parse(`<!doctype html>
<html>
<head>
<title>go-flow profiler</title>
<style>
* {
	margin: 0;
	padding: 0;
	box-sizing: border-box;
}
table {
	margin: 10px;
	border-spacing: 0;
	font-family: monospace;
}
thead {
	font-weight: bold;
}
td {
	padding: 3px 10px;
}
</style>
</head>
<body>
<table>
<thead>
	<td>ID</td>
	<td>Operations</td>
	<td>State</td>
</thead>
{{range .}}
	<tr>
	<td>{{.ID}}</td>
	<td><a href="viewer/dot/{{.ID | urlquery}}">DOT</a> | <a href="viewer/profiler/profiler.html?id={{.ID | urlquery}}">PROFILER</a></td>
	<td>{{.State}}</td>
	</tr>
{{end}}
</table>
</body>
</html>
`))

func (p *Profiler) viewerDot(id string, w http.ResponseWriter, r *http.Request) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	g, ok := p.graphs[id]
	if !ok {
		http.NotFound(w, r)
		return
	}

	t := new(types.Topology)
	err := json.Unmarshal(g.topology, &t)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	str, err := Graphviz(t)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	state := struct {
		Title string
		Dot   string
	}{
		Title: id,
		Dot:   str,
	}
	if err := viewerDotTmpl.Execute(w, state); err != nil {
		panic(err)
	}
}

var viewerDotTmpl = template.Must(template.New("viewer.dot").Parse(`<!doctype html>
<html>
<head>
<title>go-flow - {{.Title | html}}</title>
</head>
<body>
<pre id="dot">
{{.Dot | html}}
</pre>
<script src="../profiler/assets/hpcc.min.js"></script>
<script>
document.addEventListener("DOMContentLoaded", function() {
	var hpccWasm = window["@hpcc-js/wasm"];
	hpccWasm.graphviz.layout(document.getElementById("dot").innerText, "svg", "dot").then(svg => {
		document.body.innerHTML = svg;
		var r = document.getElementsByTagName('script');
		for (var i = (r.length - 1); i >= 0; i--) {
			r[i].parentNode.removeChild(r[i]);
		}
	});
});
</script>
</body>
</html>
`))
