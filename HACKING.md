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
gdbus call --system \
    --dest com.redhat.insightsd \
    --object-path /com/redhat/insightsd \
    --method com.redhat.insightsd.Upload \
    "$HOME/insights-ic-rhel8-dev-thelio-20200521100458.tar.gz" "advisor"
```
