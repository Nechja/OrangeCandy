namespace OrangeCandy.TestApp.Models;

public sealed record Order(List<OrderLine> Lines, string? DiscountCode = null);
