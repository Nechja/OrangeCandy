using System.Text.Json.Serialization;

namespace OrangeCandy.Observe;

public sealed record ObservationEvent(
    [property: JsonPropertyName("trace_id")] string TraceId,
    [property: JsonPropertyName("event_type")] string EventType,
    [property: JsonPropertyName("interface")] string Interface,
    [property: JsonPropertyName("method")] string Method,
    [property: JsonPropertyName("arguments")] string[]? Arguments,
    [property: JsonPropertyName("return_value")] string? ReturnValue,
    [property: JsonPropertyName("exception")] string? Exception,
    [property: JsonPropertyName("duration_ms")] long DurationMs,
    [property: JsonPropertyName("depth")] int Depth,
    [property: JsonPropertyName("timestamp")] DateTimeOffset Timestamp);
