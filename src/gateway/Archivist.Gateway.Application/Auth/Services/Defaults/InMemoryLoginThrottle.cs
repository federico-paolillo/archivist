namespace Archivist.Gateway.Application.Auth.Services.Defaults;

/// <summary>
/// V0 in-memory login throttle. Enforces per-IP and global failed-attempt limits.
/// Failed-attempt windows reset on gateway restart by design.
/// </summary>
public sealed class InMemoryLoginThrottle(TimeProvider timeProvider) : ILoginThrottle
{
    private const int MaxPerIpFailures = 10;
    private const int MaxGlobalFailures = 50;
    private static readonly TimeSpan FailureWindow = TimeSpan.FromMinutes(5);

    private readonly object _gate = new();
    private readonly Dictionary<string, Queue<DateTimeOffset>> _perIpFailures = new(StringComparer.OrdinalIgnoreCase);
    private readonly Queue<DateTimeOffset> _globalFailures = new();

    /// <inheritdoc />
    public bool IsThrottled(string sourceIp)
    {
        ArgumentNullException.ThrowIfNull(sourceIp);

        lock (_gate)
        {
            var now = timeProvider.GetUtcNow();
            PruneExpired(now);

            return _globalFailures.Count >= MaxGlobalFailures ||
                (_perIpFailures.TryGetValue(sourceIp, out var perIpFailures) &&
                    perIpFailures.Count >= MaxPerIpFailures);
        }
    }

    /// <inheritdoc />
    public void RecordFailure(string sourceIp)
    {
        ArgumentNullException.ThrowIfNull(sourceIp);

        lock (_gate)
        {
            var now = timeProvider.GetUtcNow();
            PruneExpired(now);

            if (_globalFailures.Count < MaxGlobalFailures)
            {
                _globalFailures.Enqueue(now);
            }

            if (!_perIpFailures.TryGetValue(sourceIp, out var perIpFailures))
            {
                perIpFailures = new Queue<DateTimeOffset>();
                _perIpFailures.Add(sourceIp, perIpFailures);
            }

            if (perIpFailures.Count < MaxPerIpFailures)
            {
                perIpFailures.Enqueue(now);
            }
        }
    }

    /// <inheritdoc />
    public void RecordSuccess(string sourceIp)
    {
        ArgumentNullException.ThrowIfNull(sourceIp);

        lock (_gate)
        {
            PruneExpired(timeProvider.GetUtcNow());
            _perIpFailures.Remove(sourceIp);
        }
    }

    private void PruneExpired(DateTimeOffset now)
    {
        var cutoff = now.Subtract(FailureWindow);

        PruneQueue(_globalFailures, cutoff);

        foreach (var (sourceIp, failures) in _perIpFailures.ToArray())
        {
            PruneQueue(failures, cutoff);
            if (failures.Count == 0)
            {
                _perIpFailures.Remove(sourceIp);
            }
        }
    }

    private static void PruneQueue(Queue<DateTimeOffset> failures, DateTimeOffset cutoff)
    {
        while (failures.TryPeek(out var recordedAt) && recordedAt <= cutoff)
        {
            failures.Dequeue();
        }
    }
}