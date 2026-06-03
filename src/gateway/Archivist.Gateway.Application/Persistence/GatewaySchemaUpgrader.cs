namespace Archivist.Gateway.Application.Persistence;

using Microsoft.EntityFrameworkCore;

/// <summary>
/// Applies idempotent Gateway SQLite schema upgrades not covered by EnsureCreated.
/// </summary>
public static class GatewaySchemaUpgrader
{
    /// <summary>
    /// Ensures queued jobs can carry W3C trace context for asynchronous Worker processing.
    /// </summary>
    public static async Task EnsureJobTraceCarrierColumnsAsync(
        ArchivistDbContext db,
        CancellationToken cancellationToken = default)
    {
        ArgumentNullException.ThrowIfNull(db);

        if (!await JobsTableHasColumnAsync(db, "traceparent", cancellationToken).ConfigureAwait(false))
        {
            await db.Database
                .ExecuteSqlRawAsync("ALTER TABLE jobs ADD COLUMN traceparent TEXT", cancellationToken)
                .ConfigureAwait(false);
        }

        if (!await JobsTableHasColumnAsync(db, "tracestate", cancellationToken).ConfigureAwait(false))
        {
            await db.Database
                .ExecuteSqlRawAsync("ALTER TABLE jobs ADD COLUMN tracestate TEXT", cancellationToken)
                .ConfigureAwait(false);
        }
    }

    private static async Task<bool> JobsTableHasColumnAsync(
        ArchivistDbContext db,
        string columnName,
        CancellationToken cancellationToken)
    {
        var connection = db.Database.GetDbConnection();
        var shouldClose = connection.State != System.Data.ConnectionState.Open;
        if (shouldClose)
        {
            await connection.OpenAsync(cancellationToken).ConfigureAwait(false);
        }

        try
        {
            await using var command = connection.CreateCommand();
            command.CommandText = "PRAGMA table_info(jobs)";
            await using var reader = await command.ExecuteReaderAsync(cancellationToken).ConfigureAwait(false);
            while (await reader.ReadAsync(cancellationToken).ConfigureAwait(false))
            {
                if (string.Equals(reader.GetString(1), columnName, StringComparison.OrdinalIgnoreCase))
                {
                    return true;
                }
            }

            return false;
        }
        finally
        {
            if (shouldClose)
            {
                await connection.CloseAsync().ConfigureAwait(false);
            }
        }
    }
}