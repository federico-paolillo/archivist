using Archivist.Gateway.Api.Ping.Models;

using Microsoft.Extensions.Time.Testing;

using Xunit.Abstractions;

namespace Archivist.Gateway.Tests.Api;

public sealed class ArticlesEndpointTests(
    ITestOutputHelper testOutputHelper
) : IntegrationTest(testOutputHelper)
{
    [Fact]
    public async Task PingEndpoint_OnGet_ReturnsPong()
    {
        var rightNow = new DateTimeOffset(2026, 4, 30, 23, 09, 00, TimeSpan.Zero);

        var fakeTimeProvider = new FakeTimeProvider(rightNow);

        PrepareEnvironment(configureTestServices: svc => svc.AddSingleton<TimeProvider>(fakeTimeProvider));

        using var httpClient = CreateHttpClient();

        var response = await httpClient.GetAsync("/ping");

        Assert.True(response.IsSuccessStatusCode);

        var pong = await response.Content.ReadFromJsonAsync<PongDto>();

        Assert.NotNull(pong);
        Assert.Equal(rightNow.ToUnixTimeSeconds(), pong.Timestamp);
    }
}