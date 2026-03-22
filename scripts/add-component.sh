#!/usr/bin/env bash
set -euo pipefail

COMPONENT="${1:-}"

if [ -z "$COMPONENT" ]; then
    echo "Usage: $0 <component-name>"
    echo ""
    echo "Scaffolds a new infrastructure component for the service."
    echo ""
    echo "Examples:"
    echo "  $0 redis"
    echo "  $0 kafka"
    echo "  $0 mongodb"
    echo ""
    echo "Or via make:"
    echo "  make add-component COMPONENT=redis"
    exit 1
fi

DIR="pkg/${COMPONENT}"

if [ -d "$DIR" ]; then
    echo "Error: ${DIR} already exists."
    exit 1
fi

mkdir -p "$DIR"

# Known components with their Go import paths
PKG=""
IMPORT_LINE=""
case "$COMPONENT" in
    redis)
        PKG="github.com/redis/go-redis/v9"
        IMPORT_LINE="\"github.com/redis/go-redis/v9\""
        ;;
    kafka)
        PKG="github.com/segmentio/kafka-go"
        IMPORT_LINE="\"github.com/segmentio/kafka-go\""
        ;;
    mongodb|mongo)
        PKG="go.mongodb.org/mongo-driver/mongo"
        IMPORT_LINE="\"go.mongodb.org/mongo-driver/mongo\""
        ;;
    nats)
        PKG="github.com/nats-io/nats.go"
        IMPORT_LINE="\"github.com/nats-io/nats.go\""
        ;;
esac

# Generate the component Go file
cat > "${DIR}/${COMPONENT}.go" << GOEOF
package ${COMPONENT}

import (
	"context"
	"fmt"
)

const ComponentName = "${COMPONENT}"

// Component implements service.Component for ${COMPONENT}.
type Component struct {
	// TODO: Add your ${COMPONENT} client/connection field here.
}

// NewComponent creates a new ${COMPONENT} component.
func NewComponent() *Component {
	return &Component{}
}

func (c *Component) Name() string { return ComponentName }

func (c *Component) Init(ctx context.Context) error {
	// TODO: Initialize ${COMPONENT} connection.
	return fmt.Errorf("${COMPONENT}: Init not implemented")
}

func (c *Component) Close(_ context.Context) error {
	// TODO: Close ${COMPONENT} connection.
	return nil
}

// HealthCheck verifies ${COMPONENT} connectivity.
func (c *Component) HealthCheck(ctx context.Context) error {
	// TODO: Implement health check.
	return fmt.Errorf("${COMPONENT}: HealthCheck not implemented")
}
GOEOF

echo "Created ${DIR}/${COMPONENT}.go"

# Add config section to config.yaml if not present
if ! grep -q "  ${COMPONENT}:" config.yaml 2>/dev/null; then
    cat >> config.yaml << YAMLEOF

  ${COMPONENT}:
    enabled: true
    # TODO: Add ${COMPONENT}-specific configuration here.
YAMLEOF
    echo "Added ${COMPONENT} section to config.yaml"
fi

# Install Go package if known
if [ -n "$PKG" ]; then
    echo "Running: go get ${PKG}"
    go get "$PKG"
fi

echo ""
echo "Next steps:"
echo "  1. Edit ${DIR}/${COMPONENT}.go — add connection logic"
echo "  2. Add config fields to internal/config/config.go (ComponentsConfig struct)"
echo "  3. Wire it up in cmd/service/main.go:"
echo ""
echo "     import \"github.com/bdtfs/go-service-template/pkg/${COMPONENT}\""
echo ""
echo "     if cfg.Components.${COMPONENT^}.Enabled {"
echo "         opts = append(opts, service.WithComponent(${COMPONENT}.NewComponent()))"
echo "     }"
echo ""
