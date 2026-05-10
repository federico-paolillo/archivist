namespace Archivist.Gateway.Application.Telegram.Extensions;

using Archivist.Gateway.Application.Persistence;
using Archivist.Gateway.Application.Persistence.Defaults;
using Archivist.Gateway.Application.Telegram.Defaults;

using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;

/// <summary>
/// Registers Telegram integration services.
/// </summary>
public static class ServiceCollectionExtensions
{
    /// <summary>
    /// Adds Telegram webhook handling, Bot API client, and notification dispatcher.
    /// </summary>
    public static IServiceCollection AddTelegram(this IServiceCollection services, IConfiguration configuration)
    {
        ArgumentNullException.ThrowIfNull(configuration);

        services.Configure<TelegramOptions>(opts =>
        {
            opts.BotToken = configuration["TELEGRAM_BOT_TOKEN"] ?? string.Empty;
            opts.WebhookSecret = configuration["TELEGRAM_WEBHOOK_SECRET"] ?? string.Empty;

            if (long.TryParse(configuration["TELEGRAM_ALLOWED_USER_ID"], out var allowedUserId))
            {
                opts.AllowedUserId = allowedUserId;
            }
        });

        services.AddHttpClient<ITelegramClient, HttpTelegramClient>();
        services.AddScoped<TelegramWebhookHandler>();

        // TELING-004: notification dispatcher
        services.AddScoped<ITelegramNotificationRepository, EfTelegramNotificationRepository>();
        services.AddScoped<TelegramNotificationDispatcher>();
        services.AddHostedService<TelegramNotificationDispatcherService>();

        return services;
    }
}