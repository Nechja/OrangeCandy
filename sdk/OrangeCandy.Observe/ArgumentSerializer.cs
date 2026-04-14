using System.Text.Json;

namespace OrangeCandy.Observe;

public static class ArgumentSerializer
{
    private static readonly JsonSerializerOptions Options = new()
    {
        MaxDepth = 3,
        WriteIndented = false,
        DefaultIgnoreCondition = System.Text.Json.Serialization.JsonIgnoreCondition.WhenWritingNull
    };

    public static string Serialize(object? value)
    {
        if (value is null) return "null";

        try
        {
            var json = JsonSerializer.Serialize(value, value.GetType(), Options);
            return json.Length > 500 ? json[..497] + "..." : json;
        }
        catch
        {
            var str = value.ToString() ?? value.GetType().Name;
            return str.Length > 500 ? str[..497] + "..." : str;
        }
    }
}
