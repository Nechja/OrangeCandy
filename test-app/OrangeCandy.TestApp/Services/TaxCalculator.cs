using Microsoft.Extensions.Logging;

namespace OrangeCandy.TestApp.Services;

public sealed class TaxCalculator(ILogger<TaxCalculator> logger) : ITaxCalculator
{
    private const decimal TaxRate = 0.08m;

    public Task<Result<decimal>> Calculate(decimal amount)
    {
        logger.LogInformation("Calculating tax on {Amount:C}", amount);

        if (amount < 0)
            return Task.FromResult(Result<decimal>.Failure("Cannot calculate tax on a negative amount"));

        var tax = Math.Round(amount * TaxRate, 2);

        logger.LogInformation("Tax calculated: {Tax:C}", tax);
        return Task.FromResult(Result<decimal>.Success(tax));
    }
}
