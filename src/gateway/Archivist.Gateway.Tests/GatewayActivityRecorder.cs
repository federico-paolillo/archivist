namespace Archivist.Gateway.Tests;

using System.Diagnostics;

using Archivist.Gateway.Application.Observability;

internal sealed class GatewayActivityRecorder : IDisposable
{
    private readonly ActivityListener _listener;
    private readonly object _sync = new();
    private readonly List<Activity> _ended = [];

    public GatewayActivityRecorder()
    {
        _listener = new ActivityListener
        {
            ShouldListenTo = static source => source.Name == ArchivistTelemetry.ActivitySourceName,
            Sample = static (ref ActivityCreationOptions<ActivityContext> _) => ActivitySamplingResult.AllData,
            ActivityStopped = activity =>
            {
                lock (_sync)
                {
                    _ended.Add(activity);
                }
            },
        };

        ActivitySource.AddActivityListener(_listener);
    }

    public IReadOnlyList<Activity> Ended
    {
        get
        {
            lock (_sync)
            {
                return _ended.ToList();
            }
        }
    }

    public void Dispose() => _listener.Dispose();
}