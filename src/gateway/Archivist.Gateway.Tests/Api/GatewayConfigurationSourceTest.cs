using Archivist.Gateway.Api;
using Archivist.Gateway.Application.Configuration;
using Archivist.Gateway.Application.Telegram;
using Archivist.Gateway.Application.Telegram.Extensions;

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
        var configuration = new ConfigurationBuilder()
            .AddInMemoryCollection(new Dictionary<string, string?>
            {
                ["Telegram:BotToken"] = "fake-token",
                ["Telegram:AllowedUserId"] = "12345",
                ["Telegram:WebhookSecret"] = "secret",
            })
            .Build();

        var services = new ServiceCollection();
        services.AddSingleton<IConfiguration>(configuration);
        services.AddTelegram(configuration);

        using var provider = services.BuildServiceProvider();

        var settings = provider.GetRequiredService<IOptions<TelegramSettings>>().Value;

        Assert.Equal("fake-token", settings.BotToken);
        Assert.Equal(12345, settings.AllowedUserId);
        Assert.Equal("secret", settings.WebhookSecret);
    }
}