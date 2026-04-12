namespace OrangeCandy.TestApp.Models;

public sealed record OrderTotal(decimal Subtotal, decimal Discount, decimal Tax, decimal Total);
