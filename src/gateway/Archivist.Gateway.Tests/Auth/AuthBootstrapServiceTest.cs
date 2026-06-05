using Archivist.Gateway.Application.Auth;
using Archivist.Gateway.Application.Auth.Services.Defaults;

using Microsoft.Data.Sqlite;
using Microsoft.Extensions.Logging.Abstractions;

namespace Archivist.Gateway.Tests.Auth;

/// <summary>
/// Integration tests for <see cref="AuthBootstrapService"/> covering bootstrap scenarios.
/// Uses a temporary SQLite file per test to provide real persistence behavior.
/// </summary>
public sealed class AuthBootstrapServiceTest : IDisposable
{
    private const string PersonalUserId = "01ASB2XFCZJY7WHZ2FNRTMQJCT";
    private const long PersonalTelegramUserId = 1559957191;

    private readonly string _dbPath = Path.Combine(
        Path.GetTempPath(),
        $"archivist-test-{Guid.NewGuid():N}.db");

    private readonly PasswordValidator _passwordValidator = new();
    private readonly Argon2idPasswordHasher _passwordHasher = new();

    public void Dispose()
    {
        // Clean up the temporary database file.
        if (File.Exists(_dbPath))
        {
            File.Delete(_dbPath);
        }
    }

    private static string ValidBootstrapPassword() => new('a', PasswordValidator.RequiredLength);

    private AuthBootstrapService CreateService(
        string? bootstrapPassword = null,
        string? sqlitePath = null)
    {
        var settings = new AuthSettings
        {
            SqlitePath = sqlitePath ?? _dbPath,
            BootstrapPassword = bootstrapPassword,
        };

        return new AuthBootstrapService(
            settings,
            _passwordValidator,
            _passwordHasher,
            NullLogger<AuthBootstrapService>.Instance);
    }

    private async Task<string?> ReadPasswordHashAsync()
    {
        await using var connection = new SqliteConnection($"Data Source={_dbPath}");
        await connection.OpenAsync();

        await using var cmd = connection.CreateCommand();
        cmd.CommandText = "SELECT password_hash FROM users WHERE id = @id;";
        cmd.Parameters.AddWithValue("@id", PersonalUserId);

        var result = await cmd.ExecuteScalarAsync();
        return result is DBNull or null ? null : (string)result;
    }

    private async Task<long?> ReadTelegramUserIdAsync()
    {
        await using var connection = new SqliteConnection($"Data Source={_dbPath}");
        await connection.OpenAsync();

        await using var cmd = connection.CreateCommand();
        cmd.CommandText = "SELECT telegram_user_id FROM users WHERE id = @id;";
        cmd.Parameters.AddWithValue("@id", PersonalUserId);

        var result = await cmd.ExecuteScalarAsync();
        return result is DBNull or null ? null : Convert.ToInt64(result);
    }

    private async Task<bool> PersonalUserExistsAsync()
    {
        await using var connection = new SqliteConnection($"Data Source={_dbPath}");
        await connection.OpenAsync();

        await using var cmd = connection.CreateCommand();
        cmd.CommandText = "SELECT COUNT(*) FROM users WHERE id = @id;";
        cmd.Parameters.AddWithValue("@id", PersonalUserId);

        var result = await cmd.ExecuteScalarAsync();
        return Convert.ToInt64(result) == 1;
    }

    private async Task SeedExistingHashAsync(string existingHash)
    {
        await using var connection = new SqliteConnection($"Data Source={_dbPath}");
        await connection.OpenAsync();

        await using var createTable = connection.CreateCommand();
        createTable.CommandText = """
            CREATE TABLE IF NOT EXISTS users (
                id TEXT NOT NULL PRIMARY KEY,
                telegram_user_id TEXT NULL UNIQUE,
                password_hash TEXT NULL
            );
            """;
        await createTable.ExecuteNonQueryAsync();

        await using var insert = connection.CreateCommand();
        insert.CommandText = "INSERT INTO users (id, telegram_user_id, password_hash) VALUES (@id, NULL, @hash);";
        insert.Parameters.AddWithValue("@id", PersonalUserId);
        insert.Parameters.AddWithValue("@hash", existingHash);
        await insert.ExecuteNonQueryAsync();
    }

    [Fact]
    public async Task InitializeAsync_WhenNoUserAndValidBootstrapPassword_StoresArgon2idHash()
    {
        var bootstrapPassword = ValidBootstrapPassword();
        var service = CreateService(bootstrapPassword: bootstrapPassword);

        await service.InitializeAsync();

        var storedHash = await ReadPasswordHashAsync();

        Assert.NotNull(storedHash);
        Assert.StartsWith("$argon2id$", storedHash);
    }

    [Fact]
    public async Task InitializeAsync_WhenNoUserAndValidBootstrapPassword_CreatesPersonalUserRow()
    {
        var service = CreateService(bootstrapPassword: ValidBootstrapPassword());

        await service.InitializeAsync();

        Assert.True(await PersonalUserExistsAsync());
    }

    [Fact]
    public async Task InitializeAsync_SeedsPersonalTelegramUserId()
    {
        var service = CreateService(bootstrapPassword: ValidBootstrapPassword());

        await service.InitializeAsync();

        Assert.Equal(PersonalTelegramUserId, await ReadTelegramUserIdAsync());
    }

    [Fact]
    public async Task InitializeAsync_WhenTelegramUserAlreadyMapped_DoesNotOverwriteMapping()
    {
        var existingHash = _passwordHasher.Hash(ValidBootstrapPassword());
        await SeedExistingHashAsync(existingHash);

        await using (var connection = new SqliteConnection($"Data Source={_dbPath}"))
        {
            await connection.OpenAsync();

            await using var update = connection.CreateCommand();
            update.CommandText = "UPDATE users SET telegram_user_id = @telegram_user_id WHERE id = @id;";
            update.Parameters.AddWithValue("@telegram_user_id", 11111);
            update.Parameters.AddWithValue("@id", PersonalUserId);
            await update.ExecuteNonQueryAsync();
        }

        var service = CreateService(bootstrapPassword: null);

        await service.InitializeAsync();

        Assert.Equal(11111, await ReadTelegramUserIdAsync());
    }

    [Fact]
    public async Task InitializeAsync_WhenHashAlreadyExists_PreservesExistingHash()
    {
        var existingHash = _passwordHasher.Hash(ValidBootstrapPassword());
        await SeedExistingHashAsync(existingHash);

        // Call without providing a bootstrap password — it should not be required.
        var service = CreateService(bootstrapPassword: null);

        await service.InitializeAsync();

        var storedHash = await ReadPasswordHashAsync();

        Assert.Equal(existingHash, storedHash);
    }

    [Fact]
    public async Task InitializeAsync_WhenHashAlreadyExists_DoesNotRequireBootstrapPassword()
    {
        var existingHash = _passwordHasher.Hash(ValidBootstrapPassword());
        await SeedExistingHashAsync(existingHash);

        var service = CreateService(bootstrapPassword: null);

        // Should not throw even though AUTH_BOOTSTRAP_PASSWORD is not set.
        var exception = await Record.ExceptionAsync(() => service.InitializeAsync());
        Assert.Null(exception);
    }

    [Fact]
    public async Task InitializeAsync_WhenNoHashAndNoBootstrapPassword_Throws()
    {
        var service = CreateService(bootstrapPassword: null);

        await Assert.ThrowsAsync<InvalidOperationException>(() => service.InitializeAsync());
    }

    [Fact]
    public async Task InitializeAsync_WhenBootstrapPasswordInvalid_Throws()
    {
        // Too short — not 2048 characters.
        var service = CreateService(bootstrapPassword: "tooshort");

        await Assert.ThrowsAsync<InvalidOperationException>(() => service.InitializeAsync());
    }

    [Fact]
    public async Task InitializeAsync_StoredHash_VerifiesAgainstBootstrapPassword()
    {
        var bootstrapPassword = ValidBootstrapPassword();
        var service = CreateService(bootstrapPassword: bootstrapPassword);

        await service.InitializeAsync();

        var storedHash = await ReadPasswordHashAsync();

        Assert.NotNull(storedHash);
        Assert.True(_passwordHasher.Verify(bootstrapPassword, storedHash));
    }

    [Fact]
    public async Task InitializeAsync_CalledTwice_DoesNotOverwriteHash()
    {
        var bootstrapPassword = ValidBootstrapPassword();
        var service = CreateService(bootstrapPassword: bootstrapPassword);

        await service.InitializeAsync();
        var firstHash = await ReadPasswordHashAsync();

        // Second call should reuse existing hash, not write a new one.
        await service.InitializeAsync();
        var secondHash = await ReadPasswordHashAsync();

        Assert.Equal(firstHash, secondHash);
    }

    [Fact]
    public async Task InitializeAsync_WithMissingSqlitePath_Throws()
    {
        var service = CreateService(bootstrapPassword: ValidBootstrapPassword(), sqlitePath: null);

        // Override sqlitePath to empty.
        var settings = new AuthSettings
        {
            SqlitePath = string.Empty,
            BootstrapPassword = ValidBootstrapPassword(),
        };

        var emptyPathService = new AuthBootstrapService(
            settings,
            _passwordValidator,
            _passwordHasher,
            NullLogger<AuthBootstrapService>.Instance);

        await Assert.ThrowsAsync<InvalidOperationException>(() => emptyPathService.InitializeAsync());
    }
}