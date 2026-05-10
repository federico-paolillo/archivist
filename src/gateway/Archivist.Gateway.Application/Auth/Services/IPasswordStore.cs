namespace Archivist.Gateway.Application.Auth.Services;

/// <summary>
/// Reads the personal user's stored Argon2id password hash from the persistence layer.
/// </summary>
public interface IPasswordStore
{
    /// <summary>
    /// Returns the stored Argon2id PHC hash for the personal user,
    /// or <c>null</c> if no hash has been stored yet.
    /// </summary>
    Task<string?> GetPasswordHashAsync(CancellationToken ct = default);
}