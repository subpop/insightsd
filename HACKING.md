# Prerequisites

In order to run `./cmd/insightsd` on the system bus, you'll need to create
the following files:

```
cp ./data/dbus/insightsd.conf /etc/dbus-1/system.d/insightsd.conf
```

Add a user policy permitting your user (`$USER`) to own `com.redhat.insightsd`
on the system bus:

```xml
  <policy user="insightsd">
    <allow own="com.redhat.insightsd"/>
  </policy>
```

# Run `./cmd/insightsd`

`go run ./cmd/insightsd --base-url https://cloud.redhat.com/api --username foo --password secret`

# GDBus

You can install D-Feet to browse the bus objects in a graphical way, or use
`gdbus` to send methods directly.

```bash
gdbus introspect --system \
    --dest com.redhat.insightsd \
    --object-path /com/redhat/insightsd
gdbus call --system \
    --dest com.redhat.insightsd \
    --object-path /com/redhat/insightsd \
    --method com.redhat.insightsd.Upload \
    "$HOME/insights-ic-rhel8-dev-thelio-20200521100458.tar.gz" "advisor"
```

# Call Graphs

Call graphs can be generated to provide a high-level overview of the interactions
between packages.

For basic call graphs, install `go-callvis` (`go get -u github.com/ofabry/go-callvis`) and run:

```bash
# Call graph of the main function of insightsd, up to calls into the insights package
go-callvis -nostd -format png -file insightsd.main ./cmd/insightsd
# Call graph of the insights package, as invoked by insightsd
go-callvis -nostd -format png -file insightsd.insights -focus github.com/subpop/insightsd/pkg ./cmd/insightsd
# Call graph of the main function of insights-exec, up to calls into the insights package
go-callvis -nostd -format png -file insights-exec.main ./cmd/insights-exec
# Call graph of the insights package, as invoked by insights-exec
go-callvis -nostd -format png -file insights-exec.insights -focus github.com/subpop/insightsd/pkg ./cmd/insights-exec
```

For more detailed, interactive call graphs, install `callgraph` and `digraph`.

```bash
go get -u golang.org/x/tools/cmd/callgraph
go get -u gilang.org/x/tools/cmd/digraph
```

Generate a call graph using `callgraph`, filter the resulting graph to
exclude standard library calls and pipe the result into `digraph`. See the `-help`
output of `digraph` for how to interact with the graph.

```bash
`callgraph -algo pta -format digraph ./cmd/insights-exec | grep github.com/subpop/insightsd | sort | uniq | digraph
```
