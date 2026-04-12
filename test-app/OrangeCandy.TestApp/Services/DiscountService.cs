using Microsoft.Extensions.Logging;

namespace OrangeCandy.TestApp.Services;

public sealed class DiscountService(ILogger<DiscountService> logger) : IDiscountService
{
    private static readonly Dictionary<string, decimal> Discounts = new()
    {
        ["SAVE10"] = 0.10m,
        ["SAVE20"] = 0.20m,
        ["VIP"] = 0.30m,
        ["WELCOME"] = 0.15m
    };

    public Task<Result<decimal>> Lookup(string code)
    {
        logger.LogInformation("Looking up discount code: {Code}", code);

        if (string.IsNullOrWhiteSpace(code))
            return Task.FromResult(Result<decimal>.Success(0m));

        var normalized = code.Trim().ToUpperInvariant();

        if (!Discounts.TryGetValue(normalized, out var rate))
        {
            logger.LogWarning("Invalid discount code: {Code}", code);
            return Task.FromResult(Result<decimal>.Failure($"Unknown discount code: {code}"));
        }

        logger.LogInformation("Discount code {Code} resolved to {Rate:P0}", code, rate);
        return Task.FromResult(Result<decimal>.Success(rate));
    }
}
