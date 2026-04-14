using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.DependencyInjection.Extensions;
using Microsoft.Extensions.Logging;

namespace OrangeCandy.Observe;

public static class ServiceCollectionExtensions
{
    private static readonly HashSet<string> ExcludedNamespacePrefixes = ["Microsoft.", "System."];

    public static IServiceCollection AddOrangeCandyObserver(
        this IServiceCollection services,
        Action<ObserverOptions>? configure = null)
    {
        var options = new ObserverOptions();
        configure?.Invoke(options);

        var depth = new AsyncLocal<int>();
        HttpObservationReporter? reporter = null;

        services.AddSingleton<IObservationReporter>(sp =>
        {
            var httpClient = new HttpClient { BaseAddress = new Uri(options.ServerUrl), Timeout = TimeSpan.FromSeconds(5) };
            var log = sp.GetRequiredService<ILogger<HttpObservationReporter>>();
            reporter = new HttpObservationReporter(httpClient, log);
            reporter.Start();
            return reporter;
        });

        var descriptors = services.ToList();

        foreach (var descriptor in descriptors)
        {
            if (!ShouldProxy(descriptor, options)) continue;

            var serviceType = descriptor.ServiceType;
            var implType = descriptor.ImplementationType!;
            var lifetime = descriptor.Lifetime;

            services.Remove(descriptor);

            services.TryAdd(new ServiceDescriptor(implType, implType, lifetime));

            services.Add(new ServiceDescriptor(
                serviceType,
                sp =>
                {
                    var target = sp.GetRequiredService(implType);
                    var rep = sp.GetRequiredService<IObservationReporter>();
                    return ObservingProxyFactory.Create(serviceType, target, rep, depth);
                },
                lifetime));
        }

        return services;
    }

    private static bool ShouldProxy(ServiceDescriptor descriptor, ObserverOptions options)
    {
        if (!descriptor.ServiceType.IsInterface) return false;
        if (descriptor.ImplementationType is null) return false;
        if (options.ExcludedInterfaces.Contains(descriptor.ServiceType)) return false;

        var ns = descriptor.ServiceType.Namespace ?? "";
        foreach (var prefix in ExcludedNamespacePrefixes)
        {
            if (ns.StartsWith(prefix, StringComparison.Ordinal)) return false;
        }

        return true;
    }
}
