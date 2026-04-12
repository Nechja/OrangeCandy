using Microsoft.Extensions.Logging;
using OrangeCandy.TestApp.Models;

namespace OrangeCandy.TestApp.Services;

public sealed class InventoryService(ILogger<InventoryService> logger) : IInventoryService
{
    private static readonly Dictionary<string, int> Stock = new()
    {
        ["Widget"] = 100,
        ["Gadget"] = 50,
        ["Premium Widget"] = 10,
        ["Deluxe Gadget"] = 5,
        ["Enterprise License"] = 999
    };

    public Task<Result<bool>> CheckStock(OrderLine line)
    {
        logger.LogInformation("Checking stock for {Product}, quantity {Qty}", line.ProductName, line.Quantity);

        var available = Stock[line.ProductName];

        if (line.Quantity > available)
        {
            logger.LogWarning("Insufficient stock for {Product}: requested {Qty}, available {Available}",
                line.ProductName, line.Quantity, available);
            return Task.FromResult(Result<bool>.Failure(
                $"Insufficient stock for {line.ProductName}: requested {line.Quantity}, available {available}"));
        }

        return Task.FromResult(Result<bool>.Success(true));
    }
}
