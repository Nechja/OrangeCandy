using System.Collections.Concurrent;
using System.Reflection;

namespace OrangeCandy.Observe;

public static class ObservingProxyFactory
{
    private static readonly ConcurrentDictionary<Type, MethodInfo> CreateMethods = new();

    public static object Create(Type interfaceType, object target, IObservationReporter reporter, AsyncLocal<int> depth)
    {
        var createMethod = CreateMethods.GetOrAdd(interfaceType, type =>
            typeof(ObservingProxy<>)
                .MakeGenericType(type)
                .GetMethod(nameof(ObservingProxy<object>.Create), BindingFlags.Public | BindingFlags.Static)!);

        return createMethod.Invoke(null, [target, reporter, depth])!;
    }
}
