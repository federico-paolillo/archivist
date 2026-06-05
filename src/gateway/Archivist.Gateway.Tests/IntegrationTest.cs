using Archivist.Gateway.Application.Auth.Services;
using Archivist.Gateway.Application.Configuration;

using Microsoft.AspNetCore.Mvc.Testing;
using Microsoft.AspNetCore.TestHost;

using Xunit.Abstractions;

namespace Archivist.Gateway.Tests;

public abstract class IntegrationTest(
    ITestOutputHelper testOutputHelper
) : IDisposable
{
    private WebApplicationFactory<Program>? _webApplicationFactory;
    private readonly string _testRoot = Path.Combine(
        Path.GetTempPath(),
        $"archivist-gateway-tests-{Guid.NewGuid():N}");

    public void Dispose()
    {
        _webApplicationFactory?.Dispose();

        if (Directory.Exists(_testRoot))
        {
            Directory.Delete(_testRoot, recursive: true);
        }
    }

    protected void PrepareEnvironment(Action<IServiceCollection>? configureTestServices = null)
    {
        PrepareEnvironment(Environments.Development, configureTestServices);
    }

    protected void PrepareEnvironment(
        string environmentName,
        Action<IServiceCollection>? configureTestServices = null,
        Action<ConfigurationBuilder>? configureConfiguration = null
    )
    {
        Directory.CreateDirectory(_testRoot);

        var cfgBuilder = new ConfigurationBuilder();
        cfgBuilder.AddInMemoryCollection(new Dictionary<string, string?>
        {
            [Settings.SqlitePathKey] = Path.Combine(_testRoot, "archive.db"),
            [Settings.DataDirectoryKey] = Path.Combine(_testRoot, "data"),
            [Settings.GatewayPublicHostsKey] = environmentName == Environments.Production ? null : "localhost",
            [$"{Settings.TelegramSection}:BotToken"] = "test-bot-token",
            [$"{Settings.TelegramSection}:WebhookSecret"] = "test-webhook-secret",
        });

        configureConfiguration?.Invoke(cfgBuilder);

        var cfg = cfgBuilder.Build();

        _webApplicationFactory = new WebApplicationFactory<Program>()
            .WithWebHostBuilder(builder =>
            {
                builder.ConfigureLogging(l => l.AddProvider(new XUnitLoggerProvider(testOutputHelper)));
                builder.UseEnvironment(environmentName);
                builder.ConfigureAppConfiguration((_, configuration) => configuration.AddConfiguration(cfg));
                builder.ConfigureTestServices(services =>
                {
                    // Replace the real auth bootstrap with a no-op so integration tests
                    // do not require a database or AUTH_BOOTSTRAP_PASSWORD by default.
                    // Tests that specifically exercise auth must override this stub.
                    services.AddSingleton<IAuthBootstrapService>(new NoOpAuthBootstrapService());

                    configureTestServices?.Invoke(services);
                });
            });
    }

    protected HttpClient CreateHttpClient()
    {
        if (_webApplicationFactory is null)
        {
            throw new InvalidOperationException($"No WebApplicationFactory Did you call {nameof(PrepareEnvironment)}");
        }

        return _webApplicationFactory.CreateClient();
    }

    protected T GetRequiredService<T>()
        where T : notnull
    {
        if (_webApplicationFactory is null)
        {
            throw new InvalidOperationException($"No WebApplicationFactory Did you call {nameof(PrepareEnvironment)}");
        }

        return _webApplicationFactory.Services.GetRequiredService<T>();
    }

    private sealed class NoOpAuthBootstrapService : IAuthBootstrapService
    {
        public Task InitializeAsync(CancellationToken ct = default) => Task.CompletedTask;
    }
}