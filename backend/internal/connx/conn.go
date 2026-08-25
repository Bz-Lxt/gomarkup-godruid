package connx

import "context"

// Connection is a single long-lived resource owned by the pool.
type Connection interface {
	Ping(ctx context.Context) error
	Close() error
	Metadata() Metadata
}

// Connector opens new underlying connections. Implementations must not
// maintain their own generic pool; each Connect returns one resource.
type Connector interface {
	Connect(ctx context.Context) (Connection, error)
	Kind() string
}

// Optional metadata for adapters that wrap a protocol client.
type Pingable = Connection
