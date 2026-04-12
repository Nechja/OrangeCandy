using OrangeCandy.TestApp.Models;

namespace OrangeCandy.TestApp.Services;

public interface IOrderProcessor
{
    Task<Result<OrderTotal>> CalculateTotal(Order order);
}
