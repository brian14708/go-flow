function escapeHtml(unsafe) {
  return unsafe
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#039;");
}

function edgeDesc(dataType, conn) {
  let html =
    "<b>Connection</b><br>" +
    "<b>Data Type:</b> " +
    escapeHtml(dataType) +
    "<br>" +
    "<b>Channel Capacity:</b> " +
    conn.capacity +
    "<br>";

  html += "<br><b>FROM:</b><br>";
  conn.src.forEach(function (s) {
    html += "&nbsp;&nbsp;" + s + "<br>";
  });

  html += "<br><b>TO:</b><br>";
  conn.dst.forEach(function (d) {
    html += "&nbsp;&nbsp;" + d + "<br>";
  });
  return html;
}

function nodeDesc(n, connections) {
  let html =
    "<b>Node Name:</b> " +
    escapeHtml(n.name) +
    "<br><br>" +
    "<b>Type:</b> " +
    escapeHtml(n.type) +
    "<br>";
  if (n.description) {
    html += "<br><b>Description:</b><br>";
    descs = n.description.split(/[;\n]/);
    for (let idx in descs) {
      html += "&nbsp;&nbsp;" + escapeHtml(descs[idx]) + "<br>";
    }
  }
  if (n.in_ports !== null) {
    html += "<br><b>Ports - IN:</b><br>";
    for (let i = 0; i < n.in_ports.length; i++) {
      let port = n.name + ":" + n.in_ports[i].name;
      let found = false;
      for (let j = 0; j < connections.length; j++) {
        if (connections[j].dst.includes(port)) {
          found = true;
          break;
        }
      }
      html += "&nbsp;&nbsp;";
      if (!found) {
        html += "<strong>(UNCONNECTED)</strong> ";
      }
      html +=
        escapeHtml(n.in_ports[i].name) +
        " - " +
        escapeHtml(n.in_ports[i].type) +
        "<br>";
    }
  }
  if (n.out_ports !== null) {
    html += "<br><b>Ports - OUT:</b><br>";
    for (let i = 0; i < n.out_ports.length; i++) {
      let port = n.name + ":" + n.out_ports[i].name;
      let found = false;
      for (let j = 0; j < connections.length; j++) {
        if (connections[j].src.includes(port)) {
          found = true;
          break;
        }
      }
      html += "&nbsp;&nbsp;";
      if (!found) {
        html += "<strong>(UNCONNECTED)</strong> ";
      }
      html +=
        escapeHtml(n.out_ports[i].name) +
        " - " +
        escapeHtml(n.out_ports[i].type) +
        "<br>";
    }
  }
  return html;
}

function checkUpdate() {
  let headers = {
    "If-None-Match": window.graphEtag,
  };
  fetch("../../graph/" + encodeURIComponent(window.graphID), {
    headers: headers,
  })
    .then((response) => {
      if (response.status === 200) {
        window.location.reload(false);
      } else {
        setTimeout(checkUpdate, 10000);
      }
    })
    .catch((err) => setTimeout(checkUpdate, 1000));
}

function init(graphInfo) {
  let svg = d3.select("#main svg"),
    inner = svg.select("g"),
    zoom = d3.zoom().on("zoom", function (event) {
      inner.attr("transform", event.transform);
    });
  svg.call(zoom);

  let render = new dagreD3.render();
  render.shapes().point = function (parent, bbox, node) {
    return render.shapes().rect(parent, { width: 0, height: 0 }, node);
  };

  let width = parseInt(svg.style("width").replace(/px/, ""));
  let height = parseInt(svg.style("height").replace(/px/, ""));

  let g = new dagreD3.graphlib.Graph({ compound: true });
  g.setGraph({
    nodesep: 20,
    ranksep: 30,
    rankdir: width < height ? "TB" : "LR",
    marginx: 20,
    marginy: 20,
  });

  let hasNamespace = {};
  let ensureNamespace = function (fullNs, ns, parent) {
    if (hasNamespace[fullNs] > 0) {
      return;
    }
    hasNamespace[fullNs] = (hasNamespace[parent] || 0) + 1;
    g.setNode(fullNs, {
      label: ns,
      clusterLabelPos: "top",
    });
    if (parent !== "") {
      g.setParent(fullNs, parent);
    }
  };

  (graphInfo.topology.nodes || []).forEach(function (n) {
    let name = n.name;
    let prev = 0;
    for (let i = 0; i < name.length - 1; i++) {
      if (name[i] === "." || name[i] === "/") {
        ensureNamespace(
          name.substr(0, i),
          name.substr(prev, i - prev),
          prev === 0 ? "" : name.substr(0, prev - 1)
        );
        prev = i + 1;
      }
    }
    g.setNode(n.name, {
      label: name.substr(prev),
      class: "flow-node",

      info: n,
      desc: nodeDesc(n, graphInfo.topology.connections || []),
    });
    if (prev !== 0) {
      g.setParent(name, name.substr(0, prev - 1));
    }
  });

  (graphInfo.topology.connections || []).forEach(function (c, idx) {
    let tmpName = c.id;
    g.setNode(tmpName, {
      edgeName: tmpName,
      conn: c,

      label: "",
      shape: "point",
      class: "edge-node",
    });

    // representive node
    let reprNode = c.dst[0].split(":");
    let reprPort = reprNode[1];
    reprNode = reprNode[0];

    let ns = reprNode;
    let prev = -1;
    if (prev === -1) {
      prev = ns.lastIndexOf(".");
    }
    if (prev === -1) {
      prev = ns.lastIndexOf("/");
    }
    if (prev !== -1) {
      g.setParent(tmpName, reprNode.substr(0, prev));
    }

    let tmp = g.node(reprNode).info.in_ports;
    let dataType = "";
    for (let i = 0; i < tmp.length; i++) {
      if (tmp[i].name === reprPort) {
        dataType = tmp[i].type;
        break;
      }
    }
    let desc = edgeDesc(dataType, c);

    let edges = [];
    let n = g.node(tmpName);
    let portRe = /^(in|out)(_[0-9]+)?$/;
    c.src.forEach(function (s) {
      s = s.split(":");
      edges.push(s[0], tmpName);
      g.setEdge(s[0], tmpName, {
        class: "flow-edge edge-" + tmpName,
        arrowhead: "undirected",
        desc: desc,
        relatedNode: n,
        relatedEdges: edges,
        label: s[1].replace(portRe, ""),
        labeloffset: 0,
      });
    });
    c.dst.forEach(function (d) {
      d = d.split(":");
      edges.push(tmpName, d[0]);
      g.setEdge(tmpName, d[0], {
        class: "flow-edge edge-" + tmpName,
        desc: desc,
        relatedNode: n,
        relatedEdges: edges,
        label: d[1].replace(portRe, ""),
        labeloffset: 0,
      });
    });
  });

  inner.call(render, g);

  // Zoom and scale to fit
  let graphWidth = g.graph().width + 80;
  let graphHeight = g.graph().height + 40;
  let zoomScale = Math.min(width / graphWidth, height / graphHeight);
  let translateX = width / 2 - (graphWidth * zoomScale) / 2;
  let translateY = height / 2 - (graphHeight * zoomScale) / 2;
  const isUpdate = false;
  let svgZoom = isUpdate ? svg.transition().duration(500) : svg;
  svgZoom.call(
    zoom.transform,
    d3.zoomIdentity.translate(translateX, translateY).scale(zoomScale)
  );

  window.info = graphInfo;

  let prevSelected = [];
  inner.selectAll("g.flow-node").on("mouseover", function (_, d) {
    for (let i = 0; i < prevSelected.length; ++i) {
      prevSelected[i].classList.remove("selected");
    }

    document.getElementById("sidebar").innerHTML = g.node(d).desc;
    prevSelected = [this];
    this.classList.add("selected");

    window.updateGraph = null;
    document.getElementById("grapher").classList.remove("selected");
  });
  inner.selectAll("g.flow-edge").on("mouseover", function (_, d) {
    if (this.classList.contains("selected")) {
      return;
    }
    for (let i = 0; i < prevSelected.length; ++i) {
      prevSelected[i].classList.remove("selected");
    }

    let e = g.edge(d);
    document.getElementById("sidebar").innerHTML = e.desc;
    let r = e.relatedEdges;
    for (let i = 0; i < r.length; i += 2) {
      let el = g.edge({ v: r[i], w: r[i + 1] }).elem;
      el.classList.add("selected");
      prevSelected.push(el);
    }

    document.getElementById("grapher").classList.add("selected");
    window.updateGraph = function () {
      let g = window.grapher;
      if (!window.metricTime) {
        return;
      }
      g.x.domain(d3.extent(metricTime));
      g.svg.select(".x.axis").call(grapher.xAxis);

      g.y.domain([0, d3.max(e.relatedNode.rate)]);
      g.svg.select(".y.axis").call(g.yAxis);

      if (e.relatedNode.conn.capacity >= 0) {
        g.y2.domain([0, e.relatedNode.conn.capacity]);
      } else {
        g.y2.domain([0, d3.max(e.relatedNode.total)]);
      }
      g.svg.select(".y2.axis").call(g.y2Axis);

      g.svg
        .select(".rate-line")
        .attr("d", grapher.rateLine(e.relatedNode.rate));
      g.svg
        .select(".total-line")
        .attr("d", grapher.totalLine(e.relatedNode.total));
    };

    updateGraph();
  });

  let metrics = [];
  for (let i = 0; i < info.metrics.length; i++) {
    let metric = graphInfo.metrics[i].split("/");
    metric[0] = g.node(metric[0]);
    metrics.push(metric);
  }
  window.metrics = metrics;
  fetchMetrics();
  checkUpdate();
}

function processMetrics(buf) {
  let l = info.metrics_history_length;

  let val = new BigInt64Array(buf, 0, l);
  if (window.metricTime && window.metricTime[l - 1].getTime() == val[l - 1]) {
    setTimeout(fetchMetrics, 2000);
    return;
  }

  let metricTime = [];
  let skip = 0;
  for (let i = 0; i < l; i++) {
    if (val[i] != 0) {
      metricTime.push(new Date(Number(val[i])));
    } else {
      skip++;
    }
  }
  if (skip === l) {
    setTimeout(fetchMetrics, 2000);
    return;
  }
  for (let i = 0; i < skip; i++) {
    metricTime.unshift(metricTime[0]);
  }
  window.metricTime = metricTime;

  let allRates = [];
  for (let i = 0; i < metrics.length; i++) {
    if (metrics[i][1] === "rate") {
      let val = new Float32Array(buf, l * 8 + i * 4 * l, l);
      metrics[i][0].rate = val;
      allRates.push(val[l - 1]);
    }
  }
  allRates.sort();
  for (let i = 0; i < metrics.length; i++) {
    if (metrics[i][1] === "size") {
      let val = new Float32Array(buf, l * 8 + i * 4 * l, l);
      let sz = metrics[i][0].conn.capacity;
      metrics[i][0].total = val;

      let rate = metrics[i][0].rate[l - 1];
      let util = sz <= 0 ? 0 : metrics[i][0].total[l - 1] / sz;

      let h = 120 - util * 120;
      let s = rate <= 0 ? util : allRates.lastIndexOf(rate) / allRates.length;
      if (util >= 1) {
        s = 1;
      }
      let v = rate <= 0 && util <= 0 ? 0 : 50;
      d3.selectAll(".edge-" + metrics[i][0].edgeName + " > path").style(
        "stroke",
        "hsl(" + h.toFixed(1) + ", " + (s * 100).toFixed(1) + "%, " + v + "%)"
      );
    }
  }
  if (window.updateGraph) {
    window.updateGraph();
  }
  setTimeout(fetchMetrics, 500);
}

window.metricsEtag = null;
function fetchMetrics() {
  let headers = {
    Accept: "application/x-flow-metrics",
  };
  if (window.metricsEtag) {
    headers["If-None-Match"] = window.metricsEtag;
  }
  fetch("../../metrics/" + encodeURIComponent(graphID) + "?timeout=5", {
    headers: headers,
  })
    .then((response) => {
      if (response.status === 200) {
        window.metricsEtag = response.headers.get("Etag");
        response.arrayBuffer().then((r) => processMetrics(r));
      } else if (response.status === 404) {
        return;
      } else {
        setTimeout(fetchMetrics, 1000);
      }
    })
    .catch((err) => setTimeout(fetchMetrics, 5000));
}

function addGrapher() {
  const marginTop = 20;
  const marginLeft = 30;
  let width = 400 - marginLeft * 2;
  let height = 260 - marginTop * 2;

  var svg = d3
    .select("#grapher")
    .append("svg")
    .attr("width", width + marginLeft * 2)
    .attr("height", height + marginTop * 2)
    .append("g")
    .attr("transform", "translate(" + marginLeft + "," + marginTop + ")");

  var x = d3.scaleTime().range([0, width]);
  var xAxis = d3.axisBottom(x);
  svg
    .append("g")
    .attr("class", "x axis")
    .attr("transform", "translate(0," + height + ")")
    .call(xAxis);

  var y = d3.scaleLinear().range([height, 0]);
  var yAxis = d3.axisLeft(y);
  svg.append("g").attr("class", "y axis").call(yAxis);

  var y2 = d3.scaleLinear().range([height, 0]);
  var y2Axis = d3.axisRight(y2);
  svg
    .append("g")
    .attr("class", "y2 axis")
    .attr("transform", "translate(" + width + ",0)")
    .attr("color", "lightcoral")
    .call(y2Axis);

  var rateLine = d3
    .line()
    .x(function (_, i) {
      return x(metricTime[i]);
    })
    .y(function (d) {
      return y(d);
    });

  var totalLine = d3
    .line()
    .x(function (_, i) {
      return x(metricTime[i]);
    })
    .y(function (d) {
      return y2(d);
    });

  svg
    .append("path")
    .attr("fill", "none")
    .attr("stroke", "aquamarine")
    .attr("stroke-width", 1.5)
    .attr("data-legend", "A")
    .attr("class", "rate-line");

  svg
    .append("path")
    .attr("fill", "none")
    .attr("stroke", "lightcoral")
    .attr("data-legend", "B")
    .attr("stroke-width", 1.5)
    .attr("class", "total-line");

  let labels = [
    ["aquamarine", "Rate"],
    ["lightcoral", "Size"],
  ];
  svg
    .selectAll("mydots")
    .data(labels)
    .enter()
    .append("circle")
    .attr("cx", 270)
    .attr("cy", function (d, i) {
      return 169 + i * 25;
    })
    .attr("r", 7)
    .style("fill", function (d) {
      return d[0];
    });

  svg
    .selectAll("mylabels")
    .data(labels)
    .enter()
    .append("text")
    .attr("x", 284)
    .attr("y", function (d, i) {
      return 170 + i * 25;
    }) // 100 is where the first dot appears. 25 is the distance between
    // dots
    .style("fill", function (d) {
      return d[0];
    })
    .text(function (d) {
      return d[1];
    })
    .attr("text-anchor", "left")
    .style("alignment-baseline", "middle");

  window.grapher = {
    svg: svg,
    x: x,
    xAxis: xAxis,
    y: y,
    yAxis: yAxis,
    y2: y2,
    y2Axis: y2Axis,
    rateLine: rateLine,
    totalLine: totalLine,
  };
}

document.addEventListener("DOMContentLoaded", function () {
  const urlParams = new URLSearchParams(window.location.search);
  const id = urlParams.get("id");
  window.graphID = id;
  fetch("../../graph/" + encodeURIComponent(id))
    .then((response) => {
      window.graphEtag = response.headers.get("Etag");
      return response.json();
    })
    .then((r) => init(r));
  addGrapher();
});
