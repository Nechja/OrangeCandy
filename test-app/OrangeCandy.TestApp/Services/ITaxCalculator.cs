namespace OrangeCandy.TestApp.Services;

public interface ITaxCalculator
{
    Task<Result<decimal>> Calculate(decimal amount);
}
