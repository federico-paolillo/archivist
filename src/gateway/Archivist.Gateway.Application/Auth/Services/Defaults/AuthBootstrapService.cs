using Archivist.Gateway.Application.Auth.Options;

using Microsoft.Data.Sqlite;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;

namespace Archivist.Gateway.Application.Auth.Services.Defaults;

/// <summary>
/// Ensures the personal user row exists and that a password hash is stored.
/// Bootstraps the password hash from the configured bootstrap secret only when the row has no stored hash.
/// </summary>
public sealed partial class AuthBootstrapService(
    IOptions<AuthOptions> options,
    IPasswordValidator passwordValidator,
    IPasswordHasher passwordHasher,
    ILogger<AuthBootstrapService> logger
) : IAuthBootstrapService
{
    private const string PersonalUserId = "01ASB2XFCZJY7WHZ2FNRTMQJCT";

    public async Task InitializeAsync(CancellationToken ct = default)
    {
        var sqlitePath = options.Value.SqlitePath;

        if (string.IsNullOrWhiteSpace(sqlitePath))
        {
            throw new InvalidOperationException("SQLITE_PATH is required for auth bootstrap.");
        }

        var connectionString = $"Data Source={sqlitePath}";

        await using var connection = new SqliteConnection(connectionString);
        await connection.OpenAsync(ct);

        await EnsureSchemaAsync(connection, ct);
        await EnsurePersonalUserAsync(connection, ct);
        await EnsurePasswordHashAsync(connection, ct);
    }

    private static async Task EnsureSchemaAsync(SqliteConnection connection, CancellationToken ct)
    {
        // Ensure users table exists with required columns.
        // This is a no-op if the table already exists with compatible schema.
        const string sql = """
            CREATE TABLE IF NOT EXISTS users (
                id TEXT NOT NULL PRIMARY KEY,
                telegram_user_id TEXT NULL UNIQUE,
                password_hash TEXT NULL
            );
            """;

        await using var cmd = connection.CreateCommand();
        cmd.CommandText = sql;
        await cmd.ExecuteNonQueryAsync(ct);
    }

    private static async Task EnsurePersonalUserAsync(SqliteConnection connection, CancellationToken ct)
    {
        // Insert the personal user row if it does not already exist.
        // Do not overwrite telegram_user_id or password_hash if the row already exists.
        const string sql = """
            INSERT INTO users (id, telegram_user_id, password_hash)
            VALUES (@id, NULL, NULL)
            ON CONFLICT(id) DO NOTHING;
            """;

        await using var cmd = connection.CreateCommand();
        cmd.CommandText = sql;
        cmd.Parameters.AddWithValue("@id", PersonalUserId);
        await cmd.ExecuteNonQueryAsync(ct);
    }

    private async Task EnsurePasswordHashAsync(SqliteConnection connection, CancellationToken ct)
    {
        var existingHash = await ReadPasswordHashAsync(connection, ct);

        if (existingHash is not null)
        {
            // Hash already exists — do not require or use bootstrap password.
            LogHashAlreadyStored(logger);
            return;
        }

        // No hash stored — bootstrap is required.
        var bootstrapPassword = options.Value.BootstrapPassword;

        if (string.IsNullOrEmpty(bootstrapPassword))
        {
            throw new InvalidOperationException(
                "AUTH_BOOTSTRAP_PASSWORD is required to initialize the personal user password hash. " +
                "Set the environment variable and restart the gateway.");
        }

        if (!passwordValidator.IsValid(bootstrapPassword))
        {
            throw new InvalidOperationException(
                "AUTH_BOOTSTRAP_PASSWORD is invalid. " +
                "The password must be exactly 2048 printable ASCII characters (0x20–0x7E).");
        }

        // Hash the bootstrap password and store it. Never log the plaintext.
        var hash = passwordHasher.Hash(bootstrapPassword);

        await StorePasswordHashAsync(connection, hash, ct);

        LogBootstrapComplete(logger);
    }

    private static async Task<string?> ReadPasswordHashAsync(SqliteConnection connection, CancellationToken ct)
    {
        const string sql = "SELECT password_hash FROM users WHERE id = @id;";

        await using var cmd = connection.CreateCommand();
        cmd.CommandText = sql;
        cmd.Parameters.AddWithValue("@id", PersonalUserId);

        var result = await cmd.ExecuteScalarAsync(ct);

        return result is DBNull or null ? null : (string)result;
    }

    private static async Task StorePasswordHashAsync(SqliteConnection connection, string hash, CancellationToken ct)
    {
        const string sql = "UPDATE users SET password_hash = @hash WHERE id = @id;";

        await using var cmd = connection.CreateCommand();
        cmd.CommandText = sql;
        cmd.Parameters.AddWithValue("@hash", hash);
        cmd.Parameters.AddWithValue("@id", PersonalUserId);
        await cmd.ExecuteNonQueryAsync(ct);
    }

    [LoggerMessage(Level = LogLevel.Information, Message = "Auth: personal user password hash is already stored; skipping bootstrap.")]
    private static partial void LogHashAlreadyStored(ILogger logger);

    [LoggerMessage(Level = LogLevel.Information, Message = "Auth: personal user password hash bootstrapped successfully.")]
    private static partial void LogBootstrapComplete(ILogger logger);
}