using Archivist.Gateway.Application.Ping.Models;

namespace Archivist.Gateway.Application.Ping.Services.Defaults;

public sealed class PingService(
  TimeProvider timeProvider
) : IPingService
{
    public Pong Ping()
    {
        var rightNow = timeProvider.GetUtcNow();

        return new Pong(rightNow);
    }
}