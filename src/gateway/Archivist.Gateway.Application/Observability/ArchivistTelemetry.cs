using System.Diagnostics;

namespace Archivist.Gateway.Application.Observability;

/// <summary>
/// Provides Gateway-owned OpenTelemetry activity sources and attribute names.
/// </summary>
public static class ArchivistTelemetry
{
    public const string ServiceName = "archivist-gateway";
    public const string ActivitySourceName = "archivist.gateway";

    public static readonly ActivitySource ActivitySource = new(ActivitySourceName);

    public const string ArticleId = "archivist.article.id";
    public const string JobId = "archivist.job.id";
    public const string TelegramUpdateId = "archivist.telegram.update_id";
    public const string Outcome = "archivist.outcome";
    public const string Stage = "archivist.stage";
}