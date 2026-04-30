using Microsoft.AspNetCore.Mvc.Testing;
using Microsoft.AspNetCore.TestHost;

using Xunit.Abstractions;

namespace Archivist.Gateway.Tests;

public abstract class IntegrationTest(
    ITestOutputHelper testOutputHelper
) : IDisposable
{
    private WebApplicationFactory<Program>? _webApplicationFactory;

    public void Dispose()
    {
        _webApplicationFactory?.Dispose();
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
        var cfgBuilder = new ConfigurationBuilder();

        configureConfiguration?.Invoke(cfgBuilder);

        var cfg = cfgBuilder.Build();

        _webApplicationFactory = new WebApplicationFactory<Program>()
            .WithWebHostBuilder(builder =>
            {
                builder.ConfigureLogging(l => l.AddProvider(new XUnitLoggerProvider(testOutputHelper)));
                builder.UseEnvironment(environmentName);
                builder.UseConfiguration(cfg);
                builder.ConfigureTestServices(services => configureTestServices?.Invoke(services));
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
}