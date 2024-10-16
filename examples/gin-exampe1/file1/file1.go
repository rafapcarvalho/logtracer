package file1

import (
	"context"
	logger "github.com/rafapcarvalho/logtracer/pkg/logtracer"
)

func CallFile1(ctx context.Context) {
	ctx = logger.StartSpan(ctx, "callFile1")
	defer logger.EndSpan(ctx)

	logger.TstLog.Error(ctx, "menssagem do novo", "arquivo", "file1")
}
