using OrangeCandy.TestApp.Models;

namespace OrangeCandy.TestApp.Services;

public interface IDiscountService
{
    Task<Result<decimal>> Lookup(string code);
}
