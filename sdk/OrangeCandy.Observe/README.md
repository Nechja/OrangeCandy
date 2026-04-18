# OrangeCandy.Observe

AOP observer for .NET. Intercepts DI-registered method calls and streams observation events (arguments, return values, timing, exceptions) to the OrangeCandy debug server for real-time analysis.

Part of the [OrangeCandy](https://github.com/Nechja/OrangeCandy) debugging suite.

## Install

```
dotnet add package OrangeCandy.Observe
```

## Usage

```csharp
using OrangeCandy.Observe;

services.AddOrangeCandyObserver();
```

All interfaces registered with DI are automatically wrapped in a proxy that
reports method calls to `http://localhost:9119/api/observe`.

`Microsoft.*` and `System.*` interfaces are excluded by default.

## Configuration

```csharp
services.AddOrangeCandyObserver(options =>
{
    options.ServerUrl = "http://localhost:9119";
    options.Exclude<IMyInterface>();
});
```

## How it works

Uses `System.Reflection.DispatchProxy` to wrap each DI interface with a transparent proxy. The proxy captures method entry, exit, and exceptions, serializes arguments and return values, and sends batched events over HTTP to the debug server.

The observer only works with interface-based DI registrations. Concrete
classes registered directly are not wrapped.

## Requirements

- .NET 10
- OrangeCandy debug server running on the configured port

## License

Apache 2.0 with Commons Clause. Free to use, modify, and redistribute. You may not sell the software or a product/service whose value derives substantially from it. See [LICENSE](https://github.com/Nechja/OrangeCandy/blob/main/LICENSE).
