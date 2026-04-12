namespace OrangeCandy.TestApp.Models;

public sealed record OrderLine(string ProductName, decimal Price, int Quantity);
