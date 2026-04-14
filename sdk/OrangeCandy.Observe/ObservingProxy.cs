using System.Diagnostics;
using System.Reflection;

namespace OrangeCandy.Observe;

public class ObservingProxy<TInterface> : DispatchProxy where TInterface : class
{
    private TInterface _target = null!;
    private IObservationReporter _reporter = null!;
    private string _interfaceName = null!;
    private AsyncLocal<int> _depth = null!;

    private static readonly MethodInfo InterceptAsyncMethod =
        typeof(ObservingProxy<TInterface>)
            .GetMethod(nameof(InterceptAsync), BindingFlags.NonPublic | BindingFlags.Static)!;

    public static TInterface Create(TInterface target, IObservationReporter reporter, AsyncLocal<int> depth)
    {
        var proxy = DispatchProxy.Create<TInterface, ObservingProxy<TInterface>>() as ObservingProxy<TInterface>;
        proxy!._target = target;
        proxy._reporter = reporter;
        proxy._interfaceName = typeof(TInterface).Name;
        proxy._depth = depth;
        return (TInterface)(object)proxy;
    }

    protected override object? Invoke(MethodInfo? targetMethod, object?[]? args)
    {
        if (targetMethod is null) return null;

        var traceId = Guid.NewGuid().ToString("N")[..8];
        var depth = _depth.Value++;
        var timestamp = Stopwatch.GetTimestamp();

        var serializedArgs = args?.Select(ArgumentSerializer.Serialize).ToArray();

        _reporter.Enqueue(new ObservationEvent(
            traceId, "method_enter", _interfaceName, targetMethod.Name,
            serializedArgs, null, null, 0, depth, DateTimeOffset.UtcNow));

        object? result;
        try
        {
            result = targetMethod.Invoke(_target, args);
        }
        catch (TargetInvocationException ex)
        {
            var elapsed = GetElapsedMs(timestamp);
            _depth.Value--;
            _reporter.Enqueue(new ObservationEvent(
                traceId, "method_exception", _interfaceName, targetMethod.Name,
                null, null, FormatException(ex.InnerException ?? ex), elapsed, depth, DateTimeOffset.UtcNow));
            throw ex.InnerException ?? ex;
        }

        var returnType = targetMethod.ReturnType;

        if (returnType == typeof(Task))
        {
            return InterceptTaskAsync((Task)result!, traceId, timestamp, depth);
        }

        if (returnType.IsGenericType && returnType.GetGenericTypeDefinition() == typeof(Task<>))
        {
            var resultType = returnType.GetGenericArguments()[0];
            return InterceptAsyncMethod
                .MakeGenericMethod(resultType)
                .Invoke(null, [result, traceId, timestamp, depth, _reporter, _interfaceName, targetMethod.Name, _depth]);
        }

        var syncElapsed = GetElapsedMs(timestamp);
        _depth.Value--;
        _reporter.Enqueue(new ObservationEvent(
            traceId, "method_exit", _interfaceName, targetMethod.Name,
            null, ArgumentSerializer.Serialize(result), null, syncElapsed, depth, DateTimeOffset.UtcNow));

        return result;
    }

    private async Task InterceptTaskAsync(Task task, string traceId, long startTimestamp, int depth)
    {
        try
        {
            await task;
            var elapsed = GetElapsedMs(startTimestamp);
            _depth.Value--;
            _reporter.Enqueue(new ObservationEvent(
                traceId, "method_exit", _interfaceName, "Task",
                null, null, null, elapsed, depth, DateTimeOffset.UtcNow));
        }
        catch (Exception ex)
        {
            var elapsed = GetElapsedMs(startTimestamp);
            _depth.Value--;
            _reporter.Enqueue(new ObservationEvent(
                traceId, "method_exception", _interfaceName, "Task",
                null, null, FormatException(ex), elapsed, depth, DateTimeOffset.UtcNow));
            throw;
        }
    }

    private static async Task<T> InterceptAsync<T>(
        Task<T> task, string traceId, long startTimestamp, int depth,
        IObservationReporter reporter, string interfaceName, string methodName, AsyncLocal<int> depthTracker)
    {
        try
        {
            var result = await task;
            var elapsed = GetElapsedMs(startTimestamp);
            depthTracker.Value--;
            reporter.Enqueue(new ObservationEvent(
                traceId, "method_exit", interfaceName, methodName,
                null, ArgumentSerializer.Serialize(result), null, elapsed, depth, DateTimeOffset.UtcNow));
            return result;
        }
        catch (Exception ex)
        {
            var elapsed = GetElapsedMs(startTimestamp);
            depthTracker.Value--;
            reporter.Enqueue(new ObservationEvent(
                traceId, "method_exception", interfaceName, methodName,
                null, null, FormatException(ex), elapsed, depth, DateTimeOffset.UtcNow));
            throw;
        }
    }

    private static long GetElapsedMs(long startTimestamp) =>
        (long)Stopwatch.GetElapsedTime(startTimestamp).TotalMilliseconds;

    private static string FormatException(Exception ex) =>
        $"{ex.GetType().Name}: {ex.Message}";
}
