using Archivist.Gateway.Application.Auth;
using Archivist.Gateway.Application.Auth.Services.Defaults;

using Microsoft.Data.Sqlite;

namespace Archivist.Gateway.Tests.Auth;

public sealed class SqlitePasswordStoreTest
{
    [Fact]
    public async Task GetPasswordCredentialsAsync_OnePasswordBearingUser_ReturnsUserIdAndHash()
    {
        var sqlitePath = await CreateUsersDatabaseAsync(
            ("01ASB2XFCZJY7WHZ2FNRTMQJCT", "$argon2id$hash-1"),
            ("01BSB2XFCZJY7WHZ2FNRTMQJCT", null));
        var store = CreateStore(sqlitePath);

        var credentials = await store.GetPasswordCredentialsAsync(CancellationToken.None);
        var credential = Assert.Single(credentials);

        Assert.Equal("01ASB2XFCZJY7WHZ2FNRTMQJCT", credential.UserId);
        Assert.Equal("$argon2id$hash-1", credential.PasswordHash);
    }

    [Fact]
    public async Task GetPasswordCredentialsAsync_NoPasswordBearingUser_ReturnsEmpty()
    {
        var sqlitePath = await CreateUsersDatabaseAsync(
            ("01ASB2XFCZJY7WHZ2FNRTMQJCT", null),
            ("01BSB2XFCZJY7WHZ2FNRTMQJCT", null));
        var store = CreateStore(sqlitePath);

        var credentials = await store.GetPasswordCredentialsAsync(CancellationToken.None);

        Assert.Empty(credentials);
    }

    [Fact]
    public async Task GetPasswordCredentialsAsync_MultiplePasswordBearingUsers_ReturnsAllOrderedByUserId()
    {
        var sqlitePath = await CreateUsersDatabaseAsync(
            ("01BSB2XFCZJY7WHZ2FNRTMQJCT", "$argon2id$hash-2"),
            ("01ASB2XFCZJY7WHZ2FNRTMQJCT", "$argon2id$hash-1"));
        var store = CreateStore(sqlitePath);

        var credentials = await store.GetPasswordCredentialsAsync(CancellationToken.None);

        Assert.Collection(
            credentials,
            credential =>
            {
                Assert.Equal("01ASB2XFCZJY7WHZ2FNRTMQJCT", credential.UserId);
                Assert.Equal("$argon2id$hash-1", credential.PasswordHash);
            },
            credential =>
            {
                Assert.Equal("01BSB2XFCZJY7WHZ2FNRTMQJCT", credential.UserId);
                Assert.Equal("$argon2id$hash-2", credential.PasswordHash);
            });
    }

    private static SqlitePasswordStore CreateStore(string sqlitePath)
    {
        return new SqlitePasswordStore(new AuthSettings { SqlitePath = sqlitePath });
    }

    private static async Task<string> CreateUsersDatabaseAsync(params (string Id, string? PasswordHash)[] users)
    {
        var sqlitePath = Path.Combine(Path.GetTempPath(), $"{Guid.NewGuid():N}.db");

        await using var connection = new SqliteConnection($"Data Source={sqlitePath}");
        await connection.OpenAsync(CancellationToken.None);

        await using (var create = connection.CreateCommand())
        {
            create.CommandText = """
                CREATE TABLE users (
                    id TEXT NOT NULL PRIMARY KEY,
                    telegram_user_id TEXT NULL UNIQUE,
                    password_hash TEXT NULL
                );
                """;
            await create.ExecuteNonQueryAsync(CancellationToken.None);
        }

        foreach (var user in users)
        {
            await using var insert = connection.CreateCommand();
            insert.CommandText = "INSERT INTO users (id, telegram_user_id, password_hash) VALUES (@id, NULL, @password_hash);";
            insert.Parameters.AddWithValue("@id", user.Id);
            insert.Parameters.AddWithValue("@password_hash", user.PasswordHash is null ? DBNull.Value : user.PasswordHash);
            await insert.ExecuteNonQueryAsync(CancellationToken.None);
        }

        return sqlitePath;
    }
}