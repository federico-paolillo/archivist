using Archivist.Gateway.Api;
using Archivist.Gateway.Application.Configuration;
using Archivist.Gateway.Application.Telegram;
using Archivist.Gateway.Application.Telegram.Defaults;
using Archivist.Gateway.Application.Telegram.Extensions;

using Microsoft.Extensions.Http;
using Microsoft.Extensions.Options;

namespace Archivist.Gateway.Tests.Api;

public sealed class GatewayConfigurationSourceTest
{
    [Fact]
    public void GatewayConfigurationSources_PreserveExistingConfigurationSources()
    {
        var configuration = new ConfigurationBuilder()
            .AddInMemoryCollection(new Dictionary<string, string?>
            {
                [Settings.SqlitePathKey] = "from-memory.db",
            })
            .AddGatewayConfigurationSources()
            .Build();

        Assert.Equal("from-memory.db", configuration.GetValue<string>(Settings.SqlitePathKey));
    }

    [Fact]
    public void GatewayConfigurationSources_MapPrefixedHierarchicalEnvironmentVariables()
    {
        const string envName = "ARCHIVIST_TestSection__Value";
        var oldValue = Environment.GetEnvironmentVariable(envName);

        try
        {
            Environment.SetEnvironmentVariable(envName, "from-env");

            var configuration = new ConfigurationBuilder()
                .AddGatewayConfigurationSources()
                .Build();

            Assert.Equal("from-env", configuration.GetValue<string>("TestSection:Value"));
        }
        finally
        {
            Environment.SetEnvironmentVariable(envName, oldValue);
        }
    }

    [Fact]
    public void TelegramSettings_BindFromHierarchicalConfiguration()
    {
        var configuration = CreateTelegramConfiguration("fake-token", "secret");

        var services = new ServiceCollection();
        services.AddSingleton<IConfiguration>(configuration);
        services.AddTelegram(configuration);

        using var provider = services.BuildServiceProvider();

        var settings = provider.GetRequiredService<IOptions<TelegramSettings>>().Value;

        Assert.Equal("fake-token", settings.BotToken);
        Assert.Equal("secret", settings.WebhookSecret);
    }

    [Fact]
    public void TelegramSettings_ValidConfiguration_PassesStartupValidation()
    {
        var configuration = CreateTelegramConfiguration("fake-token", "secret");
        var services = new ServiceCollection();
        services.AddSingleton<IConfiguration>(configuration);
        services.AddTelegram(configuration);

        using var provider = services.BuildServiceProvider();

        provider.GetRequiredService<IStartupValidator>().Validate();
    }

    [Theory]
    [InlineData("", "secret", "Telegram:BotToken is required.")]
    [InlineData("  ", "secret", "Telegram:BotToken is required.")]
    [InlineData("fake-token", "", "Telegram:WebhookSecret is required.")]
    [InlineData("fake-token", "  ", "Telegram:WebhookSecret is required.")]
    public void TelegramSettings_InvalidConfiguration_FailsStartupValidation(
        string botToken,
        string webhookSecret,
        string expectedFailure)
    {
        var configuration = CreateTelegramConfiguration(botToken, webhookSecret);
        var services = new ServiceCollection();
        services.AddSingleton<IConfiguration>(configuration);
        services.AddTelegram(configuration);

        using var provider = services.BuildServiceProvider();

        var exception = Assert.Throws<OptionsValidationException>(
            () => provider.GetRequiredService<IStartupValidator>().Validate());

        Assert.Contains(expectedFailure, exception.Failures);
    }

    [Fact]
    public void AddTelegram_RegistersHttpTelegramClient()
    {
        var configuration = CreateTelegramConfiguration("fake-token", "secret");
        var services = new ServiceCollection();
        services.AddSingleton<IConfiguration>(configuration);
        services.AddTelegram(configuration);

        using var provider = services.BuildServiceProvider();

        var client = provider.GetRequiredService<ITelegramClient>();

        Assert.IsType<HttpTelegramClient>(client);
    }

    [Fact]
    public void AddTelegram_RemovesHttpClientFactoryLoggingForTelegramClient()
    {
        var configuration = CreateTelegramConfiguration("fake-token", "secret");
        var services = new ServiceCollection();
        services.AddSingleton<IConfiguration>(configuration);
        services.AddTelegram(configuration);

        using var provider = services.BuildServiceProvider();

        var options = provider
            .GetRequiredService<IOptionsMonitor<HttpClientFactoryOptions>>()
            .Get(nameof(ITelegramClient));

        var suppressDefaultLoggingProperty = typeof(HttpClientFactoryOptions).GetProperty(
            "SuppressDefaultLogging",
            System.Reflection.BindingFlags.Instance | System.Reflection.BindingFlags.NonPublic);
        var loggingBuilderActionsProperty = typeof(HttpClientFactoryOptions).GetProperty(
            "LoggingBuilderActions",
            System.Reflection.BindingFlags.Instance | System.Reflection.BindingFlags.NonPublic);

        Assert.NotNull(suppressDefaultLoggingProperty);
        Assert.NotNull(loggingBuilderActionsProperty);
        Assert.True((bool)suppressDefaultLoggingProperty.GetValue(options)!);

        var loggingBuilderActions = Assert.IsAssignableFrom<System.Collections.IEnumerable>(
            loggingBuilderActionsProperty.GetValue(options));
        Assert.Empty(loggingBuilderActions);
    }

    private static IConfiguration CreateTelegramConfiguration(
        string botToken,
        string webhookSecret)
    {
        return new ConfigurationBuilder()
            .AddInMemoryCollection(new Dictionary<string, string?>
            {
                ["Telegram:BotToken"] = botToken,
                ["Telegram:WebhookSecret"] = webhookSecret,
            })
            .Build();
    }
}