package adapter

import (
	"encoding/json"
	"fmt"

	"github.com/meshery/schemas/models/v1beta3/connection"
)

func newAdapterConnection(host string, port int, ctxID string) ([]byte, error) {
	if host == "" {
		return nil, fmt.Errorf("adapter hostname is required")
	}

	if port < 1 || port > 65535 {
		return nil, fmt.Errorf("invalid adapter port: %d", port)
	}

	conn := connection.Connection{
		Name: host,
		Kind: "meshery-adapter",
		Metadata: map[string]interface{}{
			"host_name":  host,
			"port":       port,
			"context_id": ctxID,
		},
	}

	return json.Marshal(conn)
}
