using Archivist.Gateway.Application.Auth;

using Microsoft.Data.Sqlite;

namespace Archivist.Gateway.Application.Auth.Services.Defaults;

/// <summary>
/// Reads password-bearing users' stored Argon2id password hashes from SQLite.
/// </summary>
public sealed class SqlitePasswordStore(AuthSettings settings) : IPasswordStore
{
    public async Task<IReadOnlyList<PasswordCredential>> GetPasswordCredentialsAsync(CancellationToken ct = default)
    {
        var sqlitePath = settings.SqlitePath;

        if (string.IsNullOrWhiteSpace(sqlitePath))
        {
            return [];
        }

        var connectionString = $"Data Source={sqlitePath}";

        await using var connection = new SqliteConnection(connectionString);
        await connection.OpenAsync(ct);

        const string sql = """
            SELECT id, password_hash
            FROM users
            WHERE password_hash IS NOT NULL AND password_hash <> ''
            ORDER BY id;
            """;

        await using var cmd = connection.CreateCommand();
        cmd.CommandText = sql;

        var credentials = new List<PasswordCredential>();
        await using var reader = await cmd.ExecuteReaderAsync(ct);
        while (await reader.ReadAsync(ct))
        {
            credentials.Add(new PasswordCredential(reader.GetString(0), reader.GetString(1)));
        }

        return credentials;
    }
}