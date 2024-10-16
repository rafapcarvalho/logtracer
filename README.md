# logtracer
Esta biblioteca em Golang será um wrapper sobre o logger oficial do Go (slog) e OpenTelemetry, projetada para fornecer uma solução unificada para logs e instrumentação de tracing em aplicações HTTP e gRPC.

Mods de uso do StartSpan:

1. Sem ID personalizado: 
```go
ctx := logtracer.StartSpan(ctx, "nameSpan")
defer logtracer.EndSpan(ctx)

logtracer.SrvcLog.Info(ctx, "Esta é uma mensagem de log sem ID personalizado")
logtracer.AddAttribute(ctx, "key3", "value3")
```
2. Com ID personalizado:
```go
ctx, span := logtracer.StartSpan(ctx, "nameSpan", logtracer.WithID("customID123"))
logtracer.SrvcLog.Info(ctx, "Esta é uma mensagem de log com ID personalizado")
logtracer.AddAttribute(ctx, "key3", "value3
```
3. Com ID personalizado e opções de tracer adicionais:
```go
ctx, span := logtracer.StartSpan(ctx, "nameSpan", logtracer.WithID("customID123"), logtracer.WithAttribute("key","value"))
logtracer.SrvcLog.Info(ctx, "Esta é uma mensagem de log com ID personalizado")
logtracer.AddAttribute(ctx, "key3", "value3
```
4. Apenas com opções de tracer adicionais:
```go
ctx, span := logtracer.StartSpan(ctx, "nameSpan", tracer.WithAttributes(attribute.String("key","value")))
logtracer.SrvcLog.Info(ctx, "Esta é uma mensagem de log sem ID personalizado")
logtracer.AddAttribute(ctx, "key3", "value3
```

Nova funcionalidade WithoutTrace(), da a opção de não gerar um trace mesmo com a configuração global de carregamento com trace.
Pode-se usar da seguinte maneira:

1. Log padrão com trace
```go
logtracer.SrvcLog.Info(ctx, "Esta é uma mensagem de log com trace")
```
2. Log sem trace
```go
logtracer.SrvcLog.Info(ctx, "Esta é uma mensagem de log sem trace").WithoutTrace()
```
3. Log padrão com trace e argumentos
```go
logtracer.SrvcLog.Info(ctx, "Esta é uma mensagem de log com trace", "arg1", 1, "arg2", "string")
```
4. Log sem trace e argumentos
```go
logtracer.SrvcLog.Info(ctx, "Esta é uma mensagem de log sem trace", "arg1", 1, "arg2", "string").WithoutTrace()
```
5. Log formatado com trace
```go
logtracer.SrvcLog.Infof(ctx, "mensagem formatada: %s", "com trace")
```
6. Log formatado sem trace
```go
logtracer.SrvcLog.Infof(ctx, "mensagem formatada: %s", "sem trace").WithoutTrace()
```