package connectors

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/centralmind/gateway/model"
)

type loggingConnector struct {
	Connector
}

// WrapWithLogging wraps the given connector with logging functionality.
func WrapWithLogging(inner Connector) Connector {
	return &loggingConnector{Connector: inner}
}

func (l loggingConnector) Query(ctx context.Context, endpoint model.Endpoint, params map[string]any) ([]map[string]any, error) {
	logrus.Infof("SQL query: %s", endpoint.Query)
	res, err := l.Connector.Query(ctx, endpoint, params)
	if err != nil {
		logrus.Errorf("query failed: %v", err)
	}
	return res, err
}

func (l loggingConnector) InferQuery(ctx context.Context, query string) ([]model.ColumnSchema, error) {
	logrus.Infof("Infer query: %s", query)
	res, err := l.Connector.InferQuery(ctx, query)
	if err != nil {
		logrus.Errorf("infer query failed: %v", err)
	}
	return res, err
}
