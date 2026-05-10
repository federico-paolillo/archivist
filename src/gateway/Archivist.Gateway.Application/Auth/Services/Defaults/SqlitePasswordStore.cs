using Archivist.Gateway.Application.Auth.Options;

using Microsoft.Data.Sqlite;
using Microsoft.Extensions.Options;

namespace Archivist.Gateway.Application.Auth.Services.Defaults;

/// <summary>
/// Reads the personal user's stored Argon2id password hash from SQLite.
/// </summary>
public sealed class SqlitePasswordStore(IOptions<AuthOptions> options) : IPasswordStore
{
    private const string PersonalUserId = "01ASB2XFCZJY7WHZ2FNRTMQJCT";

    public async Task<string?> GetPasswordHashAsync(CancellationToken ct = default)
    {
        var sqlitePath = options.Value.SqlitePath;

        if (string.IsNullOrWhiteSpace(sqlitePath))
        {
            return null;
        }

        var connectionString = $"Data Source={sqlitePath}";

        await using var connection = new SqliteConnection(connectionString);
        await connection.OpenAsync(ct);

        const string sql = "SELECT password_hash FROM users WHERE id = @id;";

        await using var cmd = connection.CreateCommand();
        cmd.CommandText = sql;
        cmd.Parameters.AddWithValue("@id", PersonalUserId);

        var result = await cmd.ExecuteScalarAsync(ct);

        return result is DBNull or null ? null : (string)result;
    }
}