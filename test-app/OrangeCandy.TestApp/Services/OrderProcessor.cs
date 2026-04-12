using Microsoft.Extensions.Logging;
using OrangeCandy.TestApp.Models;

namespace OrangeCandy.TestApp.Services;

public sealed class OrderProcessor(
    IDiscountService discountService,
    ITaxCalculator taxCalculator,
    ILogger<OrderProcessor> logger) : IOrderProcessor
{
    public async Task<Result<OrderTotal>> CalculateTotal(Order order)
    {
        logger.LogInformation("Processing order with {Count} line(s)", order.Lines.Count);

        if (order.Lines.Count == 0)
            return Result<OrderTotal>.Failure("Order must have at least one line");

        var subtotal = order.Lines.Sum(line => line.Price * line.Quantity);
        logger.LogInformation("Subtotal: {Subtotal:C}", subtotal);

        var discountRate = 0m;
        if (order.DiscountCode is not null)
        {
            var discountResult = await discountService.Lookup(order.DiscountCode);
            if (!discountResult.IsSuccess)
                return Result<OrderTotal>.Failure(discountResult.Error!);

            discountRate = discountResult.Value;
        }

        var discountAmount = Math.Round(subtotal * discountRate, 2);
        var taxableAmount = subtotal - discountAmount;

        var taxResult = await taxCalculator.Calculate(taxableAmount);
        if (!taxResult.IsSuccess)
            return Result<OrderTotal>.Failure(taxResult.Error!);

        var tax = taxResult.Value;
        var total = taxableAmount + tax;

        logger.LogInformation(
            "Order complete — Subtotal: {Subtotal:C}, Discount: {Discount:C}, Tax: {Tax:C}, Total: {Total:C}",
            subtotal, discountAmount, tax, total);

        return Result<OrderTotal>.Success(new OrderTotal(subtotal, discountAmount, tax, total));
    }
}
