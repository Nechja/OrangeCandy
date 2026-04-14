using System.Net.Http.Json;
using System.Collections.Concurrent;
using Microsoft.Extensions.Logging;

namespace OrangeCandy.Observe;

public sealed class HttpObservationReporter(
    HttpClient httpClient,
    ILogger<HttpObservationReporter> logger) : IObservationReporter, IDisposable
{
    private readonly BlockingCollection<ObservationEvent> _queue = new(1000);
    private Thread? _drainThread;

    public void Start()
    {
        _drainThread = new Thread(DrainLoop) { IsBackground = true, Name = "OrangeCandy.Observer" };
        _drainThread.Start();
    }

    public void Enqueue(ObservationEvent evt)
    {
        _queue.TryAdd(evt);
    }

    private void DrainLoop()
    {
        var batch = new List<ObservationEvent>(50);

        while (!_queue.IsCompleted)
        {
            try
            {
                batch.Clear();

                if (_queue.TryTake(out var first, TimeSpan.FromMilliseconds(50)))
                {
                    batch.Add(first);
                    while (batch.Count < 50 && _queue.TryTake(out var more, TimeSpan.Zero))
                        batch.Add(more);
                }

                if (batch.Count == 0) continue;

                Send(batch);
            }
            catch (InvalidOperationException) { break; }
            catch (Exception ex)
            {
                logger.LogDebug(ex, "Failed to send observation batch");
            }
        }
    }

    private void Send(List<ObservationEvent> batch)
    {
        try
        {
            var json = System.Text.Json.JsonSerializer.Serialize(batch);
            var content = new StringContent(json, System.Text.Encoding.UTF8, "application/json");
            var response = httpClient.Send(new HttpRequestMessage(HttpMethod.Post, "/api/observe") { Content = content });
            logger.LogDebug("Sent {Count} observation events", batch.Count);
        }
        catch (Exception ex)
        {
            logger.LogDebug(ex, "Observation send failed");
        }
    }

    public void Dispose()
    {
        _queue.CompleteAdding();
        _drainThread?.Join(TimeSpan.FromSeconds(3));

        var remaining = new List<ObservationEvent>();
        while (_queue.TryTake(out var evt))
            remaining.Add(evt);
        if (remaining.Count > 0)
            Send(remaining);

        _queue.Dispose();
    }
}
