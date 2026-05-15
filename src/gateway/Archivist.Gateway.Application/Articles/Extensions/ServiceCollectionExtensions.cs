using Archivist.Gateway.Application.ArticleArtifacts;
using Archivist.Gateway.Application.ArticleArtifacts.Defaults;
using Archivist.Gateway.Application.Articles.Defaults;
using Archivist.Gateway.Application.Configuration;

using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;

namespace Archivist.Gateway.Application.Articles.Extensions;

/// <summary>
/// Registers article API application services.
/// </summary>
public static class ServiceCollectionExtensions
{
    /// <summary>
    /// Adds Gateway article endpoint services.
    /// </summary>
    public static IServiceCollection AddArticles(this IServiceCollection services, IConfiguration configuration)
    {
        ArgumentNullException.ThrowIfNull(services);
        ArgumentNullException.ThrowIfNull(configuration);

        services.AddSingleton(serviceProvider =>
        {
            var resolvedConfiguration = serviceProvider.GetRequiredService<IConfiguration>();
            var dataDirectory = resolvedConfiguration.GetValue<string>(Settings.DataDirectoryKey);
            if (string.IsNullOrWhiteSpace(dataDirectory))
            {
                dataDirectory = "/data";
            }

            return new ArticleArtifactPaths(dataDirectory);
        });
        services.AddScoped<IArticleArtifactDeletion, FileSystemArticleArtifactDeletion>();
        services.AddScoped<IArticleDeleteService, EfArticleDeleteService>();

        return services;
    }
}