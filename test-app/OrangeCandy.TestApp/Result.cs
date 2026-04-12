namespace OrangeCandy.TestApp;

public sealed record Result<T>
{
    public T? Value { get; }
    public string? Error { get; }
    public bool IsSuccess => Error is null;

    private Result(T value) => Value = value;
    private Result(string error) => Error = error;

    public static Result<T> Success(T value) => new(value);
    public static Result<T> Failure(string error) => new(error);

    public Result<TNext> Map<TNext>(Func<T, TNext> transform) =>
        IsSuccess ? Result<TNext>.Success(transform(Value!)) : Result<TNext>.Failure(Error!);

    public async Task<Result<TNext>> Bind<TNext>(Func<T, Task<Result<TNext>>> transform) =>
        IsSuccess ? await transform(Value!) : Result<TNext>.Failure(Error!);

    public T Unwrap() => IsSuccess ? Value! : throw new InvalidOperationException(Error);
}
