using OrangeCandy.TestApp.Models;

namespace OrangeCandy.TestApp.Services;

public interface IInventoryService
{
    Task<Result<bool>> CheckStock(OrderLine line);
}
