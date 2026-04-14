namespace OrangeCandy.Observe;

public sealed class ObserverOptions
{
    public string ServerUrl { get; set; } = "http://localhost:9119";
    public int MaxSerializationDepth { get; set; } = 2;
    internal HashSet<Type> ExcludedInterfaces { get; } = [];

    public ObserverOptions Exclude<T>() where T : class
    {
        ExcludedInterfaces.Add(typeof(T));
        return this;
    }
}
