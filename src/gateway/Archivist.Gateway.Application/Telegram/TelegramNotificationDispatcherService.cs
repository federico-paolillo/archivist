namespace Archivist.Gateway.Application.Telegram;

using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;

/// <summary>
/// Background service that periodically dispatches pending Telegram notifications and cleans up expired ones.
/// </summary>
public sealed partial class TelegramNotificationDispatcherService(
    IServiceScopeFactory scopeFactory,
    ILogger<TelegramNotificationDispatcherService> logger) : BackgroundService
{
    private static readonly TimeSpan PollingInterval = TimeSpan.FromSeconds(10);

    /// <inheritdoc />
    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        LogStarted(logger);

        while (!stoppingToken.IsCancellationRequested)
        {
            await RunOneCycleAsync(stoppingToken).ConfigureAwait(false);

            try
            {
                await Task.Delay(PollingInterval, stoppingToken).ConfigureAwait(false);
            }
            catch (OperationCanceledException)
            {
                break;
            }
        }

        LogStopped(logger);
    }

    private async Task RunOneCycleAsync(CancellationToken cancellationToken)
    {
        try
        {
            await using var scope = scopeFactory.CreateAsyncScope();
            var dispatcher = scope.ServiceProvider.GetRequiredService<TelegramNotificationDispatcher>();

            await dispatcher.DispatchPendingAsync(cancellationToken).ConfigureAwait(false);
            await dispatcher.CleanUpExpiredAsync(cancellationToken).ConfigureAwait(false);
        }
#pragma warning disable CA1031 // Dispatcher errors must not crash the background service loop.
        catch (Exception ex) when (!cancellationToken.IsCancellationRequested)
#pragma warning restore CA1031
        {
            LogCycleError(logger, ex);
        }
    }

    [LoggerMessage(Level = LogLevel.Information, Message = "Telegram notification dispatcher started")]
    private static partial void LogStarted(ILogger logger);

    [LoggerMessage(Level = LogLevel.Information, Message = "Telegram notification dispatcher stopped")]
    private static partial void LogStopped(ILogger logger);

    [LoggerMessage(Level = LogLevel.Error, Message = "Telegram notification dispatcher cycle error; will retry next interval")]
    private static partial void LogCycleError(ILogger logger, Exception ex);
}
