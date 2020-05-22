package insights

import (
	"fmt"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"

	insights "github.com/subpop/insightsd/pkg"
)

const interfaceXML = `
<!DOCTYPE node PUBLIC "-//freedesktop//DTD D-BUS Object Introspection 1.0//EN"
"http://www.freedesktop.org/standards/dbus/1.0/introspect.dtd">

<node name="/com/redhat/insightsd">
  <interface name="com.redhat.insightsd">
	<method name="Upload">
	  <arg name="file" direction="in" type="s"/>
	  <arg name="collector" direction="in" type="s"/>
    </method>
  </interface>
  
  <interface name="org.freedesktop.DBus.Introspectable">
    <method name="Introspect">
      <arg name="out" direction="out" type="s"/>
    </method>
  </interface>
</node>
`

// DBusServer serves insightsd functionality over the system D-Bus.
type DBusServer struct {
	client *insights.Client
	conn   *dbus.Conn
}

// NewDBusServer creates a new server. The provided HTTP Client will be used for
// all HTTP requests.
func NewDBusServer(client *insights.Client) (*DBusServer, error) {
	return &DBusServer{
		client: client,
	}, nil
}

// Connect opens a connection to the system bus, exports the server as an object
// on the bus, and requests the well-known name "com.redhat.insightsd".
func (s *DBusServer) Connect() error {
	conn, err := dbus.SystemBus()
	if err != nil {
		return err
	}
	s.conn = conn

	s.conn.Export(s, "/com/redhat/insightsd", "com.redhat.insightsd")
	s.conn.Export(introspect.Introspectable(interfaceXML),
		"/com/redhat/insightsd",
		"org.freedesktop.DBus.Introspectable")

	reply, err := s.conn.RequestName("com.redhat.insightsd", dbus.NameFlagDoNotQueue)
	if err != nil {
		return err
	}
	if reply != dbus.RequestNameReplyPrimaryOwner {
		return fmt.Errorf("name already taken")
	}
	return nil
}

// Close closes the connection.
func (s *DBusServer) Close() error {
	return s.conn.Close()
}

func (s *DBusServer) Upload(file, collector string) *dbus.Error {
	if err := insights.Upload(s.client, file, collector); err != nil {
		return &dbus.Error{
			Name: err.Error(),
		}
	}
	return nil
}
