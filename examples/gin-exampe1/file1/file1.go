package file1

import (
	"context"
	logger "github.com/rafapcarvalho/logtracer/pkg/logtracer"
)

func CallFile1(ctx context.Context) {
	ctx, span := logger.StartSpan(ctx, "callFile1")
	defer span.End()

	logger.TstLog.Error(ctx, "menssagem do novo", "arquivo", "file1")
}
