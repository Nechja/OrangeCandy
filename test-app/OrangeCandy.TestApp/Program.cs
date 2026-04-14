using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;
using OrangeCandy.Observe;
using OrangeCandy.TestApp.Models;
using OrangeCandy.TestApp.Services;

namespace OrangeCandy.TestApp;

class Program
{
    static async Task Main(string[] args)
    {
        var host = Host.CreateDefaultBuilder(args)
            .ConfigureServices(services =>
            {
                services.AddSingleton<IDiscountService, DiscountService>();
                services.AddSingleton<ITaxCalculator, TaxCalculator>();
                services.AddSingleton<IOrderProcessor, OrderProcessor>();
                services.AddSingleton<IInventoryService, InventoryService>();
                services.AddOrangeCandyObserver();
            })
            .ConfigureLogging(logging =>
            {
                logging.SetMinimumLevel(LogLevel.Information);
            })
            .Build();

        var processor = host.Services.GetRequiredService<IOrderProcessor>();
        var inventory = host.Services.GetRequiredService<IInventoryService>();
        var logger = host.Services.GetRequiredService<ILogger<Program>>();


        var allScenarios = BuildScenarios();

        var filter = args.FirstOrDefault();
        var scenarios = string.IsNullOrEmpty(filter)
            ? allScenarios
            : allScenarios.Where(s => s.Name.Contains(filter, StringComparison.OrdinalIgnoreCase)).ToList();

        if (scenarios.Count == 0)
        {
            logger.LogError("No scenarios matched filter: {Filter}", filter);
            logger.LogInformation("Available: {Names}", string.Join(", ", allScenarios.Select(s => s.Name)));
            return;
        }

        foreach (var (name, order) in scenarios)
        {
            logger.LogInformation("=== Scenario: {Name} ===", name);

            foreach (var line in order.Lines)
                await inventory.CheckStock(line);

            var result = await processor.CalculateTotal(order);

            if (result.IsSuccess)
            {
                var total = result.Value!;
                logger.LogInformation(
                    "Result — Subtotal: {Subtotal:C}, Discount: {Discount:C}, Tax: {Tax:C}, Total: {Total:C}",
                    total.Subtotal, total.Discount, total.Tax, total.Total);
            }
            else
            {
                logger.LogError("Failed: {Error}", result.Error);
            }

            logger.LogInformation("");
        }

        host.Dispose();
    }

    private static List<(string Name, Order Order)> BuildScenarios()
    {
        return
        [
            ("Simple order, no discount", new Order(
            [
                new OrderLine("Widget", 25.00m, 2),
                new OrderLine("Gadget", 49.99m, 1)
            ])),

            ("Order with valid discount", new Order(
            [
                new OrderLine("Premium Widget", 75.00m, 1),
                new OrderLine("Deluxe Gadget", 120.00m, 2)
            ], "SAVE20")),

            ("VIP discount", new Order(
            [
                new OrderLine("Enterprise License", 999.99m, 1)
            ], "VIP")),

            ("Invalid discount code", new Order(
            [
                new OrderLine("Widget", 25.00m, 1)
            ], "BOGUS")),

            ("Empty order", new Order([])),

            ("Crash unknown product", new Order(
            [
                new OrderLine("Phantom Widget", 50.00m, 1)
            ])),
        ];
    }
}
