using Archivist.Gateway.Application.Auth.Services;
using Archivist.Gateway.Application.Persistence;

namespace Archivist.Gateway.Api;

public static class WebApplicationExtensions
{
    public static async Task PrepareAsync(this WebApplication app)
    {
        ArgumentNullException.ThrowIfNull(app);

        using var scope = app.Services.CreateScope();

        var db = scope.ServiceProvider.GetRequiredService<ArchivistDbContext>();

        await db.Database.EnsureCreatedAsync();

        var authBootstrap = app.Services.GetRequiredService<IAuthBootstrapService>();

        await authBootstrap.InitializeAsync();
    }
}