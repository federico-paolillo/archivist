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

        services.AddOptions<TelegramSettings>()
            .BindConfiguration(TelegramSettings.Section)
            .Validate(static settings => !string.IsNullOrWhiteSpace(settings.BotToken), "Telegram:BotToken is required.")
            .Validate(static settings => !string.IsNullOrWhiteSpace(settings.WebhookSecret), "Telegram:WebhookSecret is required.")
            .Validate(static settings => settings.AllowedUserId > 0, "Telegram:AllowedUserId must be greater than zero.")
            .ValidateOnStart();

        services.AddHttpClient<ITelegramClient, HttpTelegramClient>()
            .RemoveAllLoggers();
        services.AddScoped<TelegramWebhookHandler>();

        // TELING-004: notification dispatcher
        services.AddScoped<ITelegramNotificationRepository, EfTelegramNotificationRepository>();
        services.AddScoped<TelegramNotificationDispatcher>();
        services.AddHostedService<TelegramNotificationDispatcherService>();

        return services;
    }
}