using System.Collections.Concurrent;

namespace Archivist.Gateway.Application.Auth.Services.Defaults;

/// <summary>
/// V0 in-memory login throttle. Enforces per-IP and global failed-attempt limits.
/// Counters reset on gateway restart by design.
/// </summary>
public sealed class InMemoryLoginThrottle : ILoginThrottle
{
    private const int MaxPerIpFailures = 10;
    private const int MaxGlobalFailures = 50;

    private readonly ConcurrentDictionary<string, int> _perIpCounters = new(StringComparer.OrdinalIgnoreCase);
    private int _globalCounter;

    public bool IsThrottled(string sourceIp)
    {
        ArgumentNullException.ThrowIfNull(sourceIp);

        if (_globalCounter >= MaxGlobalFailures)
        {
            return true;
        }

        if (_perIpCounters.TryGetValue(sourceIp, out var perIp) && perIp >= MaxPerIpFailures)
        {
            return true;
        }

        return false;
    }

    public void RecordFailure(string sourceIp)
    {
        ArgumentNullException.ThrowIfNull(sourceIp);

        Interlocked.Increment(ref _globalCounter);
        _perIpCounters.AddOrUpdate(sourceIp, 1, (_, v) => v + 1);
    }

    public void RecordSuccess(string sourceIp)
    {
        ArgumentNullException.ThrowIfNull(sourceIp);

        _perIpCounters.TryRemove(sourceIp, out _);
    }
}